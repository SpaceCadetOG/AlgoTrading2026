package series

import (
	"strconv"
	"strings"

	"AlgoTrading2026/exchanges"
)

type Frame struct {
	Rows []Row
}

type Row struct {
	Time   int64
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

func FromCandles(candles []exchanges.Candle) Frame {
	rows := make([]Row, 0, len(candles))
	for _, candle := range candles {
		rows = append(rows, Row{
			Time:   candle.StartTime,
			Open:   safeFloat(candle.Open),
			High:   safeFloat(candle.High),
			Low:    safeFloat(candle.Low),
			Close:  safeFloat(candle.Close),
			Volume: safeFloat(candle.Volume),
		})
	}

	return Frame{Rows: rows}
}

func safeFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}
