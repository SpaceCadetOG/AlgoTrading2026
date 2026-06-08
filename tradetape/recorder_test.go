package tradetape

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestRecorderRoundContinuesOnPartialVenueFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trades.csv")
	recorder := NewRecorder(RecorderConfig{OutputPath: path, MaxRounds: 1}, []VenueFetcher{
		{
			Venue: "hyperliquid", Symbol: "BTC",
			Fetch: func() ([]TradeTapePrint, error) {
				return []TradeTapePrint{testPrint("hyperliquid", "BTC", 1, 100, 1, "B")}, nil
			},
		},
		{
			Venue: "aster", Symbol: "BTCUSDT",
			Fetch: func() ([]TradeTapePrint, error) {
				return nil, errors.New("venue unavailable")
			},
		},
	})
	summary, err := recorder.RecordRound()
	if err != nil {
		t.Fatalf("record round: %v", err)
	}
	if summary.Rows != 2 || summary.Errors != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	rows, err := ReadRows(path)
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("len=%d want 2: %+v", len(rows), rows)
	}
	foundError := false
	for _, row := range rows {
		if !row.Valid && row.Error != "" {
			foundError = true
		}
	}
	if !foundError {
		t.Fatalf("expected one error row: %+v", rows)
	}
}

func TestRecorderRunStopsAtMaxRounds(t *testing.T) {
	path := filepath.Join(t.TempDir(), "trades.csv")
	recorder := NewRecorder(RecorderConfig{OutputPath: path, MaxRounds: 2, IntervalSeconds: 1}, []VenueFetcher{
		{
			Venue: "lighter", Symbol: "BTC",
			Fetch: func() ([]TradeTapePrint, error) {
				return []TradeTapePrint{testPrint("lighter", "BTC", 1, 100, 1, "")}, nil
			},
		},
	})
	summary, err := recorder.Run()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if summary.Rounds != 2 || summary.Rows != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
