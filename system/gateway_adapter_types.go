package system

type GatewayStatus string

const (
	GatewayCreated  GatewayStatus = "CREATED"
	GatewayRunning  GatewayStatus = "RUNNING"
	GatewayStopped  GatewayStatus = "STOPPED"
	GatewayDisabled GatewayStatus = "DISABLED"
	GatewayError    GatewayStatus = "ERROR"
)

type GatewayPriceUpdate struct {
	Venue   string
	Symbol  string
	BidID   string
	BidSize float64
	Bid     float64
	AskID   string
	AskSize float64
	Ask     float64
}

type GatewayOrderRequest struct {
	Venue  string
	Symbol string
	Order  OrderIntent
}

type GatewayOrderResponse struct {
	Venue    string
	Symbol   string
	Response OrderResponse
}

func (u GatewayPriceUpdate) ToLiquidityEvents() []LiquidityEvent {
	events := make([]LiquidityEvent, 0, 2)
	if u.Bid > 0 && u.BidSize > 0 {
		id := u.BidID
		if id == "" {
			id = u.Venue + ":" + u.Symbol + ":bid"
		}
		events = append(events, LiquidityEvent{
			ID:    id,
			Side:  LiquidityBid,
			Price: u.Bid,
			Size:  u.BidSize,
		})
	}
	if u.Ask > 0 && u.AskSize > 0 {
		id := u.AskID
		if id == "" {
			id = u.Venue + ":" + u.Symbol + ":ask"
		}
		events = append(events, LiquidityEvent{
			ID:    id,
			Side:  LiquidityAsk,
			Price: u.Ask,
			Size:  u.AskSize,
		})
	}
	return events
}

func (r GatewayOrderRequest) ToOrderIntent() OrderIntent {
	return r.Order
}
