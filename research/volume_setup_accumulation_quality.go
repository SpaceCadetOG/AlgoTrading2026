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

type VolumeSetupAccumulationQualityRow struct {
	GroupType              string  `json:"groupType"`
	GroupName              string  `json:"groupName"`
	SetupCount             int     `json:"setupCount"`
	AcceptedCount          int     `json:"acceptedCount"`
	RejectedCount          int     `json:"rejectedCount"`
	AcceptanceRate         float64 `json:"acceptanceRate"`
	RejectionRate          float64 `json:"rejectionRate"`
	RetestRate             float64 `json:"retestRate"`
	InvalidationRate       float64 `json:"invalidationRate"`
	AverageProfileVolume   float64 `json:"averageProfileVolume"`
	AverageShapeConfidence float64 `json:"averageShapeConfidence"`
	AverageFollowThrough5  float64 `json:"averageFollowThrough5"`
	AverageFollowThrough10 float64 `json:"averageFollowThrough10"`
	AverageFollowThrough20 float64 `json:"averageFollowThrough20"`
	Notes                  string  `json:"notes"`
}

type VolumeSetupAccumulationQualitySummary struct {
	TotalSetups                         int      `json:"totalSetups"`
	Accepted                            int      `json:"accepted"`
	Rejected                            int      `json:"rejected"`
	BestConfluenceGroup                 string   `json:"bestConfluenceGroup"`
	BestConfluenceAcceptanceRate        float64  `json:"bestConfluenceAcceptanceRate"`
	BestShape                           string   `json:"bestShape"`
	BestShapeAcceptanceRate             float64  `json:"bestShapeAcceptanceRate"`
	BestDirection                       string   `json:"bestDirection"`
	BestDirectionAverageFollowThrough20 float64  `json:"bestDirectionAverageFollowThrough20"`
	RetestAcceptanceRate                float64  `json:"retestAcceptanceRate"`
	NonRetestAcceptanceRate             float64  `json:"nonRetestAcceptanceRate"`
	Notes                               []string `json:"notes"`
}

type accumulationQualityRecord struct {
	Direction                 string
	ProfileShape              string
	ProfileVolume             float64
	ShapeConfidence           float64
	Accepted                  bool
	Rejected                  bool
	DailyPOCConfluence        bool
	Rolling3DPOCConfluence    bool
	Rolling7DPOCConfluence    bool
	Composite30DPOCConfluence bool
	RetestDetected            bool
	Invalidated               bool
	FollowThrough5            float64
	FollowThrough10           float64
	FollowThrough20           float64
}

func BuildVolumeSetupAccumulationQualityFromFile(path string) ([]VolumeSetupAccumulationQualityRow, VolumeSetupAccumulationQualitySummary, error) {
	records, err := readAccumulationQualityRecords(path)
	if err != nil {
		return nil, VolumeSetupAccumulationQualitySummary{}, err
	}
	rows := BuildVolumeSetupAccumulationQualityRows(records)
	summary := BuildVolumeSetupAccumulationQualitySummary(records, rows)
	return rows, summary, nil
}

