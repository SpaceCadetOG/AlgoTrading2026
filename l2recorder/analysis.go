package l2recorder

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

type DatasetQuality struct {
	TotalRows   int     `json:"totalRows"`
	ValidRows   int     `json:"validRows"`
	InvalidRows int     `json:"invalidRows"`
	ErrorRows   int     `json:"errorRows"`
	ValidPct    float64 `json:"validPct"`
	ErrorPct    float64 `json:"errorPct"`
}

type VenueSummary struct {
	Venue                string  `json:"venue"`
	SnapshotCount        int     `json:"snapshotCount"`
	ValidSnapshotCount   int     `json:"validSnapshotCount"`
	InvalidSnapshotCount int     `json:"invalidSnapshotCount"`
	AverageSpreadPct     float64 `json:"averageSpreadPct"`
	MinSpreadPct         float64 `json:"minSpreadPct"`
	MaxSpreadPct         float64 `json:"maxSpreadPct"`
	AverageImbalance1Pct float64 `json:"averageImbalance1Pct"`
	MinImbalance1Pct     float64 `json:"minImbalance1Pct"`
	MaxImbalance1Pct     float64 `json:"maxImbalance1Pct"`
	AverageBidDepth1Pct  float64 `json:"averageBidDepth1Pct"`
	AverageAskDepth1Pct  float64 `json:"averageAskDepth1Pct"`
	AverageMid           float64 `json:"averageMid"`
}

type CrossVenueSummary struct {
	AverageMidDifference    float64 `json:"averageMidDifference"`
	MaxMidDifference        float64 `json:"maxMidDifference"`
	AverageSpreadDifference float64 `json:"averageSpreadDifference"`
	WidestSpreadVenue       string  `json:"widestSpreadVenue"`
	TightestSpreadVenue     string  `json:"tightestSpreadVenue"`
	DeepestLiquidityVenue   string  `json:"deepestLiquidityVenue"`
	ComparableGroups        int     `json:"comparableGroups"`
}

type LiquidityRankings struct {
	AverageBidDepth    []VenueRanking `json:"averageBidDepth"`
	AverageAskDepth    []VenueRanking `json:"averageAskDepth"`
	CombinedDepth      []VenueRanking `json:"combinedDepth"`
	TightestSpreads    []VenueRanking `json:"tightestSpreads"`
	HighestBidPressure []VenueRanking `json:"highestBidPressure"`
	HighestAskPressure []VenueRanking `json:"highestAskPressure"`
}

type VenueRanking struct {
	Venue string  `json:"venue"`
	Value float64 `json:"value"`
}

type SnapshotAnalysis struct {
	DatasetQuality   DatasetQuality    `json:"datasetQuality"`
	VenueSummaries   []VenueSummary    `json:"venueSummaries"`
	CrossVenue       CrossVenueSummary `json:"crossVenueSummary"`
	LiquidityRanking LiquidityRankings `json:"liquidityRankings"`
	Notes            []string          `json:"notes"`
}

