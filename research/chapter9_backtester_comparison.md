# Chapter 9E For-Loop vs Event-Driven Backtester Comparison

## Result

- symbol: BTCUSDT
- candles: 25
- for_loop_trades: 6
- event_driven_orders: 22
- event_driven_fills: 22
- for_loop_final_equity: 9999.41
- event_driven_final_pnl: 4462.13
- pnl_difference: 4462.72
- assumptions_difference: for-loop backtester uses direct candle, signal, and simulated fill model; event-driven backtester routes candles through liquidity, order book, strategy, OMS, and market simulator queues; for-loop fills use fee and slippage models; event-driven fills use deterministic crossed-book test liquidity; results are expected to differ because market assumptions differ
- recommendation: Use for-loop backtests for fast signal research and event-driven backtests for system/OMS/market-simulator assumption validation.

## Implemented

- Runs the existing for-loop candle-driven backtester on the same candle sample.
- Runs the Chapter 9D event-driven backtester on the same candle sample.
- Computes PnL difference without forcing assumptions or results to match.
- Documents why the two backtesters are expected to diverge.

## Interpretation

- For-loop backtesting is best for fast signal research.
- Event-driven backtesting is best for validating system flow, OMS behavior, market simulator behavior, and time assumptions.
- PnL differences are diagnostic, not errors, because the engines model different execution assumptions.

## Assumption Comparison

- for-loop backtester uses direct candle, signal, and simulated fill model
- event-driven backtester routes candles through liquidity, order book, strategy, OMS, and market simulator queues
- for-loop fills use fee and slippage models; event-driven fills use deterministic crossed-book test liquidity
- results are expected to differ because market assumptions differ

## Conclusion

Chapter 9E compares for-loop and event-driven backtests honestly while keeping execution simulated.

## Next Phase

Chapter 9F should add the book's dual moving average event-driven comparison using simulated queues and simulated time.
