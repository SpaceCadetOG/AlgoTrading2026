package volumeprofile

const (
	AccumulationDirectionLong  = "long_context"
	AccumulationDirectionShort = "short_context"
)

type AccumulationSetupConfig struct {
	ConfluenceThresholdPct float64
	RetestThresholdPct     float64
	InvalidationPct        float64
	InitiationLookahead    int
	RetestLookahead        int
}

type AccumulationProfileInput struct {
	ProfileID       string
	StartTime       int64
	EndTime         int64
	Symbol          string
	POC             float64
	VAH             float64
	VAL             float64
	ProfileShape    string
	ShapeConfidence float64
	ProfileVolume   float64
	AcceptanceState string
}

type AccumulationPricePoint struct {
	Timestamp           int64
	High                float64
	Low                 float64
	Close               float64
	InitiationDetected  bool
	InitiationDirection string
}

type ScopedPOCLevels struct {
	Daily        float64
	Rolling3D    float64
	Rolling7D    float64
	Composite30D float64
}

type AccumulationSetup struct {
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

func DefaultAccumulationSetupConfig() AccumulationSetupConfig {
	return AccumulationSetupConfig{
		ConfluenceThresholdPct: 0.0025,
		RetestThresholdPct:     0.0020,
		InvalidationPct:        0.0020,
		InitiationLookahead:    20,
		RetestLookahead:        96,
	}
}

func DetectAccumulationSetups(profiles []AccumulationProfileInput, points []AccumulationPricePoint, scoped ScopedPOCLevels, cfg AccumulationSetupConfig) []AccumulationSetup {
	cfg = normalizeAccumulationConfig(cfg)
	out := make([]AccumulationSetup, 0)
	for _, profile := range profiles {
		if !validAccumulationProfile(profile) {
			continue
		}
		initiationIndex := findInitiationAfter(points, profile.EndTime, cfg.InitiationLookahead)
		if initiationIndex < 0 {
			continue
		}
		direction := directionFromInitiation(points[initiationIndex].InitiationDirection)
		if direction == "" {
			continue
		}
		retestIndex := findRetest(points, initiationIndex+1, profile.POC, cfg.RetestThresholdPct, cfg.RetestLookahead)
		anchorIndex := initiationIndex
		if retestIndex >= 0 {
			anchorIndex = retestIndex
		}
		setup := AccumulationSetup{
			Timestamp:                 points[anchorIndex].Timestamp,
			SetupID:                   profile.ProfileID,
			SetupType:                 FlexibleSidewaysAccumulation,
			Direction:                 direction,
			AccumulationStart:         profile.StartTime,
			AccumulationEnd:           profile.EndTime,
			SetupLevel:                profile.POC,
			POC:                       profile.POC,
			VAH:                       profile.VAH,
			VAL:                       profile.VAL,
			ProfileShape:              profile.ProfileShape,
			ShapeConfidence:           profile.ShapeConfidence,
			ProfileVolume:             profile.ProfileVolume,
			Accepted:                  profile.AcceptanceState == AcceptanceAccepted,
			Rejected:                  profile.AcceptanceState == AcceptanceRejected,
			DailyPOCConfluence:        NearPOC(profile.POC, scoped.Daily, cfg.ConfluenceThresholdPct),
			Rolling3DPOCConfluence:    NearPOC(profile.POC, scoped.Rolling3D, cfg.ConfluenceThresholdPct),
			Rolling7DPOCConfluence:    NearPOC(profile.POC, scoped.Rolling7D, cfg.ConfluenceThresholdPct),
			Composite30DPOCConfluence: NearPOC(profile.POC, scoped.Composite30D, cfg.ConfluenceThresholdPct),
			RetestDetected:            retestIndex >= 0,
			Invalidated:               invalidatedAfter(points, anchorIndex, direction, profile.VAL, profile.VAH, cfg.InvalidationPct),
			FollowThrough5:            accumulationFollowThrough(points, anchorIndex, 5, direction),
			FollowThrough10:           accumulationFollowThrough(points, anchorIndex, 10, direction),
			FollowThrough20:           accumulationFollowThrough(points, anchorIndex, 20, direction),
			Notes:                     "Research-only accumulation setup; setup level uses accumulation POC. HVN level selection requires future flexible profile bin-level enrichment.",
		}
		out = append(out, setup)
	}
	return out
}

func NearPOC(a float64, b float64, thresholdPct float64) bool {
	if a <= 0 || b <= 0 {
		return false
	}
	return absSetup(a-b)/b <= thresholdPct
}

func normalizeAccumulationConfig(cfg AccumulationSetupConfig) AccumulationSetupConfig {
	defaults := DefaultAccumulationSetupConfig()
	if cfg.ConfluenceThresholdPct <= 0 {
		cfg.ConfluenceThresholdPct = defaults.ConfluenceThresholdPct
	}
	if cfg.RetestThresholdPct <= 0 {
		cfg.RetestThresholdPct = defaults.RetestThresholdPct
	}
	if cfg.InvalidationPct <= 0 {
		cfg.InvalidationPct = defaults.InvalidationPct
	}
	if cfg.InitiationLookahead <= 0 {
		cfg.InitiationLookahead = defaults.InitiationLookahead
	}
	if cfg.RetestLookahead <= 0 {
		cfg.RetestLookahead = defaults.RetestLookahead
	}
	return cfg
}

func validAccumulationProfile(profile AccumulationProfileInput) bool {
	return profile.POC > 0 && profile.ProfileVolume > 0 && profile.ProfileShape != ShapeUnknown
}

func findInitiationAfter(points []AccumulationPricePoint, endTime int64, lookahead int) int {
	seenAfter := 0
	for i, point := range points {
		if point.Timestamp <= endTime {
			continue
		}
		if seenAfter >= lookahead {
			return -1
		}
		if point.InitiationDetected {
			return i
		}
		seenAfter++
	}
	return -1
}

func directionFromInitiation(direction string) string {
	switch direction {
	case "up":
		return AccumulationDirectionLong
	case "down":
		return AccumulationDirectionShort
	default:
		return ""
	}
}

func findRetest(points []AccumulationPricePoint, start int, level float64, thresholdPct float64, lookahead int) int {
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

func invalidatedAfter(points []AccumulationPricePoint, start int, direction string, val float64, vah float64, invalidationPct float64) bool {
	if start < 0 || start >= len(points) {
		return false
	}
	switch direction {
	case AccumulationDirectionLong:
		threshold := val * (1 - invalidationPct)
		return points[start].Close < threshold
	case AccumulationDirectionShort:
		threshold := vah * (1 + invalidationPct)
		return points[start].Close > threshold
	default:
		return false
	}
}

func accumulationFollowThrough(points []AccumulationPricePoint, index int, lookahead int, direction string) float64 {
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

func absSetup(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
