package priceaction

import (
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
)

func TestDailyHighRetestLongContext(t *testing.T) {
	day1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	day2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	candles := []exchanges.Candle{
		candle(100, 110, 90, 105, 10, day1),
		candle(105, 108, 95, 106, 10, day1+900000),
		candle(111, 114, 110.5, 112, 10, day2),
		candle(112, 113, 111, 112.5, 10, day2+900000),
		candle(112.5, 113, 109.9, 111, 10, day2+1800000),
	}
	rows := HighLowRetestSetups(candles, 2)
	assertPhase3Setup(t, rows, "daily_high_retest", Phase3LongContext)
}

func TestDailyLowRetestShortContext(t *testing.T) {
	day1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	day2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	candles := []exchanges.Candle{
		candle(100, 110, 90, 95, 10, day1),
		candle(95, 105, 92, 96, 10, day1+900000),
		candle(89, 89.5, 86, 88, 10, day2),
		candle(88, 89, 86, 87.5, 10, day2+900000),
		candle(87.5, 90.1, 87, 89, 10, day2+1800000),
	}
	rows := HighLowRetestSetups(candles, 2)
	assertPhase3Setup(t, rows, "daily_low_retest", Phase3ShortContext)
}

func TestWeeklyHighLowRetests(t *testing.T) {
	week1 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	week2 := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC).UnixMilli()
	highCandles := []exchanges.Candle{
		candle(100, 120, 80, 110, 10, week1),
		candle(121, 124, 120.5, 122, 10, week2),
		candle(122, 123, 121, 122.5, 10, week2+900000),
		candle(122.5, 123, 119.9, 121, 10, week2+1800000),
	}
	assertPhase3Setup(t, HighLowRetestSetups(highCandles, 2), "weekly_high_retest", Phase3LongContext)

	lowCandles := []exchanges.Candle{
		candle(100, 120, 80, 90, 10, week1),
		candle(79, 79.5, 76, 78, 10, week2),
		candle(78, 79, 76, 77.5, 10, week2+900000),
		candle(77.5, 80.1, 77, 79, 10, week2+1800000),
	}
	assertPhase3Setup(t, HighLowRetestSetups(lowCandles, 2), "weekly_low_retest", Phase3ShortContext)
}

func TestStrongWeakHighLowClassifications(t *testing.T) {
	candles := []exchanges.Candle{
		candle(100, 130, 95, 105, 10, 1),
		candle(105, 106, 60, 70, 60, 2),
		candle(70, 72, 68, 71, 10, 3),
		candle(71, 73, 69, 72, 10, 4),
	}
	rows := StrongWeakHighLowSetups(candles)
	assertPhase3Setup(t, rows, "strong_high", Phase3ShortContext)

	candles = []exchanges.Candle{
		candle(100, 105, 70, 95, 10, 1),
		candle(95, 140, 94, 130, 60, 2),
		candle(130, 131, 129, 130, 10, 3),
	}
	rows = StrongWeakHighLowSetups(candles)
	assertPhase3Setup(t, rows, "strong_low", Phase3LongContext)

	day1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	day2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	weak := []exchanges.Candle{
		candle(100, 110, 90, 100, 10, day1),
		candle(100, 110.2, 99, 109.9, 10, day2),
		candle(109.9, 111, 89.9, 90.1, 10, day2+900000),
	}
	rows = StrongWeakHighLowSetups(weak)
	assertPhase3Setup(t, rows, "weak_high", Phase3BreakoutWatch)
	assertPhase3Setup(t, rows, "weak_low", Phase3BreakoutWatch)
}

func TestFailedAuctionSetups(t *testing.T) {
	day1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	day2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli()
	highFail := []exchanges.Candle{
		candle(100, 110, 90, 105, 10, day1),
		candle(105, 112, 104, 109, 10, day2),
	}
	assertPhase3Setup(t, FailedAuctionSetups(highFail), "failed_high_auction", Phase3ShortContext)

	lowFail := []exchanges.Candle{
		candle(100, 110, 90, 95, 10, day1),
		candle(95, 96, 88, 91, 10, day2),
	}
	assertPhase3Setup(t, FailedAuctionSetups(lowFail), "failed_low_auction", Phase3LongContext)
}

func assertPhase3Setup(t *testing.T, rows []Phase3Setup, strategy string, direction Phase3Direction) {
	t.Helper()
	for _, row := range rows {
		if row.Strategy == strategy && row.Direction == direction {
			return
		}
	}
	t.Fatalf("missing setup strategy=%s direction=%s rows=%+v", strategy, direction, rows)
}
