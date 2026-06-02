package research

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVolumeProfileFinalBookPacketWriters(t *testing.T) {
	packet := BuildVolumeProfileFinalBookPacket([]string{"research/archive/example_audit.md"})
	if packet.Status == "" || packet.ArchivedAudits != 1 {
		t.Fatalf("packet=%+v", packet)
	}
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "packet.json")
	mdPath := filepath.Join(dir, "packet.md")
	if err := WriteVolumeProfileFinalBookPacketJSON(jsonPath, packet); err != nil {
		t.Fatalf("write json: %v", err)
	}
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var decoded VolumeProfileFinalBookPacket
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if decoded.ArchivedAudits != 1 {
		t.Fatalf("decoded=%+v", decoded)
	}
	if err := WriteVolumeProfileFinalBookPacketMarkdown(mdPath, packet); err != nil {
		t.Fatalf("write md: %v", err)
	}
	md, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read md: %v", err)
	}
	if !strings.Contains(string(md), "# Volume Profile Final Book Packet") {
		t.Fatalf("missing title: %s", string(md))
	}
}
