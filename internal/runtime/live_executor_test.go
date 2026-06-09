package runtime

import "testing"

func TestLiveGateRefusesWhenNotArmed(t *testing.T) {
	gate := LiveGate{VenueHealthy: true, AccountReady: true}
	if got := gate.Refusal(Candidate{Venue: "aster"}); got != "live_trading_not_enabled" {
		t.Fatalf("refusal=%q", got)
	}
	gate.LiveEnabled = true
	if got := gate.Refusal(Candidate{Venue: "aster"}); got != "execution_not_enabled" {
		t.Fatalf("refusal=%q", got)
	}
	gate.ExecutionEnabled = true
	gate.AccountReady = false
	if got := gate.Refusal(Candidate{Venue: "aster"}); got != "account_not_ready" {
		t.Fatalf("refusal=%q", got)
	}
}
