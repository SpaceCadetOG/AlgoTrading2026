package research

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strconv"

	"AlgoTrading2026/scanner"
)

type InstrumentUniverseRow = scanner.Instrument

func BuildInstrumentUniverse(orderBookRows []OrderBookFeatureRow, candleCounts map[string]int) []InstrumentUniverseRow {
	facts := make([]scanner.Instrument, 0)
	for _, row := range orderBookRows {
		facts = append(facts, scanner.Instrument{
			Venue:       row.Venue,
			Symbol:      row.Symbol,
			AssetType:   scanner.ClassifyAsset(row.Symbol),
			Price:       row.Mid,
			SpreadPct:   row.SpreadPct,
			L2Available: row.Valid,
			Notes:       "Read-only L2 snapshot metric available.",
		})
	}
	for key, count := range candleCounts {
		venue, symbol := splitVenueSymbolKey(key)
		facts = append(facts, scanner.Instrument{
			Venue:            venue,
			Symbol:           symbol,
			AssetType:        scanner.ClassifyAsset(symbol),
			CandlesAvailable: count > 0,
			ProfileReady:     count > 0,
			Notes:            "Historical candle sample available for research.",
		})
	}
	return scanner.RankInstruments(scanner.MergeUniverseFacts(scanner.DefaultUniverse(), facts))
}

func WriteInstrumentUniverseCSV(path string, rows []InstrumentUniverseRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{
		"venue", "symbol", "asset_type", "price", "volume_24h", "open_interest", "funding_rate", "spread_pct",
		"volatility_proxy", "l2_available", "candles_available", "trades_available", "profile_ready", "research_score", "rank", "notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Venue,
			row.Symbol,
			row.AssetType,
			floatToString(row.Price),
			floatToString(row.Volume24h),
			floatToString(row.OpenInterest),
			floatToString(row.FundingRate),
			floatToString(row.SpreadPct),
			floatToString(row.VolatilityProxy),
			strconv.FormatBool(row.L2Available),
			strconv.FormatBool(row.CandlesAvailable),
			strconv.FormatBool(row.TradesAvailable),
			strconv.FormatBool(row.ProfileReady),
			floatToString(row.ResearchScore),
			strconv.Itoa(row.Rank),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func splitVenueSymbolKey(key string) (string, string) {
	for i := range key {
		if key[i] == ':' {
			return key[:i], key[i+1:]
		}
	}
	return "", key
}

func WriteInstrumentUniverseSummaryJSON(path string, summary InstrumentUniverseSummary) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}
