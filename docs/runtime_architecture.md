# Shared Runtime Architecture

The operator entrypoint is `scripts/run_live_logged.sh`. It prompts `Go live? [y/N]:` and maps `y` to `TRADING_MODE=live`; any other answer runs `TRADING_MODE=paper`.

Windows operators can use `scripts/run_live_logged.ps1`, which exposes the same prompt without requiring Bash. The Bash launcher also falls back to the default Windows Go install path when Git Bash does not inherit `go` on `PATH`.

Runtime trading uses one decision path:

`MarketSnapshot -> StrategyContext -> Candidate -> RiskDecision -> ExecutionDecision -> Executor -> PositionManager`

The shared runtime layer lives in `internal/runtime`:

- `MarketSnapshot` and `StrategyContext` normalize venue/order-book data for executable strategies.
- `PlaybookCandidateBuilder` turns executable playbooks into deterministic candidates.
- `LiveExecutor` converts approved execution decisions into venue orders and refuses unless live gates are armed.

Paper mode uses the same runtime candidates and paper risk checks, then routes approved candidates through simulated fills and the paper position manager in `paper.Engine`.

Live mode uses the same universe selection, snapshots, candidate generation, and risk checks. Approved candidates are routed to the live executor. Live execution is refused unless the operator explicitly selects live mode, arms live trading with `LIVE_ENABLE_LIVE_TRADING=true`, arms real execution with `ENABLE_LIVE_ORDERS=true`, marks venue/account readiness, leaves the kill switch off, and passes the venue allowlist where configured.

`paper.Engine` exposes `RunOnce` for tests/debug and `RunLoop` for operator runtime. `TRADING_MODE=paper go run ./cmd/main.go` now enters the loop, prints one cycle summary per scan, writes heartbeat/scanner telemetry, and continues until interrupted or until `PAPER_MAX_RUNTIME_CYCLES` is set for a bounded run.

Executable runtime playbooks currently come from:

- `strategy/volume_profile_rules.go`
- `strategy/orderflow_rules.go`
- `strategy/vwap_rules.go`

Research-only setup detectors under `research`, historical strategy signal generators under `strategies`, and packet/document generation remain separate. They should not place live or paper orders unless bridged into `internal/runtime` through an explicit candidate builder.

Backtest bar-hit semantics are direction-aware in `backtest.EvaluateBarExit`. If stop and target both trade inside the same OHLC bar, the default policy is `stop_first`.

Structured runtime events are written to the existing paper event JSONL sink. The runtime emits scanner snapshots, candidate creation, risk decisions, execution requests, order results, position lifecycle updates, stop/trailing updates, scanner cycle summaries, venue health, reconciliation placeholders, and heartbeats.

Remaining gaps before production-grade real-money deployment:

- Live protective exit order placement and fill reconciliation need venue-specific hardening.
- Live position state should be reconciled continuously from the venue instead of relying on submitted order responses.
- Runtime playbook mappings are deterministic and executable, but still intentionally simple compared with the full research text.
- Backtest/replay parity has direction-aware bar-hit semantics, but is not yet a tick-level simulator.
- Kill-switch/account/venue health inputs are environment-driven and should eventually be backed by active health checks.
