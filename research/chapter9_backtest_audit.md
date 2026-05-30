# Chapter 9B Backtest Data Splits and Assumptions

- in_sample_count: 2304
- out_of_sample_count: 576
- in_sample_ratio: 0.80

## Implemented Concepts

| Book Concept | Repo Mapping | Status |
|---|---|---|
| in-sample vs out-of-sample | backtest.InSampleOutOfSampleSplit | implemented |
| correct assumptions | backtest.BacktestAssumptions | documented |
| for-loop backtester | backtest.Engine and backtest.Replay | implemented |
| portfolio cash/holdings/total | backtest.PortfolioPoint and EquityCurve | implemented |
| dual moving average backtest | strategies/chapter4.DualMAStrategy | strategy implemented; Chapter 9 comparison pending |

## Current Assumptions

- engine_model: for-loop candle-driven backtester
- fill_model: simulated fills using candle close with venue-specific slippage
- fee_model: venue-specific fee model is implemented and configurable
- slippage_model: venue-specific slippage model is implemented as configurable placeholder assumptions
- fill_ratio: assumes 100 percent fill for simulated market fills
- latency_assumption: latency is not fully modeled
- partial_fill_support: partial fills are not fully modeled
- market_impact_support: market impact is not modeled
- lookahead_bias_guard: replay preserves timestamp order and signal execution is candle-by-candle; strategy-specific lookahead must still be reviewed
- survivorship_bias_note: crypto single-symbol research avoids equity delisting survivorship issues, but multi-asset survivorship-bias-free datasets are not implemented
- data_storage_note: research CSV and JSON exports exist; no HDF5, relational database, or time-series database layer is implemented
- event_driven_support_status: Chapter 7 components exist, but event-driven backtester is not yet complete
- paper_forward_testing_status: paper or forward testing is not enabled

## Assumption Gaps

- event-driven backtester not yet complete
- paper/forward testing not enabled
- latency not fully modeled
- partial fills not fully modeled
- market impact not modeled
- fill ratio is fixed at simulated full fill
- database/HDF5/time-series storage not implemented
- strategy-specific lookahead review remains required

## Missing Chapter 9 Concepts

- paper/forward testing is not enabled
- event-driven backtester is not yet complete
- simulated clock is not implemented
- latency/fill-ratio/partial-fill market simulator assumptions are not fully modeled
- HDF5/database/time-series historical data store is not implemented

## Conclusion

Chapter 9B formalizes historical data splits and documents current backtest assumptions before event-driven backtesting.

## Next Phase

Chapter 9C should add a simulated clock and event-driven backtester skeleton using Chapter 7 queues without enabling paper or live trading.
