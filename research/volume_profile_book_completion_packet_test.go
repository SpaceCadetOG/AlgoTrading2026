package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeProfileBookCompletionPacketWriters(t *testing.T) {
	packet := BuildVolumeProfileBookCompletionPacket()
	if packet.Status != "complete" || packet.Setups != 4 || packet.Docs != 3 {
		t.Fatalf("packet=%+v", packet)
	}
	if len(packet.Gaps) == 0 || len(packet.RepoMapping) == 0 || len(packet.FinalRecommendation) == 0 {
		t.Fatalf("packet missing required sections: %+v", packet)
	}

	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "volume_profile_book_completion_packet.json")
	mdPath := filepath.Join(dir, "volume_profile_book_completion_packet.md")
	if err := WriteVolumeProfileBookCompletionPacketJSON(jsonPath, packet); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeProfileBookCompletionPacket
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.Status != "complete" {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeProfileBookCompletionPacketMarkdown(mdPath, packet); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	text := string(md)
	for _, expected := range []string{"# Volume Profile Book Completion Packet", "Trading Style Map", "Repo Mapping", "Do not add live or paper trading yet"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("missing %q in markdown:\n%s", expected, text)
		}
	}
}
