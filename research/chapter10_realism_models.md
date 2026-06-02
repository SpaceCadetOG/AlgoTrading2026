# Chapter 10B Simulation Realism Models

## Aggregate Risk

- overall_risk: high
- warnings: 6

## Latency Model

- signal_latency: 0s
- strategy_latency: 0s
- gateway_latency: 0s
- exchange_response_latency: 0s
- latency_variance: 0s
- total_expected_latency: 0s
- latency_risk: high

## Place-In-Line Estimate

- order_size: 1.0000
- visible_liquidity: 0.0000
- estimated_queue_position: 0.0000
- fill_probability: 0.0000
- place_in_line_risk: high

## Market Impact Model

- order_size: 1.0000
- average_liquidity: 0.0000
- participation_rate: 0.0000
- estimated_impact_bps: 0.0000
- impact_risk: high

## Fill Assumptions

- fill_ratio: 1.0000
- partial_fill_supported: false
- cancel_amend_supported: true
- stale_book_risk: true
- crossed_book_artificiality: true
- deterministic_fill_warning: true
- fill_risk: high

## Warnings

- latency assumptions are high risk
- place-in-line assumptions are high risk
- market impact assumptions are high risk
- crossed-book artificiality remains enabled in the validation harness
- deterministic full-fill assumption remains high risk
- stale book risk remains high

## Strengths

- latency dimensions are now modeled explicitly
- place-in-line estimation is represented as queue position and fill probability
- market impact is represented by participation rate and estimated impact bps
- fill assumptions explicitly flag deterministic fills, stale book risk, and crossed-book artificiality
- aggregate realism risk is deterministic and testable

## Gaps

- models are not yet applied to production strategy behavior
- models are not yet applied to event-driven fill decisions
- latency model is an assumption model, not calibrated from observed exchange data
- place-in-line model is an estimate, not venue-specific queue reconstruction
- market impact model is a simple participation-rate estimate
- historical market data quality and historical/live parity audit remain pending

## Conclusion

- Chapter 10B models realism assumptions only.
- No live or paper trading is enabled.
- These models are not yet used to change strategy behavior.
- Next phase is Chapter 10C market data quality and historical/live parity audit.

## Next Phase

Chapter 10C market data quality and historical/live parity audit.
