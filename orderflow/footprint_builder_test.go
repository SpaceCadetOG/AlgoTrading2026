package orderflow

import (
	"testing"

	"AlgoTrading2026/tradetape"
)

func TestBuildFootprintBarsCreatesOneMinuteBuckets(t *testing.T) {
	rows := []tradetape.TradeTapeRow{
		tradetape.RowFromPrint(builderPrint("aster", "BTCUSDT", 1000, 100, 1, "B")),
		tradetape.RowFromPrint(builderPrint("aster", "BTCUSDT", 2000, 101, 2, "A")),
		tradetape.RowFromPrint(builderPrint("aster", "BTCUSDT", 61000, 102, 3, "B")),
	}
	bars := BuildFootprintBars(rows, DefaultFootprintBuilderConfig())
	if len(bars) != 2 {
		t.Fatalf("bars=%d want 2", len(bars))
	}
	if bars[0].StartTimestamp != 0 || bars[0].EndTimestamp != 59999 {
		t.Fatalf("unexpected first bucket: %+v", bars[0])
	}
	if bars[0].VenueSymbol != "BTCUSDT" || bars[0].CanonicalSymbol != "BTC" {
		t.Fatalf("unexpected symbol identity: %+v", bars[0])
	}
	if bars[1].StartTimestamp != 60000 {
		t.Fatalf("unexpected second bucket: %+v", bars[1])
	}
}

func TestBuildFootprintBarsAggregatesPriceLevelsAndMetrics(t *testing.T) {
	rows := []tradetape.TradeTapeRow{
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 1000, 100, 1, "B")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 2000, 100, 2, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 3000, 101, 4, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 4000, 102, 8, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 5000, 103, 10, "B")),
	}
	bars := BuildFootprintBars(rows, DefaultFootprintBuilderConfig())
	if len(bars) != 1 {
		t.Fatalf("bars=%d want 1", len(bars))
	}
	bar := bars[0]
	if bar.Open != 100 || bar.High != 103 || bar.Low != 100 || bar.Close != 103 {
		t.Fatalf("unexpected OHLC: %+v", bar)
	}
	if bar.TotalBidVolume != 11 || bar.TotalAskVolume != 14 || bar.Volume != 25 || bar.Delta != 3 {
		t.Fatalf("unexpected totals: %+v", bar)
	}
	if len(bar.Levels) != 4 {
		t.Fatalf("levels=%d want 4", len(bar.Levels))
	}
	if bar.Levels[0].BidVolume != 1 || bar.Levels[0].AskVolume != 2 || bar.Levels[0].TotalVolume != 3 || bar.Levels[0].Delta != 1 {
		t.Fatalf("unexpected level aggregation: %+v", bar.Levels[0])
	}
}

func TestBuildFootprintBarsFeedsExistingMetrics(t *testing.T) {
	rows := []tradetape.TradeTapeRow{
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 1000, 100, 1, "B")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 2000, 101, 10, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 3000, 102, 12, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 4000, 103, 14, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 5000, 104, 1, "B")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 6000, 104, 1, "A")),
		tradetape.RowFromPrint(builderPrint("hyperliquid", "BTC", 7000, 100, 1, "A")),
	}
	bar := BuildFootprintBars(rows, DefaultFootprintBuilderConfig())[0]
	hvn := FootprintHVN(bar)
	if hvn.Price != 103 {
		t.Fatalf("hvn=%+v want price 103", hvn)
	}
	if len(DetectVolumeClusters(bar, 1.2)) == 0 {
		t.Fatal("expected volume clusters")
	}
	imbalances := DetectImbalances(bar, 3)
	if len(imbalances) < 3 {
		t.Fatalf("imbalances=%d want at least 3", len(imbalances))
	}
	if len(DetectStackedImbalances(imbalances, 3)) == 0 {
		t.Fatal("expected stacked imbalance")
	}
	unfinished := DetectUnfinishedBusiness(bar)
	if !unfinished.High || !unfinished.Low {
		t.Fatalf("expected unfinished high and low: %+v", unfinished)
	}
}

func builderPrint(venue string, symbol string, timestamp int64, price float64, size float64, side string) tradetape.TradeTapePrint {
	return tradetape.TradeTapePrint{
		Venue:         venue,
		Symbol:        symbol,
		Timestamp:     timestamp,
		Price:         price,
		Size:          size,
		Side:          side,
		AggressorSide: side,
		TradeID:       venue + symbol + side,
	}
}
