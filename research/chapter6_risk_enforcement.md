# Chapter 6B/6C Risk Controls Enforcement

- strategies: 14
- trades_before: 1337
- trades_after: 1051
- violations: 406

## Strategy Comparison

| Strategy | Trades Before | Trades After | PnL Before | PnL After | DD Before | DD After | Violations | Still Rejected |
|---|---:|---:|---:|---:|---:|---:|---:|---|
| sma | 164 | 118 | -28.26 | -22.05 | 0.28 | 0.22 | 51 | false |
| ema | 176 | 116 | -27.04 | -16.08 | 0.28 | 0.17 | 69 | true |
| apo | 49 | 49 | -16.44 | -13.54 | 0.18 | 0.15 | 11 | true |
| macd | 100 | 99 | -15.76 | -15.56 | 0.17 | 0.17 | 4 | false |
| bollinger | 59 | 57 | -5.43 | -5.97 | 0.10 | 0.10 | 11 | false |
| rsi | 34 | 35 | -3.72 | -8.18 | 0.07 | 0.08 | 16 | false |
| momentum | 217 | 122 | -30.58 | -18.17 | 0.31 | 0.18 | 104 | true |
| support_resistance | 36 | 35 | -15.22 | -12.38 | 0.16 | 0.14 | 10 | true |
| momentum | 146 | 108 | -27.32 | -20.67 | 0.28 | 0.22 | 44 | true |
| dual_ma | 31 | 31 | -5.84 | -2.44 | 0.09 | 0.07 | 14 | false |
| turtle | 41 | 40 | -10.22 | -8.17 | 0.12 | 0.10 | 5 | false |
| mean_reversion | 69 | 66 | -7.52 | -8.38 | 0.11 | 0.11 | 12 | false |
| vol_mean_reversion | 69 | 67 | -7.94 | -8.66 | 0.12 | 0.12 | 11 | false |
| vol_trend_following | 146 | 108 | -27.32 | -20.67 | 0.28 | 0.22 | 44 | true |

## Strategies Improved

- apo
- bollinger
- dual_ma
- ema
- macd
- momentum
- momentum
- sma
- support_resistance
- turtle
- vol_trend_following

## Strategies Still Rejected

- apo
- ema
- momentum
- momentum
- support_resistance
- vol_trend_following

## Conclusion

- Risk controls are enforced in backtest/research only.
- No live or paper trading is enabled.
- Strategy signal logic is unchanged.
- Next step is continued book-aligned risk and realism research.

## Violations By Rule

- max_hold_bars: 36
- max_trades_per_day: 304
- stop_loss: 56
- volume_participation: 10

## Next Step

Continue book-aligned risk and realism research.
