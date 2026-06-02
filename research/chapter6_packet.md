# Chapter 6 Final Risk Packet

## Best Risk-Adjusted Strategies

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 1 | bollinger | KEEP_THROTTLED | THROTTLE | -53.29 | low_sharpe, negative_expectancy, poor_risk_grade |
| 2 | mean_reversion | KEEP_THROTTLED | THROTTLE | -54.87 | low_sharpe, negative_expectancy, poor_risk_grade |
| 3 | vol_mean_reversion | KEEP_THROTTLED | THROTTLE | -55.59 | low_sharpe, negative_expectancy, poor_risk_grade |
| 4 | rsi | KEEP_THROTTLED | THROTTLE | -59.27 | low_sharpe, high_variance, negative_expectancy, poor_risk_grade |
| 5 | dual_ma | KEEP_THROTTLED | THROTTLE | -65.51 | low_sharpe, negative_expectancy, poor_risk_grade |

## Strategies Kept Throttled

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 1 | bollinger | KEEP_THROTTLED | THROTTLE | -53.29 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 2 | mean_reversion | KEEP_THROTTLED | THROTTLE | -54.87 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 3 | vol_mean_reversion | KEEP_THROTTLED | THROTTLE | -55.59 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 4 | rsi | KEEP_THROTTLED | THROTTLE | -59.27 | low_sharpe, high_variance, negative_expectancy, poor_risk_grade, top_ranked_throttle |
| 5 | dual_ma | KEEP_THROTTLED | THROTTLE | -65.51 | low_sharpe, negative_expectancy, poor_risk_grade, top_ranked_throttle |

## Strategies Requiring Rewrite

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 6 | macd | REWRITE_REQUIRED | THROTTLE | -65.89 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, low_risk_adjusted_rank |
| 7 | turtle | REWRITE_REQUIRED | THROTTLE | -68.04 | low_sharpe, negative_expectancy, poor_risk_grade, low_risk_adjusted_rank |
| 8 | sma | REWRITE_REQUIRED | THROTTLE | -72.53 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, low_risk_adjusted_rank |

## Strategies Removed From Candidates

| Rank | Strategy | Gate | Decision | Score | Reasons |
|---:|---|---|---|---:|---|
| 9 | ema | REMOVE_FROM_CANDIDATES | BLOCK | -147.80 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |
| 10 | apo | REMOVE_FROM_CANDIDATES | BLOCK | -148.34 | low_sharpe, negative_expectancy, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |
| 11 | momentum | REMOVE_FROM_CANDIDATES | BLOCK | -149.37 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |
| 12 | vol_trend_following | REMOVE_FROM_CANDIDATES | BLOCK | -149.37 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |
| 13 | momentum | REMOVE_FROM_CANDIDATES | BLOCK | -150.63 | low_sharpe, negative_expectancy, excessive_trades_per_day, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |
| 14 | support_resistance | REMOVE_FROM_CANDIDATES | BLOCK | -154.00 | low_sharpe, negative_expectancy, poor_risk_grade, blocked_by_risk_filter, severe_risk_adjusted_score |

## Top Risk Issues Observed

- low_sharpe
- negative_expectancy
- excessive_trades_per_day
- high_variance
- poor_risk_grade

## Chapter 6 Conclusion

- Chapter 6 remains a research-only risk analysis packet.
- RSI, Bollinger, mean_reversion, vol_mean_reversion, and dual_ma remain research candidates but throttled.
- EMA and momentum are removed from candidates.
- Risk layer remains rule-based.
- No ML has been introduced.

## Next Recommended Chapter/Build Phase

Chapter 7: trading system architecture and simulation components.
