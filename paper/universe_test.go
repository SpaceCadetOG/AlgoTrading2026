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
	if !selection.ManualSymbols || len(selection.Discovered) != 6 || len(selection.Qualified) != 6 {
		t.Fatalf("expected manual universe to mark selected as discovered/qualified: %+v", selection)
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
	cfg.UniverseMode = "dynamic"
	cfg.Min24hVolumeUSD = 5000000
	cfg.MaxSymbols = 2
	cfg.Venues = []string{"aster", "hyperliquid", "lighter"}

	providers := map[string]UniverseProvider{
		"aster": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC", Active: true, Tradable: true, Volume24hUSD: 9000000},
				{Venue: "aster", Symbol: "ETHUSDT", CanonicalSymbol: "ETH", Active: true, Tradable: true, Volume24hUSD: 1000000},
			}, nil
		},
		"hyperliquid": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "hyperliquid", Symbol: "ETH", CanonicalSymbol: "ETH", Active: true, Tradable: true, Volume24hUSD: 12000000},
			}, nil
		},
		"lighter": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "lighter", Symbol: "BTC", CanonicalSymbol: "BTC", Active: true, Tradable: true, Volume24hUSD: 7000000},
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
	if len(selection.Discovered) != 4 || len(selection.Qualified) != 3 {
		t.Fatalf("unexpected discovered/qualified counts: %+v", selection)
	}
	if selection.Stats[0].Venue != "aster" || selection.Stats[0].Discovered != 2 || selection.Stats[0].Qualified != 1 {
		t.Fatalf("unexpected venue stats: %+v", selection.Stats)
	}
	body, err := os.ReadFile(cfg.UniverseJSONPath)
	if err != nil {
		t.Fatalf("read universe json: %v", err)
	}
	var decoded UniverseSelection
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode universe json: %v", err)
	}
	if decoded.Mode != "dynamic" || len(decoded.Selected) != 2 {
		t.Fatalf("unexpected universe json: %+v", decoded)
	}
	for _, path := range []string{cfg.DiscoveredUniversePath, cfg.QualifiedUniversePath, cfg.SelectedUniversePath} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read universe export %s: %v", path, err)
		}
		if len(body) == 0 {
			t.Fatalf("expected universe export content in %s", path)
		}
	}
}

