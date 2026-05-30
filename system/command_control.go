package system

import (
	"fmt"
	"time"
)

type Command string

const (
	CommandStart  Command = "START"
	CommandStop   Command = "STOP"
	CommandStatus Command = "STATUS"
	CommandPause  Command = "PAUSE"
	CommandResume Command = "RESUME"
)

type SystemState string

const (
	StateCreated SystemState = "CREATED"
	StateRunning SystemState = "RUNNING"
	StatePaused  SystemState = "PAUSED"
	StateStopped SystemState = "STOPPED"
	StateError   SystemState = "ERROR"
)

type CommandAuditEntry struct {
	Time    time.Time   `json:"time"`
	Command Command     `json:"command"`
	Before  SystemState `json:"before"`
	After   SystemState `json:"after"`
	Message string      `json:"message"`
}

type CommandControl struct {
	state      SystemState
	supervisor *SystemSupervisor
	audit      []CommandAuditEntry
}

func NewCommandControl(supervisor *SystemSupervisor) *CommandControl {
	return &CommandControl{state: StateCreated, supervisor: supervisor}
}

func (c *CommandControl) State() SystemState {
	return c.state
}

func (c *CommandControl) AuditLog() []CommandAuditEntry {
	return append([]CommandAuditEntry(nil), c.audit...)
}

func (c *CommandControl) Handle(command Command) (SystemState, error) {
	before := c.state
	after := before
	message := "ok"
	var err error

	switch command {
	case CommandStart:
		if c.supervisor != nil {
			err = c.supervisor.StartAll()
		}
		if err == nil {
			after = StateRunning
		}
	case CommandStop:
		if c.supervisor != nil {
			err = c.supervisor.StopAll()
		}
		if err == nil {
			after = StateStopped
		}
	case CommandPause:
		if before == StateRunning {
			after = StatePaused
		} else {
			err = fmt.Errorf("cannot pause from %s", before)
		}
	case CommandResume:
		if before == StatePaused {
			after = StateRunning
		} else {
			err = fmt.Errorf("cannot resume from %s", before)
		}
	case CommandStatus:
		after = before
	default:
		err = fmt.Errorf("unsupported command %s", command)
	}

	if err != nil {
		after = StateError
		message = err.Error()
	}
	c.state = after
	c.audit = append(c.audit, CommandAuditEntry{
		Time:    time.Now().UTC(),
		Command: command,
		Before:  before,
		After:   after,
		Message: message,
	})
	return c.state, err
}
