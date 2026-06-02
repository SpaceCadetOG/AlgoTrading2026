package research

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AlgoTrading2026/vwap"
)

func TestBuildVWAPBehaviorSummary(t *testing.T) {
	rows := []vwap.Interaction{
		{AboveVWAP: true, TouchedVWAP: true, CrossedVWAP: true, Reaction: vwap.ReactionBounce},
		{BelowVWAP: true, TouchedVWAP: true, Reaction: vwap.ReactionBreak},
		{BelowVWAP: true, TouchedVWAP: true, Reaction: vwap.ReactionChop},
	}
	magnet := []vwap.MagnetStats{{DistanceBandPct: 0.25, Events: 2, Returns: 1, ReturnProbability: 0.5}}

	summary := BuildVWAPBehaviorSummary(rows, magnet)
	if summary.Candles != 3 || summary.TimeAboveVWAP != 1 || summary.TimeBelowVWAP != 2 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.Touches != 3 || summary.Crosses != 1 || summary.Bounces != 1 || summary.Breaks != 1 || summary.Chops != 1 {
		t.Fatalf("unexpected interaction counts: %+v", summary)
	}
	if summary.BounceRate != 1.0/3.0 || summary.BreakRate != 1.0/3.0 || summary.ChopRate != 1.0/3.0 {
		t.Fatalf("unexpected rates: %+v", summary)
	}
	if len(summary.MagnetStats) != 1 {
		t.Fatalf("magnet stats len=%d", len(summary.MagnetStats))
	}
}

func TestWriteVWAPInteractionAndReactionCSV(t *testing.T) {
	dir := t.TempDir()
	interactionPath := filepath.Join(dir, "vwap_interaction.csv")
	reactionPath := filepath.Join(dir, "vwap_reactions.csv")

	interactions := []vwap.Interaction{{
		Timestamp:        1,
		Close:            100,
		High:             101,
		Low:              99,
		Volume:           10,
		SessionVWAP:      100,
		DistanceFromVWAP: 0,
		TouchedVWAP:      true,
		Reaction:         vwap.ReactionChop,
	}}
	if err := WriteVWAPInteractionCSV(interactionPath, interactions); err != nil {
		t.Fatalf("write interaction csv: %v", err)
	}
	interactionBody, err := os.ReadFile(interactionPath)
	if err != nil {
		t.Fatalf("read interaction csv: %v", err)
	}
	if !strings.Contains(string(interactionBody), "timestamp,close,high,low,volume,session_vwap,distance_from_vwap") {
		t.Fatalf("missing interaction header: %s", string(interactionBody))
	}
	if !strings.Contains(string(interactionBody), "CHOP") {
		t.Fatalf("missing interaction reaction: %s", string(interactionBody))
	}

	reactions := []vwap.ReactionStudy{{
		Timestamp:        1,
		Reaction:         vwap.ReactionBreak,
		Close:            100,
		VWAP:             99,
		DistanceFromVWAP: 1,
	}}
	if err := WriteVWAPReactionsCSV(reactionPath, reactions); err != nil {
		t.Fatalf("write reaction csv: %v", err)
	}
	reactionBody, err := os.ReadFile(reactionPath)
	if err != nil {
		t.Fatalf("read reaction csv: %v", err)
	}
	if !strings.Contains(string(reactionBody), "timestamp,reaction,close,vwap,distance_from_vwap") {
		t.Fatalf("missing reaction header: %s", string(reactionBody))
	}
	if !strings.Contains(string(reactionBody), "BREAK") {
		t.Fatalf("missing reaction: %s", string(reactionBody))
	}
}

func TestWriteVWAPBehaviorSummaryJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "summary.json")
	summary := VWAPBehaviorSummary{
		Candles:    3,
		Touches:    2,
		Bounces:    1,
		BounceRate: 0.5,
	}
	if err := WriteVWAPBehaviorSummaryJSON(path, summary); err != nil {
		t.Fatalf("write summary json: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read summary json: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, `"candles": 3`) || !strings.Contains(text, `"bounceRate": 0.5`) {
		t.Fatalf("unexpected summary json: %s", text)
	}
}
