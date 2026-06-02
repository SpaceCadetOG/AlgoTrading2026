package research

import (
	"encoding/json"
	"os"
)

type PriceActionPhase3Summary struct {
	Candles               int      `json:"candles"`
	DailyHighRetests      int      `json:"dailyHighRetests"`
	DailyLowRetests       int      `json:"dailyLowRetests"`
	WeeklyHighRetests     int      `json:"weeklyHighRetests"`
	WeeklyLowRetests      int      `json:"weeklyLowRetests"`
	StrongHighs           int      `json:"strongHighs"`
	StrongLows            int      `json:"strongLows"`
	WeakHighs             int      `json:"weakHighs"`
	WeakLows              int      `json:"weakLows"`
	FailedHighAuctions    int      `json:"failedHighAuctions"`
	FailedLowAuctions     int      `json:"failedLowAuctions"`
	LongContexts          int      `json:"longContexts"`
	ShortContexts         int      `json:"shortContexts"`
	BreakoutWatchContexts int      `json:"breakoutWatchContexts"`
	Invalidated           int      `json:"invalidated"`
	Notes                 []string `json:"notes"`
}

func BuildPriceActionPhase3Summary(candles int, rows []PriceActionPhase3StudyRow) PriceActionPhase3Summary {
	summary := PriceActionPhase3Summary{
		Candles: candles,
		Notes: []string{
			"Research-only setup study; rows are context observations, not executable trade instructions.",
			"Daily and weekly level retests require breach acceptance before retest detection.",
			"Strong/weak highs and lows use OHLCV rejection/aggression approximations.",
			"Failed auctions use close-back-inside logic against prior daily range levels.",
		},
	}
	for _, row := range rows {
		switch row.Strategy {
		case "daily_high_retest":
			summary.DailyHighRetests++
		case "daily_low_retest":
			summary.DailyLowRetests++
		case "weekly_high_retest":
			summary.WeeklyHighRetests++
		case "weekly_low_retest":
			summary.WeeklyLowRetests++
		case "strong_high":
			summary.StrongHighs++
		case "strong_low":
			summary.StrongLows++
		case "weak_high":
			summary.WeakHighs++
		case "weak_low":
			summary.WeakLows++
		case "failed_high_auction":
			summary.FailedHighAuctions++
		case "failed_low_auction":
			summary.FailedLowAuctions++
		}
		switch string(row.Direction) {
		case "long_context":
			summary.LongContexts++
		case "short_context":
			summary.ShortContexts++
		case "breakout_watch":
			summary.BreakoutWatchContexts++
		}
		if row.Invalidated {
			summary.Invalidated++
		}
	}
	return summary
}

func WritePriceActionPhase3SummaryJSON(path string, summary PriceActionPhase3Summary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