func BuildVolumeSetupAccumulationQualityRows(records []accumulationQualityRecord) []VolumeSetupAccumulationQualityRow {
	groups := map[string]*accumulationQualityAccumulator{}
	for _, record := range records {
		addAccumulationQualityGroup(groups, "acceptance_state", boolGroupName(record.Accepted, record.Rejected)).add(record)
		addAccumulationQualityGroup(groups, "direction", record.Direction).add(record)
		addAccumulationQualityGroup(groups, "profile_shape", record.ProfileShape).add(record)
		addAccumulationQualityGroup(groups, "retest", boolName(record.RetestDetected, "retested", "not_retested")).add(record)
		addAccumulationQualityGroup(groups, "daily_poc_confluence", boolName(record.DailyPOCConfluence, "with_daily_poc", "without_daily_poc")).add(record)
		addAccumulationQualityGroup(groups, "rolling3d_poc_confluence", boolName(record.Rolling3DPOCConfluence, "with_rolling3d_poc", "without_rolling3d_poc")).add(record)
		addAccumulationQualityGroup(groups, "rolling7d_poc_confluence", boolName(record.Rolling7DPOCConfluence, "with_rolling7d_poc", "without_rolling7d_poc")).add(record)
		addAccumulationQualityGroup(groups, "composite30d_poc_confluence", boolName(record.Composite30DPOCConfluence, "with_composite30d_poc", "without_composite30d_poc")).add(record)
	}
	rows := make([]VolumeSetupAccumulationQualityRow, 0, len(groups))
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

func BuildVolumeSetupAccumulationQualitySummary(records []accumulationQualityRecord, rows []VolumeSetupAccumulationQualityRow) VolumeSetupAccumulationQualitySummary {
	summary := VolumeSetupAccumulationQualitySummary{
		TotalSetups: len(records),
		Notes: []string{
			"Research-only quality analysis; setup detection rules are unchanged.",
			"Acceptance/rejection comes from the flexible profile acceptance state.",
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
		if strings.Contains(row.GroupType, "poc_confluence") && strings.HasPrefix(row.GroupName, "with_") && row.AcceptanceRate > summary.BestConfluenceAcceptanceRate {
			summary.BestConfluenceGroup = row.GroupType + ":" + row.GroupName
			summary.BestConfluenceAcceptanceRate = row.AcceptanceRate
		}
		if row.GroupType == "profile_shape" && row.AcceptanceRate > summary.BestShapeAcceptanceRate {
			summary.BestShape = row.GroupName
			summary.BestShapeAcceptanceRate = row.AcceptanceRate
		}
		if row.GroupType == "direction" && (summary.BestDirection == "" || row.AverageFollowThrough20 > summary.BestDirectionAverageFollowThrough20) {
			summary.BestDirection = row.GroupName
			summary.BestDirectionAverageFollowThrough20 = row.AverageFollowThrough20
		}
		if row.GroupType == "retest" && row.GroupName == "retested" {
			summary.RetestAcceptanceRate = row.AcceptanceRate
		}
		if row.GroupType == "retest" && row.GroupName == "not_retested" {
			summary.NonRetestAcceptanceRate = row.AcceptanceRate
		}
	}
	return summary
}

func WriteVolumeSetupAccumulationQualityCSV(path string, rows []VolumeSetupAccumulationQualityRow) error {
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
	if err := writer.Write([]string{"group_type", "group_name", "setup_count", "accepted_count", "rejected_count", "acceptance_rate", "rejection_rate", "retest_rate", "invalidation_rate", "average_profile_volume", "average_shape_confidence", "average_follow_through_5", "average_follow_through_10", "average_follow_through_20", "notes"}); err != nil {
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
			floatToString(row.RetestRate),
			floatToString(row.InvalidationRate),
			floatToString(row.AverageProfileVolume),
			floatToString(row.AverageShapeConfidence),
			floatToString(row.AverageFollowThrough5),
			floatToString(row.AverageFollowThrough10),
			floatToString(row.AverageFollowThrough20),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVolumeSetupAccumulationQualitySummaryJSON(path string, summary VolumeSetupAccumulationQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeSetupAccumulationQualityMarkdown(path string, rows []VolumeSetupAccumulationQualityRow, summary VolumeSetupAccumulationQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Setup #1 Accumulation Quality Review\n\n")
	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("- Total setups: %d\n- Accepted: %d\n- Rejected: %d\n- Best confluence group: %s (%.2f)\n- Best shape: %s (%.2f)\n- Best direction by 20-candle follow-through: %s (%.2f)\n\n", summary.TotalSetups, summary.Accepted, summary.Rejected, summary.BestConfluenceGroup, summary.BestConfluenceAcceptanceRate, summary.BestShape, summary.BestShapeAcceptanceRate, summary.BestDirection, summary.BestDirectionAverageFollowThrough20))
	writeQualitySection(&body, "Accepted vs Rejected", rows, "acceptance_state")
	writeQualitySection(&body, "Long vs Short", rows, "direction")
	body.WriteString("## POC Confluence\n\n")
	for _, groupType := range []string{"daily_poc_confluence", "rolling3d_poc_confluence", "rolling7d_poc_confluence", "composite30d_poc_confluence"} {
		writeQualityRows(&body, rows, groupType)
	}
	writeQualitySection(&body, "Shape Quality", rows, "profile_shape")
	writeQualitySection(&body, "Retest Quality", rows, "retest")
	body.WriteString("## What This Means\n\n")
	body.WriteString("This study identifies which accumulation-profile attributes are associated with acceptance and passive follow-through. It does not change setup detection and does not simulate execution.\n\n")
	body.WriteString("## Recommended Filter Improvements Before Setup #2\n\n")
	body.WriteString("- Prefer profile types, shapes, or confluence groups with higher acceptance rates only after manual inspection.\n")
	body.WriteString("- Compare retested and non-retested follow-through before turning retests into a hard filter.\n")
	body.WriteString("- Keep the OHLCV approximation limitation in mind until tick/trade volume-at-price is available.\n")
	return os.WriteFile(path, []byte(body.String()), 0644)
}

type accumulationQualityAccumulator struct {
	groupType       string
	groupName       string
	count           int
	accepted        int
	rejected        int
	retests         int
	invalidated     int
	volumeSum       float64
	confidenceSum   float64
	followThrough5  float64
	followThrough10 float64
	followThrough20 float64
}

func addAccumulationQualityGroup(groups map[string]*accumulationQualityAccumulator, groupType, groupName string) *accumulationQualityAccumulator {
	key := groupType + ":" + groupName
	if groups[key] == nil {
		groups[key] = &accumulationQualityAccumulator{groupType: groupType, groupName: groupName}
	}
	return groups[key]
}

func (a *accumulationQualityAccumulator) add(record accumulationQualityRecord) {
	a.count++
	if record.Accepted {
		a.accepted++
	}
	if record.Rejected {
		a.rejected++
	}
	if record.RetestDetected {
		a.retests++
	}
	if record.Invalidated {
		a.invalidated++
	}
	a.volumeSum += record.ProfileVolume
	a.confidenceSum += record.ShapeConfidence
	a.followThrough5 += record.FollowThrough5
	a.followThrough10 += record.FollowThrough10
	a.followThrough20 += record.FollowThrough20
}

func (a *accumulationQualityAccumulator) row() VolumeSetupAccumulationQualityRow {
	if a.count == 0 {
		return VolumeSetupAccumulationQualityRow{GroupType: a.groupType, GroupName: a.groupName}
	}
	count := float64(a.count)
	return VolumeSetupAccumulationQualityRow{
		GroupType:              a.groupType,
		GroupName:              a.groupName,
		SetupCount:             a.count,
		AcceptedCount:          a.accepted,
		RejectedCount:          a.rejected,
		AcceptanceRate:         float64(a.accepted) / count,
		RejectionRate:          float64(a.rejected) / count,
		RetestRate:             float64(a.retests) / count,
		InvalidationRate:       float64(a.invalidated) / count,
		AverageProfileVolume:   a.volumeSum / count,
		AverageShapeConfidence: a.confidenceSum / count,
		AverageFollowThrough5:  a.followThrough5 / count,
		AverageFollowThrough10: a.followThrough10 / count,
		AverageFollowThrough20: a.followThrough20 / count,
		Notes:                  "Research-only grouping; not a trading filter yet.",
	}
}

func readAccumulationQualityRecords(path string) ([]accumulationQualityRecord, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]accumulationQualityRecord, 0, len(records))
	for _, record := range records {
		out = append(out, accumulationQualityRecord{
			Direction:                 record["direction"],
			ProfileShape:              record["profile_shape"],
			ProfileVolume:             parseFloatRecord(record["profile_volume"]),
			ShapeConfidence:           parseFloatRecord(record["shape_confidence"]),
			Accepted:                  parseBoolRecord(record["accepted"]),
			Rejected:                  parseBoolRecord(record["rejected"]),
			DailyPOCConfluence:        parseBoolRecord(record["daily_poc_confluence"]),
			Rolling3DPOCConfluence:    parseBoolRecord(record["rolling3d_poc_confluence"]),
			Rolling7DPOCConfluence:    parseBoolRecord(record["rolling7d_poc_confluence"]),
			Composite30DPOCConfluence: parseBoolRecord(record["composite30d_poc_confluence"]),
			RetestDetected:            parseBoolRecord(record["retest_detected"]),
			Invalidated:               parseBoolRecord(record["invalidated"]),
			FollowThrough5:            parseFloatRecord(record["follow_through_5"]),
			FollowThrough10:           parseFloatRecord(record["follow_through_10"]),
			FollowThrough20:           parseFloatRecord(record["follow_through_20"]),
		})
	}
	return out, nil
}

func boolGroupName(accepted bool, rejected bool) string {
	switch {
	case accepted:
		return "accepted"
	case rejected:
		return "rejected"
	default:
		return "neutral"
	}
}

func boolName(value bool, yes string, no string) string {
	if value {
		return yes
	}
	return no
}

func writeQualitySection(body *strings.Builder, title string, rows []VolumeSetupAccumulationQualityRow, groupType string) {
	body.WriteString("## " + title + "\n\n")
	writeQualityRows(body, rows, groupType)
}

func writeQualityRows(body *strings.Builder, rows []VolumeSetupAccumulationQualityRow, groupType string) {
	body.WriteString("| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |\n")
	body.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, row := range rows {
		if row.GroupType != groupType {
			continue
		}
		body.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %.2f | %.2f | %.2f | %.2f |\n", row.GroupName, row.SetupCount, row.AcceptedCount, row.RejectedCount, row.AcceptanceRate, row.RetestRate, row.InvalidationRate, row.AverageFollowThrough20))
	}
	body.WriteString("\n")
}
