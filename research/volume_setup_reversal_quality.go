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

type VolumeSetupReversalQualityRow struct {
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

type VolumeSetupReversalQualitySummary struct {
	TotalSetups                  int      `json:"totalSetups"`
	Accepted                     int      `json:"accepted"`
	Rejected                     int      `json:"rejected"`
	Neutral                      int      `json:"neutral"`
	BestDirection                string   `json:"bestDirection"`
	BestDirectionFollowThrough20 float64  `json:"bestDirectionFollowThrough20"`
	BestConfluence               string   `json:"bestConfluence"`
	BestConfluenceAcceptanceRate float64  `json:"bestConfluenceAcceptanceRate"`
	BestFilter                   string   `json:"bestFilter"`
	BestFilterFollowThrough20    float64  `json:"bestFilterFollowThrough20"`
	Notes                        []string `json:"notes"`
}

type reversalQualityRecord struct {
	Direction       string
	ProfileShape    string
	ProfileVolume   float64
	ShapeConfidence float64
	Accepted        bool
	Rejected        bool
	POCConfluence   bool
	HVNConfluence   bool
	VAHVALRejection bool
	VWAPAlignment   string
	Invalidated     bool
	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64
}

func BuildVolumeSetupReversalQualityFromFile(path string) ([]VolumeSetupReversalQualityRow, VolumeSetupReversalQualitySummary, error) {
	records, err := readReversalQualityRecords(path)
	if err != nil {
		return nil, VolumeSetupReversalQualitySummary{}, err
	}
	rows := BuildVolumeSetupReversalQualityRows(records)
	summary := BuildVolumeSetupReversalQualitySummary(records, rows)
	return rows, summary, nil
}

func BuildVolumeSetupReversalQualityRows(records []reversalQualityRecord) []VolumeSetupReversalQualityRow {
	groups := map[string]*reversalQualityAccumulator{}
	for _, record := range records {
		addReversalQualityGroup(groups, "acceptance_state", boolGroupName(record.Accepted, record.Rejected)).add(record)
		addReversalQualityGroup(groups, "direction", record.Direction).add(record)
		addReversalQualityGroup(groups, "poc_confluence", boolName(record.POCConfluence, "with_poc_confluence", "without_poc_confluence")).add(record)
		addReversalQualityGroup(groups, "hvn_confluence", boolName(record.HVNConfluence, "with_hvn_confluence", "without_hvn_confluence")).add(record)
		addReversalQualityGroup(groups, "vah_val_rejection", boolName(record.VAHVALRejection, "with_vah_val_rejection", "without_vah_val_rejection")).add(record)
		addReversalQualityGroup(groups, "vwap_alignment", boolName(reversalRecordVWAPAligned(record), "vwap_aligned", "vwap_not_aligned")).add(record)
		addReversalQualityGroup(groups, "profile_shape", profileShapeOrUnknown(record.ProfileShape)).add(record)
	}
	rows := make([]VolumeSetupReversalQualityRow, 0, len(groups))
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

func BuildVolumeSetupReversalQualitySummary(records []reversalQualityRecord, rows []VolumeSetupReversalQualityRow) VolumeSetupReversalQualitySummary {
	summary := VolumeSetupReversalQualitySummary{
		TotalSetups: len(records),
		Notes: []string{
			"Research-only quality analysis; reversal setup detection rules are unchanged.",
			"Reversal accepted/rejected labels are outcome labels based on invalidation and passive 20-candle follow-through.",
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
		if !record.Accepted && !record.Rejected {
			summary.Neutral++
		}
	}
	for _, row := range rows {
		if row.GroupType == "direction" && (summary.BestDirection == "" || row.AverageFollowThrough20 > summary.BestDirectionFollowThrough20) {
			summary.BestDirection = row.GroupName
			summary.BestDirectionFollowThrough20 = row.AverageFollowThrough20
		}
		if isReversalConfluenceGroup(row) && strings.HasPrefix(row.GroupName, "with_") && row.AcceptanceRate > summary.BestConfluenceAcceptanceRate {
			summary.BestConfluence = row.GroupType + ":" + row.GroupName
			summary.BestConfluenceAcceptanceRate = row.AcceptanceRate
		}
		if isReversalFilterGroup(row) && (summary.BestFilter == "" || row.AverageFollowThrough20 > summary.BestFilterFollowThrough20) {
			summary.BestFilter = row.GroupType + ":" + row.GroupName
			summary.BestFilterFollowThrough20 = row.AverageFollowThrough20
		}
	}
	return summary
}

func WriteVolumeSetupReversalQualityCSV(path string, rows []VolumeSetupReversalQualityRow) error {
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
			row.GroupType,
			row.GroupName,
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

func WriteVolumeSetupReversalQualitySummaryJSON(path string, summary VolumeSetupReversalQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeSetupReversalQualityMarkdown(path string, rows []VolumeSetupReversalQualityRow, summary VolumeSetupReversalQualitySummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Setup Reversal Quality Review\n\n")
	body.WriteString("## Summary\n\n")
	body.WriteString(fmt.Sprintf("- Total setups: %d\n- Accepted: %d\n- Rejected: %d\n- Neutral: %d\n- Best direction by 20-candle follow-through: %s (%.2f)\n- Best confluence by acceptance: %s (%.2f)\n- Best filter by 20-candle follow-through: %s (%.2f)\n\n", summary.TotalSetups, summary.Accepted, summary.Rejected, summary.Neutral, summary.BestDirection, summary.BestDirectionFollowThrough20, summary.BestConfluence, summary.BestConfluenceAcceptanceRate, summary.BestFilter, summary.BestFilterFollowThrough20))
	writeReversalQualitySection(&body, "Accepted vs Rejected", rows, "acceptance_state")
	writeReversalQualitySection(&body, "Long vs Short", rows, "direction")
	writeReversalQualitySection(&body, "POC Confluence", rows, "poc_confluence")
	writeReversalQualitySection(&body, "HVN Confluence", rows, "hvn_confluence")
	writeReversalQualitySection(&body, "VAH/VAL Rejection", rows, "vah_val_rejection")
	writeReversalQualitySection(&body, "VWAP Alignment", rows, "vwap_alignment")
	writeReversalQualitySection(&body, "Profile Shape", rows, "profile_shape")
	body.WriteString("## What This Means\n\n")
	body.WriteString("This report fixes the previous zero-state outcome gap by assigning explicit accepted, rejected, or neutral labels to every reversal setup. It does not alter reversal detection, POC/HVN logic, VWAP logic, or any execution behavior.\n\n")
	body.WriteString("## Updated Comparison Readiness\n\n")
	body.WriteString("Reversal can now be compared against Accumulation, Trend, and Rejection on acceptance, rejection, invalidation, and passive follow-through metrics.\n")
	return os.WriteFile(path, []byte(body.String()), 0644)
}

type reversalQualityAccumulator struct {
	groupType       string
	groupName       string
	count           int
	accepted        int
	rejected        int
	invalidated     int
	volume          float64
	shapeConfidence float64
	ft5             float64
	ft10            float64
	ft20            float64
}

func addReversalQualityGroup(groups map[string]*reversalQualityAccumulator, groupType string, groupName string) *reversalQualityAccumulator {
	key := groupType + ":" + groupName
	if groups[key] == nil {
		groups[key] = &reversalQualityAccumulator{groupType: groupType, groupName: groupName}
	}
	return groups[key]
}

func (a *reversalQualityAccumulator) add(record reversalQualityRecord) {
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
	a.volume += record.ProfileVolume
	a.shapeConfidence += record.ShapeConfidence
	a.ft5 += record.FollowThrough5
	a.ft10 += record.FollowThrough10
	a.ft20 += record.FollowThrough20
}

func (a reversalQualityAccumulator) row() VolumeSetupReversalQualityRow {
	row := VolumeSetupReversalQualityRow{
		GroupType:  a.groupType,
		GroupName:  a.groupName,
		SetupCount: a.count,
		Notes:      "Research-only reversal outcome quality group.",
	}
	if a.count == 0 {
		return row
	}
	count := float64(a.count)
	row.AcceptedCount = a.accepted
	row.RejectedCount = a.rejected
	row.AcceptanceRate = float64(a.accepted) / count
	row.RejectionRate = float64(a.rejected) / count
	row.AverageProfileVolume = a.volume / count
	row.AverageShapeConfidence = a.shapeConfidence / count
	row.AverageFollowThrough5 = a.ft5 / count
	row.AverageFollowThrough10 = a.ft10 / count
	row.AverageFollowThrough20 = a.ft20 / count
	row.InvalidationRate = float64(a.invalidated) / count
	return row
}

func readReversalQualityRecords(path string) ([]reversalQualityRecord, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]reversalQualityRecord, 0, len(records))
	for _, record := range records {
		out = append(out, reversalQualityRecord{
			Direction:       record["direction"],
			ProfileShape:    record["profile_shape"],
			ProfileVolume:   parseFloatRecord(record["profile_volume"]),
			ShapeConfidence: parseFloatRecord(record["shape_confidence"]),
			Accepted:        parseBoolRecord(record["accepted"]),
			Rejected:        parseBoolRecord(record["rejected"]),
			POCConfluence:   parseBoolRecord(record["poc_confluence"]),
			HVNConfluence:   parseBoolRecord(record["hvn_confluence"]),
			VAHVALRejection: parseBoolRecord(record["vah_val_rejection"]),
			VWAPAlignment:   record["vwap_alignment"],
			Invalidated:     parseBoolRecord(record["invalidated"]),
			FollowThrough5:  parseFloatRecord(record["follow_through_5"]),
			FollowThrough10: parseFloatRecord(record["follow_through_10"]),
			FollowThrough20: parseFloatRecord(record["follow_through_20"]),
		})
	}
	return out, nil
}

func reversalRecordVWAPAligned(record reversalQualityRecord) bool {
	return vwapAligned(record.Direction, record.VWAPAlignment)
}

func isReversalConfluenceGroup(row VolumeSetupReversalQualityRow) bool {
	switch row.GroupType {
	case "poc_confluence", "hvn_confluence", "vah_val_rejection", "vwap_alignment":
		return true
	default:
		return false
	}
}

func isReversalFilterGroup(row VolumeSetupReversalQualityRow) bool {
	switch row.GroupType {
	case "poc_confluence", "hvn_confluence", "vah_val_rejection", "vwap_alignment", "profile_shape":
		return true
	default:
		return false
	}
}

func writeReversalQualitySection(body *strings.Builder, title string, rows []VolumeSetupReversalQualityRow, groupType string) {
	body.WriteString("## " + title + "\n\n")
	writeReversalQualityRows(body, rows, groupType)
}

func writeReversalQualityRows(body *strings.Builder, rows []VolumeSetupReversalQualityRow, groupType string) {
	body.WriteString("| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |\n")
	body.WriteString("|---|---:|---:|---:|---:|---:|\n")
	for _, row := range rows {
		if row.GroupType != groupType {
			continue
		}
		body.WriteString(fmt.Sprintf("| %s | %d | %.2f | %.2f | %.2f | %.2f |\n", row.GroupName, row.SetupCount, row.AcceptanceRate, row.RejectionRate, row.AverageFollowThrough20, row.InvalidationRate))
	}
	body.WriteString("\n")
}
