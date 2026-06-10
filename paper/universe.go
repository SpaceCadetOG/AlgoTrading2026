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
	MarketID        string  `json:"marketId,omitempty"`
	Active          bool    `json:"active"`
	Tradable        bool    `json:"tradable"`
	Volume24hUSD    float64 `json:"volume24hUsd"`
	VolumeKnown     bool    `json:"volumeKnown"`
	LastPrice       float64 `json:"lastPrice,omitempty"`
	SpreadPct       float64 `json:"spreadPct,omitempty"`
	Rank            int     `json:"rank"`
}

type UniverseSelection struct {
	Mode             string                     `json:"mode"`
	ManualSymbols    bool                       `json:"manualSymbols"`
	Min24hVolumeUSD  float64                    `json:"min24hVolumeUsd"`
	RequireVolume    bool                       `json:"requireVolume"`
	MaxSymbols       int                        `json:"maxSymbols"`
	Discovered       []UniverseEntry            `json:"discovered"`
	Qualified        []UniverseEntry            `json:"qualified"`
	Selected         []UniverseEntry            `json:"selected"`
	QualifiedSkipped []QualifiedNotSelected     `json:"qualifiedSkipped,omitempty"`
	Stats            []VenueDiscoveryStats      `json:"stats"`
	Diagnostics      []QualificationDiagnostics `json:"diagnostics,omitempty"`
	BiasDiagnostics  UniverseBiasDiagnostics    `json:"biasDiagnostics,omitempty"`
	Notes            []string                   `json:"notes,omitempty"`
}

type VenueDiscoveryStats struct {
	Venue      string `json:"venue"`
	Discovered int    `json:"discovered"`
	Qualified  int    `json:"qualified"`
	Selected   int    `json:"selected"`
	Failed     bool   `json:"failed"`
	Error      string `json:"error,omitempty"`
}

type UniverseProvider func() ([]UniverseEntry, error)

type QualifiedNotSelected struct {
	Venue           string  `json:"venue"`
	Symbol          string  `json:"symbol"`
	CanonicalSymbol string  `json:"canonicalSymbol"`
	Rank            int     `json:"rank"`
	Volume24hUSD    float64 `json:"volume24hUsd"`
	Reason          string  `json:"reason"`
}

type UniverseBiasDiagnostics struct {
	ManualSymbols              bool                       `json:"manualSymbols"`
	DiscoveredByVenue          map[string]int             `json:"discoveredByVenue"`
	QualifiedByVenue           map[string]int             `json:"qualifiedByVenue"`
	SelectedByVenue            map[string]int             `json:"selectedByVenue"`
	SelectedMajors             int                        `json:"selectedMajors"`
	SelectedNonMajors          int                        `json:"selectedNonMajors"`
	QualifiedMajors            int                        `json:"qualifiedMajors"`
	QualifiedNonMajors         int                        `json:"qualifiedNonMajors"`
	MajorSelectedPct           float64                    `json:"majorSelectedPct"`
	TopQualifiedNotSelected    []QualifiedNotSelected     `json:"topQualifiedNotSelected"`
	QualificationRejectReasons []QualificationDiagnostics `json:"qualificationRejectReasons,omitempty"`
	Policy                     string                     `json:"policy"`
}

type QualificationDiagnostics struct {
	Venue                      string         `json:"venue"`
	Discovered                 int            `json:"discovered"`
	Qualified                  int            `json:"qualified"`
	Rejected                   int            `json:"rejected"`
	Reasons                    map[string]int `json:"reasons"`
	MissingVolume              int            `json:"missingVolume"`
	LowVolume                  int            `json:"lowVolume"`
	MissingPrice               int            `json:"missingPrice"`
	SpreadTooWide              int            `json:"spreadTooWide"`
	MissingOrderbook           int            `json:"missingOrderbook"`
	MissingSymbolNormalization int            `json:"missingSymbolNormalization"`
	UnsupportedMetadata        int            `json:"unsupportedMetadata"`
	Inactive                   int            `json:"inactive"`
	Untradable                 int            `json:"untradable"`
	AdmittedUnknownVolume      int            `json:"admittedUnknownVolume"`
}

