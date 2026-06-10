package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type VenueGateStatus struct {
	Venue                 string   `json:"venue"`
	LiveAllowed           bool     `json:"liveAllowed"`
	Reason                string   `json:"reason,omitempty"`
	Healthy               bool     `json:"healthy"`
	AccountReady          bool     `json:"accountReady"`
	StreamReady           bool     `json:"streamReady"`
	ReconciliationReady   bool     `json:"reconciliationReady"`
	ProtectionReady       bool     `json:"protectionReady"`
	PollingFallbackActive bool     `json:"pollingFallbackActive"`
	Reasons               []string `json:"reasons,omitempty"`
}

type VenueReadinessAudit struct {
	Timestamp             int64       `json:"timestamp"`
	TimeUTC               string      `json:"timeUtc"`
	Venue                 string      `json:"venue"`
	Tier                  string      `json:"tier"`
	SupportedCapabilities []string    `json:"supportedCapabilities"`
	MissingCapabilities   []string    `json:"missingCapabilities"`
	NativeProtections     []string    `json:"nativeProtections"`
	RuntimeManaged        []string    `json:"runtimeManaged"`
	StreamFresh           bool        `json:"streamFresh"`
	PollingHealing        bool        `json:"pollingHealing"`
	PositionSyncTrusted   bool        `json:"positionSyncTrusted"`
	Reasons               []string    `json:"reasons"`
	RawHealth             VenueHealth `json:"rawHealth"`
}

func BuildVenueGateStatus(snapshot AccountHealthSnapshot) []VenueGateStatus {
	out := make([]VenueGateStatus, 0, len(snapshot.Venues))
	for _, health := range snapshot.Venues {
		reason := ""
		if !health.Healthy {
			reason = health.RefusalReason()
		}
		out = append(out, VenueGateStatus{
			Venue:                 health.Venue,
			LiveAllowed:           health.Healthy,
			Reason:                reason,
			Healthy:               health.Healthy,
			AccountReady:          health.AccountReady,
			StreamReady:           health.StreamReady,
			ReconciliationReady:   health.ReconciliationReady,
			ProtectionReady:       health.ProtectionReady,
			PollingFallbackActive: health.Stream.PollingFallback,
			Reasons:               append([]string(nil), health.Reasons...),
		})
	}
	return out
}

func WriteVenueGateStatus(root string, rows []VenueGateStatus) error {
	return writeReadinessJSON(filepath.Join(root, "venue_gate_status.json"), rows)
}

func BuildVenueReadinessAudit(health VenueHealth) VenueReadinessAudit {
	now := time.Now().UTC()
	audit := VenueReadinessAudit{
		Timestamp:           now.UnixMilli(),
		TimeUTC:             now.Format(time.RFC3339),
		Venue:               health.Venue,
		StreamFresh:         health.StreamReady,
		PollingHealing:      health.Stream.PollingFallback,
		PositionSyncTrusted: health.PositionSyncReady,
		Reasons:             append([]string(nil), health.Reasons...),
		RawHealth:           health,
	}
	if health.Protection.NativeStop {
		audit.NativeProtections = append(audit.NativeProtections, "native_stop")
	}
	if health.Protection.NativeTakeProfit {
		audit.NativeProtections = append(audit.NativeProtections, "native_take_profit")
	}
	if health.Protection.ReduceOnly {
		audit.SupportedCapabilities = append(audit.SupportedCapabilities, "reduce_only")
	}
	if health.Protection.RuntimeManagedTrailing {
		audit.RuntimeManaged = append(audit.RuntimeManaged, "trailing")
	}
	if health.ReconciliationReady {
		audit.SupportedCapabilities = append(audit.SupportedCapabilities, "reconciliation")
	} else {
		audit.MissingCapabilities = append(audit.MissingCapabilities, "reconciliation")
	}
	if health.Stream.Supported {
		audit.SupportedCapabilities = append(audit.SupportedCapabilities, "account_stream_shape")
	}
	if !health.StreamReady && health.Stream.RequiresFreshStream {
		audit.MissingCapabilities = append(audit.MissingCapabilities, "fresh_account_stream")
	}
	audit.Tier = readinessTier(health)
	return audit
}

func WriteVenueReadinessAudit(root string, audit VenueReadinessAudit) error {
	name := strings.ToLower(strings.TrimSpace(audit.Venue)) + "_readiness_audit.json"
	if audit.Venue == "lighter" {
		name = "lighter_readiness_audit.json"
	}
	return writeReadinessJSON(filepath.Join(root, name), audit)
}

func readinessTier(health VenueHealth) string {
	if !health.AccountReady || !health.PrivateEndpointsReady {
		return "not_ready"
	}
	if !health.ReconciliationReady || !health.ProtectionReady {
		return "degraded"
	}
	if health.Stream.RequiresFreshStream && !health.StreamReady {
		return "degraded"
	}
	if health.Stream.Supported && health.Stream.PollingFallback {
		return "guarded_live"
	}
	if health.Healthy {
		return "guarded_live"
	}
	return "degraded"
}

func writeReadinessJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}
