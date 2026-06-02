package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"time"

	"AlgoTrading2026/exchanges"
	setupFeatures "AlgoTrading2026/features"
	setupLabels "AlgoTrading2026/labels"
)

type SetupFeatureLabelRow struct {
	Feature setupFeatures.SetupFeatureRow
	Label   setupLabels.SetupLabelRow
}

type SetupLabelsSummary struct {
	Rows                int            `json:"rows"`
	BySetupType         map[string]int `json:"bySetupType"`
	AcceptanceLabels    map[string]int `json:"acceptanceLabels"`
	DirectionalLabels   map[string]int `json:"directionalLabels"`
	TripleBarrierLabels map[string]int `json:"tripleBarrierLabels"`
	Notes               []string       `json:"notes"`
}

type SetupFeatureLabelPaths struct {
	ReversalPath     string
	AccumulationPath string
	TrendPath        string
	RejectionPath    string
	VWAPPath         string
	ContextPath      string
	ScopedPath       string
	PriceActionPath  string
}

func BuildSetupFeatureLabelExport(paths SetupFeatureLabelPaths, candles []exchanges.Candle) ([]SetupFeatureLabelRow, SetupLabelsSummary, error) {
	vwapRows, err := readCSVRecords(paths.VWAPPath)
	if err != nil {
		return nil, SetupLabelsSummary{}, err
	}
	contextRows, err := readCSVRecords(paths.ContextPath)
	if err != nil {
		return nil, SetupLabelsSummary{}, err
	}
	if _, err := readCSVRecords(paths.ScopedPath); err != nil {
		return nil, SetupLabelsSummary{}, err
	}
	priceActionRows, err := readCSVRecords(paths.PriceActionPath)
	if err != nil {
		return nil, SetupLabelsSummary{}, err
	}
	setupInputs := []struct {
		setupType string
		path      string
	}{
		{"REVERSAL", paths.ReversalPath},
		{"ACCUMULATION", paths.AccumulationPath},
		{"TREND", paths.TrendPath},
		{"REJECTION", paths.RejectionPath},
	}

	vwapByTime := recordsByTime(vwapRows)
	contextByTime := recordsByTime(contextRows)
	priceActionByTime := recordsByTime(priceActionRows)
	candleCtx := buildSetupCandleContext(candles)

	rows := make([]SetupFeatureLabelRow, 0)
	for _, input := range setupInputs {
		records, err := readCSVRecords(input.path)
		if err != nil {
			return nil, SetupLabelsSummary{}, err
		}
		for _, record := range records {
			row := buildSetupFeatureLabelRow(input.setupType, record, vwapByTime, contextByTime, priceActionByTime, candleCtx)
			rows = append(rows, row)
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Feature.Timestamp == rows[j].Feature.Timestamp {
			return rows[i].Feature.SetupType < rows[j].Feature.SetupType
		}
		return rows[i].Feature.Timestamp < rows[j].Feature.Timestamp
	})
	return rows, BuildSetupLabelsSummary(rows), nil
}

func BuildSetupLabelsSummary(rows []SetupFeatureLabelRow) SetupLabelsSummary {
	summary := SetupLabelsSummary{
		Rows:                len(rows),
		BySetupType:         map[string]int{},
		AcceptanceLabels:    map[string]int{},
		DirectionalLabels:   map[string]int{},
		TripleBarrierLabels: map[string]int{},
		Notes: []string{
			"Normalized labels are research labels only.",
			"No ML model is trained in this phase.",
			"L2 fields remain descriptive until historical replay exists.",
		},
	}
	for _, row := range rows {
		summary.BySetupType[row.Feature.SetupType]++
		summary.AcceptanceLabels[row.Label.LabelAcceptance]++
		summary.DirectionalLabels[row.Label.LabelDirectional20]++
		summary.TripleBarrierLabels[row.Label.LabelTripleBarrier]++
	}
	return summary
}

