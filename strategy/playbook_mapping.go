package strategy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func WriteBookTradeRulesPacketJSON(path string, packet BookTradeRulesPacket) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func WriteBookTradeRulesPacketMarkdown(path string, packet BookTradeRulesPacket) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var builder strings.Builder
	builder.WriteString("# Book Trade Rules Packet\n\n")
	builder.WriteString(fmt.Sprintf("Playbooks: %d\n\n", len(packet.Playbooks)))
	builder.WriteString(fmt.Sprintf("Execution Enabled: %t\n\n", packet.ExecutionEnabled))
	builder.WriteString(fmt.Sprintf("Paper Trading Enabled: %t\n\n", packet.PaperTradingEnabled))
	builder.WriteString(fmt.Sprintf("Status: %s\n\n", packet.Status))
	for _, playbook := range packet.Playbooks {
		writeBookTradeRuleSection(&builder, playbook)
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func WriteBookTradeRulesMarkdown(path string, playbooks []ExecutablePlaybook) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var builder strings.Builder
	builder.WriteString("# Book Trade Rules\n\n")
	builder.WriteString("These rules convert the three books into explicit bot-ready trade playbooks while keeping execution and paper trading disabled.\n\n")
	for _, playbook := range playbooks {
		writeBookTradeRuleSection(&builder, playbook)
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func writeBookTradeRuleSection(builder *strings.Builder, playbook ExecutablePlaybook) {
	builder.WriteString("## " + playbook.Name + "\n\n")
	builder.WriteString("Setup:\n")
	builder.WriteString(playbook.Setup + "\n\n")
	builder.WriteString("Entry:\n")
	builder.WriteString(playbook.Entry.Trigger + "\n\n")
	builder.WriteString("Stop:\n")
	builder.WriteString(playbook.Stop.Placement + "\n\n")
	builder.WriteString("Target:\n")
	builder.WriteString("TP1: " + playbook.Target.Target1 + "\n")
	builder.WriteString("TP2: " + playbook.Target.Target2 + "\n")
	builder.WriteString("TP3: " + playbook.Target.Target3 + "\n\n")
	builder.WriteString("Management:\n")
	builder.WriteString("Break-even after: " + playbook.Management.BreakEvenAfter + "\n")
	builder.WriteString("Trail after: " + playbook.Management.TrailAfter + "\n")
	builder.WriteString("What confirms it:\n")
	for _, item := range playbook.RequiredConfirmations {
		builder.WriteString("- " + item + "\n")
	}
	builder.WriteString("\nWhat cancels it:\n")
	for _, item := range playbook.Entry.InvalidIf {
		builder.WriteString("- " + item + "\n")
	}
	builder.WriteString("\nRisk:\n")
	builder.WriteString(fmt.Sprintf("Risk per trade: %.2f%%\n", playbook.Risk.RiskPerTradePct))
	builder.WriteString(fmt.Sprintf("Min reward-to-risk: %.2f\n", playbook.Risk.MinRewardToRisk))
	builder.WriteString("\n")
}

func WritePlaybookPacketJSON(path string, packet PlaybookPacket) error {
	return WriteBookTradeRulesPacketJSON(path, packet)
}

func WritePlaybookPacketMarkdown(path string, packet PlaybookPacket) error {
	return WriteBookTradeRulesPacketMarkdown(path, packet)
}

func WriteTradingPlaybookMarkdown(path string, playbooks []TradeSetup) error {
	return WriteBookTradeRulesMarkdown(path, playbooks)
}
