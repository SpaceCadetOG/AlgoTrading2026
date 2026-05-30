# Chapter 6 Final Risk Packet

## Best Risk-Adjusted Strategies

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 1 | rsi | KEEP_THROTTLED | THROTTLE | -41.47 | high_variance |
| 2 | bollinger | KEEP_THROTTLED | THROTTLE | -50.88 | low_sharpe, negative_expectancy, poor_risk_grade |
| 3 | mean_reversion | KEEP_THROTTLED | THROTTLE | -52.04 | low_sharpe, negative_expectancy, poor_risk_grade |
| 4 | vol_mean_reversion | KEEP_THROTTLED | THROTTLE | -53.39 | low_sharpe, negative_expectancy, poor_risk_grade |
| 5 | macd | KEEP_THROTTLED | THROTTLE | -60.60 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade |

## Strategies Kept Throttled

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 1 | rsi | KEEP_THROTTLED | THROTTLE | -41.47 | high_variance, top_ranked_throttle |
| 2 | bollinger | KEEP_THROTTLED | THROTTLE | -50.88 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 3 | mean_reversion | KEEP_THROTTLED | THROTTLE | -52.04 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 4 | vol_mean_reversion | KEEP_THROTTLED | THROTTLE | -53.39 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 5 | macd | KEEP_THROTTLED | THROTTLE | -60.60 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, top_ranked_throttle |

## Strategies Requiring Rewrite

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 6 | dual_ma | REWRITE_REQUIRED | THROTTLE | -65.13 | low_sharpe, negative_expectancy, poor_risk_grade, low_risk_adjusted_rank |
| 7 | turtle | REWRITE_REQUIRED | THROTTLE | -66.85 | low_sharpe, negative_expectancy, poor_risk_grade, low_risk_adjusted_rank |
| 8 | apo | REWRITE_REQUIRED | THROTTLE | -68.23 | low_sharpe, negative_expectancy, poor_risk_grade, low_risk_adjusted_rank |
| 9 | sma | REWRITE_REQUIRED | THROTTLE | -69.65 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, low_risk_adjusted_rank |
| 10 | momentum | REWRITE_REQUIRED | THROTTLE | -71.79 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, low_risk_adjusted_rank |
| 11 | vol_trend_following | REWRITE_REQUIRED | THROTTLE | -71.79 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, low_risk_adjusted_rank |
| 12 | support_resistance | REWRITE_REQUIRED | THROTTLE | -73.69 | low_sharpe, high_variance, negative_expectancy, poor_risk_grade, low_risk_adjusted_rank |

## Strategies Removed From Candidates

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 13 | ema | REMOVE_FROM_CANDIDATES | BLOCK | -144.52 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |
| 14 | momentum | REMOVE_FROM_CANDIDATES | BLOCK | -147.84 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |

## Top Risk Issues Observed

- low_sharpe
- negative_expectancy
- excessive_trades_per_day
- high_variance
- poor_risk_grade

## Chapter 6 Conclusion

- No strategy is promoted to active/paper execution yet.
- RSI, Bollinger, mean_reversion, vol_mean_reversion, and dual_ma remain research candidates but throttled.
- EMA and momentum are removed from candidates.
- Risk layer remains rule-based.
- No ML has been introduced.

## Next Recommended Chapter/Build Phase

Chapter 7: controlled paper-trading infrastructure and monitoring gates, without enabling active execution.
