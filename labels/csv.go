package labels

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
)

func BuildRows(times []int64, futureReturns []float64, directions []int, tpsl []int) []LabelRow {
	n := min(len(times), min(len(futureReturns), min(len(directions), len(tpsl))))
	rows := make([]LabelRow, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, LabelRow{
			Time:         times[i],
			FutureReturn: futureReturns[i],
			Direction:    directions[i],
			TPSL:         tpsl[i],
		})
	}
	return rows
}

func WriteCSV(path string, rows []LabelRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"time", "future_return", "direction", "tpsl"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			strconv.FormatInt(row.Time, 10),
			fmtFloat(row.FutureReturn),
			strconv.Itoa(row.Direction),
			strconv.Itoa(row.TPSL),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func fmtFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 8, 64)
}
