package paper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type StrategyReviewRecord struct {
	Timestamp       int64    `json:"timestamp"`
	TimeUTC         string   `json:"timeUtc"`
	Kind            string   `json:"kind"`
	Venue           string   `json:"venue,omitempty"`
	Symbol          string   `json:"symbol,omitempty"`
	CanonicalSymbol string   `json:"canonicalSymbol,omitempty"`
	Major           bool     `json:"major"`
	Strategy        string   `json:"strategy,omitempty"`
	Playbook        string   `json:"playbook,omitempty"`
	Side            string   `json:"side,omitempty"`
	Confidence      float64  `json:"confidence,omitempty"`
	Score           float64  `json:"score,omitempty"`
	Rank            int      `json:"rank,omitempty"`
	Decision        string   `json:"decision,omitempty"`
	Status          string   `json:"status,omitempty"`
	Reasons         []string `json:"reasons,omitempty"`
	Entry           float64  `json:"entry,omitempty"`
	Stop            float64  `json:"stop,omitempty"`
	TP1             float64  `json:"tp1,omitempty"`
	TP2             float64  `json:"tp2,omitempty"`
	TP3             float64  `json:"tp3,omitempty"`
	Provenance      string   `json:"provenance,omitempty"`
}

type OpportunityCoverage struct {
	Timestamp             int64          `json:"timestamp"`
	TimeUTC               string         `json:"timeUtc"`
	MajorSelected         int            `json:"majorSelected"`
	NonMajorSelected      int            `json:"nonMajorSelected"`
	MajorCandidates       int            `json:"majorCandidates"`
	NonMajorCandidates    int            `json:"nonMajorCandidates"`
	MajorApproved         int            `json:"majorApproved"`
	NonMajorApproved      int            `json:"nonMajorApproved"`
	MajorApprovalRate     float64        `json:"majorApprovalRate"`
	NonMajorApprovalRate  float64        `json:"nonMajorApprovalRate"`
	TopNonMajorSelected   map[string]int `json:"topNonMajorSelected,omitempty"`
	TopNonMajorRejected   map[string]int `json:"topNonMajorRejected,omitempty"`
	NonMajorRejectReasons map[string]int `json:"nonMajorRejectReasons,omitempty"`
}

type OpportunityCoverageVenue struct {
	Venue              string         `json:"venue"`
	MajorSelected      int            `json:"majorSelected"`
	NonMajorSelected   int            `json:"nonMajorSelected"`
	MajorCandidates    int            `json:"majorCandidates"`
	NonMajorCandidates int            `json:"nonMajorCandidates"`
	Approved           int            `json:"approved"`
	Rejected           int            `json:"rejected"`
	RejectReasons      map[string]int `json:"rejectReasons,omitempty"`
}

type OvernightStrategySummary struct {
	Timestamp             int64          `json:"timestamp"`
	TimeUTC               string         `json:"timeUtc"`
	TotalCycles           int            `json:"totalCycles"`
	SelectedSymbols       int            `json:"selectedSymbols"`
	Candidates            int            `json:"candidates"`
	Approved              int            `json:"approved"`
	Rejected              int            `json:"rejected"`
	MajorCandidates       int            `json:"majorCandidates"`
	NonMajorCandidates    int            `json:"nonMajorCandidates"`
	AverageConfidence     float64        `json:"averageConfidence"`
	NonMajorAvgConfidence float64        `json:"nonMajorAvgConfidence"`
	CandidatesByVenue     map[string]int `json:"candidatesByVenue"`
	CandidatesBySymbol    map[string]int `json:"candidatesBySymbol"`
	RejectReasons         map[string]int `json:"rejectReasons"`
	TopNonMajorSymbols    map[string]int `json:"topNonMajorSymbols"`
}

func AppendStrategyReviewEvent(cfg Config, kind string, event TelemetryEvent) error {
	record := StrategyReviewRecord{
		Timestamp:       event.Timestamp,
		TimeUTC:         unixMillisUTC(event.Timestamp),
		Kind:            kind,
		Venue:           event.Venue,
		Symbol:          event.Symbol,
		CanonicalSymbol: canonicalFromSymbol(event.Symbol),
		Major:           isMajorReviewSymbol(canonicalFromSymbol(event.Symbol)),
		Strategy:        event.Strategy,
		Playbook:        event.Playbook,
		Side:            event.Side,
		Confidence:      event.Confidence,
		Score:           event.Score,
		Decision:        event.Decision,
		Status:          event.Type,
		Reasons:         append([]string(nil), event.Reasons...),
		Entry:           event.Entry,
	}
	switch kind {
	case "candidate":
		return appendJSONL(cfg.StrategyReviewCandidatesPath, record)
	case "rejection":
		return appendJSONL(cfg.StrategyReviewRejectionsPath, record)
	default:
		return nil
	}
}

