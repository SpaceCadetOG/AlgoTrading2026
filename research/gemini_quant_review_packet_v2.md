# Gemini Quant Review Packet V2

## Purpose

This is an updated quantitative research handoff after fixing the Reversal outcome-label pipeline.

The prior Gemini packet was created when Reversal had broken outcome labels:

- old Reversal accepted: 0
- old Reversal rejected: 0

That made Reversal incomparable to Accumulation, Trend, and Rejection.

The labeling issue is now fixed. This packet focuses on the corrected setup comparison.

Scope remains unchanged:

- No live trading.
- No paper trading.
- No strategy execution.
- No OMS changes.
- No gateway changes.
- No ML.
- Documentation and quantitative review only.

## 1. What Changed Since V1

### Reversal Labels Fixed

Reversal setup rows now include explicit outcome labels:

- accepted
- rejected
- neutral

New Reversal stats:

- setups: 118
- accepted: 44
- rejected: 74
- neutral: 0
- averageFollowThrough20: -6.39
- pocConfluence: 80
- hvnConfluence: 26

### Reversal Is Now Comparable

Reversal can now be compared against:

- Accumulation
- Trend
- Rejection

on:

- acceptance rate
- rejection rate
- follow-through
- VWAP confluence
- POC confluence
- HVN confluence
- invalidation

### Cross-Setup Rankings Changed

Updated rankings:

- bestAcceptanceSetup: REVERSAL
- bestFT20Setup: ACCUMULATION
- bestVWAPSetup: TREND
- bestPOCSetup: REVERSAL
- bestHVNSetup: TREND
- largestSampleSetup: REJECTION

## 2. Updated Setup Comparison Table

| Setup | Setups | Accepted | Rejected | Acceptance Rate | Rejection Rate | FT20 | VWAP Confluence | POC Confluence | HVN Confluence |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| Accumulation | 230 | 10 | 173 | 4.35% | 75.22% | 81.83 | 0.00% | 20.00% | 0.00% |
| Trend | 73 | 18 | 33 | 24.66% | 45.21% | -74.14 | 61.64% | 0.00% | 63.01% |
| Rejection | 647 | 50 | 443 | 7.73% | 68.47% | -11.69 | 53.01% | 20.87% | 20.87% |
| Reversal | 118 | 44 | 74 | 37.29% | 62.71% | -6.39 | 37.29% | 67.80% | 22.03% |

## 3. Updated Findings

### Reversal Now Has Best Acceptance

Reversal now has the highest acceptance rate:

- Reversal: 37.29%
- Trend: 24.66%
- Rejection: 7.73%
- Accumulation: 4.35%

This changes the priority discussion. Reversal can no longer be treated as an incomplete or unusable comparison row.

### Accumulation Still Has Best FT20

Accumulation remains the only setup with strongly positive 20-candle passive follow-through:

- Accumulation FT20: 81.83
- Reversal FT20: -6.39
- Rejection FT20: -11.69
- Trend FT20: -74.14

This suggests Accumulation may still be more interesting for delayed continuation even though its acceptance rate is low.

### Trend Still Has Best VWAP/HVN Confluence

Trend remains strongest on:

- VWAP confluence: 61.64%
- HVN confluence: 63.01%

But Trend has the worst FT20 in the comparison:

- Trend FT20: -74.14

This may indicate that Trend is better suited to shorter-horizon behavior, scalp-style labels, exhaustion labels, or a different outcome definition.

### Rejection Remains Broad And Noisy

Rejection has the largest sample:

- Rejection setups: 647

But its outcome quality is weak:

- acceptance rate: 7.73%
- rejection rate: 68.47%
- FT20: -11.69

This still looks too broad. It likely needs stronger filters, spatial exclusion zones, or stricter profile/VWAP/L2 requirements before becoming a useful candidate.

### Acceptance And Follow-Through Still Disagree

The largest finding is not a single winner. It is a label-design conflict:

