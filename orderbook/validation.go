package orderbook

import "fmt"

func ValidateSnapshot(snapshot OrderBookSnapshot) error {
	if snapshot.Venue == "" {
		return fmt.Errorf("venue is required")
	}
	if snapshot.Symbol == "" {
		return fmt.Errorf("symbol is required")
	}
	if len(snapshot.Bids) == 0 {
		return fmt.Errorf("at least one bid level is required")
	}
	if len(snapshot.Asks) == 0 {
		return fmt.Errorf("at least one ask level is required")
	}

	for i, level := range snapshot.Bids {
		if err := validateLevel("bid", i, level); err != nil {
			return err
		}
	}
	for i, level := range snapshot.Asks {
		if err := validateLevel("ask", i, level); err != nil {
			return err
		}
	}

	bid := BestBid(snapshot).PriceFloat()
	ask := BestAsk(snapshot).PriceFloat()
	if bid >= ask {
		return fmt.Errorf("crossed or locked book: best bid %.12f >= best ask %.12f", bid, ask)
	}

	return nil
}

func validateLevel(side string, index int, level BookLevel) error {
	if level.PriceFloat() <= 0 {
		return fmt.Errorf("%s level %d has invalid price %q", side, index, level.Price)
	}
	if level.SizeFloat() <= 0 {
		return fmt.Errorf("%s level %d has invalid size %q", side, index, level.Size)
	}
	return nil
}