func ReadSnapshotRowsCSV(path string) ([]SnapshotRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) <= 1 {
		return nil, nil
	}

	rows := make([]SnapshotRow, 0, len(records)-1)
	for i, record := range records[1:] {
		row, err := parseSnapshotRow(record)
		if err != nil {
			return nil, fmt.Errorf("parse row %d: %w", i+2, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func AnalyzeSnapshotRows(rows []SnapshotRow) SnapshotAnalysis {
	quality := datasetQuality(rows)
	venueSummaries := venueSummaries(rows)
	crossVenue := crossVenueSummary(rows, venueSummaries)
	rankings := liquidityRankings(venueSummaries)

	notes := []string{
		"Historical L2 recorder Phase 1 samples snapshots at fixed intervals; it is not tick-by-tick book replay.",
		"Cross-venue comparison groups rows by recorder round order because venue timestamps are collected sequentially, not simultaneously.",
		"CSV storage is local and append-only; larger recordings should manage file rotation and metadata.",
	}
	if quality.ErrorRows > 0 {
		notes = append(notes, "Some rows contain venue fetch errors; inspect the error column before using the dataset for research.")
	}
	if quality.ValidPct < 95 {
		notes = append(notes, "Valid row percentage is below 95%; collect a larger sample or inspect venue reliability before extending research.")
	}

	return SnapshotAnalysis{
		DatasetQuality:   quality,
		VenueSummaries:   venueSummaries,
		CrossVenue:       crossVenue,
		LiquidityRanking: rankings,
		Notes:            notes,
	}
}

func AnalyzeSnapshotCSV(path string) (SnapshotAnalysis, error) {
	rows, err := ReadSnapshotRowsCSV(path)
	if err != nil {
		return SnapshotAnalysis{}, err
	}
	return AnalyzeSnapshotRows(rows), nil
}

func parseSnapshotRow(record []string) (SnapshotRow, error) {
	if len(record) < 13 {
		return SnapshotRow{}, fmt.Errorf("expected 13 columns, got %d", len(record))
	}
	valid, err := strconv.ParseBool(record[11])
	if err != nil {
		return SnapshotRow{}, err
	}
	return SnapshotRow{
		Timestamp:     parseInt(record[0]),
		Venue:         record[1],
		Symbol:        record[2],
		BestBid:       parseFloat(record[3]),
		BestAsk:       parseFloat(record[4]),
		Spread:        parseFloat(record[5]),
		SpreadPct:     parseFloat(record[6]),
		Mid:           parseFloat(record[7]),
		BidDepth1Pct:  parseFloat(record[8]),
		AskDepth1Pct:  parseFloat(record[9]),
		Imbalance1Pct: parseFloat(record[10]),
		Valid:         valid,
		Error:         record[12],
	}, nil
}

func datasetQuality(rows []SnapshotRow) DatasetQuality {
	var quality DatasetQuality
	quality.TotalRows = len(rows)
	for _, row := range rows {
		if row.Valid {
			quality.ValidRows++
		} else {
			quality.InvalidRows++
		}
		if row.Error != "" {
			quality.ErrorRows++
		}
	}
	if quality.TotalRows > 0 {
		quality.ValidPct = float64(quality.ValidRows) / float64(quality.TotalRows) * 100
		quality.ErrorPct = float64(quality.ErrorRows) / float64(quality.TotalRows) * 100
	}
	return quality
}

func venueSummaries(rows []SnapshotRow) []VenueSummary {
	groups := map[string][]SnapshotRow{}
	for _, row := range rows {
		groups[row.Venue] = append(groups[row.Venue], row)
	}

	out := make([]VenueSummary, 0, len(groups))
	for venue, venueRows := range groups {
		out = append(out, summarizeVenue(venue, venueRows))
	}
	sort.Slice(out, func(i int, j int) bool { return out[i].Venue < out[j].Venue })
	return out
}

func summarizeVenue(venue string, rows []SnapshotRow) VenueSummary {
	summary := VenueSummary{
		Venue:            venue,
		SnapshotCount:    len(rows),
		MinSpreadPct:     math.Inf(1),
		MinImbalance1Pct: math.Inf(1),
		MaxSpreadPct:     math.Inf(-1),
		MaxImbalance1Pct: math.Inf(-1),
	}
	var spreadSum, imbalanceSum, bidDepthSum, askDepthSum, midSum float64
	for _, row := range rows {
		if !row.Valid {
			summary.InvalidSnapshotCount++
			continue
		}
		summary.ValidSnapshotCount++
		spreadSum += row.SpreadPct
		imbalanceSum += row.Imbalance1Pct
		bidDepthSum += row.BidDepth1Pct
		askDepthSum += row.AskDepth1Pct
		midSum += row.Mid
		if row.SpreadPct < summary.MinSpreadPct {
			summary.MinSpreadPct = row.SpreadPct
		}
		if row.SpreadPct > summary.MaxSpreadPct {
			summary.MaxSpreadPct = row.SpreadPct
		}
		if row.Imbalance1Pct < summary.MinImbalance1Pct {
			summary.MinImbalance1Pct = row.Imbalance1Pct
		}
		if row.Imbalance1Pct > summary.MaxImbalance1Pct {
			summary.MaxImbalance1Pct = row.Imbalance1Pct
		}
	}
	if summary.ValidSnapshotCount > 0 {
		divisor := float64(summary.ValidSnapshotCount)
		summary.AverageSpreadPct = spreadSum / divisor
		summary.AverageImbalance1Pct = imbalanceSum / divisor
		summary.AverageBidDepth1Pct = bidDepthSum / divisor
		summary.AverageAskDepth1Pct = askDepthSum / divisor
		summary.AverageMid = midSum / divisor
	} else {
		summary.MinSpreadPct = 0
		summary.MaxSpreadPct = 0
		summary.MinImbalance1Pct = 0
		summary.MaxImbalance1Pct = 0
	}
	return summary
}

func crossVenueSummary(rows []SnapshotRow, summaries []VenueSummary) CrossVenueSummary {
	groups := recorderRoundGroups(rows)
	var summary CrossVenueSummary
	var midDiffSum, spreadDiffSum float64
	for _, group := range groups {
		validRows := validRows(group)
		if len(validRows) < 2 {
			continue
		}
		minMid, maxMid := minMax(validRows, func(row SnapshotRow) float64 { return row.Mid })
		minSpread, maxSpread := minMax(validRows, func(row SnapshotRow) float64 { return row.SpreadPct })
		midDiff := maxMid - minMid
		spreadDiff := maxSpread - minSpread
		midDiffSum += midDiff
		spreadDiffSum += spreadDiff
		if midDiff > summary.MaxMidDifference {
			summary.MaxMidDifference = midDiff
		}
		summary.ComparableGroups++
	}
	if summary.ComparableGroups > 0 {
		divisor := float64(summary.ComparableGroups)
		summary.AverageMidDifference = midDiffSum / divisor
		summary.AverageSpreadDifference = spreadDiffSum / divisor
	}
	summary.WidestSpreadVenue = venueWithExtreme(summaries, func(s VenueSummary) float64 { return s.AverageSpreadPct }, true)
	summary.TightestSpreadVenue = venueWithExtreme(summaries, func(s VenueSummary) float64 { return s.AverageSpreadPct }, false)
	summary.DeepestLiquidityVenue = venueWithExtreme(summaries, func(s VenueSummary) float64 { return s.AverageBidDepth1Pct + s.AverageAskDepth1Pct }, true)
	return summary
}

func recorderRoundGroups(rows []SnapshotRow) [][]SnapshotRow {
	groups := make([][]SnapshotRow, 0)
	current := make([]SnapshotRow, 0, 3)
	seen := map[string]bool{}
	for _, row := range rows {
		if seen[row.Venue] && len(current) > 0 {
			groups = append(groups, current)
			current = make([]SnapshotRow, 0, 3)
			seen = map[string]bool{}
		}
		current = append(current, row)
		seen[row.Venue] = true
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}
	return groups
}

func liquidityRankings(summaries []VenueSummary) LiquidityRankings {
	return LiquidityRankings{
		AverageBidDepth:    rankVenues(summaries, func(s VenueSummary) float64 { return s.AverageBidDepth1Pct }, true),
		AverageAskDepth:    rankVenues(summaries, func(s VenueSummary) float64 { return s.AverageAskDepth1Pct }, true),
		CombinedDepth:      rankVenues(summaries, func(s VenueSummary) float64 { return s.AverageBidDepth1Pct + s.AverageAskDepth1Pct }, true),
		TightestSpreads:    rankVenues(summaries, func(s VenueSummary) float64 { return s.AverageSpreadPct }, false),
		HighestBidPressure: rankVenues(summaries, func(s VenueSummary) float64 { return s.AverageImbalance1Pct }, true),
		HighestAskPressure: rankVenues(summaries, func(s VenueSummary) float64 { return s.AverageImbalance1Pct }, false),
	}
}

func rankVenues(summaries []VenueSummary, value func(VenueSummary) float64, descending bool) []VenueRanking {
	out := make([]VenueRanking, 0, len(summaries))
	for _, summary := range summaries {
		if summary.ValidSnapshotCount == 0 {
			continue
		}
		out = append(out, VenueRanking{Venue: summary.Venue, Value: value(summary)})
	}
	sort.Slice(out, func(i int, j int) bool {
		if descending {
			return out[i].Value > out[j].Value
		}
		return out[i].Value < out[j].Value
	})
	return out
}

func venueWithExtreme(summaries []VenueSummary, value func(VenueSummary) float64, max bool) string {
	rankings := rankVenues(summaries, value, max)
	if len(rankings) == 0 {
		return ""
	}
	return rankings[0].Venue
}

func validRows(rows []SnapshotRow) []SnapshotRow {
	out := make([]SnapshotRow, 0, len(rows))
	for _, row := range rows {
		if row.Valid {
			out = append(out, row)
		}
	}
	return out
}

func minMax(rows []SnapshotRow, value func(SnapshotRow) float64) (float64, float64) {
	minValue := math.Inf(1)
	maxValue := math.Inf(-1)
	for _, row := range rows {
		v := value(row)
		if v < minValue {
			minValue = v
		}
		if v > maxValue {
			maxValue = v
		}
	}
	return minValue, maxValue
}

func parseInt(value string) int64 {
	parsed, _ := strconv.ParseInt(value, 10, 64)
	return parsed
}

func parseFloat(value string) float64 {
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}
