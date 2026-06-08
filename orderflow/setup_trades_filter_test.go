package orderflow

import "testing"

func TestRollingAverageTradeSizeAndLargeTrade(t *testing.T) {
	trades := []TradesFilterTrade{
		tradesFilterTrade(1, 100, 1, "BUY"),
		tradesFilterTrade(2, 100, 2, "BUY"),
		tradesFilterTrade(3, 100, 9, "BUY"),
	}
	avg := RollingAverageTradeSize(trades, 2, 20)
	if avg != 1.5 {
		t.Fatalf("expected avg 1.5, got %f", avg)
	}
	if !IsLargeTrade(9, avg, 3) {
		t.Fatalf("expected large trade")
	}
}

func TestDetectTradesFilterSetupsBuyAccepted(t *testing.T) {
	trades := []TradesFilterTrade{
		tradesFilterTrade(1, 100, 1, "BUY"),
		tradesFilterTrade(2, 100, 1, "BUY"),
		tradesFilterTrade(3, 100, 10, "BUY"),
	}
	footprints := []TradesFilterFootprint{
		tradesFilterFootprint(0, 59999, 100, 101, 99, 101, 5),
		tradesFilterFootprint(60000, 119999, 100, 102, 99, 101.5, 4),
	}
	got := DetectTradesFilterSetups(trades, footprints, nil, DefaultTradesFilterConfig())
	if len(got) != 1 {
		t.Fatalf("expected 1 setup, got %d", len(got))
	}
	if !got[0].BuyAggressor || !got[0].LongContext || !got[0].Accepted || got[0].Rejected {
		t.Fatalf("unexpected setup state: %+v", got[0])
	}
	if !got[0].NearHVN || !got[0].NearVolumeCluster {
		t.Fatalf("expected HVN and volume cluster proximity: %+v", got[0])
	}
}

func TestDetectTradesFilterSetupsSellRejected(t *testing.T) {
	trades := []TradesFilterTrade{
		tradesFilterTrade(1, 100, 1, "SELL"),
		tradesFilterTrade(2, 100, 1, "SELL"),
		tradesFilterTrade(3, 100, 10, "SELL"),
	}
	footprints := []TradesFilterFootprint{
		tradesFilterFootprint(0, 59999, 100, 101, 99, 100.5, -5),
		tradesFilterFootprint(60000, 119999, 100, 102, 99, 101.5, 4),
	}
	got := DetectTradesFilterSetups(trades, footprints, nil, DefaultTradesFilterConfig())
	if len(got) != 1 {
		t.Fatalf("expected 1 setup, got %d", len(got))
	}
	if !got[0].SellAggressor || !got[0].ShortContext || got[0].Accepted || !got[0].Rejected {
		t.Fatalf("unexpected setup state: %+v", got[0])
	}
}

func TestDetectTradesFilterSetupsMultipleHVNProximity(t *testing.T) {
	trades := []TradesFilterTrade{
		tradesFilterTrade(1, 100, 1, "BUY"),
		tradesFilterTrade(2, 100, 1, "BUY"),
		tradesFilterTrade(3, 105, 10, "BUY"),
	}
	footprints := []TradesFilterFootprint{
		tradesFilterFootprint(0, 59999, 100, 106, 99, 105, 1),
		tradesFilterFootprint(60000, 119999, 100, 106, 99, 105.5, 1),
	}
	zones := []TradesFilterZone{{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		ZoneLow: 104.9, ZoneHigh: 105.1,
	}}
	got := DetectTradesFilterSetups(trades, footprints, zones, DefaultTradesFilterConfig())
	if len(got) != 1 || !got[0].NearMultipleHVN {
		t.Fatalf("expected multiple HVN proximity setup, got %+v", got)
	}
}

func TestDetectTradesFilterSetupsFollowThrough(t *testing.T) {
	trades := []TradesFilterTrade{
		tradesFilterTrade(1, 100, 1, "BUY"),
		tradesFilterTrade(2, 100, 1, "BUY"),
		tradesFilterTrade(3, 100, 10, "BUY"),
	}
	footprints := []TradesFilterFootprint{tradesFilterFootprint(0, 59999, 100, 101, 99, 100, 1)}
	for i := 1; i <= 20; i++ {
		close := 100 + float64(i)
		footprints = append(footprints, tradesFilterFootprint(int64(i)*60000, int64(i)*60000+59999, 100, 125, 99, close, 1))
	}
	got := DetectTradesFilterSetups(trades, footprints, nil, DefaultTradesFilterConfig())
	if len(got) != 1 {
		t.Fatalf("expected 1 setup")
	}
	if got[0].FollowThrough5 != 5 || got[0].FollowThrough10 != 10 || got[0].FollowThrough20 != 20 {
		t.Fatalf("unexpected follow through: %+v", got[0])
	}
}

func tradesFilterTrade(timestamp int64, price float64, size float64, side string) TradesFilterTrade {
	return TradesFilterTrade{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		Timestamp: timestamp, Price: price, Size: size, Side: side, AggressorSide: side,
	}
}

func tradesFilterFootprint(start int64, end int64, hvn float64, high float64, low float64, close float64, delta float64) TradesFilterFootprint {
	return TradesFilterFootprint{
		Venue: "aster", VenueSymbol: "BTCUSDT", CanonicalSymbol: "BTC",
		StartTimestamp: start, EndTimestamp: end, High: high, Low: low, Close: close,
		HVNPrice: hvn, VolumeClusterCount: 1, Delta: delta, ImbalanceCount: 2, StackedImbalanceCount: 1,
	}
}