func AppendStrategyReviewSelected(cfg Config, rows []UniverseEntry, timestamp int64) error {
	for _, row := range rows {
		record := StrategyReviewRecord{
			Timestamp:       timestamp,
			TimeUTC:         unixMillisUTC(timestamp),
			Kind:            "selected",
			Venue:           row.Venue,
			Symbol:          row.Symbol,
			CanonicalSymbol: row.CanonicalSymbol,
			Major:           isMajorReviewSymbol(row.CanonicalSymbol),
			Rank:            row.Rank,
			Status:          "selected",
			Reasons:         []string{"selected_universe"},
		}
		if err := appendJSONL(cfg.StrategyReviewSelectedPath, record); err != nil {
			return err
		}
	}
	return nil
}

func AppendStrategyReviewPositions(cfg Config, positions []PaperPosition, timestamp int64) error {
	for _, position := range positions {
		record := StrategyReviewRecord{
			Timestamp:       timestamp,
			TimeUTC:         unixMillisUTC(timestamp),
			Kind:            "position",
			Venue:           position.Venue,
			Symbol:          position.Symbol,
			CanonicalSymbol: position.CanonicalSymbol,
			Major:           isMajorReviewSymbol(position.CanonicalSymbol),
			Strategy:        position.Strategy,
			Playbook:        position.Playbook,
			Side:            position.Side,
			Confidence:      position.Confidence,
			Score:           position.Score,
			Status:          "open",
			Entry:           position.EntryPrice,
			Stop:            position.Stop,
			TP1:             position.TP1,
			TP2:             position.TP2,
			TP3:             position.TP3,
			Provenance:      position.Provenance,
		}
		if err := appendJSONL(cfg.StrategyReviewPositionsPath, record); err != nil {
			return err
		}
	}
	return nil
}

func AppendStrategyReviewCycle(cfg Config, summary RuntimeSummary, timestamp int64) error {
	record := map[string]any{
		"timestamp":          timestamp,
		"timeUtc":            unixMillisUTC(timestamp),
		"mode":               summary.Mode,
		"executionMode":      summary.ExecutionMode,
		"universeMode":       summary.UniverseMode,
		"selectedSymbols":    summary.SelectedSymbols,
		"decisions":          summary.Decisions,
		"approved":           summary.Approved,
		"rejected":           summary.Rejected,
		"openPositions":      summary.OpenCount,
		"rejectReasons":      summary.RejectReasons,
		"candidateAdmission": summary.CandidateAdmission,
	}
	return appendJSONL(cfg.StrategyReviewCyclesPath, record)
}

func WriteOpportunityCoverageJSON(cfg Config, summary RuntimeSummary) error {
	now := time.Now().UTC()
	coverage, byVenue := BuildOpportunityCoverage(summary, now)
	if err := writeJSONFile(cfg.OpportunityCoveragePath, coverage); err != nil {
		return err
	}
	return writeJSONFile(cfg.OpportunityCoverageByVenuePath, byVenue)
}

func BuildOpportunityCoverage(summary RuntimeSummary, now time.Time) (OpportunityCoverage, []OpportunityCoverageVenue) {
	coverage := OpportunityCoverage{
		Timestamp:             now.UnixMilli(),
		TimeUTC:               now.Format(time.RFC3339),
		TopNonMajorSelected:   map[string]int{},
		TopNonMajorRejected:   map[string]int{},
		NonMajorRejectReasons: map[string]int{},
	}
	venueRows := map[string]*OpportunityCoverageVenue{}
	for _, row := range summary.SelectedUniverse {
		venue := venueCoverageRow(venueRows, row.Venue)
		if isMajorReviewSymbol(row.CanonicalSymbol) {
			coverage.MajorSelected++
			venue.MajorSelected++
		} else {
			coverage.NonMajorSelected++
			venue.NonMajorSelected++
			coverage.TopNonMajorSelected[row.Venue+":"+row.Symbol]++
		}
	}
	for _, row := range summary.CandidateCoverage {
		venue := venueCoverageRow(venueRows, row.Venue)
		if row.Major {
			coverage.MajorCandidates += row.Candidates
			coverage.MajorApproved += row.Approved
			venue.MajorCandidates += row.Candidates
		} else {
			coverage.NonMajorCandidates += row.Candidates
			coverage.NonMajorApproved += row.Approved
			venue.NonMajorCandidates += row.Candidates
			if row.Rejected > 0 {
				coverage.TopNonMajorRejected[row.Venue+":"+row.Symbol] += row.Rejected
			}
			for reason, count := range row.RejectReasons {
				coverage.NonMajorRejectReasons[reason] += count
			}
		}
		venue.Approved += row.Approved
		venue.Rejected += row.Rejected
		for reason, count := range row.RejectReasons {
			venue.RejectReasons[reason] += count
		}
	}
	if coverage.MajorCandidates > 0 {
		coverage.MajorApprovalRate = float64(coverage.MajorApproved) / float64(coverage.MajorCandidates)
	}
	if coverage.NonMajorCandidates > 0 {
		coverage.NonMajorApprovalRate = float64(coverage.NonMajorApproved) / float64(coverage.NonMajorCandidates)
	}
	venues := make([]string, 0, len(venueRows))
	for venue := range venueRows {
		venues = append(venues, venue)
	}
	sort.Strings(venues)
	out := make([]OpportunityCoverageVenue, 0, len(venues))
	for _, venue := range venues {
		out = append(out, *venueRows[venue])
	}
	return coverage, out
}

