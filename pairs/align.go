package pairs

import (
	"fmt"
	"sort"

	"AlgoTrading2026/exchanges"
)

type AlignedPoint struct {
	Time   int64
	AClose float64
	BClose float64
}

func AlignByTime(a []exchanges.Candle, b []exchanges.Candle) ([]AlignedPoint, error) {
	aByTime := make(map[int64]float64, len(a))
	for _, candle := range a {
		aByTime[candle.StartTime] = candle.CloseFloat()
	}

	var points []AlignedPoint
	for _, candle := range b {
		aClose, ok := aByTime[candle.StartTime]
		if !ok {
			continue
		}
		points = append(points, AlignedPoint{
			Time:   candle.StartTime,
			AClose: aClose,
			BClose: candle.CloseFloat(),
		})
	}

	sort.Slice(points, func(i int, j int) bool {
		return points[i].Time < points[j].Time
	})

	if len(points) == 0 {
		return nil, fmt.Errorf("no shared candle timestamps")
	}
	return points, nil
}