- Reversal has best acceptance but negative FT20.
- Accumulation has poor acceptance but best FT20.
- Trend has strong confluence but poor FT20.
- Rejection has large sample size but noisy outcomes.

This suggests acceptance/rejection and passive follow-through should probably be treated as separate labels, not one merged quality score.

## 4. Questions For Gemini

Please review the corrected comparison as a quantitative research reviewer.

1. Does corrected Reversal change the setup priority?
2. Is Reversal now the best candidate to refine first?
3. Should Accumulation still be prioritized because of FT20?
4. Should Trend be treated as a short-horizon scalp or exhaustion signal?
5. Should Rejection be tightened with spatial exclusion zones?
6. Which setup should become Strategy Research Packet #1?
7. Which filters should be tested first?
8. Should FT labels be normalized to bps or ATR multiples before the next comparison?
9. What features should be engineered next?
10. What labels should be created next?

Additional specific questions:

- Should acceptance be considered a market-structure label and FT20 a forward-return label?
- Should Reversal acceptance be rechecked using shorter follow-through windows?
- Should Accumulation be studied with delayed-return labels instead of immediate acceptance labels?
- Should Trend be tested with FT5 or FT10 rather than FT20?
- Should Rejection require distance from VWAP, distance from VAH/VAL, or exclusion from mid-profile chop?
- Should POC/HVN confluence be made directional rather than boolean?
- Should confluence be normalized by profile scope: daily, rolling3d, rolling7d, composite30d?

## 5. Explicit Constraints For Gemini

Do not propose:

- Live trading.
- Production execution.
- Paper trading.
- Portfolio optimization.

Focus on:

- Research quality.
- Setup quality.
- Statistical weaknesses.
- Feature engineering.
- Label design.
- ML readiness.

The desired output from Gemini:

- Strongest corrected findings.
- Weakest corrected findings.
- Whether Reversal now deserves priority.
- Whether Accumulation remains the best forward-return candidate.
- Recommended filter tests.
- Recommended feature engineering roadmap.
- Recommended label roadmap.
- ML readiness assessment.

## 6. Current Interpretation To Challenge

A conservative current interpretation is:

```text
Reversal = best acceptance candidate
Accumulation = best FT20 candidate
Trend = best VWAP/HVN confluence candidate
Rejection = best sample-size candidate but too noisy
```

The next research decision should not be "which setup trades best."

The better next question is:

```text
Which label family should each setup be evaluated against?
```

Possible mapping:

- Reversal: acceptance/rejection and short-horizon confirmation labels.
- Accumulation: delayed continuation and FT20 labels.
- Trend: FT5/FT10, exhaustion, or scalp-style labels.
- Rejection: stricter filter discovery before ML.

## 7. Files To Review

Primary corrected files:

- `research/volume_setup_reversal.csv`
- `research/volume_setup_reversal_summary.json`
- `research/volume_setup_reversal_quality.csv`
- `research/volume_setup_reversal_quality_summary.json`
- `research/volume_setup_reversal_quality.md`
- `research/volume_setup_comparison.csv`
- `research/volume_setup_comparison.json`
- `research/volume_setup_comparison.md`

Prior setup files:

- `research/volume_setup_accumulation.csv`
- `research/volume_setup_accumulation_quality.csv`
- `research/volume_setup_accumulation_quality.md`
- `research/volume_setup_trend.csv`
- `research/volume_setup_trend_quality.csv`
- `research/volume_setup_trend_quality.md`
- `research/volume_setup_rejection.csv`
- `research/volume_setup_rejection_quality.csv`
- `research/volume_setup_rejection_quality.md`

Context files:

- `research/flexible_volume_profile.csv`
- `research/profile_acceptance_quality.csv`
- `research/volume_profile_scoped_features.csv`
- `research/volume_profile_shape_study.csv`
- `research/vwap_features.csv`
- `research/context_features.csv`
- `research/context_features_l2.csv`
- `research/vwap_l2_research_review.md`

