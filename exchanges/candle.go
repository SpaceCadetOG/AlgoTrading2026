package exchanges

import "time"

type Candle struct {
	Venue    string
	Symbol   string
	Interval string

	Open   string
	High   string
	Low    string
	Close  string
	Volume string

	StartTime int64 // milliseconds
	EndTime   int64 // milliseconds
	Closed    bool
}

type CandleHandler func(Candle)

type CandleReader interface {
	GetCandles(symbol string, interval string, limit int) ([]Candle, error)
}

type CandleStreamer interface {
	StreamCandles(symbol string, interval string, handler CandleHandler) error
}

func (c Candle) StartUTC() time.Time {
	return time.UnixMilli(c.StartTime).UTC()
}

func (c Candle) EndUTC() time.Time {
	return time.UnixMilli(c.EndTime).UTC()
}

func (c Candle) OpenFloat() float64 {
	return parseFloat(c.Open)
}

func (c Candle) HighFloat() float64 {
	return parseFloat(c.High)
}

func (c Candle) LowFloat() float64 {
	return parseFloat(c.Low)
}

func (c Candle) CloseFloat() float64 {
	return parseFloat(c.Close)
}

func (c Candle) VolumeFloat() float64 {
	return parseFloat(c.Volume)
}

func (c Candle) Range() float64 {
	return c.HighFloat() - c.LowFloat()
}

func (c Candle) Body() float64 {
	body := c.CloseFloat() - c.OpenFloat()
	if body < 0 {
		return -body
	}
	return body
}

func (c Candle) IsBullish() bool {
	return c.CloseFloat() > c.OpenFloat()
}

func (c Candle) IsBearish() bool {
	return c.CloseFloat() < c.OpenFloat()
}