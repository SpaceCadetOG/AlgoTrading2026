# Chapter 10 Master Plan

Audit-only plan based on Chapter 10, `research/chapter10_realism_audit.md`, `research/chapter9_validation_audit.md`, and the current repo structure.

No code, strategy behavior, live trading, paper trading, ML training, real exchange calls, or portfolio optimization is added by this document.

## Current Position

Chapter 10A is complete as a measurement/audit layer:

- simulation dislocation is measured;
- current event-driven PnL is classified as system-validation only;
- current for-loop backtests are classified as useful for strategy research;
- realism gaps are documented;
- the system is not ready for paper/live execution.

The remaining Chapter 10 work should focus on making the simulator and research workflow more realistic before any forward-testing shell is considered.

## Remaining Chapter 10 Concepts

| Concept | Status | Existing repo support | Remaining work |
|---|---|---|---|
| Backtester vs live market dislocations | Partial | `realism/`, `research/chapter10_realism_audit.*`, `research/chapter9_validation_audit.*` | Add concrete latency, place-in-line, market-impact, and data-quality models. |
| Signal validation | Partial | `series/`, `indicators/`, `signals/`, `features/`, `labels/`, `datasets/`, research CSVs | Add signal dictionary/database and per-signal validation/decay records. |
| Strategy validation | Partial | `backtest/`, `strategies/`, `research/chapter*_analysis.*` | Separate system-flow event tests from performance event tests; add realism-adjusted validation reports. |
| Risk estimates | Partial | `riskmetrics/`, `risk/`, Chapter 6 reports | Adjust risk estimates for simulation dislocation and realism grade. |
| Risk management system | Partial | `risk/`, `riskmetrics/`, Chapter 6 promotion gates | Add realistic risk-adjustment recommendations only; do not enforce live controls yet. |
| Choice of strategies for deployment | Partial | Chapter 6 rankings and promotion gates | Add Chapter 10 realism gate before any candidate can be considered for forward testing. |
| Expected performance | Partial | `backtest.Report`, Chapter 9 comparison, Chapter 10A audit | Add expected-performance haircut/adjustment from realism assumptions. |
| Slippage | Partial | `backtest/slippage.go` | Calibrate assumptions and distinguish research slippage from event-driven fill slippage. |
| Fees | Mostly exists | `backtest/fees.go` | Keep; only document calibration status. |
| Operational issues | Missing | `system/command_control.go`, `system/supervisor.go` provide lifecycle foundation | Add operational issue catalog/checklist; no live ops. |
| Market data issues | Partial | `marketdata/`, `exchanges.Candle`, normalized candles | Add historical data quality checks and historical/live format parity audit. |
| Latency variance | Missing | `backtest/clock.go` provides deterministic time source | Add static latency and latency variance model for simulation only. |
| Place-in-line estimates | Missing | `system/order_book.go` has sorted book state | Add queue-position estimate model; no real venue claims. |
| Market impact | Missing | None dedicated | Add simple notional/size-based impact model for audit/backtest assumptions. |
| Historical market data accuracy | Partial | REST candle fetching and normalized candle structs | Add quality report: gaps, duplicates, sort order, abnormal OHLCV, venue count mismatch. |
| Backtester bias | Partial | `realism.SimulationDislocation` | Add expected-performance adjustment and bias/haircut report. |
| Live strategy analytics | Missing | No live/paper trading; research analytics exist | Defer active live analytics. Add schema only later if needed. |
| Profit decay | Missing | Strategy historical reports exist | Add research-only decay metrics across rolling windows. |
| Signal dictionary/database | Missing | `signals/`, `features/`, `labels/` exist | Add CSV/JSON dictionary of signal definitions, parameters, validation stats, and status. |
| Optimizing signals/models/parameters | Partial | Many fixed strategy configs exist | Defer optimizer engine; add audit/staging only after signal dictionary exists. |
| Researching new signals | Partial | Indicator/signal packages exist | Add research workflow metadata, not new strategies yet. |
| Portfolio optimization | Missing by design | Some statarb/pairs research exists | Defer; do not implement in this Chapter 10 finish pass. |

## Already Exists

- For-loop backtester for fast signal and strategy research.
- Event-driven backtester skeleton for OMS/system/market-simulator validation.
- Simulated clock and candle iterator.
- Backtester assumptions packet.
- For-loop vs event-driven comparison.
- Chapter 9 validation audit that explains artificial event-driven PnL.
- Chapter 10A realism audit and realism grade.
- Fee and slippage models.
- Risk metrics, risk filters, rankings, and promotion gates.
- Normalized market data and exchange state.
- Research CSV/JSON/Markdown export pattern.

## Partially Implemented

- Realism-adjusted expected performance.
- Market data quality validation.
- Signal validation.
- Strategy validation under realistic assumptions.
- Risk estimate adjustment.
- Deployment candidate selection.
- Slippage and fees calibration.
- Backtester bias classification.
- Event-driven simulator realism.

## Missing

- Latency model.
- Latency variance model.
- Event response scheduling.
- Place-in-line estimate.
- Partial fill and fill-ratio realism in event-driven simulator.
- Market impact model.
- Historical market data accuracy report.
- Signal dictionary/database.
- Profit decay tracking.
- Operational issue catalog.
- Realism-aware candidate gate.
- Expected-performance haircut.

## Grouping

The remaining work should be grouped by dependency order, not by every individual Chapter 10 subsection.

### Group 1: Simulator Realism Inputs

Includes:

- latency;
- latency variance;
- response delays;
- place-in-line;
- partial fill assumptions;
- market impact.

Why together: these are all causes of simulation dislocation and directly affect event-driven fill realism.

