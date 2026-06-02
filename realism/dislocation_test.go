package realism

import "testing"

func TestSimulationDislocationClassifiesOptimisticHigh(t *testing.T) {
	d := NewSimulationDislocation("test", "BTC", 25, -0.26, 2252.12)

	if d.BiasDirection != BiasOptimistic {
		t.Fatalf("expected optimistic bias, got %s", d.BiasDirection)
	}
	if d.Severity != SeverityHigh {
		t.Fatalf("expected high severity, got %s", d.Severity)
	}
	if d.PnLDifference <= 0 {
		t.Fatalf("expected positive pnl difference, got %.2f", d.PnLDifference)
	}
}

func TestRealismAssumptionsMissingAndStrengths(t *testing.T) {
	assumptions := DefaultRealismAssumptions()

	if assumptions.MissingCount() == 0 {
		t.Fatal("expected missing realism assumptions")
	}
	if len(assumptions.Strengths()) != 2 {
		t.Fatalf("expected fee/slippage strengths, got %+v", assumptions.Strengths())
	}
}

func TestGradeRealism(t *testing.T) {
	if grade := GradeRealism(DefaultRealismAssumptions(), SeverityHigh); grade != GradeF {
		t.Fatalf("expected F for current realism assumptions and high severity, got %s", grade)
	}

	full := RealismAssumptions{
		SlippageModeled:                   true,
		FeesModeled:                       true,
		LatencyModeled:                    true,
		LatencyVarianceModeled:            true,
		PlaceInLineModeled:                true,
		MarketImpactModeled:               true,
		MarketDataAccuracyChecked:         true,
		HistoricalLiveFormatParityChecked: true,
		OperationalInterventionModeled:    true,
		LiveAnalyticsAvailable:            true,
		ProfitDecayTracked:                true,
	}
	if grade := GradeRealism(full, SeverityLow); grade != GradeA {
		t.Fatalf("expected A for complete low-severity realism, got %s", grade)
	}
}
