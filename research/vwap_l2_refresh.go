package research

import (
	"encoding/csv"
	"encoding/json"
	"math"
	"os"
	"sort"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/orderbook"
	"AlgoTrading2026/vwap"
)

type SpreadQuality string

const (
	SpreadQualityTight  SpreadQuality = "tight"
	SpreadQualityNormal SpreadQuality = "normal"
	SpreadQualityWide   SpreadQuality = "wide"
)

type BookPressure string

const (
	BookPressureBid      BookPressure = "bid_pressure"
	BookPressureAsk      BookPressure = "ask_pressure"
	BookPressureBalanced BookPressure = "balanced"
)

type LiquidityQuality string

const (
	LiquidityQualityStrongNearVWAP LiquidityQuality = "strong_near_vwap"
	LiquidityQualityWeakNearVWAP   LiquidityQuality = "weak_near_vwap"
	LiquidityQualityNoneNearVWAP   LiquidityQuality = "no_near_vwap_liquidity"
)

type VWAPL2RefreshInput struct {
	Venue    string
	Symbol   string
	Candles  []exchanges.Candle
	Snapshot orderbook.OrderBookSnapshot
}

type VWAPL2FeatureRow struct {
	Timestamp         int64
	Venue             string
	Symbol            string
	Close             float64
	Volume            float64
	SessionVWAP       float64
	DistanceFromVWAP  float64
	VWAPSlope         float64
	VWAPRegime        vwap.Regime
	SpreadPct         float64
	Imbalance1Pct     float64
	BidDepth1Pct      float64
	AskDepth1Pct      float64
	LiquidityNearVWAP float64
}

type ContextL2FeatureRow struct {
	Timestamp         int64
	Venue             string
	Symbol            string
	Close             float64
	SessionVWAP       float64
	DistanceFromVWAP  float64
	EMA9              float64
	EMA20             float64
	EMAAlignment      vwap.EMAAlignment
	VWAPSlope         float64
	TrendAlignment    TrendAlignment
	SpreadPct         float64
	Imbalance1Pct     float64
	BidDepth1Pct      float64
	AskDepth1Pct      float64
	LiquidityNearVWAP float64
	SpreadQuality     SpreadQuality
	BookPressure      BookPressure
	LiquidityQuality  LiquidityQuality
}

type VWAPL2BehaviorSummary struct {
	Candles                  int      `json:"candles"`
	Touches                  int      `json:"touches"`
	Crosses                  int      `json:"crosses"`
	Bounces                  int      `json:"bounces"`
	Breaks                   int      `json:"breaks"`
	Chops                    int      `json:"chops"`
	BounceBidSupport         int      `json:"bounceBidSupport"`
	BounceAskWeakness        int      `json:"bounceAskWeakness"`
	BreakAskPressure         int      `json:"breakAskPressure"`
	BreakBidCollapse         int      `json:"breakBidCollapse"`
	AverageSpreadPct         float64  `json:"averageSpreadPct"`
	AverageImbalance1Pct     float64  `json:"averageImbalance1Pct"`
	AverageLiquidityNearVWAP float64  `json:"averageLiquidityNearVWAP"`
	SnapshotOnly             bool     `json:"snapshotOnly"`
	Notes                    []string `json:"notes"`
}

type VWAPL2RefreshResult struct {
	FeatureRows []VWAPL2FeatureRow
	ContextRows []ContextL2FeatureRow
	Summary     VWAPL2BehaviorSummary
}

func BuildVWAPL2Refresh(inputs []VWAPL2RefreshInput) VWAPL2RefreshResult {
	featureRows := make([]VWAPL2FeatureRow, 0)
	contextRows := make([]ContextL2FeatureRow, 0)
	summary := VWAPL2BehaviorSummary{
		SnapshotOnly: true,
		Notes: []string{
			"Historical candles are joined with the latest valid L2 snapshot per venue.",
			"L2 fields are current snapshot context, not historical order book replay.",
		},
	}

	for _, input := range inputs {
		if len(input.Candles) == 0 || orderbook.ValidateSnapshot(input.Snapshot) != nil {
			continue
		}
		features, context, venueSummary := buildVenueVWAPL2Refresh(input)
		featureRows = append(featureRows, features...)
		contextRows = append(contextRows, context...)
		mergeVWAPL2Summary(&summary, venueSummary)
	}

	tightThreshold, wideThreshold := spreadPercentileThresholds(contextRows)
	for i := range contextRows {
		contextRows[i].SpreadQuality = ClassifySpreadQuality(contextRows[i].SpreadPct, tightThreshold, wideThreshold)
	}

	finalizeVWAPL2Averages(&summary)

	return VWAPL2RefreshResult{
		FeatureRows: featureRows,
		ContextRows: contextRows,
		Summary:     summary,
	}
}

