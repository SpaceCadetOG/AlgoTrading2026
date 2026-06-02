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

type VolumeSetupTrendQualityRow struct {
	GroupType              string  `json:"groupType"`
	GroupName              string  `json:"groupName"`
	SetupCount             int     `json:"setupCount"`
	AcceptedCount          int     `json:"acceptedCount"`
	RejectedCount          int     `json:"rejectedCount"`
	AcceptanceRate         float64 `json:"acceptanceRate"`
	RejectionRate          float64 `json:"rejectionRate"`
	AverageTrendStrength   float64 `json:"averageTrendStrength"`
	AverageProfileVolume   float64 `json:"averageProfileVolume"`
	AverageShapeConfidence float64 `json:"averageShapeConfidence"`
	AverageFollowThrough5  float64 `json:"averageFollowThrough5"`
	AverageFollowThrough10 float64 `json:"averageFollowThrough10"`
	AverageFollowThrough20 float64 `json:"averageFollowThrough20"`
	InvalidationRate       float64 `json:"invalidationRate"`
	Notes                  string  `json:"notes"`
}

type VolumeSetupTrendQualitySummary struct {
	TotalSetups                     int      `json:"totalSetups"`
	Accepted                        int      `json:"accepted"`
	Rejected                        int      `json:"rejected"`
	BestDirection                   string   `json:"bestDirection"`
	BestDirectionFollowThrough20    float64  `json:"bestDirectionFollowThrough20"`
	BestShape                       string   `json:"bestShape"`
	BestShapeAcceptanceRate         float64  `json:"bestShapeAcceptanceRate"`
	BestTrendStrengthBucket         string   `json:"bestTrendStrengthBucket"`
	BestTrendStrengthAcceptanceRate float64  `json:"bestTrendStrengthAcceptanceRate"`
	BestConfluence                  string   `json:"bestConfluence"`
	BestConfluenceAcceptanceRate    float64  `json:"bestConfluenceAcceptanceRate"`
	Notes                           []string `json:"notes"`
}

type trendQualityRecord struct {
	Direction          string
	TrendStrengthScore float64
	ProfileShape       string
	ProfileVolume      float64
	ShapeConfidence    float64
	Accepted           bool
	Rejected           bool
	POCRetestDetected  bool
	HVNRetestDetected  bool
	VWAPAlignment      string
	Invalidated        bool
	FollowThrough5     float64
	FollowThrough10    float64
	FollowThrough20    float64
}

func BuildVolumeSetupTrendQualityFromFile(path string) ([]VolumeSetupTrendQualityRow, VolumeSetupTrendQualitySummary, error) {
	records, err := readTrendQualityRecords(path)
	if err != nil {
		return nil, VolumeSetupTrendQualitySummary{}, err
	}
	rows := BuildVolumeSetupTrendQualityRows(records)
	summary := BuildVolumeSetupTrendQualitySummary(records, rows)
	return rows, summary, nil
}

func BuildVolumeSetupTrendQualityRows(records []trendQualityRecord) []VolumeSetupTrendQualityRow {
	groups := map[string]*trendQualityAccumulator{}
	for _, record := range records {
		addTrendQualityGroup(groups, "acceptance_state", boolGroupName(record.Accepted, record.Rejected)).add(record)
		addTrendQualityGroup(groups, "direction", record.Direction).add(record)
		addTrendQualityGroup(groups, "poc_retest", boolName(record.POCRetestDetected, "with_poc_retest", "without_poc_retest")).add(record)
		addTrendQualityGroup(groups, "hvn_retest", boolName(record.HVNRetestDetected, "with_hvn_retest", "without_hvn_retest")).add(record)
		addTrendQualityGroup(groups, "vwap_alignment", boolName(trendRecordVWAPAligned(record), "vwap_aligned", "vwap_not_aligned")).add(record)
		addTrendQualityGroup(groups, "profile_shape", profileShapeOrUnknown(record.ProfileShape)).add(record)
		addTrendQualityGroup(groups, "trend_strength_bucket", trendStrengthBucket(record.TrendStrengthScore)).add(record)
	}
	rows := make([]VolumeSetupTrendQualityRow, 0, len(groups))
	for _, group := range groups {
		rows = append(rows, group.row())
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].GroupType == rows[j].GroupType {
			return rows[i].GroupName < rows[j].GroupName
		}
		return rows[i].GroupType < rows[j].GroupType
	})
	return rows
}