func SelectUniverse(cfg Config) (UniverseSelection, error) {
	return selectUniverseWithProviders(cfg, defaultUniverseProviders())
}

func selectUniverseWithProviders(cfg Config, providers map[string]UniverseProvider) (UniverseSelection, error) {
	mode := strings.ToLower(strings.TrimSpace(cfg.UniverseMode))
	if mode == "" {
		mode = "dynamic"
	}
	selection := UniverseSelection{
		Mode:            mode,
		Min24hVolumeUSD: cfg.Min24hVolumeUSD,
		RequireVolume:   cfg.RequireVolumeForQualification,
		MaxSymbols:      cfg.MaxSymbols,
	}
	switch mode {
	case "manual":
		selection.ManualSymbols = true
		selection.Selected = buildManualUniverse(cfg)
		selection.Discovered = append([]UniverseEntry(nil), selection.Selected...)
		selection.Qualified = append([]UniverseEntry(nil), selection.Selected...)
	case "volume_filter", "dynamic":
		for _, venue := range cfg.Venues {
			venueKey := strings.ToLower(strings.TrimSpace(venue))
			stat := VenueDiscoveryStats{Venue: venueKey}
			provider, ok := providers[venueKey]
			if !ok {
				stat.Failed = true
				stat.Error = "universe provider unavailable"
				selection.Stats = append(selection.Stats, stat)
				selection.Notes = append(selection.Notes, fmt.Sprintf("universe provider unavailable for venue=%s", venue))
				continue
			}
			rows, err := provider()
			if err != nil {
				stat.Failed = true
				stat.Error = err.Error()
				selection.Stats = append(selection.Stats, stat)
				selection.Notes = append(selection.Notes, fmt.Sprintf("universe provider failed for venue=%s: %v", venue, err))
				continue
			}
			selection.Discovered = append(selection.Discovered, rows...)
			stat.Discovered = len(rows)
			diagnostics := QualificationDiagnostics{Venue: venueKey, Discovered: len(rows), Reasons: map[string]int{}}
			for _, row := range rows {
				qualified, reasons := qualifyUniverseEntry(row, cfg)
				for _, reason := range reasons {
					diagnostics.Reasons[reason]++
					switch reason {
					case "inactive":
						diagnostics.Inactive++
					case "untradable":
						diagnostics.Untradable++
					case "missing_volume":
						diagnostics.MissingVolume++
					case "low_volume":
						diagnostics.LowVolume++
					case "missing_price":
						diagnostics.MissingPrice++
					case "spread_too_wide":
						diagnostics.SpreadTooWide++
					case "missing_orderbook":
						diagnostics.MissingOrderbook++
					case "missing_symbol_normalization":
						diagnostics.MissingSymbolNormalization++
					case "unsupported_metadata":
						diagnostics.UnsupportedMetadata++
					case "admitted_unknown_volume":
						diagnostics.AdmittedUnknownVolume++
					}
				}
				if qualified {
					selection.Qualified = append(selection.Qualified, row)
					stat.Qualified++
				}
			}
			diagnostics.Qualified = stat.Qualified
			diagnostics.Rejected = diagnostics.Discovered - diagnostics.Qualified
			selection.Diagnostics = append(selection.Diagnostics, diagnostics)
			selection.Stats = append(selection.Stats, stat)
		}
		selection.Selected = selectDynamicUniverse(selection.Qualified, cfg)
		selection.QualifiedSkipped = qualifiedNotSelected(selection.Qualified, selection.Selected)
		selectedByVenue := countUniverseByVenue(selection.Selected)
		for i := range selection.Stats {
			selection.Stats[i].Selected = selectedByVenue[selection.Stats[i].Venue]
		}
	default:
		return UniverseSelection{}, fmt.Errorf("unsupported PAPER_UNIVERSE_MODE %q", cfg.UniverseMode)
	}
	assignUniverseRanks(selection.Discovered)
	assignUniverseRanks(selection.Qualified)
	assignUniverseRanks(selection.Selected)
	if len(selection.Stats) == 0 {
		selection.Stats = statsFromManualSelection(selection.Selected)
	}
	selection.BiasDiagnostics = buildUniverseBiasDiagnostics(selection)
	if err := writeUniverseCSV(cfg.UniverseCSVPath, selection.Selected); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeUniverseJSON(cfg.UniverseJSONPath, selection); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeUniverseEntriesJSON(cfg.DiscoveredUniversePath, selection.Discovered); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeUniverseEntriesJSON(cfg.QualifiedUniversePath, selection.Qualified); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeUniverseEntriesJSON(cfg.SelectedUniversePath, selection.Selected); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeQualificationDiagnosticsJSON(cfg.QualificationDiagnosticsPath, selection.Diagnostics); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeQualifiedNotSelectedJSON(cfg.QualifiedNotSelectedPath, selection.QualifiedSkipped); err != nil {
		return UniverseSelection{}, err
	}
	if err := writeUniverseBiasDiagnosticsJSON(cfg.UniverseBiasDiagnosticsPath, selection.BiasDiagnostics); err != nil {
		return UniverseSelection{}, err
	}
	return selection, nil
}

