package exchanges

type AccountSnapshot struct {
	Venue       string
	AccountID   string
	WalletValue string
	Available   string
	OpenPnL     string
	Positions   []Position
	Balances    []Balance
}

type Balance struct {
	Asset     string
	Total     string
	Available string
}

type Position struct {
	Venue  string
	Symbol string
	Side   string
	Size   string
	Entry  string
	PnL    string
	Lev    string
}

type Ticker struct {
	Venue  string
	Symbol string

	Bid     string
	BidSize string
	Ask     string
	AskSize string
	Last    string

	Time int64
}

type Trade struct {
	Venue  string
	Symbol string
	Side   string
	Price  string
	Size   string
	Time   int64
}

type TickerHandler func(Ticker)

type TradeHandler func(Trade)

type AccountReader interface {
	GetAccountSnapshot() (*AccountSnapshot, error)
}

type PositionProvider interface {
	GetPositions() ([]Position, error)
}

type AccountProvider interface {
	GetAccountSnapshot() (*AccountSnapshot, error)
}

type BalanceProvider interface {
	GetBalances() ([]Balance, error)
}

type TickerStreamer interface {
	StreamTicker(symbol string, handler TickerHandler) error
}

type TradeStreamer interface {
	StreamTrades(symbol string, handler TradeHandler) error
}

type Venue interface {
	Name() string
}
