# Volume Setup #2 Trend Quality Review

## Summary

- Total setups: 74
- Accepted: 19
- Rejected: 33
- Best direction by 20-candle follow-through: short_context (-12.06)
- Best shape: B_PROFILE (0.33)
- Best trend-strength bucket: weak (0.34)
- Best confluence: vwap_alignment:vwap_aligned (0.34)

## Accepted vs Rejected

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| accepted | 19 | 19 | 0 | 1.00 | 0.00 | 1.96 | -135.18 |
| neutral | 22 | 0 | 0 | 0.00 | 0.00 | 2.68 | -95.02 |
| rejected | 33 | 0 | 33 | 0.00 | 0.18 | 2.14 | -9.66 |

## Long vs Short

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| long_context | 37 | 8 | 17 | 0.22 | 0.08 | 1.94 | -122.48 |
| short_context | 37 | 11 | 16 | 0.30 | 0.08 | 2.56 | -12.06 |

## POC Retest Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| with_poc_retest | 67 | 19 | 27 | 0.28 | 0.09 | 2.27 | -58.56 |
| without_poc_retest | 7 | 0 | 6 | 0.00 | 0.00 | 2.09 | -150.56 |

## HVN Retest Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| with_hvn_retest | 46 | 11 | 19 | 0.24 | 0.11 | 2.32 | -80.28 |
| without_hvn_retest | 28 | 8 | 14 | 0.29 | 0.04 | 2.14 | -45.88 |

## VWAP Alignment Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| vwap_aligned | 47 | 16 | 21 | 0.34 | 0.06 | 2.39 | -26.79 |
| vwap_not_aligned | 27 | 3 | 12 | 0.11 | 0.11 | 2.01 | -137.73 |

## Profile Shape Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| B_PROFILE | 24 | 8 | 8 | 0.33 | 0.04 | 2.10 | -140.18 |
| D_PROFILE | 6 | 2 | 4 | 0.33 | 0.17 | 1.86 | 330.07 |
| P_PROFILE | 39 | 8 | 18 | 0.21 | 0.10 | 2.47 | -106.98 |
| UNKNOWN | 5 | 1 | 3 | 0.20 | 0.00 | 1.78 | 115.74 |

## Trend Strength Quality

| Group | Setups | Accepted | Rejected | Acceptance | Invalidated | Avg Strength | Avg FT20 |
|---|---:|---:|---:|---:|---:|---:|---:|
| medium | 24 | 4 | 14 | 0.17 | 0.04 | 2.02 | 6.44 |
| strong | 21 | 5 | 9 | 0.24 | 0.14 | 3.41 | -243.56 |
| weak | 29 | 10 | 10 | 0.34 | 0.07 | 1.60 | -0.60 |

## What This Means

This review separates trend-leg continuation context by acceptance, retest behavior, VWAP alignment, shape, and strength. It does not change detection rules and does not simulate orders.

## Comparison With Setup #1 Accumulation

Trend setups are evaluated after directional initiation, while Setup #1 accumulation studies the volume level left behind by sideways positioning. The two reports should be compared manually before turning either into a hard filter.

## Recommended Filters Before Setup #3

- Inspect whether POC or HVN retests add acceptance quality before requiring them.
- Prefer shapes and strength buckets with better acceptance only after visual review.
- Keep VWAP/L2 fields contextual because current L2 state is not historical replay.
