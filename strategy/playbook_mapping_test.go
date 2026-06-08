package strategy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBookTradeRuleWriters(t *testing.T) {
	dir := t.TempDir()
	packet := DefaultBookTradeRulesPacket()
	jsonPath := filepath.Join(dir, "book_trade_rules_packet.json")
	mdPath := filepath.Join(dir, "book_trade_rules_packet.md")
	docPath := filepath.Join(dir, "book_trade_rules.md")

	if err := WriteBookTradeRulesPacketJSON(jsonPath, packet); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteBookTradeRulesPacketMarkdown(mdPath, packet); err != nil {
		t.Fatalf("write markdown: %v", err)
	}
	if err := WriteBookTradeRulesMarkdown(docPath, packet.Playbooks); err != nil {
		t.Fatalf("write docs: %v", err)
	}

	var decoded BookTradeRulesPacket
	body, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if len(decoded.Playbooks) != 9 {
		t.Fatalf("decoded playbooks = %d", len(decoded.Playbooks))
	}

	for _, path := range []string{mdPath, docPath} {
		text, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read markdown %s: %v", path, err)
		}
		if !strings.Contains(string(text), "VWAP First Touch Combo") {
			t.Fatalf("expected rule content in %s", path)
		}
	}
}
