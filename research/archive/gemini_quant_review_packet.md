# Gemini Quant Review Packet

## Section 1 - Project Context

AlgoTrading2026 is a Go research framework for learning the practical workflow behind algorithmic trading in a crypto/perpetuals context.

Primary goals:

- Learn indicators.
- Learn signals.
- Learn backtesting.
- Learn risk analytics.
- Learn a repeatable research process.

Current progression:

```text
Indicators
-> Signals
-> Backtesting
-> Risk
-> VWAP
-> Volume Profile
-> L2 Research
-> Feature Engineering later
-> Labels later
-> ML later
```

Current scope:

- No live trading.
- No paper trading.
- No ML.
- Research-only.
- No production execution or OMS activation.

## Section 2 - Dataset

Current primary candle dataset:

- Venue: Aster
- Symbol: BTCUSDT
- Timeframe: 15m
- Lookback: 30 days
- Candles: 2880
- Date range: 2026-05-02 -> 2026-06-01

Current L2 sources:

- Hyperliquid
- Aster
- Lighter

Historical L2 recorder:

- Implemented as a read-only snapshot recorder.
- Current L2 research supports historical snapshot collection going forward.
- Existing VWAP L2 refresh used latest available L2 snapshots with historical candles, not true historical L2 replay.

Volume Profile data approach:

- Uses OHLCV candle approximation.
- Candle volume is distributed across price bins between each candle high and low.
- This is not tick-accurate volume-at-price.
- Trade tape reconstruction and historical L2 replay remain future improvements.

## Section 3 - Implemented Research Layers

### VWAP

- Session VWAP
- Anchored VWAP
- VWAP distance
- VWAP slope
- VWAP regime
- VWAP behavior study
- VWAP interaction and reaction classification
- VWAP magnet study
- VWAP Chapter 3 context layer
- VWAP L2 refresh

### Order Book

- Normalized L2 snapshot schema
- Best bid
- Best ask
- Spread
- Spread %
- Mid
- Bid depth within percentage band
- Ask depth within percentage band
- Total depth within percentage band
- Imbalance
- Liquidity near price
- Snapshot validation
- Hyperliquid L2 adapter
- Aster L2 adapter
- Lighter L2 adapter
- Historical L2 recorder
- L2 snapshot analysis

### Volume Profile

- Price bins
- Candle-volume distribution
- POC
- VAH
- VAL
- HVN
- LVN
- Profile shapes
- Scoped profiles:
  - daily_session
  - rolling_3d
  - rolling_7d
  - composite_30d
- Flexible profiles:
  - sideways accumulation
  - open drive
  - failed auction
  - high/low retest
- Acceptance studies
- Setup quality reviews
- Cross-setup comparison

## Section 4 - Setup Results

The current comparison uses passive follow-through, acceptance/rejection labels, confluence rates, and setup counts. It does not simulate orders, stops, targets, slippage, or fills.

| Setup | Setups | Accepted | Rejected | Acceptance Rate | Rejection Rate | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| ACCUMULATION | 228 | 10 | 172 | 4.39% | 75.44% | 81.17 |
| TREND | 73 | 18 | 33 | 24.66% | 45.21% | -74.14 |
| REJECTION | 649 | 51 | 444 | 7.86% | 68.41% | -14.28 |
| REVERSAL | 129 | 0 | 0 | 0.00% | 0.00% | -26.83 |

Current rankings:

- bestAcceptanceSetup: TREND
- bestFT20Setup: ACCUMULATION
- bestVWAPSetup: TREND
- bestPOCSetup: REVERSAL
- bestHVNSetup: TREND
- largestSampleSetup: REJECTION

Additional comparison notes:

- ACCUMULATION has the best FT5, FT10, and FT20 rankings.
- TREND has the best acceptance ranking and strongest VWAP/HVN alignment rates.
- REJECTION has the largest sample size and lowest invalidation rate, but negative average follow-through.
- REVERSAL has the strongest POC confluence, but its current accepted/rejected fields are not populated in the same way as the other setup studies.

## Section 5 - Important Findings

- Trend had the best acceptance rate at 24.66%.
- Accumulation had the best passive FT20 at 81.17.
- Rejection had the largest sample size with 649 setups.
- Reversal had the strongest POC confluence rate at 70.54%.
- VWAP mattered most for Trend, with a 61.64% VWAP alignment rate.
- HVN mattered most for Trend, with a 63.01% HVN confluence rate.
- POC alone appears weak or incomplete as a standalone filter because Reversal had the strongest POC confluence but negative FT20 and no accepted/rejected labels.
- Acceptance and passive follow-through often disagreed:
  - Trend accepted most often but had negative FT20.
  - Accumulation accepted rarely but had positive FT20.