func isQualifiedUniverseEntry(row UniverseEntry, cfg Config) bool {
	qualified, _ := qualifyUniverseEntry(row, cfg)
	return qualified
}

func qualifyUniverseEntry(row UniverseEntry, cfg Config) (bool, []string) {
	reasons := []string{}
	if !row.Active {
		reasons = append(reasons, "inactive")
	}
	if !row.Tradable {
		reasons = append(reasons, "untradable")
	}
	if strings.TrimSpace(row.Symbol) == "" || strings.TrimSpace(row.CanonicalSymbol) == "" {
		reasons = append(reasons, "missing_symbol_normalization")
	}
	if cfg.Min24hVolumeUSD > 0 {
		if row.Volume24hUSD <= 0 {
			if cfg.RequireVolumeForQualification {
				reasons = append(reasons, "missing_volume")
			} else {
				reasons = append(reasons, "admitted_unknown_volume")
			}
		} else if row.Volume24hUSD < cfg.Min24hVolumeUSD {
			reasons = append(reasons, "low_volume")
		}
	}
	if row.LastPrice < 0 {
		reasons = append(reasons, "missing_price")
	}
	if row.SpreadPct > 0.5 {
		reasons = append(reasons, "spread_too_wide")
	}
	return !hasBlockingQualificationReason(reasons), reasons
}

func hasBlockingQualificationReason(reasons []string) bool {
	for _, reason := range reasons {
		if reason != "admitted_unknown_volume" {
			return true
		}
	}
	return false
}

func sortUniverseEntries(rows []UniverseEntry) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Volume24hUSD <= 0 && rows[j].Volume24hUSD > 0 {
			return false
		}
		if rows[i].Volume24hUSD > 0 && rows[j].Volume24hUSD <= 0 {
			return true
		}
		if rows[i].Volume24hUSD == rows[j].Volume24hUSD {
			if rows[i].Venue == rows[j].Venue {
				return rows[i].Symbol < rows[j].Symbol
			}
			return rows[i].Venue < rows[j].Venue
		}
		return rows[i].Volume24hUSD > rows[j].Volume24hUSD
	})
}

func selectDynamicUniverse(qualified []UniverseEntry, cfg Config) []UniverseEntry {
	rows := append([]UniverseEntry(nil), qualified...)
	sortUniverseEntries(rows)
	if cfg.MaxSymbols <= 0 || len(rows) <= cfg.MaxSymbols {
		return rows
	}
	venues := sortedUniverseVenues(rows)
	quota := cfg.MaxSymbols / len(venues)
	if quota < 1 {
		quota = 1
	}
	selected := make([]UniverseEntry, 0, cfg.MaxSymbols)
	selectedKeys := map[string]bool{}
	byVenue := universeByVenue(rows)
	for _, venue := range venues {
		venueRows := byVenue[venue]
		limit := quota
		if limit > len(venueRows) {
			limit = len(venueRows)
		}
		for i := 0; i < limit && len(selected) < cfg.MaxSymbols; i++ {
			addUniverseSelection(&selected, selectedKeys, venueRows[i])
		}
	}
	for _, row := range rows {
		if len(selected) >= cfg.MaxSymbols {
			break
		}
		addUniverseSelection(&selected, selectedKeys, row)
	}
	sortUniverseEntries(selected)
	return selected
}

