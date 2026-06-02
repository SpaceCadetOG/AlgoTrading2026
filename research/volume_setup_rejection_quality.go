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

type VolumeSetupRejectionQualityRow struct {
	GroupType              string  `json:"groupType"`
	GroupName              string  `json:"groupName"`
	SetupCount             int     `json:"setupCount"`
	AcceptedCount          int     `json:"acceptedCount"`
	RejectedCount          int     `json:"rejectedCount"`
	AcceptanceRate         float64 `json:"acceptanceRate"`
	RejectionRate          float64 `json:"rejectionRate"`
	AverageProfileVolume   float64 `json:"averageProfileVolume"`
	AverageShapeConfidence float64 `json:"averageShapeConfidence"`
	AverageFollowThrough5  float64 `json:"averageFollowThrough5"`
	AverageFollowThrough10 float64 `json:"averageFollowThrough10"`
	AverageFollowThrough20 float64 `json:"averageFollowThrough20"`
	InvalidationRate       float64 `json:"invalidationRate"`
	Notes                  string  `json:"notes"`
}

type VolumeSetupRejectionQualitySummary struct {
	TotalSetups                  int      `json:"totalSetups"`
	Accepted                     int      `json:"accepted"`
	Rejected                     int      `json:"rejected"`
	BestDirection                string   `json:"bestDirection"`
	BestDirectionFollowThrough20 float64  `json:"bestDirectionFollowThrough20"`
	BestShape                    string   `json:"bestShape"`
	BestShapeAcceptanceRate      float64  `json:"bestShapeAcceptanceRate"`
	BestConfluence               string   `json:"bestConfluence"`
	BestConfluenceAcceptanceRate float64  `json:"bestConfluenceAcceptanceRate"`
	BestFilter                   string   `json:"bestFilter"`
	BestFilterFollowThrough20    float64  `json:"bestFilterFollowThrough20"`
	Notes                        []string `json:"notes"`
}

type rejectionQualityRecord struct {
	Direction                 string
	ProfileShape              string
	ProfileVolume             float64
	ShapeConfidence           float64
	Accepted                  bool
	Rejected                  bool
	POCRetestDetected         bool
	HVNRetestDetected         bool
	VWAPAlignment             string
	DailyPOCConfluence        bool
	Rolling3DPOCConfluence    bool
	Rolling7DPOCConfluence    bool
	Composite30DPOCConfluence bool
	Invalidated               bool
	FollowThrough5            float64
	FollowThrough10           float64
	FollowThrough20           float64
}

func BuildVolumeSetupRejectionQualityFromFile(path string) ([]VolumeSetupRejectionQualityRow, VolumeSetupRejectionQualitySummary, error) {
	records, err := readRejectionQualityRecords(path)
	if err != nil {
		return nil, VolumeSetupRejectionQualitySummary{}, err
	}
	rows := BuildVolumeSetupRejectionQualityRows(records)
	summary := BuildVolumeSetupRejectionQualitySummary(records, rows)
	return rows, summary, nil
}

