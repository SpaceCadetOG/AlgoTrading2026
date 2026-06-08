package orderflow

import "testing"

func fixtureBar() FootprintBar {
	return FootprintBar{
		Timestamp: 1,
		Symbol:    "BTC",
		Open:      100,
		High:      104,
		Low:       100,
		Close:     103,
		Levels: []FootprintLevel{
			{Price: 100, BidVolume: 10, AskVolume: 10},
			{Price: 101, BidVolume: 30, AskVolume: 5},
			{Price: 102, BidVolume: 40, AskVolume: 10},
			{Price: 103, BidVolume: 5, AskVolume: 25},
			{Price: 104, BidVolume: 5, AskVolume: 20},
		},
	}
}

func TestFootprintVolumesAndDelta(t *testing.T) {
	bar := fixtureBar()
	if TotalBidVolume(bar) != 90 {
		t.Fatalf("bid volume=%f", TotalBidVolume(bar))
	}
	if TotalAskVolume(bar) != 70 {
		t.Fatalf("ask volume=%f", TotalAskVolume(bar))
	}
	if TotalVolume(bar) != 160 {
		t.Fatalf("total volume=%f", TotalVolume(bar))
	}
	if Delta(bar) != -20 {
		t.Fatalf("delta=%f", Delta(bar))
	}
}

func TestCumulativeDelta(t *testing.T) {
	got := CumulativeDelta([]FootprintBar{fixtureBar(), {Levels: []FootprintLevel{{Price: 1, BidVolume: 1, AskVolume: 4}}}})
	if got[0] != -20 || got[1] != -17 {
		t.Fatalf("cumulative delta=%+v", got)
	}
}

func TestFootprintHVN(t *testing.T) {
	hvn := FootprintHVN(fixtureBar())
	if hvn.Price != 102 {
		t.Fatalf("hvn=%+v", hvn)
	}
}

func TestVolumeClusters(t *testing.T) {
	clusters := DetectVolumeClusters(fixtureBar(), 1.4)
	if len(clusters) == 0 || clusters[0].Price != 102 {
		t.Fatalf("clusters=%+v", clusters)
	}
}

func TestImbalancesAndStackedImbalances(t *testing.T) {
	bar := FootprintBar{Levels: []FootprintLevel{
		{Price: 100, BidVolume: 30, AskVolume: 5},
		{Price: 101, BidVolume: 40, AskVolume: 10},
		{Price: 102, BidVolume: 50, AskVolume: 10},
		{Price: 103, BidVolume: 5, AskVolume: 30},
	}}
	imbalances := DetectImbalances(bar, 3)
	if len(imbalances) != 4 {
		t.Fatalf("imbalances=%+v", imbalances)
	}
	stacked := DetectStackedImbalances(imbalances, 3)
	if len(stacked) != 1 || stacked[0].Direction != ImbalanceBid || stacked[0].Count != 3 {
		t.Fatalf("stacked=%+v", stacked)
	}
}

func TestUnfinishedBusiness(t *testing.T) {
	unfinished := DetectUnfinishedBusiness(fixtureBar())
	if !unfinished.High || !unfinished.Low {
		t.Fatalf("unfinished=%+v", unfinished)
	}
}

func TestTradesFilter(t *testing.T) {
	trades := TradesFilter(fixtureBar(), 45)
	if len(trades) != 1 || trades[0].Price != 102 {
		t.Fatalf("large trades=%+v", trades)
	}
}

func TestValidateFootprintBar(t *testing.T) {
	if err := ValidateFootprintBar(fixtureBar()); err != nil {
		t.Fatalf("valid fixture failed: %v", err)
	}
	if err := ValidateFootprintBar(FootprintBar{}); err == nil {
		t.Fatal("empty levels should fail")
	}
	if err := ValidateFootprintBar(FootprintBar{Levels: []FootprintLevel{{Price: -1}}}); err == nil {
		t.Fatal("negative price should fail")
	}
	if err := ValidateFootprintBar(FootprintBar{Levels: []FootprintLevel{{Price: 1, BidVolume: -1}}}); err == nil {
		t.Fatal("negative volume should fail")
	}
}
