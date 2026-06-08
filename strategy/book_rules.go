package strategy

func DefaultBookTradeRules() []ExecutablePlaybook {
	playbooks := append([]ExecutablePlaybook{}, volumeProfileRulePack()...)
	playbooks = append(playbooks, orderFlowRulePack()...)
	playbooks = append(playbooks, vwapRulePack()...)
	return playbooks
}

func DefaultBookTradeRulesPacket() BookTradeRulesPacket {
	return BookTradeRulesPacket{
		Playbooks:           DefaultBookTradeRules(),
		ExecutionEnabled:    false,
		PaperTradingEnabled: false,
		Status:              "ready_for_paper_engine",
	}
}

func DefaultPlaybooks() []TradeSetup {
	return DefaultBookTradeRules()
}

func DefaultPlaybookPacket() PlaybookPacket {
	return DefaultBookTradeRulesPacket()
}
