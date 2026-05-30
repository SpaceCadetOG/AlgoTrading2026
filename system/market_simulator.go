package system

type MarketSimulator struct {
	in       <-chan OrderIntent
	out      chan<- OrderResponse
	auditOut chan<- MarketSimulatorEvent
	orders   map[string]OrderIntent
}

func NewMarketSimulator(queues *Queues) *MarketSimulator {
	return &MarketSimulator{
		in:       queues.OM2GW,
		out:      queues.GW2OM,
		auditOut: queues.MS2OM,
		orders:   make(map[string]OrderIntent),
	}
}

func (s *MarketSimulator) ProcessNext() bool {
	select {
	case order := <-s.in:
		response := s.handle(order)
		s.out <- response
		s.auditOut <- MarketSimulatorEvent{OrderID: order.ID, Event: string(response.Status)}
		return true
	default:
		return false
	}
}

func (s *MarketSimulator) FillAllOrders() int {
	orders := make([]OrderIntent, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}

	for _, order := range orders {
		delete(s.orders, order.ID)
		response := OrderResponse{
			OrderID:   order.ID,
			ClientID:  order.ClientID,
			Status:    OrderFilled,
			Side:      order.Side,
			FillPrice: order.Price,
			FillSize:  order.Size,
			Reason:    "simulated fill",
		}
		s.out <- response
		s.auditOut <- MarketSimulatorEvent{OrderID: order.ID, Event: string(OrderFilled)}
	}
	return len(orders)
}

func (s *MarketSimulator) handle(order OrderIntent) OrderResponse {
	if order.Action == "" {
		order.Action = ActionNew
	}
	switch order.Action {
	case ActionNew:
		if _, exists := s.orders[order.ID]; exists {
			return simulatorReject(order, "duplicate order")
		}
		if order.ID == "" || order.Price <= 0 || order.Size <= 0 || (order.Side != Buy && order.Side != Sell) {
			return simulatorReject(order, "invalid simulated order")
		}
		s.orders[order.ID] = order
		return OrderResponse{
			OrderID:  order.ID,
			ClientID: order.ClientID,
			Status:   OrderAccepted,
			Side:     order.Side,
			Reason:   "simulated accept",
		}
	case ActionCancel:
		existing, exists := s.orders[order.ID]
		if !exists {
			return simulatorReject(order, "cancel unknown order")
		}
		delete(s.orders, order.ID)
		return OrderResponse{
			OrderID:  existing.ID,
			ClientID: existing.ClientID,
			Status:   OrderCanceled,
			Side:     existing.Side,
			Reason:   "simulated cancel",
		}
	case ActionAmend:
		existing, exists := s.orders[order.ID]
		if !exists {
			return simulatorReject(order, "amend unknown order")
		}
		if order.Price <= 0 || order.Size <= 0 {
			return simulatorReject(order, "invalid amend")
		}
		existing.Price = order.Price
		existing.Size = order.Size
		s.orders[order.ID] = existing
		return OrderResponse{
			OrderID:  existing.ID,
			ClientID: existing.ClientID,
			Status:   OrderAmended,
			Side:     existing.Side,
			Reason:   "simulated amend",
		}
	default:
		return simulatorReject(order, "unsupported action")
	}
}

func simulatorReject(order OrderIntent, reason string) OrderResponse {
	return OrderResponse{
		OrderID:  order.ID,
		ClientID: order.ClientID,
		Status:   OrderRejected,
		Side:     order.Side,
		Reason:   reason,
	}
}
