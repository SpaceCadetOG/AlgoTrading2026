package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"AlgoTrading2026/exchanges"
)

type VenueHealth struct {
	Venue                 string                 `json:"venue"`
	Healthy               bool                   `json:"healthy"`
	AccountReady          bool                   `json:"accountReady"`
	MarketDataReady       bool                   `json:"marketDataReady"`
	PrivateEndpointsReady bool                   `json:"privateEndpointsReady"`
	PositionSyncReady     bool                   `json:"positionSyncReady"`
	FillHistoryReady      bool                   `json:"fillHistoryReady"`
	ProtectionReady       bool                   `json:"protectionReady"`
	ReconciliationReady   bool                   `json:"reconciliationReady"`
	StreamReady           bool                   `json:"streamReady"`
	Protection            ProtectionCapabilities `json:"protection"`
	Stream                StreamStatus           `json:"stream"`
	AvailableBalance      float64                `json:"availableBalance,omitempty"`
	Reasons               []string               `json:"reasons,omitempty"`
}

type AccountHealthSnapshot struct {
	Timestamp int64         `json:"timestamp"`
	TimeUTC   string        `json:"timeUtc"`
	Venues    []VenueHealth `json:"venues"`
}

type VenueHealthChecker struct {
	Venue                 string
	Account               exchanges.AccountProvider
	Positions             exchanges.PositionProvider
	RequirePositions      bool
	RequireFills          bool
	Fills                 FillReader
	MarketDataReady       bool
	PrivateEndpointsReady bool
	Protection            ProtectionCapabilities
	Stream                StreamStatus
	RequireFreshStream    bool
}

func (c VenueHealthChecker) Check() VenueHealth {
	health := VenueHealth{
		Venue:                 strings.ToLower(strings.TrimSpace(c.Venue)),
		MarketDataReady:       c.MarketDataReady,
		PrivateEndpointsReady: c.PrivateEndpointsReady,
		PositionSyncReady:     !c.RequirePositions,
		FillHistoryReady:      !c.RequireFills,
		Protection:            c.Protection,
		Stream:                c.Stream,
	}
	health.StreamReady = !c.RequireFreshStream
	if c.Stream.Venue != "" {
		health.Stream = c.Stream
		health.StreamReady = c.Stream.Connected && !c.Stream.Stale
	}
	if health.Protection.Venue == "" {
		health.Protection = ProtectionForVenue(c.Venue)
	}
	health.ProtectionReady = health.Protection.Ready
	if !health.ProtectionReady {
		health.Reasons = append(health.Reasons, "protection_unavailable:"+health.Protection.Reason)
	}
	if !health.MarketDataReady {
		health.Reasons = append(health.Reasons, "market_metadata_unavailable")
	}
	if !health.PrivateEndpointsReady {
		health.Reasons = append(health.Reasons, "private_endpoints_unavailable")
	}
	if c.Account == nil {
		health.Reasons = append(health.Reasons, "account_provider_unavailable")
	} else if snapshot, err := c.Account.GetAccountSnapshot(); err != nil {
		health.Reasons = append(health.Reasons, "account_snapshot_failed:"+err.Error())
	} else {
		health.AccountReady = true
		health.AvailableBalance = parseFloat(snapshot.Available)
		if snapshot.Available == "" && snapshot.WalletValue == "" {
			health.Reasons = append(health.Reasons, "account_balance_unavailable")
			health.AccountReady = false
		}
	}
	if c.RequirePositions {
		if c.Positions == nil {
			health.Reasons = append(health.Reasons, "position_provider_unavailable")
		} else if _, err := c.Positions.GetPositions(); err != nil {
			health.Reasons = append(health.Reasons, "position_snapshot_failed:"+err.Error())
		} else {
			health.PositionSyncReady = true
		}
	}
	if c.RequireFills {
		if c.Fills == nil {
			health.Reasons = append(health.Reasons, "fill_history_provider_unavailable")
		} else {
			health.FillHistoryReady = true
		}
	}
	health.ReconciliationReady = health.PositionSyncReady && health.FillHistoryReady
	if c.RequireFreshStream && !health.StreamReady {
		health.Reasons = append(health.Reasons, "stream_not_ready")
	}
	if !health.ReconciliationReady {
		health.Reasons = append(health.Reasons, "reconciliation_not_ready")
	}
	health.Healthy = health.AccountReady &&
		health.MarketDataReady &&
		health.PrivateEndpointsReady &&
		health.PositionSyncReady &&
		health.FillHistoryReady &&
		health.ProtectionReady &&
		health.StreamReady
	return health
}

func HealthGateFromSnapshot(base LiveGate, health VenueHealth) LiveGate {
	gate := base
	gate.VenueHealthy = health.Healthy
	gate.AccountReady = health.AccountReady
	return gate
}

func WriteHealthSnapshots(root string, snapshot AccountHealthSnapshot) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(root, "venue_health.json"), append(body, '\n'), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "account_health.json"), append(body, '\n'), 0o644)
}

func BuildHealthSnapshot(checkers map[string]VenueHealthChecker) AccountHealthSnapshot {
	venues := make([]string, 0, len(checkers))
	for venue := range checkers {
		venues = append(venues, venue)
	}
	sort.Strings(venues)
	now := time.Now().UTC()
	snapshot := AccountHealthSnapshot{
		Timestamp: now.UnixMilli(),
		TimeUTC:   now.Format(time.RFC3339),
	}
	for _, venue := range venues {
		snapshot.Venues = append(snapshot.Venues, checkers[venue].Check())
	}
	return snapshot
}

func HealthByVenue(snapshot AccountHealthSnapshot) map[string]VenueHealth {
	out := map[string]VenueHealth{}
	for _, health := range snapshot.Venues {
		out[strings.ToLower(strings.TrimSpace(health.Venue))] = health
	}
	return out
}

func (h VenueHealth) RefusalReason() string {
	if h.Healthy {
		return ""
	}
	if len(h.Reasons) == 0 {
		return "venue_health_failed"
	}
	return fmt.Sprintf("venue_health_failed:%s", strings.Join(h.Reasons, "|"))
}

func FloatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
