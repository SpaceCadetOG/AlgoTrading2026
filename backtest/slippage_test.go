package backtest

import "testing"

func TestBuySellTakerSlippageDirection(t *testing.T) {
	model := DefaultSlippageModel().WithVenue("test", VenueSlippageConfig{TakerBps: 10})

	buy := model.Apply("test", "BUY", 100, 1000, LiquidityTaker)
	if buy <= 100 {
		t.Fatalf("buy taker price = %f, want above 100", buy)
	}

	sell := model.Apply("test", "SELL", 100, 1000, LiquidityTaker)
	if sell >= 100 {
		t.Fatalf("sell taker price = %f, want below 100", sell)
	}
}

func TestMakerNoSlippageDefault(t *testing.T) {
	model := DefaultSlippageModel()

	got := model.Apply("aster", "BUY", 100, 1000, LiquidityMaker)
	if got != 100 {
		t.Fatalf("maker price = %f, want unchanged 100", got)
	}
}
