# Chapter 10C Market Data Quality and Historical/Live Parity Audit

- venue: aster
- symbol: BTCUSDT
- interval: 15m

## Market Data Quality

- total_candles_checked: 2880
- quality_risk: low
- missing_candles: 0
- duplicate_timestamps: 0
- out_of_order_timestamps: 0
- invalid_ohlc_values: 0
- invalid_volume: 0
- large_time_gaps: 0
- suspicious_price_jumps: 0
- expected_interval_ms: 900000
- max_observed_gap_ms: 900000
- max_price_jump_pct: 1.5469

## Historical/Live Parity

- parity_risk: medium
- historical_schema_documented: true
- live_schema_documented: true
- timestamp_format_parity: true
- symbol_format_parity: true
- price_precision_parity: true
- volume_precision_parity: true
- candle_interval_parity: true

## Known Venue Differences

- REST candles are closed historical bars while WebSocket candles may include in-progress updates

## Warnings

- known_venue_differences

## Recommendations

- record historical and live candle samples for schema parity tests

## Strengths

- normalized candle data is checked before being used for realism-sensitive research
- timestamp gaps, duplicates, ordering, invalid prices, invalid volume, and price jumps are measured
- historical/live candle schema parity is represented explicitly
- quality and parity risks are separate so data issues are not hidden inside strategy metrics

## Gaps

- historical/live parity is documented from adapter behavior, not verified with live recorded samples
- market data accuracy is not yet cross-checked against an independent reference feed
- venue-specific candle repair rules are not implemented
- data quality gates are not yet applied to strategy promotion decisions

## Conclusion

- Chapter 10C audits market data quality and historical/live parity only.
- No live or paper trading is enabled.
- No real exchange calls are made.
- Next phase is continued Chapter 10 market realism research.

## Next Phase

Continue Chapter 10 market realism research.
