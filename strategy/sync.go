package strategy

import (
	"time"

	"AlgoTrading2026/exchanges"
)

func SyncSnapshot(manager *PositionManager, snapshot *exchanges.AccountSnapshot) AccountState {
	if snapshot == nil {
		return AccountState{}
	}

	manager.SyncPositions(snapshot.Positions)
	return AccountStateFromSnapshot(snapshot)
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
