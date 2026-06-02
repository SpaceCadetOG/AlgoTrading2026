package volumeprofile

const (
	ShapeDProfile    = "D_PROFILE"
	ShapePProfile    = "P_PROFILE"
	ShapeBProfile    = "B_PROFILE"
	ShapeThinProfile = "THIN_PROFILE"
	ShapeUnknown     = "UNKNOWN"
)

func ClassifyShape(profile VolumeProfile) string {
	bins := profile.Bins
	if len(bins) < 3 {
		return ShapeUnknown
	}
	total := profile.TotalVolume
	if total <= 0 {
		total = TotalVolume(bins)
	}
	if total <= 0 {
		return ShapeUnknown
	}
	maxVolume := 0.0
	for _, bin := range bins {
		if bin.Volume > maxVolume {
			maxVolume = bin.Volume
		}
	}
	avgVolume := total / float64(len(bins))
	if avgVolume > 0 && maxVolume/avgVolume < 1.35 {
		return ShapeThinProfile
	}

	lower, middle, upper := thirdVolumes(bins)
	balance := absFloat(lower-upper) / total
	if middle >= lower && middle >= upper && balance <= 0.25 {
		return ShapeDProfile
	}
	if upper > lower*1.35 && upper >= middle*0.8 {
		return ShapePProfile
	}
	if lower > upper*1.35 && lower >= middle*0.8 {
		return ShapeBProfile
	}
	return ShapeUnknown
}

func thirdVolumes(bins []PriceBin) (float64, float64, float64) {
	if len(bins) == 0 {
		return 0, 0, 0
	}
	lowerEnd := len(bins) / 3
	upperStart := len(bins) * 2 / 3
	var lower, middle, upper float64
	for i, bin := range bins {
		switch {
		case i < lowerEnd:
			lower += bin.Volume
		case i >= upperStart:
			upper += bin.Volume
		default:
			middle += bin.Volume
		}
	}
	return lower, middle, upper
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
