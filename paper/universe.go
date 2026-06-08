package paper

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"AlgoTrading2026/config"
	"AlgoTrading2026/symbols"
)

type UniverseEntry struct {
	Venue           string  `json:"venue"`
	Symbol          string  `json:"symbol"`
	CanonicalSymbol string  `json:"canonicalSymbol"`
	Volume24hUSD    float64 `json:"volume24hUsd"`
	Rank            int     `json:"rank"`
}

type UniverseSelection struct {
	Mode             string          `json:"mode"`
	Min24hVolumeUSD  float64         `json:"min24hVolumeUsd"`
	MaxSymbols       int             `json:"maxSymbols"`
	Selected         []UniverseEntry `json:"selected"`
	Notes            []string        `json:"notes,omitempty"`
}

type UniverseProvider func() ([]UniverseEntry, error)

func SelectUniverse(cfg Config) (UniverseSelection, error) {
	return selectUniverseWithProviders(cfg, defaultUniverseProviders())
}

func selectUniverseWithProviders(cfg Config, providers map[string]UniverseProvider) (UniverseSelection, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.UniverseMode))
	if mode == "" {
		mode = "manual"
	}
	selection := UniverseSelection{
		Mode:            mode,
		Min24hVolumeUSD: cfg.Min24hVolumeUSD,
		MaxSymbols:      cfg.MaxSymbols,
	}
	switch mode {
	case "manual":
		selection.Selected = buildManualUniverse(cfg)
	case "volume_filter":
		for _, venue := range cfg.Venues {
			provider, ok := providers[strings.ToLower(strings.TrimSpace(venue))]
			if !ok {
				selection.Notes = append(selection.Notes, fmt.Sprintf("universe provider unavailable for venue=%s", venue))
				continue
			}
			rows, err := provider()
			if err != nil {
				selection.Notes = append(selection.Notes, fmt.Sprintf("universe provider failed for venue=%s: %v", venue, err))
				continue
			}
			for _, row := range rows {
				if row.Volume24hUSD >= cfg.Min24hVolumeUSD {
					selection.Selected = append(selection.Selected, row)
				}
			}
		}
		sort.Slice(selection.Selected, func(i, j int) bool {
			if selection.Selected[i].Volume24hUSD == selection.Selected[j].Volume24hUSD {
				if selection.Selected[i].Venue == selection.Selected[j].Venue {
					return selection.Selected[i].Symbol < selection.Selected[j].Symbol
				}
				return selection.Selected[i].Venue < selection.Selected[j].Venue
			}
			return selection.Selected[i].Volume24hUSD > selection.Selected[j].Volume24hUSD
		})
		if cfg.MaxSymbols > 0 && len(selection.Selected) > cfg.MaxSymbols {
			selection.Selected = selection.Selected[:cfg.MaxSymbols]
		}
	default:
		return UniverseSelection{}, fmt.Errorf("unsupported PAPER_UNIVERSE_MODE %q", cfg.UniverseMode)
	}
	assignUniverseRanks(selection.Selected)
	if err := writeUniverseCSV(cfg.UniverseCSVPath, selection.Selected); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeUniverseJSON(cfg.UniverseJSONPath, selection); err != nil {
		return UniverseSelection{}, err
	}
	return selection, nil
}

func buildManualUniverse(cfg Config) []UniverseEntry {
	rows := make([]UniverseEntry, 0, len(cfg.Venues)*len(cfg.Symbols))
	seen := map[string]bool{}
	for _, venue := range cfg.Venues {
		for _, rawSymbol := range cfg.Symbols {
			symbol := universeVenueSymbol(venue, rawSymbol)
			if symbol == "" {
				continue
			}
			identity := symbols.NormalizeSymbol(venue, symbol)
			key := strings.ToLower(strings.TrimSpace(venue)) + ":" + identity.VenueSymbol
			if seen[key] {
				continue
			}
			seen[key] = true
			rows = append(rows, UniverseEntry{
				Venue:           strings.ToLower(strings.TrimSpace(venue)),
				Symbol:          identity.VenueSymbol,
				CanonicalSymbol: identity.CanonicalSymbol,
			})
		}
	}
	assignUniverseRanks(rows)
	return rows
}

