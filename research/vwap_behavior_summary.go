package research

import (
	"encoding/json"
	"os"

	"AlgoTrading2026/vwap"
)

type VWAPBehaviorSummary struct {
	Candles       int                `json:"candles"`
	TimeAboveVWAP int                `json:"timeAboveVWAP"`
	TimeBelowVWAP int                `json:"timeBelowVWAP"`
	Touches       int                `json:"touches"`
	Crosses       int                `json:"crosses"`
	Bounces       int                `json:"bounces"`
	Breaks        int                `json:"breaks"`
	Chops         int                `json:"chops"`
	BounceRate    float64            `json:"bounceRate"`
	BreakRate     float64            `json:"breakRate"`
	ChopRate      float64            `json:"chopRate"`
	MagnetStats   []vwap.MagnetStats `json:"magnetStats"`
}

func BuildVWAPBehaviorSummary(interactions []vwap.Interaction, magnetStats []vwap.MagnetStats) VWAPBehaviorSummary {
	summary := VWAPBehaviorSummary{
		Candles:     len(interactions),
		MagnetStats: magnetStats,
	}

	for _, row := range interactions {
		if row.AboveVWAP {
			summary.TimeAboveVWAP++
		}
		if row.BelowVWAP {
			summary.TimeBelowVWAP++
		}
		if row.TouchedVWAP {
			summary.Touches++
		}
		if row.CrossedVWAP {
			summary.Crosses++
		}
		switch row.Reaction {
		case vwap.ReactionBounce:
			summary.Bounces++
		case vwap.ReactionBreak:
			summary.Breaks++
		case vwap.ReactionChop:
			summary.Chops++
		}
	}

	if summary.Touches > 0 {
		denominator := float64(summary.Touches)
		summary.BounceRate = float64(summary.Bounces) / denominator
		summary.BreakRate = float64(summary.Breaks) / denominator
		summary.ChopRate = float64(summary.Chops) / denominator
	}

	return summary
}

func WriteVWAPBehaviorSummaryJSON(path string, summary VWAPBehaviorSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}
