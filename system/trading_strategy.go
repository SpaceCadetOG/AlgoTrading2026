package system

import "fmt"

type TradingStrategy struct {
	bookIn    <-chan BookEvent
	orderOut  chan<- OrderIntent
	respIn    <-chan OrderResponse
	nextID    int
	waiting   int
	Position  float64
	Cash      float64
	Realized  float64
	Fills     int
	LastOrder string
}

func NewTradingStrategy(queues *Queues) *TradingStrategy {
	return &TradingStrategy{
		bookIn:   queues.OB2TS,
		orderOut: queues.TS2OM,
		respIn:   queues.OM2TS,
	}
}

func (s *TradingStrategy) ProcessBookEvent() bool {
	select {
	case event := <-s.bookIn:
		if s.waiting > 0 || event.BestBid <= 0 || event.BestAsk <= 0 || event.BestBid <= event.BestAsk {
			return true
		}
		size := minPositive(event.BestBidSize, event.BestAskSize)
		if size <= 0 {
			return true
		}
		s.nextID++
		buyID := fmt.Sprintf("sim-order-%d-buy", s.nextID)
		sellID := fmt.Sprintf("sim-order-%d-sell", s.nextID)
		s.waiting = 2
		s.LastOrder = sellID
		s.orderOut <- OrderIntent{
			ClientID: buyID,
			Action:   ActionNew,
			Side:     Buy,
			Price:    event.BestAsk,
			Size:     size,
		}
		s.orderOut <- OrderIntent{
			ClientID: sellID,
			Action:   ActionNew,
			Side:     Sell,
			Price:    event.BestBid,
			Size:     size,
		}
		return true
	default:
		return false
	}
}

func (s *TradingStrategy) ProcessOrderResponse() bool {
	select {
	case response := <-s.respIn:
		if response.Status == OrderFilled {
			s.Fills++
			switch response.Side {
			case Buy:
				s.Position += response.FillSize
				s.Cash -= response.FillPrice * response.FillSize
			case Sell:
				s.Position -= response.FillSize
				s.Cash += response.FillPrice * response.FillSize
			}
			if s.waiting > 0 {
				s.waiting--
			}
			if s.Position == 0 {
				s.Realized = s.Cash
			}
		}
		if response.Status == OrderRejected || response.Status == OrderCanceled {
			if s.waiting > 0 {
				s.waiting--
			}
		}
		return true
	default:
		return false
	}
}

func minPositive(a float64, b float64) float64 {
	if a <= 0 {
		return b
	}
	if b <= 0 {
		return a
	}
	if a < b {
		return a
	}
	return b
}
