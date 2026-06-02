package l2recorder

import (
	"fmt"
	"strings"
	"time"

	"AlgoTrading2026/exchanges/aster"
	"AlgoTrading2026/exchanges/hyperliquid"
	"AlgoTrading2026/exchanges/lighter"
	"AlgoTrading2026/orderbook"
)

type SnapshotFetcher interface {
	FetchOrderBook(venue string, symbol string) (orderbook.OrderBookSnapshot, error)
}

type SnapshotStorage interface {
	AppendRows(rows []SnapshotRow) error
}

type Recorder struct {
	Config  RecorderConfig
	Fetcher SnapshotFetcher
	Storage SnapshotStorage
	Now     func() time.Time
	Sleep   func(time.Duration)
}

type RunSummary struct {
	Rounds int
	Rows   int
	Errors int
	Output string
}

func NewRecorder(config RecorderConfig, fetcher SnapshotFetcher, storage SnapshotStorage) *Recorder {
	normalized := config.normalized()
	if fetcher == nil {
		fetcher = MainnetFetcher{}
	}
	if storage == nil {
		storage = NewCSVStorage(normalized.OutputPath)
	}
	return &Recorder{
		Config:  normalized,
		Fetcher: fetcher,
		Storage: storage,
		Now:     time.Now,
		Sleep:   time.Sleep,
	}
}

func (r *Recorder) Run() (RunSummary, error) {
	if r.Now == nil {
		r.Now = time.Now
	}
	if r.Sleep == nil {
		r.Sleep = time.Sleep
	}
	summary := RunSummary{Output: r.Config.OutputPath}
	for round := 0; round < r.Config.MaxSnapshots; round++ {
		rows := r.collectOnce()
		if err := r.Storage.AppendRows(rows); err != nil {
			return summary, err
		}
		summary.Rounds++
		summary.Rows += len(rows)
		for _, row := range rows {
			if !row.Valid || row.Error != "" {
				summary.Errors++
			}
		}
		if round < r.Config.MaxSnapshots-1 && r.Config.IntervalSeconds > 0 {
			r.Sleep(time.Duration(r.Config.IntervalSeconds) * time.Second)
		}
	}
	return summary, nil
}

func (r *Recorder) collectOnce() []SnapshotRow {
	rows := make([]SnapshotRow, 0, len(r.Config.Venues)*len(r.Config.Symbols))
	for _, symbol := range r.Config.Symbols {
		for _, venue := range r.Config.Venues {
			now := r.Now()
			snapshot, err := r.Fetcher.FetchOrderBook(venue, symbol)
			if err != nil {
				rows = append(rows, ErrorRow(venue, symbol, err, now))
				continue
			}
			if snapshot.Venue == "" {
				snapshot.Venue = venue
			}
			if snapshot.Symbol == "" {
				snapshot.Symbol = symbol
			}
			rows = append(rows, SnapshotToRow(snapshot, nil, now))
		}
	}
	return rows
}

type MainnetFetcher struct{}

func (MainnetFetcher) FetchOrderBook(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
	switch strings.ToLower(strings.TrimSpace(venue)) {
	case "hyperliquid":
		snapshot, err := hyperliquid.GetMainnetL2OrderBookSnapshot(normalizeHyperliquidSymbol(symbol))
		if err != nil {
			return orderbook.OrderBookSnapshot{}, err
		}
		return *snapshot, nil
	case "aster":
		return aster.GetMainnetOrderBook(normalizeAsterSymbol(symbol))
	case "lighter":
		return lighter.GetMainnetOrderBook(normalizeLighterSymbol(symbol))
	default:
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("unsupported venue %q", venue)
	}
}

func normalizeHyperliquidSymbol(symbol string) string {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.HasSuffix(normalized, "USDT") {
		normalized = strings.TrimSuffix(normalized, "USDT")
	}
	return normalized
}

func normalizeAsterSymbol(symbol string) string {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if normalized == "BTC" {
		return "BTCUSDT"
	}
	if !strings.HasSuffix(normalized, "USDT") {
		return normalized + "USDT"
	}
	return normalized
}

func normalizeLighterSymbol(symbol string) string {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.HasSuffix(normalized, "USDT") {
		normalized = strings.TrimSuffix(normalized, "USDT")
	}
	return normalized
}
