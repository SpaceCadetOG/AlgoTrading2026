package series

type SignalResult struct {
	Times     []int64
	Close     []float64
	Diff      []float64
	Signal    []float64
	Positions []float64
}

func SignalPositive(diff []float64) []float64 {
	signal := make([]float64, len(diff))
	for i, value := range diff {
		if value > 0 {
			signal[i] = 1
		}
	}
	return signal
}

func BuyLowSellHighSignals(frame Frame) SignalResult {
	close := frame.CloseColumn()
	diff := Diff(close)
	signal := SignalPositive(diff)

	return SignalResult{
		Times:     frame.TimeColumn(),
		Close:     close,
		Diff:      diff,
		Signal:    signal,
		Positions: PositionChanges(signal),
	}
}
