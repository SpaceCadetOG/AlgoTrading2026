package labels

import (
	"os"
	"strings"
	"testing"
)

func TestFutureReturn(t *testing.T) {
	got := FutureReturn([]float64{100, 110, 121}, 2)
	if got[0] != 0.21 {
		t.Fatalf("future return = %f, want 0.21", got[0])
	}
	if got[1] != 0 {
		t.Fatalf("tail future return = %f, want 0", got[1])
	}
}

func TestDirectionLabel(t *testing.T) {
	got := DirectionLabel([]float64{0.003, -0.004, 0.001}, 0.002)
	want := []int{1, -1, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("direction[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestTPSLLabel(t *testing.T) {
	got := TPSLLabel([]float64{100, 101, 99, 98}, TPSLConfig{
		Horizon:       2,
		TakeProfitPct: 0.01,
		StopLossPct:   0.01,
	})
	if got[0] != 1 {
		t.Fatalf("first tp/sl = %d, want 1", got[0])
	}
	if got[1] != -1 {
		t.Fatalf("second tp/sl = %d, want -1", got[1])
	}
}

func TestLabelsCSVWriter(t *testing.T) {
	path := t.TempDir() + "/labels.csv"
	rows := []LabelRow{{Time: 1, FutureReturn: 0.1, Direction: 1, TPSL: 1}}

	if err := WriteCSV(path, rows); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.Contains(text, "time,future_return,direction,tpsl") {
		t.Fatalf("missing header: %s", text)
	}
	if !strings.Contains(text, "1,0.10000000,1,1") {
		t.Fatalf("missing row: %s", text)
	}
}
