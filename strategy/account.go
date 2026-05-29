package strategy

import "AlgoTrading2026/exchanges"

type AccountState struct {
	Venue           string
	AccountID       string
	TotalEquity     float64
	AvailableEquity float64
	OpenExposure    float64
	TotalOpenPnL    float64
	PositionCount   int
}

func AccountStateFromSnapshot(snapshot *exchanges.AccountSnapshot) AccountState {
	if snapshot == nil {
		return AccountState{}
	}

	positions := make([]PositionState, 0, len(snapshot.Positions))
	for _, position := range snapshot.Positions {
		positions = append(positions, PositionStateFromExchangePosition(position, nowUTC()))
	}

	return AccountState{
		Venue:           snapshot.Venue,
		AccountID:       snapshot.AccountID,
		TotalEquity:     SafeFloat(snapshot.WalletValue),
		AvailableEquity: SafeFloat(snapshot.Available),
		OpenExposure:    SumExposure(positions),
		TotalOpenPnL:    SafeFloat(snapshot.OpenPnL),
		PositionCount:   len(positions),
	}
}

func AggregateAccountState(accounts []AccountState) AccountState {
	total := AccountState{}
	for _, account := range accounts {
		total.TotalEquity += account.TotalEquity
		total.AvailableEquity += account.AvailableEquity
		total.OpenExposure += account.OpenExposure
		total.TotalOpenPnL += account.TotalOpenPnL
		total.PositionCount += account.PositionCount
	}

	return total
}
