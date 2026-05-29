package strategy

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"AlgoTrading2026/exchanges"
)

type PositionState struct {
	Venue         string
	Symbol        string
	Side          string
	Size          float64
	EntryPrice    float64
	MarkPrice     float64
	UnrealizedPnL float64
	ExposureUSD   float64
	Leverage      float64
	LastUpdate    time.Time
}

type PositionManager struct {
	positions map[string]PositionState
}

func NewPositionManager() *PositionManager {
	return &PositionManager{
		positions: make(map[string]PositionState),
	}
}

func (m *PositionManager) SyncPositions(positions []exchanges.Position) {
	now := time.Now()
	for _, position := range positions {
		state := PositionStateFromExchangePosition(position, now)
		if state.Venue == "" || state.Symbol == "" {
			continue
		}

		m.positions[positionKey(state.Venue, state.Symbol)] = state
	}
}

func (m *PositionManager) GetPosition(venue string, symbol string) (PositionState, bool) {
	position, ok := m.positions[positionKey(venue, symbol)]
	return position, ok
}

func (m *PositionManager) AllPositions() []PositionState {
	positions := make([]PositionState, 0, len(m.positions))
	for _, position := range m.positions {
		positions = append(positions, position)
	}

	sort.Slice(positions, func(i int, j int) bool {
		if positions[i].Venue == positions[j].Venue {
			return positions[i].Symbol < positions[j].Symbol
		}
		return positions[i].Venue < positions[j].Venue
	})

	return positions
}

func PositionStateFromExchangePosition(position exchanges.Position, now time.Time) PositionState {
	size := SafeFloat(position.Size)
	entry := SafeFloat(position.Entry)
	pnl := SafeFloat(position.PnL)
	lev := SafeFloat(position.Lev)

	return PositionState{
		Venue:         position.Venue,
		Symbol:        position.Symbol,
		Side:          NormalizeSide(position.Side, size),
		Size:          size,
		EntryPrice:    entry,
		UnrealizedPnL: pnl,
		ExposureUSD:   ExposureUSD(size, entry),
		Leverage:      lev,
		LastUpdate:    now,
	}
}

func positionKey(venue string, symbol string) string {
	return fmt.Sprintf("%s:%s", strings.ToLower(strings.TrimSpace(venue)), strings.ToUpper(strings.TrimSpace(symbol)))
}

func NormalizeSide(side string, size float64) string {
	upper := strings.ToUpper(strings.TrimSpace(side))
	switch upper {
	case "LONG", "SHORT", "BUY", "SELL":
		if upper == "BUY" {
			return "LONG"
		}
		if upper == "SELL" {
			return "SHORT"
		}
		return upper
	}

	if size > 0 {
		return "LONG"
	}
	if size < 0 {
		return "SHORT"
	}

	return "FLAT"
}
