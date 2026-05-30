package system

import (
	"fmt"
	"sort"
	"time"
)

type GatewaySessionStatus string

const (
	SessionCreated      GatewaySessionStatus = "CREATED"
	SessionConnecting   GatewaySessionStatus = "CONNECTING"
	SessionConnected    GatewaySessionStatus = "CONNECTED"
	SessionHeartbeatOK  GatewaySessionStatus = "HEARTBEAT_OK"
	SessionDegraded     GatewaySessionStatus = "DEGRADED"
	SessionDisconnected GatewaySessionStatus = "DISCONNECTED"
	SessionError        GatewaySessionStatus = "ERROR"
)

type GatewaySession struct {
	SessionID       string
	Venue           string
	Status          GatewaySessionStatus
	ConnectedAt     time.Time
	LastHeartbeatAt time.Time
	LastError       string
	MessageCount    int
}

func NewGatewaySession(venue string) *GatewaySession {
	return &GatewaySession{
		SessionID: fmt.Sprintf("%s-session", venue),
		Venue:     venue,
		Status:    SessionCreated,
	}
}

func (s *GatewaySession) Connect() {
	now := time.Now().UTC()
	s.Status = SessionConnecting
	s.ConnectedAt = now
	s.LastHeartbeatAt = now
	s.LastError = ""
	s.Status = SessionConnected
}

func (s *GatewaySession) Heartbeat() {
	now := time.Now().UTC()
	s.LastHeartbeatAt = now
	if s.Status != SessionError && s.Status != SessionDisconnected {
		s.Status = SessionHeartbeatOK
	}
}

func (s *GatewaySession) Disconnect() {
	s.Status = SessionDisconnected
}

func (s *GatewaySession) MarkError(err error) {
	s.Status = SessionError
	if err != nil {
		s.LastError = err.Error()
		return
	}
	s.LastError = "unknown gateway session error"
}

func (s *GatewaySession) RecordMessage() {
	s.MessageCount++
	if s.Status == SessionCreated || s.Status == SessionDisconnected {
		s.Status = SessionDegraded
	}
}

type GatewaySessionSummary struct {
	Total        int
	Connected    int
	HeartbeatOK  int
	Degraded     int
	Disconnected int
	Errors       int
	Messages     int
	Statuses     map[string]GatewaySessionStatus
}

type GatewaySessionManager struct {
	sessions map[string]*GatewaySession
}

func NewGatewaySessionManager() *GatewaySessionManager {
	return &GatewaySessionManager{sessions: make(map[string]*GatewaySession)}
}

func (m *GatewaySessionManager) Register(session *GatewaySession) {
	if session == nil {
		return
	}
	m.sessions[session.Venue] = session
}

func (m *GatewaySessionManager) Session(venue string) (*GatewaySession, bool) {
	session, ok := m.sessions[venue]
	return session, ok
}

func (m *GatewaySessionManager) ConnectAll() {
	for _, session := range m.sessions {
		session.Connect()
	}
}

func (m *GatewaySessionManager) HeartbeatAll() {
	for _, session := range m.sessions {
		session.Heartbeat()
	}
}

func (m *GatewaySessionManager) DisconnectAll() {
	for _, session := range m.sessions {
		session.Disconnect()
	}
}

func (m *GatewaySessionManager) Summary() GatewaySessionSummary {
	summary := GatewaySessionSummary{
		Total:    len(m.sessions),
		Statuses: make(map[string]GatewaySessionStatus, len(m.sessions)),
	}

	venues := make([]string, 0, len(m.sessions))
	for venue := range m.sessions {
		venues = append(venues, venue)
	}
	sort.Strings(venues)

	for _, venue := range venues {
		session := m.sessions[venue]
		summary.Statuses[venue] = session.Status
		summary.Messages += session.MessageCount
		switch session.Status {
		case SessionConnected:
			summary.Connected++
		case SessionHeartbeatOK:
			summary.HeartbeatOK++
		case SessionDegraded:
			summary.Degraded++
		case SessionDisconnected:
			summary.Disconnected++
		case SessionError:
			summary.Errors++
		}
	}

	return summary
}
