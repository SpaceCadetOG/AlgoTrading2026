package research

import (
	"encoding/json"
	"os"
)

type ResearchSummary struct {
	Candles int `json:"candles"`

	BestChapter2Strategy string  `json:"bestChapter2Strategy"`
	BestChapter2NetPnL   float64 `json:"bestChapter2NetPnL"`

	BestChapter4Strategy string  `json:"bestChapter4Strategy"`
	BestChapter4NetPnL   float64 `json:"bestChapter4NetPnL"`

	BestChapter5Strategy string  `json:"bestChapter5Strategy"`
	BestChapter5NetPnL   float64 `json:"bestChapter5NetPnL"`

	WorstDrawdownStrategy string  `json:"worstDrawdownStrategy"`
	WorstDrawdown         float64 `json:"worstDrawdown"`

	MostActiveStrategy string `json:"mostActiveStrategy"`
	MostTrades         int    `json:"mostTrades"`

	FeatureRows  int `json:"featureRows"`
	LabelRows    int `json:"labelRows"`
	TrainingRows int `json:"trainingRows"`

	PairCandidates int `json:"pairCandidates"`
}

func WriteResearchSummaryJSON(path string, summary ResearchSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
