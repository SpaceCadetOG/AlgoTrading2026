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

type ExecutionRequest struct {
	Decision      ExecutionDecision
	OrderType     string
	RequestedQty  float64
	ExpectedEntry float64
	ReduceOnly    bool
	Protection    ProtectionPlan
}

type OrderSnapshot struct {
	Venue        string  `json:"venue"`
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side,omitempty"`
	OrderID      string  `json:"orderId"`
	ClientID     string  `json:"clientId,omitempty"`
	Status       string  `json:"status"`
	OrderType    string  `json:"orderType,omitempty"`
	RequestedQty float64 `json:"requestedQty,omitempty"`
	FilledQty    float64 `json:"filledQty,omitempty"`
	Price        float64 `json:"price,omitempty"`
	Raw          any     `json:"raw,omitempty"`
}

type FillSnapshot struct {
	Venue     string  `json:"venue"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	OrderID   string  `json:"orderId,omitempty"`
	FillID    string  `json:"fillId,omitempty"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	Fee       float64 `json:"fee,omitempty"`
	Timestamp int64   `json:"timestamp,omitempty"`
	Raw       any     `json:"raw,omitempty"`
}

type PositionSnapshot struct {
	Venue         string  `json:"venue"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Quantity      float64 `json:"quantity"`
	Entry         float64 `json:"entry"`
	UnrealizedPnL float64 `json:"unrealizedPnl,omitempty"`
	Leverage      float64 `json:"leverage,omitempty"`
	Raw           any     `json:"raw,omitempty"`
}

type ReconcileRequest struct {
	Decision ExecutionDecision
	Result   ExecutionResult
}

type ReconcileResult struct {
	Venue      string             `json:"venue"`
	Symbol     string             `json:"symbol"`
	OrderID    string             `json:"orderId,omitempty"`
	Status     string             `json:"status"`
	Source     string             `json:"source,omitempty"`
	Events     []string           `json:"events"`
	Orders     []OrderSnapshot    `json:"orders,omitempty"`
	Fills      []FillSnapshot     `json:"fills,omitempty"`
	Positions  []PositionSnapshot `json:"positions,omitempty"`
	Mismatches []string           `json:"mismatches,omitempty"`
	Message    string             `json:"message,omitempty"`
}

type ProtectionCapabilities struct {
	Venue                    string `json:"venue"`
	NativeBracket            bool   `json:"nativeBracket"`
	NativeStop               bool   `json:"nativeStop"`
	NativeTakeProfit         bool   `json:"nativeTakeProfit"`
	ReduceOnly               bool   `json:"reduceOnly"`
	RuntimeManagedStop       bool   `json:"runtimeManagedStop"`
	RuntimeManagedTakeProfit bool   `json:"runtimeManagedTakeProfit"`
	RuntimeManagedTrailing   bool   `json:"runtimeManagedTrailing"`
	Ready                    bool   `json:"ready"`
	Reason                   string `json:"reason,omitempty"`
}

type ProtectionPlan struct {
	Venue         string `json:"venue"`
	Mode          string `json:"mode"`
	StopArmed     bool   `json:"stopArmed"`
	TPLadderArmed bool   `json:"tpLadderArmed"`
	TrailingArmed bool   `json:"trailingArmed"`
	ReduceOnly    bool   `json:"reduceOnly"`
	Reason        string `json:"reason,omitempty"`
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

type Reconciler interface {
	Reconcile(request ReconcileRequest) (ReconcileResult, error)
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
	Request  ExecutionRequest
	Order    *OrderSnapshot
	Raw      any
}