func assignUniverseRanks(rows []UniverseEntry) {
	for i := range rows {
		rows[i].Rank = i + 1
	}
}

func writeUniverseCSV(path string, rows []UniverseEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"venue", "symbol", "canonical_symbol", "volume_24h_usd", "rank"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Venue,
			row.Symbol,
			row.CanonicalSymbol,
			strconv.FormatFloat(row.Volume24hUSD, 'f', -1, 64),
			strconv.Itoa(row.Rank),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeUniverseJSON(path string, selection UniverseSelection) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(selection, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func uniqueUniverseVenues(rows []UniverseEntry) int {
	seen := map[string]bool{}
	for _, row := range rows {
		seen[strings.ToLower(strings.TrimSpace(row.Venue))] = true
	}
	return len(seen)
}

func universeVenueSymbol(venue string, symbol string) string {
	switch strings.ToLower(strings.TrimSpace(venue)) {
	case "aster":
		return normalizePaperAsterSymbol(symbol)
	case "hyperliquid", "lighter":
		return normalizePaperPerpSymbol(symbol)
	default:
		return stringsUpperTrim(symbol)
	}
}

func defaultUniverseProviders() map[string]UniverseProvider {
	return map[string]UniverseProvider{
		"aster":       fetchAsterUniverse,
		"hyperliquid": fetchHyperliquidUniverse,
		"lighter":     fetchLighterUniverse,
	}
}

func fetchAsterUniverse() ([]UniverseEntry, error) {
	baseURL := "https://fapi.asterdex.com"
	if config.IsTestnet() {
		baseURL = "https://fapi.asterdex-testnet.com"
	}
	resp, err := http.Get(baseURL + "/fapi/v1/ticker/24hr")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("aster ticker 24hr bad status %d: %s", resp.StatusCode, string(body))
	}
	return parseAsterUniverse(body)
}

func parseAsterUniverse(body []byte) ([]UniverseEntry, error) {
	type asterTicker struct {
		Symbol      string `json:"symbol"`
		QuoteVolume string `json:"quoteVolume"`
	}
	var list []asterTicker
	if err := json.Unmarshal(body, &list); err != nil {
		var single asterTicker
		if errSingle := json.Unmarshal(body, &single); errSingle != nil {
			return nil, err
		}
		list = []asterTicker{single}
	}
	rows := make([]UniverseEntry, 0, len(list))
	for _, item := range list {
		identity := symbols.NormalizeSymbol("aster", item.Symbol)
		rows = append(rows, UniverseEntry{
			Venue:           "aster",
			Symbol:          identity.VenueSymbol,
			CanonicalSymbol: identity.CanonicalSymbol,
			Volume24hUSD:    parseUniverseFloat(item.QuoteVolume),
		})
	}
	return rows, nil
}

func fetchHyperliquidUniverse() ([]UniverseEntry, error) {
	payload := []byte(`{"type":"metaAndAssetCtxs"}`)
	baseURL := "https://api.hyperliquid.xyz"
	if config.IsTestnet() {
		baseURL = "https://api.hyperliquid-testnet.xyz"
	}
	resp, err := http.Post(baseURL+"/info", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("hyperliquid metaAndAssetCtxs bad status %d: %s", resp.StatusCode, string(body))
	}
	return parseHyperliquidUniverse(body)
}

func parseHyperliquidUniverse(body []byte) ([]UniverseEntry, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if len(raw) < 2 {
		return nil, fmt.Errorf("unexpected hyperliquid universe payload")
	}
	var meta struct {
		Universe []struct {
			Name string `json:"name"`
		} `json:"universe"`
	}
	if err := json.Unmarshal(raw[0], &meta); err != nil {
		return nil, err
	}
	var ctxs []map[string]any
	if err := json.Unmarshal(raw[1], &ctxs); err != nil {
		return nil, err
	}
	limit := len(meta.Universe)
	if len(ctxs) < limit {
		limit = len(ctxs)
	}
	rows := make([]UniverseEntry, 0, limit)
	for i := 0; i < limit; i++ {
		symbol := stringsUpperTrim(meta.Universe[i].Name)
		identity := symbols.NormalizeSymbol("hyperliquid", symbol)
		rows = append(rows, UniverseEntry{
			Venue:           "hyperliquid",
			Symbol:          identity.VenueSymbol,
			CanonicalSymbol: identity.CanonicalSymbol,
			Volume24hUSD: parseUniverseFloat(
				firstUniverseValue(ctxs[i], "dayNtlVlm", "dailyNtlVlm", "dayNtlVolume", "volume24h"),
			),
		})
	}
	return rows, nil
}

func fetchLighterUniverse() ([]UniverseEntry, error) {
	symbolsList := []string{"BTC", "ETH"}
	rows := make([]UniverseEntry, 0, len(symbolsList))
	var errs []string
	for _, symbol := range symbolsList {
		volume, err := fetchLighterExchangeMetricVolume(symbol)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", symbol, err))
			continue
		}
		identity := symbols.NormalizeSymbol("lighter", symbol)
		rows = append(rows, UniverseEntry{
			Venue:           "lighter",
			Symbol:          identity.VenueSymbol,
			CanonicalSymbol: identity.CanonicalSymbol,
			Volume24hUSD:    volume,
		})
	}
	if len(rows) == 0 && len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}
	return rows, nil
}

