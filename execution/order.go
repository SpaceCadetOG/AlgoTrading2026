package execution

type OrderSide string
type OrderType string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

const (
	Market OrderType = "MARKET"
	Limit  OrderType = "LIMIT"
)

type OrderRequest struct {
	Venue string

	Symbol string

	Side OrderSide
	Type OrderType

	Size  string
	Price string

	ReduceOnly bool
}

type OrderPlacer interface {
	PlaceOrder(order OrderRequest) (*OrderResult, error)
}