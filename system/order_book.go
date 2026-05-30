package system

import "sort"

type bookOrder struct {
	id       string
	side     LiquiditySide
	price    float64
	size     float64
	sequence int64
}

type OrderBook struct {
	in  <-chan LiquidityEvent
	out chan<- BookEvent

	bids     []bookOrder
	asks     []bookOrder
	lastTop  BookEvent
	sequence int64
}

func NewOrderBook(queues *Queues) *OrderBook {
	return &OrderBook{
		in:  queues.LP2Gateway,
		out: queues.OB2TS,
	}
}

func (b *OrderBook) ProcessNext() bool {
	select {
	case event := <-b.in:
		before := b.Snapshot()
		b.apply(event)
		after := b.Snapshot()
		if topChanged(before, after) {
			b.lastTop = after
			b.out <- after
		}
		return true
	default:
		return false
	}
}

func (b *OrderBook) Snapshot() BookEvent {
	event := BookEvent{}
	if len(b.bids) > 0 {
		event.BestBid = b.bids[0].price
		event.BestBidSize = b.bids[0].size
	}
	if len(b.asks) > 0 {
		event.BestAsk = b.asks[0].price
		event.BestAskSize = b.asks[0].size
	}
	return event
}

func (b *OrderBook) Bids() []LiquidityEvent {
	return bookOrdersToLiquidity(b.bids)
}

func (b *OrderBook) Asks() []LiquidityEvent {
	return bookOrdersToLiquidity(b.asks)
}

func (b *OrderBook) apply(event LiquidityEvent) {
	b.sequence++
	order := bookOrder{
		id:       event.ID,
		side:     event.Side,
		price:    event.Price,
		size:     event.Size,
		sequence: b.sequence,
	}
	switch event.Side {
	case LiquidityBid:
		b.bids = append(b.bids, order)
		sort.SliceStable(b.bids, func(i int, j int) bool {
			if b.bids[i].price == b.bids[j].price {
				return b.bids[i].sequence < b.bids[j].sequence
			}
			return b.bids[i].price > b.bids[j].price
		})
	case LiquidityAsk:
		b.asks = append(b.asks, order)
		sort.SliceStable(b.asks, func(i int, j int) bool {
			if b.asks[i].price == b.asks[j].price {
				return b.asks[i].sequence < b.asks[j].sequence
			}
			return b.asks[i].price < b.asks[j].price
		})
	}
}

func topChanged(a BookEvent, b BookEvent) bool {
	return a.BestBid != b.BestBid ||
		a.BestBidSize != b.BestBidSize ||
		a.BestAsk != b.BestAsk ||
		a.BestAskSize != b.BestAskSize
}

func bookOrdersToLiquidity(orders []bookOrder) []LiquidityEvent {
	out := make([]LiquidityEvent, 0, len(orders))
	for _, order := range orders {
		out = append(out, LiquidityEvent{
			ID:    order.id,
			Side:  order.side,
			Price: order.price,
			Size:  order.size,
		})
	}
	return out
}
