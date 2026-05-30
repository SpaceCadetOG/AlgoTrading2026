package backtest

import (
	"fmt"
	"time"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/system"
)

type EventDrivenBacktestConfig struct {
	Symbol              string
	StartingCash        float64
	MaxCandles          int
	UseSimulatedGateway bool
	AllowLiveOrders     bool
}

type EventDrivenBacktestResult struct {
	CandlesProcessed int
	OrdersCreated    int
	OrdersFilled     int
	FinalCash        float64
	FinalPosition    float64
	FinalPnL         float64
	ClockStart       time.Time
	ClockEnd         time.Time
	AuditEvents      int
}

type EventDrivenBacktester struct {
	Config EventDrivenBacktestConfig
}

func NewEventDrivenBacktester(config EventDrivenBacktestConfig) *EventDrivenBacktester {
	if config.StartingCash == 0 {
		config.StartingCash = 10000
	}
	config.AllowLiveOrders = false
	return &EventDrivenBacktester{Config: config}
}

func (b *EventDrivenBacktester) Run(candles []exchanges.Candle) (EventDrivenBacktestResult, error) {
	if len(candles) == 0 {
		return EventDrivenBacktestResult{FinalCash: b.Config.StartingCash}, nil
	}

	if b.Config.MaxCandles > 0 && len(candles) > b.Config.MaxCandles {
		candles = candles[:b.Config.MaxCandles]
	}

	iterator := NewCandleTimeIterator(nil, candles)
	clock := iterator.Clock()
	start := clock.Current()

	queues := system.NewQueues(64)
	lp := system.NewLiquidityProvider(queues)
	ob := system.NewOrderBook(queues)
	ts := system.NewTradingStrategy(queues)
	ts.Cash = b.Config.StartingCash
	om := system.NewOrderManager(queues)
	ms := system.NewMarketSimulator(queues)
	var gw *system.SimulatedGateway
	if b.Config.UseSimulatedGateway {
		gw = system.NewSimulatedGateway(queues)
		if err := gw.Start(); err != nil {
			return EventDrivenBacktestResult{}, err
		}
	}

	processed := 0
	for {
		candle, ok, err := iterator.Next()
		if err != nil {
			return EventDrivenBacktestResult{}, err
		}
		if !ok {
			break
		}

		b.publishCrossedBook(lp, candle, processed)
		processEventCycle(ob, ts, om, ms, gw)
		processed++
	}

	return EventDrivenBacktestResult{
		CandlesProcessed: processed,
		OrdersCreated:    len(om.Orders),
		OrdersFilled:     ts.Fills,
		FinalCash:        ts.Cash,
		FinalPosition:    ts.Position,
		FinalPnL:         ts.Cash - b.Config.StartingCash,
		ClockStart:       start,
		ClockEnd:         clock.Current(),
		AuditEvents:      len(om.Audits),
	}, nil
}

func (b *EventDrivenBacktester) publishCrossedBook(lp *system.LiquidityProvider, candle exchanges.Candle, index int) {
	close := candle.CloseFloat()
	if close <= 0 {
		close = 1
	}
	spread := close * 0.001
	if spread < 0.01 {
		spread = 0.01
	}

	ask := close - spread/2
	if ask <= 0 {
		ask = close
	}
	bid := close + spread/2
	size := 1.0
	symbol := b.Config.Symbol
	if symbol == "" {
		symbol = candle.Symbol
	}

	lp.InsertAsk(fmt.Sprintf("%s-%d-ask", symbol, index), ask, size)
	lp.InsertBid(fmt.Sprintf("%s-%d-bid", symbol, index), bid, size)
}

func processEventCycle(
	ob *system.OrderBook,
	ts *system.TradingStrategy,
	om *system.OrderManager,
	ms *system.MarketSimulator,
	gw *system.SimulatedGateway,
) {
	processUntilIdle(ob, ts, om, ms, gw)
	if gw != nil {
		gw.FillAllOrders()
	} else {
		ms.FillAllOrders()
	}
	processUntilIdle(ob, ts, om, ms, gw)
}

func processUntilIdle(
	ob *system.OrderBook,
	ts *system.TradingStrategy,
	om *system.OrderManager,
	ms *system.MarketSimulator,
	gw *system.SimulatedGateway,
) {
	for {
		worked := false
		if ob.ProcessNext() {
			worked = true
		}
		if ts.ProcessBookEvent() {
			worked = true
		}
		if om.ProcessStrategyOrder() {
			worked = true
		}
		if gw != nil {
			if gw.ProcessNextOrder() {
				worked = true
			}
		} else if ms.ProcessNext() {
			worked = true
		}
		if om.ProcessSimulatorAudit() {
			worked = true
		}
		if om.ProcessMarketResponse() {
			worked = true
		}
		if ts.ProcessOrderResponse() {
			worked = true
		}
		if !worked {
			return
		}
	}
}
