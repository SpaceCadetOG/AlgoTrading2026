package orderflow

import "sort"

func DetectStackedImbalances(imbalances []Imbalance, minAdjacent int) []StackedImbalance {
	if minAdjacent <= 0 {
		minAdjacent = 3
	}
	if len(imbalances) == 0 {
		return nil
	}
	sorted := append([]Imbalance(nil), imbalances...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Price < sorted[j].Price
	})
	out := make([]StackedImbalance, 0)
	start := sorted[0]
	prev := sorted[0]
	count := 1
	flush := func() {
		if count >= minAdjacent {
			out = append(out, StackedImbalance{Direction: start.Direction, StartPrice: start.Price, EndPrice: prev.Price, Count: count})
		}
	}
	for i := 1; i < len(sorted); i++ {
		current := sorted[i]
		if current.Direction == prev.Direction && current.Price > prev.Price {
			prev = current
			count++
			continue
		}
		flush()
		start = current
		prev = current
		count = 1
	}
	flush()
	return out
}
