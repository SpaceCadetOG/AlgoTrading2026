package priceaction

import "AlgoTrading2026/exchanges"

func OpenDriveSetups(candles []exchanges.Candle) []StrategySetup {
	aggression := Aggression(candles, 20)
	sideways := Sideways(candles, 8)
	initiations := Initiations(candles, aggression, sideways)
	out := make([]StrategySetup, 0)

	var activeLevel float64
	var activeDirection StudyDirection
	var activeIndex int
	for i, initiation := range initiations {
		if initiation.Detected {
			activeLevel = initiation.StartPrice
			activeIndex = i
			activeDirection = StudyLong
			if initiation.Direction == InitiationDown {
				activeDirection = StudyShort
			}
		}
		if activeLevel == 0 || i <= activeIndex {
			continue
		}
		if touchesLevel(candles[i], activeLevel) {
			out = append(out, setupAt(candles, i, "open_drive", activeDirection, activeLevel, activeLevel, true, invalidatedByClose(candles[i], activeDirection, activeLevel), "return to initiating candle open after one-sided drive"))
			activeLevel = 0
		}
	}
	return out
}
