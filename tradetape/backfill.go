package tradetape

import (
	"os"
	"strconv"
	"strings"
	"time"

	appconfig "AlgoTrading2026/config"
)

type BackfillProvider interface {
	Venue() string
	BackfillTrades(symbol string, startMS int64, endMS int64) ([]TradeTapePrint, error)
}

type BackfillConfig struct {
	RunAsterAggTradesBackfill bool
	Symbol                    string
	StartMS                   int64
	EndMS                     int64
	OutputPath                string
	DedupedOutputPath         string
}

func DefaultBackfillConfig() BackfillConfig {
	output := envString("ASTER_BACKFILL_OUTPUT", appconfig.DataPath("trade_tape", "backfill", "aster_btcusdt.csv"))
	end := envInt64("ASTER_BACKFILL_END_MS", time.Now().UnixMilli())
	start := envInt64("ASTER_BACKFILL_START_MS", end-int64(24*time.Hour/time.Millisecond))
	return BackfillConfig{
		RunAsterAggTradesBackfill: strings.EqualFold(os.Getenv("RUN_ASTER_AGGTRADES_BACKFILL"), "true"),
		Symbol:                    envString("ASTER_BACKFILL_SYMBOL", "BTCUSDT"),
		StartMS:                   start,
		EndMS:                     end,
		OutputPath:                output,
		DedupedOutputPath:         dedupedPath(output),
	}
}

func RunBackfill(provider BackfillProvider, cfg BackfillConfig) ([]TradeTapeRow, []TradeTapeRow, DedupeSummary, error) {
	prints, err := provider.BackfillTrades(cfg.Symbol, cfg.StartMS, cfg.EndMS)
	if err != nil {
		return nil, nil, DedupeSummary{}, err
	}
	rows := make([]TradeTapeRow, 0, len(prints))
	for _, print := range prints {
		rows = append(rows, RowFromPrint(print))
	}
	if err := WriteRows(cfg.OutputPath, rows); err != nil {
		return nil, nil, DedupeSummary{}, err
	}
	deduped, summary := DeduplicatePrints(prints)
	dedupedRows := make([]TradeTapeRow, 0, len(deduped))
	for _, print := range deduped {
		dedupedRows = append(dedupedRows, RowFromPrint(print))
	}
	if err := WriteRows(cfg.DedupedOutputPath, dedupedRows); err != nil {
		return nil, nil, DedupeSummary{}, err
	}
	return rows, dedupedRows, summary, nil
}

func envInt64(name string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func dedupedPath(path string) string {
	if strings.HasSuffix(path, ".csv") {
		return strings.TrimSuffix(path, ".csv") + "_deduped.csv"
	}
	return path + "_deduped.csv"
}
