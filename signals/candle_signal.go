package signals

import "AlgoTrading2026/exchanges"

type SignalDirection string

const (
	SignalBullish SignalDirection = "bullish"
	SignalBearish SignalDirection = "bearish"
	SignalNeutral SignalDirection = "neutral"
)

type CandleSignal struct {
	Venue     string
	Symbol    string
	Interval  string
	Direction SignalDirection

	Close  float64
	Range  float64
	Body   float64
	Volume float64

	BodyPctOfRange float64
	StartTime      int64
	EndTime        int64
}

func FromCandle(c exchanges.Candle) CandleSignal {
	direction := SignalNeutral

	if c.IsBullish() {
		direction = SignalBullish
	} else if c.IsBearish() {
		direction = SignalBearish
	}

	bodyPct := 0.0
	if c.Range() != 0 {
		bodyPct = c.Body() / c.Range() * 100
	}

	return CandleSignal{
		Venue:          c.Venue,
		Symbol:         c.Symbol,
		Interval:       c.Interval,
		Direction:      direction,
		Close:          c.CloseFloat(),
		Range:          c.Range(),
		Body:           c.Body(),
		Volume:         c.VolumeFloat(),
		BodyPctOfRange: bodyPct,
		StartTime:      c.StartTime,
		EndTime:        c.EndTime,
	}
}
