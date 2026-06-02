package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type VolumeSetupRejectionRow struct {
	Timestamp                 int64
	SetupID                   string
	SetupType                 string
	Direction                 string
	RejectionStart            int64
	RejectionEnd              int64
	RejectionLevel            float64
	RejectionHigh             float64
	RejectionLow              float64
	SetupLevel                float64
	POC                       float64
	VAH                       float64
	VAL                       float64
	NearestHVN                float64
	NearestLVN                float64
	ProfileShape              string
	ShapeConfidence           float64
	ProfileVolume             float64
	Accepted                  bool
	Rejected                  bool
	POCRetestDetected         bool
	HVNRetestDetected         bool
	VWAPAlignment             string
	L2BookPressure            string
	DailyPOCConfluence        bool
	Rolling3DPOCConfluence    bool
	Rolling7DPOCConfluence    bool
	Composite30DPOCConfluence bool
	Invalidated               bool
	FollowThrough5            float64
	FollowThrough10           float64
	FollowThrough20           float64
	Notes                     string
}

func BuildVolumeSetupRejectionFromFiles(priceActionPath string, phase3Path string, flexiblePath string, scopedPath string, acceptancePath string, vwapPath string, contextPath string, contextL2Path string, cfg volumeprofile.RejectionSetupConfig) ([]VolumeSetupRejectionRow, VolumeSetupRejectionSummary, error) {
	points, err := readRejectionPricePoints(priceActionPath)
	if err != nil {
		return nil, VolumeSetupRejectionSummary{}, err
	}
	if _, err := readCSVRecords(phase3Path); err != nil {
		return nil, VolumeSetupRejectionSummary{}, err
	}
	if _, err := readCSVRecords(flexiblePath); err != nil {
		return nil, VolumeSetupRejectionSummary{}, err
	}
	if _, err := readCSVRecords(acceptancePath); err != nil {
		return nil, VolumeSetupRejectionSummary{}, err
	}
	if vwapPath != "" {
		applyRejectionVWAPAlignment(points, vwapPath)
	}
	if contextPath != "" {
		applyRejectionContextAlignment(points, contextPath)
	}
	if contextL2Path != "" {
		applyRejectionL2Context(points, contextL2Path)
	}
	if pressure := latestL2BookPressure("research/multi_venue_orderbook_features.csv"); pressure != "unknown" {
		for i := range points {
			if points[i].L2BookPressure == "" || points[i].L2BookPressure == "unknown" {
				points[i].L2BookPressure = pressure
			}
		}
	}
	scoped, err := readScopedPOCLevels(scopedPath)
	if err != nil {
		return nil, VolumeSetupRejectionSummary{}, err
	}
	setups := volumeprofile.DetectRejectionSetups(points, scoped, cfg)
	rows := BuildVolumeSetupRejectionRows(setups)
	summary := BuildVolumeSetupRejectionSummary(rows)
	return rows, summary, nil
}

func BuildVolumeSetupRejectionRows(setups []volumeprofile.RejectionSetup) []VolumeSetupRejectionRow {
	rows := make([]VolumeSetupRejectionRow, 0, len(setups))
	for _, setup := range setups {
		rows = append(rows, VolumeSetupRejectionRow{
			Timestamp:                 setup.Timestamp,
			SetupID:                   setup.SetupID,
			SetupType:                 setup.SetupType,
			Direction:                 setup.Direction,
			RejectionStart:            setup.RejectionStart,
			RejectionEnd:              setup.RejectionEnd,
			RejectionLevel:            setup.RejectionLevel,
			RejectionHigh:             setup.RejectionHigh,
			RejectionLow:              setup.RejectionLow,
			SetupLevel:                setup.SetupLevel,
			POC:                       setup.POC,
			VAH:                       setup.VAH,
			VAL:                       setup.VAL,
			NearestHVN:                setup.NearestHVN,
			NearestLVN:                setup.NearestLVN,
			ProfileShape:              setup.ProfileShape,
			ShapeConfidence:           setup.ShapeConfidence,
			ProfileVolume:             setup.ProfileVolume,
			Accepted:                  setup.Accepted,
			Rejected:                  setup.Rejected,
			POCRetestDetected:         setup.POCRetestDetected,
			HVNRetestDetected:         setup.HVNRetestDetected,
			VWAPAlignment:             setup.VWAPAlignment,
			L2BookPressure:            setup.L2BookPressure,
			DailyPOCConfluence:        setup.DailyPOCConfluence,
			Rolling3DPOCConfluence:    setup.Rolling3DPOCConfluence,
			Rolling7DPOCConfluence:    setup.Rolling7DPOCConfluence,
			Composite30DPOCConfluence: setup.Composite30DPOCConfluence,
			Invalidated:               setup.Invalidated,
			FollowThrough5:            setup.FollowThrough5,
			FollowThrough10:           setup.FollowThrough10,
			FollowThrough20:           setup.FollowThrough20,
			Notes:                     setup.Notes,
		})
	}
	return rows
}

