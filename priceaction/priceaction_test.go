package priceaction

import (
	"strconv"
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
)

func TestAggressionClassification(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 101, 99, 100.5, 10, 1),
		candle(100.5, 101, 100, 100.7, 10, 2),
		candle(100.7, 111, 100.5, 110, 40, 3),
	}
	rows := Aggression(candles, 2)
	if rows[2].Direction != AggressionBullish {
		t.Fatalf("direction=%s want bullish score=%.4f", rows[2].Direction, rows[2].Score)
	}
}

func TestSidewaysDetection(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 100.2, 99.8, 100.0, 10, 1),
		candle(100, 100.3, 99.9, 100.1, 10, 2),
		candle(100.1, 100.2, 99.8, 100.0, 10, 3),
		candle(100, 100.3, 99.9, 100.1, 10, 4),
	}
	rows := Sideways(candles, 4)
	if !rows[3].Detected {
		t.Fatalf("expected sideways row: %+v", rows[3])
	}
}

func TestInitiationAfterSideways(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 100.2, 99.8, 100.0, 10, 1),
		candle(100, 100.3, 99.9, 100.1, 10, 2),
		candle(100.1, 100.2, 99.8, 100.0, 10, 3),
		candle(100, 100.3, 99.9, 100.1, 10, 4),
		candle(100.1, 103, 100, 102.8, 50, 5),
	}
	aggression := Aggression(candles, 4)
	sideways := Sideways(candles, 4)
	rows := Initiations(candles, aggression, sideways)
	if !rows[4].Detected || rows[4].Direction != InitiationUp {
		t.Fatalf("expected up initiation: %+v aggression=%+v sideways=%+v", rows[4], aggression[4], sideways[3])
	}
}

func TestRejections(t *testing.T) {
	bullish := Rejections([]exchanges.Candle{candle(100, 101, 90, 100, 10, 1)})
	if !bullish[0].Detected || bullish[0].Direction != RejectionBullish {
		t.Fatalf("expected bullish rejection: %+v", bullish[0])
	}
	bearish := Rejections([]exchanges.Candle{candle(100, 110, 99, 100, 10, 1)})
	if !bearish[0].Detected || bearish[0].Direction != RejectionBearish {
		t.Fatalf("expected bearish rejection: %+v", bearish[0])
	}
}

func TestSupportResistanceFlipDetection(t *testing.T) {
	candles := []exchanges.Candle{
		candle(101, 102, 99, 101, 10, 1),
		candle(101, 101.5, 98, 99, 10, 2),
		candle(99, 100.1, 98.5, 99.5, 10, 3),
	}
	rejections := []RejectionFeature{
		{Detected: true, Direction: RejectionBullish, Level: 100},
		{},
		{},
	}
	rows := SupportResistanceFlips(candles, rejections)
	if !rows[2].Detected || rows[2].Direction != FlipSupportToResistance {
		t.Fatalf("expected support to resistance flip: %+v", rows[2])
	}
}

func TestDailyOpenAssignment(t *testing.T) {
	day1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	day2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	rows := Opens([]exchanges.Candle{
		candle(100, 101, 99, 100, 10, day1),
		candle(101, 102, 100, 101, 10, day1+900000),
		candle(200, 201, 199, 200, 10, day2),
	})
	if rows[0].DailyOpenLevel != 100 || rows[1].DailyOpenLevel != 100 || rows[2].DailyOpenLevel != 200 {
		t.Fatalf("bad opens: %+v", rows)
	}
}

func TestPreviousHighLowAndFailedAuction(t *testing.T) {
	day1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	day2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	candles := []exchanges.Candle{
		candle(100, 110, 95, 105, 10, day1),
		candle(105, 108, 100, 106, 10, day1+900000),
		candle(106, 112, 104, 109, 10, day2),
	}
	rejections := Rejections(candles)
	aggression := Aggression(candles, 2)
	levels := HighsLows(candles, rejections, aggression)
	if levels[2].PreviousDailyHigh != 110 || levels[2].PreviousDailyLow != 95 {
		t.Fatalf("bad previous daily levels: %+v", levels[2])
	}
	failed := FailedAuctions(candles, levels)
	if !failed[2].Detected || failed[2].Direction != FailedAuctionUp {
		t.Fatalf("expected failed auction up: %+v", failed[2])
	}
}

func candle(open, high, low, close, volume float64, ts int64) exchanges.Candle {
	return exchanges.Candle{
		Open:      f(open),
		High:      f(high),
		Low:       f(low),
		Close:     f(close),
		Volume:    f(volume),
		StartTime: ts,
		EndTime:   ts + 900000,
	}
}

func f(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
