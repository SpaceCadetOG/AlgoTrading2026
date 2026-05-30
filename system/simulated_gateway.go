package system

import "errors"

var ErrGatewayNotRunning = errors.New("gateway is not running")

type SimulatedGateway struct {
	queues       *Queues
	simulator    *MarketSimulator
	priceUpdates chan LiquidityEvent
	status       GatewayStatus
}

func NewSimulatedGateway(queues *Queues) *SimulatedGateway {
	if queues == nil {
		queues = NewQueues(16)
	}
	return &SimulatedGateway{
		queues:       queues,
		simulator:    NewMarketSimulator(queues),
		priceUpdates: make(chan LiquidityEvent, capOrDefault(queues.LP2Gateway, 16)),
		status:       GatewayCreated,
	}
}

func (g *SimulatedGateway) Start() error {
	g.status = GatewayRunning
	return nil
}

func (g *SimulatedGateway) Stop() error {
	g.status = GatewayStopped
	return nil
}

func (g *SimulatedGateway) Status() GatewayStatus {
	return g.status
}

func (g *SimulatedGateway) InjectPriceUpdate(update GatewayPriceUpdate) error {
	if g.status != GatewayRunning {
		return ErrGatewayNotRunning
	}
	for _, event := range update.ToLiquidityEvents() {
		g.priceUpdates <- event
		g.queues.LP2Gateway <- event
	}
	return nil
}

func (g *SimulatedGateway) ReceivePriceUpdate() (LiquidityEvent, bool) {
	select {
	case event := <-g.priceUpdates:
		return event, true
	default:
		return LiquidityEvent{}, false
	}
}

func (g *SimulatedGateway) SendOrder(order OrderIntent) error {
	if g.status != GatewayRunning {
		return ErrGatewayNotRunning
	}
	g.queues.OM2GW <- order
	return nil
}

func (g *SimulatedGateway) ProcessNextOrder() bool {
	return g.simulator.ProcessNext()
}

func (g *SimulatedGateway) FillAllOrders() int {
	return g.simulator.FillAllOrders()
}

func (g *SimulatedGateway) ReceiveOrderResponse() (OrderResponse, bool) {
	select {
	case response := <-g.queues.GW2OM:
		return response, true
	default:
		return OrderResponse{}, false
	}
}

func capOrDefault[T any](ch chan T, fallback int) int {
	if cap(ch) > 0 {
		return cap(ch)
	}
	return fallback
}
