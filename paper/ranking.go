package paper

import (
	"math"
	"sort"
	"strings"
)

type RankingRow struct {
	Rank                    int                `json:"rank"`
	Venue                   string             `json:"venue"`
	Symbol                  string             `json:"symbol"`
	CanonicalSymbol         string             `json:"canonicalSymbol"`
	Side                    string             `json:"side"`
	State                   string             `json:"state"`
	Score                   float64            `json:"score"`
	RankReason              string             `json:"rankReason"`
	SourceUniverseRank      int                `json:"sourceUniverseRank"`
	ExecutableCandidateRank int                `json:"executableCandidateRank,omitempty"`
	Candidates              int                `json:"candidates"`
	Approved                int                `json:"approved"`
	Rejected                int                `json:"rejected"`
	Open                    bool               `json:"open"`
	Mark                    float64            `json:"mark"`
	Volume24hUSD            float64            `json:"volume24hUsd"`
	VolumeKnown             bool               `json:"volumeKnown"`
	RejectReasons           map[string]int     `json:"rejectReasons,omitempty"`
	Components              RankingComponents  `json:"components"`
	Penalties               map[string]float64 `json:"penalties,omitempty"`
	Bonuses                 map[string]float64 `json:"bonuses,omitempty"`
	DuplicateCount          int                `json:"duplicateCount"`
}

type RankingComponents struct {
	Confidence        float64 `json:"confidence"`
	CandidateScore    float64 `json:"candidateScore"`
	PlaybookQuality   float64 `json:"playbookQuality"`
	RiskReward        float64 `json:"riskReward"`
	Liquidity         float64 `json:"liquidity"`
	OrderBookQuality  float64 `json:"orderBookQuality"`
	VenueHealth       float64 `json:"venueHealth"`
	BaseWeightedScore float64 `json:"baseWeightedScore"`
}

type RankingDiagnostics struct {
	Formula           string       `json:"formula"`
	Rows              []RankingRow `json:"rows"`
	OperatorKnobs     []ConfigKnob `json:"operatorKnobs"`
	DeprecatedAliases []ConfigKnob `json:"deprecatedAliases"`
	DebugDevKnobs     []ConfigKnob `json:"debugDevKnobs"`
	LiveSafetyKnobs   []ConfigKnob `json:"liveSafetyKnobs"`
	CredentialKnobs   []ConfigKnob `json:"credentialKnobs"`
	MovedDefaultKnobs []ConfigKnob `json:"movedDefaultKnobs"`
	DeletedKnobs      []ConfigKnob `json:"deletedKnobs"`
}

func BuildRanking(summary RuntimeSummary, state EngineState) []RankingRow {
	coverage := map[string]CandidateCoverageRow{}
	for _, row := range summary.CandidateCoverage {
		coverage[rankingRouteKey(row.Venue, row.Symbol)] = row
	}
	openPositions := map[string]PaperPosition{}
	for _, pos := range state.OpenPositions {
		openPositions[rankingRouteKey(pos.Venue, pos.Symbol)] = pos
	}
	decisions := bestRecentDecisionsByRoute(state.RecentDecisions)
	venueCounts := countUniverseByVenue(summary.SelectedUniverse)
	rows := make([]RankingRow, 0, len(summary.SelectedUniverse))
	for _, entry := range summary.SelectedUniverse {
		key := rankingRouteKey(entry.Venue, entry.Symbol)
		cov := coverage[key]
		decision := decisions[key]
		pos, isOpen := openPositions[key]
		row := buildRankingRow(entry, cov, decision, isOpen, pos, venueCounts)
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].State != rows[j].State {
			return rankingStatePriority(rows[i].State) < rankingStatePriority(rows[j].State)
		}
		if rows[i].Score != rows[j].Score {
			return rows[i].Score > rows[j].Score
		}
		if rows[i].Volume24hUSD != rows[j].Volume24hUSD {
			return rows[i].Volume24hUSD > rows[j].Volume24hUSD
		}
		if rows[i].Venue == rows[j].Venue {
			return rows[i].Symbol < rows[j].Symbol
		}
		return rows[i].Venue < rows[j].Venue
	})
	executableRank := 0
	for i := range rows {
		rows[i].Rank = i + 1
		if rows[i].Candidates > 0 && rows[i].State != "unavailable" {
			executableRank++
			rows[i].ExecutableCandidateRank = executableRank
		}
	}
	return rows
}

