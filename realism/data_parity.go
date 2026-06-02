package realism

type HistoricalLiveParityCheck struct {
	Venue                      string              `json:"venue"`
	Symbol                     string              `json:"symbol"`
	Interval                   string              `json:"interval"`
	HistoricalSchemaDocumented bool                `json:"historicalSchemaDocumented"`
	LiveSchemaDocumented       bool                `json:"liveSchemaDocumented"`
	TimestampFormatParity      bool                `json:"timestampFormatParity"`
	SymbolFormatParity         bool                `json:"symbolFormatParity"`
	PricePrecisionParity       bool                `json:"pricePrecisionParity"`
	VolumePrecisionParity      bool                `json:"volumePrecisionParity"`
	CandleIntervalParity       bool                `json:"candleIntervalParity"`
	KnownVenueDifferences      []string            `json:"knownVenueDifferences"`
	ParityRisk                 DislocationSeverity `json:"parityRisk"`
	Warnings                   []string            `json:"warnings"`
}

type DataQualityReport struct {
	TotalCandlesChecked int                       `json:"totalCandlesChecked"`
	IssueCounts         map[string]int            `json:"issueCounts"`
	QualityRisk         DislocationSeverity       `json:"qualityRisk"`
	ParityRisk          DislocationSeverity       `json:"parityRisk"`
	Warnings            []string                  `json:"warnings"`
	Recommendations     []string                  `json:"recommendations"`
	Quality             MarketDataQualityCheck    `json:"quality"`
	Parity              HistoricalLiveParityCheck `json:"parity"`
}

func NewHistoricalLiveParityCheck(venue string, symbol string, interval string, differences []string) HistoricalLiveParityCheck {
	check := HistoricalLiveParityCheck{
		Venue:                      venue,
		Symbol:                     symbol,
		Interval:                   interval,
		HistoricalSchemaDocumented: true,
		LiveSchemaDocumented:       true,
		TimestampFormatParity:      true,
		SymbolFormatParity:         true,
		PricePrecisionParity:       true,
		VolumePrecisionParity:      true,
		CandleIntervalParity:       true,
		KnownVenueDifferences:      append([]string(nil), differences...),
	}
	check.Warnings = parityWarnings(check)
	check.ParityRisk = classifyParityRisk(check)
	return check
}

func DefaultHistoricalLiveParityCheck(venue string, symbol string, interval string) HistoricalLiveParityCheck {
	return NewHistoricalLiveParityCheck(venue, symbol, interval, defaultVenueDifferences(venue))
}

func BuildDataQualityReport(quality MarketDataQualityCheck, parity HistoricalLiveParityCheck) DataQualityReport {
	issues := map[string]int{
		"missing_candles":         quality.MissingCandles,
		"duplicate_timestamps":    quality.DuplicateTimestamps,
		"out_of_order_timestamps": quality.OutOfOrderTimestamps,
		"invalid_ohlc_values":     quality.ZeroNegativeOHLC,
		"invalid_volume":          quality.InvalidVolume,
		"large_time_gaps":         quality.LargeTimeGaps,
		"suspicious_price_jumps":  quality.SuspiciousPriceJumps,
	}
	warnings := append([]string{}, quality.Warnings...)
	warnings = append(warnings, parity.Warnings...)
	return DataQualityReport{
		TotalCandlesChecked: quality.TotalCandles,
		IssueCounts:         issues,
		QualityRisk:         quality.QualityRisk,
		ParityRisk:          parity.ParityRisk,
		Warnings:            warnings,
		Recommendations:     dataQualityRecommendations(quality, parity),
		Quality:             quality,
		Parity:              parity,
	}
}

func classifyParityRisk(check HistoricalLiveParityCheck) DislocationSeverity {
	missing := 0
	fields := []bool{
		check.HistoricalSchemaDocumented,
		check.LiveSchemaDocumented,
		check.TimestampFormatParity,
		check.SymbolFormatParity,
		check.PricePrecisionParity,
		check.VolumePrecisionParity,
		check.CandleIntervalParity,
	}
	for _, ok := range fields {
		if !ok {
			missing++
		}
	}
	switch {
	case missing >= 2:
		return SeverityHigh
	case missing == 1:
		return SeverityMedium
	case len(check.KnownVenueDifferences) > 0:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

func parityWarnings(check HistoricalLiveParityCheck) []string {
	var warnings []string
	if !check.HistoricalSchemaDocumented {
		warnings = append(warnings, "historical_schema_not_documented")
	}
	if !check.LiveSchemaDocumented {
		warnings = append(warnings, "live_schema_not_documented")
	}
	if !check.TimestampFormatParity {
		warnings = append(warnings, "timestamp_format_mismatch")
	}
	if !check.SymbolFormatParity {
		warnings = append(warnings, "symbol_format_mismatch")
	}
	if !check.PricePrecisionParity {
		warnings = append(warnings, "price_precision_mismatch")
	}
	if !check.VolumePrecisionParity {
		warnings = append(warnings, "volume_precision_mismatch")
	}
	if !check.CandleIntervalParity {
		warnings = append(warnings, "candle_interval_mismatch")
	}
	if len(check.KnownVenueDifferences) > 0 {
		warnings = append(warnings, "known_venue_differences")
	}
	return warnings
}

func dataQualityRecommendations(quality MarketDataQualityCheck, parity HistoricalLiveParityCheck) []string {
	var recommendations []string
	if quality.QualityRisk != SeverityLow {
		recommendations = append(recommendations, "block strategy promotion when market data quality risk is elevated")
	}
	if quality.MissingCandles > 0 || quality.LargeTimeGaps > 0 {
		recommendations = append(recommendations, "add gap repair or explicit gap rejection before forward testing")
	}
	if quality.DuplicateTimestamps > 0 || quality.OutOfOrderTimestamps > 0 {
		recommendations = append(recommendations, "deduplicate and sort candles before research runs")
	}
	if quality.SuspiciousPriceJumps > 0 {
		recommendations = append(recommendations, "cross-check suspicious jumps against another venue before trusting fills")
	}
	if parity.ParityRisk != SeverityLow {
		recommendations = append(recommendations, "record historical and live candle samples for schema parity tests")
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "continue monitoring candle quality before any forward-testing shell")
	}
	return recommendations
}

func defaultVenueDifferences(venue string) []string {
	switch venue {
	case "aster":
		return []string{"REST candles are closed historical bars while WebSocket candles may include in-progress updates"}
	case "hyperliquid":
		return []string{"REST and WebSocket candle symbols use venue-native coin names such as BTC instead of BTCUSDT"}
	case "lighter":
		return []string{"REST candle lookup may use market index mapping while normalized output uses symbol names"}
	default:
		return []string{"venue-specific historical/live parity requires recorded sample comparison"}
	}
}
