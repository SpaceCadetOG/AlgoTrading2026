package backtest

import (
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestInSampleOutOfSampleSplitDefaultsAndSorts(t *testing.T) {
	candles := []exchanges.Candle{
		{StartTime: 300},
		{StartTime: 100},
		{StartTime: 500},
		{StartTime: 200},
		{StartTime: 400},
	}

	inSample, outOfSample := InSampleOutOfSampleSplit(candles, 0)
	if len(inSample) != 4 || len(outOfSample) != 1 {
		t.Fatalf("expected 4/1 split, got %d/%d", len(inSample), len(outOfSample))
	}
	if inSample[0].StartTime != 100 || inSample[3].StartTime != 400 || outOfSample[0].StartTime != 500 {
		t.Fatalf("split did not preserve timestamp order: in=%+v out=%+v", inSample, outOfSample)
	}
}

func TestInSampleOutOfSampleSplitCustomRatio(t *testing.T) {
	candles := []exchanges.Candle{
		{StartTime: 1},
		{StartTime: 2},
		{StartTime: 3},
		{StartTime: 4},
	}

	inSample, outOfSample := InSampleOutOfSampleSplit(candles, 0.5)
	if len(inSample) != 2 || len(outOfSample) != 2 {
		t.Fatalf("expected 2/2 split, got %d/%d", len(inSample), len(outOfSample))
	}
}

func TestInSampleOutOfSampleSplitSmallInputs(t *testing.T) {
	inSample, outOfSample := InSampleOutOfSampleSplit(nil, 0.8)
	if len(inSample) != 0 || len(outOfSample) != 0 {
		t.Fatalf("expected empty split, got %d/%d", len(inSample), len(outOfSample))
	}

	inSample, outOfSample = InSampleOutOfSampleSplit([]exchanges.Candle{{StartTime: 1}}, 0.8)
	if len(inSample) != 1 || len(outOfSample) != 0 {
		t.Fatalf("expected one candle in sample, got %d/%d", len(inSample), len(outOfSample))
	}
}
