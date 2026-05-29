package risk

func (e *Engine) ActivateKillSwitch() {
	e.Limits.KillSwitchActive = true
}

func (e *Engine) ClearKillSwitch() {
	e.Limits.KillSwitchActive = false
}

func (e *Engine) IsKilled() bool {
	return e.Limits.KillSwitchActive
}
