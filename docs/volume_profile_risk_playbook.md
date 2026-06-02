# Volume Profile Risk And Position Management Playbook

## Position Management Concepts

Profit target concepts:

- Prior POC, VAH, VAL, HVN, LVN.
- Prior daily or weekly high/low.
- VWAP or anchored VWAP.
- Opposite side of a value area.

Stop placement concepts:

- Beyond rejection high/low.
- Beyond VAH/VAL when the setup depends on value-area defense.
- Beyond failed-auction level when reversal context fails.
- Beyond accumulation range when accumulation is the setup thesis.

Early exit conditions:

- Acceptance fails where rejection was expected.
- Price returns into a rejected value area.
- VWAP alignment breaks against the setup.
- L2 context deteriorates, such as widening spread or opposing imbalance.

## Money Management

Research notes only:

- Risk per trade should be fixed before any future execution work.
- R:R should be measured against profile-defined targets and invalidation levels.
- Position sizing should account for volatility, spread, and liquidity.
- Correlated setups across venues or symbols should not be treated as independent exposure.

## Psychology Notes

Good winner:

- Followed rules and accepted planned target or trailing logic.

Bad winner:

- Profitable only because rules were ignored.

Good loser:

- Followed invalidation logic and kept risk controlled.

Bad loser:

- Ignored stop, widened risk, or added outside the plan.

Rule-breaking prevention:

- Keep setup labels separate from execution.
- Require manual review before turning any research setup into a simulated strategy.
- Preserve logs and research exports for post-study review.

## Existing Risk Artifacts

- `riskmetrics/`
- `risk/`
- `research/chapter6_risk_metrics.csv`
- `research/chapter6_risk_enforcement.csv`
- `research/chapter10_realism_audit.md`

No new risk rules are added by this playbook.

