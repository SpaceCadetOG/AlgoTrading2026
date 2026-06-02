# Volume Profile Book Audit vs Existing Framework

Source boundary: no local Volume Profile book/PDF was found in the workspace during this audit. This review is based on the Volume Profile methodology concepts named in the task and compares them against the current `vwap/`, `orderbook/`, `features/`, `labels/`, `backtest/`, and `research/` layers.

## SECTION 1 - Concepts Already Covered

The current framework already covers several adjacent ideas that Volume Profile can build on, but it does not yet implement Volume Profile itself.

| Volume Profile-adjacent concept | Current representation | Repo files/packages |
|---|---|---|
| Fair value / reference price | Session VWAP, anchored VWAP, distance from VWAP | `vwap/session_vwap.go`, `vwap/anchored_vwap.go`, `vwap/distance.go`, `research/vwap_features.go` |
| Price relative to context | Price vs VWAP, price vs EMA9/EMA20, trend alignment | `vwap/ema_relationships.go`, `research/context_features.go`, `research/trend_alignment.go` |
| Acceptance/rejection proxy | VWAP touches, crosses, bounces, breaks, chops | `vwap/interaction.go`, `vwap/reaction.go`, `research/vwap_interaction.go`, `research/vwap_behavior_summary.go` |
| Liquidity near a reference price | L2 liquidity near VWAP | `orderbook/metrics.go`, `research/orderbook_features.go`, `research/vwap_l2_refresh.go` |
| Spread quality | Spread and spread percent from normalized L2 snapshots | `orderbook/metrics.go`, `research/orderbook_features.go`, `research/vwap_l2_refresh.go` |
| Book imbalance | Bid/ask depth imbalance inside 1% | `orderbook/metrics.go`, `research/orderbook_features.go` |
| Depth context | Bid depth, ask depth, total depth within percent bands | `orderbook/metrics.go` |
| Multi-venue microstructure context | Hyperliquid, Aster, and Lighter L2 snapshots normalized into one schema | `orderbook/types.go`, `exchanges/*/orderbook.go`, `research/multi_venue_orderbook_features.csv` |
| Feature export foundation | Candle, indicator, VWAP, context, L2, and ML-staging feature exports | `features/`, `research/context_features_l2.csv`, `research/vwap_features_l2.csv` |

Current output files that overlap with future Volume Profile work:

- `research/vwap_features.csv`
- `research/context_features.csv`
- `research/trend_alignment.csv`
- `research/orderbook_features.csv`
- `research/vwap_features_l2.csv`
- `research/context_features_l2.csv`
- `research/vwap_behavior_l2_summary.json`

The closest existing equivalent to Volume Profile logic is the VWAP + L2 layer: it measures price relative to a volume-weighted reference and then asks whether displayed liquidity exists near that reference. That is not the same as a price-by-volume distribution, but it is compatible with one.

## SECTION 2 - Missing Concepts

| Concept | Description | Why it matters | Difficulty | Potential future data requirements |
|---|---|---|---|---|
| Point of Control (POC) | The price level with the highest traded volume over a profile window. | Identifies the main accepted price/fair-value area for a session or range. | Medium | Historical candles can approximate it using volume binned by price; trade prints give better precision. |
| High Volume Nodes (HVN) | Price zones with unusually high accumulated volume. | Often interpreted as acceptance areas, congestion, or magnet zones. | Medium | OHLCV approximation is possible; trade-level volume by price is better. |
| Low Volume Nodes (LVN) | Price zones with low accumulated volume between acceptance areas. | Often treated as rejection zones, fast-travel zones, or weak auction areas. | Medium | Needs a volume-by-price histogram; precision improves with trades/tick data. |
| Value Area | Price range containing a chosen share of volume, commonly around 70%. | Helps distinguish accepted value from excess/rejection. | Medium | Volume profile histogram by session/range. |
| Value Area High / Low | Upper/lower bounds of the value area. | Useful context levels for acceptance, rejection, and range transitions. | Medium | Same as value area. |
| D-profile | Balanced, bell-shaped volume distribution. | Suggests two-sided trade and acceptance around fair value. | Medium | Profile-shape classification over a session/window. |
| P-profile | Profile with volume concentration near highs and thinner lower distribution. | Often associated with short covering or upward auction imbalance. | Medium | Profile-shape classification; ideally sessionized intraday data. |
| b-profile | Profile with volume concentration near lows and thinner upper distribution. | Often associated with long liquidation or downward auction imbalance. | Medium | Profile-shape classification; ideally sessionized intraday data. |
| Double distribution profile | Two separate high-volume areas in one session/range. | Indicates regime shift or two accepted value areas. | Medium/High | More robust histogram peak detection. |
| Volume accumulation zones | Price bands where volume builds over time. | Potential acceptance, congestion, and future reaction zones. | Medium | Rolling volume-by-price profiles. |
| Rejection zones | Price bands quickly visited with low volume or sharp return away. | Helps identify prices the market did not accept. | Medium | Candle-only approximation possible; trade/tick data improves quality. |
| Acceptance vs rejection | Whether price spends time and volume at a level or rejects quickly. | Core auction-market framing for Volume Profile. | Medium | Time-at-price plus volume-at-price; candles can approximate time, trades improve volume allocation. |
| Session profile structures | Profiles calculated per day/session instead of one rolling window. | Crypto trades 24/7, so session definition becomes a research choice. | Medium | Timestamp/session boundaries and volume-by-price bins. |
| Composite profile | Profile over multiple sessions/ranges. | Useful for longer-term accepted value and major nodes. | Medium | Multi-day candle/trade history. |
| Volume-weighted profile bins | Allocation of candle volume across price bins. | Needed for candle-only profile approximation. | Medium | OHLCV candles; better with lower timeframe candles or trades. |
| Initial balance / range context | Early-session high/low or opening distribution context. | Common market profile companion concept; useful for session structure. | Medium | Session definitions and intraday candles. |
| Profile migration | How POC/value area shifts over time. | Helps distinguish trend acceptance from range rotation. | Medium | Rolling or session profile sequence. |
| POC retest behavior | How price behaves when returning to POC. | Potential future research label/feature, not a strategy yet. | Medium | Profile levels plus candles; L2/trades add confirmation. |

