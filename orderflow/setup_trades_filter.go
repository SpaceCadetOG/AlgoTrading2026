package orderflow

import "strings"

type TradesFilterSetup struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string

	Timestamp int64

	Price float64
	Size  float64

	LargeTrade bool

	NearHVN           bool
	NearVolumeCluster bool
	NearMultipleHVN   bool

	BuyAggressor  bool
	SellAggressor bool

	LongContext  bool
	ShortContext bool

	Accepted bool
	Rejected bool

	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64

	Delta float64

	ImbalanceCount        int
	StackedImbalanceCount int
}

type TradesFilterTrade struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	Timestamp       int64
	Price           float64
	Size            float64
	Side            string
	AggressorSide   string
}

type TradesFilterFootprint struct {
	Venue                 string
	VenueSymbol           string
	CanonicalSymbol       string
	StartTimestamp        int64
	EndTimestamp          int64
	High                  float64
	Low                   float64
	Close                 float64
	HVNPrice              float64
	VolumeClusterCount    int
	Delta                 float64
	ImbalanceCount        int
	StackedImbalanceCount int
}

type TradesFilterZone struct {
	Venue           string
	VenueSymbol     string
	CanonicalSymbol string
	StartTimestamp  int64
	EndTimestamp    int64
	ZoneLow         float64
	ZoneHigh        float64
}

type TradesFilterConfig struct {
	SizeMultiplier float64
	RollingWindow  int
	ProximityPct   float64
}

func DefaultTradesFilterConfig() TradesFilterConfig {
	return TradesFilterConfig{
		SizeMultiplier: 3.0,
		RollingWindow:  20,
		ProximityPct:   0.0025,
	}
}

func DetectTradesFilterSetups(trades []TradesFilterTrade, footprints []TradesFilterFootprint, zones []TradesFilterZone, config TradesFilterConfig) []TradesFilterSetup {
	if config.SizeMultiplier <= 0 {
		config.SizeMultiplier = DefaultTradesFilterConfig().SizeMultiplier
	}
	if config.RollingWindow <= 0 {
		config.RollingWindow = DefaultTradesFilterConfig().RollingWindow
	}
	if config.ProximityPct <= 0 {
		config.ProximityPct = DefaultTradesFilterConfig().ProximityPct
	}
	setups := make([]TradesFilterSetup, 0)
	for i, trade := range trades {
		averageSize := RollingAverageTradeSize(trades, i, config.RollingWindow)
		if !IsLargeTrade(trade.Size, averageSize, config.SizeMultiplier) {
			continue
		}
		barIndex := matchingFootprintIndex(trade, footprints)
		if barIndex < 0 {
			continue
		}
		footprint := footprints[barIndex]
		nearHVN := withinPct(trade.Price, footprint.HVNPrice, config.ProximityPct)
		nearVolumeCluster := footprint.VolumeClusterCount > 0 && nearHVN
		nearMultipleHVN := nearAnyTradesFilterZone(trade, zones, config.ProximityPct)
		if !nearHVN && !nearVolumeCluster && !nearMultipleHVN {
			continue
		}
		setup := TradesFilterSetup{
			Venue:                 trade.Venue,
			VenueSymbol:           trade.VenueSymbol,
			CanonicalSymbol:       trade.CanonicalSymbol,
			Timestamp:             trade.Timestamp,
			Price:                 trade.Price,
			Size:                  trade.Size,
			LargeTrade:            true,
			NearHVN:               nearHVN,
			NearVolumeCluster:     nearVolumeCluster,
			NearMultipleHVN:       nearMultipleHVN,
			BuyAggressor:          isBuyAggressor(trade),
			SellAggressor:         isSellAggressor(trade),
			Delta:                 footprint.Delta,
			ImbalanceCount:        footprint.ImbalanceCount,
			StackedImbalanceCount: footprint.StackedImbalanceCount,
		}
		setup.LongContext = setup.BuyAggressor || (!setup.SellAggressor && footprint.Delta > 0)
		setup.ShortContext = setup.SellAggressor || (!setup.BuyAggressor && footprint.Delta < 0)
		outcomeIndex := nextSameMarketFootprint(footprints, barIndex)
		if outcomeIndex >= 0 {
			outcome := footprints[outcomeIndex]
			if setup.LongContext {
				setup.Accepted = outcome.Close >= trade.Price
				setup.Rejected = outcome.Close < trade.Price
			}
			if setup.ShortContext {
				setup.Accepted = outcome.Close <= trade.Price
				setup.Rejected = outcome.Close > trade.Price
			}
		}
		setup.FollowThrough5 = tradesFilterFollowThrough(footprints, barIndex, 5, trade.Price, setup.LongContext, setup.ShortContext)
		setup.FollowThrough10 = tradesFilterFollowThrough(footprints, barIndex, 10, trade.Price, setup.LongContext, setup.ShortContext)
		setup.FollowThrough20 = tradesFilterFollowThrough(footprints, barIndex, 20, trade.Price, setup.LongContext, setup.ShortContext)
		setups = append(setups, setup)
	}
	return setups
}

