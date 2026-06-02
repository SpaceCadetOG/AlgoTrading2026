package volumeprofile

import (
	"math"

	"AlgoTrading2026/exchanges"
)

type BinConfig struct {
	TickSize float64
	BinSize  float64
}

type PriceBin struct {
	Low    float64
	High   float64
	Mid    float64
	Volume float64
}

func DefaultBinConfig() BinConfig {
	return BinConfig{BinSize: 10.0}
}

func BuildPriceBins(candles []exchanges.Candle, cfg BinConfig) []PriceBin {
	if len(candles) == 0 {
		return nil
	}
	low, high := priceRange(candles)
	if low <= 0 || high <= 0 || high < low {
		return nil
	}
	binSize := effectiveBinSize(low, high, cfg)
	start := math.Floor(low/binSize) * binSize
	end := math.Ceil(high/binSize) * binSize
	if end <= start {
		end = start + binSize
	}
	count := int(math.Ceil((end - start) / binSize))
	bins := make([]PriceBin, 0, count)
	for i := 0; i < count; i++ {
		binLow := start + float64(i)*binSize
		binHigh := binLow + binSize
		bins = append(bins, PriceBin{
			Low:  binLow,
			High: binHigh,
			Mid:  (binLow + binHigh) / 2,
		})
	}
	return bins
}

func DistributeCandleVolume(candles []exchanges.Candle, bins []PriceBin) []PriceBin {
	out := append([]PriceBin(nil), bins...)
	for _, candle := range candles {
		volume := candle.VolumeFloat()
		if volume <= 0 {
			continue
		}
		indices := overlappingBinIndexes(out, candle.LowFloat(), candle.HighFloat(), candle.CloseFloat())
		if len(indices) == 0 {
			continue
		}
		share := volume / float64(len(indices))
		for _, index := range indices {
			out[index].Volume += share
		}
	}
	return out
}

func priceRange(candles []exchanges.Candle) (float64, float64) {
	low := math.MaxFloat64
	high := 0.0
	for _, candle := range candles {
		candleLow := candle.LowFloat()
		candleHigh := candle.HighFloat()
		if candleLow > 0 && candleLow < low {
			low = candleLow
		}
		if candleHigh > high {
			high = candleHigh
		}
	}
	if low == math.MaxFloat64 {
		low = 0
	}
	return low, high
}

func effectiveBinSize(low float64, high float64, cfg BinConfig) float64 {
	if cfg.BinSize > 0 {
		return cfg.BinSize
	}
	if cfg.TickSize > 0 {
		return cfg.TickSize
	}
	priceRange := high - low
	if priceRange <= 0 {
		return 10
	}
	derived := priceRange / 100
	if derived < 1 {
		return 1
	}
	return math.Ceil(derived)
}

func overlappingBinIndexes(bins []PriceBin, low float64, high float64, close float64) []int {
	if len(bins) == 0 {
		return nil
	}
	if low <= 0 || high <= 0 || high < low {
		low = close
		high = close
	}
	out := make([]int, 0)
	for i, bin := range bins {
		if high >= bin.Low && low < bin.High {
			out = append(out, i)
		}
	}
	if len(out) == 0 {
		for i, bin := range bins {
			if close >= bin.Low && close < bin.High {
				return []int{i}
			}
		}
	}
	return out
}
