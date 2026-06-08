package paper

type TelemetryEvent struct {
	Timestamp    int64    `json:"timestamp"`
	Type         string   `json:"type"`
	Symbol       string   `json:"symbol"`
	Venue        string   `json:"venue"`
	Side         string   `json:"side"`
	Strategy     string   `json:"strategy"`
	Playbook     string   `json:"playbook"`
	Score        float64  `json:"score"`
	Confidence   float64  `json:"confidence"`
	Decision     string   `json:"decision"`
	Entry        float64  `json:"entry"`
	Exit         float64  `json:"exit"`
	Mark         float64  `json:"mark"`
	Last         float64  `json:"last"`
	Reasons      []string `json:"reasons"`
	RealizedPnL  float64  `json:"realizedPnl"`
	MFE          float64  `json:"mfe"`
	MAE          float64  `json:"mae"`
	HoldTimeSec  int64    `json:"holdTimeSec"`
	StopDistance float64  `json:"stopDistance"`
}

type RecorderStatus struct {
	TradeTape string `json:"tradeTape"`
	L2        string `json:"l2"`
	Equity    string `json:"equity"`
	Events    string `json:"events"`
	Funding   string `json:"funding"`
}
