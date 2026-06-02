# Volume Setup #3 Rejection Quality Review

## Summary

- Total setups: 644
- Accepted: 51
- Rejected: 438
- Best direction by 20-candle follow-through: short_context (12.72)
- Best shape by acceptance: UNKNOWN (0.18)
- Best confluence by acceptance: composite30d_poc_confluence:with_composite30d_poc (0.05)
- Best filter by 20-candle follow-through: daily_poc_confluence:with_daily_poc (568.57)

## Accepted vs Rejected

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| accepted | 51 | 51 | 0 | 1.00 | 0.00 | -20.92 |
| neutral | 155 | 0 | 0 | 0.00 | 0.00 | 102.29 |
| rejected | 438 | 0 | 438 | 0.00 | 0.01 | -50.76 |

## Long vs Short

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| long_context | 321 | 18 | 221 | 0.06 | 0.01 | -35.99 |
| short_context | 323 | 33 | 217 | 0.10 | 0.00 | 12.72 |

## VWAP Alignment

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| vwap_aligned | 336 | 31 | 239 | 0.09 | 0.00 | 3.03 |
| vwap_not_aligned | 308 | 20 | 199 | 0.06 | 0.01 | -27.48 |

## POC Retest Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| with_poc_retest | 628 | 48 | 426 | 0.08 | 0.00 | -17.78 |
| without_poc_retest | 16 | 3 | 12 | 0.19 | 0.00 | 232.62 |

## HVN Retest Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| with_hvn_retest | 134 | 8 | 97 | 0.06 | 0.00 | 97.16 |
| without_hvn_retest | 510 | 43 | 341 | 0.08 | 0.01 | -40.13 |

## POC Confluence

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| with_daily_poc | 10 | 0 | 8 | 0.00 | 0.00 | 568.57 |
| without_daily_poc | 634 | 51 | 430 | 0.08 | 0.00 | -20.71 |

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| with_rolling3d_poc | 48 | 2 | 38 | 0.04 | 0.00 | -65.37 |
| without_rolling3d_poc | 596 | 49 | 400 | 0.08 | 0.01 | -7.23 |

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| with_rolling7d_poc | 48 | 2 | 38 | 0.04 | 0.00 | -65.37 |
| without_rolling7d_poc | 596 | 49 | 400 | 0.08 | 0.01 | -7.23 |

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| with_composite30d_poc | 77 | 4 | 53 | 0.05 | 0.00 | -15.38 |
| without_composite30d_poc | 567 | 47 | 385 | 0.08 | 0.01 | -11.04 |

## Profile Shape Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|
| B_PROFILE | 108 | 11 | 71 | 0.10 | 0.01 | 40.21 |
| D_PROFILE | 121 | 8 | 80 | 0.07 | 0.01 | -9.82 |
| P_PROFILE | 331 | 19 | 236 | 0.06 | 0.00 | -29.86 |
| THIN_PROFILE | 29 | 3 | 18 | 0.10 | 0.00 | -17.98 |
| UNKNOWN | 55 | 10 | 33 | 0.18 | 0.00 | -3.59 |

## What This Means

This report identifies which rejection-profile attributes are associated with acceptance and passive follow-through. It does not change the rejection setup detector and does not simulate execution.

## Comparison With Setup #1 and Setup #2

Setup #1 studies accumulated volume after sideways positioning, Setup #2 studies trend-leg continuation, and Setup #3 studies reversal context after strong price rejection. These reports should be compared manually before building a Reversal Trade study.

## Recommendation Before Reversal Trade

- Inspect filters with positive or less-negative follow-through before using rejection setups as reversal candidates.
- Treat high setup count and low acceptance as a warning that raw rejection detection is too broad for execution research.
- Keep L2 context limitations in mind until historical L2 replay is available.
