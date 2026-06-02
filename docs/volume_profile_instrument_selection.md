# Volume Profile Instrument Selection

## Purpose

Define what makes an instrument worth researching with Volume Profile, VWAP, and L2 context.

## Selection Criteria

- Liquidity: enough depth to make profile levels meaningful.
- Spread: tight enough that passive follow-through is not swallowed by costs.
- Volatility: enough range to create auction structure, but not so erratic that levels are meaningless.
- Data availability: candles, L2 snapshots, and ideally trades.
- Venue reliability: consistent public market-data access.
- Open interest: useful for perps when available.
- Correlation: avoid duplicating the same exposure across highly related instruments.

## Asset Types

- Crypto: BTC, ETH, SOL, and major alts.
- RWA: tokenized stocks, treasuries, or real-world-asset tokens where detectable.
- Commodity: tokenized gold, oil, or similar markets where detectable.
- Index: index-style instruments where detectable.
- Unknown: fallback when symbol metadata is not enough.

## Excessive Exposure

BTC, ETH, and SOL can still behave like one crypto-beta basket during risk-off conditions. Instrument selection should track correlation before treating them as independent research candidates.

## Repo Mapping

- `scanner/` ranks instruments for research.
- `orderbook/` measures spread, depth, and imbalance.
- `l2recorder/` supports longer-term L2 dataset collection.
- `research/instrument_universe.csv` stores the research universe snapshot.

