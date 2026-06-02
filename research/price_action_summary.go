package research

import (
	"encoding/json"
	"os"
)

type PriceActionSummary struct {
	Candles                int      `json:"candles"`
	SidewaysAreas          int      `json:"sidewaysAreas"`
	BullishInitiations     int      `json:"bullishInitiations"`
	BearishInitiations     int      `json:"bearishInitiations"`
	BullishRejections      int      `json:"bullishRejections"`
	BearishRejections      int      `json:"bearishRejections"`
	SupportResistanceFlips int      `json:"supportResistanceFlips"`
	OpenDrives             int      `json:"openDrives"`
	StrongHighs            int      `json:"strongHighs"`
	StrongLows             int      `json:"strongLows"`
	WeakHighs              int      `json:"weakHighs"`
	WeakLows               int      `json:"weakLows"`
	FailedAuctions         int      `json:"failedAuctions"`
	Notes                  []string `json:"notes"`
}

func BuildPriceActionSummary(rows []PriceActionFeatureRow) PriceActionSummary {
	summary := PriceActionSummary{
		Candles: len(rows),
		Notes: []string{
			"Research-only price action features use OHLCV candle approximations.",
			"Session and daily open use UTC day reset for crypto; 00:00 UTC displays as 19:00 CT during standard time planning context.",
			"Institutional aggression, failed auctions, and support/resistance flips are approximations until trade tape and historical L2 are joined.",
		},
	}
	for _, row := range rows {
		if row.SidewaysDetected {
			summary.SidewaysAreas++
		}
		if row.InitiationDetected {
			switch row.InitiationDirection {
			case "up":
				summary.BullishInitiations++
			case "down":
				summary.BearishInitiations++
			}
			if row.SessionOpenLevel == row.InitiationStartPrice {
				summary.OpenDrives++
			}
		}
		if row.RejectionDetected {
			switch row.RejectionDirection {
			case "bullish":
				summary.BullishRejections++
			case "bearish":
				summary.BearishRejections++
			}
		}
		if row.SupportResistanceFlipDetected {
			summary.SupportResistanceFlips++
		}
		if row.StrongHigh {
			summary.StrongHighs++
		}
		if row.StrongLow {
			summary.StrongLows++
		}
		if row.WeakHigh {
			summary.WeakHighs++
		}
		if row.WeakLow {
			summary.WeakLows++
		}
		if row.FailedAuctionDetected {
			summary.FailedAuctions++
		}
	}
	return summary
}

func WritePriceActionSummaryJSON(path string, summary PriceActionSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
