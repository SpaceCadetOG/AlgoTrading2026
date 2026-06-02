# Volume Profile Trading Style Playbook

## Trading Style Map

### Intraday

Intraday work should use the smallest context windows first:

- `daily_session` scoped profile
- `rolling_3d` scoped profile
- VWAP session features
- price-action open, daily open, failed auction, and strong/weak high-low studies
- latest L2 context as current market context, not historical replay

Intraday research should focus on whether price accepts or rejects around POC, VAH, VAL, HVN, LVN, VWAP, daily open, and prior daily high/low.

### Swing

Swing research should emphasize:

- `rolling_7d` scoped profile
- `composite_30d` scoped profile
- accumulation and trend setup comparisons
- confluence between flexible profiles and broader scoped POCs
- risk metrics and drawdown behavior from existing backtest reports

Swing setups should be reviewed for whether value migrates, whether rejected areas remain defended, and whether passive follow-through persists beyond 20 candles.

### Long-Term

Long-term investing context is not an execution path in this repo. It is represented as research context through:

- `composite_30d` profile
- profile shape study
- volume accumulation and broad value-area location
- major POC/VAH/VAL levels

No long-term investing strategy has been implemented.

## Instrument Selection

Instrument selection should consider:

- Liquidity: depth near price and near VWAP from `orderbook/` and L2 research exports.
- Spread: spread percent from normalized order book metrics.
- Volatility: Chapter 5 volatility metrics.
- Data availability: candles, trades where available, L2 snapshots, and recorder coverage.
- Venue reliability: successful REST candle and L2 snapshot reads across Aster, Hyperliquid, and Lighter.

## Market Analysis A To Z

The current workflow is:

1. Price action context.
2. VWAP and EMA context.
3. Volume Profile foundation and scoped profiles.
4. Flexible profiles around market events.
5. L2 context for spread, imbalance, and nearby liquidity.
6. Setup studies and quality reviews.
7. Cross-setup comparison.
8. Risk and realism reports.

This is research-only. It does not create executable entries or exits.

