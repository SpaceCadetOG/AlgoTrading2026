package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/priceaction"
)

type PriceActionFeatureRow struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64

	RangePct     float64
	BodyPct      float64
	UpperWickPct float64
	LowerWickPct float64

	AggressionDirection priceaction.AggressionDirection
	AggressionScore     float64

	SidewaysDetected bool
	SidewaysHigh     float64
	SidewaysLow      float64
	SidewaysDuration int

	InitiationDetected   bool
	InitiationDirection  priceaction.InitiationDirection
	InitiationStartPrice float64

	RejectionDetected  bool
	RejectionDirection priceaction.RejectionDirection
	RejectionLevel     float64

	SupportResistanceFlipDetected bool
	FlipLevel                     float64
	FlipDirection                 priceaction.FlipDirection

	SessionOpenLevel float64
	DailyOpenLevel   float64

	PreviousDailyHigh  float64
	PreviousDailyLow   float64
	PreviousWeeklyHigh float64
	PreviousWeeklyLow  float64

	NearPreviousHigh bool
	NearPreviousLow  bool

	StrongHigh bool
	StrongLow  bool
	WeakHigh   bool
	WeakLow    bool

	FailedAuctionDetected  bool
	FailedAuctionDirection priceaction.FailedAuctionDirection
}

func BuildPriceActionFeatureRows(candles []exchanges.Candle) []PriceActionFeatureRow {
	aggression := priceaction.Aggression(candles, 20)
	sideways := priceaction.Sideways(candles, 8)
	initiations := priceaction.Initiations(candles, aggression, sideways)
	rejections := priceaction.Rejections(candles)
	flips := priceaction.SupportResistanceFlips(candles, rejections)
	opens := priceaction.Opens(candles)
	highsLows := priceaction.HighsLows(candles, rejections, aggression)
	failedAuctions := priceaction.FailedAuctions(candles, highsLows)

	rows := make([]PriceActionFeatureRow, 0, len(candles))
	for i, candle := range candles {
		metrics := priceaction.Metrics(candle)
		row := PriceActionFeatureRow{
			Timestamp:    candle.StartTime,
			Open:         candle.OpenFloat(),
			High:         candle.HighFloat(),
			Low:          candle.LowFloat(),
			Close:        candle.CloseFloat(),
			Volume:       candle.VolumeFloat(),
			RangePct:     metrics.RangePct,
			BodyPct:      metrics.BodyPct,
			UpperWickPct: metrics.UpperWickPct,
			LowerWickPct: metrics.LowerWickPct,
		}
		if i < len(aggression) {
			row.AggressionDirection = aggression[i].Direction
			row.AggressionScore = aggression[i].Score
		}
		if i < len(sideways) {
			row.SidewaysDetected = sideways[i].Detected
			row.SidewaysHigh = sideways[i].High
			row.SidewaysLow = sideways[i].Low
			row.SidewaysDuration = sideways[i].Duration
		}
		if i < len(initiations) {
			row.InitiationDetected = initiations[i].Detected
			row.InitiationDirection = initiations[i].Direction
			row.InitiationStartPrice = initiations[i].StartPrice
		}
		if i < len(rejections) {
			row.RejectionDetected = rejections[i].Detected
			row.RejectionDirection = rejections[i].Direction
			row.RejectionLevel = rejections[i].Level
		}
		if i < len(flips) {
			row.SupportResistanceFlipDetected = flips[i].Detected
			row.FlipLevel = flips[i].Level
			row.FlipDirection = flips[i].Direction
		}
		if i < len(opens) {
			row.SessionOpenLevel = opens[i].SessionOpenLevel
			row.DailyOpenLevel = opens[i].DailyOpenLevel
		}
		if i < len(highsLows) {
			row.PreviousDailyHigh = highsLows[i].PreviousDailyHigh
			row.PreviousDailyLow = highsLows[i].PreviousDailyLow
			row.PreviousWeeklyHigh = highsLows[i].PreviousWeeklyHigh
			row.PreviousWeeklyLow = highsLows[i].PreviousWeeklyLow
			row.NearPreviousHigh = highsLows[i].NearPreviousHigh
			row.NearPreviousLow = highsLows[i].NearPreviousLow
			row.StrongHigh = highsLows[i].StrongHigh
			row.StrongLow = highsLows[i].StrongLow
			row.WeakHigh = highsLows[i].WeakHigh
			row.WeakLow = highsLows[i].WeakLow
		}
		if i < len(failedAuctions) {
			row.FailedAuctionDetected = failedAuctions[i].Detected
			row.FailedAuctionDirection = failedAuctions[i].Direction
		}
		rows = append(rows, row)
	}
	return rows
}

func WritePriceActionFeaturesCSV(path string, rows []PriceActionFeatureRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{
		"timestamp", "open", "high", "low", "close", "volume",
		"range_pct", "body_pct", "upper_wick_pct", "lower_wick_pct",
		"aggression_direction", "aggression_score",
		"sideways_detected", "sideways_high", "sideways_low", "sideways_duration",
		"initiation_detected", "initiation_direction", "initiation_start_price",
		"rejection_detected", "rejection_direction", "rejection_level",
		"support_resistance_flip_detected", "flip_level", "flip_direction",
		"session_open_level", "daily_open_level",
		"previous_daily_high", "previous_daily_low", "previous_weekly_high", "previous_weekly_low",
		"near_previous_high", "near_previous_low",
		"strong_high", "strong_low", "weak_high", "weak_low",
		"failed_auction_detected", "failed_auction_direction",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			floatToString(row.Open),
			floatToString(row.High),
			floatToString(row.Low),
			floatToString(row.Close),
			floatToString(row.Volume),
			floatToString(row.RangePct),
			floatToString(row.BodyPct),
			floatToString(row.UpperWickPct),
			floatToString(row.LowerWickPct),
			string(row.AggressionDirection),
			floatToString(row.AggressionScore),
			strconv.FormatBool(row.SidewaysDetected),
			floatToString(row.SidewaysHigh),
			floatToString(row.SidewaysLow),
			strconv.Itoa(row.SidewaysDuration),
			strconv.FormatBool(row.InitiationDetected),
			string(row.InitiationDirection),
			floatToString(row.InitiationStartPrice),
			strconv.FormatBool(row.RejectionDetected),
			string(row.RejectionDirection),
			floatToString(row.RejectionLevel),
			strconv.FormatBool(row.SupportResistanceFlipDetected),
			floatToString(row.FlipLevel),
			string(row.FlipDirection),
			floatToString(row.SessionOpenLevel),
			floatToString(row.DailyOpenLevel),
			floatToString(row.PreviousDailyHigh),
			floatToString(row.PreviousDailyLow),
			floatToString(row.PreviousWeeklyHigh),
			floatToString(row.PreviousWeeklyLow),
			strconv.FormatBool(row.NearPreviousHigh),
			strconv.FormatBool(row.NearPreviousLow),
			strconv.FormatBool(row.StrongHigh),
			strconv.FormatBool(row.StrongLow),
			strconv.FormatBool(row.WeakHigh),
			strconv.FormatBool(row.WeakLow),
			strconv.FormatBool(row.FailedAuctionDetected),
			string(row.FailedAuctionDirection),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
