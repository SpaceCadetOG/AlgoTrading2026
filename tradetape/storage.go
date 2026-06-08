package tradetape

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type TradeTapeRow struct {
	Print TradeTapePrint
	Valid bool
	Error string
}

func RowFromPrint(print TradeTapePrint) TradeTapeRow {
	print = NormalizePrintSymbols(print)
	err := ValidatePrint(print)
	if err != nil {
		return TradeTapeRow{Print: print, Valid: false, Error: err.Error()}
	}
	return TradeTapeRow{Print: print, Valid: true}
}

func ErrorRow(venue string, symbol string, err error) TradeTapeRow {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return TradeTapeRow{
		Print: TradeTapePrint{
			Venue:           venue,
			Symbol:          symbol,
			VenueSymbol:     symbol,
			CanonicalSymbol: NormalizePrintSymbols(TradeTapePrint{Venue: venue, Symbol: symbol}).CanonicalSymbol,
			AggressorSide:   UnknownAggressorSide,
		},
		Valid: false,
		Error: message,
	}
}

func AppendRows(path string, rows []TradeTapeRow) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	needsHeader := true
	if info, err := os.Stat(path); err == nil && info.Size() > 0 {
		needsHeader = false
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if needsHeader {
		if err := writer.Write(tradeTapeHeader()); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := writer.Write(rowToCSV(row)); err != nil {
			return err
		}
	}
	return writer.Error()
}

func WriteRows(path string, rows []TradeTapeRow) error {
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
	if err := writer.Write(tradeTapeHeader()); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write(rowToCSV(row)); err != nil {
			return err
		}
	}
	return writer.Error()
}

func ReadRows(path string) ([]TradeTapeRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) <= 1 {
		return nil, nil
	}
	rows := make([]TradeTapeRow, 0, len(records)-1)
	for _, record := range records[1:] {
		row, err := csvToRow(record)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func tradeTapeHeader() []string {
	return []string{
		"venue", "venue_symbol", "canonical_symbol", "market_id", "timestamp",
		"price", "size",
		"side", "aggressor_side",
		"trade_id",
		"valid", "error",
	}
}

func rowToCSV(row TradeTapeRow) []string {
	row.Print = NormalizePrintSymbols(row.Print)
	return []string{
		row.Print.Venue,
		row.Print.VenueSymbol,
		row.Print.CanonicalSymbol,
		row.Print.MarketID,
		strconv.FormatInt(row.Print.Timestamp, 10),
		floatString(row.Print.Price),
		floatString(row.Print.Size),
		row.Print.Side,
		row.Print.AggressorSide,
		row.Print.TradeID,
		strconv.FormatBool(row.Valid),
		row.Error,
	}
}

func csvToRow(record []string) (TradeTapeRow, error) {
	if len(record) < 11 {
		return TradeTapeRow{}, fmt.Errorf("trade tape row has %d columns, want 11", len(record))
	}
	if len(record) >= 12 {
		timestamp, _ := strconv.ParseInt(record[4], 10, 64)
		price, _ := strconv.ParseFloat(record[5], 64)
		size, _ := strconv.ParseFloat(record[6], 64)
		valid, _ := strconv.ParseBool(record[10])
		row := TradeTapeRow{
			Print: TradeTapePrint{
				Venue:           record[0],
				Symbol:          record[1],
				VenueSymbol:     record[1],
				CanonicalSymbol: record[2],
				MarketID:        record[3],
				Timestamp:       timestamp,
				Price:           price,
				Size:            size,
				Side:            record[7],
				AggressorSide:   record[8],
				TradeID:         record[9],
			},
			Valid: valid,
			Error: record[11],
		}
		row.Print = NormalizePrintSymbols(row.Print)
		return row, nil
	}
	timestamp, _ := strconv.ParseInt(record[0], 10, 64)
	price, _ := strconv.ParseFloat(record[4], 64)
	size, _ := strconv.ParseFloat(record[5], 64)
	valid, _ := strconv.ParseBool(record[9])
	row := TradeTapeRow{
		Print: TradeTapePrint{
			Timestamp:     timestamp,
			Venue:         record[1],
			Symbol:        record[2],
			VenueSymbol:   record[2],
			MarketID:      record[3],
			Price:         price,
			Size:          size,
			Side:          record[6],
			AggressorSide: record[7],
			TradeID:       record[8],
		},
		Valid: valid,
		Error: record[10],
	}
	row.Print = NormalizePrintSymbols(row.Print)
	return row, nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func floatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
