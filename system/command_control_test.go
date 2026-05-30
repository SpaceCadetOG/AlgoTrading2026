package system

import "testing"

func TestCommandControlStartStopFlow(t *testing.T) {
	supervisor := NewChapter7Supervisor()
	control := NewCommandControl(supervisor)

	state, err := control.Handle(CommandStart)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if state != StateRunning {
		t.Fatalf("state = %s, want RUNNING", state)
	}
	if supervisor.CountByStatus(ServiceRunning) != supervisor.ServiceCount() {
		t.Fatalf("running services = %d, want %d", supervisor.CountByStatus(ServiceRunning), supervisor.ServiceCount())
	}

	state, err = control.Handle(CommandStop)
	if err != nil {
		t.Fatalf("stop: %v", err)
	}
	if state != StateStopped {
		t.Fatalf("state = %s, want STOPPED", state)
	}
	if supervisor.CountByStatus(ServiceStopped) != supervisor.ServiceCount() {
		t.Fatalf("stopped services = %d, want %d", supervisor.CountByStatus(ServiceStopped), supervisor.ServiceCount())
	}
}

func TestCommandControlPauseResumeFlow(t *testing.T) {
	control := NewCommandControl(NewChapter7Supervisor())

	if _, err := control.Handle(CommandStart); err != nil {
		t.Fatalf("start: %v", err)
	}
	state, err := control.Handle(CommandPause)
	if err != nil {
		t.Fatalf("pause: %v", err)
	}
	if state != StatePaused {
		t.Fatalf("state = %s, want PAUSED", state)
	}

	state, err = control.Handle(CommandResume)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if state != StateRunning {
		t.Fatalf("state = %s, want RUNNING", state)
	}
}

func TestSupervisorRegistrationStartStopAndSummary(t *testing.T) {
	supervisor := NewSystemSupervisor()
	supervisor.Register(NewLoggingService())
	supervisor.Register(NewPositionService())

	if supervisor.ServiceCount() != 2 {
		t.Fatalf("service count = %d, want 2", supervisor.ServiceCount())
	}
	if err := supervisor.StartAll(); err != nil {
		t.Fatalf("start all: %v", err)
	}
	if supervisor.CountByStatus(ServiceRunning) != 2 {
		t.Fatalf("running services = %d, want 2", supervisor.CountByStatus(ServiceRunning))
	}

	summary := supervisor.Summary()
	if len(summary) != 2 {
		t.Fatalf("summary len = %d, want 2", len(summary))
	}
	if summary[0].Name == "" || summary[1].Name == "" {
		t.Fatalf("summary missing service names: %+v", summary)
	}

	if err := supervisor.StopAll(); err != nil {
		t.Fatalf("stop all: %v", err)
	}
	if supervisor.CountByStatus(ServiceStopped) != 2 {
		t.Fatalf("stopped services = %d, want 2", supervisor.CountByStatus(ServiceStopped))
	}
}

func TestCommandAuditLogRecordsCommands(t *testing.T) {
	control := NewCommandControl(NewChapter7Supervisor())
	control.Handle(CommandStart)
	control.Handle(CommandStatus)
	control.Handle(CommandPause)

	audit := control.AuditLog()
	if len(audit) != 3 {
		t.Fatalf("audit entries = %d, want 3", len(audit))
	}
	if audit[0].Command != CommandStart || audit[1].Command != CommandStatus || audit[2].Command != CommandPause {
		t.Fatalf("unexpected audit log: %+v", audit)
	}
	if audit[2].After != StatePaused {
		t.Fatalf("pause audit after = %s, want PAUSED", audit[2].After)
	}
}
