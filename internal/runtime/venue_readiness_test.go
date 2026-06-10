package runtime

import "testing"

func TestBuildVenueGateStatusExplainsBlockedVenue(t *testing.T) {
	rows := BuildVenueGateStatus(AccountHealthSnapshot{Venues: []VenueHealth{{
		Venue:               "lighter",
		Healthy:             false,
		AccountReady:        true,
		StreamReady:         false,
		ReconciliationReady: true,
		ProtectionReady:     true,
		Stream:              StreamStatus{Venue: "lighter", PollingFallback: true},
		Reasons:             []string{"stream_not_ready"},
	}}})
	if len(rows) != 1 || rows[0].LiveAllowed || rows[0].Reason == "" || !rows[0].PollingFallbackActive {
		t.Fatalf("unexpected gate status: %+v", rows)
	}
}

func TestBuildVenueReadinessAuditClassifiesLighterGuardedLive(t *testing.T) {
	health := VenueHealth{
		Venue:                 "lighter",
		Healthy:               true,
		AccountReady:          true,
		PrivateEndpointsReady: true,
		PositionSyncReady:     true,
		FillHistoryReady:      true,
		ProtectionReady:       true,
		ReconciliationReady:   true,
		StreamReady:           true,
		Protection:            ProtectionForVenue("lighter"),
		Stream:                StreamStatus{Venue: "lighter", Supported: true, Connected: true, PollingFallback: true},
	}
	audit := BuildVenueReadinessAudit(health)
	if audit.Tier != "guarded_live" || !audit.PositionSyncTrusted || !audit.PollingHealing {
		t.Fatalf("unexpected readiness audit: %+v", audit)
	}
}