func addUniverseSelection(selected *[]UniverseEntry, selectedKeys map[string]bool, row UniverseEntry) {
	key := universeRouteKey(row)
	if selectedKeys[key] {
		return
	}
	selectedKeys[key] = true
	*selected = append(*selected, row)
}

func qualifiedNotSelected(qualified []UniverseEntry, selected []UniverseEntry) []QualifiedNotSelected {
	selectedKeys := map[string]bool{}
	for _, row := range selected {
		selectedKeys[universeRouteKey(row)] = true
	}
	rows := append([]UniverseEntry(nil), qualified...)
	sortUniverseEntries(rows)
	out := make([]QualifiedNotSelected, 0)
	for _, row := range rows {
		if selectedKeys[universeRouteKey(row)] {
			continue
		}
		out = append(out, QualifiedNotSelected{
			Venue:           row.Venue,
			Symbol:          row.Symbol,
			CanonicalSymbol: row.CanonicalSymbol,
			Rank:            row.Rank,
			Volume24hUSD:    row.Volume24hUSD,
			Reason:          "qualified_but_below_selection_cutoff",
		})
	}
	return out
}

func buildUniverseBiasDiagnostics(selection UniverseSelection) UniverseBiasDiagnostics {
	selectedMajors, selectedNonMajors := majorSplit(selection.Selected)
	qualifiedMajors, qualifiedNonMajors := majorSplit(selection.Qualified)
	majorPct := 0.0
	if len(selection.Selected) > 0 {
		majorPct = float64(selectedMajors) / float64(len(selection.Selected)) * 100
	}
	topSkipped := selection.QualifiedSkipped
	if len(topSkipped) > 25 {
		topSkipped = append([]QualifiedNotSelected(nil), topSkipped[:25]...)
	}
	return UniverseBiasDiagnostics{
		ManualSymbols:              selection.ManualSymbols,
		DiscoveredByVenue:          countUniverseByVenue(selection.Discovered),
		QualifiedByVenue:           countUniverseByVenue(selection.Qualified),
		SelectedByVenue:            countUniverseByVenue(selection.Selected),
		SelectedMajors:             selectedMajors,
		SelectedNonMajors:          selectedNonMajors,
		QualifiedMajors:            qualifiedMajors,
		QualifiedNonMajors:         qualifiedNonMajors,
		MajorSelectedPct:           majorPct,
		TopQualifiedNotSelected:    topSkipped,
		QualificationRejectReasons: selection.Diagnostics,
		Policy:                     "venue_balanced_quota_then_global_volume_fill",
	}
}

func majorSplit(rows []UniverseEntry) (int, int) {
	majors := 0
	nonMajors := 0
	for _, row := range rows {
		if isMajorCanonical(row.CanonicalSymbol) {
			majors++
		} else {
			nonMajors++
		}
	}
	return majors, nonMajors
}

func isMajorCanonical(symbol string) bool {
	switch strings.ToUpper(strings.TrimSpace(symbol)) {
	case "BTC", "ETH", "SOL":
		return true
	default:
		return false
	}
}

func sortedUniverseVenues(rows []UniverseEntry) []string {
	seen := map[string]bool{}
	for _, row := range rows {
		venue := strings.ToLower(strings.TrimSpace(row.Venue))
		if venue != "" {
			seen[venue] = true
		}
	}
	venues := make([]string, 0, len(seen))
	for venue := range seen {
		venues = append(venues, venue)
	}
	sort.Strings(venues)
	return venues
}

