# Volume Profile Backtesting Playbook

## Progression

### Rough Backtest

Use current research exports to inspect setup frequency, acceptance, rejection, confluence, and passive follow-through. This is not execution.

### Thorough Backtest

Only after manual review, convert a setup into an explicit signal model and run the existing for-loop backtester with fees, slippage, risk controls, and equity curves.

### Micro Trading

Not enabled. This would come after book-aligned research, realism checks, and a future paper-trading shell.

### Half Positions

Not enabled. This belongs to future paper/live operational planning, not the current research layer.

### Full Positions

Not enabled. No live or paper trading path is active.

## Existing Backtesting Artifacts

- `backtest/`
- `research/chapter9_packet.md`
- `research/chapter9_backtester_comparison.md`
- `research/chapter10_realism_audit.md`
- `research/chapter10_data_quality.md`

## Volume Profile Backtest Readiness

The setup studies currently produce context and passive follow-through only. They should not be treated as executable strategies until a separate, explicit strategy research task defines entries, exits, stops, and sizing.

