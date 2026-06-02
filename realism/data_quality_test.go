package realism

import (
	"testing"

	"AlgoTrading2026/exchanges"
)

func TestCheckMarketDataQualityLowRisk(t *testing.T) {
	candles := []exchanges.Candle{
		testCandle(0, "100", "101", "99", "100", "10"),
		testCandle(900000, "100", "102", "99", "101", "11"),
		testCandle(1800000, "101", "103", "100", "102", "12"),
	}

	check := CheckMarketDataQuality(candles, 900000)
	if check.QualityRisk != SeverityLow {
		t.Fatalf("expected low risk, got %s", check.QualityRisk)
	}
	if check.MissingCandles != 0 || check.DuplicateTimestamps != 0 || check.OutOfOrderTimestamps != 0 {
		t.Fatalf("unexpected data issues: %+v", check)
	}
}

func TestCheckMarketDataQualityFindsTimestampIssues(t *testing.T) {
	candles := []exchanges.Candle{
		testCandle(0, "100", "101", "99", "100", "10"),
		testCandle(2700000, "100", "102", "99", "101", "10"),
		testCandle(2700000, "101", "102", "100", "101", "10"),
		testCandle(1800000, "101", "103", "100", "102", "10"),
	}

	check := CheckMarketDataQuality(candles, 900000)
	if check.QualityRisk != SeverityHigh {
		t.Fatalf("expected high risk for out-of-order timestamps, got %s", check.QualityRisk)
	}
	if check.MissingCandles == 0 || check.DuplicateTimestamps == 0 || check.OutOfOrderTimestamps == 0 {
		t.Fatalf("expected missing, duplicate, and out-of-order issues: %+v", check)
	}
}

func TestCheckMarketDataQualityFindsInvalidValuesAndJumps(t *testing.T) {
	candles := []exchanges.Candle{
		testCandle(0, "100", "101", "99", "100", "10"),
		testCandle(900000, "100", "101", "99", "150", "-1"),
		testCandle(1800000, "0", "101", "99", "100", "10"),
	}

	check := CheckMarketDataQuality(candles, 900000)
	if check.QualityRisk != SeverityHigh {
		t.Fatalf("expected high risk, got %s", check.QualityRisk)
	}
	if check.ZeroNegativeOHLC == 0 || check.InvalidVolume == 0 || check.SuspiciousPriceJumps == 0 {
		t.Fatalf("expected invalid values and jump detection: %+v", check)
	}
}

func TestHistoricalLiveParityRisk(t *testing.T) {
	low := NewHistoricalLiveParityCheck("test", "BTC", "15m", nil)
	if low.ParityRisk != SeverityLow {
		t.Fatalf("expected low parity risk, got %s", low.ParityRisk)
	}

	medium := DefaultHistoricalLiveParityCheck("aster", "BTCUSDT", "15m")
	if medium.ParityRisk != SeverityMedium {
		t.Fatalf("expected medium parity risk for known differences, got %s", medium.ParityRisk)
	}

	high := NewHistoricalLiveParityCheck("test", "BTC", "15m", nil)
	high.LiveSchemaDocumented = false
	high.TimestampFormatParity = false
	high.Warnings = parityWarnings(high)
	high.ParityRisk = classifyParityRisk(high)
	if high.ParityRisk != SeverityHigh {
		t.Fatalf("expected high parity risk, got %s", high.ParityRisk)
	}
}

func TestBuildDataQualityReport(t *testing.T) {
	quality := CheckMarketDataQuality([]exchanges.Candle{
		testCandle(0, "100", "101", "99", "100", "10"),
		testCandle(900000, "100", "101", "99", "101", "10"),
	}, 900000)
	parity := DefaultHistoricalLiveParityCheck("aster", "BTCUSDT", "15m")

	report := BuildDataQualityReport(quality, parity)
	if report.TotalCandlesChecked != 2 {
		t.Fatalf("expected two checked candles, got %d", report.TotalCandlesChecked)
	}
	if report.ParityRisk != SeverityMedium {
		t.Fatalf("expected medium parity risk, got %s", report.ParityRisk)
	}
	if len(report.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
}

func testCandle(start int64, open string, high string, low string, closePrice string, volume string) exchanges.Candle {
	return exchanges.Candle{
		Venue:     "test",
		Symbol:    "BTC",
		Interval:  "15m",
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePrice,
		Volume:    volume,
		StartTime: start,
		EndTime:   start + 900000,
		Closed:    true,
	}
}