func BuildVolumeSetupRejectionQualityRows(records []rejectionQualityRecord) []VolumeSetupRejectionQualityRow {
	groups := map[string]*rejectionQualityAccumulator{}
	for _, record := range records {
		addRejectionQualityGroup(groups, "acceptance_state", boolGroupName(record.Accepted, record.Rejected)).add(record)
		addRejectionQualityGroup(groups, "direction", record.Direction).add(record)
		addRejectionQualityGroup(groups, "vwap_alignment", boolName(rejectionRecordVWAPAligned(record), "vwap_aligned", "vwap_not_aligned")).add(record)
		addRejectionQualityGroup(groups, "poc_retest", boolName(record.POCRetestDetected, "with_poc_retest", "without_poc_retest")).add(record)
		addRejectionQualityGroup(groups, "hvn_retest", boolName(record.HVNRetestDetected, "with_hvn_retest", "without_hvn_retest")).add(record)
		addRejectionQualityGroup(groups, "daily_poc_confluence", boolName(record.DailyPOCConfluence, "with_daily_poc", "without_daily_poc")).add(record)
		addRejectionQualityGroup(groups, "rolling3d_poc_confluence", boolName(record.Rolling3DPOCConfluence, "with_rolling3d_poc", "without_rolling3d_poc")).add(record)
		addRejectionQualityGroup(groups, "rolling7d_poc_confluence", boolName(record.Rolling7DPOCConfluence, "with_rolling7d_poc", "without_rolling7d_poc")).add(record)
		addRejectionQualityGroup(groups, "composite30d_poc_confluence", boolName(record.Composite30DPOCConfluence, "with_composite30d_poc", "without_composite30d_poc")).add(record)
		addRejectionQualityGroup(groups, "profile_shape", profileShapeOrUnknown(record.ProfileShape)).add(record)
	}
	rows := make([]VolumeSetupRejectionQualityRow, 0, len(groups))
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

func BuildVolumeSetupRejectionQualitySummary(records []rejectionQualityRecord, rows []VolumeSetupRejectionQualityRow) VolumeSetupRejectionQualitySummary {
	summary := VolumeSetupRejectionQualitySummary{
		TotalSetups: len(records),
		Notes: []string{
			"Research-only quality analysis; rejection setup detection rules are unchanged.",
			"Acceptance/rejection comes from rejection-profile value-area behavior.",
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
		if isRejectionConfluenceGroup(row) && strings.HasPrefix(row.GroupName, "with_") && row.AcceptanceRate > summary.BestConfluenceAcceptanceRate {
			summary.BestConfluence = row.GroupType + ":" + row.GroupName
			summary.BestConfluenceAcceptanceRate = row.AcceptanceRate
		}
		if isRejectionFilterGroup(row) && (summary.BestFilter == "" || row.AverageFollowThrough20 > summary.BestFilterFollowThrough20) {
			summary.BestFilter = row.GroupType + ":" + row.GroupName
			summary.BestFilterFollowThrough20 = row.AverageFollowThrough20
		}
	}
	return summary
}

func WriteVolumeSetupRejectionQualityCSV(path string, rows []VolumeSetupRejectionQualityRow) error {
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
	if err := writer.Write([]string{"group_type", "group_name", "setup_count", "accepted_count", "rejected_count", "acceptance_rate", "rejection_rate", "average_profile_volume", "average_shape_confidence", "average_follow_through_5", "average_follow_through_10", "average_follow_through_20", "invalidation_rate", "notes"}); err != nil {
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

func WriteVolumeSetupRejectionQualitySummaryJSON(path string, summary VolumeSetupRejectionQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeSetupRejectionQualityMarkdown(path string, rows []VolumeSetupRejectionQualityRow, summary VolumeSetupRejectionQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Setup #3 Rejection Quality Review\n\n")
	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("- Total setups: %d\n- Accepted: %d\n- Rejected: %d\n- Best direction by 20-candle follow-through: %s (%.2f)\n- Best shape by acceptance: %s (%.2f)\n- Best confluence by acceptance: %s (%.2f)\n- Best filter by 20-candle follow-through: %s (%.2f)\n\n", summary.TotalSetups, summary.Accepted, summary.Rejected, summary.BestDirection, summary.BestDirectionFollowThrough20, summary.BestShape, summary.BestShapeAcceptanceRate, summary.BestConfluence, summary.BestConfluenceAcceptanceRate, summary.BestFilter, summary.BestFilterFollowThrough20))
	writeRejectionQualitySection(&body, "Accepted vs Rejected", rows, "acceptance_state")
	writeRejectionQualitySection(&body, "Long vs Short", rows, "direction")
	writeRejectionQualitySection(&body, "VWAP Alignment", rows, "vwap_alignment")
	writeRejectionQualitySection(&body, "POC Retest Quality", rows, "poc_retest")
	writeRejectionQualitySection(&body, "HVN Retest Quality", rows, "hvn_retest")
	body.WriteString("## POC Confluence\n\n")
	for _, groupType := range []string{"daily_poc_confluence", "rolling3d_poc_confluence", "rolling7d_poc_confluence", "composite30d_poc_confluence"} {
		writeRejectionQualityRows(&body, rows, groupType)
	}
	writeRejectionQualitySection(&body, "Profile Shape Quality", rows, "profile_shape")
	body.WriteString("## What This Means\n\n")
	body.WriteString("This report identifies which rejection-profile attributes are associated with acceptance and passive follow-through. It does not change the rejection setup detector and does not simulate execution.\n\n")
	body.WriteString("## Comparison With Setup #1 and Setup #2\n\n")
	body.WriteString("Setup #1 studies accumulated volume after sideways positioning, Setup #2 studies trend-leg continuation, and Setup #3 studies reversal context after strong price rejection. These reports should be compared manually before building a Reversal Trade study.\n\n")
	body.WriteString("## Recommendation Before Reversal Trade\n\n")
	body.WriteString("- Inspect filters with positive or less-negative follow-through before using rejection setups as reversal candidates.\n")
	body.WriteString("- Treat high setup count and low acceptance as a warning that raw rejection detection is too broad for execution research.\n")
	body.WriteString("- Keep L2 context limitations in mind until historical L2 replay is available.\n")
	return os.WriteFile(path, []byte(body.String()), 0644)
}

type rejectionQualityAccumulator struct {
	groupType       string
	groupName       string
	count           int
	accepted        int
	rejected        int
	invalidated     int
	volumeSum       float64
	confidenceSum   float64
	followThrough5  float64
	followThrough10 float64
	followThrough20 float64
}

func addRejectionQualityGroup(groups map[string]*rejectionQualityAccumulator, groupType, groupName string) *rejectionQualityAccumulator {
	key := groupType + ":" + groupName
	if groups[key] == nil {
		groups[key] = &rejectionQualityAccumulator{groupType: groupType, groupName: groupName}
	}
	return groups[key]
}

func (a *rejectionQualityAccumulator) add(record rejectionQualityRecord) {
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
	a.volumeSum += record.ProfileVolume
	a.confidenceSum += record.ShapeConfidence
	a.followThrough5 += record.FollowThrough5
	a.followThrough10 += record.FollowThrough10
	a.followThrough20 += record.FollowThrough20
}

func (a *rejectionQualityAccumulator) row() VolumeSetupRejectionQualityRow {
	if a.count == 0 {
		return VolumeSetupRejectionQualityRow{GroupType: a.groupType, GroupName: a.groupName}
	}
	count := float64(a.count)
	return VolumeSetupRejectionQualityRow{
		GroupType:              a.groupType,
		GroupName:              a.groupName,
		SetupCount:             a.count,
		AcceptedCount:          a.accepted,
		RejectedCount:          a.rejected,
		AcceptanceRate:         float64(a.accepted) / count,
		RejectionRate:          float64(a.rejected) / count,
		AverageProfileVolume:   a.volumeSum / count,
		AverageShapeConfidence: a.confidenceSum / count,
		AverageFollowThrough5:  a.followThrough5 / count,
		AverageFollowThrough10: a.followThrough10 / count,
		AverageFollowThrough20: a.followThrough20 / count,
		InvalidationRate:       float64(a.invalidated) / count,
		Notes:                  "Research-only grouping; not a trading filter yet.",
	}
}

func readRejectionQualityRecords(path string) ([]rejectionQualityRecord, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]rejectionQualityRecord, 0, len(records))
	for _, record := range records {
		out = append(out, rejectionQualityRecord{
			Direction:                 record["direction"],
			ProfileShape:              record["profile_shape"],
			ProfileVolume:             parseFloatRecord(record["profile_volume"]),
			ShapeConfidence:           parseFloatRecord(record["shape_confidence"]),
			Accepted:                  parseBoolRecord(record["accepted"]),
			Rejected:                  parseBoolRecord(record["rejected"]),
			POCRetestDetected:         parseBoolRecord(record["poc_retest_detected"]),
			HVNRetestDetected:         parseBoolRecord(record["hvn_retest_detected"]),
			VWAPAlignment:             record["vwap_alignment"],
			DailyPOCConfluence:        parseBoolRecord(record["daily_poc_confluence"]),
			Rolling3DPOCConfluence:    parseBoolRecord(record["rolling3d_poc_confluence"]),
			Rolling7DPOCConfluence:    parseBoolRecord(record["rolling7d_poc_confluence"]),
			Composite30DPOCConfluence: parseBoolRecord(record["composite30d_poc_confluence"]),
			Invalidated:               parseBoolRecord(record["invalidated"]),
			FollowThrough5:            parseFloatRecord(record["follow_through_5"]),
			FollowThrough10:           parseFloatRecord(record["follow_through_10"]),
			FollowThrough20:           parseFloatRecord(record["follow_through_20"]),
		})
	}
	return out, nil
}

func rejectionRecordVWAPAligned(record rejectionQualityRecord) bool {
	return vwapAligned(record.Direction, record.VWAPAlignment)
}

func isRejectionConfluenceGroup(row VolumeSetupRejectionQualityRow) bool {
	return strings.Contains(row.GroupType, "poc_confluence")
}

func isRejectionFilterGroup(row VolumeSetupRejectionQualityRow) bool {
	switch row.GroupType {
	case "vwap_alignment", "poc_retest", "hvn_retest", "profile_shape", "daily_poc_confluence", "rolling3d_poc_confluence", "rolling7d_poc_confluence", "composite30d_poc_confluence":
		return true
	default:
		return false
	}
}

func writeRejectionQualitySection(body *strings.Builder, title string, rows []VolumeSetupRejectionQualityRow, groupType string) {
	body.WriteString("## " + title + "\n\n")
	writeRejectionQualityRows(body, rows, groupType)
}

func writeRejectionQualityRows(body *strings.Builder, rows []VolumeSetupRejectionQualityRow, groupType string) {
	body.WriteString("| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |\n")
	body.WriteString("|---|---:|---:|---:|---:|---:|---:|\n")
	for _, row := range rows {
		if row.GroupType != groupType {
			continue
		}
		body.WriteString(fmt.Sprintf("| %s | %d | %d | %d | %.2f | %.2f | %.2f |\n", row.GroupName, row.SetupCount, row.AcceptedCount, row.RejectedCount, row.AcceptanceRate, row.InvalidationRate, row.AverageFollowThrough20))
	}
	body.WriteString("\n")
}
