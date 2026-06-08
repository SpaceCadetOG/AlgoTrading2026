package paper

type ExitDecision struct {
	Action           string
	ClosePct         float64
	Reason           string
	UpgradeBreakEven bool
	NewTrailingStop  float64
	ForceFlat        bool
}

func EvaluateExit(position PaperPosition, mark float64, last float64, noFollowThrough bool, fundingHazard bool, endOfDay bool) ExitDecision {
	switch normalizeSide(position.Side) {
	case "LONG":
		if position.TrailingStop > 0 && mark <= position.TrailingStop {
			return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "trailing_stop", ForceFlat: true}
		}
		if mark <= position.Stop {
			return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "hard_stop", ForceFlat: true}
		}
		if !position.TP1Taken && mark >= position.TP1 {
			return ExitDecision{Action: "partial", ClosePct: 0.50, Reason: "tp1", UpgradeBreakEven: true}
		}
		if position.TP1Taken && !position.TP2Taken && mark >= position.TP2 {
			return ExitDecision{Action: "partial", ClosePct: 0.30, Reason: "tp2", NewTrailingStop: position.TP1}
		}
		if position.TP2Taken && mark >= position.TP3 {
			return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "tp3"}
		}
	case "SHORT":
		if position.TrailingStop > 0 && mark >= position.TrailingStop {
			return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "trailing_stop", ForceFlat: true}
		}
		if mark >= position.Stop {
			return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "hard_stop", ForceFlat: true}
		}
		if !position.TP1Taken && mark <= position.TP1 {
			return ExitDecision{Action: "partial", ClosePct: 0.50, Reason: "tp1", UpgradeBreakEven: true}
		}
		if position.TP1Taken && !position.TP2Taken && mark <= position.TP2 {
			return ExitDecision{Action: "partial", ClosePct: 0.30, Reason: "tp2", NewTrailingStop: position.TP1}
		}
		if position.TP2Taken && mark <= position.TP3 {
			return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "tp3"}
		}
	}
	if noFollowThrough {
		return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "no_follow_through"}
	}
	if fundingHazard {
		return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "funding_exit"}
	}
	if endOfDay {
		return ExitDecision{Action: "close", ClosePct: 1.0, Reason: "force_flat_eod", ForceFlat: true}
	}
	return ExitDecision{}
}
