package research

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"AlgoTrading2026/riskmetrics"
)

type Chapter6RiskPacket struct {
	BestRiskAdjustedStrategies []Chapter6PacketStrategy `json:"bestRiskAdjustedStrategies"`
	KeptThrottled              []Chapter6PacketStrategy `json:"keptThrottled"`
	RewriteRequired            []Chapter6PacketStrategy `json:"rewriteRequired"`
	RemovedFromCandidates      []Chapter6PacketStrategy `json:"removedFromCandidates"`
	TopRiskIssues              []string                 `json:"topRiskIssues"`
	Conclusion                 []string                 `json:"conclusion"`
	NextRecommendedPhase       string                   `json:"nextRecommendedPhase"`
}

type Chapter6PacketStrategy struct {
	Rank     int      `json:"rank"`
	Strategy string   `json:"strategy"`
	Gate     string   `json:"gate"`
	Decision string   `json:"decision"`
	Score    float64  `json:"riskAdjustedScore"`
	Reasons  []string `json:"reasons"`
}

func BuildChapter6RiskPacket(rankings []riskmetrics.RankedStrategy, gates []riskmetrics.PromotionDecision) Chapter6RiskPacket {
	gatesByStrategyRank := make(map[string]riskmetrics.PromotionDecision, len(gates))
	for _, gate := range gates {
		gatesByStrategyRank[packetKey(gate.Rank, gate.Strategy)] = gate
	}

	packet := Chapter6RiskPacket{
		TopRiskIssues: []string{
			"low_sharpe",
			"negative_expectancy",
			"excessive_trades_per_day",
			"high_variance",
			"poor_risk_grade",
		},
		Conclusion: []string{
			"No strategy is promoted to active/paper execution yet.",
			"RSI, Bollinger, mean_reversion, vol_mean_reversion, and dual_ma remain research candidates but throttled.",
			"EMA and momentum are removed from candidates.",
			"Risk layer remains rule-based.",
			"No ML has been introduced.",
		},
		NextRecommendedPhase: "Chapter 7: controlled paper-trading infrastructure and monitoring gates, without enabling active execution.",
	}

	for i, row := range rankings {
		if i < 5 {
			packet.BestRiskAdjustedStrategies = append(packet.BestRiskAdjustedStrategies, packetStrategyFromRanking(row, gatesByStrategyRank))
		}
	}

	for _, gate := range gates {
		item := packetStrategyFromGate(gate)
		switch gate.Gate {
		case riskmetrics.KeepThrottled:
			packet.KeptThrottled = append(packet.KeptThrottled, item)
		case riskmetrics.RewriteRequired:
			packet.RewriteRequired = append(packet.RewriteRequired, item)
		case riskmetrics.RemoveFromCandidates:
			packet.RemovedFromCandidates = append(packet.RemovedFromCandidates, item)
		}
	}

	sortPacketStrategies(packet.KeptThrottled)
	sortPacketStrategies(packet.RewriteRequired)
	sortPacketStrategies(packet.RemovedFromCandidates)

	return packet
}

func WriteChapter6RiskPacketJSON(path string, packet Chapter6RiskPacket) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter6RiskPacketMarkdown(path string, packet Chapter6RiskPacket) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter6RiskPacketMarkdown(packet)), 0644)
}

func Chapter6RiskPacketMarkdown(packet Chapter6RiskPacket) string {
	var b strings.Builder
	b.WriteString("# Chapter 6 Final Risk Packet\n\n")
	b.WriteString("## Best Risk-Adjusted Strategies\n\n")
	writePacketTable(&b, packet.BestRiskAdjustedStrategies)
	b.WriteString("\n## Strategies Kept Throttled\n\n")
	writePacketTable(&b, packet.KeptThrottled)
	b.WriteString("\n## Strategies Requiring Rewrite\n\n")
	writePacketTable(&b, packet.RewriteRequired)
	b.WriteString("\n## Strategies Removed From Candidates\n\n")
	writePacketTable(&b, packet.RemovedFromCandidates)
	b.WriteString("\n## Top Risk Issues Observed\n\n")
	for _, issue := range packet.TopRiskIssues {
		b.WriteString("- " + issue + "\n")
	}
	b.WriteString("\n## Chapter 6 Conclusion\n\n")
	for _, line := range packet.Conclusion {
		b.WriteString("- " + line + "\n")
	}
	b.WriteString("\n## Next Recommended Chapter/Build Phase\n\n")
	b.WriteString(packet.NextRecommendedPhase + "\n")
	return b.String()
}

func packetStrategyFromRanking(row riskmetrics.RankedStrategy, gatesByStrategyRank map[string]riskmetrics.PromotionDecision) Chapter6PacketStrategy {
	gate := ""
	if decision, ok := gatesByStrategyRank[packetKey(row.Rank, row.Strategy)]; ok {
		gate = string(decision.Gate)
	}
	return Chapter6PacketStrategy{
		Rank:     row.Rank,
		Strategy: row.Strategy,
		Gate:     gate,
		Decision: string(row.Decision),
		Score:    row.Score,
		Reasons:  append([]string(nil), row.Reasons...),
	}
}

func packetStrategyFromGate(gate riskmetrics.PromotionDecision) Chapter6PacketStrategy {
	return Chapter6PacketStrategy{
		Rank:     gate.Rank,
		Strategy: gate.Strategy,
		Gate:     string(gate.Gate),
		Decision: string(gate.FilterDecision),
		Score:    gate.Score,
		Reasons:  append([]string(nil), gate.Reasons...),
	}
}

func writePacketTable(b *strings.Builder, rows []Chapter6PacketStrategy) {
	if len(rows) == 0 {
		b.WriteString("_None._\n")
		return
	}

	b.WriteString("| Rank | Strategy | Gate | Decision | Score | Reasons |\n")
	b.WriteString("|---:|---|---|---|---:|---|\n")
	for _, row := range rows {
		b.WriteString(fmt.Sprintf(
			"| %d | %s | %s | %s | %.2f | %s |\n",
			row.Rank,
			row.Strategy,
			row.Gate,
			row.Decision,
			row.Score,
			strings.Join(row.Reasons, ", "),
		))
	}
}

func sortPacketStrategies(rows []Chapter6PacketStrategy) {
	sort.SliceStable(rows, func(i int, j int) bool {
		return rows[i].Rank < rows[j].Rank
	})
}

func packetKey(rank int, strategy string) string {
	return fmt.Sprintf("%d:%s", rank, strategy)
}
