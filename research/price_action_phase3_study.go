package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/priceaction"
)

type PriceActionPhase3StudyRow struct {
	Timestamp       int64
	Strategy        string
	Direction       priceaction.Phase3Direction
	Level           float64
	LevelType       string
	Confirmation    bool
	Invalidated     bool
	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64
	Notes           string
}

func BuildPriceActionPhase3StudyRows(candles []exchanges.Candle) []PriceActionPhase3StudyRow {
	setups := make([]priceaction.Phase3Setup, 0)
	setups = append(setups, priceaction.HighLowRetestSetups(candles, 2)...)
	setups = append(setups, priceaction.StrongWeakHighLowSetups(candles)...)
	setups = append(setups, priceaction.FailedAuctionSetups(candles)...)

	rows := make([]PriceActionPhase3StudyRow, 0, len(setups))
	for _, setup := range setups {
		rows = append(rows, PriceActionPhase3StudyRow{
			Timestamp:       setup.Timestamp,
			Strategy:        setup.Strategy,
			Direction:       setup.Direction,
			Level:           setup.Level,
			LevelType:       setup.LevelType,
			Confirmation:    setup.Confirmation,
			Invalidated:     setup.Invalidated,
			FollowThrough5:  setup.FollowThrough5,
			FollowThrough10: setup.FollowThrough10,
			FollowThrough20: setup.FollowThrough20,
			Notes:           setup.Notes,
		})
	}
	return rows
}

func WritePriceActionPhase3StudyCSV(path string, rows []PriceActionPhase3StudyRow) error {
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
		"timestamp",
		"strategy",
		"direction",
		"level",
		"level_type",
		"confirmation",
		"invalidated",
		"follow_through_5",
		"follow_through_10",
		"follow_through_20",
		"notes",
	}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.Strategy,
			string(row.Direction),
			floatToString(row.Level),
			row.LevelType,
			strconv.FormatBool(row.Confirmation),
			strconv.FormatBool(row.Invalidated),
			floatToString(row.FollowThrough5),
			floatToString(row.FollowThrough10),
			floatToString(row.FollowThrough20),
			row.Notes,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
