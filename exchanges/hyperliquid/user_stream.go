package hyperliquid

import (
	"context"

	hl "github.com/sonirico/go-hyperliquid"
)

type HLOrderUpdateHandler func([]hl.WsOrder, error)

func (c *Client) StreamOrderUpdates(ctx context.Context, handler HLOrderUpdateHandler) error {
	ws := hl.NewWebsocketClient(getBaseURL())
	if err := ws.Connect(ctx); err != nil {
		return err
	}

	sub, err := ws.OrderUpdates(hl.OrderUpdatesSubscriptionParams{
		User: c.Address,
	}, handler)
	if err != nil {
		_ = ws.Close()
		return err
	}

	go func() {
		<-ctx.Done()
		sub.Close()
		_ = ws.Close()
	}()

	return nil
}
