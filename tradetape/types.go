package tradetape

import (
	"fmt"

	"AlgoTrading2026/symbols"
)

const UnknownAggressorSide = "UNKNOWN"

type TradeTapePrint struct {
	Venue           string
	Symbol          string
	VenueSymbol     string
	CanonicalSymbol string
	MarketID        string
	Timestamp       int64
	Price           float64
	Size            float64
	Side            string
	AggressorSide   string
	TradeID         string
}

func NormalizePrintSymbols(print TradeTapePrint) TradeTapePrint {
	if print.VenueSymbol == "" {
		print.VenueSymbol = print.Symbol
	}
	identity := symbols.NormalizeSymbol(print.Venue, print.VenueSymbol)
	if print.CanonicalSymbol == "" {
		print.CanonicalSymbol = identity.CanonicalSymbol
	}
	if print.Symbol == "" {
		print.Symbol = print.VenueSymbol
	}
	return print
}

func ValidatePrint(print TradeTapePrint) error {
	print = NormalizePrintSymbols(print)
	if print.Venue == "" {
		return fmt.Errorf("venue is required")
	}
	if print.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if print.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if print.Size < 0 {
		return fmt.Errorf("size must be non-negative")
	}
	if print.Timestamp <= 0 {
		return fmt.Errorf("timestamp must be positive")
	}
	return nil
}
