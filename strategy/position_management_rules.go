package strategy

func conservativeManagement(name string, sourceBook string, breakEvenAfter string, trailAfter string, earlyExitIf []string, forceFlatIf []string) TradeManagementRule {
	return TradeManagementRule{
		Name:           name,
		SourceBook:     sourceBook,
		BreakEvenAfter: breakEvenAfter,
		TrailAfter:     trailAfter,
		EarlyExitIf:    earlyExitIf,
		ForceFlatIf:    forceFlatIf,
	}
}
