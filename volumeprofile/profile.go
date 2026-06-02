package volumeprofile

import "AlgoTrading2026/exchanges"

type VolumeProfile struct {
	StartTime int64
	EndTime   int64
	Symbol    string
	Venue     string

	Bins []PriceBin

	POC float64
	VAH float64
	VAL float64

	TotalVolume float64

	HVNs []PriceBin
	LVNs []PriceBin

	Shape string
}

func BuildProfile(candles []exchanges.Candle, cfg BinConfig) VolumeProfile {
	profile := VolumeProfile{}
	if len(candles) == 0 {
		return profile
	}
	profile.StartTime = candles[0].StartTime
	profile.EndTime = candles[len(candles)-1].EndTime
	profile.Symbol = candles[0].Symbol
	profile.Venue = candles[0].Venue
	profile.Bins = DistributeCandleVolume(candles, BuildPriceBins(candles, cfg))
	profile.TotalVolume = TotalVolume(profile.Bins)
	profile.POC = PointOfControl(profile)
	profile.VAL, profile.VAH = ValueArea(profile, 0.70)
	profile.HVNs = HighVolumeNodes(profile.Bins)
	profile.LVNs = LowVolumeNodes(profile.Bins)
	profile.Shape = ClassifyShape(profile)
	return profile
}

func TotalVolume(bins []PriceBin) float64 {
	total := 0.0
	for _, bin := range bins {
		total += bin.Volume
	}
	return total
}
