package system

type LiquidityProvider struct {
	out chan<- LiquidityEvent
}

func NewLiquidityProvider(queues *Queues) *LiquidityProvider {
	return &LiquidityProvider{out: queues.LP2Gateway}
}

func (p *LiquidityProvider) Insert(event LiquidityEvent) {
	p.out <- event
}

func (p *LiquidityProvider) InsertBid(id string, price float64, size float64) {
	p.Insert(LiquidityEvent{ID: id, Side: LiquidityBid, Price: price, Size: size})
}

func (p *LiquidityProvider) InsertAsk(id string, price float64, size float64) {
	p.Insert(LiquidityEvent{ID: id, Side: LiquidityAsk, Price: price, Size: size})
}
