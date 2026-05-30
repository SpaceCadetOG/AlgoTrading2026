package backtest

import (
	"fmt"
	"sort"
	"time"

	"AlgoTrading2026/exchanges"
)

type SimulatedClock struct {
	current time.Time
}

func NewSimulatedClock(start time.Time) *SimulatedClock {
	return &SimulatedClock{current: start.UTC()}
}

func (c *SimulatedClock) Current() time.Time {
	return c.current
}

func (c *SimulatedClock) Advance(duration time.Duration) error {
	if duration < 0 {
		return fmt.Errorf("cannot advance simulated clock backward")
	}
	c.current = c.current.Add(duration)
	return nil
}

func (c *SimulatedClock) Set(next time.Time) error {
	next = next.UTC()
	if next.Before(c.current) {
		return fmt.Errorf("cannot move simulated clock backward from %s to %s", c.current, next)
	}
	c.current = next
	return nil
}

func (c *SimulatedClock) Reset(next time.Time) {
	c.current = next.UTC()
}

func (c *SimulatedClock) StepTo(next time.Time) error {
	return c.Set(next)
}

type CandleTimeIterator struct {
	clock   *SimulatedClock
	candles []exchanges.Candle
	index   int
}

func NewCandleTimeIterator(clock *SimulatedClock, candles []exchanges.Candle) *CandleTimeIterator {
	ordered := append([]exchanges.Candle(nil), candles...)
	sort.SliceStable(ordered, func(i int, j int) bool {
		return ordered[i].StartTime < ordered[j].StartTime
	})
	if clock == nil {
		start := time.Time{}
		if len(ordered) > 0 {
			start = time.UnixMilli(ordered[0].StartTime).UTC()
		}
		clock = NewSimulatedClock(start)
	}
	return &CandleTimeIterator{clock: clock, candles: ordered}
}

func (i *CandleTimeIterator) Next() (exchanges.Candle, bool, error) {
	if i.index >= len(i.candles) {
		return exchanges.Candle{}, false, nil
	}

	candle := i.candles[i.index]
	i.index++
	if err := i.clock.StepTo(time.UnixMilli(candle.StartTime).UTC()); err != nil {
		return exchanges.Candle{}, false, err
	}

	return candle, true, nil
}

func (i *CandleTimeIterator) Clock() *SimulatedClock {
	return i.clock
}

func (i *CandleTimeIterator) Len() int {
	return len(i.candles)
}
