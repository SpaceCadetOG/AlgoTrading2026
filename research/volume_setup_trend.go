package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type VolumeSetupTrendRow struct {
	Timestamp          int64
	SetupID            string
	SetupType          string
	Direction          string
	TrendStart         int64
	TrendEnd           int64
	TrendDirection     string
	TrendStrengthScore float64
	SetupLevel         float64
	POC                float64
	VAH                float64
	VAL                float64
	NearestHVN         float64
	ProfileShape       string
	ShapeConfidence    float64
	ProfileVolume      float64
	Accepted           bool
	Rejected           bool
	POCRetestDetected  bool
	HVNRetestDetected  bool
	VWAPAlignment      string
	L2BookPressure     string
	Invalidated        bool
	FollowThrough5     float64
	FollowThrough10    float64
	FollowThrough20    float64
	Notes              string
}

func BuildVolumeSetupTrendFromFiles(priceActionPath string, strategyPath string, phase3Path string, flexiblePath string, scopedPath string, acceptancePath string, cfg volumeprofile.TrendSetupConfig) ([]VolumeSetupTrendRow, VolumeSetupTrendSummary, error) {
	points, err := readTrendPricePoints(priceActionPath)
	if err != nil {
		return nil, VolumeSetupTrendSummary{}, err
	}
	if _, err := readCSVRecords(strategyPath); err != nil {
		return nil, VolumeSetupTrendSummary{}, err
	}
	if _, err := readCSVRecords(phase3Path); err != nil {
		return nil, VolumeSetupTrendSummary{}, err
	}
	if _, err := readCSVRecords(flexiblePath); err != nil {
		return nil, VolumeSetupTrendSummary{}, err
	}
	if _, err := readCSVRecords(scopedPath); err != nil {
		return nil, VolumeSetupTrendSummary{}, err
	}
	if _, err := readCSVRecords(acceptancePath); err != nil {
		return nil, VolumeSetupTrendSummary{}, err
	}
	applyVWAPAlignment(points, "research/vwap_features.csv")
	pressure := latestL2BookPressure("research/multi_venue_orderbook_features.csv")
	for i := range points {
		points[i].L2BookPressure = pressure
	}
	setups := volumeprofile.DetectTrendSetups(points, cfg)
	rows := BuildVolumeSetupTrendRows(setups)
	summary := BuildVolumeSetupTrendSummary(rows)
	return rows, summary, nil
}

func BuildVolumeSetupTrendRows(setups []volumeprofile.TrendSetup) []VolumeSetupTrendRow {
	rows := make([]VolumeSetupTrendRow, 0, len(setups))
	for _, setup := range setups {
		rows = append(rows, VolumeSetupTrendRow{
			Timestamp:          setup.Timestamp,
			SetupID:            setup.SetupID,
			SetupType:          setup.SetupType,
			Direction:          setup.Direction,
			TrendStart:         setup.TrendStart,
			TrendEnd:           setup.TrendEnd,
			TrendDirection:     setup.TrendDirection,
			TrendStrengthScore: setup.TrendStrengthScore,
			SetupLevel:         setup.SetupLevel,
			POC:                setup.POC,
			VAH:                setup.VAH,
			VAL:                setup.VAL,
			NearestHVN:         setup.NearestHVN,
			ProfileShape:       setup.ProfileShape,
			ShapeConfidence:    setup.ShapeConfidence,
			ProfileVolume:      setup.ProfileVolume,
			Accepted:           setup.Accepted,
			Rejected:           setup.Rejected,
			POCRetestDetected:  setup.POCRetestDetected,
			HVNRetestDetected:  setup.HVNRetestDetected,
			VWAPAlignment:      setup.VWAPAlignment,
			L2BookPressure:     setup.L2BookPressure,
			Invalidated:        setup.Invalidated,
			FollowThrough5:     setup.FollowThrough5,
			FollowThrough10:    setup.FollowThrough10,
			FollowThrough20:    setup.FollowThrough20,
			Notes:              setup.Notes,
		})
	}
	return rows
}

