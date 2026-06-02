# Order Book Inventory

This audit checks the current repo before adding any Chapter 4 VWAP/microstructure research features. It is inventory only: no strategy, OMS, gateway, execution, paper/live trading, ML, or risk behavior is changed.

## Requested Files

The requested venue files do not currently exist:

| Requested file | Exists | Current equivalent |
|---|---:|---|
| `exchanges/hyperliquid/orderbook.go` | no | no venue order book implementation found |
| `exchanges/aster/orderbook.go` | no | no venue order book implementation found |
| `exchanges/lighter/orderbook.go` | no | `exchanges/lighter/markets.go` calls `/api/v1/orderBooks` for market metadata, not normalized depth |

## Current Capabilities

| Venue / package | Best bid | Best ask | Full depth | Top N levels | Incremental updates | Snapshots | Notes |
|---|---:|---:|---:|---:|---:|---:|---|
| Aster | no | no | no | no | no | no | `exchanges/aster/ws.go` streams trades only. Candles are available via REST/WS. No depth stream or REST order book endpoint is implemented. |
| Hyperliquid | no | no | no | no | no | no | `exchanges/hyperliquid/ws.go` streams trades only. Candles are available via REST/WS. No L2 book stream or snapshot endpoint is implemented. |
| Lighter | yes, ticker only | yes, ticker only | no | no | no | no | `exchanges/lighter/ws.go` subscribes to `ticker/<market>` and maps bid/ask plus sizes to `exchanges.Ticker`. This is top-of-book ticker data, not a depth book. |
| Lighter market metadata | no | no | no | no | no | metadata snapshot only | `exchanges/lighter/markets.go` calls `/api/v1/orderBooks?filter=all` to infer market precision fields. It stores raw JSON on `LighterMarket.Raw`, but does not normalize book levels. |
| Shared exchange types | yes, ticker type | yes, ticker type | no | no | no | no | `exchanges.Ticker` has `Bid`, `BidSize`, `Ask`, `AskSize`, `Last`, and time. |
| Chapter 7 system simulation | yes | yes | simulated depth only | all simulated inserted levels | insert-only events | in-memory snapshot | `system.OrderBook` maintains sorted bids/asks with FIFO at same price and emits top-of-book changes. This is a simulator component, not venue market data. |

## Normalization Status

Current normalized market data types:

| Type | File | Coverage |
|---|---|---|
| `exchanges.Ticker` | `exchanges/types.go` | Normalizes best bid/ask ticker-style data only. |
| `exchanges.Trade` | `exchanges/types.go` | Normalizes trade prints. Aster and Hyperliquid stream trades; Lighter currently streams ticker. |
| `exchanges.Candle` | `exchanges/candle.go` | Normalizes candles across venues. |
| `system.BookEvent` | `system/types.go` | Simulated top-of-book event for Chapter 7 system path. |
| `system.LiquidityEvent` | `system/types.go` | Simulated bid/ask liquidity insertion for Chapter 7 system path. |

There is no common normalized venue order book structure for real exchange L2 data.

## Existing Calculations

| Calculation | Exists | File | Scope |
|---|---:|---|---|
| Spread | yes | `exchanges/math.go` | `Ticker.Spread()` calculates `ask - bid`. |
| Spread percent | yes | `exchanges/math.go` | `Ticker.SpreadPct()` calculates spread over mid price. |
| Bid/ask USD | yes | `exchanges/math.go` | `Ticker.BidUSD()` and `Ticker.AskUSD()` calculate top-of-book notional from ticker fields. |
| Best bid / best ask from simulated book | yes | `system/order_book.go` | `OrderBook.Snapshot()` returns `system.BookEvent`. |
| Imbalance | no | none found | Not implemented for ticker or depth. |
| Depth calculations | no | none found | No sum-by-level, top-N depth, cumulative depth, or depth notional helpers. |
| Liquidity near price | no | none found | Not implemented. |
| Order book recording/export | no | none found | No L2 snapshot/delta recorder, CSV writer, or replay file found. |

## Venue Notes

### Aster

Implemented market data:

- REST candles: `exchanges/aster/candles.go`
- WS candles: `exchanges/aster/candles_ws.go`
- WS trades: `exchanges/aster/ws.go`

Missing order book support:

