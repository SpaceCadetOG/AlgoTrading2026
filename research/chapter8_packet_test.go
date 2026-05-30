package research

import (
	"os"
	"strings"
	"testing"
)

func TestChapter8PacketContainsRequiredConclusion(t *testing.T) {
	packet := BuildChapter8Packet()
	required := []string{
		"Chapter 8 exchange connectivity layer is mapped and safely abstracted.",
		"Real venue adapters remain disabled by default.",
		"No live or paper execution is enabled.",
		"The system is ready for Chapter 9 backtester audit.",
	}

	for _, want := range required {
		found := false
		for _, got := range packet.Conclusion {
			if got == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing conclusion %q in %+v", want, packet.Conclusion)
		}
	}
}

func TestChapter8PacketWriters(t *testing.T) {
	packet := BuildChapter8Packet()
	dir := t.TempDir()
	jsonPath := dir + "/chapter8_packet.json"
	mdPath := dir + "/chapter8_packet.md"

	if err := WriteChapter8PacketJSON(jsonPath, packet); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter8PacketMarkdown(mdPath, packet); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(data)
	for _, want := range []string{"Aster", "Hyperliquid", "Lighter", "Missing FIX Items", "Readiness For Chapter 9"} {
		if !strings.Contains(strings.ToLower(text), strings.ToLower(want)) {
			t.Fatalf("expected markdown to contain %q", want)
		}
	}
}
