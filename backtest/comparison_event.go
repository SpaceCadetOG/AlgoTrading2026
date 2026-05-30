package backtest

import (
	"strings"

	"AlgoTrading2026/exchanges"
)

type BacktesterComparisonResult struct {
	Symbol                string   `json:"symbol"`
	Candles               int      `json:"candles"`
	ForLoopTrades         int      `json:"forLoopTrades"`
	EventDrivenOrders     int      `json:"eventDrivenOrders"`
	EventDrivenFills      int      `json:"eventDrivenFills"`
	ForLoopFinalEquity    float64  `json:"forLoopFinalEquity"`
	EventDrivenFinalPnL   float64  `json:"eventDrivenFinalPnL"`
	PnLDifference         float64  `json:"pnlDifference"`
	AssumptionsDifference string   `json:"assumptionsDifference"`
	Recommendation        string   `json:"recommendation"`
	AssumptionComparisons []string `json:"assumptionComparisons"`
}

func CompareForLoopVsEventDriven(
	candles []exchanges.Candle,
	forLoopConfig Config,
	eventConfig EventDrivenBacktestConfig,
) (BacktesterComparisonResult, error) {
	sample := append([]exchanges.Candle(nil), candles...)
	if eventConfig.MaxCandles > 0 && len(sample) > eventConfig.MaxCandles {
		sample = sample[:eventConfig.MaxCandles]
	}

	symbol := comparisonSymbol(sample, forLoopConfig, eventConfig)
	if forLoopConfig.Symbol == "" {
		forLoopConfig.Symbol = symbol
	}
	if forLoopConfig.StartingBalance == 0 {
		forLoopConfig.StartingBalance = 10000
	}
	if eventConfig.Symbol == "" {
		eventConfig.Symbol = symbol
	}
	if eventConfig.StartingCash == 0 {
		eventConfig.StartingCash = forLoopConfig.StartingBalance
	}
	eventConfig.AllowLiveOrders = false

	forLoopReport := RunBookSignalBacktest(sample, forLoopConfig)
	eventResult, err := NewEventDrivenBacktester(eventConfig).Run(sample)
	if err != nil {
		return BacktesterComparisonResult{}, err
	}

	forLoopPnL := forLoopReport.EndingBalance - forLoopReport.StartingBalance
	assumptions := []string{
		"for-loop backtester uses direct candle, signal, and simulated fill model",
		"event-driven backtester routes candles through liquidity, order book, strategy, OMS, and market simulator queues",
		"for-loop fills use fee and slippage models; event-driven fills use deterministic crossed-book test liquidity",
		"results are expected to differ because market assumptions differ",
	}

	return BacktesterComparisonResult{
		Symbol:                symbol,
		Candles:               len(sample),
		ForLoopTrades:         forLoopReport.Metrics.TotalTrades,
		EventDrivenOrders:     eventResult.OrdersCreated,
		EventDrivenFills:      eventResult.OrdersFilled,
		ForLoopFinalEquity:    forLoopReport.EndingBalance,
		EventDrivenFinalPnL:   eventResult.FinalPnL,
		PnLDifference:         eventResult.FinalPnL - forLoopPnL,
		AssumptionsDifference: strings.Join(assumptions, "; "),
		Recommendation:        "Use for-loop backtests for fast signal research and event-driven backtests for system/OMS/market-simulator assumption validation.",
		AssumptionComparisons: assumptions,
	}, nil
}

func comparisonSymbol(candles []exchanges.Candle, forLoopConfig Config, eventConfig EventDrivenBacktestConfig) string {
	switch {
	case forLoopConfig.Symbol != "":
		return forLoopConfig.Symbol
	case eventConfig.Symbol != "":
		return eventConfig.Symbol
	case len(candles) > 0:
		return candles[0].Symbol
	default:
		return ""
	}
}
