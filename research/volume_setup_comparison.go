package research

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	VolumeSetupAccumulation = "ACCUMULATION"
	VolumeSetupTrend        = "TREND"
	VolumeSetupRejection    = "REJECTION"
	VolumeSetupReversal     = "REVERSAL"
)

type VolumeSetupComparisonRow struct {
	SetupType                     string  `json:"setupType"`
	SetupCount                    int     `json:"setupCount"`
	AcceptedCount                 int     `json:"acceptedCount"`
	RejectedCount                 int     `json:"rejectedCount"`
	AcceptanceRate                float64 `json:"acceptanceRate"`
	RejectionRate                 float64 `json:"rejectionRate"`
	AverageFollowThrough5         float64 `json:"averageFollowThrough5"`
	AverageFollowThrough10        float64 `json:"averageFollowThrough10"`
	AverageFollowThrough20        float64 `json:"averageFollowThrough20"`
	InvalidationRate              float64 `json:"invalidationRate"`
	VWAPAlignmentRate             float64 `json:"vwapAlignmentRate"`
	POCConfluenceRate             float64 `json:"pocConfluenceRate"`
	HVNConfluenceRate             float64 `json:"hvnConfluenceRate"`
	DailyPOCConfluenceRate        float64 `json:"dailyPocConfluenceRate"`
	Rolling3DPOCConfluenceRate    float64 `json:"rolling3dPocConfluenceRate"`
	Rolling7DPOCConfluenceRate    float64 `json:"rolling7dPocConfluenceRate"`
	Composite30DPOCConfluenceRate float64 `json:"composite30dPocConfluenceRate"`
	LongCount                     int     `json:"longCount"`
	ShortCount                    int     `json:"shortCount"`
	AverageProfileVolume          float64 `json:"averageProfileVolume"`
	RankAcceptance                int     `json:"rankAcceptance"`
	RankFT5                       int     `json:"rankFt5"`
	RankFT10                      int     `json:"rankFt10"`
	RankFT20                      int     `json:"rankFt20"`
	RankInvalidation              int     `json:"rankInvalidation"`
	RankVWAP                      int     `json:"rankVwap"`
	RankPOC                       int     `json:"rankPoc"`
	RankHVN                       int     `json:"rankHvn"`
	RankSampleSize                int     `json:"rankSampleSize"`
	Notes                         string  `json:"notes"`
}

type VolumeSetupComparisonSummary struct {
	BestAcceptanceSetup string                     `json:"bestAcceptanceSetup"`
	BestAcceptanceRate  float64                    `json:"bestAcceptanceRate"`
	BestFT5Setup        string                     `json:"bestFT5Setup"`
	BestFT10Setup       string                     `json:"bestFT10Setup"`
	BestFT20Setup       string                     `json:"bestFT20Setup"`
	BestVWAPSetup       string                     `json:"bestVWAPSetup"`
	BestPOCSetup        string                     `json:"bestPOCSetup"`
	BestHVNSetup        string                     `json:"bestHVNSetup"`
	LargestSampleSetup  string                     `json:"largestSampleSetup"`
	Rankings            []VolumeSetupComparisonRow `json:"rankings"`
}

type volumeSetupComparisonAccumulator struct {
	setupType       string
	count           int
	accepted        int
	rejected        int
	invalidated     int
	vwapAligned     int
	pocConfluence   int
	hvnConfluence   int
	dailyPOC        int
	rolling3DPOC    int
	rolling7DPOC    int
	composite30DPOC int
	longCount       int
	shortCount      int
	volumeSum       float64
	ft5             float64
	ft10            float64
	ft20            float64
}

