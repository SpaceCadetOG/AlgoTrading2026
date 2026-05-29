package features

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/series"
)

func TestBuildFeaturesRowCount(t *testing.T) {
	frame := series.Frame{Rows: []series.Row{
		{Time: 1, Close: 100, High: 101, Low: 99, Volume: 1},
		{Time: 2, Close: 101, High: 102, Low: 100, Volume: 2},
	}}

	rows := BuildFeatures("aster", "BTCUSDT", "15m", frame)
	if len(rows) != len(frame.Rows) {
		t.Fatalf("rows = %d, want %d", len(rows), len(frame.Rows))
	}
	if rows[0].Venue != "aster" || rows[0].Symbol != "BTCUSDT" {
		t.Fatalf("unexpected metadata: %+v", rows[0])
	}
}

func TestFeaturesCSVNoFutureLeakageColumns(t *testing.T) {
	path := t.TempDir() + "/features.csv"
	rows := []FeatureRow{{Time: 1, Venue: "aster", Symbol: "BTCUSDT", Interval: "15m", Close: 100}}

	if err := WriteCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	header := strings.Split(string(body), "\n")[0]
	for _, forbidden := range []string{"future_return", "direction", "tpsl"} {
		if strings.Contains(header, forbidden) {
			t.Fatalf("feature header leaks label column %s: %s", forbidden, header)
		}
	}
}
