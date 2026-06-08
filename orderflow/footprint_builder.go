package orderflow

import (
	"sort"
	"strings"

	"AlgoTrading2026/tradetape"
)

const OneMinuteMillis int64 = 60 * 1000

type FootprintBuilderConfig struct {
	BarMillis int64
}

func DefaultFootprintBuilderConfig() FootprintBuilderConfig {
	return FootprintBuilderConfig{BarMillis: OneMinuteMillis}
}

func BuildFootprintBars(rows []tradetape.TradeTapeRow, config FootprintBuilderConfig) []FootprintBar {
	if config.BarMillis <= 0 {
		config = DefaultFootprintBuilderConfig()
	}
	validRows := make([]tradetape.TradeTapeRow, 0, len(rows))
	for _, row := range rows {
		if row.Valid && row.Error == "" && tradetape.ValidatePrint(row.Print) == nil {
			validRows = append(validRows, row)
		}
	}
	sort.Slice(validRows, func(i int, j int) bool {
		if validRows[i].Print.Timestamp == validRows[j].Print.Timestamp {
			return validRows[i].Print.TradeID < validRows[j].Print.TradeID
		}
		return validRows[i].Print.Timestamp < validRows[j].Print.Timestamp
	})

	builders := map[string]*barBuilder{}
	keys := make([]string, 0)
	for _, row := range validRows {
		print := tradetape.NormalizePrintSymbols(row.Print)
		start := print.Timestamp - (print.Timestamp % config.BarMillis)
		end := start + config.BarMillis - 1
		key := print.Venue + "|" + print.VenueSymbol + "|" + intKey(start)
		builder := builders[key]
		if builder == nil {
			builder = &barBuilder{
				bar: FootprintBar{
					Timestamp:       start,
					Venue:           print.Venue,
					Symbol:          print.VenueSymbol,
					VenueSymbol:     print.VenueSymbol,
					CanonicalSymbol: print.CanonicalSymbol,
					StartTimestamp:  start,
					EndTimestamp:    end,
					Open:            print.Price,
					High:            print.Price,
					Low:             print.Price,
					Close:           print.Price,
				},
				levels: map[float64]*FootprintLevel{},
			}
			builders[key] = builder
			keys = append(keys, key)
		}
		builder.add(print)
	}

	sort.Slice(keys, func(i int, j int) bool {
		a := builders[keys[i]].bar
		b := builders[keys[j]].bar
		if a.StartTimestamp == b.StartTimestamp {
			if a.Venue == b.Venue {
				return a.Symbol < b.Symbol
			}
			return a.Venue < b.Venue
		}
		return a.StartTimestamp < b.StartTimestamp
	})
	out := make([]FootprintBar, 0, len(keys))
	for _, key := range keys {
		out = append(out, builders[key].finish())
	}
	return out
}

type barBuilder struct {
	bar    FootprintBar
	levels map[float64]*FootprintLevel
}

func (b *barBuilder) add(print tradetape.TradeTapePrint) {
	if print.Price > b.bar.High {
		b.bar.High = print.Price
	}
	if print.Price < b.bar.Low {
		b.bar.Low = print.Price
	}
	b.bar.Close = print.Price
	level := b.levels[print.Price]
	if level == nil {
		level = &FootprintLevel{Price: print.Price}
		b.levels[print.Price] = level
	}
	switch sideClass(print.Side, print.AggressorSide) {
	case "ask":
		level.AskVolume += print.Size
	case "bid":
		level.BidVolume += print.Size
	default:
		level.BidVolume += print.Size
	}
}

func (b *barBuilder) finish() FootprintBar {
	prices := make([]float64, 0, len(b.levels))
	for price := range b.levels {
		prices = append(prices, price)
	}
	sort.Float64s(prices)
	levels := make([]FootprintLevel, 0, len(prices))
	for _, price := range prices {
		level := *b.levels[price]
		level.TotalVolume = level.BidVolume + level.AskVolume
		level.Delta = level.AskVolume - level.BidVolume
		levels = append(levels, level)
	}
	b.bar.Levels = levels
	b.bar.TotalBidVolume = TotalBidVolume(b.bar)
	b.bar.TotalAskVolume = TotalAskVolume(b.bar)
	b.bar.Volume = TotalVolume(b.bar)
	b.bar.Delta = Delta(b.bar)
	return b.bar
}

func sideClass(side string, aggressorSide string) string {
	value := strings.ToUpper(strings.TrimSpace(aggressorSide))
	if value == "" || value == tradetape.UnknownAggressorSide {
		value = strings.ToUpper(strings.TrimSpace(side))
	}
	switch value {
	case "A", "ASK", "SELL", "S", "SHORT":
		return "ask"
	case "B", "BID", "BUY", "LONG":
		return "bid"
	default:
		return "unknown"
	}
}

func intKey(value int64) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	digits := make([]byte, 0, 20)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	if negative {
		digits = append(digits, '-')
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}
