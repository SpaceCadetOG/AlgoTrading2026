# Volume Profile Setup Comparison

## Executive Summary

- Best acceptance setup: REVERSAL (0.36)
- Best FT20 setup: ACCUMULATION
- Best VWAP setup: TREND
- Best POC setup: REVERSAL
- Best HVN setup: TREND
- Largest sample: REJECTION

## Accumulation

- Setups: 226
- Acceptance rate: 0.04
- Rejection rate: 0.75
- Average FT20: 75.81
- VWAP alignment rate: 0.00
- POC confluence rate: 0.20
- HVN confluence rate: 0.00

## Trend

- Setups: 74
- Acceptance rate: 0.26
- Rejection rate: 0.45
- Average FT20: -67.27
- VWAP alignment rate: 0.64
- POC confluence rate: 0.00
- HVN confluence rate: 0.62

## Rejection

- Setups: 644
- Acceptance rate: 0.08
- Rejection rate: 0.68
- Average FT20: -11.56
- VWAP alignment rate: 0.52
- POC confluence rate: 0.21
- HVN confluence rate: 0.21

## Reversal

- Setups: 121
- Acceptance rate: 0.36
- Rejection rate: 0.64
- Average FT20: -16.45
- VWAP alignment rate: 0.35
- POC confluence rate: 0.66
- HVN confluence rate: 0.21

## Acceptance Ranking

| Rank | Setup | Value |
|---:|---|---:|
| 1 | REVERSAL | 0.36 |
| 2 | TREND | 0.26 |
| 3 | REJECTION | 0.08 |
| 4 | ACCUMULATION | 0.04 |

## Follow Through Ranking

| Rank | Setup | Value |
|---:|---|---:|
| 1 | ACCUMULATION | 75.81 |
| 2 | REJECTION | -11.56 |
| 3 | REVERSAL | -16.45 |
| 4 | TREND | -67.27 |

## Confluence Ranking

| Setup | VWAP | POC | HVN |
|---|---:|---:|---:|
| ACCUMULATION | 0.00 | 0.20 | 0.00 |
| TREND | 0.64 | 0.00 | 0.62 |
| REJECTION | 0.52 | 0.21 | 0.21 |
| REVERSAL | 0.35 | 0.66 | 0.21 |

## Sample Size Ranking

| Rank | Setup | Value |
|---:|---|---:|
| 1 | REJECTION | 644.00 |
| 2 | ACCUMULATION | 226.00 |
| 3 | REVERSAL | 121.00 |
| 4 | TREND | 74.00 |

## Strengths

- Side-by-side setup metrics now make acceptance, follow-through, confluence, and sample size comparable.
- Reversal and rejection studies can be inspected against failed-auction and profile-level context before any executable strategy work.

## Weaknesses

- Follow-through remains passive candle movement, not simulated execution.
- Missing confluence fields stay at zero instead of being inferred.
- Volume Profile still uses OHLCV volume-at-price approximation.

## Recommended Research Focus

- Prioritize setups with non-negative FT20 and enough sample size for manual chart review.
- Treat tiny-sample winners as hypotheses, not conclusions.

## Recommendation Before Pacifica

Finish manual review of the strongest Volume Profile setup filters before adding any venue-specific Pacifica work.

## Recommendation Before ML

Keep these setup comparison outputs as labeled research context, but do not add ML until the book-aligned research layer is stable.
