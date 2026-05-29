package backtest

import "testing"

func TestAsterFeeCalculationMakerTaker(t *testing.T) {
	model := DefaultFeeModel()

	if got := model.CalculateFee("aster", 1000, LiquidityMaker); got != 0 {
		t.Fatalf("aster maker fee = %f, want 0", got)
	}
	if got := model.CalculateFee("aster", 1000, LiquidityTaker); got != 0.4 {
		t.Fatalf("aster taker fee = %f, want 0.4", got)
	}
}

func TestHyperliquidFeeOverride(t *testing.T) {
	model := DefaultFeeModel().WithVenue("hyperliquid", VenueFeeConfig{
		MakerRate: 0.0001,
		TakerRate: 0.0005,
	})

	if got := model.CalculateFee("hyperliquid", 1000, LiquidityMaker); got != 0.1 {
		t.Fatalf("hyperliquid maker fee = %f, want 0.1", got)
	}
	if got := model.CalculateFee("hyperliquid", 1000, LiquidityTaker); got != 0.5 {
		t.Fatalf("hyperliquid taker fee = %f, want 0.5", got)
	}
}

func TestLighterZeroFeeDefault(t *testing.T) {
	model := DefaultFeeModel()

	if got := model.CalculateFee("lighter", 1000, LiquidityMaker); got != 0 {
		t.Fatalf("lighter maker fee = %f, want 0", got)
	}
	if got := model.CalculateFee("lighter", 1000, LiquidityTaker); got != 0 {
		t.Fatalf("lighter taker fee = %f, want 0", got)
	}
}
