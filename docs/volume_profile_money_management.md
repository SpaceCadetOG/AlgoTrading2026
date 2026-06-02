# Volume Profile Money Management

## Risk Per Trade

Risk per trade should be defined before any executable strategy exists. Current research does not activate position sizing.

## R:R

Reward-to-risk should be estimated from profile-defined target and invalidation levels:

- target: POC, VAH, VAL, HVN, LVN, prior high/low
- invalidation: failed auction level, rejection high/low, range boundary

## Position Sizing

Future sizing research should account for:

- volatility
- spread
- depth
- liquidity near price
- correlation
- maximum notional

## Correlation And Excessive Exposure

Multiple crypto instruments may share the same beta. A basket of BTC, ETH, SOL, and major alts can still be one directional exposure.

## Current Status

Money management is documented only. Risk controls exist for backtest research, not live execution.

