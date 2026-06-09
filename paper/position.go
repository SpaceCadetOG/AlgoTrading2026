package paper

type PaperPosition struct {
	ID              string   `json:"id"`
	Venue           string   `json:"venue"`
	Symbol          string   `json:"symbol"`
	CanonicalSymbol string   `json:"canonicalSymbol"`
	Strategy        string   `json:"strategy"`
	Playbook        string   `json:"playbook"`
	Side            string   `json:"side"`
	Score           float64  `json:"score"`
	Confidence      float64  `json:"confidence"`
	EntryPrice      float64  `json:"entryPrice"`
	MarkPrice       float64  `json:"markPrice"`
	LastPrice       float64  `json:"lastPrice"`
	Quantity        float64  `json:"quantity"`
	InitialQuantity float64  `json:"initialQuantity"`
	Margin          float64  `json:"margin"`
	Leverage        float64  `json:"leverage"`
	Stop            float64  `json:"stop"`
	TP1             float64  `json:"tp1"`
	TP2             float64  `json:"tp2"`
	TP3             float64  `json:"tp3"`
	TrailingStop    float64  `json:"trailingStop"`
	BreakEvenPrice  float64  `json:"breakEvenPrice"`
	RealizedPnL     float64  `json:"realizedPnL"`
	OpenPnL         float64  `json:"openPnL"`
	FeesPaid        float64  `json:"feesPaid"`
	FundingPaid     float64  `json:"fundingPaid"`
	OpenedTime      int64    `json:"openedTime"`
	UpdatedTime     int64    `json:"updatedTime"`
	ClosedTime      int64    `json:"closedTime,omitempty"`
	ExitReason      string   `json:"exitReason,omitempty"`
	MFER            float64  `json:"mfer"`
	MAER            float64  `json:"maer"`
	EntryReasons    []string `json:"entryReasons"`
	DecisionReasons []string `json:"decisionReasons"`
	Provenance      string   `json:"provenance"`
	TP1Taken        bool     `json:"tp1Taken"`
	TP2Taken        bool     `json:"tp2Taken"`
	TP3Taken        bool     `json:"tp3Taken"`
}
