package system

import (
	"errors"
	"testing"
)

func TestGatewaySessionLifecycleTransitions(t *testing.T) {
	session := NewGatewaySession("aster")
	if session.Status != SessionCreated {
		t.Fatalf("expected created, got %s", session.Status)
	}

	session.Connect()
	if session.Status != SessionConnected {
		t.Fatalf("expected connected, got %s", session.Status)
	}
	if session.ConnectedAt.IsZero() {
		t.Fatal("expected connected timestamp")
	}

	session.Heartbeat()
	if session.Status != SessionHeartbeatOK {
		t.Fatalf("expected heartbeat ok, got %s", session.Status)
	}
	if session.LastHeartbeatAt.IsZero() {
		t.Fatal("expected heartbeat timestamp")
	}

	session.Disconnect()
	if session.Status != SessionDisconnected {
		t.Fatalf("expected disconnected, got %s", session.Status)
	}
}

func TestGatewaySessionRecordMessageAndError(t *testing.T) {
	session := NewGatewaySession("hyperliquid")

	session.RecordMessage()
	if session.MessageCount != 1 {
		t.Fatalf("expected one message, got %d", session.MessageCount)
	}
	if session.Status != SessionDegraded {
		t.Fatalf("expected degraded for message before connect, got %s", session.Status)
	}

	session.MarkError(errors.New("heartbeat timeout"))
	if session.Status != SessionError {
		t.Fatalf("expected error status, got %s", session.Status)
	}
	if session.LastError != "heartbeat timeout" {
		t.Fatalf("unexpected last error: %s", session.LastError)
	}
}

func TestGatewaySessionManagerMultipleVenues(t *testing.T) {
	manager := NewGatewaySessionManager()
	manager.Register(NewGatewaySession("aster"))
	manager.Register(NewGatewaySession("hyperliquid"))
	manager.Register(NewGatewaySession("lighter"))

	manager.ConnectAll()
	summary := manager.Summary()
	if summary.Total != 3 || summary.Connected != 3 {
		t.Fatalf("unexpected connected summary: %+v", summary)
	}

	manager.HeartbeatAll()
	summary = manager.Summary()
	if summary.HeartbeatOK != 3 {
		t.Fatalf("unexpected heartbeat summary: %+v", summary)
	}

	aster, ok := manager.Session("aster")
	if !ok {
		t.Fatal("expected aster session")
	}
	aster.RecordMessage()
	aster.RecordMessage()
	summary = manager.Summary()
	if summary.Messages != 2 {
		t.Fatalf("expected two messages, got %+v", summary)
	}

	lighter, ok := manager.Session("lighter")
	if !ok {
		t.Fatal("expected lighter session")
	}
	lighter.MarkError(errors.New("decode failure"))
	summary = manager.Summary()
	if summary.Errors != 1 {
		t.Fatalf("expected one error, got %+v", summary)
	}

	manager.DisconnectAll()
	summary = manager.Summary()
	if summary.Disconnected != 3 {
		t.Fatalf("unexpected disconnected summary: %+v", summary)
	}
}
