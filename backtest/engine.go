package backtest

import (
	"fmt"
	"strconv"
	"time"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/execution"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/series"
	"AlgoTrading2026/strategy"
)

type Config struct {
	Venue           string
	Symbol          string
	Interval        string
	StartingBalance float64
	FixedNotional   float64
	FeeModel        FeeModel
	SlippageModel   SlippageModel
	CSVPath         string
	EquityCSVPath   string
	SignalCSVPath   string
}

type Engine struct {
	Config Config

	StartingBalance float64
	CurrentBalance  float64
	OpenPosition    *strategy.PositionState
	OpenEntryFee    float64
	OpenEntryTime   time.Time
	ClosedTrades    []strategy.Trade
	RejectReasons   []string
	Drawdown        DrawdownTracker
	EquityCurve     EquityCurve

	RiskEngine      *risk.Engine
	PositionManager *strategy.PositionManager
}

func NewEngine(config Config, riskEngine *risk.Engine) *Engine {
	if config.FixedNotional <= 0 {
		config.FixedNotional = 100
	}
	if riskEngine == nil {
		limits := risk.DefaultLimits()
		limits.LiveTradingEnabled = true
		riskEngine = risk.NewEngine(limits)
	}
	if config.FeeModel == nil {
		model := DefaultFeeModel()
		config.FeeModel = model
	}
	if config.SlippageModel == nil {
		model := DefaultSlippageModel()
		config.SlippageModel = model
	}
	if config.EquityCSVPath == "" {
		config.EquityCSVPath = "research/equity_curve.csv"
	}
	if config.SignalCSVPath == "" {
		config.SignalCSVPath = "research/signals.csv"
	}

	return &Engine{
		Config:          config,
		StartingBalance: config.StartingBalance,
		CurrentBalance:  config.StartingBalance,
		Drawdown:        NewDrawdownTracker(config.StartingBalance),
		RiskEngine:      riskEngine,
		PositionManager: strategy.NewPositionManager(),
	}
}

func (e *Engine) Run(candles []exchanges.Candle) Report {
	frame := series.FromCandles(candles)
	signals := series.BuyLowSellHighSignals(frame)

	return e.RunWithSignals(candles, signals)
}

func RunWithSignals(candles []exchanges.Candle, signals series.SignalResult, config Config) Report {
	return NewEngine(config, nil).RunWithSignals(candles, signals)
}

func RunBookSignalBacktest(candles []exchanges.Candle, config Config) Report {
	frame := series.FromCandles(candles)
	signals := series.BuyLowSellHighSignals(frame)

	return RunWithSignals(candles, signals, config)
}

func (e *Engine) RunWithSignals(candles []exchanges.Candle, signals series.SignalResult) Report {
	replay := NewReplay(candles)
	signalByTime := signalsByTime(signals)
	positionByTime := signalPositionsByTime(signals)
	for _, candle := range replay.Candles {
		e.onCandleSignal(candle, signalByTime[candle.StartTime], positionByTime[candle.StartTime])
	}

	metrics := e.metrics()
	return Report{
		Venue:           e.Config.Venue,
		Symbol:          e.Config.Symbol,
		Interval:        e.Config.Interval,
		CandleCount:     replay.Len(),
		StartingBalance: e.StartingBalance,
		EndingBalance:   e.EquityCurve.EndingEquity(e.CurrentBalance),
		Metrics:         metrics,
		RejectReasons:   append([]string(nil), e.RejectReasons...),
		CSVPath:         e.Config.CSVPath,
		EquityCSVPath:   e.Config.EquityCSVPath,
		SignalCSVPath:   e.Config.SignalCSVPath,
	}
}

func signalsByTime(signals series.SignalResult) map[int64]float64 {
	values := make(map[int64]float64, len(signals.Times))
	for i, timestamp := range signals.Times {
		if i >= len(signals.Signal) {
			break
		}
		values[timestamp] = signals.Signal[i]
	}
	return values
}

func signalPositionsByTime(signals series.SignalResult) map[int64]float64 {
	positions := make(map[int64]float64, len(signals.Times))
	for i, timestamp := range signals.Times {
		if i >= len(signals.Positions) {
			break
		}
		positions[timestamp] = signals.Positions[i]
	}
	return positions
}

func (e *Engine) onCandleSignal(candle exchanges.Candle, signal float64, positionChange float64) {
	e.markOpenPosition(candle)

	switch {
	case positionChange > 0 && e.OpenPosition == nil:
		e.openLong(candle)
	case positionChange < 0 && e.OpenPosition != nil:
		e.closeLong(candle, "book signal exit")
	}

	e.markOpenPosition(candle)
	e.appendPortfolioPoint(candle, signal, positionChange)
}

func (e *Engine) openLong(candle exchanges.Candle) {
	price := candle.CloseFloat()
	if price <= 0 {
		e.RejectReasons = append(e.RejectReasons, "invalid candle close price")
		return
	}

	size := e.Config.FixedNotional / price
	order := execution.OrderRequest{
		Venue:  e.Config.Venue,
		Symbol: e.Config.Symbol,
		Side:   execution.Buy,
		Type:   execution.Market,
		Size:   strconv.FormatFloat(size, 'f', -1, 64),
		Price:  strconv.FormatFloat(price, 'f', -1, 64),
	}

	decision := e.RiskEngine.CheckOrder(e.accountState(), e.positions(), order)
	if !decision.Approved {
		e.RejectReasons = append(e.RejectReasons, decision.Reason)
		return
	}

	fill, err := SimulateFill(order, candle, LiquidityTaker, e.Config.FeeModel, e.Config.SlippageModel)
	if err != nil {
		e.RejectReasons = append(e.RejectReasons, err.Error())
		return
	}

	e.CurrentBalance -= fill.Notional + fill.Fee
	e.OpenEntryFee = fill.Fee
	e.OpenEntryTime = fill.Timestamp
	position := strategy.PositionState{
		Venue:         fill.Venue,
		Symbol:        fill.Symbol,
		Side:          "LONG",
		Size:          fill.Size,
		EntryPrice:    fill.Price,
		MarkPrice:     fill.Price,
		UnrealizedPnL: 0,
		ExposureUSD:   fill.Notional,
		LastUpdate:    fill.Timestamp,
	}
	e.OpenPosition = &position
	e.syncPositionManager()
}

