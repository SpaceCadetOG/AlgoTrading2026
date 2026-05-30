package system

type TradingSimulation struct {
	Queues            *Queues
	LiquidityProvider *LiquidityProvider
	OrderBook         *OrderBook
	TradingStrategy   *TradingStrategy
	OrderManager      *OrderManager
	MarketSimulator   *MarketSimulator
}

type SimulationSummary struct {
	OrderID       string
	OrderStatus   OrderStatus
	RealizedPnL   float64
	Fills         int
	AuditEvents   int
	BookBestBid   float64
	BookBestAsk   float64
	Accepted      int
	OpenAfterFill int
}

func NewTestTradingSimulation(buffer int) *TradingSimulation {
	queues := NewQueues(buffer)
	return &TradingSimulation{
		Queues:            queues,
		LiquidityProvider: NewLiquidityProvider(queues),
		OrderBook:         NewOrderBook(queues),
		TradingStrategy:   NewTradingStrategy(queues),
		OrderManager:      NewOrderManager(queues),
		MarketSimulator:   NewMarketSimulator(queues),
	}
}

func (s *TradingSimulation) RunArbitrageExample() SimulationSummary {
	s.LiquidityProvider.InsertAsk("ask-1", 99, 1)
	s.LiquidityProvider.InsertBid("bid-1", 100, 1)

	s.OrderBook.ProcessNext()
	s.TradingStrategy.ProcessBookEvent()
	s.OrderBook.ProcessNext()
	s.TradingStrategy.ProcessBookEvent()
	s.OrderManager.ProcessStrategyOrder()
	s.OrderManager.ProcessStrategyOrder()
	s.MarketSimulator.ProcessNext()
	s.MarketSimulator.ProcessNext()
	s.OrderManager.ProcessSimulatorAudit()
	s.OrderManager.ProcessSimulatorAudit()
	s.OrderManager.ProcessMarketResponse()
	s.OrderManager.ProcessMarketResponse()
	s.TradingStrategy.ProcessOrderResponse()
	s.TradingStrategy.ProcessOrderResponse()
	accepted := len(s.MarketSimulator.orders)
	s.MarketSimulator.FillAllOrders()
	s.OrderManager.ProcessSimulatorAudit()
	s.OrderManager.ProcessSimulatorAudit()
	s.OrderManager.ProcessMarketResponse()
	s.OrderManager.ProcessMarketResponse()
	s.TradingStrategy.ProcessOrderResponse()
	s.TradingStrategy.ProcessOrderResponse()

	snapshot := s.OrderBook.Snapshot()
	internalID := s.OrderManager.ClientOrderIDs[s.TradingStrategy.LastOrder]
	status := s.OrderManager.Orders[internalID]
	return SimulationSummary{
		OrderID:       s.TradingStrategy.LastOrder,
		OrderStatus:   status.Status,
		RealizedPnL:   s.TradingStrategy.Realized,
		Fills:         s.TradingStrategy.Fills,
		AuditEvents:   len(s.OrderManager.Audits),
		BookBestBid:   snapshot.BestBid,
		BookBestAsk:   snapshot.BestAsk,
		Accepted:      accepted,
		OpenAfterFill: len(s.MarketSimulator.orders),
	}
}
