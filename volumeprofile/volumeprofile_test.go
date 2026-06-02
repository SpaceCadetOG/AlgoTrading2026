package volumeprofile

import (
	"strconv"
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestPriceBinCreation(t *testing.T) {
	bins := BuildPriceBins([]exchanges.Candle{
		vpCandle(100, 125, 100, 120, 10, 1),
	}, BinConfig{BinSize: 10})
	if len(bins) != 3 {
		t.Fatalf("bins=%d want 3 %+v", len(bins), bins)
	}
	if bins[0].Low != 100 || bins[0].High != 110 || bins[0].Mid != 105 {
		t.Fatalf("bad first bin: %+v", bins[0])
	}
}

func TestCandleVolumeDistribution(t *testing.T) {
	bins := []PriceBin{
		{Low: 100, High: 110, Mid: 105},
		{Low: 110, High: 120, Mid: 115},
	}
	out := DistributeCandleVolume([]exchanges.Candle{vpCandle(100, 120, 100, 110, 20, 1)}, bins)
	if out[0].Volume != 10 || out[1].Volume != 10 {
		t.Fatalf("bad distribution: %+v", out)
	}
}

func TestPOCAndValueArea(t *testing.T) {
	profile := VolumeProfile{Bins: []PriceBin{
		{Low: 0, High: 10, Mid: 5, Volume: 10},
		{Low: 10, High: 20, Mid: 15, Volume: 80},
		{Low: 20, High: 30, Mid: 25, Volume: 10},
	}, TotalVolume: 100}
	if poc := PointOfControl(profile); poc != 15 {
		t.Fatalf("poc=%v", poc)
	}
	val, vah := ValueArea(profile, 0.70)
	if val != 10 || vah != 20 {
		t.Fatalf("value area val=%v vah=%v", val, vah)
	}
}

func TestHVNAndLVNDetection(t *testing.T) {
	bins := []PriceBin{
		{Low: 0, High: 10, Volume: 10},
		{Low: 10, High: 20, Volume: 30},
		{Low: 20, High: 30, Volume: 5},
		{Low: 30, High: 40, Volume: 25},
		{Low: 40, High: 50, Volume: 10},
	}
	if hvns := HighVolumeNodes(bins); len(hvns) != 2 {
		t.Fatalf("hvns=%+v", hvns)
	}
	if lvns := LowVolumeNodes(bins); len(lvns) != 1 || lvns[0].Low != 20 {
		t.Fatalf("lvns=%+v", lvns)
	}
}

func TestProfileShapeClassification(t *testing.T) {
	tests := []struct {
		name string
		bins []PriceBin
		want string
	}{
		{name: "d", bins: shapeBins(10, 60, 10), want: ShapeDProfile},
		{name: "p", bins: shapeBins(5, 20, 60), want: ShapePProfile},
		{name: "b", bins: shapeBins(60, 20, 5), want: ShapeBProfile},
		{name: "thin", bins: []PriceBin{{Volume: 10}, {Volume: 11}, {Volume: 10}, {Volume: 9}, {Volume: 10}}, want: ShapeThinProfile},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := VolumeProfile{Bins: tt.bins, TotalVolume: TotalVolume(tt.bins)}
			if got := ClassifyShape(profile); got != tt.want {
				t.Fatalf("shape=%s want %s", got, tt.want)
			}
		})
	}
}

func TestBuildProfile(t *testing.T) {
	profile := BuildProfile([]exchanges.Candle{
		vpCandle(100, 110, 100, 105, 20, 1),
		vpCandle(110, 120, 110, 115, 40, 2),
	}, BinConfig{BinSize: 10})
	if profile.TotalVolume != 60 || profile.POC == 0 || len(profile.Bins) != 2 {
		t.Fatalf("bad profile: %+v", profile)
	}
}

func shapeBins(lower, middle, upper float64) []PriceBin {
	return []PriceBin{
		{Volume: lower},
		{Volume: middle},
		{Volume: upper},
	}
}

func vpCandle(open, high, low, close, volume float64, ts int64) exchanges.Candle {
	return exchanges.Candle{
		Venue:     "test",
		Symbol:    "BTCUSDT",
		Interval:  "15m",
		Open:      vpFloat(open),
		High:      vpFloat(high),
		Low:       vpFloat(low),
		Close:     vpFloat(close),
		Volume:    vpFloat(volume),
		StartTime: ts,
		EndTime:   ts + 900000,
	}
}

func vpFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
