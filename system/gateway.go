package system

type GatewayLifecycle interface {
	Start() error
	Stop() error
	Status() GatewayStatus
}

type PriceGateway interface {
	GatewayLifecycle
	ReceivePriceUpdate() (LiquidityEvent, bool)
}

type OrderGateway interface {
	GatewayLifecycle
	SendOrder(order OrderIntent) error
	ReceiveOrderResponse() (OrderResponse, bool)
}

type Gateway interface {
	PriceGateway
	OrderGateway
}
