package marketdata

type Router struct {
	Events chan MarketEvent
}

func NewRouter(buffer int) *Router {
	return &Router{
		Events: make(chan MarketEvent, buffer),
	}
}

func (r *Router) Publish(event MarketEvent) {
	r.Events <- event
}