func fetchLighterExchangeMetricVolume(symbol string) (float64, error) {
	baseURL := "https://mainnet.zklighter.elliot.ai"
	if config.IsTestnet() {
		baseURL = "https://testnet.zklighter.elliot.ai"
	}
	params := url.Values{}
	params.Set("period", "d")
	params.Set("kind", "volume")
	params.Set("filter", "byMarket")
	params.Set("value", stringsUpperTrim(symbol))

	resp, err := http.Get(baseURL + "/api/v1/exchangeMetrics?" + params.Encode())
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("lighter exchangeMetrics bad status %d: %s", resp.StatusCode, string(body))
	}
	volume, ok := parseLighterExchangeMetricVolume(body)
	if !ok {
		return 0, fmt.Errorf("lighter exchangeMetrics volume not found for %s", symbol)
	}
	return volume, nil
}

func parseLighterExchangeMetricVolume(body []byte) (float64, bool) {
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return 0, false
	}
	return extractLighterMetric(raw)
}

func extractLighterMetric(value any) (float64, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"value", "volume", "daily_quote_token_volume"} {
			if raw, ok := typed[key]; ok {
				if parsed := parseUniverseFloat(raw); parsed > 0 {
					return parsed, true
				}
			}
		}
		for _, key := range []string{"data", "result", "metrics"} {
			if raw, ok := typed[key]; ok {
				if parsed, ok := extractLighterMetric(raw); ok {
					return parsed, true
				}
			}
		}
		for _, raw := range typed {
			if parsed, ok := extractLighterMetric(raw); ok {
				return parsed, true
			}
		}
	case []any:
		for i := len(typed) - 1; i >= 0; i-- {
			if parsed, ok := extractLighterMetric(typed[i]); ok {
				return parsed, true
			}
		}
	case float64, string, json.Number:
		if parsed := parseUniverseFloat(typed); parsed > 0 {
			return parsed, true
		}
	}
	return 0, false
}

func firstUniverseValue(m map[string]any, names ...string) any {
	for _, name := range names {
		if value, ok := m[name]; ok {
			return value
		}
	}
	return nil
}

func parseUniverseFloat(value any) float64 {
	switch typed := value.(type) {
	case nil:
		return 0
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed
	default:
		return 0
	}
}