func (e *Engine) closeLong(candle exchanges.Candle, exitReason string) {
	if e.OpenPosition == nil {
		return
	}

	order := execution.OrderRequest{
		Venue:      e.OpenPosition.Venue,
		Symbol:     e.OpenPosition.Symbol,
		Side:       execution.Sell,
		Type:       execution.Market,
		Size:       strconv.FormatFloat(e.OpenPosition.Size, 'f', -1, 64),
		Price:      candle.Close,
		ReduceOnly: true,
	}

	fill, err := SimulateFill(order, candle, LiquidityTaker, e.Config.FeeModel, e.Config.SlippageModel)
	if err != nil {
		e.RejectReasons = append(e.RejectReasons, err.Error())
		return
	}

	opened := strategy.NewOpenedTrade(
		e.OpenPosition.Venue,
		e.OpenPosition.Symbol,
		e.OpenPosition.Side,
		e.OpenPosition.Size,
		e.OpenPosition.EntryPrice,
		e.OpenEntryTime,
	)
	closed := opened.Close(fill.Price, fill.Fee, exitReason, fill.Timestamp)
	closed.Fees += e.OpenEntryFee
	closed.RealizedPnL = closed.GrossPnL - closed.Fees
	e.CurrentBalance += fill.Notional - fill.Fee
	e.ClosedTrades = append(e.ClosedTrades, closed)
	e.OpenPosition = nil
	e.OpenEntryFee = 0
	e.OpenEntryTime = time.Time{}
	e.syncPositionManager()
}

func (e *Engine) accountState() strategy.AccountState {
	openPnL := 0.0
	openExposure := 0.0
	positionCount := 0
	if e.OpenPosition != nil {
		openPnL = e.OpenPosition.UnrealizedPnL
		openExposure = e.OpenPosition.ExposureUSD
		positionCount = 1
	}

	return strategy.AccountState{
		Venue:           e.Config.Venue,
		TotalEquity:     e.CurrentBalance + openExposure,
		AvailableEquity: e.CurrentBalance,
		OpenExposure:    openExposure,
		TotalOpenPnL:    openPnL,
		PositionCount:   positionCount,
	}
}

func (e *Engine) positions() []strategy.PositionState {
	if e.OpenPosition == nil {
		return nil
	}

	return []strategy.PositionState{*e.OpenPosition}
}

func (e *Engine) syncPositionManager() {
	e.PositionManager = strategy.NewPositionManager()
	if e.OpenPosition == nil {
		return
	}

	e.PositionManager.SyncPositions([]exchanges.Position{{
		Venue:  e.OpenPosition.Venue,
		Symbol: e.OpenPosition.Symbol,
		Side:   e.OpenPosition.Side,
		Size:   fmt.Sprintf("%f", e.OpenPosition.Size),
		Entry:  fmt.Sprintf("%f", e.OpenPosition.EntryPrice),
		PnL:    fmt.Sprintf("%f", e.OpenPosition.UnrealizedPnL),
		Lev:    fmt.Sprintf("%f", e.OpenPosition.Leverage),
	}})
}

func (e *Engine) markOpenPosition(candle exchanges.Candle) {
	if e.OpenPosition == nil {
		return
	}

	mark := candle.CloseFloat()
	e.OpenPosition.MarkPrice = mark
	e.OpenPosition.UnrealizedPnL = strategy.UnrealizedPnL(
		e.OpenPosition.Side,
		e.OpenPosition.Size,
		e.OpenPosition.EntryPrice,
		mark,
	)
	e.OpenPosition.ExposureUSD = strategy.ExposureUSD(e.OpenPosition.Size, mark)
	e.OpenPosition.LastUpdate = candle.EndUTC()
}

func (e *Engine) appendPortfolioPoint(candle exchanges.Candle, signal float64, positionChange float64) {
	close := candle.CloseFloat()
	holdings := 0.0
	unrealizedPnL := 0.0
	if e.OpenPosition != nil {
		holdings = e.OpenPosition.Size * close
		unrealizedPnL = e.OpenPosition.UnrealizedPnL
	}

	point := PortfolioPoint{
		Time:          candle.StartTime,
		Close:         close,
		Signal:        signal,
		Position:      positionChange,
		Holdings:      holdings,
		Cash:          e.CurrentBalance,
		TotalEquity:   e.CurrentBalance + holdings,
		UnrealizedPnL: unrealizedPnL,
		RealizedPnL:   e.realizedPnL(),
	}
	e.EquityCurve.Append(point)
	e.Drawdown = e.EquityCurve.Drawdown(e.StartingBalance)
}

func (e *Engine) realizedPnL() float64 {
	total := 0.0
	for _, trade := range e.ClosedTrades {
		total += trade.RealizedPnL
	}
	return total
}

func (e *Engine) metrics() Metrics {
	metrics := CalculateMetrics(e.ClosedTrades)
	drawdown := e.EquityCurve.Drawdown(e.StartingBalance)
	metrics.PeakEquity = drawdown.PeakEquity
	metrics.MaxDrawdown = drawdown.MaxDrawdown
	metrics.MaxDrawdownPct = drawdown.MaxDrawdownPct
	return metrics
}
