package tradetape

import "strconv"

type DedupeSummary struct {
	InputRows     int `json:"inputRows"`
	OutputRows    int `json:"outputRows"`
	DuplicateRows int `json:"duplicateRows"`
	InvalidRows   int `json:"invalidRows"`
}

func DeduplicatePrints(prints []TradeTapePrint) ([]TradeTapePrint, DedupeSummary) {
	summary := DedupeSummary{InputRows: len(prints)}
	seen := map[string]bool{}
	out := make([]TradeTapePrint, 0, len(prints))
	for _, print := range prints {
		print = NormalizePrintSymbols(print)
		if err := ValidatePrint(print); err != nil {
			summary.InvalidRows++
			continue
		}
		key := dedupeKey(print)
		if seen[key] {
			summary.DuplicateRows++
			continue
		}
		seen[key] = true
		out = append(out, print)
	}
	summary.OutputRows = len(out)
	return out, summary
}

func dedupeKey(print TradeTapePrint) string {
	print = NormalizePrintSymbols(print)
	if print.TradeID != "" {
		return print.Venue + "|" + print.CanonicalSymbol + "|" + print.TradeID
	}
	return print.Venue + "|" + print.CanonicalSymbol + "|" +
		strconv.FormatInt(print.Timestamp, 10) + "|" +
		strconv.FormatFloat(print.Price, 'f', -1, 64) + "|" +
		strconv.FormatFloat(print.Size, 'f', -1, 64) + "|" +
		print.Side
}
