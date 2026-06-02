package research

import (
	"encoding/json"
	"os"
)

type PriceActionStrategySummary struct {
	Candles           int      `json:"candles"`
	SRFlipSetups      int      `json:"srFlipSetups"`
	OpenDriveSetups   int      `json:"openDriveSetups"`
	ABCDSetups        int      `json:"abcdSetups"`
	SessionOpenSetups int      `json:"sessionOpenSetups"`
	DailyOpenSetups   int      `json:"dailyOpenSetups"`
	LongContexts      int      `json:"longContexts"`
	ShortContexts     int      `json:"shortContexts"`
	Invalidated       int      `json:"invalidated"`
	Notes             []string `json:"notes"`
}

func BuildPriceActionStrategySummary(candles int, rows []PriceActionStrategyStudyRow) PriceActionStrategySummary {
	summary := PriceActionStrategySummary{
		Candles: candles,
		Notes: []string{
			"Research-only setup study; rows are contexts, not orders or executable entries.",
			"Open-drive, S/R flips, AB=CD, session open, and daily open use OHLCV approximations.",
			"UTC day reset is used for crypto daily open; UTC 00:00 maps to the user's 19:00 CT planning context when applicable.",
		},
	}
	for _, row := range rows {
		switch row.Strategy {
		case "sr_flip":
			summary.SRFlipSetups++
		case "open_drive":
			summary.OpenDriveSetups++
		case "abcd":
			summary.ABCDSetups++
		case "session_open":
			summary.SessionOpenSetups++
		case "daily_open":
			summary.DailyOpenSetups++
		}
		switch row.Direction {
		case "long":
			summary.LongContexts++
		case "short":
			summary.ShortContexts++
		}
		if row.Invalidated {
			summary.Invalidated++
		}
	}
	return summary
}

func WritePriceActionStrategySummaryJSON(path string, summary PriceActionStrategySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