func buildVenueVWAPL2Refresh(input VWAPL2RefreshInput) ([]VWAPL2FeatureRow, []ContextL2FeatureRow, VWAPL2BehaviorSummary) {
	closeValues := closeColumn(input.Candles)
	session := vwap.SessionVWAP(input.Candles)
	distance := vwap.Distance(closeValues, session)
	slopes := vwap.Slope(session, 5)
	regimes := vwap.Regimes(closeValues, session, slopes)
	ema9 := vwap.EMA9(closeValues)
	ema20 := vwap.EMA20(closeValues)
	relationships := vwap.EMARelationships(closeValues)
	interactions := vwap.Interactions(input.Candles, session)

	spreadPct := orderbook.SpreadPct(input.Snapshot)
	bidDepth := orderbook.BidDepthWithinPct(input.Snapshot, 1)
	askDepth := orderbook.AskDepthWithinPct(input.Snapshot, 1)
	imbalance := orderbook.Imbalance(input.Snapshot, 1)

	featureRows := make([]VWAPL2FeatureRow, 0, len(input.Candles))
	contextRows := make([]ContextL2FeatureRow, 0, len(input.Candles))
	summary := VWAPL2BehaviorSummary{SnapshotOnly: true}

	for i, candle := range input.Candles {
		sessionValue := valueAt(session, i)
		bidLiquidity, askLiquidity := orderbook.LiquidityNearPrice(input.Snapshot, sessionValue, 1)
		liquidityNearVWAP := bidLiquidity + askLiquidity
		closeValue := valueAt(closeValues, i)
		slope := valueAt(slopes, i)
		relationship := emaRelationshipAt(relationships, i)

		featureRows = append(featureRows, VWAPL2FeatureRow{
			Timestamp:         candle.StartTime,
			Venue:             input.Venue,
			Symbol:            input.Symbol,
			Close:             closeValue,
			Volume:            candle.VolumeFloat(),
			SessionVWAP:       sessionValue,
			DistanceFromVWAP:  valueAt(distance, i),
			VWAPSlope:         slope,
			VWAPRegime:        regimeAt(regimes, i),
			SpreadPct:         spreadPct,
			Imbalance1Pct:     imbalance,
			BidDepth1Pct:      bidDepth,
			AskDepth1Pct:      askDepth,
			LiquidityNearVWAP: liquidityNearVWAP,
		})

		contextRows = append(contextRows, ContextL2FeatureRow{
			Timestamp:         candle.StartTime,
			Venue:             input.Venue,
			Symbol:            input.Symbol,
			Close:             closeValue,
			SessionVWAP:       sessionValue,
			DistanceFromVWAP:  valueAt(distance, i),
			EMA9:              valueAt(ema9, i),
			EMA20:             valueAt(ema20, i),
			EMAAlignment:      relationship.Alignment,
			VWAPSlope:         slope,
			TrendAlignment:    ClassifyTrendAlignment(closeValue, sessionValue, valueAt(ema9, i), valueAt(ema20, i), slope),
			SpreadPct:         spreadPct,
			Imbalance1Pct:     imbalance,
			BidDepth1Pct:      bidDepth,
			AskDepth1Pct:      askDepth,
			LiquidityNearVWAP: liquidityNearVWAP,
			BookPressure:      ClassifyBookPressure(imbalance),
			LiquidityQuality:  ClassifyLiquidityQuality(liquidityNearVWAP),
		})

		if i < len(interactions) {
			updateVWAPL2BehaviorSummary(&summary, interactions[i], bidDepth, askDepth, imbalance, spreadPct, liquidityNearVWAP)
		}
	}

	return featureRows, contextRows, summary
}

func ClassifySpreadQuality(spreadPct float64, tightThreshold float64, wideThreshold float64) SpreadQuality {
	switch {
	case spreadPct <= tightThreshold:
		return SpreadQualityTight
	case spreadPct >= wideThreshold:
		return SpreadQualityWide
	default:
		return SpreadQualityNormal
	}
}

func ClassifyBookPressure(imbalance1Pct float64) BookPressure {
	switch {
	case imbalance1Pct > 0.10:
		return BookPressureBid
	case imbalance1Pct < -0.10:
		return BookPressureAsk
	default:
		return BookPressureBalanced
	}
}

func ClassifyLiquidityQuality(liquidityNearVWAP float64) LiquidityQuality {
	switch {
	case liquidityNearVWAP <= 0:
		return LiquidityQualityNoneNearVWAP
	case liquidityNearVWAP < 100000:
		return LiquidityQualityWeakNearVWAP
	default:
		return LiquidityQualityStrongNearVWAP
	}
}

