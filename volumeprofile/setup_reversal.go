package volumeprofile

const (
	ReversalSetupType = "REVERSAL_TRADE"

	FailedHighAuction = "failed_high_auction"
	FailedLowAuction  = "failed_low_auction"
)

type ReversalSetupConfig struct {
	MaxEventDistanceMillis float64
	ConfluenceThresholdPct float64
	InvalidationPct        float64
}

type ReversalRejectionInput struct {
	Timestamp                 int64
	SetupID                   string
	Direction                 string
	RejectionLevel            float64
	RejectionHigh             float64
	RejectionLow              float64
	POC                       float64
	VAH                       float64
	VAL                       float64
	NearestHVN                float64
	NearestLVN                float64
	ProfileShape              string
	ShapeConfidence           float64
	ProfileVolume             float64
	VWAPAlignment             string
	DailyPOCConfluence        bool
	Rolling3DPOCConfluence    bool
	Rolling7DPOCConfluence    bool
	Composite30DPOCConfluence bool
	Invalidated               bool
	FollowThrough5            float64
	FollowThrough10           float64
	FollowThrough20           float64
}

type ReversalFailedAuctionInput struct {
	Timestamp       int64
	Strategy        string
	Direction       string
	Level           float64
	Invalidated     bool
	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64
}

type ReversalSetup struct {
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

func DefaultReversalSetupConfig() ReversalSetupConfig {
	return ReversalSetupConfig{
		MaxEventDistanceMillis: 6 * 60 * 60 * 1000,
		ConfluenceThresholdPct: 0.0025,
		InvalidationPct:        0.0020,
	}
}

func DetectReversalSetups(rejections []ReversalRejectionInput, auctions []ReversalFailedAuctionInput, cfg ReversalSetupConfig) []ReversalSetup {
	cfg = normalizeReversalConfig(cfg)
	out := make([]ReversalSetup, 0)
	for _, auction := range auctions {
		direction := reversalDirectionFromAuction(auction.Strategy)
		if direction == "" {
			continue
		}
		rejection, ok := closestRejectionForAuction(rejections, auction, direction, cfg.MaxEventDistanceMillis)
		if !ok {
			continue
		}
		reversalLevel := rejection.POC
		if reversalLevel <= 0 {
			reversalLevel = auction.Level
		}
		invalidated := reversalInvalidated(rejection, auction, direction, cfg.InvalidationPct)
		out = append(out, ReversalSetup{
			Timestamp:         auction.Timestamp,
			SetupID:           "reversal_" + intString(len(out)+1),
			SetupType:         ReversalSetupType,
			Direction:         direction,
			ReversalLevel:     reversalLevel,
			FailedAuctionType: auction.Strategy,
			POC:               rejection.POC,
			VAH:               rejection.VAH,
			VAL:               rejection.VAL,
			NearestHVN:        rejection.NearestHVN,
			NearestLVN:        rejection.NearestLVN,
			ProfileShape:      rejection.ProfileShape,
			ShapeConfidence:   rejection.ShapeConfidence,
			ProfileVolume:     rejection.ProfileVolume,
			Accepted:          reversalAccepted(auction, invalidated),
			Rejected:          reversalRejected(auction, invalidated),
			VWAPAlignment:     rejection.VWAPAlignment,
			POCConfluence:     reversalPOCConfluence(rejection, auction.Level, cfg.ConfluenceThresholdPct),
			HVNConfluence:     rejection.NearestHVN > 0 && NearPOC(rejection.NearestHVN, auction.Level, cfg.ConfluenceThresholdPct),
			VAHVALRejection:   vahValRejection(rejection, auction.Level, cfg.ConfluenceThresholdPct),
			Invalidated:       invalidated,
			FollowThrough5:    auction.FollowThrough5,
			FollowThrough10:   auction.FollowThrough10,
			FollowThrough20:   auction.FollowThrough20,
			Notes:             "Research-only reversal setup; joins failed auction context with rejection profile levels.",
		})
	}
	return out
}

func normalizeReversalConfig(cfg ReversalSetupConfig) ReversalSetupConfig {
	defaults := DefaultReversalSetupConfig()
	if cfg.MaxEventDistanceMillis <= 0 {
		cfg.MaxEventDistanceMillis = defaults.MaxEventDistanceMillis
	}
	if cfg.ConfluenceThresholdPct <= 0 {
		cfg.ConfluenceThresholdPct = defaults.ConfluenceThresholdPct
	}
	if cfg.InvalidationPct <= 0 {
		cfg.InvalidationPct = defaults.InvalidationPct
	}
	return cfg
}

func reversalDirectionFromAuction(strategy string) string {
	switch strategy {
	case FailedHighAuction:
		return AccumulationDirectionShort
	case FailedLowAuction:
		return AccumulationDirectionLong
	default:
		return ""
	}
}

func reversalAccepted(auction ReversalFailedAuctionInput, invalidated bool) bool {
	return !invalidated && auction.FollowThrough20 > 0
}

func reversalRejected(auction ReversalFailedAuctionInput, invalidated bool) bool {
	return invalidated || auction.FollowThrough20 < 0
}

func closestRejectionForAuction(rejections []ReversalRejectionInput, auction ReversalFailedAuctionInput, direction string, maxDistance float64) (ReversalRejectionInput, bool) {
	var best ReversalRejectionInput
	found := false
	bestDistance := maxDistance
	for _, rejection := range rejections {
		if rejection.Direction != direction || rejection.POC <= 0 {
			continue
		}
		distance := absSetup(float64(auction.Timestamp - rejection.Timestamp))
		if distance <= bestDistance {
			best = rejection
			bestDistance = distance
			found = true
		}
	}
	return best, found
}

func reversalPOCConfluence(rejection ReversalRejectionInput, auctionLevel float64, thresholdPct float64) bool {
	return NearPOC(rejection.POC, auctionLevel, thresholdPct) ||
		rejection.DailyPOCConfluence ||
		rejection.Rolling3DPOCConfluence ||
		rejection.Rolling7DPOCConfluence ||
		rejection.Composite30DPOCConfluence
}

func vahValRejection(rejection ReversalRejectionInput, auctionLevel float64, thresholdPct float64) bool {
	return NearPOC(rejection.VAH, auctionLevel, thresholdPct) ||
		NearPOC(rejection.VAL, auctionLevel, thresholdPct) ||
		NearPOC(rejection.RejectionLevel, rejection.VAH, thresholdPct) ||
		NearPOC(rejection.RejectionLevel, rejection.VAL, thresholdPct)
}

func reversalInvalidated(rejection ReversalRejectionInput, auction ReversalFailedAuctionInput, direction string, invalidationPct float64) bool {
	if rejection.Invalidated || auction.Invalidated {
		return true
	}
	switch direction {
	case AccumulationDirectionLong:
		return auction.Level > 0 && rejection.RejectionLow > 0 && rejection.RejectionLow < auction.Level*(1-invalidationPct)
	case AccumulationDirectionShort:
		return auction.Level > 0 && rejection.RejectionHigh > auction.Level*(1+invalidationPct)
	default:
		return false
	}
}
