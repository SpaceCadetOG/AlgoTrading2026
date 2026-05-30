package backtest

import (
	"testing"
	"time"

	"AlgoTrading2026/exchanges"
)

func TestSimulatedClockRejectsBackwardMove(t *testing.T) {
	start := time.Unix(100, 0).UTC()
	clock := NewSimulatedClock(start)

	if err := clock.Advance(10 * time.Second); err != nil {
		t.Fatalf("advance clock: %v", err)
	}
	if got := clock.Current(); !got.Equal(start.Add(10 * time.Second)) {
		t.Fatalf("unexpected current time: %s", got)
	}
	if err := clock.Set(start); err == nil {
		t.Fatal("expected backward set to fail")
	}
	clock.Reset(start)
	if got := clock.Current(); !got.Equal(start) {
		t.Fatalf("reset did not move backward explicitly: %s", got)
	}
}

func TestSimulatedClockStepTo(t *testing.T) {
	start := time.Unix(100, 0).UTC()
	next := time.Unix(120, 0).UTC()
	clock := NewSimulatedClock(start)

	if err := clock.StepTo(next); err != nil {
		t.Fatalf("step clock: %v", err)
	}
	if !clock.Current().Equal(next) {
		t.Fatalf("expected %s, got %s", next, clock.Current())
	}
}

func TestCandleTimeIteratorOrdersAndStepsClock(t *testing.T) {
	candles := []exchanges.Candle{
		{StartTime: 3000},
		{StartTime: 1000},
		{StartTime: 2000},
	}
	clock := NewSimulatedClock(time.UnixMilli(1000).UTC())
	iterator := NewCandleTimeIterator(clock, candles)

	var seen []int64
	for {
		candle, ok, err := iterator.Next()
		if err != nil {
			t.Fatalf("next candle: %v", err)
		}
		if !ok {
			break
		}
		seen = append(seen, candle.StartTime)
		if got := clock.Current().UnixMilli(); got != candle.StartTime {
			t.Fatalf("clock did not step to candle: got %d want %d", got, candle.StartTime)
		}
	}

	if len(seen) != 3 || seen[0] != 1000 || seen[1] != 2000 || seen[2] != 3000 {
		t.Fatalf("unexpected candle order: %+v", seen)
	}
}

func TestCandleTimeIteratorEmpty(t *testing.T) {
	iterator := NewCandleTimeIterator(nil, nil)
	_, ok, err := iterator.Next()
	if err != nil {
		t.Fatalf("empty iterator returned error: %v", err)
	}
	if ok {
		t.Fatal("expected empty iterator to be done")
	}
}