func buildRankingRow(entry UniverseEntry, cov CandidateCoverageRow, decision TelemetryEvent, open bool, pos PaperPosition, venueCounts map[string]int) RankingRow {
	confidence := clamp01(firstPositive(decision.Confidence, cov.AverageConfidence))
	candidateScore := clamp01(firstPositive(decision.Score, cov.AverageConfidence))
	playbookQuality := playbookQualityScore(decision.Playbook, cov.Candidates)
	rr := riskRewardScore(decision)
	liquidity := liquidityScore(entry.Volume24hUSD)
	bookQuality := orderBookQualityScore(cov)
	venueHealth := 1.0

	// Ranking formula: confidence 30%, candidate score 20%, playbook quality 15%,
	// risk/reward 15%, liquidity 10%, orderbook quality 5%, venue health 5%.
	base := confidence*30 + candidateScore*20 + playbookQuality*15 + rr*15 + liquidity*10 + bookQuality*5 + venueHealth*5
	penalties := map[string]float64{}
	bonuses := map[string]float64{}
	state := rankingState(cov, decision, open)
	if open {
		bonuses["open_position_management"] = 12
	}
	if cov.Approved > 0 || decision.Decision == "approved" {
		bonuses["approved_actionable"] = 15
	}
	if !entry.VolumeKnown {
		penalties["unknown_volume"] = 8
	}
	if !cov.SnapshotFetched {
		penalties["no_orderbook"] = 35
	}
	if cov.RejectReasons["symbol_cooldown"] > 0 || cov.RejectReasons["max_trades_per_symbol_per_day"] > 0 {
		penalties["cooldown_or_daily_limit"] = 18
	}
	if cov.RejectReasons["max_open_positions"] > 0 || cov.RejectReasons["position_policy_duplicate"] > 0 {
		penalties["portfolio_blocked"] = 16
	}
	if cov.RejectReasons["orderbook_fetch_failed"] > 0 {
		penalties["market_data_unavailable"] = 28
	}
	if cov.RejectReasons["deduped_out"] > 0 {
		penalties["duplicate_candidates"] = math.Min(10, float64(cov.RejectReasons["deduped_out"]))
	}
	if venueCounts[strings.ToLower(entry.Venue)] <= 2 {
		bonuses["venue_diversity"] = 2
	}
	score := base
	for _, value := range bonuses {
		score += value
	}
	for _, value := range penalties {
		score -= value
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	side := "-"
	if open {
		side = strings.ToUpper(pos.Side)
	} else if decision.Side != "" {
		side = strings.ToUpper(decision.Side)
	}
	return RankingRow{
		Venue:              strings.ToLower(entry.Venue),
		Symbol:             strings.ToUpper(entry.Symbol),
		CanonicalSymbol:    strings.ToUpper(entry.CanonicalSymbol),
		Side:               side,
		State:              state,
		Score:              score,
		RankReason:         rankingReason(state, cov),
		SourceUniverseRank: entry.Rank,
		Candidates:         cov.Candidates,
		Approved:           cov.Approved,
		Rejected:           cov.Rejected,
		Open:               open,
		Mark:               entry.LastPrice,
		Volume24hUSD:       entry.Volume24hUSD,
		VolumeKnown:        entry.VolumeKnown,
		RejectReasons:      cov.RejectReasons,
		Components: RankingComponents{
			Confidence:        confidence,
			CandidateScore:    candidateScore,
			PlaybookQuality:   playbookQuality,
			RiskReward:        rr,
			Liquidity:         liquidity,
			OrderBookQuality:  bookQuality,
			VenueHealth:       venueHealth,
			BaseWeightedScore: base,
		},
		Penalties:      compactAdjustments(penalties),
		Bonuses:        compactAdjustments(bonuses),
		DuplicateCount: cov.RejectReasons["deduped_out"],
	}
}

func WriteRankingJSON(cfg Config, summary RuntimeSummary, state EngineState) error {
	rows := BuildRanking(summary, state)
	if err := writeJSONFile(cfg.RankingPath, rows); err != nil {
		return err
	}
	return writeJSONFile(cfg.RankingDiagnosticsPath, RankingDiagnostics{
		Formula:           "30% confidence + 20% candidate score + 15% playbook quality + 15% risk/reward + 10% liquidity + 5% orderbook quality + 5% venue health, then actionability bonuses and blocker penalties",
		Rows:              rows,
		OperatorKnobs:     ConfigKnobsByClass("operator"),
		DeprecatedAliases: ConfigKnobsByClass("deprecated_alias"),
		DebugDevKnobs:     ConfigKnobsByClass("debug_dev"),
		LiveSafetyKnobs:   ConfigKnobsByClass("live_safety"),
		CredentialKnobs:   ConfigKnobsByClass("credential"),
		MovedDefaultKnobs: ConfigKnobsByClass("code_default"),
		DeletedKnobs:      ConfigKnobsByClass("delete"),
	})
}

func bestRecentDecisionsByRoute(events []TelemetryEvent) map[string]TelemetryEvent {
	out := map[string]TelemetryEvent{}
	for _, event := range events {
		if event.Type != "decision" {
			continue
		}
		key := rankingRouteKey(event.Venue, event.Symbol)
		current, ok := out[key]
		if !ok || eventBetterForRanking(event, current) {
			out[key] = event
		}
	}
	return out
}

func eventBetterForRanking(a TelemetryEvent, b TelemetryEvent) bool {
	if a.Decision != b.Decision {
		return a.Decision == "approved"
	}
	if a.Confidence != b.Confidence {
		return a.Confidence > b.Confidence
	}
	return a.Score > b.Score
}

func rankingState(cov CandidateCoverageRow, decision TelemetryEvent, open bool) string {
	if open {
		return "open"
	}
	if !cov.SnapshotFetched {
		return "unavailable"
	}
	if cov.Approved > 0 || decision.Decision == "approved" {
		return "approved"
	}
	if cov.RejectReasons["symbol_cooldown"] > 0 || cov.RejectReasons["max_trades_per_symbol_per_day"] > 0 {
		return "cooldown"
	}
	if cov.RejectReasons["max_open_positions"] > 0 || cov.RejectReasons["position_policy_duplicate"] > 0 {
		return "blocked"
	}
	if cov.Candidates > 0 {
		return "rejected"
	}
	return "unavailable"
}

func rankingStatePriority(state string) int {
	switch state {
	case "open":
		return 0
	case "approved":
		return 1
	case "rejected":
		return 2
	case "cooldown":
		return 3
	case "blocked":
		return 4
	default:
		return 5
	}
}

func rankingReason(state string, cov CandidateCoverageRow) string {
	if state == "approved" || state == "open" {
		return state
	}
	reason := ""
	count := 0
	for key, value := range cov.RejectReasons {
		if value > count || (value == count && key < reason) {
			reason = key
			count = value
		}
	}
	if reason == "" {
		return state
	}
	return reason
}

func playbookQualityScore(playbook string, candidates int) float64 {
	if strings.TrimSpace(playbook) == "" && candidates == 0 {
		return 0.35
	}
	if strings.TrimSpace(playbook) == "" {
		return 0.65
	}
	return 0.8
}

func riskRewardScore(event TelemetryEvent) float64 {
	if event.StopDistance <= 0 || event.Entry <= 0 {
		return 0.55
	}
	return 0.7
}

func liquidityScore(volume float64) float64 {
	switch {
	case volume >= 1_000_000_000:
		return 1
	case volume >= 100_000_000:
		return 0.9
	case volume >= 25_000_000:
		return 0.75
	case volume >= 5_000_000:
		return 0.6
	case volume > 0:
		return 0.35
	default:
		return 0.2
	}
}

func orderBookQualityScore(cov CandidateCoverageRow) float64 {
	if !cov.SnapshotFetched {
		return 0
	}
	if cov.RejectReasons["orderbook_fetch_failed"] > 0 {
		return 0.2
	}
	return 0.8
}

func firstPositive(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func compactAdjustments(values map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for key, value := range values {
		if value != 0 {
			out[key] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func rankingRouteKey(venue string, symbol string) string {
	return strings.ToLower(strings.TrimSpace(venue)) + ":" + strings.ToUpper(strings.TrimSpace(symbol))
}