func BuildVolumeSetupComparisonFromFiles(accumulationPath string, trendPath string, rejectionPath string, reversalPath string) ([]VolumeSetupComparisonRow, VolumeSetupComparisonSummary, error) {
	rows := make([]VolumeSetupComparisonRow, 0, 4)
	for _, input := range []struct {
		setupType string
		path      string
	}{
		{VolumeSetupAccumulation, accumulationPath},
		{VolumeSetupTrend, trendPath},
		{VolumeSetupRejection, rejectionPath},
		{VolumeSetupReversal, reversalPath},
	} {
		records, err := readCSVRecords(input.path)
		if err != nil {
			return nil, VolumeSetupComparisonSummary{}, err
		}
		rows = append(rows, buildVolumeSetupComparisonRow(input.setupType, records))
	}
	applyVolumeSetupRanks(rows)
	summary := BuildVolumeSetupComparisonSummary(rows)
	return rows, summary, nil
}

func BuildVolumeSetupComparisonSummary(rows []VolumeSetupComparisonRow) VolumeSetupComparisonSummary {
	summary := VolumeSetupComparisonSummary{Rankings: append([]VolumeSetupComparisonRow(nil), rows...)}
	sort.Slice(summary.Rankings, func(i, j int) bool {
		return summary.Rankings[i].RankFT20 < summary.Rankings[j].RankFT20
	})
	for _, row := range rows {
		if summary.BestAcceptanceSetup == "" || row.AcceptanceRate > summary.BestAcceptanceRate {
			summary.BestAcceptanceSetup = row.SetupType
			summary.BestAcceptanceRate = row.AcceptanceRate
		}
		if summary.BestFT5Setup == "" || row.AverageFollowThrough5 > rowBySetup(rows, summary.BestFT5Setup).AverageFollowThrough5 {
			summary.BestFT5Setup = row.SetupType
		}
		if summary.BestFT10Setup == "" || row.AverageFollowThrough10 > rowBySetup(rows, summary.BestFT10Setup).AverageFollowThrough10 {
			summary.BestFT10Setup = row.SetupType
		}
		if summary.BestFT20Setup == "" || row.AverageFollowThrough20 > rowBySetup(rows, summary.BestFT20Setup).AverageFollowThrough20 {
			summary.BestFT20Setup = row.SetupType
		}
		if summary.BestVWAPSetup == "" || row.VWAPAlignmentRate > rowBySetup(rows, summary.BestVWAPSetup).VWAPAlignmentRate {
			summary.BestVWAPSetup = row.SetupType
		}
		if summary.BestPOCSetup == "" || row.POCConfluenceRate > rowBySetup(rows, summary.BestPOCSetup).POCConfluenceRate {
			summary.BestPOCSetup = row.SetupType
		}
		if summary.BestHVNSetup == "" || row.HVNConfluenceRate > rowBySetup(rows, summary.BestHVNSetup).HVNConfluenceRate {
			summary.BestHVNSetup = row.SetupType
		}
		if summary.LargestSampleSetup == "" || row.SetupCount > rowBySetup(rows, summary.LargestSampleSetup).SetupCount {
			summary.LargestSampleSetup = row.SetupType
		}
	}
	return summary
}

