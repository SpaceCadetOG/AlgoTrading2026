# Volume Profile Style Playbook

## Purpose

Map Volume Profile research outputs to trading styles without creating executable strategies.

## Intraday

Primary context:

- `daily_session` profiles
- VWAP session features
- daily open
- session open
- failed auction
- latest L2 spread and imbalance

Best suited research questions:

- Does price accept above or below the daily POC?
- Does a failed auction reverse back through VWAP?
- Does the session open become support or resistance?

## Swing

Primary context:

- `rolling_3d`
- `rolling_7d`
- accumulation setups
- trend setups
- rejection quality reports

Best suited research questions:

- Does a broad accumulation POC produce delayed continuation?
- Do rolling profile levels align with VWAP regime?
- Are strong highs/lows being defended across several sessions?

## Long-Term

Primary context:

- `composite_30d`
- POC
- VAH
- VAL
- profile shape
- major HVN/LVN zones

Best suited research questions:

- Is price accepting in a major value area?
- Is the 30-day profile balanced, top-heavy, bottom-heavy, or thin?
- Are broader value levels useful as context for shorter setups?

## Repo Mapping

- Intraday: `research/volume_profile_scoped_features.csv`, `research/vwap_features.csv`, `research/price_action_phase3_study.csv`
- Swing: `research/volume_setup_accumulation.csv`, `research/volume_setup_trend.csv`, `research/volume_setup_comparison.csv`
- Long-term: `research/volume_profile_shape_study.csv`, `research/volume_profile_summary.json`

## Limitation

These are research frames only. No style is activated as a trading system.

