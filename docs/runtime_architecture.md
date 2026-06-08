# Shared Runtime Architecture

The operator entrypoint is `scripts/run_live_logged.sh`. It prompts `Go live? [y/N]:` and maps `y` to `TRADING_MODE=live`; any other answer runs `TRADING_MODE=paper`.

Runtime trading uses one decision path:

`MarketSnapshot -> StrategyContext -> Candidate -> RiskDecision -> ExecutionDecision -> Executor -> PositionManager`

The shared runtime layer lives in `internal/runtime`:

- `MarketSnapshot` and `StrategyContext` normalize venue/order-book data for executable strategies.
- `PlaybookCandidateBuilder` turns executable playbooks into deterministic candidates.
- `LiveExecutor` converts approved execution decisions into venue orders and refuses unless live gates are armed.

Paper mode uses the same runtime candidates and paper risk checks, then routes approved candidates through simulated fills and the paper position manager in `paper.Engine`.

Live mode uses the same universe selection, snapshots, candidate generation, and risk checks. Approved candidates are routed to the live executor. Live execution is refused unless the operator explicitly arms live trading with `LIVE_ENABLE_LIVE_TRADING=true`, marks venue/account readiness, and passes the venue allowlist where configured. The Aster venue adapter still also requires `ENABLE_LIVE_ORDERS=true`.

Executable runtime playbooks currently come from:

- `strategy/volume_profile_rules.go`
- `strategy/orderflow_rules.go`
- `strategy/vwap_rules.go`

Research-only setup detectors under `research`, historical strategy signal generators under `strategies`, and packet/document generation remain separate. They should not place live or paper orders unless bridged into `internal/runtime` through an explicit candidate builder.

Backtest bar-hit semantics are direction-aware in `backtest.EvaluateBarExit`. If stop and target both trade inside the same OHLC bar, the default policy is `stop_first`.
