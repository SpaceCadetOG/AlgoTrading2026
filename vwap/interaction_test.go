package vwap

import (
	"strconv"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestTouchedAndCrossedVWAP(t *testing.T) {
	candles := []exchanges.Candle{
		testInteractionCandle(0, 100, 101, 99, 99),
		testInteractionCandle(1, 100, 102, 99, 101),
	}
	values := []float64{100, 100}
	if !TouchedVWAP(candles[1].HighFloat(), candles[1].LowFloat(), values[1]) {
		t.Fatal("expected touch")
	}
	if !CrossedVWAP(candles, values, 1) {
		t.Fatal("expected cross")
	}
}

func TestReactionClassification(t *testing.T) {
	bounceCandles := []exchanges.Candle{
		testInteractionCandle(0, 100, 101, 99, 101),
		testInteractionCandle(1, 100, 102, 99, 101),
	}
	if got := ClassifyReaction(bounceCandles, []float64{100, 100}, 1); got != ReactionBounce {
		t.Fatalf("bounce got %s", got)
	}

	breakCandles := []exchanges.Candle{
		testInteractionCandle(0, 100, 101, 99, 101),
		testInteractionCandle(1, 100, 101, 99, 99),
	}
	if got := ClassifyReaction(breakCandles, []float64{100, 100}, 1); got != ReactionBreak {
		t.Fatalf("break got %s", got)
	}

	chopCandles := []exchanges.Candle{
		testInteractionCandle(0, 100, 101, 99, 101),
		testInteractionCandle(1, 100, 101, 99, 100.01),
	}
	if got := ClassifyReaction(chopCandles, []float64{100, 100}, 1); got != ReactionChop {
		t.Fatalf("chop got %s", got)
	}
}

func TestMagnetStudy(t *testing.T) {
	candles := []exchanges.Candle{
		testInteractionCandle(0, 100, 101, 99, 102),
		testInteractionCandle(1, 100, 100.5, 99.5, 100),
		testInteractionCandle(2, 100, 101, 99, 98),
	}
	stats := MagnetStudy(candles, []float64{100, 100, 100}, []float64{1}, 2)
	if len(stats) != 1 {
		t.Fatalf("stats len=%d", len(stats))
	}
	if stats[0].Events != 2 || stats[0].Returns != 1 {
		t.Fatalf("unexpected stats: %+v", stats[0])
	}
	if stats[0].ReturnProbability != 0.5 {
		t.Fatalf("return probability %.4f", stats[0].ReturnProbability)
	}
}

func testInteractionCandle(t int64, open float64, high float64, low float64, close float64) exchanges.Candle {
	return exchanges.Candle{
		Open:      floatString(open),
		High:      floatString(high),
		Low:       floatString(low),
		Close:     floatString(close),
		Volume:    "1",
		StartTime: t,
		EndTime:   t + 1,
	}
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
