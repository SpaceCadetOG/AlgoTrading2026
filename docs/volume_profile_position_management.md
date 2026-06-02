# Volume Profile Position Management

## Purpose

Document position-management concepts from the book without creating executable trade rules.

## Targets

Fixed target concepts:

- Prior high or low
- Session open
- Daily open
- Fixed R multiple

Volume-based target concepts:

- POC
- VAH
- VAL
- HVN
- LVN
- VWAP

## Stops

Fixed stop concepts:

- Candle extreme
- Failed auction level
- Rejection high or low

Volume-based stop concepts:

- Beyond VAH or VAL
- Beyond LVN rejection boundary
- Outside accumulation range

## Early Exit Conditions

- Acceptance fails.
- VWAP alignment breaks.
- Price returns into chop.
- Spread widens materially.
- L2 liquidity disappears near planned level.

## Constraint

No stop, target, or exit logic is active in execution.

