# Chapter 7 Trading System Architecture

This package implements the Chapter 7 trading-system skeleton in simulated form only.

## Components

- `LiquidityProvider`: inserts simulated bid/ask liquidity into `lp_2_gateway`.
- `OrderBook`: consumes `lp_2_gateway`, maintains sorted bid and ask books, preserves FIFO ordering at the same price, and emits `ob_2_ts` only when top-of-book changes.
- `TradingStrategy`: consumes `ob_2_ts`, detects crossed-book arbitrage, creates simulated buy-at-ask and sell-at-bid intents on `ts_2_om`, receives `om_2_ts`, and tracks position, cash, and realized PnL.
- `OrderManager`: consumes `ts_2_om`, validates orders, assigns internal `om-*` IDs, tracks order state, forwards valid orders to `om_2_gw`, consumes `gw_2_om`, and routes responses back on `om_2_ts`.
- `MarketSimulator`: consumes `om_2_gw`, accepts valid new orders first, rejects duplicates and invalid cancel/amend requests, supports cancel/amend, fills accepted orders through `FillAllOrders`, emits `gw_2_om`, and writes audit events to `ms_2_om`.
- `TestTradingSimulation`: wires every component and verifies the full simulated lifecycle.

## Queues

- `lp_2_gateway`
- `ob_2_ts`
- `ts_2_om`
- `ms_2_om`
- `om_2_ts`
- `gw_2_om`
- `om_2_gw`

## Safety

This chapter layer does not import venue adapters, does not call exchange APIs, does not enable paper trading, and does not place live orders.

## Chapter 7B Notes

The current implementation mirrors the book's educational trading-system flow more closely:

- Liquidity updates flow from `LiquidityProvider` to `OrderBook`.
- Crossed top-of-book events create two simulated strategy orders.
- The `OrderManager` assigns internal IDs and owns lifecycle state.
- The `MarketSimulator` accepts orders before fills, then `FillAllOrders` produces final fill responses.
- The strategy PnL example captures the bid/ask spread from the crossed book.

Remaining Chapter 7 work is still simulated architecture work, not exchange connectivity:

- Expand the order book beyond simple insert-only behavior if needed.
- Add richer OMS validation policies before any paper/live path exists.

## Chapter 7C Command/Control and Services

The simulated command/control layer now models the book's non-critical operational components:

- `CommandControl` supports `START`, `STOP`, `STATUS`, `PAUSE`, and `RESUME`.
- System states are `CREATED`, `RUNNING`, `PAUSED`, `STOPPED`, and `ERROR`.
- Every command is recorded in an audit log.
- `Service` defines `Name`, `Start`, `Stop`, and `Status`.
- Core simulated services are `LoggingService`, `PositionService`, `MarketDataService`, `OrderService`, and `RiskService`.
- `SystemSupervisor` registers services, starts/stops all services, and reports service status summaries.

The `RiskService` is intentionally a placeholder for lifecycle wiring only. It does not introduce new risk logic.