func RollingAverageTradeSize(trades []TradesFilterTrade, index int, window int) float64 {
	if index <= 0 || len(trades) == 0 {
		return 0
	}
	if window <= 0 {
		window = DefaultTradesFilterConfig().RollingWindow
	}
	start := index - window
	if start < 0 {
		start = 0
	}
	var total float64
	var count int
	for i := start; i < index; i++ {
		if sameTradesFilterMarket(trades[index], trades[i]) && trades[i].Size >= 0 {
			total += trades[i].Size
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func IsLargeTrade(size float64, averageSize float64, multiplier float64) bool {
	return size > 0 && averageSize > 0 && size >= averageSize*multiplier
}

func matchingFootprintIndex(trade TradesFilterTrade, footprints []TradesFilterFootprint) int {
	for i, footprint := range footprints {
		if !sameTradeFootprintMarket(trade, footprint) {
			continue
		}
		if trade.Timestamp >= footprint.StartTimestamp && trade.Timestamp <= footprint.EndTimestamp {
			return i
		}
	}
	return -1
}

func nearAnyTradesFilterZone(trade TradesFilterTrade, zones []TradesFilterZone, proximityPct float64) bool {
	for _, zone := range zones {
		if trade.Venue != zone.Venue || trade.VenueSymbol != zone.VenueSymbol || trade.CanonicalSymbol != zone.CanonicalSymbol {
			continue
		}
		mid := (zone.ZoneLow + zone.ZoneHigh) / 2
		if mid <= 0 {
			continue
		}
		buffer := mid * proximityPct
		if trade.Price >= zone.ZoneLow-buffer && trade.Price <= zone.ZoneHigh+buffer {
			return true
		}
	}
	return false
}

func withinPct(price float64, level float64, proximityPct float64) bool {
	if price <= 0 || level <= 0 {
		return false
	}
	buffer := level * proximityPct
	return price >= level-buffer && price <= level+buffer
}

func isBuyAggressor(trade TradesFilterTrade) bool {
	side := strings.ToUpper(firstNonEmptyOrderFlow(trade.AggressorSide, trade.Side))
	return side == "BUY"
}

func isSellAggressor(trade TradesFilterTrade) bool {
	side := strings.ToUpper(firstNonEmptyOrderFlow(trade.AggressorSide, trade.Side))
	return side == "SELL"
}

func nextSameMarketFootprint(footprints []TradesFilterFootprint, start int) int {
	for i := start + 1; i < len(footprints); i++ {
		if sameTradesFilterFootprintMarket(footprints[start], footprints[i]) {
			return i
		}
	}
	return -1
}

func tradesFilterFollowThrough(footprints []TradesFilterFootprint, start int, horizon int, price float64, longContext bool, shortContext bool) float64 {
	count := 0
	for i := start + 1; i < len(footprints); i++ {
		if !sameTradesFilterFootprintMarket(footprints[start], footprints[i]) {
			continue
		}
		count++
		if count == horizon {
			if longContext {
				return footprints[i].Close - price
			}
			if shortContext {
				return price - footprints[i].Close
			}
			return footprints[i].Close - price
		}
	}
	return 0
}

func sameTradesFilterMarket(a TradesFilterTrade, b TradesFilterTrade) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}

func sameTradeFootprintMarket(trade TradesFilterTrade, footprint TradesFilterFootprint) bool {
	return trade.Venue == footprint.Venue && trade.VenueSymbol == footprint.VenueSymbol && trade.CanonicalSymbol == footprint.CanonicalSymbol
}

func sameTradesFilterFootprintMarket(a TradesFilterFootprint, b TradesFilterFootprint) bool {
	return a.Venue == b.Venue && a.VenueSymbol == b.VenueSymbol && a.CanonicalSymbol == b.CanonicalSymbol
}

func firstNonEmptyOrderFlow(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
