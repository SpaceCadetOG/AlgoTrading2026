package research

import (
	"os"
	"strings"
	"testing"

	"AlgoTrading2026/volatility"
)

func TestWriteChapter5VolatilityCSV(t *testing.T) {
	path := t.TempDir() + "/chapter5_volatility.csv"
	rows := []Chapter5VolatilityRow{{
		Time:      1,
		Close:     100,
		High:      101,
		Low:       99,
		TrueRange: 2,
		Regime:    volatility.RegimeNormal,
	}}

	if err := WriteChapter5VolatilityCSV(path, rows); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "time,close,high,low,true_range,atr,log_return,realized_vol,rolling_stddev,regime") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "1,100.00000000,101.00000000,99.00000000,2.00000000") {
		t.Fatalf("missing row: %s", text)
	}
}
