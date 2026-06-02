package volumeprofile

const (
	RejectionSetupType = "REJECTION"

	RejectionDirectionBullish = "bullish"
	RejectionDirectionBearish = "bearish"
)

type RejectionSetupConfig struct {
	ConfluenceThresholdPct float64
	RetestThresholdPct     float64
	InvalidationPct        float64
	LookbackCandles        int
	ConfirmationLookahead  int
	RetestLookahead        int
}

type RejectionPricePoint struct {
	Timestamp           int64
	Open                float64
	High                float64
	Low                 float64
	Close               float64
	Volume              float64
	AggressionDirection string
	AggressionScore     float64
	RejectionDetected   bool
	RejectionDirection  string
	RejectionLevel      float64
	VWAPAlignment       string
	L2BookPressure      string
}

type RejectionSetup struct {
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

func DefaultRejectionSetupConfig() RejectionSetupConfig {
	return RejectionSetupConfig{
		ConfluenceThresholdPct: 0.0025,
		RetestThresholdPct:     0.0020,
		InvalidationPct:        0.0020,
		LookbackCandles:        3,
		ConfirmationLookahead:  5,
		RetestLookahead:        96,
	}
}

func DetectRejectionSetups(points []RejectionPricePoint, scoped ScopedPOCLevels, cfg RejectionSetupConfig) []RejectionSetup {
	cfg = normalizeRejectionConfig(cfg)
	out := make([]RejectionSetup, 0)
	for i, point := range points {
		if !point.RejectionDetected {
			continue
		}
		context := rejectionContext(point.RejectionDirection)
		if context == "" {
			continue
		}
		end := findRejectionConfirmation(points, i, context, cfg.ConfirmationLookahead)
		if end < i {
			continue
		}
		start := findRejectionStart(points, i, context, cfg.LookbackCandles)
		profile := rejectionProfile(points[start : end+1])
		if len(profile.Bins) == 0 || profile.POC <= 0 {
			continue
		}
		shape := AnalyzeProfileShape(profile)
		nearestHVN := nearestNodeTo(profile.POC, profile.HVNs)
		nearestLVN := nearestNodeTo(profile.POC, profile.LVNs)
		pocRetest := findRejectionRetest(points, end+1, profile.POC, cfg.RetestThresholdPct, cfg.RetestLookahead)
		hvnRetest := -1
		if nearestHVN > 0 {
			hvnRetest = findRejectionRetest(points, end+1, nearestHVN, cfg.RetestThresholdPct, cfg.RetestLookahead)
		}
		anchor := end
		if pocRetest >= 0 {
			anchor = pocRetest
		} else if hvnRetest >= 0 {
			anchor = hvnRetest
		}
		rejectionHigh, rejectionLow := rejectionHighLow(points[start : end+1])
		acceptance := classifyRejectionAcceptance(points, profile, end+1, 20)
		out = append(out, RejectionSetup{
			Timestamp:                 points[anchor].Timestamp,
			SetupID:                   "rejection_" + intString(len(out)+1),
			SetupType:                 RejectionSetupType,
			Direction:                 context,
			RejectionStart:            points[start].Timestamp,
			RejectionEnd:              points[end].Timestamp,
			RejectionLevel:            rejectionLevel(point),
			RejectionHigh:             rejectionHigh,
			RejectionLow:              rejectionLow,
			SetupLevel:                profile.POC,
			POC:                       profile.POC,
			VAH:                       profile.VAH,
			VAL:                       profile.VAL,
			NearestHVN:                nearestHVN,
			NearestLVN:                nearestLVN,
			ProfileShape:              shape.ProfileShape,
			ShapeConfidence:           shape.ShapeConfidence,
			ProfileVolume:             profile.TotalVolume,
			Accepted:                  acceptance == AcceptanceAccepted,
			Rejected:                  acceptance == AcceptanceRejected,
			POCRetestDetected:         pocRetest >= 0,
			HVNRetestDetected:         hvnRetest >= 0,
			VWAPAlignment:             points[anchor].VWAPAlignment,
			L2BookPressure:            points[anchor].L2BookPressure,
			DailyPOCConfluence:        NearPOC(profile.POC, scoped.Daily, cfg.ConfluenceThresholdPct),
			Rolling3DPOCConfluence:    NearPOC(profile.POC, scoped.Rolling3D, cfg.ConfluenceThresholdPct),
			Rolling7DPOCConfluence:    NearPOC(profile.POC, scoped.Rolling7D, cfg.ConfluenceThresholdPct),
			Composite30DPOCConfluence: NearPOC(profile.POC, scoped.Composite30D, cfg.ConfluenceThresholdPct),
			Invalidated:               rejectionInvalidated(points[anchor], context, rejectionLow, rejectionHigh, profile.VAL, profile.VAH, cfg.InvalidationPct),
			FollowThrough5:            rejectionFollowThrough(points, anchor, 5, context),
			FollowThrough10:           rejectionFollowThrough(points, anchor, 10, context),
			FollowThrough20:           rejectionFollowThrough(points, anchor, 20, context),
			Notes:                     "Research-only rejection setup; rejection profile uses OHLCV candle-volume approximation.",
		})
	}
	return out
}

func normalizeRejectionConfig(cfg RejectionSetupConfig) RejectionSetupConfig {
	defaults := DefaultRejectionSetupConfig()
	if cfg.ConfluenceThresholdPct <= 0 {
		cfg.ConfluenceThresholdPct = defaults.ConfluenceThresholdPct
	}
	if cfg.RetestThresholdPct <= 0 {
		cfg.RetestThresholdPct = defaults.RetestThresholdPct
	}
	if cfg.InvalidationPct <= 0 {
		cfg.InvalidationPct = defaults.InvalidationPct
	}
	if cfg.LookbackCandles <= 0 {
		cfg.LookbackCandles = defaults.LookbackCandles
	}
	if cfg.ConfirmationLookahead <= 0 {
		cfg.ConfirmationLookahead = defaults.ConfirmationLookahead
	}
	if cfg.RetestLookahead <= 0 {
		cfg.RetestLookahead = defaults.RetestLookahead
	}
	return cfg
}

func rejectionContext(direction string) string {
	switch direction {
	case RejectionDirectionBullish:
		return AccumulationDirectionLong
	case RejectionDirectionBearish:
		return AccumulationDirectionShort
	default:
		return ""
	}
}

func findRejectionStart(points []RejectionPricePoint, index int, context string, lookback int) int {
	start := index - lookback
	if start < 0 {
		start = 0
	}
	for i := index - 1; i >= start; i-- {
		if context == AccumulationDirectionLong && points[i].AggressionDirection == TrendDirectionBearish {
			return i
		}
		if context == AccumulationDirectionShort && points[i].AggressionDirection == TrendDirectionBullish {
			return i
		}
	}
	return start
}

func findRejectionConfirmation(points []RejectionPricePoint, index int, context string, lookahead int) int {
	end := index + lookahead
	if end >= len(points) {
		end = len(points) - 1
	}
	for i := index; i <= end; i++ {
		if context == AccumulationDirectionLong && points[i].AggressionDirection == TrendDirectionBullish {
			return i
		}
		if context == AccumulationDirectionShort && points[i].AggressionDirection == TrendDirectionBearish {
			return i
		}
	}
	return index
}

func rejectionProfile(points []RejectionPricePoint) VolumeProfile {
	trendPoints := make([]TrendPricePoint, 0, len(points))
	for _, point := range points {
		trendPoints = append(trendPoints, TrendPricePoint{
			Timestamp: point.Timestamp,
			Open:      point.Open,
			High:      point.High,
			Low:       point.Low,
			Close:     point.Close,
			Volume:    point.Volume,
		})
	}
	return trendProfile(trendPoints)
}

func rejectionHighLow(points []RejectionPricePoint) (float64, float64) {
	if len(points) == 0 {
		return 0, 0
	}
	high := points[0].High
	low := points[0].Low
	for _, point := range points[1:] {
		if point.High > high {
			high = point.High
		}
		if point.Low > 0 && point.Low < low {
			low = point.Low
		}
	}
	return high, low
}

func rejectionLevel(point RejectionPricePoint) float64 {
	if point.RejectionLevel > 0 {
		return point.RejectionLevel
	}
	if point.RejectionDirection == RejectionDirectionBullish {
		return point.Low
	}
	if point.RejectionDirection == RejectionDirectionBearish {
		return point.High
	}
	return point.Close
}

func findRejectionRetest(points []RejectionPricePoint, start int, level float64, thresholdPct float64, lookahead int) int {
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

func classifyRejectionAcceptance(points []RejectionPricePoint, profile VolumeProfile, start int, lookahead int) string {
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

func rejectionInvalidated(point RejectionPricePoint, direction string, rejectionLow float64, rejectionHigh float64, val float64, vah float64, invalidationPct float64) bool {
	switch direction {
	case AccumulationDirectionLong:
		level := rejectionLow
		if val > 0 && val < level {
			level = val
		}
		return level > 0 && point.Close < level*(1-invalidationPct)
	case AccumulationDirectionShort:
		level := rejectionHigh
		if vah > level {
			level = vah
		}
		return level > 0 && point.Close > level*(1+invalidationPct)
	default:
		return false
	}
}

func rejectionFollowThrough(points []RejectionPricePoint, index int, lookahead int, direction string) float64 {
	if index < 0 || index >= len(points) || lookahead <= 0 {
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