Important distinction: the framework currently has VWAP and L2 liquidity near VWAP, but it does not yet have volume-at-price. Volume Profile begins when volume is accumulated by price bins, not just by candle timestamp.

## SECTION 3 - Impact On Book Framework

| Layer | Does Volume Profile change it now? | Audit conclusion |
|---|---|---|
| Indicators layer | Additive only | Future Volume Profile can be implemented as another research/indicator-like layer, but existing indicators do not need redesign. |
| Signals layer | No immediate change | No entries/exits should be added from this audit. Volume Profile may later produce signal inputs after research validation. |
| Backtester | No immediate change | Existing backtester can consume future profile-derived signals/features. It does not need to change until profile-aware fills/slippage are explicitly modeled. |
| Feature layer | Yes, future extension | Feature rows can later add POC, distance from POC, HVN/LVN proximity, value-area location, profile shape, and POC migration. |
| Label layer | Possible future extension | Labels could later measure return-to-POC, HVN bounce/rejection, LVN traversal, or value-area acceptance. Current labels remain valid. |
| Risk layer | No immediate change | Risk controls do not change. Later, volume-profile liquidity context may inform max size, participation, or avoid-low-liquidity zones. |
| VWAP research | Enhances, not replaces | VWAP can be compared against POC/value area. VWAP is a flow-weighted mean; POC is the most traded price bin. Divergence between them is useful research context. |
| Orderbook research | Enhances, not replaces | L2 tells current displayed liquidity; Volume Profile tells historically traded liquidity by price. Together they answer different questions. |
| ML staging | Future feature source only | No ML scoring now. Later, profile features can join existing `features/` rows. |

Framework answer: Volume Profile does not require a redesign. It should be added as a research/feature layer after the current VWAP + L2 outputs are manually inspected.

## SECTION 4 - Impact On Existing L2 Layer

The normalized L2 layer can enhance future Volume Profile work, but it cannot replace volume-at-price.

| Existing L2 metric | Impact on future Volume Profile |
|---|---|
| Spread | Helps judge whether profile levels are tradable or noisy. Wide spread near POC/HVN weakens execution assumptions. |
| Imbalance | Adds current pressure context when price is near POC, HVN, LVN, VAH, or VAL. |
| Depth | Helps distinguish historical volume nodes from current displayed liquidity. A high-volume historical node with thin current depth may not behave as support/resistance. |
| LiquidityNearPrice | Directly reusable for liquidity near POC, HVN, LVN, VAH, VAL, or VWAP. |
| ValidateSnapshot | Needed before combining any live snapshot with profile levels. Prevents bad book data from contaminating research exports. |

Useful future combinations:

- `distance_from_vwap` plus `distance_from_poc`
- `liquidity_near_vwap` plus `liquidity_near_poc`
- `imbalance_1pct` when price is inside value area
- `spread_pct` near LVN traversal zones
- `depth_near_hvn` vs `depth_near_lvn`

The L2 layer answers "what liquidity is displayed now?" Volume Profile answers "where did volume historically trade?" Both are complementary.

## SECTION 5 - Recommended Roadmap

Recommended order:

1. **A. Historical L2 Recorder**
2. **B. Volume Profile Research Layer**
3. **C. Volume Profile Feature Exports**
4. **D. Volume Profile Signals**
5. **E. Volume Profile Backtests**

Rationale:

1. Historical L2 should come first because the most recent VWAP L2 audit showed that latest-snapshot joins can create misleading historical context. A recorder would allow future profile studies to compare historical traded volume with historical displayed liquidity.
2. Volume Profile Research Layer should come second and remain research-only: POC, HVN, LVN, value area, profile shape, and acceptance/rejection summaries.
3. Feature Exports should come third after the research definitions are stable.
4. Signals should wait until manual review shows that profile features are meaningful.
5. Backtests should be last because strategy behavior before stable profile definitions would overfit the first implementation.

## Recommended Next Move

Do not start Volume Profile signals or backtests yet.

The conservative next step is:

**A. Historical L2 Recorder**

If you want to stay strictly inside candle-only research before recording L2, the next-best step is:

**B. Volume Profile Research Layer**

But the cleaner research sequence is to solve historical L2 first, then build Volume Profile with both traded-volume context and recorded book context available.

## Summary

Existing framework alignment is strong:

- VWAP foundation covers fair-value/reference-price research.
- Context features cover price/EMA/VWAP alignment.
- L2 exports cover spread, depth, imbalance, and liquidity near price.
- Feature and label layers are ready for future profile fields.

Missing Volume Profile core:

- POC
- HVN/LVN
- Value area
- Profile shape classification
- Session/composite profiles
- Acceptance/rejection from volume-at-price

No current Book 1 framework layer needs to be redesigned. Volume Profile should be added as a future research layer, not as a framework rewrite.