func WriteSetupFeatureLabelCSV(path string, rows []SetupFeatureLabelRow) error {
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
		"setup_type", "setup_id", "timestamp", "direction",
		"price", "poc", "vah", "val", "nearest_hvn", "nearest_lvn",
		"profile_shape", "shape_confidence", "profile_volume",
		"vwap", "vwap_slope", "distance_to_vwap_pct",
		"distance_to_poc_pct", "distance_to_vah_pct", "distance_to_val_pct", "distance_to_hvn_pct", "distance_to_lvn_pct",
		"time_in_value_area", "atr_14", "atr_pct", "volatility_state", "session", "daily_open_distance_pct",
		"vwap_alignment", "poc_confluence", "hvn_confluence", "vah_val_rejection",
		"accepted", "rejected", "invalidated",
		"follow_through_5", "follow_through_10", "follow_through_20",
		"ft5_bps", "ft10_bps", "ft20_bps",
		"ft5_atr", "ft10_atr", "ft20_atr",
		"label_acceptance", "label_directional_20", "label_triple_barrier",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		f := row.Feature
		l := row.Label
		if err := writer.Write([]string{
			f.SetupType, f.SetupID, strconv.FormatInt(f.Timestamp, 10), f.Direction,
			floatToString(f.Price), floatToString(f.POC), floatToString(f.VAH), floatToString(f.VAL), floatToString(f.NearestHVN), floatToString(f.NearestLVN),
			f.ProfileShape, floatToString(f.ShapeConfidence), floatToString(f.ProfileVolume),
			floatToString(f.VWAP), floatToString(f.VWAPSlope), floatToString(f.DistanceToVWAPPct),
			floatToString(f.DistanceToPOCPct), floatToString(f.DistanceToVAHPct), floatToString(f.DistanceToVALPct), floatToString(f.DistanceToHVNPct), floatToString(f.DistanceToLVNPct),
			strconv.FormatBool(f.TimeInValueArea), floatToString(f.ATR14), floatToString(f.ATRPct), f.VolatilityState, f.Session, floatToString(f.DailyOpenDistancePct),
			f.VWAPAlignment, strconv.FormatBool(f.POCConfluence), strconv.FormatBool(f.HVNConfluence), strconv.FormatBool(f.VAHVALRejection),
			strconv.FormatBool(f.Accepted), strconv.FormatBool(f.Rejected), strconv.FormatBool(f.Invalidated),
			floatToString(l.FollowThrough5), floatToString(l.FollowThrough10), floatToString(l.FollowThrough20),
			floatToString(l.FT5Bps), floatToString(l.FT10Bps), floatToString(l.FT20Bps),
			floatToString(l.FT5ATR), floatToString(l.FT10ATR), floatToString(l.FT20ATR),
			l.LabelAcceptance, l.LabelDirectional20, l.LabelTripleBarrier,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteSetupLabelsSummaryJSON(path string, summary SetupLabelsSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

type setupCandleContext struct {
	high        []float64
	low         []float64
	close       []float64
	atr         []float64
	indexByTime map[int64]int
}

func buildSetupCandleContext(candles []exchanges.Candle) setupCandleContext {
	ctx := setupCandleContext{
		high:        make([]float64, len(candles)),
		low:         make([]float64, len(candles)),
		close:       make([]float64, len(candles)),
		indexByTime: map[int64]int{},
	}
	for i, candle := range candles {
		ctx.high[i] = candle.HighFloat()
		ctx.low[i] = candle.LowFloat()
		ctx.close[i] = candle.CloseFloat()
		ctx.indexByTime[candle.StartTime] = i
	}
	ctx.atr = setupFeatures.ATR(ctx.high, ctx.low, ctx.close, 14)
	return ctx
}

func buildSetupFeatureLabelRow(setupType string, record map[string]string, vwapByTime, contextByTime, priceActionByTime map[int64]map[string]string, candles setupCandleContext) SetupFeatureLabelRow {
	timestamp := parseIntRecord(record["timestamp"])
	vwapRecord := vwapByTime[timestamp]
	contextRecord := contextByTime[timestamp]
	priceActionRecord := priceActionByTime[timestamp]
	candleIndex := candles.indexByTime[timestamp]
	price := setupRecordPrice(record)
	if price == 0 {
		price = parseFloatRecord(vwapRecord["close"])
	}
	atr := setupValueAt(candles.atr, candleIndex)
	atrPct := 0.0
	if price != 0 {
		atrPct = atr / price
	}
	vwapValue := parseFloatRecord(vwapRecord["session_vwap"])
	vwapAlignment := record["vwap_alignment"]
	if vwapAlignment == "" {
		vwapAlignment = vwapAlignmentFromDistance(parseFloatRecord(contextRecord["distance_from_vwap"]))
	}
	feature := setupFeatures.SetupFeatureRow{
		SetupType:            setupType,
		SetupID:              record["setup_id"],
		Timestamp:            timestamp,
		Direction:            record["direction"],
		Price:                price,
		POC:                  parseFloatRecord(record["poc"]),
		VAH:                  parseFloatRecord(record["vah"]),
		VAL:                  parseFloatRecord(record["val"]),
		NearestHVN:           parseFloatRecord(record["nearest_hvn"]),
		NearestLVN:           parseFloatRecord(record["nearest_lvn"]),
		ProfileShape:         record["profile_shape"],
		ShapeConfidence:      parseFloatRecord(record["shape_confidence"]),
		ProfileVolume:        parseFloatRecord(record["profile_volume"]),
		VWAP:                 vwapValue,
		VWAPSlope:            parseFloatRecord(vwapRecord["vwap_slope"]),
		TimeInValueArea:      setupFeatures.InValueArea(price, parseFloatRecord(record["vah"]), parseFloatRecord(record["val"])),
		ATR14:                atr,
		ATRPct:               atrPct,
		VolatilityState:      setupFeatures.VolatilityState(atrPct),
		Session:              sessionFromTimestamp(timestamp),
		DailyOpenDistancePct: setupFeatures.DistancePct(price, parseFloatRecord(priceActionRecord["daily_open_level"])),
		VWAPAlignment:        vwapAlignment,
		POCConfluence:        setupRecordPOCConfluence(record),
		HVNConfluence:        setupRecordHVNConfluence(record),
		VAHVALRejection:      parseBoolRecord(record["vah_val_rejection"]),
		Accepted:             parseBoolRecord(record["accepted"]),
		Rejected:             parseBoolRecord(record["rejected"]),
		Invalidated:          parseBoolRecord(record["invalidated"]),
	}
	feature.DistanceToVWAPPct = setupFeatures.DistancePct(price, feature.VWAP)
	feature.DistanceToPOCPct = setupFeatures.DistancePct(price, feature.POC)
	feature.DistanceToVAHPct = setupFeatures.DistancePct(price, feature.VAH)
	feature.DistanceToVALPct = setupFeatures.DistancePct(price, feature.VAL)
	feature.DistanceToHVNPct = setupFeatures.DistancePct(price, feature.NearestHVN)
	feature.DistanceToLVNPct = setupFeatures.DistancePct(price, feature.NearestLVN)

	ft5 := parseFloatRecord(record["follow_through_5"])
	ft10 := parseFloatRecord(record["follow_through_10"])
	ft20 := parseFloatRecord(record["follow_through_20"])
	label := setupLabels.SetupLabelRow{
		FollowThrough5:     ft5,
		FollowThrough10:    ft10,
		FollowThrough20:    ft20,
		FT5Bps:             setupLabels.Bps(ft5, price),
		FT10Bps:            setupLabels.Bps(ft10, price),
		FT20Bps:            setupLabels.Bps(ft20, price),
		FT5ATR:             setupLabels.ATRMultiple(ft5, atr),
		FT10ATR:            setupLabels.ATRMultiple(ft10, atr),
		FT20ATR:            setupLabels.ATRMultiple(ft20, atr),
		LabelAcceptance:    setupLabels.AcceptanceLabel(feature.Accepted, feature.Rejected),
		LabelDirectional20: setupLabels.DirectionalLabel20(ft20),
		LabelTripleBarrier: setupLabels.TripleBarrierLabel(candles.high, candles.low, candles.close, candleIndex, 20, atr, feature.Direction),
	}
	return SetupFeatureLabelRow{Feature: feature, Label: label}
}

func recordsByTime(records []map[string]string) map[int64]map[string]string {
	out := map[int64]map[string]string{}
	for _, record := range records {
		out[parseIntRecord(record["timestamp"])] = record
	}
	return out
}

func setupRecordPrice(record map[string]string) float64 {
	for _, key := range []string{"setup_level", "reversal_level", "rejection_level", "poc"} {
		value := parseFloatRecord(record[key])
		if value != 0 {
			return value
		}
	}
	return 0
}

func setupValueAt(values []float64, index int) float64 {
	if index < 0 || index >= len(values) {
		return 0
	}
	return values[index]
}

func vwapAlignmentFromDistance(distance float64) string {
	switch {
	case distance > 0:
		return "above_vwap"
	case distance < 0:
		return "below_vwap"
	default:
		return "at_vwap"
	}
}

func sessionFromTimestamp(timestamp int64) string {
	hour := time.UnixMilli(timestamp).UTC().Hour()
	switch {
	case hour < 8:
		return "asia"
	case hour < 13:
		return "europe"
	case hour < 21:
		return "us"
	default:
		return "late_us"
	}
}