func updateVWAPL2BehaviorSummary(summary *VWAPL2BehaviorSummary, interaction vwap.Interaction, bidDepth float64, askDepth float64, imbalance float64, spreadPct float64, liquidityNearVWAP float64) {
	summary.Candles++
	if interaction.TouchedVWAP {
		summary.Touches++
	}
	if interaction.CrossedVWAP {
		summary.Crosses++
	}

	switch interaction.Reaction {
	case vwap.ReactionBounce:
		summary.Bounces++
		if imbalance > 0 {
			summary.BounceBidSupport++
		}
		if askDepth < bidDepth {
			summary.BounceAskWeakness++
		}
	case vwap.ReactionBreak:
		summary.Breaks++
		if imbalance < 0 {
			summary.BreakAskPressure++
		}
		if bidDepth < askDepth {
			summary.BreakBidCollapse++
		}
	case vwap.ReactionChop:
		summary.Chops++
	}

	summary.AverageSpreadPct += spreadPct
	summary.AverageImbalance1Pct += imbalance
	summary.AverageLiquidityNearVWAP += liquidityNearVWAP
}

func mergeVWAPL2Summary(dst *VWAPL2BehaviorSummary, src VWAPL2BehaviorSummary) {
	dst.Candles += src.Candles
	dst.Touches += src.Touches
	dst.Crosses += src.Crosses
	dst.Bounces += src.Bounces
	dst.Breaks += src.Breaks
	dst.Chops += src.Chops
	dst.BounceBidSupport += src.BounceBidSupport
	dst.BounceAskWeakness += src.BounceAskWeakness
	dst.BreakAskPressure += src.BreakAskPressure
	dst.BreakBidCollapse += src.BreakBidCollapse
	dst.AverageSpreadPct += src.AverageSpreadPct
	dst.AverageImbalance1Pct += src.AverageImbalance1Pct
	dst.AverageLiquidityNearVWAP += src.AverageLiquidityNearVWAP
}

func finalizeVWAPL2Averages(summary *VWAPL2BehaviorSummary) {
	if summary.Candles == 0 {
		return
	}
	divisor := float64(summary.Candles)
	summary.AverageSpreadPct /= divisor
	summary.AverageImbalance1Pct /= divisor
	summary.AverageLiquidityNearVWAP /= divisor
}

func spreadPercentileThresholds(rows []ContextL2FeatureRow) (float64, float64) {
	if len(rows) == 0 {
		return 0, 0
	}
	values := make([]float64, 0, len(rows))
	for _, row := range rows {
		values = append(values, row.SpreadPct)
	}
	sort.Float64s(values)
	return percentile(values, 0.33), percentile(values, 0.66)
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := int(math.Round(p * float64(len(sorted)-1)))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func WriteVWAPL2FeaturesCSV(path string, rows []VWAPL2FeatureRow) error {
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
		"venue",
		"symbol",
		"close",
		"volume",
		"session_vwap",
		"distance_from_vwap",
		"vwap_slope",
		"vwap_regime",
		"spread_pct",
		"imbalance_1pct",
		"bid_depth_1pct",
		"ask_depth_1pct",
		"liquidity_near_vwap",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.Venue,
			row.Symbol,
			floatToString(row.Close),
			floatToString(row.Volume),
			floatToString(row.SessionVWAP),
			floatToString(row.DistanceFromVWAP),
			floatToString(row.VWAPSlope),
			string(row.VWAPRegime),
			floatToString(row.SpreadPct),
			floatToString(row.Imbalance1Pct),
			floatToString(row.BidDepth1Pct),
			floatToString(row.AskDepth1Pct),
			floatToString(row.LiquidityNearVWAP),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteContextL2FeaturesCSV(path string, rows []ContextL2FeatureRow) error {
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
		"venue",
		"symbol",
		"close",
		"session_vwap",
		"distance_from_vwap",
		"ema9",
		"ema20",
		"ema_alignment",
		"vwap_slope",
		"trend_alignment",
		"spread_pct",
		"imbalance_1pct",
		"bid_depth_1pct",
		"ask_depth_1pct",
		"liquidity_near_vwap",
		"spread_quality",
		"book_pressure",
		"liquidity_quality",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.Venue,
			row.Symbol,
			floatToString(row.Close),
			floatToString(row.SessionVWAP),
			floatToString(row.DistanceFromVWAP),
			floatToString(row.EMA9),
			floatToString(row.EMA20),
			string(row.EMAAlignment),
			floatToString(row.VWAPSlope),
			string(row.TrendAlignment),
			floatToString(row.SpreadPct),
			floatToString(row.Imbalance1Pct),
			floatToString(row.BidDepth1Pct),
			floatToString(row.AskDepth1Pct),
			floatToString(row.LiquidityNearVWAP),
			string(row.SpreadQuality),
			string(row.BookPressure),
			string(row.LiquidityQuality),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVWAPL2BehaviorSummaryJSON(path string, summary VWAPL2BehaviorSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0644)
}
