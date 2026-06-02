package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type VolumeSetupAccumulationRow struct {
	Timestamp                 int64
	SetupID                   string
	SetupType                 string
	Direction                 string
	AccumulationStart         int64
	AccumulationEnd           int64
	SetupLevel                float64
	POC                       float64
	VAH                       float64
	VAL                       float64
	ProfileShape              string
	ShapeConfidence           float64
	ProfileVolume             float64
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
	Notes                     string
}

func BuildVolumeSetupAccumulationFromFiles(flexiblePath string, priceActionPath string, strategyPath string, scopedPath string, acceptancePath string, cfg volumeprofile.AccumulationSetupConfig) ([]VolumeSetupAccumulationRow, VolumeSetupAccumulationSummary, error) {
	profiles, err := readAccumulationProfiles(flexiblePath)
	if err != nil {
		return nil, VolumeSetupAccumulationSummary{}, err
	}
	points, err := readAccumulationPricePoints(priceActionPath)
	if err != nil {
		return nil, VolumeSetupAccumulationSummary{}, err
	}
	if _, err := readCSVRecords(strategyPath); err != nil {
		return nil, VolumeSetupAccumulationSummary{}, err
	}
	if _, err := readCSVRecords(acceptancePath); err != nil {
		return nil, VolumeSetupAccumulationSummary{}, err
	}
	scoped, err := readScopedPOCLevels(scopedPath)
	if err != nil {
		return nil, VolumeSetupAccumulationSummary{}, err
	}
	setups := volumeprofile.DetectAccumulationSetups(profiles, points, scoped, cfg)
	rows := BuildVolumeSetupAccumulationRows(setups)
	summary := BuildVolumeSetupAccumulationSummary(rows)
	return rows, summary, nil
}

func BuildVolumeSetupAccumulationRows(setups []volumeprofile.AccumulationSetup) []VolumeSetupAccumulationRow {
	rows := make([]VolumeSetupAccumulationRow, 0, len(setups))
	for _, setup := range setups {
		rows = append(rows, VolumeSetupAccumulationRow{
			Timestamp:                 setup.Timestamp,
			SetupID:                   setup.SetupID,
			SetupType:                 setup.SetupType,
			Direction:                 setup.Direction,
			AccumulationStart:         setup.AccumulationStart,
			AccumulationEnd:           setup.AccumulationEnd,
			SetupLevel:                setup.SetupLevel,
			POC:                       setup.POC,
			VAH:                       setup.VAH,
			VAL:                       setup.VAL,
			ProfileShape:              setup.ProfileShape,
			ShapeConfidence:           setup.ShapeConfidence,
			ProfileVolume:             setup.ProfileVolume,
			Accepted:                  setup.Accepted,
			Rejected:                  setup.Rejected,
			DailyPOCConfluence:        setup.DailyPOCConfluence,
			Rolling3DPOCConfluence:    setup.Rolling3DPOCConfluence,
			Rolling7DPOCConfluence:    setup.Rolling7DPOCConfluence,
			Composite30DPOCConfluence: setup.Composite30DPOCConfluence,
			RetestDetected:            setup.RetestDetected,
			Invalidated:               setup.Invalidated,
			FollowThrough5:            setup.FollowThrough5,
			FollowThrough10:           setup.FollowThrough10,
			FollowThrough20:           setup.FollowThrough20,
			Notes:                     setup.Notes,
		})
	}
	return rows
}

func WriteVolumeSetupAccumulationCSV(path string, rows []VolumeSetupAccumulationRow) error {
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
		"timestamp",
		"setup_id",
		"setup_type",
		"direction",
		"accumulation_start",
		"accumulation_end",
		"setup_level",
		"poc",
		"vah",
		"val",
		"profile_shape",
		"shape_confidence",
		"profile_volume",
		"accepted",
		"rejected",
		"daily_poc_confluence",
		"rolling3d_poc_confluence",
		"rolling7d_poc_confluence",
		"composite30d_poc_confluence",
		"retest_detected",
		"invalidated",
		"follow_through_5",
		"follow_through_10",
		"follow_through_20",
		"notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.SetupID,
			row.SetupType,
			row.Direction,
			strconv.FormatInt(row.AccumulationStart, 10),
			strconv.FormatInt(row.AccumulationEnd, 10),
			floatToString(row.SetupLevel),
			floatToString(row.POC),
			floatToString(row.VAH),
			floatToString(row.VAL),
			row.ProfileShape,
			floatToString(row.ShapeConfidence),
			floatToString(row.ProfileVolume),
			strconv.FormatBool(row.Accepted),
			strconv.FormatBool(row.Rejected),
			strconv.FormatBool(row.DailyPOCConfluence),
			strconv.FormatBool(row.Rolling3DPOCConfluence),
			strconv.FormatBool(row.Rolling7DPOCConfluence),
			strconv.FormatBool(row.Composite30DPOCConfluence),
			strconv.FormatBool(row.RetestDetected),
			strconv.FormatBool(row.Invalidated),
			floatToString(row.FollowThrough5),
			floatToString(row.FollowThrough10),
			floatToString(row.FollowThrough20),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func readAccumulationProfiles(path string) ([]volumeprofile.AccumulationProfileInput, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]volumeprofile.AccumulationProfileInput, 0)
	for _, record := range records {
		if record["profile_type"] != volumeprofile.FlexibleSidewaysAccumulation {
			continue
		}
		out = append(out, volumeprofile.AccumulationProfileInput{
			ProfileID:       record["profile_id"],
			StartTime:       parseIntRecord(record["start_time"]),
			EndTime:         parseIntRecord(record["end_time"]),
			Symbol:          record["symbol"],
			POC:             parseFloatRecord(record["poc"]),
			VAH:             parseFloatRecord(record["vah"]),
			VAL:             parseFloatRecord(record["val"]),
			ProfileShape:    record["profile_shape"],
			ShapeConfidence: parseFloatRecord(record["shape_confidence"]),
			ProfileVolume:   parseFloatRecord(record["profile_volume"]),
			AcceptanceState: record["acceptance_state"],
		})
	}
	return out, nil
}

func readAccumulationPricePoints(path string) ([]volumeprofile.AccumulationPricePoint, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]volumeprofile.AccumulationPricePoint, 0, len(records))
	for _, record := range records {
		out = append(out, volumeprofile.AccumulationPricePoint{
			Timestamp:           parseIntRecord(record["timestamp"]),
			High:                parseFloatRecord(record["high"]),
			Low:                 parseFloatRecord(record["low"]),
			Close:               parseFloatRecord(record["close"]),
			InitiationDetected:  parseBoolRecord(record["initiation_detected"]),
			InitiationDirection: record["initiation_direction"],
		})
	}
	return out, nil
}

func readScopedPOCLevels(path string) (volumeprofile.ScopedPOCLevels, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return volumeprofile.ScopedPOCLevels{}, err
	}
	var levels volumeprofile.ScopedPOCLevels
	seen := map[string]bool{}
	for _, record := range records {
		scope := record["scope"]
		if seen[scope] {
			continue
		}
		seen[scope] = true
		poc := parseFloatRecord(record["poc"])
		switch scope {
		case "daily_session":
			levels.Daily = poc
		case "rolling_3d":
			levels.Rolling3D = poc
		case "rolling_7d":
			levels.Rolling7D = poc
		case "composite_30d":
			levels.Composite30D = poc
		}
	}
	return levels, nil
}

func parseBoolRecord(value string) bool {
	parsed, _ := strconv.ParseBool(value)
	return parsed
}

func WriteVolumeSetupAccumulationSummaryJSON(path string, summary VolumeSetupAccumulationSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