func universeByVenue(rows []UniverseEntry) map[string][]UniverseEntry {
	out := map[string][]UniverseEntry{}
	for _, row := range rows {
		venue := strings.ToLower(strings.TrimSpace(row.Venue))
		out[venue] = append(out[venue], row)
	}
	for venue := range out {
		sortUniverseEntries(out[venue])
	}
	return out
}

func universeRouteKey(row UniverseEntry) string {
	return strings.ToLower(strings.TrimSpace(row.Venue)) + ":" + strings.ToUpper(strings.TrimSpace(row.Symbol))
}

func countUniverseByVenue(rows []UniverseEntry) map[string]int {
	out := map[string]int{}
	for _, row := range rows {
		out[strings.ToLower(strings.TrimSpace(row.Venue))]++
	}
	return out
}

func statsFromManualSelection(rows []UniverseEntry) []VenueDiscoveryStats {
	counts := countUniverseByVenue(rows)
	venues := make([]string, 0, len(counts))
	for venue := range counts {
		venues = append(venues, venue)
	}
	sort.Strings(venues)
	stats := make([]VenueDiscoveryStats, 0, len(venues))
	for _, venue := range venues {
		count := counts[venue]
		stats = append(stats, VenueDiscoveryStats{Venue: venue, Discovered: count, Qualified: count, Selected: count})
	}
	return stats
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
				Active:          true,
				Tradable:        true,
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

	if err := writer.Write([]string{"venue", "symbol", "canonical_symbol", "market_id", "active", "tradable", "volume_24h_usd", "rank"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Venue,
			row.Symbol,
			row.CanonicalSymbol,
			row.MarketID,
			strconv.FormatBool(row.Active),
			strconv.FormatBool(row.Tradable),
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

func writeUniverseEntriesJSON(path string, rows []UniverseEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func writeQualificationDiagnosticsJSON(path string, diagnostics []QualificationDiagnostics) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func writeQualifiedNotSelectedJSON(path string, rows []QualifiedNotSelected) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0o644)
}

func writeUniverseBiasDiagnosticsJSON(path string, diagnostics UniverseBiasDiagnostics) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(diagnostics, "", "  ")
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
		LastPrice   string `json:"lastPrice"`
		BidPrice    string `json:"bidPrice"`
		AskPrice    string `json:"askPrice"`
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
			Active:          true,
			Tradable:        true,
			Volume24hUSD:    parseUniverseFloat(item.QuoteVolume),
			VolumeKnown:     parseUniverseFloat(item.QuoteVolume) > 0,
			LastPrice:       parseUniverseFloat(item.LastPrice),
			SpreadPct:       spreadPctFromBidAsk(parseUniverseFloat(item.BidPrice), parseUniverseFloat(item.AskPrice)),
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
		volume := parseUniverseFloat(
			firstUniverseValue(ctxs[i], "dayNtlVlm", "dailyNtlVlm", "dayNtlVolume", "volume24h"),
		)
		rows = append(rows, UniverseEntry{
			Venue:           "hyperliquid",
			Symbol:          identity.VenueSymbol,
			CanonicalSymbol: identity.CanonicalSymbol,
			Active:          true,
			Tradable:        true,
			Volume24hUSD:    volume,
			VolumeKnown:     volume > 0,
			LastPrice: parseUniverseFloat(
				firstUniverseValue(ctxs[i], "midPx", "markPx", "oraclePx", "prevDayPx"),
			),
		})
	}
	return rows, nil
}

func fetchLighterUniverse() ([]UniverseEntry, error) {
	resp, err := http.Get(lighterBaseURL() + "/api/v1/orderBooks?filter=all")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lighter orderBooks bad status %d: %s", resp.StatusCode, string(body))
	}
	rows, err := parseLighterUniverse(body)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("lighter orderBooks discovery returned no markets")
	}
	return rows, nil
}

