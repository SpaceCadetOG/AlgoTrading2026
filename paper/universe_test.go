package paper

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildManualUniverse(t *testing.T) {
	cfg := tempConfig(t)
	cfg.UniverseMode = "manual"
	cfg.Symbols = []string{"BTC", "ETH"}
	cfg.Venues = []string{"aster", "hyperliquid", "lighter"}

	selection, err := selectUniverseWithProviders(cfg, nil)
	if err != nil {
		t.Fatalf("select manual universe: %v", err)
	}
	if len(selection.Selected) != 6 {
		t.Fatalf("expected 6 manual rows, got %d", len(selection.Selected))
	}
	if selection.Selected[0].Venue != "aster" || selection.Selected[0].Symbol != "BTCUSDT" || selection.Selected[0].CanonicalSymbol != "BTC" {
		t.Fatalf("unexpected first manual row: %+v", selection.Selected[0])
	}
	body, err := os.ReadFile(cfg.UniverseCSVPath)
	if err != nil {
		t.Fatalf("read universe csv: %v", err)
	}
	if !strings.Contains(string(body), "aster,BTCUSDT,BTC") {
		t.Fatalf("expected aster btc row in csv: %s", string(body))
	}
}

func TestVolumeFilterUniverseSortsAndCaps(t *testing.T) {
	cfg := tempConfig(t)
	cfg.UniverseMode = "volume_filter"
	cfg.Min24hVolumeUSD = 5000000
	cfg.MaxSymbols = 2
	cfg.Venues = []string{"aster", "hyperliquid", "lighter"}

	providers := map[string]UniverseProvider{
		"aster": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC", Volume24hUSD: 9000000},
				{Venue: "aster", Symbol: "ETHUSDT", CanonicalSymbol: "ETH", Volume24hUSD: 1000000},
			}, nil
		},
		"hyperliquid": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "hyperliquid", Symbol: "ETH", CanonicalSymbol: "ETH", Volume24hUSD: 12000000},
			}, nil
		},
		"lighter": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "lighter", Symbol: "BTC", CanonicalSymbol: "BTC", Volume24hUSD: 7000000},
			}, nil
		},
	}

	selection, err := selectUniverseWithProviders(cfg, providers)
	if err != nil {
		t.Fatalf("select volume-filter universe: %v", err)
	}
	if len(selection.Selected) != 2 {
		t.Fatalf("expected 2 capped rows, got %d", len(selection.Selected))
	}
	if selection.Selected[0].Venue != "hyperliquid" || selection.Selected[1].Venue != "aster" {
		t.Fatalf("unexpected sort order: %+v", selection.Selected)
	}
	body, err := os.ReadFile(cfg.UniverseJSONPath)
	if err != nil {
		t.Fatalf("read universe json: %v", err)
	}
	var decoded UniverseSelection
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode universe json: %v", err)
	}
	if decoded.Mode != "volume_filter" || len(decoded.Selected) != 2 {
		t.Fatalf("unexpected universe json: %+v", decoded)
	}
}

func TestParseHyperliquidUniverse(t *testing.T) {
	body := []byte(`[
		{"universe":[{"name":"BTC"},{"name":"ETH"}]},
		[
			{"dayNtlVlm":"12000000","funding":"0.0001"},
			{"dayNtlVlm":"4500000","funding":"0.0002"}
		]
	]`)
	rows, err := parseHyperliquidUniverse(body)
	if err != nil {
		t.Fatalf("parse hyperliquid universe: %v", err)
	}
	if len(rows) != 2 || rows[0].CanonicalSymbol != "BTC" || rows[0].Volume24hUSD != 12000000 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestParseAsterUniverse(t *testing.T) {
	body := []byte(`[
		{"symbol":"BTCUSDT","quoteVolume":"12345678.9"},
		{"symbol":"ETHUSDT","quoteVolume":"9876543.2"}
	]`)
	rows, err := parseAsterUniverse(body)
	if err != nil {
		t.Fatalf("parse aster universe: %v", err)
	}
	if len(rows) != 2 || rows[0].CanonicalSymbol != "BTC" || rows[0].Volume24hUSD != 12345678.9 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestParseLighterExchangeMetricVolume(t *testing.T) {
	volume, ok := parseLighterExchangeMetricVolume([]byte(`{"data":[{"timestamp":1,"value":"7654321.5"}]}`))
	if !ok || volume != 7654321.5 {
		t.Fatalf("unexpected lighter volume: %v %t", volume, ok)
	}
}

func TestDefaultConfigUniversePathsUsePaperRoot(t *testing.T) {
	t.Setenv("DATA_ROOT", filepath.Join("C:", "tmp", "algo-data"))

	cfg := DefaultConfig()
	if cfg.UniverseCSVPath != filepath.Join("C:", "tmp", "algo-data", "paper", "universe.csv") {
		t.Fatalf("unexpected universe csv path: %s", cfg.UniverseCSVPath)
	}
	if cfg.UniverseJSONPath != filepath.Join("C:", "tmp", "algo-data", "paper", "universe.json") {
		t.Fatalf("unexpected universe json path: %s", cfg.UniverseJSONPath)
	}
}