func TestDynamicUniversePreservesVenueSpecificSymbolCollisions(t *testing.T) {
	cfg := tempConfig(t)
	cfg.UniverseMode = "dynamic"
	cfg.Min24hVolumeUSD = 1
	cfg.MaxSymbols = 10
	cfg.Venues = []string{"aster", "hyperliquid", "lighter"}
	providers := map[string]UniverseProvider{
		"aster": func() ([]UniverseEntry, error) {
			return []UniverseEntry{{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC", Active: true, Tradable: true, Volume24hUSD: 3}}, nil
		},
		"hyperliquid": func() ([]UniverseEntry, error) {
			return []UniverseEntry{{Venue: "hyperliquid", Symbol: "BTC", CanonicalSymbol: "BTC", Active: true, Tradable: true, Volume24hUSD: 2}}, nil
		},
		"lighter": func() ([]UniverseEntry, error) {
			return []UniverseEntry{{Venue: "lighter", Symbol: "BTC", CanonicalSymbol: "BTC", Active: true, Tradable: true, Volume24hUSD: 1}}, nil
		},
	}
	selection, err := selectUniverseWithProviders(cfg, providers)
	if err != nil {
		t.Fatalf("select dynamic universe: %v", err)
	}
	if len(selection.Selected) != 3 {
		t.Fatalf("expected all venue-specific BTC markets, got %+v", selection.Selected)
	}
	seen := map[string]bool{}
	for _, row := range selection.Selected {
		seen[row.Venue+":"+row.Symbol] = true
	}
	for _, key := range []string{"aster:BTCUSDT", "hyperliquid:BTC", "lighter:BTC"} {
		if !seen[key] {
			t.Fatalf("missing venue-specific market %s in %+v", key, selection.Selected)
		}
	}
}

func TestDynamicUniverseRecordsProviderFailureWithoutPlaceholders(t *testing.T) {
	cfg := tempConfig(t)
	cfg.UniverseMode = "dynamic"
	cfg.Min24hVolumeUSD = 1
	cfg.Venues = []string{"aster", "lighter"}
	providers := map[string]UniverseProvider{
		"aster": func() ([]UniverseEntry, error) {
			return []UniverseEntry{{Venue: "aster", Symbol: "BTCUSDT", CanonicalSymbol: "BTC", Active: true, Tradable: true, Volume24hUSD: 10}}, nil
		},
		"lighter": func() ([]UniverseEntry, error) {
			return nil, os.ErrNotExist
		},
	}
	selection, err := selectUniverseWithProviders(cfg, providers)
	if err != nil {
		t.Fatalf("select dynamic universe: %v", err)
	}
	if len(selection.Selected) != 1 || selection.Selected[0].Venue != "aster" {
		t.Fatalf("expected runtime to continue without fake lighter placeholders: %+v", selection.Selected)
	}
	if len(selection.Stats) != 2 || !selection.Stats[1].Failed {
		t.Fatalf("expected lighter failure stats, got %+v", selection.Stats)
	}
}

func TestDynamicUniverseDiagnosticsExplainZeroQualified(t *testing.T) {
	cfg := tempConfig(t)
	cfg.UniverseMode = "dynamic"
	cfg.Min24hVolumeUSD = 1000
	cfg.Venues = []string{"hyperliquid"}

	selection, err := selectUniverseWithProviders(cfg, map[string]UniverseProvider{
		"hyperliquid": func() ([]UniverseEntry, error) {
			return []UniverseEntry{
				{Venue: "hyperliquid", Symbol: "BTC", CanonicalSymbol: "BTC", Active: true, Tradable: true},
				{Venue: "hyperliquid", Symbol: "ETH", CanonicalSymbol: "ETH", Active: true, Tradable: true, Volume24hUSD: 50},
			}, nil
		},
	})
	if err != nil {
		t.Fatalf("select universe: %v", err)
	}
	if len(selection.Qualified) != 0 || len(selection.Diagnostics) != 1 {
		t.Fatalf("expected zero qualified with diagnostics, got %+v", selection)
	}
	diagnostics := selection.Diagnostics[0]
	if diagnostics.Discovered != 2 || diagnostics.Qualified != 0 || diagnostics.MissingVolume != 1 || diagnostics.LowVolume != 1 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	body, err := os.ReadFile(cfg.QualificationDiagnosticsPath)
	if err != nil {
		t.Fatalf("read diagnostics export: %v", err)
	}
	if !strings.Contains(string(body), "missing_volume") || !strings.Contains(string(body), "low_volume") {
		t.Fatalf("expected persisted diagnostics reasons, got %s", string(body))
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
	if !rows[0].Active || !rows[0].Tradable {
		t.Fatalf("expected hyperliquid rows active/tradable: %+v", rows[0])
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

func TestParseLighterUniverse(t *testing.T) {
	body := []byte(`{"order_books":[{"market_id":1,"symbol":"BTC","daily_quote_token_volume":"1000000","active":true},{"market_id":2,"symbol":"ETH","daily_quote_token_volume":"500000","active":true}]}`)
	rows, err := parseLighterUniverse(body)
	if err != nil {
		t.Fatalf("parse lighter universe: %v", err)
	}
	if len(rows) != 2 || rows[0].Venue != "lighter" || rows[0].MarketID == "" || !rows[0].Tradable {
		t.Fatalf("unexpected lighter rows: %+v", rows)
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
	if cfg.DiscoveredUniversePath != filepath.Join("C:", "tmp", "algo-data", "paper", "discovered_universe.json") {
		t.Fatalf("unexpected discovered universe path: %s", cfg.DiscoveredUniversePath)
	}
}
