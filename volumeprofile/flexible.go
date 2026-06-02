package volumeprofile

import "AlgoTrading2026/exchanges"

const (
	FlexibleSidewaysAccumulation = "SIDEWAYS_ACCUMULATION"
	FlexibleOpenDrive            = "OPEN_DRIVE"
	FlexibleFailedAuction        = "FAILED_AUCTION"
	FlexibleHighLowRetest        = "HIGH_LOW_RETEST"
)

const (
	AcceptanceAccepted = "accepted"
	AcceptanceRejected = "rejected"
	AcceptanceNeutral  = "neutral"
)

type FlexibleProfileWindow struct {
	ProfileType string
	ProfileID   string
	StartTime   int64
	EndTime     int64
}

type FlexibleProfile struct {
	ProfileType     string
	ProfileID       string
	StartTime       int64
	EndTime         int64
	Profile         VolumeProfile
	ShapeStudy      ProfileShapeStudy
	AcceptanceState string
}

func BuildFlexibleProfile(candles []exchanges.Candle, window FlexibleProfileWindow, cfg BinConfig) FlexibleProfile {
	sliced := CandlesBetween(candles, window.StartTime, window.EndTime)
	profile := BuildProfile(sliced, cfg)
	shape := AnalyzeProfileShape(profile)
	return FlexibleProfile{
		ProfileType:     window.ProfileType,
		ProfileID:       window.ProfileID,
		StartTime:       window.StartTime,
		EndTime:         window.EndTime,
		Profile:         profile,
		ShapeStudy:      shape,
		AcceptanceState: ClassifyAcceptance(candles, profile, window.EndTime, 20),
	}
}

func BuildFlexibleProfiles(candles []exchanges.Candle, windows []FlexibleProfileWindow, cfg BinConfig) []FlexibleProfile {
	out := make([]FlexibleProfile, 0, len(windows))
	for _, window := range windows {
		profile := BuildFlexibleProfile(candles, window, cfg)
		if len(profile.Profile.Bins) == 0 {
			continue
		}
		out = append(out, profile)
	}
	return out
}

func CandlesBetween(candles []exchanges.Candle, startTime int64, endTime int64) []exchanges.Candle {
	if len(candles) == 0 {
		return nil
	}
	if endTime < startTime {
		startTime, endTime = endTime, startTime
	}
	out := make([]exchanges.Candle, 0)
	for _, candle := range candles {
		if candle.StartTime >= startTime && candle.StartTime <= endTime {
			out = append(out, candle)
		}
	}
	return out
}

func ClassifyAcceptance(candles []exchanges.Candle, profile VolumeProfile, eventEndTime int64, lookahead int) string {
	if len(candles) == 0 || profile.VAH <= profile.VAL || lookahead <= 0 {
		return AcceptanceNeutral
	}
	startIndex := -1
	for i, candle := range candles {
		if candle.StartTime >= eventEndTime {
			startIndex = i + 1
			break
		}
	}
	if startIndex < 0 || startIndex >= len(candles) {
		return AcceptanceNeutral
	}
	end := startIndex + lookahead
	if end > len(candles) {
		end = len(candles)
	}
	if end <= startIndex {
		return AcceptanceNeutral
	}
	valueWidth := profile.VAH - profile.VAL
	accepted := 0
	rejected := false
	for i := startIndex; i < end; i++ {
		close := candles[i].CloseFloat()
		if close >= profile.VAL && close <= profile.VAH {
			accepted++
		}
		if i-startIndex < 5 && (close > profile.VAH+valueWidth*0.5 || close < profile.VAL-valueWidth*0.5) {
			rejected = true
		}
	}
	if rejected {
		return AcceptanceRejected
	}
	if float64(accepted)/float64(end-startIndex) >= 0.60 {
		return AcceptanceAccepted
	}
	return AcceptanceNeutral
}
