package runtime

import (
	"math"
	"sort"
	"strings"

	"AlgoTrading2026/orderbook"
	"AlgoTrading2026/strategy"
)

type PlaybookCandidateBuilder struct {
	Playbooks       []strategy.ExecutablePlaybook
	NotionalUSD     float64
	DefaultLeverage float64
	MinTP1RR        float64
}

func NewPlaybookCandidateBuilder(playbooks []strategy.ExecutablePlaybook) PlaybookCandidateBuilder {
	return PlaybookCandidateBuilder{
		Playbooks:       append([]strategy.ExecutablePlaybook(nil), playbooks...),
		NotionalUSD:     10,
		DefaultLeverage: 1,
		MinTP1RR:        1,
	}
}

func ContextFromOrderBook(snapshot orderbook.OrderBookSnapshot) StrategyContext {
	return StrategyContext{
		Snapshot: MarketSnapshot{
			Venue:     snapshot.Venue,
			Symbol:    snapshot.Symbol,
			OrderBook: snapshot,
		},
		Price:     orderbook.Mid(snapshot),
		SpreadPct: orderbook.SpreadPct(snapshot),
		Liquidity: orderbook.DepthWithinPct(snapshot, 1),
		Imbalance: orderbook.Imbalance(snapshot, 1),
	}
}

func (b PlaybookCandidateBuilder) BuildCandidates(ctx StrategyContext) []Candidate {
	if ctx.Price <= 0 || ctx.Liquidity <= 0 || len(b.Playbooks) == 0 {
		return nil
	}
	notional := b.NotionalUSD
	if notional <= 0 {
		notional = 10
	}
	leverage := b.DefaultLeverage
	if leverage <= 0 {
		leverage = 1
	}

	var out []Candidate
	for _, playbook := range b.Playbooks {
		side, ok := runtimeSide(playbook, ctx)
		if !ok {
			continue
		}
		candidate := bracketedCandidate(playbook, ctx, side, notional, leverage, b.MinTP1RR)
		if candidate.EntryPrice > 0 {
			out = append(out, candidate)
		}
	}
	sort.SliceStable(out, func(i int, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Playbook < out[j].Playbook
		}
		return out[i].Score > out[j].Score
	})
	return out
}

func runtimeSide(playbook strategy.ExecutablePlaybook, ctx StrategyContext) (string, bool) {
	if len(ctx.Snapshot.OrderBook.Bids) == 0 || len(ctx.Snapshot.OrderBook.Asks) == 0 {
		return "", false
	}
	switch strings.ToLower(strings.TrimSpace(playbook.DirectionBias)) {
	case "long":
		return "LONG", ctx.Imbalance >= -0.25
	case "short":
		return "SHORT", ctx.Imbalance <= 0.25
	case "both", "":
		if strings.Contains(strings.ToLower(playbook.Name), "rotation") && math.Abs(ctx.Imbalance) < 0.15 {
			if ctx.Snapshot.OrderBook.Asks[0].Notional() >= ctx.Snapshot.OrderBook.Bids[0].Notional() {
				return "SHORT", true
			}
			return "LONG", true
		}
		if ctx.Imbalance < -0.10 {
			return "SHORT", true
		}
		return "LONG", true
	default:
		return "", false
	}
}

func bracketedCandidate(playbook strategy.ExecutablePlaybook, ctx StrategyContext, side string, notional float64, leverage float64, minRR float64) Candidate {
	entry := ctx.Price
	bufferBps := math.Max(playbook.Stop.BufferBps, 4)
	stopDistance := entry * bufferBps / 10000
	if stopDistance <= 0 {
		stopDistance = entry * 0.003
	}
	rr := math.Max(playbook.Target.MinRR, minRR)
	if rr <= 0 {
		rr = 1
	}
	tp1Distance := stopDistance * rr
	qty := notional / entry
	score := 0.55 + math.Min(math.Abs(ctx.Imbalance), 0.35)
	if strings.Contains(strings.ToLower(playbook.Entry.SourceBook), "vwap") {
		score += 0.03
	}
	if strings.Contains(strings.ToLower(playbook.Entry.SourceBook), "order") {
		score += 0.03
	}
	candidate := Candidate{
		Venue:           ctx.Snapshot.Venue,
		Symbol:          ctx.Snapshot.Symbol,
		CanonicalSymbol: CanonicalFromSymbol(ctx.Snapshot.Symbol),
		Strategy:        playbook.Name,
		Playbook:        playbook.Name,
		Side:            side,
		Score:           math.Min(score, 0.95),
		Confidence:      math.Min(score+0.05, 0.95),
		EntryPrice:      entry,
		Quantity:        qty,
		Leverage:        leverage,
		SpreadPct:       ctx.SpreadPct,
		Liquidity:       ctx.Liquidity,
		RequiredRR:      rr,
		Reasons: []string{
			"runtime_playbook_candidate",
			playbook.Entry.EntryType,
			playbook.Entry.SourceBook,
			"orderbook_priced",
		},
		Provenance: "runtime_playbook:" + playbook.Entry.SourceBook,
	}
	if side == "LONG" {
		candidate.StopPrice = entry - stopDistance
		candidate.TP1 = entry + tp1Distance
		candidate.TP2 = entry + tp1Distance*2
		candidate.TP3 = entry + tp1Distance*3
	} else {
		candidate.StopPrice = entry + stopDistance
		candidate.TP1 = entry - tp1Distance
		candidate.TP2 = entry - tp1Distance*2
		candidate.TP3 = entry - tp1Distance*3
	}
	return candidate
}

func CanonicalFromSymbol(symbol string) string {
	value := strings.ToUpper(strings.TrimSpace(symbol))
	for _, suffix := range []string{"USDT", "USD", "USDC"} {
		if strings.HasSuffix(value, suffix) {
			return strings.TrimSuffix(value, suffix)
		}
	}
	return value
}
