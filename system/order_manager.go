package system

import "fmt"

type ManagedOrder struct {
	ID       string
	ClientID string
	Action   OrderAction
	Side     OrderSide
	Price    float64
	Size     float64
	Status   OrderStatus
}

type OrderManager struct {
	strategyIn  <-chan OrderIntent
	gatewayOut  chan<- OrderIntent
	gatewayIn   <-chan OrderResponse
	strategyOut chan<- OrderResponse
	simIn       <-chan MarketSimulatorEvent

	nextID         int
	Orders         map[string]ManagedOrder
	ClientOrderIDs map[string]string
	Audits         []MarketSimulatorEvent
}

func NewOrderManager(queues *Queues) *OrderManager {
	return &OrderManager{
		strategyIn:     queues.TS2OM,
		gatewayOut:     queues.OM2GW,
		gatewayIn:      queues.GW2OM,
		strategyOut:    queues.OM2TS,
		simIn:          queues.MS2OM,
		Orders:         make(map[string]ManagedOrder),
		ClientOrderIDs: make(map[string]string),
	}
}

func (m *OrderManager) ProcessStrategyOrder() bool {
	select {
	case order := <-m.strategyIn:
		forwarded, response, ok := m.prepareOrder(order)
		if !ok {
			m.strategyOut <- response
			return true
		}
		m.Orders[forwarded.ID] = ManagedOrder{
			ID:       forwarded.ID,
			ClientID: forwarded.ClientID,
			Action:   forwarded.Action,
			Side:     forwarded.Side,
			Price:    forwarded.Price,
			Size:     forwarded.Size,
			Status:   OrderNew,
		}
		m.ClientOrderIDs[forwarded.ClientID] = forwarded.ID
		m.gatewayOut <- forwarded
		return true
	default:
		return false
	}
}

func (m *OrderManager) ProcessMarketResponse() bool {
	select {
	case response := <-m.gatewayIn:
		if order, ok := m.Orders[response.OrderID]; ok {
			order.Status = response.Status
			m.Orders[response.OrderID] = order
		}
		m.strategyOut <- response
		return true
	default:
		return false
	}
}

func (m *OrderManager) prepareOrder(order OrderIntent) (OrderIntent, OrderResponse, bool) {
	if order.Action == "" {
		order.Action = ActionNew
	}

	switch order.Action {
	case ActionNew:
		if order.Price <= 0 || order.Size <= 0 || (order.Side != Buy && order.Side != Sell) {
			return order, OrderResponse{
				ClientID: order.ClientID,
				Status:   OrderRejected,
				Reason:   "invalid new order",
			}, false
		}
		m.nextID++
		order.ID = fmt.Sprintf("om-%d", m.nextID)
		if order.ClientID == "" {
			order.ClientID = order.ID
		}
		return order, OrderResponse{}, true
	case ActionCancel, ActionAmend:
		existing, ok := m.Orders[order.ID]
		if !ok {
			return order, OrderResponse{
				OrderID:  order.ID,
				ClientID: order.ClientID,
				Status:   OrderRejected,
				Reason:   "unknown order",
			}, false
		}
		if existing.Status == OrderFilled || existing.Status == OrderCanceled || existing.Status == OrderRejected {
			return order, OrderResponse{
				OrderID:  order.ID,
				ClientID: existing.ClientID,
				Status:   OrderRejected,
				Reason:   "terminal order state",
			}, false
		}
		if order.ClientID == "" {
			order.ClientID = existing.ClientID
		}
		if order.Side == "" {
			order.Side = existing.Side
		}
		if order.Action == ActionAmend {
			if order.Price <= 0 {
				order.Price = existing.Price
			}
			if order.Size <= 0 {
				order.Size = existing.Size
			}
		}
		return order, OrderResponse{}, true
	default:
		return order, OrderResponse{
			OrderID:  order.ID,
			ClientID: order.ClientID,
			Status:   OrderRejected,
			Reason:   "unsupported action",
		}, false
	}
}

func (m *OrderManager) ProcessSimulatorAudit() bool {
	select {
	case event := <-m.simIn:
		m.Audits = append(m.Audits, event)
		return true
	default:
		return false
	}
}
