package research

import (
	"strings"
	"testing"
)

func TestBuildChapter7PacketContainsConclusion(t *testing.T) {
	packet := BuildChapter7Packet()
	body := strings.Join(packet.Conclusion, "\n")

	for _, want := range []string{
		"Chapter 7 trading-system skeleton is complete.",
		"No live/paper trading is enabled.",
		"Exchange adapters remain outside this system path until Chapter 8.",
		"RiskService is placeholder only.",
		"The system is ready to proceed to Chapter 8 exchange connectivity audit.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("conclusion missing %q", want)
		}
	}
}

func TestChapter7PacketMarkdownIncludesCoreSections(t *testing.T) {
	body := Chapter7PacketMarkdown(BuildChapter7Packet())
	for _, want := range []string{
		"## Components Implemented",
		"## Critical Components",
		"## Queue/Channel Data Flow",
		"## Readiness For Chapter 8",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("markdown missing %q", want)
		}
	}
}
