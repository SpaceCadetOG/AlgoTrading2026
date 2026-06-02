# VWAP L2 Research Review

## Data Sources

Reviewed outputs:

- `research/vwap_features_l2.csv`
- `research/context_features_l2.csv`
- `research/vwap_behavior_l2_summary.json`
- `research/orderbook_features.csv`
- `research/multi_venue_orderbook_features.csv`

The refresh joins 30-day historical 15m candles with the latest valid L2 snapshot per venue. This is useful for checking schema, export shape, and first-pass microstructure context, but it is not historical L2 replay.

L2 venues included:

| Venue | Symbol | Rows | Snapshot status |
|---|---:|---:|---|
| Hyperliquid | BTC | 2881 | valid |
| Aster | BTCUSDT | 2880 | valid |
| Lighter | BTC | 2880 | valid |

## Order Book Helper Usage

The L2 feature exports are derived from the normalized order book layer and existing helper methods:

- `orderbook.BestBid`
- `orderbook.BestAsk`
- `orderbook.Spread`
- `orderbook.SpreadPct`
- `orderbook.Mid`
- `orderbook.BidDepthWithinPct`
- `orderbook.AskDepthWithinPct`
- `orderbook.DepthWithinPct`
- `orderbook.Imbalance`
- `orderbook.LiquidityNearPrice`
- `orderbook.ValidateSnapshot`

This review consumes those exported helper-derived fields rather than introducing a separate order book calculation path.

## Venue Spread Quality

| Venue | Avg spread % | Avg imbalance 1% | Avg bid depth 1% | Avg ask depth 1% | Avg liquidity near VWAP |
|---|---:|---:|---:|---:|---:|
| Aster | 0.000136 | -0.061757 | 6,063,326 | 6,861,530 | 0 |
| Hyperliquid | 0.001363 | -0.432697 | 5,590,761 | 14,119,199 | 0 |
| Lighter | 0.000136 | 0.119754 | 52,990,863 | 41,656,490 | 2,567,786 |

Rankings:

| Category | Rank |
|---|---|
| Tightest spread | 1. Aster, 2. Lighter, 3. Hyperliquid |
| Deepest 1% liquidity | 1. Lighter, 2. Hyperliquid, 3. Aster |
| Strongest bid pressure | 1. Lighter |
| Strongest ask pressure | 1. Hyperliquid, 2. Aster |

Strongest finding: Aster and Lighter both showed extremely tight top-of-book spread in the latest snapshot, while Lighter had by far the deepest 1% displayed liquidity and the only non-zero liquidity near session VWAP.

Weakest finding: these venue rankings are from one latest snapshot applied across all historical rows, so they should not be interpreted as persistent venue behavior.

## Venue Book Pressure

Hyperliquid showed strong ask-side pressure at `-0.4327`, meaning ask depth inside 1% was much larger than bid depth at the captured snapshot.

Aster also leaned ask-side, but mildly, at `-0.0618`.

Lighter leaned bid-side at `0.1198`, with substantially more total displayed depth than the other venues.

This is a useful snapshot-level contrast. It is not enough to infer venue personality or a stable market condition.

## VWAP Liquidity Context

Overall VWAP L2 behavior summary:

| Metric | Value |
|---|---:|
| Candles | 8641 |
| Touches | 278 |
| Crosses | 135 |
| Bounces | 112 |
| Breaks | 71 |
| Chops | 93 |
| Bounce + bid support | 7 |
| Bounce + ask weakness | 7 |
| Break + ask pressure | 67 |
| Break + bid collapse | 67 |
| Avg spread % | 0.000545 |
| Avg imbalance 1% | -0.124936 |
| Avg liquidity near VWAP | 855,830 |

The most useful observation is the contrast between current price and VWAP liquidity:

- Hyperliquid and Aster had zero liquidity near session VWAP in the latest snapshot.
- Lighter had meaningful liquidity near VWAP.

That confirms the research question is valid: price can be far from VWAP, and the book can show whether there is actually displayed liquidity near the VWAP zone.

## Trend Alignment + L2 Context

| Trend alignment | Rows | Avg spread % | Avg imbalance 1% | Avg liquidity near VWAP | Spread quality distribution | Book pressure distribution | Liquidity quality distribution |
|---|---:|---:|---:|---:|---|---|---|
| strong_bull | 1064 | 0.000592 | -0.158352 | 209,597 | tight:428; wide:636 | ask_pressure:395; balanced:428; bid_pressure:241 | no_near_vwap_liquidity:823; strong_near_vwap:241 |
| bull | 783 | 0.000378 | -0.033642 | 514,260 | tight:193; wide:590 | ask_pressure:154; balanced:193; bid_pressure:436 | no_near_vwap_liquidity:347; strong_near_vwap:436 |
| neutral | 3303 | 0.000597 | -0.155955 | 1,342,720 | tight:1240; wide:2063 | ask_pressure:1241; balanced:1240; bid_pressure:822 | no_near_vwap_liquidity:2487; strong_near_vwap:816 |
| bear | 1813 | 0.000352 | -0.008186 | 998,974 | tight:307; wide:1506 | ask_pressure:319; balanced:307; bid_pressure:1187 | no_near_vwap_liquidity:626; strong_near_vwap:1187 |
| strong_bear | 1678 | 0.000701 | -0.211431 | 311,921 | tight:712; wide:966 | ask_pressure:772; balanced:712; bid_pressure:194 | no_near_vwap_liquidity:1484; strong_near_vwap:194 |