func fetchLighterExchangeMetricVolume(symbol string) (float64, error) {
	params := url.Values{}
	params.Set("period", "d")
	params.Set("kind", "volume")
	params.Set("filter", "byMarket")
	params.Set("value", stringsUpperTrim(symbol))

	resp, err := http.Get(lighterBaseURL() + "/api/v1/exchangeMetrics?" + params.Encode())
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

func lighterBaseURL() string {
	if config.IsTestnet() {
		return "https://testnet.zklighter.elliot.ai"
	}
	return "https://mainnet.zklighter.elliot.ai"
}

func parseLighterUniverse(body []byte) ([]UniverseEntry, error) {
	var raw any
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	var rows []UniverseEntry
	collectLighterMarkets(raw, &rows)
	deduped := make([]UniverseEntry, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		key := strings.ToLower(row.Venue) + ":" + row.Symbol
		if seen[key] || row.Symbol == "" {
			continue
		}
		seen[key] = true
		deduped = append(deduped, row)
	}
	sortUniverseEntries(deduped)
	return deduped, nil
}

func collectLighterMarkets(value any, rows *[]UniverseEntry) {
	switch typed := value.(type) {
	case map[string]any:
		if row, ok := lighterMarketEntryFromMap(typed); ok {
			*rows = append(*rows, row)
		}
		for _, child := range typed {
			collectLighterMarkets(child, rows)
		}
	case []any:
		for _, child := range typed {
			collectLighterMarkets(child, rows)
		}
	}
}

func lighterMarketEntryFromMap(m map[string]any) (UniverseEntry, bool) {
	symbol := stringsUpperTrim(stringFromAny(firstUniverseValue(m, "symbol", "market", "name", "ticker", "base_asset", "baseAsset")))
	if symbol == "" {
		return UniverseEntry{}, false
	}
	if strings.Contains(symbol, "-") {
		symbol = strings.Split(symbol, "-")[0]
	}
	if strings.Contains(symbol, "/") {
		symbol = strings.Split(symbol, "/")[0]
	}
	identity := symbols.NormalizeSymbol("lighter", symbol)
	volume := parseUniverseFloat(firstUniverseValue(
		m,
		"daily_quote_token_volume",
		"volume24h",
		"quoteVolume",
		"dayNtlVlm",
		"volume",
	))
	row := UniverseEntry{
		Venue:           "lighter",
		Symbol:          identity.VenueSymbol,
		CanonicalSymbol: identity.CanonicalSymbol,
		MarketID:        stringFromAny(firstUniverseValue(m, "market_id", "marketIndex", "market_index", "id")),
		Active:          boolFromAny(firstUniverseValue(m, "active", "enabled", "tradable"), true),
		Tradable:        boolFromAny(firstUniverseValue(m, "tradable", "active", "enabled"), true),
		Volume24hUSD:    volume,
		VolumeKnown:     volume > 0,
		LastPrice: parseUniverseFloat(firstUniverseValue(
			m,
			"mid_price",
			"midPrice",
			"mark_price",
			"markPrice",
			"last_price",
			"lastPrice",
			"index_price",
			"indexPrice",
		)),
		SpreadPct: spreadPctFromBidAsk(
			parseUniverseFloat(firstUniverseValue(m, "best_bid", "bestBid", "bid_price", "bidPrice")),
			parseUniverseFloat(firstUniverseValue(m, "best_ask", "bestAsk", "ask_price", "askPrice")),
		),
	}
	return row, true
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

func spreadPctFromBidAsk(bid float64, ask float64) float64 {
	if bid <= 0 || ask <= 0 || ask < bid {
		return 0
	}
	mid := (bid + ask) / 2
	if mid <= 0 {
		return 0
	}
	return (ask - bid) / mid * 100
}

func stringFromAny(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

func boolFromAny(value any, fallback bool) bool {
	switch typed := value.(type) {
	case nil:
		return fallback
	case bool:
		return typed
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "1", "yes", "active", "enabled", "tradable":
			return true
		case "false", "0", "no", "inactive", "disabled":
			return false
		default:
			return fallback
		}
	case float64:
		return typed != 0
	default:
		return fallback
	}
}
