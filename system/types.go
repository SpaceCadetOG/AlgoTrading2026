package system

const (
	QueueLP2Gateway = "lp_2_gateway"
	QueueOB2TS      = "ob_2_ts"
	QueueTS2OM      = "ts_2_om"
	QueueMS2OM      = "ms_2_om"
	QueueOM2TS      = "om_2_ts"
	QueueGW2OM      = "gw_2_om"
	QueueOM2GW      = "om_2_gw"
)

type LiquiditySide string

const (
	LiquidityBid LiquiditySide = "BID"
	LiquidityAsk LiquiditySide = "ASK"
)

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type OrderStatus string

const (
	OrderNew      OrderStatus = "NEW"
	OrderAccepted OrderStatus = "ACCEPTED"
	OrderFilled   OrderStatus = "FILLED"
	OrderCanceled OrderStatus = "CANCELED"
	OrderAmended  OrderStatus = "AMENDED"
	OrderRejected OrderStatus = "REJECTED"
)

type OrderAction string

const (
	ActionNew    OrderAction = "NEW"
	ActionCancel OrderAction = "CANCEL"
	ActionAmend  OrderAction = "AMEND"
)

type LiquidityEvent struct {
	ID    string
	Side  LiquiditySide
	Price float64
	Size  float64
}

type BookEvent struct {
	BestBid     float64
	BestBidSize float64
	BestAsk     float64
	BestAskSize float64
}

type OrderIntent struct {
	ID                string
	ClientID          string
	Action            OrderAction
	Side              OrderSide
	Price             float64
	Size              float64
	ExpectedExitPrice float64
}

type OrderResponse struct {
	OrderID     string
	ClientID    string
	Status      OrderStatus
	Side        OrderSide
	FillPrice   float64
	FillSize    float64
	RealizedPnL float64
	Reason      string
}

type MarketSimulatorEvent struct {
	OrderID string
	Event   string
}

type Queues struct {
	LP2Gateway chan LiquidityEvent
	OB2TS      chan BookEvent
	TS2OM      chan OrderIntent
	MS2OM      chan MarketSimulatorEvent
	OM2TS      chan OrderResponse
	GW2OM      chan OrderResponse
	OM2GW      chan OrderIntent
}

func NewQueues(buffer int) *Queues {
	return &Queues{
		LP2Gateway: make(chan LiquidityEvent, buffer),
		OB2TS:      make(chan BookEvent, buffer),
		TS2OM:      make(chan OrderIntent, buffer),
		MS2OM:      make(chan MarketSimulatorEvent, buffer),
		OM2TS:      make(chan OrderResponse, buffer),
		GW2OM:      make(chan OrderResponse, buffer),
		OM2GW:      make(chan OrderIntent, buffer),
	}
}