func WriteVolumeSetupComparisonCSV(path string, rows []VolumeSetupComparisonRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{
		"setup_type", "setup_count", "accepted_count", "rejected_count", "acceptance_rate", "rejection_rate",
		"average_follow_through_5", "average_follow_through_10", "average_follow_through_20", "invalidation_rate",
		"vwap_alignment_rate", "poc_confluence_rate", "hvn_confluence_rate", "long_count", "short_count",
		"average_profile_volume", "rank_acceptance", "rank_ft5", "rank_ft10", "rank_ft20", "rank_invalidation", "notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.SetupType,
			strconv.Itoa(row.SetupCount),
			strconv.Itoa(row.AcceptedCount),
			strconv.Itoa(row.RejectedCount),
			floatToString(row.AcceptanceRate),
			floatToString(row.RejectionRate),
			floatToString(row.AverageFollowThrough5),
			floatToString(row.AverageFollowThrough10),
			floatToString(row.AverageFollowThrough20),
			floatToString(row.InvalidationRate),
			floatToString(row.VWAPAlignmentRate),
			floatToString(row.POCConfluenceRate),
			floatToString(row.HVNConfluenceRate),
			strconv.Itoa(row.LongCount),
			strconv.Itoa(row.ShortCount),
			floatToString(row.AverageProfileVolume),
			strconv.Itoa(row.RankAcceptance),
			strconv.Itoa(row.RankFT5),
			strconv.Itoa(row.RankFT10),
			strconv.Itoa(row.RankFT20),
			strconv.Itoa(row.RankInvalidation),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVolumeSetupComparisonJSON(path string, summary VolumeSetupComparisonSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeSetupComparisonMarkdown(path string, rows []VolumeSetupComparisonRow, summary VolumeSetupComparisonSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Profile Setup Comparison\n\n")
	body.WriteString("## Executive Summary\n\n")
	body.WriteString(fmt.Sprintf("- Best acceptance setup: %s (%.2f)\n- Best FT20 setup: %s\n- Best VWAP setup: %s\n- Best POC setup: %s\n- Best HVN setup: %s\n- Largest sample: %s\n\n", summary.BestAcceptanceSetup, summary.BestAcceptanceRate, summary.BestFT20Setup, summary.BestVWAPSetup, summary.BestPOCSetup, summary.BestHVNSetup, summary.LargestSampleSetup))
	for _, setupType := range []string{VolumeSetupAccumulation, VolumeSetupTrend, VolumeSetupRejection, VolumeSetupReversal} {
		writeVolumeSetupComparisonSection(&body, setupType, rows)
	}
	body.WriteString("## Acceptance Ranking\n\n")
	writeVolumeSetupRanking(&body, rows, "acceptance")
	body.WriteString("## Follow Through Ranking\n\n")
	writeVolumeSetupRanking(&body, rows, "ft20")
	body.WriteString("## Confluence Ranking\n\n")
	writeVolumeSetupConfluenceRanking(&body, rows)
	body.WriteString("## Sample Size Ranking\n\n")
	writeVolumeSetupRanking(&body, rows, "sample")
	body.WriteString("## Strengths\n\n")
	body.WriteString("- Side-by-side setup metrics now make acceptance, follow-through, confluence, and sample size comparable.\n")
	body.WriteString("- Reversal and rejection studies can be inspected against failed-auction and profile-level context before any executable strategy work.\n\n")
	body.WriteString("## Weaknesses\n\n")
	body.WriteString("- Follow-through remains passive candle movement, not simulated execution.\n")
	body.WriteString("- Missing confluence fields stay at zero instead of being inferred.\n")
	body.WriteString("- Volume Profile still uses OHLCV volume-at-price approximation.\n\n")
	body.WriteString("## Recommended Research Focus\n\n")
	body.WriteString("- Prioritize setups with non-negative FT20 and enough sample size for manual chart review.\n")
	body.WriteString("- Treat tiny-sample winners as hypotheses, not conclusions.\n\n")
	body.WriteString("## Recommendation Before Pacifica\n\n")
	body.WriteString("Finish manual review of the strongest Volume Profile setup filters before adding any venue-specific Pacifica work.\n\n")
	body.WriteString("## Recommendation Before ML\n\n")
	body.WriteString("Keep these setup comparison outputs as labeled research context, but do not add ML until the book-aligned research layer is stable.\n")
	return os.WriteFile(path, []byte(body.String()), 0644)
}

func buildVolumeSetupComparisonRow(setupType string, records []map[string]string) VolumeSetupComparisonRow {
	acc := volumeSetupComparisonAccumulator{setupType: setupType}
	for _, record := range records {
		acc.add(record)
	}
	return acc.row()
}

func (a *volumeSetupComparisonAccumulator) add(record map[string]string) {
	a.count++
	if parseBoolRecord(record["accepted"]) {
		a.accepted++
	}
	if parseBoolRecord(record["rejected"]) {
		a.rejected++
	}
	if parseBoolRecord(record["invalidated"]) {
		a.invalidated++
	}
	direction := record["direction"]
	if direction == "long_context" {
		a.longCount++
	}
	if direction == "short_context" {
		a.shortCount++
	}
	a.volumeSum += parseFloatRecord(record["profile_volume"])
	a.ft5 += parseFloatRecord(record["follow_through_5"])
	a.ft10 += parseFloatRecord(record["follow_through_10"])
	a.ft20 += parseFloatRecord(record["follow_through_20"])
	if vwapAligned(direction, record["vwap_alignment"]) {
		a.vwapAligned++
	}
	if setupRecordPOCConfluence(record) {
		a.pocConfluence++
	}
	if setupRecordHVNConfluence(record) {
		a.hvnConfluence++
	}
	if parseBoolRecord(record["daily_poc_confluence"]) {
		a.dailyPOC++
	}
	if parseBoolRecord(record["rolling3d_poc_confluence"]) {
		a.rolling3DPOC++
	}
	if parseBoolRecord(record["rolling7d_poc_confluence"]) {
		a.rolling7DPOC++
	}
	if parseBoolRecord(record["composite30d_poc_confluence"]) {
		a.composite30DPOC++
	}
}

func (a volumeSetupComparisonAccumulator) row() VolumeSetupComparisonRow {
	row := VolumeSetupComparisonRow{
		SetupType: a.setupType,
		Notes:     "Research-only comparison; no setup logic or execution behavior changed.",
	}
	if a.count == 0 {
		return row
	}
	count := float64(a.count)
	row.SetupCount = a.count
	row.AcceptedCount = a.accepted
	row.RejectedCount = a.rejected
	row.AcceptanceRate = float64(a.accepted) / count
	row.RejectionRate = float64(a.rejected) / count
	row.AverageFollowThrough5 = a.ft5 / count
	row.AverageFollowThrough10 = a.ft10 / count
	row.AverageFollowThrough20 = a.ft20 / count
	row.InvalidationRate = float64(a.invalidated) / count
	row.VWAPAlignmentRate = float64(a.vwapAligned) / count
	row.POCConfluenceRate = float64(a.pocConfluence) / count
	row.HVNConfluenceRate = float64(a.hvnConfluence) / count
	row.DailyPOCConfluenceRate = float64(a.dailyPOC) / count
	row.Rolling3DPOCConfluenceRate = float64(a.rolling3DPOC) / count
	row.Rolling7DPOCConfluenceRate = float64(a.rolling7DPOC) / count
	row.Composite30DPOCConfluenceRate = float64(a.composite30DPOC) / count
	row.LongCount = a.longCount
	row.ShortCount = a.shortCount
	row.AverageProfileVolume = a.volumeSum / count
	return row
}

func setupRecordPOCConfluence(record map[string]string) bool {
	return parseBoolRecord(record["poc_confluence"]) ||
		parseBoolRecord(record["daily_poc_confluence"]) ||
		parseBoolRecord(record["rolling3d_poc_confluence"]) ||
		parseBoolRecord(record["rolling7d_poc_confluence"]) ||
		parseBoolRecord(record["composite30d_poc_confluence"])
}

func setupRecordHVNConfluence(record map[string]string) bool {
	return parseBoolRecord(record["hvn_confluence"]) || parseBoolRecord(record["hvn_retest_detected"])
}

func applyVolumeSetupRanks(rows []VolumeSetupComparisonRow) {
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.AcceptanceRate }, false, func(i int, rank int) { rows[i].RankAcceptance = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.AverageFollowThrough5 }, false, func(i int, rank int) { rows[i].RankFT5 = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.AverageFollowThrough10 }, false, func(i int, rank int) { rows[i].RankFT10 = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.AverageFollowThrough20 }, false, func(i int, rank int) { rows[i].RankFT20 = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.InvalidationRate }, true, func(i int, rank int) { rows[i].RankInvalidation = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.VWAPAlignmentRate }, false, func(i int, rank int) { rows[i].RankVWAP = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.POCConfluenceRate }, false, func(i int, rank int) { rows[i].RankPOC = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return row.HVNConfluenceRate }, false, func(i int, rank int) { rows[i].RankHVN = rank })
	assignVolumeSetupRank(rows, func(row VolumeSetupComparisonRow) float64 { return float64(row.SetupCount) }, false, func(i int, rank int) { rows[i].RankSampleSize = rank })
}

func assignVolumeSetupRank(rows []VolumeSetupComparisonRow, value func(VolumeSetupComparisonRow) float64, ascending bool, set func(int, int)) {
	indexes := make([]int, len(rows))
	for i := range rows {
		indexes[i] = i
	}
	sort.Slice(indexes, func(i, j int) bool {
		left := value(rows[indexes[i]])
		right := value(rows[indexes[j]])
		if ascending {
			return left < right
		}
		return left > right
	})
	for rank, index := range indexes {
		set(index, rank+1)
	}
}

func rowBySetup(rows []VolumeSetupComparisonRow, setupType string) VolumeSetupComparisonRow {
	for _, row := range rows {
		if row.SetupType == setupType {
			return row
		}
	}
	return VolumeSetupComparisonRow{}
}

func writeVolumeSetupComparisonSection(body *strings.Builder, setupType string, rows []VolumeSetupComparisonRow) {
	row := rowBySetup(rows, setupType)
	body.WriteString("## " + setupTypeTitle(setupType) + "\n\n")
	body.WriteString(fmt.Sprintf("- Setups: %d\n- Acceptance rate: %.2f\n- Rejection rate: %.2f\n- Average FT20: %.2f\n- VWAP alignment rate: %.2f\n- POC confluence rate: %.2f\n- HVN confluence rate: %.2f\n\n", row.SetupCount, row.AcceptanceRate, row.RejectionRate, row.AverageFollowThrough20, row.VWAPAlignmentRate, row.POCConfluenceRate, row.HVNConfluenceRate))
}

func writeVolumeSetupRanking(body *strings.Builder, rows []VolumeSetupComparisonRow, metric string) {
	body.WriteString("| Rank | Setup | Value |\n|---:|---|---:|\n")
	sorted := append([]VolumeSetupComparisonRow(nil), rows...)
	sort.Slice(sorted, func(i, j int) bool {
		return volumeSetupRankingValue(sorted[i], metric) > volumeSetupRankingValue(sorted[j], metric)
	})
	if metric == "invalidation" {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].InvalidationRate < sorted[j].InvalidationRate
		})
	}
	for i, row := range sorted {
		body.WriteString(fmt.Sprintf("| %d | %s | %.2f |\n", i+1, row.SetupType, volumeSetupRankingValue(row, metric)))
	}
	body.WriteString("\n")
}

func writeVolumeSetupConfluenceRanking(body *strings.Builder, rows []VolumeSetupComparisonRow) {
	body.WriteString("| Setup | VWAP | POC | HVN |\n|---|---:|---:|---:|\n")
	for _, row := range rows {
		body.WriteString(fmt.Sprintf("| %s | %.2f | %.2f | %.2f |\n", row.SetupType, row.VWAPAlignmentRate, row.POCConfluenceRate, row.HVNConfluenceRate))
	}
	body.WriteString("\n")
}

func volumeSetupRankingValue(row VolumeSetupComparisonRow, metric string) float64 {
	switch metric {
	case "acceptance":
		return row.AcceptanceRate
	case "ft5":
		return row.AverageFollowThrough5
	case "ft10":
		return row.AverageFollowThrough10
	case "ft20":
		return row.AverageFollowThrough20
	case "sample":
		return float64(row.SetupCount)
	default:
		return 0
	}
}

func setupTypeTitle(setupType string) string {
	return strings.Title(strings.ToLower(setupType))
}
