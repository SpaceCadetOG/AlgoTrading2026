# Chapter 10A Backtester Realism and Simulation Dislocation Audit

## Simulation Dislocation

- strategy: chapter9_backtester_comparison
- symbol: BTCUSDT
- candles: 25
- for_loop_pnl: -0.59
- event_driven_pnl: 4462.13
- pnl_difference: 4462.72
- pnl_difference_pct: 10000.00
- bias_direction: optimistic
- severity: high
- realism_grade: F

## Event-Driven Classification

system-validation only; not strategy-performance PnL

## Realism Strengths

- slippage modeled
- fees modeled
- for-loop and event-driven backtesters are separated
- Chapter 9 validation identifies crossed-book artificial profit
- simulated clock exists for deterministic event replay

## Realism Gaps

- latency not modeled
- latency variance not modeled
- place-in-line not modeled
- market impact not modeled
- market data accuracy not checked
- historical/live format parity not checked
- operational intervention not modeled
- live analytics not available
- profit decay not tracked

## Validation Findings

- event-driven PnL is dominated by deterministic crossed-book spread capture
- event-driven fills are full fills at submitted prices
- event-driven fills do not currently include fees, slippage, latency, queue position, partial fill ratio, or market impact
- OrderBook accumulates liquidity and does not consume or expire filled resting liquidity
- OMS does not yet enforce idempotency against duplicate fill responses

## Must Be Solved Before Paper/Live Trading

- replace artificial crossed-book generation for performance tests
- consume, expire, or reset order-book liquidity
- add latency and latency variance modeling
- add place-in-line and partial-fill modeling
- add market impact modeling
- calibrate fees and slippage assumptions
- add historical market data accuracy checks
- add OMS response idempotency

## Conclusion

- Chapter 10A measures simulation dislocation risk only.
- No live or paper trading is enabled.
- Current for-loop backtests are useful for strategy research.
- Current event-driven backtester is useful for OMS/system validation, not performance estimates.
- Next phase should model latency, market impact, and place-in-line before any forward-testing shell.

## Next Phase

Model latency, market impact, and place-in-line before any forward-testing shell.
