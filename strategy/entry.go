package strategy

type EntryRule struct {
	Name                 string   `json:"name"`
	SourceBook           string   `json:"sourceBook"`
	EntryType            string   `json:"entryType"`
	Trigger              string   `json:"trigger"`
	ConfirmationRequired []string `json:"confirmationRequired"`
	InvalidIf            []string `json:"invalidIf"`
	OffsetBps            float64  `json:"offsetBps"`
}
