package research

import (
	"os"
	"strings"
	"testing"
)

func TestChapter8GatewayAuditWriters(t *testing.T) {
	audit := BuildChapter8GatewayAudit()
	if len(audit.Venues) != 3 {
		t.Fatalf("expected three venue audits, got %d", len(audit.Venues))
	}

	dir := t.TempDir()
	jsonPath := dir + "/chapter8_gateway_audit.json"
	mdPath := dir + "/chapter8_gateway_audit.md"

	if err := WriteChapter8GatewayAuditJSON(jsonPath, audit); err != nil {
		t.Fatalf("write json: %v", err)
	}
	if err := WriteChapter8GatewayAuditMarkdown(mdPath, audit); err != nil {
		t.Fatalf("write markdown: %v", err)
	}

	data, err := os.ReadFile(mdPath)
	if err != nil {
		t.Fatalf("read markdown: %v", err)
	}
	text := string(data)
	for _, expected := range []string{"aster", "hyperliquid", "lighter", "system.Gateway"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected markdown to contain %q", expected)
		}
	}
}
