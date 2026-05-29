package features

type FeatureRow struct {
	Time     int64
	Venue    string
	Symbol   string
	Interval string

	Close  float64
	Volume float64

	SMA float64
	EMA float64
	APO float64

	MACD          float64
	MACDSignal    float64
	MACDHistogram float64

	RSI      float64
	Momentum float64
	StdDev   float64

	BollingerMiddle float64
	BollingerUpper  float64
	BollingerLower  float64
	BollingerWidth  float64

	DistanceFromSMA float64
	DistanceFromEMA float64

	Support                float64
	Resistance             float64
	DistanceFromSupport    float64
	DistanceFromResistance float64

	HourOfDay int
	DayOfWeek int
}
