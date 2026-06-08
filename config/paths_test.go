package config

import (
	"path/filepath"
	"testing"
)

func TestDataRootDefaultsAndSubdirs(t *testing.T) {
	t.Setenv("DATA_ROOT", "")

	if got := DataRoot(); got != filepath.Clean(filepath.Join(".", "data")) {
		t.Fatalf("unexpected default data root: %s", got)
	}
	if got := TradeTapeDir(); got != filepath.Join(".", "data", "trade_tape") {
		t.Fatalf("unexpected trade tape dir: %s", got)
	}
	if got := L2Dir(); got != filepath.Join(".", "data", "l2") {
		t.Fatalf("unexpected l2 dir: %s", got)
	}
	if got := PaperDir(); got != filepath.Join(".", "data", "paper") {
		t.Fatalf("unexpected paper dir: %s", got)
	}
}

func TestDataRootOverride(t *testing.T) {
	t.Setenv("DATA_ROOT", filepath.Join("C:", "tmp", "algo-data"))

	if got := DataRoot(); got != filepath.Join("C:", "tmp", "algo-data") {
		t.Fatalf("unexpected override data root: %s", got)
	}
	if got := TradeTapeDir(); got != filepath.Join("C:", "tmp", "algo-data", "trade_tape") {
		t.Fatalf("unexpected override trade tape dir: %s", got)
	}
	if got := L2Dir(); got != filepath.Join("C:", "tmp", "algo-data", "l2") {
		t.Fatalf("unexpected override l2 dir: %s", got)
	}
	if got := PaperDir(); got != filepath.Join("C:", "tmp", "algo-data", "paper") {
		t.Fatalf("unexpected override paper dir: %s", got)
	}
	if got := ArchiveDir(); got != filepath.Join("C:", "tmp", "algo-data", "archive") {
		t.Fatalf("unexpected override archive dir: %s", got)
	}
	if got := LogsDir(); got != filepath.Join("C:", "tmp", "algo-data", "logs") {
		t.Fatalf("unexpected override logs dir: %s", got)
	}
	if got := ExportsDir(); got != filepath.Join("C:", "tmp", "algo-data", "exports") {
		t.Fatalf("unexpected override exports dir: %s", got)
	}
}
