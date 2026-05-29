package series

func (f Frame) CloseColumn() []float64 {
	values := make([]float64, len(f.Rows))
	for i, row := range f.Rows {
		values[i] = row.Close
	}
	return values
}

func (f Frame) HighColumn() []float64 {
	values := make([]float64, len(f.Rows))
	for i, row := range f.Rows {
		values[i] = row.High
	}
	return values
}

func (f Frame) LowColumn() []float64 {
	values := make([]float64, len(f.Rows))
	for i, row := range f.Rows {
		values[i] = row.Low
	}
	return values
}

func (f Frame) TimeColumn() []int64 {
	values := make([]int64, len(f.Rows))
	for i, row := range f.Rows {
		values[i] = row.Time
	}
	return values
}
