package volumeprofile

const (
	ShapeReasonBalancedDistribution      = "BALANCED_DISTRIBUTION"
	ShapeReasonUpperDistributionDominant = "UPPER_DISTRIBUTION_DOMINANT"
	ShapeReasonLowerDistributionDominant = "LOWER_DISTRIBUTION_DOMINANT"
	ShapeReasonMultipleThinZones         = "MULTIPLE_THIN_ZONES"
	ShapeReasonLowAcceptance             = "LOW_ACCEPTANCE"
)

type ProfileShapeStudy struct {
	ProfileShape        string
	UpperVolumePct      float64
	LowerVolumePct      float64
	DistributionBalance float64
	ShapeConfidence     float64
	ShapeReason         string
}

func AnalyzeProfileShape(profile VolumeProfile) ProfileShapeStudy {
	total := profile.TotalVolume
	if total <= 0 {
		total = TotalVolume(profile.Bins)
	}
	if len(profile.Bins) == 0 || total <= 0 || profile.POC <= 0 {
		return ProfileShapeStudy{ProfileShape: ShapeUnknown}
	}

	above, below, pocVolume := volumeAroundPOC(profile)
	upperPct := above / total
	lowerPct := below / total
	balance := (above - below) / total
	maxRatio := maxVolumeRatio(profile.Bins, total)
	lvnDensity := 0.0
	if len(profile.Bins) > 0 {
		lvnDensity = float64(len(profile.LVNs)) / float64(len(profile.Bins))
	}
	pocPct := pocVolume / total

	switch {
	case maxRatio < 1.35 || (lvnDensity >= 0.18 && pocPct < 0.05):
		return ProfileShapeStudy{
			ProfileShape:        ShapeThinProfile,
			UpperVolumePct:      upperPct,
			LowerVolumePct:      lowerPct,
			DistributionBalance: balance,
			ShapeConfidence:     clamp01(0.55 + lvnDensity + (1.35-maxRatio)*0.2),
			ShapeReason:         thinReason(lvnDensity, maxRatio),
		}
	case absFloat(balance) <= 0.15 && pocPct >= 0.05:
		return ProfileShapeStudy{
			ProfileShape:        ShapeDProfile,
			UpperVolumePct:      upperPct,
			LowerVolumePct:      lowerPct,
			DistributionBalance: balance,
			ShapeConfidence:     clamp01(1 - absFloat(balance)*3),
			ShapeReason:         ShapeReasonBalancedDistribution,
		}
	case balance > 0.15:
		return ProfileShapeStudy{
			ProfileShape:        ShapePProfile,
			UpperVolumePct:      upperPct,
			LowerVolumePct:      lowerPct,
			DistributionBalance: balance,
			ShapeConfidence:     clamp01(0.55 + absFloat(balance)),
			ShapeReason:         ShapeReasonUpperDistributionDominant,
		}
	case balance < -0.15:
		return ProfileShapeStudy{
			ProfileShape:        ShapeBProfile,
			UpperVolumePct:      upperPct,
			LowerVolumePct:      lowerPct,
			DistributionBalance: balance,
			ShapeConfidence:     clamp01(0.55 + absFloat(balance)),
			ShapeReason:         ShapeReasonLowerDistributionDominant,
		}
	default:
		return ProfileShapeStudy{
			ProfileShape:        ShapeUnknown,
			UpperVolumePct:      upperPct,
			LowerVolumePct:      lowerPct,
			DistributionBalance: balance,
			ShapeConfidence:     0,
			ShapeReason:         ShapeReasonLowAcceptance,
		}
	}
}

func volumeAroundPOC(profile VolumeProfile) (float64, float64, float64) {
	var above, below, pocVolume float64
	for _, bin := range profile.Bins {
		switch {
		case bin.Mid > profile.POC:
			above += bin.Volume
		case bin.Mid < profile.POC:
			below += bin.Volume
		default:
			pocVolume += bin.Volume
		}
	}
	return above, below, pocVolume
}

func maxVolumeRatio(bins []PriceBin, total float64) float64 {
	if len(bins) == 0 || total <= 0 {
		return 0
	}
	avg := total / float64(len(bins))
	if avg <= 0 {
		return 0
	}
	maxVolume := 0.0
	for _, bin := range bins {
		if bin.Volume > maxVolume {
			maxVolume = bin.Volume
		}
	}
	return maxVolume / avg
}

func thinReason(lvnDensity float64, maxRatio float64) string {
	if lvnDensity >= 0.18 {
		return ShapeReasonMultipleThinZones
	}
	if maxRatio < 1.35 {
		return ShapeReasonLowAcceptance
	}
	return ShapeReasonMultipleThinZones
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
