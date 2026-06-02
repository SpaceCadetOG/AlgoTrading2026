package research

import (
	"strings"
	"testing"

	"AlgoTrading2026/riskmetrics"
)

func TestBuildChapter6RiskPacketBucketsStrategies(t *testing.T) {
	rankings := []riskmetrics.RankedStrategy{
		{Rank: 1, Strategy: "rsi", Decision: riskmetrics.DecisionThrottle, Score: -1, Reasons: []string{"high_variance"}},
		{Rank: 2, Strategy: "ema", Decision: riskmetrics.DecisionBlock, Score: -130, Reasons: []string{"low_sharpe"}},
	}
	gates := []riskmetrics.PromotionDecision{
		{Rank: 1, Strategy: "rsi", Gate: riskmetrics.KeepThrottled, FilterDecision: riskmetrics.DecisionThrottle, Score: -1, Reasons: []string{"high_variance"}},
		{Rank: 2, Strategy: "ema", Gate: riskmetrics.RemoveFromCandidates, FilterDecision: riskmetrics.DecisionBlock, Score: -130, Reasons: []string{"blocked_by_risk_filter"}},
	}

	packet := BuildChapter6RiskPacket(rankings, gates)

	if len(packet.BestRiskAdjustedStrategies) != 2 {
		t.Fatalf("best strategies = %d, want 2", len(packet.BestRiskAdjustedStrategies))
	}
	if len(packet.KeptThrottled) != 1 || packet.KeptThrottled[0].Strategy != "rsi" {
		t.Fatalf("kept throttled = %+v", packet.KeptThrottled)
	}
	if len(packet.RemovedFromCandidates) != 1 || packet.RemovedFromCandidates[0].Strategy != "ema" {
		t.Fatalf("removed = %+v", packet.RemovedFromCandidates)
	}
}

func TestChapter6RiskPacketMarkdownIncludesConclusion(t *testing.T) {
	packet := BuildChapter6RiskPacket(nil, nil)
	body := Chapter6RiskPacketMarkdown(packet)

	for _, want := range []string{
		"Chapter 6 remains a research-only risk analysis packet.",
		"Risk layer remains rule-based.",
		"No ML has been introduced.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("markdown missing %q\n%s", want, body)
		}
	}
}
