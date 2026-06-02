package risk

import "testing"

func TestCountViolationsByRule(t *testing.T) {
	counts := CountViolationsByRule([]RiskViolation{
		{Rule: RuleStopLoss},
		{Rule: RuleStopLoss},
		{Rule: RuleMaxNotional},
	})
	if counts[RuleStopLoss] != 2 || counts[RuleMaxNotional] != 1 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}

func TestHasBlockingViolation(t *testing.T) {
	if HasBlockingViolation([]RiskViolation{{Action: ActionAllow}}) {
		t.Fatal("allow-only violations should not block")
	}
	if !HasBlockingViolation([]RiskViolation{{Action: ActionForceExit}}) {
		t.Fatal("force exit should count as blocking")
	}
}
