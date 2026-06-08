package orderflow

import "fmt"

func ValidateFootprintBar(bar FootprintBar) error {
	if len(bar.Levels) == 0 {
		return fmt.Errorf("footprint levels must not be empty")
	}
	for _, level := range bar.Levels {
		if level.Price <= 0 {
			return fmt.Errorf("footprint level price must be positive")
		}
		if level.BidVolume < 0 || level.AskVolume < 0 {
			return fmt.Errorf("footprint level volumes must be non-negative")
		}
		if bar.High > 0 && bar.Low > 0 && (level.Price < bar.Low || level.Price > bar.High) {
			return fmt.Errorf("footprint level price outside high/low range")
		}
	}
	return nil
}