func WriteVolumeSetupTrendCSV(path string, rows []VolumeSetupTrendRow) error {
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
		"timestamp", "setup_id", "setup_type", "direction", "trend_start", "trend_end", "trend_direction", "trend_strength_score",
		"setup_level", "poc", "vah", "val", "nearest_hvn", "profile_shape", "shape_confidence", "profile_volume",
		"accepted", "rejected", "poc_retest_detected", "hvn_retest_detected", "vwap_alignment", "l2_book_pressure",
		"invalidated", "follow_through_5", "follow_through_10", "follow_through_20", "notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10), row.SetupID, row.SetupType, row.Direction,
			strconv.FormatInt(row.TrendStart, 10), strconv.FormatInt(row.TrendEnd, 10), row.TrendDirection, floatToString(row.TrendStrengthScore),
			floatToString(row.SetupLevel), floatToString(row.POC), floatToString(row.VAH), floatToString(row.VAL), floatToString(row.NearestHVN),
			row.ProfileShape, floatToString(row.ShapeConfidence), floatToString(row.ProfileVolume),
			strconv.FormatBool(row.Accepted), strconv.FormatBool(row.Rejected), strconv.FormatBool(row.POCRetestDetected), strconv.FormatBool(row.HVNRetestDetected),
			row.VWAPAlignment, row.L2BookPressure, strconv.FormatBool(row.Invalidated),
			floatToString(row.FollowThrough5), floatToString(row.FollowThrough10), floatToString(row.FollowThrough20), row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVolumeSetupTrendSummaryJSON(path string, summary VolumeSetupTrendSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func readTrendPricePoints(path string) ([]volumeprofile.TrendPricePoint, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	points := make([]volumeprofile.TrendPricePoint, 0, len(records))
	for _, record := range records {
		points = append(points, volumeprofile.TrendPricePoint{
			Timestamp:           parseIntRecord(record["timestamp"]),
			Open:                parseFloatRecord(record["open"]),
			High:                parseFloatRecord(record["high"]),
			Low:                 parseFloatRecord(record["low"]),
			Close:               parseFloatRecord(record["close"]),
			Volume:              parseFloatRecord(record["volume"]),
			AggressionDirection: record["aggression_direction"],
			AggressionScore:     parseFloatRecord(record["aggression_score"]),
			InitiationDetected:  parseBoolRecord(record["initiation_detected"]),
			InitiationDirection: record["initiation_direction"],
			VWAPAlignment:       "unknown",
			L2BookPressure:      "unknown",
		})
	}
	return points, nil
}

func applyVWAPAlignment(points []volumeprofile.TrendPricePoint, path string) {
	records, err := readCSVRecords(path)
	if err != nil {
		return
	}
	vwapByTime := map[int64]float64{}
	for _, record := range records {
		vwapByTime[parseIntRecord(record["timestamp"])] = parseFloatRecord(record["session_vwap"])
	}
	for i := range points {
		vwap := vwapByTime[points[i].Timestamp]
		switch {
		case vwap <= 0:
			points[i].VWAPAlignment = "unknown"
		case points[i].Close > vwap:
			points[i].VWAPAlignment = "above_vwap"
		case points[i].Close < vwap:
			points[i].VWAPAlignment = "below_vwap"
		default:
			points[i].VWAPAlignment = "at_vwap"
		}
	}
}

func latestL2BookPressure(path string) string {
	records, err := readCSVRecords(path)
	if err != nil || len(records) == 0 {
		return "unknown"
	}
	total := 0.0
	count := 0.0
	for _, record := range records {
		if record["valid"] != "true" {
			continue
		}
		total += parseFloatRecord(record["imbalance_1pct"])
		count++
	}
	if count == 0 {
		return "unknown"
	}
	avg := total / count
	switch {
	case avg > 0.05:
		return "bid_pressure"
	case avg < -0.05:
		return "ask_pressure"
	default:
		return "balanced"
	}
}