Does book pressure align with trend alignment?

Not reliably. Strong bearish rows show the most negative average imbalance, which directionally fits ask pressure. But bear rows show many bid-pressure observations because Lighter rows carry bid-pressure and deep liquidity. Bull rows also show a mixed picture.

The core issue is that historical trend labels are being compared to a latest L2 snapshot. This makes the trend/L2 relationship more of a venue-composition artifact than a true time-series relationship.

Useful hypothesis: if historical L2 confirms that strong bearish trend alignment tends to occur with ask pressure, that may support VWAP break/bid-collapse studies. Current outputs cannot prove it.

## VWAP Regime + L2 Context

| VWAP regime | Rows | Avg spread % | Avg imbalance 1% | Avg bid depth 1% | Avg ask depth 1% | Avg liquidity near VWAP |
|---|---:|---:|---:|---:|---:|---:|
| above_falling | 69 | 0.000794 | -0.234361 | 12,611,014 | 15,796,071 | 114,244 |
| above_rising | 2273 | 0.000607 | -0.173318 | 13,830,605 | 15,539,377 | 154,820 |
| below_falling | 4098 | 0.000711 | -0.218762 | 10,159,185 | 13,461,131 | 494,665 |
| below_rising | 82 | 0.000854 | -0.256757 | 11,509,573 | 15,353,209 | 173,449 |
| neutral | 2119 | 0.000139 | 0.117080 | 52,524,680 | 41,328,786 | 2,356,806 |

Strongest bid support by average imbalance:

1. `neutral`: 0.117080

Strongest ask pressure by average imbalance:

1. `below_rising`: -0.256757
2. `above_falling`: -0.234361
3. `below_falling`: -0.218762

The neutral regime had the largest liquidity near VWAP and the only positive average imbalance. This is likely because Lighter contributes both deeper liquidity and bid-pressure in the current snapshot.

The regime result is interesting, but it should not drive strategy work yet because the L2 values are current snapshot context rather than historical L2 at each candle.

## Book 1 Framework Recheck

L2 does not require a redesign of the Book 1 framework.

| Area | Does L2 change it now? | Notes |
|---|---|---|
| Indicators | No | Existing indicator layer remains price/volume based. L2 can become optional feature input later. |
| Signals | No | No signal behavior should change from this review. |
| Backtesting | No immediate redesign | Existing for-loop and event-driven backtests remain valid for their current purposes. Historical L2 would be needed before L2-aware fill assumptions are realistic. |
| Risk analysis | No | Current risk analytics and controls are still strategy/backtest level. L2 may later improve liquidity/participation checks. |
| Realism models | Partial future enhancement | L2 can strengthen slippage, market impact, and place-in-line assumptions, but latest-snapshot joins are not enough. |
| Data quality | Partial future enhancement | L2 snapshots add a new data quality surface: crossed books, empty sides, stale timestamps, depth anomalies. |
| Strategy research workflow | No redesign | L2 should be added as research context/features, not as a reason to rewrite the framework. |

Book 1 assumptions remain intact: candle-based indicators and research outputs are still the foundation. L2 expands the research dataset; it does not replace the original architecture.

## Data Limitations

The main limitation is decisive:

Current VWAP L2 refresh equals historical candles plus latest available valid L2 snapshots.

Implications:

- L2 values repeat across all historical rows for a venue.
- Trend and VWAP regime comparisons are dominated by venue/snapshot composition.
- Bounce/break labels are historical, but L2 support/pressure labels are current snapshot context.
- Liquidity near VWAP is meaningful for the latest market state, not for the historical candle being labeled.
- Strong conclusions about behavior require historical L2 recording or replay.

Other limitations:

- Only 1% depth metrics are included.
- No trade prints/tape are joined here.
- No intrabar order book path exists.
- No queue position or fill probability is inferred from historical L2.
- Spread quality is relative to this small three-venue sample, not a universal exchange quality grade.

## Hypotheses Worth Testing Later

Good hypotheses for later, once historical L2 exists:

1. VWAP bounces are more reliable when bid depth near VWAP expands before touch.
2. VWAP breaks are more likely when same-side liquidity collapses before touch.
3. Strong bull alignment with bid imbalance has better continuation than strong bull alignment with ask pressure.
4. Strong bear alignment with ask pressure has higher continuation risk than strong bear alignment with bid support.
5. Lighter's deeper book may improve liquidity-near-VWAP studies, but venue-specific spread and feed behavior must be tracked over time.
6. Aster's tight spread may matter more for execution quality than directional pressure.
7. Hyperliquid's ask-heavy snapshot should be retested over time to see whether it is persistent or momentary.

## Recommended Next Move

Recommendation: **A. Pause VWAP and inspect results manually.**

Reason:

The output shape is correct, the multi-venue L2 schema is working, and the initial review surfaced useful questions. But the current join is latest-L2-only, so starting Chapter 4 or VWAP strategy research now would overfit to a snapshot artifact.

Conservative next action:

Inspect:

- `research/vwap_features_l2.csv`
- `research/context_features_l2.csv`
- `research/vwap_behavior_l2_summary.json`
- `research/multi_venue_orderbook_features.csv`

Then decide whether to build **C. historical L2 recorder** before moving to VWAP Chapter 4.

