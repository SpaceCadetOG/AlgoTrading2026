package series

import "testing"

func TestSignalPositive(t *testing.T) {
	got := SignalPositive([]float64{0, -1, 2, 1, -2})
	want := []float64{0, 0, 1, 1, 0}

	assertFloatSlices(t, got, want)
}

func TestBookStyleSignalFlow(t *testing.T) {
	frame := Frame{Rows: []Row{
		{Time: 1, Close: 10},
		{Time: 2, Close: 9},
		{Time: 3, Close: 11},
		{Time: 4, Close: 12},
		{Time: 5, Close: 10},
	}}

	got := BuyLowSellHighSignals(frame)

	assertFloatSlices(t, got.Close, []float64{10, 9, 11, 12, 10})
	assertFloatSlices(t, got.Diff, []float64{0, -1, 2, 1, -2})
	assertFloatSlices(t, got.Signal, []float64{0, 0, 1, 1, 0})
	assertFloatSlices(t, got.Positions, []float64{0, 0, 1, 0, -1})
}
