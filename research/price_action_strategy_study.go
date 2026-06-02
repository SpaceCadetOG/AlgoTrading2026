package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/priceaction"
)

type PriceActionStrategyStudyRow struct {
	Timestamp       int64
	Strategy        string
	Direction       priceaction.StudyDirection
	Level           float64
	EntryZone       float64
	Confirmation    bool
	Invalidated     bool
	FollowThrough5  float64
	FollowThrough10 float64
	FollowThrough20 float64
	Notes           string
}

func BuildPriceActionStrategyStudyRows(candles []exchanges.Candle) []PriceActionStrategyStudyRow {
	setups := make([]priceaction.StrategySetup, 0)
	setups = append(setups, priceaction.SRFlipSetups(candles)...)
	setups = append(setups, priceaction.OpenDriveSetups(candles)...)
	setups = append(setups, priceaction.ABCDSetups(candles)...)
	setups = append(setups, priceaction.SessionOpenSetups(candles)...)
	setups = append(setups, priceaction.DailyOpenSetups(candles)...)

	rows := make([]PriceActionStrategyStudyRow, 0, len(setups))
	for _, setup := range setups {
		rows = append(rows, PriceActionStrategyStudyRow{
			Timestamp:       setup.Timestamp,
			Strategy:        setup.Strategy,
			Direction:       setup.Direction,
			Level:           setup.Level,
			EntryZone:       setup.EntryZone,
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

func WritePriceActionStrategyStudyCSV(path string, rows []PriceActionStrategyStudyRow) error {
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
		"entry_zone",
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
			floatToString(row.EntryZone),
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
