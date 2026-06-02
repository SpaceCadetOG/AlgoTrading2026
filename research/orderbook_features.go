package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/orderbook"
	"AlgoTrading2026/vwap"
)

type OrderBookFeatureRow struct {
	Timestamp     int64
	Venue         string
	Symbol        string
	BestBid       float64
	BestAsk       float64
	Spread        float64
	SpreadPct     float64
	Mid           float64
	BidDepth1Pct  float64
	AskDepth1Pct  float64
	Imbalance1Pct float64
	Valid         bool
}

type VWAPL2InteractionRow struct {
	Timestamp         int64
	Venue             string
	Symbol            string
	Mid               float64
	SessionVWAP       float64
	DistanceFromVWAP  float64
	SpreadPct         float64
	Imbalance1Pct     float64
	LiquidityNearVWAP float64
}

func BuildOrderBookFeatureRow(snapshot orderbook.OrderBookSnapshot) OrderBookFeatureRow {
	return OrderBookFeatureRow{
		Timestamp:     snapshot.Time,
		Venue:         snapshot.Venue,
		Symbol:        snapshot.Symbol,
		BestBid:       orderbook.BestBid(snapshot).PriceFloat(),
		BestAsk:       orderbook.BestAsk(snapshot).PriceFloat(),
		Spread:        orderbook.Spread(snapshot),
		SpreadPct:     orderbook.SpreadPct(snapshot),
		Mid:           orderbook.Mid(snapshot),
		BidDepth1Pct:  orderbook.BidDepthWithinPct(snapshot, 1),
		AskDepth1Pct:  orderbook.AskDepthWithinPct(snapshot, 1),
		Imbalance1Pct: orderbook.Imbalance(snapshot, 1),
		Valid:         orderbook.ValidateSnapshot(snapshot) == nil,
	}
}

func BuildVWAPL2InteractionRow(snapshot orderbook.OrderBookSnapshot, candles []exchanges.Candle) VWAPL2InteractionRow {
	session := vwap.SessionVWAP(candles)
	sessionVWAP := latestNonZero(session)
	bidLiquidity, askLiquidity := orderbook.LiquidityNearPrice(snapshot, sessionVWAP, 1)
	mid := orderbook.Mid(snapshot)

	return VWAPL2InteractionRow{
		Timestamp:         snapshot.Time,
		Venue:             snapshot.Venue,
		Symbol:            snapshot.Symbol,
		Mid:               mid,
		SessionVWAP:       sessionVWAP,
		DistanceFromVWAP:  mid - sessionVWAP,
		SpreadPct:         orderbook.SpreadPct(snapshot),
		Imbalance1Pct:     orderbook.Imbalance(snapshot, 1),
		LiquidityNearVWAP: bidLiquidity + askLiquidity,
	}
}

func WriteOrderBookFeaturesCSV(path string, rows []OrderBookFeatureRow) error {
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
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.Venue,
			row.Symbol,
			floatToString(row.BestBid),
			floatToString(row.BestAsk),
			floatToString(row.Spread),
			floatToString(row.SpreadPct),
			floatToString(row.Mid),
			floatToString(row.BidDepth1Pct),
			floatToString(row.AskDepth1Pct),
			floatToString(row.Imbalance1Pct),
			strconv.FormatBool(row.Valid),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteVWAPL2InteractionCSV(path string, rows []VWAPL2InteractionRow) error {
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
		"venue",
		"symbol",
		"mid",
		"session_vwap",
		"distance_from_vwap",
		"spread_pct",
		"imbalance_1pct",
		"liquidity_near_vwap",
	}); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Timestamp, 10),
			row.Venue,
			row.Symbol,
			floatToString(row.Mid),
			floatToString(row.SessionVWAP),
			floatToString(row.DistanceFromVWAP),
			floatToString(row.SpreadPct),
			floatToString(row.Imbalance1Pct),
			floatToString(row.LiquidityNearVWAP),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func latestNonZero(values []float64) float64 {
	for i := len(values) - 1; i >= 0; i-- {
		if values[i] != 0 {
			return values[i]
		}
	}
	return 0
}
