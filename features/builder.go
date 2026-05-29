package features

import (
	"AlgoTrading2026/indicators"
	"AlgoTrading2026/series"
)

func BuildFeatures(venue string, symbol string, interval string, frame series.Frame) []FeatureRow {
	close := frame.CloseColumn()
	high := frame.HighColumn()
	low := frame.LowColumn()
	times := frame.TimeColumn()

	const (
		trendPeriod    = 20
		fastPeriod     = 12
		slowPeriod     = 26
		signalPeriod   = 9
		rsiPeriod      = 14
		momentumPeriod = 10
		stdevFactor    = 2
	)

	sma := indicators.SMA(close, trendPeriod)
	ema := indicators.EMA(close, trendPeriod)
	apo := indicators.APO(close, fastPeriod, slowPeriod)
	macd := indicators.MACD(close, fastPeriod, slowPeriod, signalPeriod)
	rsi := indicators.RSI(close, rsiPeriod)
	momentum := indicators.Momentum(close, momentumPeriod)
	stddev := indicators.StdDev(close, trendPeriod)
	bands := indicators.BollingerBands(close, trendPeriod, stdevFactor)
	sr := indicators.RollingSupportResistance(high, low, trendPeriod)
	hour := indicators.HourOfDay(times)
	day := indicators.DayOfWeek(times)

	rows := make([]FeatureRow, 0, len(frame.Rows))
	for i, row := range frame.Rows {
		rows = append(rows, FeatureRow{
			Time:     row.Time,
			Venue:    venue,
			Symbol:   symbol,
			Interval: interval,

			Close:  row.Close,
			Volume: row.Volume,

			SMA: sma[i],
			EMA: ema[i],
			APO: apo[i],

			MACD:          macd.MACD[i],
			MACDSignal:    macd.Signal[i],
			MACDHistogram: macd.Histogram[i],

			RSI:      rsi[i],
			Momentum: momentum[i],
			StdDev:   stddev[i],

			BollingerMiddle: bands.Middle[i],
			BollingerUpper:  bands.Upper[i],
			BollingerLower:  bands.Lower[i],
			BollingerWidth:  bands.Upper[i] - bands.Lower[i],

			DistanceFromSMA: row.Close - sma[i],
			DistanceFromEMA: row.Close - ema[i],

			Support:                sr.Support[i],
			Resistance:             sr.Resistance[i],
			DistanceFromSupport:    row.Close - sr.Support[i],
			DistanceFromResistance: row.Close - sr.Resistance[i],

			HourOfDay: hour[i],
			DayOfWeek: day[i],
		})
	}

	return rows
}
