# Chapter 9C Simulated Clock and Value-of-Time Layer


## Implemented

- backtest.SimulatedClock exposes Current, Advance, Set, Reset, and StepTo.
- SimulatedClock rejects backward movement unless Reset is used explicitly.
- backtest.CandleTimeIterator sorts candles by timestamp and advances the simulated clock candle-by-candle.
- The main harness validates the time model against 2880 candles.

## Current Candle Timestamp Usage

The existing for-loop backtester uses candle StartTime and EndUTC timestamps for replay, fills, equity points, and trade timestamps.

## Why Wall-Clock Time Is Not Used

Backtests need deterministic simulated time so repeated runs do not depend on machine clock, runtime duration, network timing, or wall-clock scheduling.

## Remaining Gaps

- OMS timeout modeling is not implemented.
- Latency simulation is not implemented.
- Event scheduling is not implemented.
- Exchange response delays are not modeled.

## Readiness For Event Backtest

Ready to use as the deterministic time source for a Chapter 9 event-driven backtester skeleton.