func BuildVolumeSetupTrendQualitySummary(records []trendQualityRecord, rows []VolumeSetupTrendQualityRow) VolumeSetupTrendQualitySummary {
	summary := VolumeSetupTrendQualitySummary{
		TotalSetups: len(records),
		Notes: []string{
			"Research-only quality analysis; trend setup detection rules are unchanged.",
			"Acceptance/rejection comes from trend-leg value-area behavior.",
			"Follow-through is passive candle movement, not simulated execution.",
		},
	}
	for _, record := range records {
		if record.Accepted {
			summary.Accepted++
		}
		if record.Rejected {
			summary.Rejected++
		}
	}
	for _, row := range rows {
		if row.GroupType == "direction" && (summary.BestDirection == "" || row.AverageFollowThrough20 > summary.BestDirectionFollowThrough20) {
			summary.BestDirection = row.GroupName
			summary.BestDirectionFollowThrough20 = row.AverageFollowThrough20
		}
		if row.GroupType == "profile_shape" && row.AcceptanceRate > summary.BestShapeAcceptanceRate {
			summary.BestShape = row.GroupName
			summary.BestShapeAcceptanceRate = row.AcceptanceRate
		}
		if row.GroupType == "trend_strength_bucket" && row.AcceptanceRate > summary.BestTrendStrengthAcceptanceRate {
			summary.BestTrendStrengthBucket = row.GroupName
			summary.BestTrendStrengthAcceptanceRate = row.AcceptanceRate
		}
		isPositiveConfluence := isTrendConfluenceGroup(row) && strings.HasPrefix(row.GroupName, "with_") ||
			row.GroupType == "vwap_alignment" && row.GroupName == "vwap_aligned"
		if isPositiveConfluence {
			if row.AcceptanceRate > summary.BestConfluenceAcceptanceRate {
				summary.BestConfluence = row.GroupType + ":" + row.GroupName
				summary.BestConfluenceAcceptanceRate = row.AcceptanceRate
			}
		}
	}
	return summary
}

