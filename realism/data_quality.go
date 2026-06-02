package realism

import (
	"math"
	"sort"
	"strconv"

	"AlgoTrading2026/exchanges"
)

type MarketDataQualityCheck struct {
	TotalCandles         int                 `json:"totalCandles"`
	MissingCandles       int                 `json:"missingCandles"`
	DuplicateTimestamps  int                 `json:"duplicateTimestamps"`
	OutOfOrderTimestamps int                 `json:"outOfOrderTimestamps"`
	ZeroNegativeOHLC     int                 `json:"zeroNegativeOhlc"`
	InvalidVolume        int                 `json:"invalidVolume"`
	LargeTimeGaps        int                 `json:"largeTimeGaps"`
	SuspiciousPriceJumps int                 `json:"suspiciousPriceJumps"`
	ExpectedIntervalMS   int64               `json:"expectedIntervalMs"`
	MaxObservedGapMS     int64               `json:"maxObservedGapMs"`
	MaxPriceJumpPct      float64             `json:"maxPriceJumpPct"`
	QualityRisk          DislocationSeverity `json:"qualityRisk"`
	Warnings             []string            `json:"warnings"`
}

func CheckMarketDataQuality(candles []exchanges.Candle, expectedIntervalMS int64) MarketDataQualityCheck {
	check := MarketDataQualityCheck{
		TotalCandles:       len(candles),
		ExpectedIntervalMS: expectedIntervalMS,
	}
	if len(candles) == 0 {
		check.QualityRisk = SeverityHigh
		check.Warnings = append(check.Warnings, "no_candles_available")
		return check
	}
	if check.ExpectedIntervalMS <= 0 {
		check.ExpectedIntervalMS = inferIntervalMS(candles)
	}

	seen := make(map[int64]bool, len(candles))
	var previousTime int64
	var previousClose float64
	for i, candle := range candles {
		open, openOK := parseFinitePositive(candle.Open)
		high, highOK := parseFinitePositive(candle.High)
		low, lowOK := parseFinitePositive(candle.Low)
		closePrice, closeOK := parseFinitePositive(candle.Close)
		if !openOK || !highOK || !lowOK || !closeOK || high < low || open < low || open > high || closePrice < low || closePrice > high {
			check.ZeroNegativeOHLC++
		}

		volume, volumeOK := parseFinite(candle.Volume)
		if !volumeOK || volume < 0 {
			check.InvalidVolume++
		}

		if seen[candle.StartTime] {
			check.DuplicateTimestamps++
		}
		seen[candle.StartTime] = true

		if i > 0 {
			gap := candle.StartTime - previousTime
			if gap < 0 {
				check.OutOfOrderTimestamps++
			}
			if gap > check.MaxObservedGapMS {
				check.MaxObservedGapMS = gap
			}
			if check.ExpectedIntervalMS > 0 && gap > check.ExpectedIntervalMS*3/2 {
				check.LargeTimeGaps++
				missing := int(math.Round(float64(gap)/float64(check.ExpectedIntervalMS))) - 1
				if missing > 0 {
					check.MissingCandles += missing
				}
			}
			if previousClose > 0 && closeOK {
				jumpPct := math.Abs(closePrice-previousClose) / previousClose * 100
				if jumpPct > check.MaxPriceJumpPct {
					check.MaxPriceJumpPct = jumpPct
				}
				if jumpPct >= 10 {
					check.SuspiciousPriceJumps++
				}
			}
		}

		previousTime = candle.StartTime
		if closeOK {
			previousClose = closePrice
		}
		_ = volume
	}

	check.Warnings = marketDataWarnings(check)
	check.QualityRisk = classifyMarketDataQualityRisk(check)
	return check
}

func classifyMarketDataQualityRisk(check MarketDataQualityCheck) DislocationSeverity {
	switch {
	case check.TotalCandles == 0:
		return SeverityHigh
	case check.OutOfOrderTimestamps > 0 || check.ZeroNegativeOHLC > 0 || check.InvalidVolume > 0:
		return SeverityHigh
	case check.SuspiciousPriceJumps > 2 || check.MissingCandles > maxInt(3, check.TotalCandles/100):
		return SeverityHigh
	case check.DuplicateTimestamps > 0 || check.LargeTimeGaps > 0 || check.SuspiciousPriceJumps > 0 || check.MissingCandles > 0:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func marketDataWarnings(check MarketDataQualityCheck) []string {
	var warnings []string
	if check.TotalCandles == 0 {
		warnings = append(warnings, "no_candles_available")
	}
	if check.MissingCandles > 0 {
		warnings = append(warnings, "missing_candles")
	}
	if check.DuplicateTimestamps > 0 {
		warnings = append(warnings, "duplicate_timestamps")
	}
	if check.OutOfOrderTimestamps > 0 {
		warnings = append(warnings, "out_of_order_timestamps")
	}
	if check.ZeroNegativeOHLC > 0 {
		warnings = append(warnings, "invalid_ohlc_values")
	}
	if check.InvalidVolume > 0 {
		warnings = append(warnings, "invalid_volume")
	}
	if check.LargeTimeGaps > 0 {
		warnings = append(warnings, "large_time_gaps")
	}
	if check.SuspiciousPriceJumps > 0 {
		warnings = append(warnings, "suspicious_price_jumps")
	}
	return warnings
}

func inferIntervalMS(candles []exchanges.Candle) int64 {
	if len(candles) < 2 {
		return 0
	}
	var gaps []int64
	for i := 1; i < len(candles); i++ {
		gap := candles[i].StartTime - candles[i-1].StartTime
		if gap > 0 {
			gaps = append(gaps, gap)
		}
	}
	if len(gaps) == 0 {
		return 0
	}
	sort.Slice(gaps, func(i, j int) bool { return gaps[i] < gaps[j] })
	return gaps[len(gaps)/2]
}

func parseFinitePositive(value string) (float64, bool) {
	parsed, ok := parseFinite(value)
	return parsed, ok && parsed > 0
}

func parseFinite(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}
	return parsed, true
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
