# Volume Setup Reversal Quality Review

## Summary

- Total setups: 121
- Accepted: 44
- Rejected: 77
- Neutral: 0
- Best direction by 20-candle follow-through: short_context (51.95)
- Best confluence by acceptance: poc_confluence:with_poc_confluence (0.36)
- Best filter by 20-candle follow-through: profile_shape:UNKNOWN (303.43)

## Accepted vs Rejected

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| accepted | 44 | 1.00 | 0.00 | 418.69 | 0.00 |
| rejected | 77 | 0.00 | 1.00 | -265.10 | 0.58 |

## Long vs Short

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| long_context | 57 | 0.32 | 0.68 | -93.25 | 0.30 |
| short_context | 64 | 0.41 | 0.59 | 51.95 | 0.44 |

## POC Confluence

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| with_poc_confluence | 80 | 0.36 | 0.64 | -0.99 | 0.36 |
| without_poc_confluence | 41 | 0.37 | 0.63 | -46.61 | 0.39 |

## HVN Confluence

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| with_hvn_confluence | 26 | 0.27 | 0.73 | -33.15 | 0.58 |
| without_hvn_confluence | 95 | 0.39 | 0.61 | -11.88 | 0.32 |

## VAH/VAL Rejection

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| with_vah_val_rejection | 112 | 0.34 | 0.66 | -41.14 | 0.39 |
| without_vah_val_rejection | 9 | 0.67 | 0.33 | 290.76 | 0.11 |

## VWAP Alignment

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| vwap_aligned | 42 | 0.48 | 0.52 | 53.02 | 0.36 |
| vwap_not_aligned | 79 | 0.30 | 0.70 | -53.39 | 0.38 |

## Profile Shape

| Group | Setups | Acceptance | Rejection | Avg FT20 | Invalidation |
|---|---:|---:|---:|---:|---:|
| B_PROFILE | 19 | 0.58 | 0.42 | 86.66 | 0.32 |
| D_PROFILE | 24 | 0.25 | 0.75 | 3.46 | 0.42 |
| P_PROFILE | 66 | 0.33 | 0.67 | -85.75 | 0.41 |
| THIN_PROFILE | 6 | 0.50 | 0.50 | 19.82 | 0.00 |
| UNKNOWN | 6 | 0.33 | 0.67 | 303.43 | 0.33 |

## What This Means

This report fixes the previous zero-state outcome gap by assigning explicit accepted, rejected, or neutral labels to every reversal setup. It does not alter reversal detection, POC/HVN logic, VWAP logic, or any execution behavior.

## Updated Comparison Readiness

Reversal can now be compared against Accumulation, Trend, and Rejection on acceptance, rejection, invalidation, and passive follow-through metrics.
