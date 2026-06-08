package tradetape

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type TradeTapeAnalysis struct {
	Rows              int      `json:"rows"`
	ValidRows         int      `json:"validRows"`
	ErrorRows         int      `json:"errorRows"`
	Venues            int      `json:"venues"`
	Symbols           int      `json:"symbols"`
	VenueSymbols      int      `json:"venueSymbols"`
	CanonicalSymbols  int      `json:"canonicalSymbols"`
	TotalVolume       float64  `json:"totalVolume"`
	BuyTrades         int      `json:"buyTrades"`
	SellTrades        int      `json:"sellTrades"`
	UnknownSideTrades int      `json:"unknownSideTrades"`
	AverageTradeSize  float64  `json:"averageTradeSize"`
	LargestTrade      float64  `json:"largestTrade"`
	FirstTimestamp    int64    `json:"firstTimestamp"`
	LastTimestamp     int64    `json:"lastTimestamp"`
	Notes             []string `json:"notes"`
}

func AnalyzeCSV(path string) (TradeTapeAnalysis, error) {
	rows, err := ReadRows(path)
	if err != nil {
		return TradeTapeAnalysis{}, err
	}
	return AnalyzeRows(rows), nil
}

func AnalyzeRows(rows []TradeTapeRow) TradeTapeAnalysis {
	analysis := TradeTapeAnalysis{
		Rows: len(rows),
		Notes: []string{
			"Trade tape analysis uses normalized recorded prints.",
			"Footprint bars are not built in this phase.",
			"Aggressor-side UNKNOWN means venue semantics still need verification or later inference.",
		},
	}
	venues := map[string]bool{}
	venueSymbols := map[string]bool{}
	canonicalSymbols := map[string]bool{}
	for _, row := range rows {
		row.Print = NormalizePrintSymbols(row.Print)
		if row.Print.Venue != "" {
			venues[row.Print.Venue] = true
		}
		if row.Print.VenueSymbol != "" {
			venueSymbols[row.Print.VenueSymbol] = true
		}
		if row.Print.CanonicalSymbol != "" {
			canonicalSymbols[row.Print.CanonicalSymbol] = true
		}
		if !row.Valid || row.Error != "" {
			analysis.ErrorRows++
			continue
		}
		analysis.ValidRows++
		analysis.TotalVolume += row.Print.Size
		if row.Print.Size > analysis.LargestTrade {
			analysis.LargestTrade = row.Print.Size
		}
		if analysis.FirstTimestamp == 0 || row.Print.Timestamp < analysis.FirstTimestamp {
			analysis.FirstTimestamp = row.Print.Timestamp
		}
		if row.Print.Timestamp > analysis.LastTimestamp {
			analysis.LastTimestamp = row.Print.Timestamp
		}
		switch normalizedSide(row.Print.Side, row.Print.AggressorSide) {
		case "buy":
			analysis.BuyTrades++
		case "sell":
			analysis.SellTrades++
		default:
			analysis.UnknownSideTrades++
		}
	}
	analysis.Venues = len(venues)
	analysis.Symbols = len(venueSymbols)
	analysis.VenueSymbols = len(venueSymbols)
	analysis.CanonicalSymbols = len(canonicalSymbols)
	if analysis.ValidRows > 0 {
		analysis.AverageTradeSize = analysis.TotalVolume / float64(analysis.ValidRows)
	}
	return analysis
}

func WriteAnalysisJSON(path string, analysis TradeTapeAnalysis) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteAnalysisMarkdown(path string, analysis TradeTapeAnalysis) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(AnalysisMarkdown(analysis)), 0644)
}

func AnalysisMarkdown(analysis TradeTapeAnalysis) string {
	var b strings.Builder
	b.WriteString("# Trade Tape Analysis\n\n")
	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- Rows: %d\n", analysis.Rows)
	fmt.Fprintf(&b, "- Valid rows: %d\n", analysis.ValidRows)
	fmt.Fprintf(&b, "- Error rows: %d\n", analysis.ErrorRows)
	fmt.Fprintf(&b, "- Venues: %d\n", analysis.Venues)
	fmt.Fprintf(&b, "- Venue symbols: %d\n", analysis.VenueSymbols)
	fmt.Fprintf(&b, "- Canonical symbols: %d\n", analysis.CanonicalSymbols)
	fmt.Fprintf(&b, "- Total volume: %.8f\n", analysis.TotalVolume)
	fmt.Fprintf(&b, "- Average trade size: %.8f\n", analysis.AverageTradeSize)
	fmt.Fprintf(&b, "- Largest trade: %.8f\n", analysis.LargestTrade)
	fmt.Fprintf(&b, "- First timestamp: %d\n", analysis.FirstTimestamp)
	fmt.Fprintf(&b, "- Last timestamp: %d\n\n", analysis.LastTimestamp)
	b.WriteString("## Side Counts\n\n")
	fmt.Fprintf(&b, "- Buy trades: %d\n", analysis.BuyTrades)
	fmt.Fprintf(&b, "- Sell trades: %d\n", analysis.SellTrades)
	fmt.Fprintf(&b, "- Unknown side trades: %d\n\n", analysis.UnknownSideTrades)
	b.WriteString("## Notes\n\n")
	for _, note := range analysis.Notes {
		fmt.Fprintf(&b, "- %s\n", note)
	}
	return b.String()
}

func normalizedSide(values ...string) string {
	for _, value := range values {
		switch strings.ToUpper(strings.TrimSpace(value)) {
		case "B", "BUY", "BID", "LONG":
			return "buy"
		case "A", "ASK", "SELL", "S", "SHORT":
			return "sell"
		}
	}
	return "unknown"
}

func SortedVenues(rows []TradeTapeRow) []string {
	seen := map[string]bool{}
	for _, row := range rows {
		if row.Print.Venue != "" {
			seen[row.Print.Venue] = true
		}
	}
	out := make([]string, 0, len(seen))
	for venue := range seen {
		out = append(out, venue)
	}
	sort.Strings(out)
	return out
}
