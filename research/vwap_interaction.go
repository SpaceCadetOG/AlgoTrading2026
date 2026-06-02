package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/vwap"
)

func BuildVWAPInteractions(candles []exchanges.Candle) ([]vwap.Interaction, []vwap.ReactionStudy, []vwap.MagnetStats) {
	session := vwap.SessionVWAP(candles)
	return vwap.Interactions(candles, session), vwap.ReactionStudies(candles, session), vwap.DefaultMagnetStudy(candles, session)
}

func WriteVWAPInteractionCSV(path string, rows []vwap.Interaction) error {
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
		"close",
		"high",
		"low",
		"volume",
		"session_vwap",
		"distance_from_vwap",
		"above_vwap",
		"below_vwap",
		"touched_vwap",
		"crossed_vwap",
		"reaction",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			floatToString(row.Close),
			floatToString(row.High),
			floatToString(row.Low),
			floatToString(row.Volume),
			floatToString(row.SessionVWAP),
			floatToString(row.DistanceFromVWAP),
			strconv.FormatBool(row.AboveVWAP),
			strconv.FormatBool(row.BelowVWAP),
			strconv.FormatBool(row.TouchedVWAP),
			strconv.FormatBool(row.CrossedVWAP),
			string(row.Reaction),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVWAPReactionsCSV(path string, rows []vwap.ReactionStudy) error {
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
		"reaction",
		"close",
		"vwap",
		"distance_from_vwap",
		"follow_through_5",
		"follow_through_10",
		"follow_through_20",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			string(row.Reaction),
			floatToString(row.Close),
			floatToString(row.VWAP),
			floatToString(row.DistanceFromVWAP),
			floatToString(row.FollowThrough5),
			floatToString(row.FollowThrough10),
			floatToString(row.FollowThrough20),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}
