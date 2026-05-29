package backtest

type PortfolioPoint struct {
	Time int64

	Close    float64
	Signal   float64
	Position float64

	Holdings float64
	Cash     float64

	TotalEquity float64

	UnrealizedPnL float64
	RealizedPnL   float64
}
