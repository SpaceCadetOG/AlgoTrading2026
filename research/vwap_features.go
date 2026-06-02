package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/vwap"
)

type VWAPFeatureRow struct {
	Timestamp          int64
	Close              float64
	Volume             float64
	SessionVWAP        float64
	AnchoredVWAP       float64
	DistanceFromVWAP   float64
	DistanceFromAnchor float64
	VWAPSlope          float64
	VWAPRegime         vwap.Regime
}

func BuildVWAPFeatureRows(candles []exchanges.Candle, anchorTime int64, slopePeriod int) []VWAPFeatureRow {
	session := vwap.SessionVWAP(candles)
	anchored := vwap.AnchoredVWAP(candles, anchorTime)
	close := make([]float64, len(candles))
	for i, candle := range candles {
		close[i] = candle.CloseFloat()
	}
	distanceSession := vwap.Distance(close, session)
	distanceAnchor := vwap.Distance(close, anchored)
	slopes := vwap.Slope(session, slopePeriod)
	regimes := vwap.Regimes(close, session, slopes)

	rows := make([]VWAPFeatureRow, 0, len(candles))
	for i, candle := range candles {
		rows = append(rows, VWAPFeatureRow{
			Timestamp:          candle.StartTime,
			Close:              close[i],
			Volume:             candle.VolumeFloat(),
			SessionVWAP:        valueAt(session, i),
			AnchoredVWAP:       valueAt(anchored, i),
			DistanceFromVWAP:   valueAt(distanceSession, i),
			DistanceFromAnchor: valueAt(distanceAnchor, i),
			VWAPSlope:          valueAt(slopes, i),
			VWAPRegime:         regimeAt(regimes, i),
		})
	}
	return rows
}

func WriteVWAPFeaturesCSV(path string, rows []VWAPFeatureRow) error {
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
		"volume",
		"session_vwap",
		"anchored_vwap",
		"distance_from_vwap",
		"distance_from_anchor",
		"vwap_slope",
		"vwap_regime",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			floatToString(row.Close),
			floatToString(row.Volume),
			floatToString(row.SessionVWAP),
			floatToString(row.AnchoredVWAP),
			floatToString(row.DistanceFromVWAP),
			floatToString(row.DistanceFromAnchor),
			floatToString(row.VWAPSlope),
			string(row.VWAPRegime),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func valueAt(values []float64, i int) float64 {
	if i < 0 || i >= len(values) {
		return 0
	}
	return values[i]
}

func regimeAt(values []vwap.Regime, i int) vwap.Regime {
	if i < 0 || i >= len(values) {
		return vwap.RegimeNeutral
	}
	return values[i]
}
