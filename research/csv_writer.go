package research

import (
	"encoding/csv"
	"os"
	"strconv"

	"AlgoTrading2026/signals"
)

type CSVSignalWriter struct {
	file   *os.File
	writer *csv.Writer
}

func NewCSVSignalWriter(path string) (*CSVSignalWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)

	err = writer.Write([]string{
		"venue",
		"symbol",
		"interval",
		"direction",
		"close",
		"range",
		"body",
		"body_pct",
		"volume",
		"start_time",
		"end_time",
	})
	if err != nil {
		return nil, err
	}

	return &CSVSignalWriter{
		file:   file,
		writer: writer,
	}, nil
}

func (w *CSVSignalWriter) WriteCandleSignal(s signals.CandleSignal) error {
	row := []string{
		s.Venue,
		s.Symbol,
		s.Interval,
		string(s.Direction),
		floatToString(s.Close),
		floatToString(s.Range),
		floatToString(s.Body),
		floatToString(s.BodyPctOfRange),
		floatToString(s.Volume),
		strconv.FormatInt(s.StartTime, 10),
		strconv.FormatInt(s.EndTime, 10),
	}

	if err := w.writer.Write(row); err != nil {
		return err
	}

	w.writer.Flush()

	return w.writer.Error()
}

func (w *CSVSignalWriter) Close() error {
	w.writer.Flush()

	if err := w.writer.Error(); err != nil {
		return err
	}

	return w.file.Close()
}

func floatToString(v float64) string {
	return strconv.FormatFloat(v, 'f', 8, 64)
}
