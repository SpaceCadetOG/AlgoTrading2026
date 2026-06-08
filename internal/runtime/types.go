package runtime

import "AlgoTrading2026/orderbook"

type MarketSnapshot struct {
	Venue     string
	Symbol    string
	OrderBook orderbook.OrderBookSnapshot
}

type StrategyContext struct {
	Snapshot  MarketSnapshot
	Price     float64
	SpreadPct float64
	Liquidity float64
	Imbalance float64
}

type Candidate struct {
	Venue           string
	Symbol          string
	CanonicalSymbol string
	Strategy        string
	Playbook        string
	Side            string
	Score           float64
	Confidence      float64
	EntryPrice      float64
	StopPrice       float64
	TP1             float64
	TP2             float64
	TP3             float64
	Quantity        float64
	Leverage        float64
	SpreadPct       float64
	Liquidity       float64
	FundingRate     float64
	RequiredRR      float64
	Reasons         []string
	Provenance      string
}

type RiskDecision struct {
	Allowed bool
	Reasons []string
}

type ExecutionDecision struct {
	Candidate Candidate
	Risk      RiskDecision
	Mode      string
}

type SnapshotSource interface {
	Snapshot(venue string, symbol string) (MarketSnapshot, error)
}

type CandidateBuilder interface {
	BuildCandidates(ctx StrategyContext) []Candidate
}

type RiskChecker interface {
	Check(candidate Candidate) RiskDecision
}

type Executor interface {
	Execute(decision ExecutionDecision) (ExecutionResult, error)
}

type PositionService interface {
	Manage(ctx StrategyContext) error
}

type ExecutionResult struct {
	Accepted bool
	Venue    string
	Symbol   string
	OrderID  string
	Status   string
	Message  string
	Raw      any
}