### Group 2: Data Quality And Historical/Live Parity

Includes:

- historical candle accuracy;
- gaps/duplicates/out-of-order data;
- abnormal OHLCV checks;
- cross-venue count mismatch;
- historical/live normalized format parity documentation.

Why together: Chapter 10 says signal validation fails if historical playback differs from what live systems would see.

### Group 3: Realism-Adjusted Performance And Risk

Includes:

- expected-performance haircut;
- backtester optimism/pessimism adjustment;
- risk estimate adjustment;
- realism-aware strategy candidate gate;
- deployment-readiness report.

Why together: once realism assumptions and data quality are known, strategy/risk outputs can be adjusted.

### Group 4: Signal Dictionary And Profit Decay

Includes:

- signal dictionary/database;
- signal parameter catalog;
- rolling validation stats;
- profit decay metrics;
- new-signal research workflow metadata.

Why together: this is the "continued profitability" half of Chapter 10.

### Group 5: Final Chapter 10 Packet

Includes:

- final summary;
- what is complete;
- what is deferred;
- why no paper/live trading is enabled;
- readiness statement.

Why together: this closes the book-aligned Chapter 10 track.

## Smallest Number Of Remaining Phases

Chapter 10 can be finished in **five remaining phases** after 10A.

### Chapter 10B: Simulation Realism Models

Build research/simulation-only models for:

- latency;
- latency variance;
- event response delay;
- place-in-line estimate;
- partial fill/fill ratio assumptions;
- market impact.

Do not alter production strategies. Do not enable paper/live trading.

### Chapter 10C: Market Data Quality And Parity Audit

Build research audits for:

- candle gaps;
- duplicate timestamps;
- out-of-order data;
- abnormal OHLCV rows;
- cross-venue candle count mismatch;
- historical/live normalized format parity checklist.

No real-time subscriptions. Use existing fetched historical candles only.

### Chapter 10D: Signal Dictionary And Profit Decay Research

Build:

- signal dictionary CSV/JSON;
- signal parameter catalog;
- rolling signal validation;
- rolling strategy PnL/decay metrics;
- "research new signals" workflow metadata.

No new signal logic and no optimization engine yet.

### Chapter 10E: Final Chapter 10 Packet

Generate:

- final Chapter 10 packet JSON/Markdown;
- complete/deferred map;
- final readiness statement.

Conclusion should still say no live/paper trading is enabled unless a later explicit task changes that.

## Exact Order

1. Chapter 10B - Simulation realism models.
2. Chapter 10C - Market data quality and historical/live parity audit.
3. Chapter 10D - Signal dictionary and profit decay research.
4. Chapter 10E - Final Chapter 10 packet.

This is the smallest clean sequence because each phase feeds the next:

- 10B defines execution realism assumptions.
- 10C defines data realism assumptions.
- 10D handles ongoing signal tracking, profitability, and decay.
- 10E closes the chapter.

## Never Implement Because It Duplicates Existing Work

Do not rebuild these:

- another generic for-loop backtester;
- another generic event-driven backtester skeleton;
- another fee model package;
- another slippage model package;
- another risk ranking/promotion gate system;
- another strategy comparison CSV framework;
- another Chapter 9 packet;
- another gateway abstraction;
- another normalized market data type layer;
- another indicator/signal package for Chapter 2 indicators.

Instead, Chapter 10 should extend the current layers:

- `backtest/` for simulation assumptions;
- `realism/` for dislocation and realism scoring;
- `research/` for audits and packets;
- `marketdata/` for data quality checks;
- `riskmetrics/` and `risk/` for realism-adjusted risk reporting.

## Deferred Beyond Chapter 10 Finish

These are book concepts, but should not be implemented in the remaining Chapter 10 finish pass:

- paper trading;
- live trading;
- real exchange execution;
- ML training;
- regime predictive allocation;
- Markowitz optimizer;
- full portfolio optimization engine;
- automated parameter optimizer;
- new production strategies.

Reason: the current repo is still in research/simulation validation. Chapter 10 should finish by making research and simulation realism explicit, not by turning on deployment machinery.

## Recommended Next Implementation Task

```text
Implement Chapter 10B: simulation realism models.

Do not add:
- live trading
- paper trading
- ML training
- real exchange calls
- new strategy behavior
- portfolio optimization

Use existing:
- realism/
- backtest/
- system/
- research/chapter10_realism_audit.*
- research/chapter9_validation_audit.*

Add:
- realism/latency.go
- realism/place_in_line.go
- realism/market_impact.go
- realism/fill_assumptions.go
- realism/latency_test.go
- realism/place_in_line_test.go
- realism/market_impact_test.go
- research/chapter10_simulation_realism.go
- research/chapter10_simulation_realism_test.go

Implement:
1. LatencyModel
   - static latency
   - jitter/variance
   - deterministic seeded calculation for tests

2. PlaceInLineEstimate
   - price level
   - size ahead
   - our size
   - estimated fill probability

3. MarketImpactEstimate
   - notional
   - depth
   - impact bps
   - adjusted expected price

4. FillAssumption
   - fill ratio
   - partial fill support flag
   - one-leg risk flag

5. Research report:
   - research/chapter10_simulation_realism.json
   - research/chapter10_simulation_realism.md

Update cmd/main.go:
- print Chapter 10B simulation realism summary

Validation:
- go test ./...
- go run ./cmd/main.go

Conclusion must state:
- Chapter 10B models realism assumptions only.
- No live or paper trading is enabled.
- These models are not yet applied to production strategy behavior.
- Next phase should audit historical market data quality and historical/live parity.
```
