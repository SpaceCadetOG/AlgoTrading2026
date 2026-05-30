# Chapter 9D Event-Driven Backtester

## Result

- candles_processed: 25
- orders_created: 8
- orders_filled: 8
- final_cash: 12252.50
- final_position: 0.0000
- final_pnl: 2252.50
- clock_start: 2026-04-30 02:15:00 +0000 UTC
- clock_end: 2026-04-30 08:15:00 +0000 UTC
- audit_events: 16

## Implemented

- Historical candles replay through backtest.CandleTimeIterator.
- SimulatedClock advances to each candle timestamp.
- Each candle becomes deterministic crossed-book liquidity for the Chapter 7 test strategy.
- Chapter 7 LiquidityProvider, OrderBook, TradingStrategy, OrderManager, and MarketSimulator are reused.
- Simulated fills return through the OMS to the strategy.

## Queue Flow

- lp_2_gateway -> OrderBook
- ob_2_ts -> TradingStrategy
- ts_2_om -> OrderManager
- om_2_gw -> MarketSimulator or SimulatedGateway
- gw_2_om -> OrderManager
- om_2_ts -> TradingStrategy
- ms_2_om -> OrderManager audit stream

## Safety

- AllowLiveOrders is forced to false.
- No venue adapters are called.
- No paper trading is enabled.
- No live trading is enabled.
- No WebSocket subscriptions or real API calls are made.

## Remaining Gaps

- Event-driven strategy is a deterministic crossed-book test strategy only.
- Latency and response scheduling are not modeled.
- OMS timeout behavior is not integrated with the simulated clock yet.
- Partial fills and fill ratios are not modeled in the event-driven simulator.
- Dual moving average event-driven strategy comparison remains pending.

## Conclusion

Chapter 9D adds an event-driven backtester skeleton using Chapter 7 system queues and Chapter 9 simulated time without enabling paper or live trading.

## Next Phase

Chapter 9E should add the book's dual moving average event-driven comparison while keeping execution simulated.
