package orderflow

type FootprintBar struct {
	Timestamp       int64
	Venue           string
	Symbol          string
	VenueSymbol     string
	CanonicalSymbol string
	StartTimestamp  int64
	EndTimestamp    int64
	Open            float64
	High            float64
	Low             float64
	Close           float64
	Levels          []FootprintLevel
	TotalBidVolume  float64
	TotalAskVolume  float64
	Delta           float64
	Volume          float64
}

type FootprintLevel struct {
	Price       float64
	BidVolume   float64
	AskVolume   float64
	TotalVolume float64
	Delta       float64
}

type VolumeCluster struct {
	Price  float64
	Volume float64
}

type ImbalanceDirection string

const (
	ImbalanceAsk ImbalanceDirection = "ask"
	ImbalanceBid ImbalanceDirection = "bid"
)

type Imbalance struct {
	Price     float64
	Direction ImbalanceDirection
	Ratio     float64
}

type StackedImbalance struct {
	Direction  ImbalanceDirection
	StartPrice float64
	EndPrice   float64
	Count      int
}

type UnfinishedBusiness struct {
	High bool
	Low  bool
}

type LargeTrade struct {
	Price  float64
	Volume float64
}
