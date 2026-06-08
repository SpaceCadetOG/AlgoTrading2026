package backtest

import (
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestEvaluateBarExitDirectionAwareAndStopFirstCollision(t *testing.T) {
	candle := exchanges.Candle{High: "106", Low: "94"}

	longExit := EvaluateBarExit(SideLong, candle, 95, 105, SameBarStopFirst)
	if !longExit.Hit || longExit.Reason != "stop" || longExit.Price != 95 {
		t.Fatalf("expected long stop-first collision, got %+v", longExit)
	}
	shortExit := EvaluateBarExit(SideShort, candle, 105, 95, SameBarStopFirst)
	if !shortExit.Hit || shortExit.Reason != "stop" || shortExit.Price != 105 {
		t.Fatalf("expected short stop-first collision, got %+v", shortExit)
	}
	targetFirst := EvaluateBarExit(SideShort, candle, 105, 95, SameBarTargetFirst)
	if !targetFirst.Hit || targetFirst.Reason != "target" || targetFirst.Price != 95 {
		t.Fatalf("expected target-first override, got %+v", targetFirst)
	}
}
