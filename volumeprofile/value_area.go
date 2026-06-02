package volumeprofile

func ValueArea(profile VolumeProfile, targetPct float64) (float64, float64) {
	if len(profile.Bins) == 0 {
		return 0, 0
	}
	if targetPct <= 0 || targetPct > 1 {
		targetPct = 0.70
	}
	total := profile.TotalVolume
	if total <= 0 {
		total = TotalVolume(profile.Bins)
	}
	if total <= 0 {
		return 0, 0
	}
	poc := POCIndex(profile.Bins)
	if poc < 0 {
		return 0, 0
	}
	lowIndex := poc
	highIndex := poc
	included := profile.Bins[poc].Volume
	target := total * targetPct
	for included < target && (lowIndex > 0 || highIndex < len(profile.Bins)-1) {
		downVolume := -1.0
		upVolume := -1.0
		if lowIndex > 0 {
			downVolume = profile.Bins[lowIndex-1].Volume
		}
		if highIndex < len(profile.Bins)-1 {
			upVolume = profile.Bins[highIndex+1].Volume
		}
		if upVolume >= downVolume && highIndex < len(profile.Bins)-1 {
			highIndex++
			included += profile.Bins[highIndex].Volume
			continue
		}
		if lowIndex > 0 {
			lowIndex--
			included += profile.Bins[lowIndex].Volume
		}
	}
	return profile.Bins[lowIndex].Low, profile.Bins[highIndex].High
}