func WriteOvernightStrategySummary(cfg Config, summary RuntimeSummary) error {
	now := time.Now().UTC()
	payload := BuildOvernightStrategySummary(summary, now)
	if err := writeJSONFile(cfg.OvernightStrategySummaryJSONPath, payload); err != nil {
		return err
	}
	return os.WriteFile(cfg.OvernightStrategySummaryMarkdownPath, []byte(OvernightStrategySummaryMarkdown(payload)), 0o644)
}

func BuildOvernightStrategySummary(summary RuntimeSummary, now time.Time) OvernightStrategySummary {
	payload := OvernightStrategySummary{
		Timestamp:          now.UnixMilli(),
		TimeUTC:            now.Format(time.RFC3339),
		TotalCycles:        1,
		SelectedSymbols:    summary.SelectedSymbols,
		Candidates:         0,
		Approved:           summary.Approved,
		Rejected:           summary.Rejected,
		CandidatesByVenue:  map[string]int{},
		CandidatesBySymbol: map[string]int{},
		RejectReasons:      map[string]int{},
		TopNonMajorSymbols: map[string]int{},
	}
	confidenceSum := 0.0
	nonMajorConfidenceSum := 0.0
	for _, row := range summary.CandidateCoverage {
		payload.Candidates += row.Candidates
		confidenceSum += row.ConfidenceSum
		payload.CandidatesByVenue[row.Venue] += row.Candidates
		payload.CandidatesBySymbol[row.Venue+":"+row.Symbol] += row.Candidates
		if row.Major {
			payload.MajorCandidates += row.Candidates
		} else {
			payload.NonMajorCandidates += row.Candidates
			nonMajorConfidenceSum += row.ConfidenceSum
			payload.TopNonMajorSymbols[row.Venue+":"+row.Symbol] += row.Candidates
		}
		for reason, count := range row.RejectReasons {
			payload.RejectReasons[reason] += count
		}
	}
	if payload.Candidates > 0 {
		payload.AverageConfidence = confidenceSum / float64(payload.Candidates)
	}
	if payload.NonMajorCandidates > 0 {
		payload.NonMajorAvgConfidence = nonMajorConfidenceSum / float64(payload.NonMajorCandidates)
	}
	return payload
}

func OvernightStrategySummaryMarkdown(summary OvernightStrategySummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Overnight Strategy Summary\n\n")
	fmt.Fprintf(&b, "- Time UTC: %s\n", summary.TimeUTC)
	fmt.Fprintf(&b, "- Selected symbols: %d\n", summary.SelectedSymbols)
	fmt.Fprintf(&b, "- Candidates: %d\n", summary.Candidates)
	fmt.Fprintf(&b, "- Approved: %d\n", summary.Approved)
	fmt.Fprintf(&b, "- Rejected: %d\n", summary.Rejected)
	fmt.Fprintf(&b, "- Major candidates: %d\n", summary.MajorCandidates)
	fmt.Fprintf(&b, "- Non-major candidates: %d\n", summary.NonMajorCandidates)
	fmt.Fprintf(&b, "- Average confidence: %.4f\n", summary.AverageConfidence)
	fmt.Fprintf(&b, "- Non-major average confidence: %.4f\n\n", summary.NonMajorAvgConfidence)
	b.WriteString("## Reject Reasons\n\n")
	for _, key := range sortedMapKeys(summary.RejectReasons) {
		fmt.Fprintf(&b, "- %s: %d\n", key, summary.RejectReasons[key])
	}
	b.WriteString("\n## Top Non-Major Symbols\n\n")
	for _, key := range sortedMapKeys(summary.TopNonMajorSymbols) {
		fmt.Fprintf(&b, "- %s: %d\n", key, summary.TopNonMajorSymbols[key])
	}
	return b.String()
}

func appendJSONL(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = file.Write(append(body, '\n'))
	return err
}

func venueCoverageRow(rows map[string]*OpportunityCoverageVenue, venue string) *OpportunityCoverageVenue {
	venue = strings.ToLower(strings.TrimSpace(venue))
	if rows[venue] == nil {
		rows[venue] = &OpportunityCoverageVenue{Venue: venue, RejectReasons: map[string]int{}}
	}
	return rows[venue]
}

func sortedMapKeys(rows map[string]int) []string {
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if rows[keys[i]] == rows[keys[j]] {
			return keys[i] < keys[j]
		}
		return rows[keys[i]] > rows[keys[j]]
	})
	return keys
}

func unixMillisUTC(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.UnixMilli(ts).UTC().Format(time.RFC3339)
}

func isMajorReviewSymbol(symbol string) bool {
	switch strings.ToUpper(strings.TrimSpace(symbol)) {
	case "BTC", "ETH", "SOL":
		return true
	default:
		return false
	}
}