func WriteVolumeSetupTrendQualityCSV(path string, rows []VolumeSetupTrendQualityRow) error {
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
	if err := writer.Write([]string{"group_type", "group_name", "setup_count", "accepted_count", "rejected_count", "acceptance_rate", "rejection_rate", "average_trend_strength", "average_profile_volume", "average_shape_confidence", "average_follow_through_5", "average_follow_through_10", "average_follow_through_20", "invalidation_rate", "notes"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.GroupType, row.GroupName,
			strconv.Itoa(row.SetupCount),
			strconv.Itoa(row.AcceptedCount),
			strconv.Itoa(row.RejectedCount),
			floatToString(row.AcceptanceRate),
			floatToString(row.RejectionRate),
			floatToString(row.AverageTrendStrength),
			floatToString(row.AverageProfileVolume),
			floatToString(row.AverageShapeConfidence),
			floatToString(row.AverageFollowThrough5),
			floatToString(row.AverageFollowThrough10),
			floatToString(row.AverageFollowThrough20),
			floatToString(row.InvalidationRate),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVolumeSetupTrendQualitySummaryJSON(path string, summary VolumeSetupTrendQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeSetupTrendQualityMarkdown(path string, rows []VolumeSetupTrendQualityRow, summary VolumeSetupTrendQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Setup #2 Trend Quality Review\n\n")
	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("- Total setups: %d\n- Accepted: %d\n- Rejected: %d\n- Best direction by 20-candle follow-through: %s (%.2f)\n- Best shape: %s (%.2f)\n- Best trend-strength bucket: %s (%.2f)\n- Best confluence: %s (%.2f)\n\n", summary.TotalSetups, summary.Accepted, summary.Rejected, summary.BestDirection, summary.BestDirectionFollowThrough20, summary.BestShape, summary.BestShapeAcceptanceRate, summary.BestTrendStrengthBucket, summary.BestTrendStrengthAcceptanceRate, summary.BestConfluence, summary.BestConfluenceAcceptanceRate))
	writeTrendQualitySection(&body, "Accepted vs Rejected", rows, "acceptance_state")
	writeTrendQualitySection(&body, "Long vs Short", rows, "direction")
	writeTrendQualitySection(&body, "POC Retest Quality", rows, "poc_retest")
	writeTrendQualitySection(&body, "HVN Retest Quality", rows, "hvn_retest")
	writeTrendQualitySection(&body, "VWAP Alignment Quality", rows, "vwap_alignment")
	writeTrendQualitySection(&body, "Profile Shape Quality", rows, "profile_shape")
	writeTrendQualitySection(&body, "Trend Strength Quality", rows, "trend_strength_bucket")
	body.WriteString("## What This Means\n\n")
	body.WriteString("This review separates trend-leg continuation context by acceptance, retest behavior, VWAP alignment, shape, and strength. It does not change detection rules and does not simulate orders.\n\n")
	body.WriteString("## Comparison With Setup #1 Accumulation\n\n")
	body.WriteString("Trend setups are evaluated after directional initiation, while Setup #1 accumulation studies the volume level left behind by sideways positioning. The two reports should be compared manually before turning either into a hard filter.\n\n")
	body.WriteString("## Recommended Filters Before Setup #3\n\n")
	body.WriteString("- Inspect whether POC or HVN retests add acceptance quality before requiring them.\n")
	body.WriteString("- Prefer shapes and strength buckets with better acceptance only after visual review.\n")
	body.WriteString("- Keep VWAP/L2 fields contextual because current L2 state is not historical replay.\n")
	return os.WriteFile(path, []byte(body.String()), 0644)
}

type trendQualityAccumulator struct {
	groupType       string
	groupName       string
	count           int
	accepted        int
	rejected        int
	invalidated     int
	strengthSum     float64
	volumeSum       float64
	confidenceSum   float64
	followThrough5  float64
	followThrough10 float64
	followThrough20 float64
}

func addTrendQualityGroup(groups map[string]*trendQualityAccumulator, groupType, groupName string) *trendQualityAccumulator {
	key := groupType + ":" + groupName
	if groups[key] == nil {
		groups[key] = &trendQualityAccumulator{groupType: groupType, groupName: groupName}
	}
	return groups[key]
}

func (a *trendQualityAccumulator) add(record trendQualityRecord) {
	a.count++
	if record.Accepted {
		a.accepted++
	}
	if record.Rejected {
		a.rejected++
	}
	if record.Invalidated {
		a.invalidated++
	}
	a.strengthSum += record.TrendStrengthScore
	a.volumeSum += record.ProfileVolume
	a.confidenceSum += record.ShapeConfidence
	a.followThrough5 += record.FollowThrough5
	a.followThrough10 += record.FollowThrough10
	a.followThrough20 += record.FollowThrough20
}

func (a *trendQualityAccumulator) row() VolumeSetupTrendQualityRow {
	if a.count == 0 {
		return VolumeSetupTrendQualityRow{GroupType: a.groupType, GroupName: a.groupName}
	}
	count := float64(a.count)
	return VolumeSetupTrendQualityRow{
		GroupType:              a.groupType,
		GroupName:              a.groupName,
		SetupCount:             a.count,
		AcceptedCount:          a.accepted,
		RejectedCount:          a.rejected,
		AcceptanceRate:         float64(a.accepted) / count,
		RejectionRate:          float64(a.rejected) / count,
		AverageTrendStrength:   a.strengthSum / count,
		AverageProfileVolume:   a.volumeSum / count,
		AverageShapeConfidence: a.confidenceSum / count,
		AverageFollowThrough5:  a.followThrough5 / count,
		AverageFollowThrough10: a.followThrough10 / count,
		AverageFollowThrough20: a.followThrough20 / count,
		InvalidationRate:       float64(a.invalidated) / count,
		Notes:                  "Research-only grouping; not a trading filter yet.",
	}
}

func readTrendQualityRecords(path string) ([]trendQualityRecord, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]trendQualityRecord, 0, len(records))
	for _, record := range records {
		out = append(out, trendQualityRecord{
			Direction:          record["direction"],
			TrendStrengthScore: parseFloatRecord(record["trend_strength_score"]),
			ProfileShape:       record["profile_shape"],
			ProfileVolume:      parseFloatRecord(record["profile_volume"]),
			ShapeConfidence:    parseFloatRecord(record["shape_confidence"]),
			Accepted:           parseBoolRecord(record["accepted"]),
			Rejected:           parseBoolRecord(record["rejected"]),
			POCRetestDetected:  parseBoolRecord(record["poc_retest_detected"]),
			HVNRetestDetected:  parseBoolRecord(record["hvn_retest_detected"]),
			VWAPAlignment:      record["vwap_alignment"],
			Invalidated:        parseBoolRecord(record["invalidated"]),
			FollowThrough5:     parseFloatRecord(record["follow_through_5"]),
			FollowThrough10:    parseFloatRecord(record["follow_through_10"]),
			FollowThrough20:    parseFloatRecord(record["follow_through_20"]),
		})
	}
	return out, nil
}

func trendRecordVWAPAligned(record trendQualityRecord) bool {
	return vwapAligned(record.Direction, record.VWAPAlignment)
}

func profileShapeOrUnknown(shape string) string {
	if shape == "" {
		return "UNKNOWN"
	}
	return shape
}

func trendStrengthBucket(strength float64) string {
	switch {
	case strength < 1.80:
		return "weak"
	case strength < 2.40:
		return "medium"
	default:
		return "strong"
	}
}

func isTrendConfluenceGroup(row VolumeSetupTrendQualityRow) bool {
	return row.GroupType == "poc_retest" || row.GroupType == "hvn_retest"
}

func writeTrendQualitySection(body *strings.Builder, title string, rows []VolumeSetupTrendQualityRow, groupType string) {
	body.WriteString("## " + title + "\n\n")
	body.WriteString("| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |\n")
	body.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, row := range rows {
		if row.GroupType != groupType {
			continue
		}
		body.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %.2f | %.2f | %.2f | %.2f |\n", row.GroupName, row.SetupCount, row.AcceptedCount, row.RejectedCount, row.AcceptanceRate, row.InvalidationRate, row.AverageTrendStrength, row.AverageFollowThrough20))
	}
	body.WriteString("\n")
}
