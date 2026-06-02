package l2recorder

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSnapshotRowsCSV(t *testing.T) {
	path := writeTestSnapshotCSV(t)
	rows, err := ReadSnapshotRowsCSV(path)
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	if len(rows) != 7 {
		t.Fatalf("rows=%d want 7", len(rows))
	}
	if rows[0].Venue != "hyperliquid" || !rows[0].Valid {
		t.Fatalf("bad first row: %+v", rows[0])
	}
	if rows[6].Error == "" || rows[6].Valid {
		t.Fatalf("expected error row: %+v", rows[6])
	}
}

func TestAnalyzeSnapshotRows(t *testing.T) {
	rows, err := ReadSnapshotRowsCSV(writeTestSnapshotCSV(t))
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	analysis := AnalyzeSnapshotRows(rows)

	if analysis.DatasetQuality.TotalRows != 7 {
		t.Fatalf("total rows=%d want 7", analysis.DatasetQuality.TotalRows)
	}
	if analysis.DatasetQuality.ValidRows != 6 || analysis.DatasetQuality.ErrorRows != 1 {
		t.Fatalf("bad quality: %+v", analysis.DatasetQuality)
	}
	if len(analysis.VenueSummaries) != 3 {
		t.Fatalf("venue summaries=%d want 3", len(analysis.VenueSummaries))
	}

	hl := findVenueSummary(t, analysis.VenueSummaries, "hyperliquid")
	if hl.SnapshotCount != 2 || hl.ValidSnapshotCount != 2 {
		t.Fatalf("bad hl counts: %+v", hl)
	}
	if !closeEnough(hl.AverageSpreadPct, 0.15) || !closeEnough(hl.MinSpreadPct, 0.1) || !closeEnough(hl.MaxSpreadPct, 0.2) {
		t.Fatalf("bad spread stats: %+v", hl)
	}
	if !closeEnough(hl.AverageImbalance1Pct, 0.15) || !closeEnough(hl.MinImbalance1Pct, 0.1) || !closeEnough(hl.MaxImbalance1Pct, 0.2) {
		t.Fatalf("bad imbalance stats: %+v", hl)
	}
	if hl.AverageBidDepth1Pct != 1100 || hl.AverageAskDepth1Pct != 900 {
		t.Fatalf("bad depth stats: %+v", hl)
	}

	if analysis.CrossVenue.ComparableGroups != 2 {
		t.Fatalf("comparable groups=%d want 2", analysis.CrossVenue.ComparableGroups)
	}
	if analysis.CrossVenue.AverageMidDifference != 10 {
		t.Fatalf("avg mid diff=%.4f want 10", analysis.CrossVenue.AverageMidDifference)
	}
	if !closeEnough(analysis.CrossVenue.AverageSpreadDifference, 0.2) {
		t.Fatalf("avg spread diff=%.4f want 0.2", analysis.CrossVenue.AverageSpreadDifference)
	}
}

func TestLiquidityRankings(t *testing.T) {
	rows, err := ReadSnapshotRowsCSV(writeTestSnapshotCSV(t))
	if err != nil {
		t.Fatalf("read rows: %v", err)
	}
	analysis := AnalyzeSnapshotRows(rows)
	if analysis.LiquidityRanking.AverageBidDepth[0].Venue != "lighter" {
		t.Fatalf("bid depth ranking: %+v", analysis.LiquidityRanking.AverageBidDepth)
	}
	if analysis.LiquidityRanking.AverageAskDepth[0].Venue != "lighter" {
		t.Fatalf("ask depth ranking: %+v", analysis.LiquidityRanking.AverageAskDepth)
	}
	if analysis.LiquidityRanking.CombinedDepth[0].Venue != "lighter" {
		t.Fatalf("combined depth ranking: %+v", analysis.LiquidityRanking.CombinedDepth)
	}
	if analysis.LiquidityRanking.TightestSpreads[0].Venue != "aster" {
		t.Fatalf("tight spread ranking: %+v", analysis.LiquidityRanking.TightestSpreads)
	}
}

func findVenueSummary(t *testing.T, rows []VenueSummary, venue string) VenueSummary {
	t.Helper()
	for _, row := range rows {
		if row.Venue == venue {
			return row
		}
	}
	t.Fatalf("missing venue %s in %+v", venue, rows)
	return VenueSummary{}
}

func writeTestSnapshotCSV(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "l2.csv")
	body := `timestamp,venue,symbol,best_bid,best_ask,spread,spread_pct,mid,bid_depth_1pct,ask_depth_1pct,imbalance_1pct,valid,error
1,hyperliquid,BTC,100,101,1,0.1,100,1000,800,0.1,true,
2,aster,BTCUSDT,100,100.5,0.5,0.05,100,900,900,0,true,
3,lighter,BTC,90,91,1,0.2,90,2000,2500,-0.2,true,
4,hyperliquid,BTC,110,111,1,0.2,110,1200,1000,0.2,true,
5,aster,BTCUSDT,109,109.5,0.5,0.05,109,950,900,0.05,true,
6,lighter,BTC,100,101,1,0.3,100,2200,2400,-0.1,true,
7,aster,BTCUSDT,0,0,0,0,0,0,0,0,false,mock failure
`
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write test csv: %v", err)
	}
	return path
}

func closeEnough(a float64, b float64) bool {
	const epsilon = 1e-9
	if a > b {
		return a-b < epsilon
	}
	return b-a < epsilon
}
