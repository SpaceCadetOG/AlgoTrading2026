# Volume Setup #1 Accumulation Quality Review

## Summary

- Total setups: 226
- Accepted: 10
- Rejected: 170
- Best confluence group: rolling3d_poc_confluence:with_rolling3d_poc (0.06)
- Best shape: B_PROFILE (0.10)
- Best direction by 20-candle follow-through: long_context (76.80)

## Accepted vs Rejected

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| accepted | 10 | 10 | 0 | 1.00 | 1.00 | 0.00 | -52.25 |
| neutral | 46 | 0 | 0 | 0.00 | 0.83 | 0.04 | 11.52 |
| rejected | 170 | 0 | 170 | 0.00 | 0.84 | 0.01 | 100.73 |

## Long vs Short

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| long_context | 111 | 7 | 83 | 0.06 | 0.88 | 0.00 | 76.80 |
| short_context | 115 | 3 | 87 | 0.03 | 0.80 | 0.03 | 74.84 |

## POC Confluence

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| with_daily_poc | 1 | 0 | 1 | 0.00 | 0.00 | 0.00 | 771.70 |
| without_daily_poc | 225 | 10 | 169 | 0.04 | 0.84 | 0.01 | 72.71 |

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| with_rolling3d_poc | 18 | 1 | 12 | 0.06 | 1.00 | 0.00 | 64.32 |
| without_rolling3d_poc | 208 | 9 | 158 | 0.04 | 0.83 | 0.01 | 76.80 |

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| with_rolling7d_poc | 18 | 1 | 12 | 0.06 | 1.00 | 0.00 | 64.32 |
| without_rolling7d_poc | 208 | 9 | 158 | 0.04 | 0.83 | 0.01 | 76.80 |

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| with_composite30d_poc | 27 | 1 | 24 | 0.04 | 0.89 | 0.00 | 85.78 |
| without_composite30d_poc | 199 | 9 | 146 | 0.05 | 0.83 | 0.02 | 74.45 |

## Shape Quality

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| B_PROFILE | 62 | 6 | 39 | 0.10 | 0.84 | 0.03 | 34.10 |
| D_PROFILE | 58 | 2 | 52 | 0.03 | 0.78 | 0.00 | 76.47 |
| P_PROFILE | 106 | 2 | 79 | 0.02 | 0.88 | 0.01 | 99.83 |

## Retest Quality

| Group | Setups | Accepted | Rejected | Acceptance | Retest | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| not_retested | 36 | 0 | 28 | 0.00 | 0.00 | 0.00 | 471.27 |
| retested | 190 | 10 | 142 | 0.05 | 1.00 | 0.02 | 0.88 |

## What This Means

This study identifies which accumulation-profile attributes are associated with acceptance and passive follow-through. It does not change setup detection and does not simulate execution.

## Recommended Filter Improvements Before Setup #2

- Prefer profile types, shapes, or confluence groups with higher acceptance rates only after manual inspection.
- Compare retested and non-retested follow-through before turning retests into a hard filter.
- Keep the OHLCV approximation limitation in mind until tick/trade volume-at-price is available.
