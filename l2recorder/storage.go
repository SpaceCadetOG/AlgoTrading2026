package l2recorder

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
)

type CSVStorage struct {
	path string
}

func NewCSVStorage(path string) CSVStorage {
	return CSVStorage{path: path}
}

func (s CSVStorage) AppendRows(rows []SnapshotRow) error {
	if len(rows) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	needsHeader := true
	if info, err := os.Stat(s.path); err == nil && info.Size() > 0 {
		needsHeader = false
	}

	file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if needsHeader {
		if err := writer.Write(snapshotRowHeader()); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := writer.Write(snapshotRowCSV(row)); err != nil {
			return err
		}
	}
	return writer.Error()
}

func snapshotRowHeader() []string {
	return []string{
		"timestamp",
		"venue",
		"symbol",
		"best_bid",
		"best_ask",
		"spread",
		"spread_pct",
		"mid",
		"bid_depth_1pct",
		"ask_depth_1pct",
		"imbalance_1pct",
		"valid",
		"error",
	}
}

func snapshotRowCSV(row SnapshotRow) []string {
	return []string{
		strconv.FormatInt(row.Timestamp, 10),
		row.Venue,
		row.Symbol,
		floatString(row.BestBid),
		floatString(row.BestAsk),
		floatString(row.Spread),
		floatString(row.SpreadPct),
		floatString(row.Mid),
		floatString(row.BidDepth1Pct),
		floatString(row.AskDepth1Pct),
		floatString(row.Imbalance1Pct),
		strconv.FormatBool(row.Valid),
		row.Error,
	}
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}
