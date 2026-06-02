package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"AlgoTrading2026/volumeprofile"
)

type VolumeSetupReversalRow struct {
	Timestamp         int64
	SetupID           string
	SetupType         string
	Direction         string
	ReversalLevel     float64
	FailedAuctionType string
	POC               float64
	VAH               float64
	VAL               float64
	NearestHVN        float64
	NearestLVN        float64
	ProfileShape      string
	ShapeConfidence   float64
	ProfileVolume     float64
	Accepted          bool
	Rejected          bool
	VWAPAlignment     string
	POCConfluence     bool
	HVNConfluence     bool
	VAHVALRejection   bool
	Invalidated       bool
	FollowThrough5    float64
	FollowThrough10   float64
	FollowThrough20   float64
	Notes             string
}

func BuildVolumeSetupReversalFromFiles(rejectionPath string, rejectionQualityPath string, phase3Path string, flexiblePath string, scopedPath string, vwapPath string, contextPath string, cfg volumeprofile.ReversalSetupConfig) ([]VolumeSetupReversalRow, VolumeSetupReversalSummary, error) {
	rejections, err := readReversalRejections(rejectionPath)
	if err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	if _, err := readCSVRecords(rejectionQualityPath); err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	auctions, err := readReversalFailedAuctions(phase3Path)
	if err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	if _, err := readCSVRecords(flexiblePath); err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	if _, err := readCSVRecords(scopedPath); err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	if _, err := readCSVRecords(vwapPath); err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	if _, err := readCSVRecords(contextPath); err != nil {
		return nil, VolumeSetupReversalSummary{}, err
	}
	setups := volumeprofile.DetectReversalSetups(rejections, auctions, cfg)
	rows := BuildVolumeSetupReversalRows(setups)
	summary := BuildVolumeSetupReversalSummary(rows)
	return rows, summary, nil
}

func BuildVolumeSetupReversalRows(setups []volumeprofile.ReversalSetup) []VolumeSetupReversalRow {
	rows := make([]VolumeSetupReversalRow, 0, len(setups))
	for _, setup := range setups {
		rows = append(rows, VolumeSetupReversalRow{
			Timestamp:         setup.Timestamp,
			SetupID:           setup.SetupID,
			SetupType:         setup.SetupType,
			Direction:         setup.Direction,
			ReversalLevel:     setup.ReversalLevel,
			FailedAuctionType: setup.FailedAuctionType,
			POC:               setup.POC,
			VAH:               setup.VAH,
			VAL:               setup.VAL,
			NearestHVN:        setup.NearestHVN,
			NearestLVN:        setup.NearestLVN,
			ProfileShape:      setup.ProfileShape,
			ShapeConfidence:   setup.ShapeConfidence,
			ProfileVolume:     setup.ProfileVolume,
			Accepted:          setup.Accepted,
			Rejected:          setup.Rejected,
			VWAPAlignment:     setup.VWAPAlignment,
			POCConfluence:     setup.POCConfluence,
			HVNConfluence:     setup.HVNConfluence,
			VAHVALRejection:   setup.VAHVALRejection,
			Invalidated:       setup.Invalidated,
			FollowThrough5:    setup.FollowThrough5,
			FollowThrough10:   setup.FollowThrough10,
			FollowThrough20:   setup.FollowThrough20,
			Notes:             setup.Notes,
		})
	}
	return rows
}

func WriteVolumeSetupReversalCSV(path string, rows []VolumeSetupReversalRow) error {
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
		"timestamp", "setup_id", "setup_type", "direction", "reversal_level", "failed_auction_type",
		"poc", "vah", "val", "nearest_hvn", "nearest_lvn", "profile_shape", "shape_confidence", "profile_volume",
		"accepted", "rejected", "vwap_alignment", "poc_confluence", "hvn_confluence", "vah_val_rejection", "invalidated",
		"follow_through_5", "follow_through_10", "follow_through_20", "notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10), row.SetupID, row.SetupType, row.Direction,
			floatToString(row.ReversalLevel), row.FailedAuctionType,
			floatToString(row.POC), floatToString(row.VAH), floatToString(row.VAL), floatToString(row.NearestHVN), floatToString(row.NearestLVN),
			row.ProfileShape, floatToString(row.ShapeConfidence), floatToString(row.ProfileVolume),
			strconv.FormatBool(row.Accepted), strconv.FormatBool(row.Rejected), row.VWAPAlignment,
			strconv.FormatBool(row.POCConfluence), strconv.FormatBool(row.HVNConfluence), strconv.FormatBool(row.VAHVALRejection), strconv.FormatBool(row.Invalidated),
			floatToString(row.FollowThrough5), floatToString(row.FollowThrough10), floatToString(row.FollowThrough20), row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVolumeSetupReversalSummaryJSON(path string, summary VolumeSetupReversalSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func readReversalRejections(path string) ([]volumeprofile.ReversalRejectionInput, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]volumeprofile.ReversalRejectionInput, 0, len(records))
	for _, record := range records {
		out = append(out, volumeprofile.ReversalRejectionInput{
			Timestamp:                 parseIntRecord(record["timestamp"]),
			SetupID:                   record["setup_id"],
			Direction:                 record["direction"],
			RejectionLevel:            parseFloatRecord(record["rejection_level"]),
			RejectionHigh:             parseFloatRecord(record["rejection_high"]),
			RejectionLow:              parseFloatRecord(record["rejection_low"]),
			POC:                       parseFloatRecord(record["poc"]),
			VAH:                       parseFloatRecord(record["vah"]),
			VAL:                       parseFloatRecord(record["val"]),
			NearestHVN:                parseFloatRecord(record["nearest_hvn"]),
			NearestLVN:                parseFloatRecord(record["nearest_lvn"]),
			ProfileShape:              record["profile_shape"],
			ShapeConfidence:           parseFloatRecord(record["shape_confidence"]),
			ProfileVolume:             parseFloatRecord(record["profile_volume"]),
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

func readReversalFailedAuctions(path string) ([]volumeprofile.ReversalFailedAuctionInput, error) {
	records, err := readCSVRecords(path)
	if err != nil {
		return nil, err
	}
	out := make([]volumeprofile.ReversalFailedAuctionInput, 0)
	for _, record := range records {
		strategy := record["strategy"]
		if strategy != volumeprofile.FailedHighAuction && strategy != volumeprofile.FailedLowAuction {
			continue
		}
		out = append(out, volumeprofile.ReversalFailedAuctionInput{
			Timestamp:       parseIntRecord(record["timestamp"]),
			Strategy:        strategy,
			Direction:       record["direction"],
			Level:           parseFloatRecord(record["level"]),
			Invalidated:     parseBoolRecord(record["invalidated"]),
			FollowThrough5:  parseFloatRecord(record["follow_through_5"]),
			FollowThrough10: parseFloatRecord(record["follow_through_10"]),
			FollowThrough20: parseFloatRecord(record["follow_through_20"]),
		})
	}
	return out, nil
}