- Rejection generated many observations, but the broad version appears noisy.
- Current results are hypothesis-generating, not execution-ready.

## Section 6 - Questions For Gemini

Please act as a quantitative research reviewer.

1. What findings appear statistically meaningful?
2. Which setup currently looks most promising?
3. What filters would Gemini test first?
4. Does Gemini agree that `Trend > Rejection > Accumulation > Reversal` is the current usefulness ordering?
5. What features should be engineered next?
6. What labels should be created next?
7. What research appears noisy or misleading?
8. Should setup logic be refined before ML?
9. What would Gemini inspect manually on charts?
10. What is the best path from here to:

```text
Feature Engineering
-> Labels
-> ML Scoring
```

## Section 7 - Files

### VWAP

- `research/vwap_features.csv`
- `research/vwap_interaction.csv`
- `research/vwap_reactions.csv`
- `research/vwap_behavior_summary.json`
- `research/context_features.csv`
- `research/trend_alignment.csv`
- `research/vwap_features_l2.csv`
- `research/context_features_l2.csv`
- `research/vwap_behavior_l2_summary.json`
- `research/vwap_l2_research_review.md`

### Order Book And L2

- `research/orderbook_inventory.md`
- `research/orderbook_features.csv`
- `research/multi_venue_orderbook_features.csv`
- `research/vwap_l2_interaction.csv`
- `research/multi_venue_vwap_l2_interaction.csv`
- `research/l2_snapshot_analysis.json`
- `research/l2_snapshot_analysis.md`

### Price Action

- `research/price_action_features.csv`
- `research/price_action_summary.json`
- `research/price_action_strategy_study.csv`
- `research/price_action_strategy_summary.json`
- `research/price_action_phase3_study.csv`
- `research/price_action_phase3_summary.json`

### Volume Profile Foundation

- `research/volume_profile_features.csv`
- `research/volume_profile_summary.json`
- `research/volume_profile_scoped_features.csv`
- `research/volume_profile_scoped_summary.json`
- `research/volume_profile_shape_study.csv`
- `research/volume_profile_shape_summary.json`
- `research/flexible_volume_profile.csv`
- `research/flexible_volume_profile_summary.json`
- `research/profile_acceptance_quality.csv`
- `research/profile_acceptance_quality_summary.json`

### Setup Studies

- `research/volume_setup_accumulation.csv`
- `research/volume_setup_accumulation_summary.json`
- `research/volume_setup_trend.csv`
- `research/volume_setup_trend_summary.json`
- `research/volume_setup_rejection.csv`
- `research/volume_setup_rejection_summary.json`
- `research/volume_setup_reversal.csv`
- `research/volume_setup_reversal_summary.json`

### Quality Studies

- `research/volume_setup_accumulation_quality.csv`
- `research/volume_setup_accumulation_quality_summary.json`
- `research/volume_setup_accumulation_quality.md`
- `research/volume_setup_trend_quality.csv`
- `research/volume_setup_trend_quality_summary.json`
- `research/volume_setup_trend_quality.md`
- `research/volume_setup_rejection_quality.csv`
- `research/volume_setup_rejection_quality_summary.json`
- `research/volume_setup_rejection_quality.md`

### Cross-Comparison

- `research/volume_setup_comparison.csv`
- `research/volume_setup_comparison.json`
- `research/volume_setup_comparison.md`

### Book Completion Packet

- `docs/volume_profile_playbook.md`
- `docs/volume_profile_risk_playbook.md`
- `docs/volume_profile_backtesting_playbook.md`
- `research/volume_profile_book_completion_packet.json`
- `research/volume_profile_book_completion_packet.md`

## Section 8 - Request

Gemini, please review this as a quantitative research handoff.

Do not propose:

- Live trading.
- Production execution.
- Portfolio optimization.

Focus on:

- Research quality.
- Setup quality.
- Feature engineering.
- Labels.
- ML readiness.
- Statistical weaknesses.

Requested output:

- Strongest findings.
- Weakest findings.
- Recommended next steps.
- Feature engineering roadmap.
- Label design roadmap.
- ML readiness assessment.

Specific review emphasis:

- Whether the current setup counts are sufficient for inference.
- Whether acceptance and passive follow-through should be treated as separate labels.
- Whether Trend's high acceptance but negative FT20 implies bad follow-through labeling, poor setup construction, or a useful short-horizon feature.
- Whether Accumulation's low acceptance but positive FT20 suggests delayed continuation.
- Whether Rejection is too broad and needs stronger filters.
- Whether Reversal should be fixed to populate accepted/rejected outcomes before further comparison.
- Whether L2 context should remain descriptive until historical L2 replay exists.

