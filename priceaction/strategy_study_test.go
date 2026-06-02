package priceaction

import (
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
)

func TestSRFlipLongContext(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 110, 99, 100, 10, 1),
		candle(100, 112, 100, 111, 10, 2),
		candle(111, 112, 109.95, 110.5, 10, 3),
	}
	setups := SRFlipSetups(candles)
	if len(setups) == 0 || setups[0].Direction != StudyLong {
		t.Fatalf("expected long sr flip: %+v", setups)
	}
}

func TestSRFlipShortContext(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 101, 90, 100, 10, 1),
		candle(100, 100, 88, 89, 10, 2),
		candle(89, 90.05, 88, 89.5, 10, 3),
	}
	setups := SRFlipSetups(candles)
	if len(setups) == 0 || setups[0].Direction != StudyShort {
		t.Fatalf("expected short sr flip: %+v", setups)
	}
}

func TestOpenDriveSetup(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 100.2, 99.8, 100.0, 10, 1),
		candle(100, 100.3, 99.9, 100.1, 10, 2),
		candle(100.1, 100.2, 99.8, 100.0, 10, 3),
		candle(100, 100.3, 99.9, 100.1, 10, 4),
		candle(100.1, 100.2, 99.8, 100.0, 10, 5),
		candle(100, 100.3, 99.9, 100.1, 10, 6),
		candle(100.1, 100.2, 99.8, 100.0, 10, 7),
		candle(100, 100.3, 99.9, 100.1, 10, 8),
		candle(100.1, 103, 100, 102.8, 50, 9),
		candle(102.8, 103, 100.05, 101, 15, 10),
	}
	setups := OpenDriveSetups(candles)
	if len(setups) == 0 || setups[0].Direction != StudyLong {
		t.Fatalf("expected open-drive long setup: %+v", setups)
	}
}

func TestABCDBullishProjection(t *testing.T) {
	candles := []exchanges.Candle{
		candle(108, 110, 107, 109, 10, 1),
		candle(105, 106, 99, 100, 10, 2),
		candle(103, 105, 102, 104, 10, 3),
		candle(96, 97, 94, 95, 10, 4),
	}
	setups := ABCDSetups(candles)
	if len(setups) == 0 || setups[0].Direction != StudyLong {
		t.Fatalf("expected bullish abcd: %+v", setups)
	}
}

func TestABCDBearishProjection(t *testing.T) {
	candles := []exchanges.Candle{
		candle(92, 93, 90, 91, 10, 1),
		candle(95, 101, 94, 100, 10, 2),
		candle(97, 98, 95.5, 97, 10, 3),
		candle(104, 106.5, 103, 105, 10, 4),
	}
	setups := ABCDSetups(candles)
	if len(setups) == 0 || setups[0].Direction != StudyShort {
		t.Fatalf("expected bearish abcd: %+v", setups)
	}
}

func TestSessionAndDailyOpenRetests(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	candles := []exchanges.Candle{
		candle(100, 101, 99, 100, 10, start),
		candle(100, 103, 100, 102, 10, start+900000),
		candle(102, 103, 99.95, 100.5, 10, start+1800000),
	}
	session := SessionOpenSetups(candles)
	daily := DailyOpenSetups(candles)
	if len(session) == 0 || session[0].Direction != StudyLong {
		t.Fatalf("expected session open retest: %+v", session)
	}
	if len(daily) == 0 || daily[0].Direction != StudyLong {
		t.Fatalf("expected daily open retest: %+v", daily)
	}
}
