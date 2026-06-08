package paper

import "testing"

func TestFundingChargeAndHazard(t *testing.T) {
	position := PaperPosition{
		Side:      "LONG",
		Quantity:  1,
		MarkPrice: 100,
	}
	charge := FundingCharge(position, 0.001)
	if charge >= 0 {
		t.Fatalf("expected long positive funding to be a debit, got %.4f", charge)
	}

	cfg := DefaultConfig()
	cfg.FundingHazardRate = 0.0005
	if !FundingHazardBlocked(0.001, "LONG", cfg) {
		t.Fatal("expected funding hazard block for long positive funding")
	}
	if FundingHazardBlocked(0.001, "SHORT", cfg) {
		t.Fatal("expected short not blocked on positive funding")
	}
}
