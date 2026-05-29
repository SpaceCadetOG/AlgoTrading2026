package signals

type AggregatedSignal struct {
	Venue    string
	Symbol   string
	Interval string

	Direction SignalDirection

	Score float64

	Close  float64
	Range  float64
	Body   float64
	Volume float64

	StartTime int64
	EndTime   int64
}

func ScoreCandle(s CandleSignal) AggregatedSignal {
	score := 0.0

	// Direction bias
	switch s.Direction {

	case SignalBullish:
		score += 25

	case SignalBearish:
		score -= 25
	}

	// Strong candle body
	if s.BodyPctOfRange >= 70 {
		if score > 0 {
			score += 40
		} else {
			score -= 40
		}

	} else if s.BodyPctOfRange >= 50 {
		if score > 0 {
			score += 25
		} else {
			score -= 25
		}

	} else if s.BodyPctOfRange >= 30 {
		if score > 0 {
			score += 10
		} else {
			score -= 10
		}
	}

	// Volatility boost
	if s.Range >= 100 {
		if score > 0 {
			score += 10
		} else if score < 0 {
			score -= 10
		}
	}

	// Clamp score
	if score > 100 {
		score = 100
	}

	if score < -100 {
		score = -100
	}

	return AggregatedSignal{
		Venue:      s.Venue,
		Symbol:     s.Symbol,
		Interval:   s.Interval,
		Direction:  s.Direction,
		Score:      score,
		Close:      s.Close,
		Range:      s.Range,
		Body:       s.Body,
		Volume:     s.Volume,
		StartTime:  s.StartTime,
		EndTime:    s.EndTime,
	}
}