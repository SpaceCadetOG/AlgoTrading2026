package system

type LoggingService struct {
	*BaseService
}

func NewLoggingService() *LoggingService {
	return &LoggingService{BaseService: NewBaseService("logging")}
}

type PositionService struct {
	*BaseService
}

func NewPositionService() *PositionService {
	return &PositionService{BaseService: NewBaseService("position")}
}

type MarketDataService struct {
	*BaseService
}

func NewMarketDataService() *MarketDataService {
	return &MarketDataService{BaseService: NewBaseService("market_data")}
}

type OrderService struct {
	*BaseService
}

func NewOrderService() *OrderService {
	return &OrderService{BaseService: NewBaseService("order")}
}

type RiskService struct {
	*BaseService
}

func NewRiskService() *RiskService {
	return &RiskService{BaseService: NewBaseService("risk")}
}
