package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"AlgoTrading2026/system"
)

type Chapter8SessionLifecycle struct {
	Title      string                       `json:"title"`
	Sessions   []Chapter8SessionDescription `json:"sessions"`
	Commands   []string                     `json:"commands"`
	Statuses   []string                     `json:"statuses"`
	Summary    system.GatewaySessionSummary `json:"summary"`
	FilesAdded []string                     `json:"filesAdded"`
	Outputs    []string                     `json:"outputs"`
	Conclusion string                       `json:"conclusion"`
	NextPhase  string                       `json:"nextPhase"`
}

type Chapter8SessionDescription struct {
	Venue           string `json:"venue"`
	SessionID       string `json:"sessionID"`
	Status          string `json:"status"`
	MessageCount    int    `json:"messageCount"`
	LastError       string `json:"lastError"`
	Communication   string `json:"communication"`
	NetworkBehavior string `json:"networkBehavior"`
}

func BuildChapter8SessionLifecycle() Chapter8SessionLifecycle {
	manager := system.NewGatewaySessionManager()
	for _, venue := range []string{"aster", "hyperliquid", "lighter"} {
		manager.Register(system.NewGatewaySession(venue))
	}
	manager.ConnectAll()
	manager.HeartbeatAll()
	for _, venue := range []string{"aster", "hyperliquid", "lighter"} {
		if session, ok := manager.Session(venue); ok {
			session.RecordMessage()
		}
	}

	sessions := make([]Chapter8SessionDescription, 0, 3)
	for _, venue := range []string{"aster", "hyperliquid", "lighter"} {
		session, _ := manager.Session(venue)
		sessions = append(sessions, Chapter8SessionDescription{
			Venue:           venue,
			SessionID:       session.SessionID,
			Status:          string(session.Status),
			MessageCount:    session.MessageCount,
			LastError:       session.LastError,
			Communication:   "modeled gateway session only",
			NetworkBehavior: "no real sockets, REST calls, WebSockets, or order placement",
		})
	}

	return Chapter8SessionLifecycle{
		Title:    "Chapter 8D Gateway Session Lifecycle",
		Sessions: sessions,
		Commands: []string{
			"Connect",
			"Heartbeat",
			"Disconnect",
			"MarkError",
			"RecordMessage",
		},
		Statuses: []string{
			string(system.SessionCreated),
			string(system.SessionConnecting),
			string(system.SessionConnected),
			string(system.SessionHeartbeatOK),
			string(system.SessionDegraded),
			string(system.SessionDisconnected),
			string(system.SessionError),
		},
		Summary: manager.Summary(),
		FilesAdded: []string{
			"system/gateway_session.go",
			"system/gateway_session_test.go",
			"research/chapter8_session_lifecycle.go",
		},
		Outputs: []string{
			"research/chapter8_session_lifecycle.json",
			"research/chapter8_session_lifecycle.md",
		},
		Conclusion: "Chapter 8D models gateway session lifecycle, heartbeats, errors, disconnects, and message accounting without real network calls.",
		NextPhase:  "Chapter 8E can model request IDs and response correlation before any real adapter bridge is enabled.",
	}
}

func WriteChapter8SessionLifecycleJSON(path string, lifecycle Chapter8SessionLifecycle) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(lifecycle, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func WriteChapter8SessionLifecycleMarkdown(path string, lifecycle Chapter8SessionLifecycle) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", lifecycle.Title)
	fmt.Fprintf(&b, "## Statuses\n\n")
	for _, status := range lifecycle.Statuses {
		fmt.Fprintf(&b, "- %s\n", status)
	}

	fmt.Fprintf(&b, "\n## Commands\n\n")
	for _, command := range lifecycle.Commands {
		fmt.Fprintf(&b, "- %s\n", command)
	}

	fmt.Fprintf(&b, "\n## Sessions\n\n")
	fmt.Fprintf(&b, "| Venue | Session ID | Status | Messages | Last error | Network behavior |\n")
	fmt.Fprintf(&b, "|---|---|---|---|---|---|\n")
	for _, session := range lifecycle.Sessions {
		fmt.Fprintf(
			&b,
			"| %s | %s | %s | %d | %s | %s |\n",
			session.Venue,
			session.SessionID,
			session.Status,
			session.MessageCount,
			session.LastError,
			session.NetworkBehavior,
		)
	}

	fmt.Fprintf(&b, "\n## Summary\n\n")
	fmt.Fprintf(&b, "- total: %d\n", lifecycle.Summary.Total)
	fmt.Fprintf(&b, "- heartbeat_ok: %d\n", lifecycle.Summary.HeartbeatOK)
	fmt.Fprintf(&b, "- messages: %d\n", lifecycle.Summary.Messages)
	fmt.Fprintf(&b, "- errors: %d\n", lifecycle.Summary.Errors)

	fmt.Fprintf(&b, "\n## Conclusion\n\n%s\n\n", lifecycle.Conclusion)
	fmt.Fprintf(&b, "## Next Phase\n\n%s\n", lifecycle.NextPhase)

	return os.WriteFile(path, []byte(b.String()), 0644)
}