func WriteVolumeSetupRejectionCSV(path string, rows []VolumeSetupRejectionRow) error {
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
		"timestamp", "setup_id", "setup_type", "direction", "rejection_start", "rejection_end", "rejection_level", "rejection_high", "rejection_low",
		"setup_level", "poc", "vah", "val", "nearest_hvn", "nearest_lvn", "profile_shape", "shape_confidence", "profile_volume",
		"accepted", "rejected", "poc_retest_detected", "hvn_retest_detected", "vwap_alignment", "l2_book_pressure",
		"daily_poc_confluence", "rolling3d_poc_confluence", "rolling7d_poc_confluence", "composite30d_poc_confluence",
		"invalidated", "follow_through_5", "follow_through_10", "follow_through_20", "notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10), row.SetupID, row.SetupType, row.Direction,
			strconv.FormatInt(row.RejectionStart, 10), strconv.FormatInt(row.RejectionEnd, 10),
			floatToString(row.RejectionLevel), floatToString(row.RejectionHigh), floatToString(row.RejectionLow),
			floatToString(row.SetupLevel), floatToString(row.POC), floatToString(row.VAH), floatToString(row.VAL),
			floatToString(row.NearestHVN), floatToString(row.NearestLVN), row.ProfileShape, floatToString(row.ShapeConfidence), floatToString(row.ProfileVolume),
			strconv.FormatBool(row.Accepted), strconv.FormatBool(row.Rejected), strconv.FormatBool(row.POCRetestDetected), strconv.FormatBool(row.HVNRetestDetected),
			row.VWAPAlignment, row.L2BookPressure,
			strconv.FormatBool(row.DailyPOCConfluence), strconv.FormatBool(row.Rolling3DPOCConfluence), strconv.FormatBool(row.Rolling7DPOCConfluence), strconv.FormatBool(row.Composite30DPOCConfluence),
			strconv.FormatBool(row.Invalidated), floatToString(row.FollowThrough5), floatToString(row.FollowThrough10), floatToString(row.FollowThrough20), row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVolumeSetupRejectionSummaryJSON(path string, summary VolumeSetupRejectionSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func readRejectionPricePoints(path string) ([]volumeprofile.RejectionPricePoint, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]volumeprofile.RejectionPricePoint, 0, len(records))
	for _, record := range records {
		out = append(out, volumeprofile.RejectionPricePoint{
			Timestamp:           parseIntRecord(record["timestamp"]),
			Open:                parseFloatRecord(record["open"]),
			High:                parseFloatRecord(record["high"]),
			Low:                 parseFloatRecord(record["low"]),
			Close:               parseFloatRecord(record["close"]),
			Volume:              parseFloatRecord(record["volume"]),
			AggressionDirection: record["aggression_direction"],
			AggressionScore:     parseFloatRecord(record["aggression_score"]),
			RejectionDetected:   parseBoolRecord(record["rejection_detected"]),
			RejectionDirection:  record["rejection_direction"],
			RejectionLevel:      parseFloatRecord(record["rejection_level"]),
			VWAPAlignment:       "unknown",
			L2BookPressure:      "unknown",
		})
	}
	return out, nil
}

func applyRejectionVWAPAlignment(points []volumeprofile.RejectionPricePoint, path string) {
	records, err := readCSVRecords(path)
	if err != nil {
		return
	}
	byTime := make(map[int64]string, len(records))
	for _, record := range records {
		timestamp := parseIntRecord(record["timestamp"])
		distance := parseFloatRecord(record["distance_from_vwap"])
		slope := parseFloatRecord(record["vwap_slope"])
		switch {
		case distance >= 0 || slope > 0:
			byTime[timestamp] = "above_vwap"
		case distance < 0 || slope < 0:
			byTime[timestamp] = "below_vwap"
		default:
			byTime[timestamp] = "near_vwap"
		}
	}
	for i := range points {
		if alignment, ok := byTime[points[i].Timestamp]; ok {
			points[i].VWAPAlignment = alignment
		}
	}
}

func applyRejectionContextAlignment(points []volumeprofile.RejectionPricePoint, path string) {
	records, err := readCSVRecords(path)
	if err != nil {
		return
	}
	byTime := make(map[int64]string, len(records))
	for _, record := range records {
		timestamp := parseIntRecord(record["timestamp"])
		priceVsVWAP := record["price_vs_vwap"]
		switch priceVsVWAP {
		case "above_vwap", "below_vwap", "near_vwap":
			byTime[timestamp] = priceVsVWAP
		}
	}
	for i := range points {
		if alignment, ok := byTime[points[i].Timestamp]; ok && alignment != "" {
			points[i].VWAPAlignment = alignment
		}
	}
}

func applyRejectionL2Context(points []volumeprofile.RejectionPricePoint, path string) {
	records, err := readCSVRecords(path)
	if err != nil {
		return
	}
	byTime := make(map[int64]string, len(records))
	for _, record := range records {
		timestamp := parseIntRecord(record["timestamp"])
		pressure := record["book_pressure"]
		if pressure != "" {
			byTime[timestamp] = pressure
		}
	}
	for i := range points {
		if pressure, ok := byTime[points[i].Timestamp]; ok {
			points[i].L2BookPressure = pressure
		}
	}
}
