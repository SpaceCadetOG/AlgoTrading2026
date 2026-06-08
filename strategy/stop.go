package strategy

type StopRule struct {
	Name                   string  `json:"name"`
	SourceBook             string  `json:"sourceBook"`
	StopType               string  `json:"stopType"`
	Placement              string  `json:"placement"`
	BufferBps              float64 `json:"bufferBps"`
	CatastrophicMultiplier float64 `json:"catastrophicMultiplier"`
}
