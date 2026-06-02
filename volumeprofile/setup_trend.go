package volumeprofile

import "math"

const (
	TrendSetupType = "TREND_LEG"

	TrendDirectionBullish = "bullish"
	TrendDirectionBearish = "bearish"
)

type TrendSetupConfig struct {
	RetestThresholdPct  float64
	InvalidationPct     float64
	MinTrendStrength    float64
	MaxLegCandles       int
	RetestLookahead     int
	MinDirectionalCount int
}

type TrendPricePoint struct {
	Timestamp           int64
	Open                float64
	High                float64
	Low                 float64
	Close               float64
	Volume              float64
	AggressionDirection string
	AggressionScore     float64
	InitiationDetected  bool
	InitiationDirection string
	VWAPAlignment       string
	L2BookPressure      string
}

type TrendSetup struct {
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

func DefaultTrendSetupConfig() TrendSetupConfig {
	return TrendSetupConfig{
		RetestThresholdPct:  0.0020,
		InvalidationPct:     0.0020,
		MinTrendStrength:    1.10,
		MaxLegCandles:       16,
		RetestLookahead:     96,
		MinDirectionalCount: 2,
	}
}

func DetectTrendSetups(points []TrendPricePoint, cfg TrendSetupConfig) []TrendSetup {
	cfg = normalizeTrendConfig(cfg)
	out := make([]TrendSetup, 0)
	for i := 0; i < len(points); i++ {
		direction := trendDirectionAt(points[i])
		if direction == "" {
			continue
		}
		end := trendLegEnd(points, i, direction, cfg)
		if end <= i {
			continue
		}
		strength, directionalCount := trendStrength(points[i:end+1], direction)
		if strength < cfg.MinTrendStrength || directionalCount < cfg.MinDirectionalCount {
			continue
		}
		profile := trendProfile(points[i : end+1])
		if len(profile.Bins) == 0 || profile.POC <= 0 {
			continue
		}
		shape := AnalyzeProfileShape(profile)
		nearestHVN := nearestNodeTo(profile.POC, profile.HVNs)
		pocRetest := findTrendRetest(points, end+1, profile.POC, cfg.RetestThresholdPct, cfg.RetestLookahead)
		hvnRetest := -1
		if nearestHVN > 0 {
			hvnRetest = findTrendRetest(points, end+1, nearestHVN, cfg.RetestThresholdPct, cfg.RetestLookahead)
		}
		anchor := end
		if pocRetest >= 0 {
			anchor = pocRetest
		} else if hvnRetest >= 0 {
			anchor = hvnRetest
		}
		context := AccumulationDirectionLong
		if direction == TrendDirectionBearish {
			context = AccumulationDirectionShort
		}
		acceptance := classifyTrendAcceptance(points, profile, end+1, 20)
		out = append(out, TrendSetup{
			Timestamp:          points[anchor].Timestamp,
			SetupID:            "trend_leg_" + intString(len(out)+1),
			SetupType:          TrendSetupType,
			Direction:          context,
			TrendStart:         points[i].Timestamp,
			TrendEnd:           points[end].Timestamp,
			TrendDirection:     direction,
			TrendStrengthScore: strength,
			SetupLevel:         profile.POC,
			POC:                profile.POC,
			VAH:                profile.VAH,
			VAL:                profile.VAL,
			NearestHVN:         nearestHVN,
			ProfileShape:       shape.ProfileShape,
			ShapeConfidence:    shape.ShapeConfidence,
			ProfileVolume:      profile.TotalVolume,
			Accepted:           acceptance == AcceptanceAccepted,
			Rejected:           acceptance == AcceptanceRejected,
			POCRetestDetected:  pocRetest >= 0,
			HVNRetestDetected:  hvnRetest >= 0,
			VWAPAlignment:      points[anchor].VWAPAlignment,
			L2BookPressure:     points[anchor].L2BookPressure,
			Invalidated:        trendInvalidated(points[anchor], context, profile.VAL, profile.VAH, cfg.InvalidationPct),
			FollowThrough5:     trendFollowThrough(points, anchor, 5, context),
			FollowThrough10:    trendFollowThrough(points, anchor, 10, context),
			FollowThrough20:    trendFollowThrough(points, anchor, 20, context),
			Notes:              "Research-only trend setup; trend-leg volume profile uses OHLCV candle-volume approximation.",
		})
		i = end
	}
	return out
}

func normalizeTrendConfig(cfg TrendSetupConfig) TrendSetupConfig {
	defaults := DefaultTrendSetupConfig()
	if cfg.RetestThresholdPct <= 0 {
		cfg.RetestThresholdPct = defaults.RetestThresholdPct
	}
	if cfg.InvalidationPct <= 0 {
		cfg.InvalidationPct = defaults.InvalidationPct
	}
	if cfg.MinTrendStrength <= 0 {
		cfg.MinTrendStrength = defaults.MinTrendStrength
	}
	if cfg.MaxLegCandles <= 0 {
		cfg.MaxLegCandles = defaults.MaxLegCandles
	}
	if cfg.RetestLookahead <= 0 {
		cfg.RetestLookahead = defaults.RetestLookahead
	}
	if cfg.MinDirectionalCount <= 0 {
		cfg.MinDirectionalCount = defaults.MinDirectionalCount
	}
	return cfg
}

func trendDirectionAt(point TrendPricePoint) string {
	switch {
	case point.InitiationDetected && point.InitiationDirection == "up":
		return TrendDirectionBullish
	case point.InitiationDetected && point.InitiationDirection == "down":
		return TrendDirectionBearish
	case point.AggressionDirection == "bullish":
		return TrendDirectionBullish
	case point.AggressionDirection == "bearish":
		return TrendDirectionBearish
	default:
		return ""
	}
}

func trendLegEnd(points []TrendPricePoint, start int, direction string, cfg TrendSetupConfig) int {
	end := start
	limit := start + cfg.MaxLegCandles
	if limit >= len(points) {
		limit = len(points) - 1
	}
	oppositeSeen := 0
	for i := start + 1; i <= limit; i++ {
		if isOppositeTrendPoint(points[i], direction) {
			oppositeSeen++
			if oppositeSeen >= 1 {
				break
			}
		}
		if direction == TrendDirectionBullish && points[i].Close < points[i-1].Close && points[i].AggressionDirection == "bearish" {
			break
		}
		if direction == TrendDirectionBearish && points[i].Close > points[i-1].Close && points[i].AggressionDirection == "bullish" {
			break
		}
		end = i
	}
	return end
}

func isOppositeTrendPoint(point TrendPricePoint, direction string) bool {
	return direction == TrendDirectionBullish && point.InitiationDirection == "down" ||
		direction == TrendDirectionBearish && point.InitiationDirection == "up"
}

func trendStrength(points []TrendPricePoint, direction string) (float64, int) {
	if len(points) == 0 {
		return 0, 0
	}
	var total float64
	var directional int
	for _, point := range points {
		if direction == TrendDirectionBullish && (point.AggressionDirection == "bullish" || point.InitiationDirection == "up") {
			directional++
			total += point.AggressionScore
		}
		if direction == TrendDirectionBearish && (point.AggressionDirection == "bearish" || point.InitiationDirection == "down") {
			directional++
			total += point.AggressionScore
		}
	}
	if directional == 0 {
		return 0, 0
	}
	sequenceBonus := float64(directional) / float64(len(points))
	return total/float64(directional) + sequenceBonus, directional
}

func trendProfile(points []TrendPricePoint) VolumeProfile {
	profile := VolumeProfile{}
	if len(points) == 0 {
		return profile
	}
	profile.StartTime = points[0].Timestamp
	profile.EndTime = points[len(points)-1].Timestamp
	profile.Bins = distributeTrendVolume(points, BuildTrendBins(points, DefaultBinConfig()))
	profile.TotalVolume = TotalVolume(profile.Bins)
	profile.POC = PointOfControl(profile)
	profile.VAL, profile.VAH = ValueArea(profile, 0.70)
	profile.HVNs = HighVolumeNodes(profile.Bins)
	profile.LVNs = LowVolumeNodes(profile.Bins)
	profile.Shape = ClassifyShape(profile)
	return profile
}

func BuildTrendBins(points []TrendPricePoint, cfg BinConfig) []PriceBin {
	if len(points) == 0 {
		return nil
	}
	low := points[0].Low
	high := points[0].High
	for _, point := range points {
		if point.Low > 0 && point.Low < low {
			low = point.Low
		}
		if point.High > high {
			high = point.High
		}
	}
	binSize := effectiveBinSize(low, high, cfg)
	count := int((high-low)/binSize) + 1
	bins := make([]PriceBin, 0, count)
	start := floorToBin(low, binSize)
	for i := 0; i < count; i++ {
		binLow := start + float64(i)*binSize
		bins = append(bins, PriceBin{Low: binLow, High: binLow + binSize, Mid: binLow + binSize/2})
	}
	return bins
}

func distributeTrendVolume(points []TrendPricePoint, bins []PriceBin) []PriceBin {
	out := append([]PriceBin(nil), bins...)
	for _, point := range points {
		indices := overlappingBinIndexes(out, point.Low, point.High, point.Close)
		if len(indices) == 0 || point.Volume <= 0 {
			continue
		}
		share := point.Volume / float64(len(indices))
		for _, index := range indices {
			out[index].Volume += share
		}
	}
	return out
}

func floorToBin(price float64, binSize float64) float64 {
	return math.Floor(price/binSize) * binSize
}

func nearestNodeTo(level float64, nodes []PriceBin) float64 {
	if level <= 0 || len(nodes) == 0 {
		return 0
	}
	best := nodes[0].Mid
	bestDistance := absSetup(best - level)
	for _, node := range nodes[1:] {
		distance := absSetup(node.Mid - level)
		if distance < bestDistance {
			best = node.Mid
			bestDistance = distance
		}
	}
	return best
}

func findTrendRetest(points []TrendPricePoint, start int, level float64, thresholdPct float64, lookahead int) int {
	if level <= 0 || start >= len(points) {
		return -1
	}
	end := start + lookahead
	if end > len(points) {
		end = len(points)
	}
	for i := start; i < end; i++ {
		tolerance := level * thresholdPct
		if points[i].Low <= level+tolerance && points[i].High >= level-tolerance {
			return i
		}
	}
	return -1
}

func classifyTrendAcceptance(points []TrendPricePoint, profile VolumeProfile, start int, lookahead int) string {
	if profile.VAH <= profile.VAL || start >= len(points) {
		return AcceptanceNeutral
	}
	end := start + lookahead
	if end > len(points) {
		end = len(points)
	}
	if end <= start {
		return AcceptanceNeutral
	}
	width := profile.VAH - profile.VAL
	inside := 0
	for i := start; i < end; i++ {
		if points[i].Close >= profile.VAL && points[i].Close <= profile.VAH {
			inside++
		}
		if i-start < 5 && (points[i].Close < profile.VAL-width*0.5 || points[i].Close > profile.VAH+width*0.5) {
			return AcceptanceRejected
		}
	}
	if float64(inside)/float64(end-start) >= 0.60 {
		return AcceptanceAccepted
	}
	return AcceptanceNeutral
}

func trendInvalidated(point TrendPricePoint, direction string, val float64, vah float64, invalidationPct float64) bool {
	switch direction {
	case AccumulationDirectionLong:
		return point.Close < val*(1-invalidationPct)
	case AccumulationDirectionShort:
		return point.Close > vah*(1+invalidationPct)
	default:
		return false
	}
}

func trendFollowThrough(points []TrendPricePoint, index int, lookahead int, direction string) float64 {
	if index < 0 || index >= len(points) {
		return 0
	}
	target := index + lookahead
	if target >= len(points) {
		target = len(points) - 1
	}
	move := points[target].Close - points[index].Close
	if direction == AccumulationDirectionShort {
		return -move
	}
	return move
}

func intString(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0)
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
