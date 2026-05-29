package pairs

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
)

func WriteCandidatesCSV(path string, rows []PairCandidate) error {
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
		"venue_a",
		"symbol_a",
		"venue_b",
		"symbol_b",
		"samples",
		"correlation",
		"latest_spread",
		"latest_zscore",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			row.VenueA,
			row.SymbolA,
			row.VenueB,
			row.SymbolB,
			strconv.Itoa(row.Samples),
			formatFloat(row.Correlation),
			formatFloat(row.LatestSpread),
			formatFloat(row.LatestZScore),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}