- no REST depth snapshot wrapper
- no WebSocket depth stream
- no best bid/ask ticker adapter
- no normalized book levels
- no depth, imbalance, or near-price liquidity calculations

### Hyperliquid

Implemented market data:

- REST candles: `exchanges/hyperliquid/candles.go`
- WS candles: `exchanges/hyperliquid/candles_ws.go`
- WS trades: `exchanges/hyperliquid/ws.go`

Missing order book support:

- no L2 book subscription wrapper
- no book snapshot wrapper
- no normalized book levels
- no depth, imbalance, or near-price liquidity calculations

### Lighter

Implemented market data:

- REST candles: `exchanges/lighter/candles.go`
- WS candles: `exchanges/lighter/candles_ws.go`
- WS ticker: `exchanges/lighter/ws.go`
- market precision metadata from `/api/v1/orderBooks`: `exchanges/lighter/markets.go`

Available L2-like data:

- top bid and ask price/size through ticker messages
- no normalized multi-level depth
- no incremental book deltas
- no snapshot-to-book structure

Important distinction: `GetOrderBookDetails()` uses an endpoint named `orderBooks`, but current code consumes it for market metadata and precision discovery, not as a normalized order book.

## Chapter 7 Simulation Is Separate

The `system.OrderBook` is useful for the book-derived Chapter 7 simulation path:

- sorted bids high to low
- sorted asks low to high
- FIFO at same price by sequence
- emits only top-of-book changes
- returns a top-of-book `BookEvent`

It should not be treated as real venue L2 support. It is insert-only simulated liquidity, does not consume/expire venue liquidity, and is already documented in Chapter 9/10 realism audits as system-validation infrastructure.

## Missing Capabilities

Before building a serious Chapter 4 microstructure research layer, the repo is missing:

- normalized real venue order book schema
- REST snapshot adapters where each venue supports them
- WebSocket depth or L2 subscriptions where each venue supports them
- snapshot/delta sequencing and update application
- best bid/ask derived from real L2 levels
- top-N depth calculations
- cumulative bid/ask depth
- imbalance calculations
- liquidity-near-price calculations
- spread history recording
- order book snapshot/delta CSV export
- replay loader for recorded book data
- stale-book and sequence-gap detection

## Recommended Unified Order Book Schema

For future implementation, add a venue-neutral schema outside strategy logic, likely under `exchanges` or `marketdata`:

```go
type BookSide string

const (
    BookBid BookSide = "bid"
    BookAsk BookSide = "ask"
)

type BookLevel struct {
    Price string
    Size  string
}

type OrderBookSnapshot struct {
    Venue     string
    Symbol    string
    Time      int64
    Sequence  string
    Bids      []BookLevel
    Asks      []BookLevel
    IsSnapshot bool
}

type OrderBookDelta struct {
    Venue    string
    Symbol   string
    Time     int64
    Sequence string
    Bids     []BookLevel
    Asks     []BookLevel
}

type OrderBookProvider interface {
    GetOrderBookSnapshot(symbol string, depth int) (*OrderBookSnapshot, error)
}

type OrderBookStreamer interface {
    StreamOrderBook(symbol string, depth int, handler func(OrderBookDelta)) error
}
```

Recommended helper calculations:

```go
func BestBid(book OrderBookSnapshot) BookLevel
func BestAsk(book OrderBookSnapshot) BookLevel
func Spread(book OrderBookSnapshot) float64
func SpreadPct(book OrderBookSnapshot) float64
func DepthNotional(levels []BookLevel, n int) float64
func Imbalance(book OrderBookSnapshot, n int) float64
func LiquidityNearPrice(book OrderBookSnapshot, center float64, pctBand float64) (bidLiquidity float64, askLiquidity float64)
```

## Recommended Next Step

Do not build Chapter 4 microstructure features directly against current venue adapters yet.

First build a small normalized order book foundation:

1. shared book DTOs and interfaces;
2. pure helper functions for spread, depth, imbalance, and near-price liquidity;
3. CSV export for snapshots;
4. tests using synthetic book snapshots only;
5. no real API calls in tests;
6. no strategy, OMS, gateway, execution, paper/live trading, ML, or risk changes.

After that foundation exists, venue-specific snapshot/stream adapters can be added one at a time and kept disabled/read-only until verified.
