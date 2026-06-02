package realism

import (
	"testing"
	"time"
)

func TestLatencyModelRisk(t *testing.T) {
	low := NewLatencyModel(10*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, 5*time.Millisecond)
	if low.Risk != SeverityLow {
		t.Fatalf("expected low latency risk, got %s", low.Risk)
	}

	high := DefaultLatencyModel()
	if high.Risk != SeverityHigh {
		t.Fatalf("expected missing/default latency to be high risk, got %s", high.Risk)
	}
}

func TestPlaceInLineEstimateRisk(t *testing.T) {
	low := NewPlaceInLineEstimate(1, 10, 0)
	if low.FillProbability != 1 || low.Risk != SeverityLow {
		t.Fatalf("expected low place-in-line risk, got %+v", low)
	}

	high := NewPlaceInLineEstimate(1, 0, 0)
	if high.Risk != SeverityHigh {
		t.Fatalf("expected no liquidity to be high risk, got %s", high.Risk)
	}
}

func TestMarketImpactModelRisk(t *testing.T) {
	low := NewMarketImpactModel(1, 1000)
	if low.Risk != SeverityLow {
		t.Fatalf("expected low market impact risk, got %+v", low)
	}

	high := NewMarketImpactModel(25, 100)
	if high.Risk != SeverityHigh {
		t.Fatalf("expected high participation to be high risk, got %+v", high)
	}
}

func TestFillAssumptionRisk(t *testing.T) {
	current := DefaultFillAssumptionModel()
	if current.Risk != SeverityHigh {
		t.Fatalf("expected current deterministic crossed-book assumptions to be high risk, got %s", current.Risk)
	}

	low := NewFillAssumptionModel(0.75, true, true, false, false, false)
	if low.Risk != SeverityLow {
		t.Fatalf("expected realistic fill assumptions to be low risk, got %s", low.Risk)
	}
}

func TestAggregateRealismModel(t *testing.T) {
	current := DefaultAggregateRealismModel()
	if current.OverallRisk != SeverityHigh {
		t.Fatalf("expected current aggregate model to be high risk, got %s", current.OverallRisk)
	}
	if len(current.Warnings) == 0 {
		t.Fatal("expected current aggregate model warnings")
	}

	low := NewAggregateRealismModel(
		NewLatencyModel(10*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, 10*time.Millisecond, 5*time.Millisecond),
		NewPlaceInLineEstimate(1, 10, 0),
		NewMarketImpactModel(1, 1000),
		NewFillAssumptionModel(0.75, true, true, false, false, false),
	)
	if low.OverallRisk != SeverityLow {
		t.Fatalf("expected low aggregate risk, got %s", low.OverallRisk)
	}
}
