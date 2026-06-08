package binance

import (
	"fmt"

	"AlgoTrading2026/tradetape"
)

const PublicArchiveNote = "Binance public data archives are available through https://data.binance.vision/ for daily/monthly market data downloads."

type BackfillStubProvider struct{}

func (BackfillStubProvider) Venue() string { return "binance" }

func (BackfillStubProvider) BackfillTrades(symbol string, startMS int64, endMS int64) ([]tradetape.TradeTapePrint, error) {
	return nil, fmt.Errorf("binance historical archive backfill not implemented yet; %s", PublicArchiveNote)
}
