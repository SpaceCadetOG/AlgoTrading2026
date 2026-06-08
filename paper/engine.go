package paper

import (
	"fmt"
	"math"
	"strings"
	"time"

	"AlgoTrading2026/exchanges/aster"
	"AlgoTrading2026/exchanges/hyperliquid"
	"AlgoTrading2026/exchanges/lighter"
	runtime "AlgoTrading2026/internal/runtime"
	"AlgoTrading2026/l2recorder"
	"AlgoTrading2026/orderbook"
	"AlgoTrading2026/strategy"
	"AlgoTrading2026/tradetape"
)

type Engine struct {
	Config           Config
	State            EngineState
	Playbooks        []strategy.ExecutablePlaybook
	Now              func() time.Time
	SelectUniverse   func(cfg Config) (UniverseSelection, error)
	FetchOrderBook   func(venue string, symbol string) (orderbook.OrderBookSnapshot, error)
	RunTradeRecorder func() error
	RunL2Recorder    func() error
	BuildCandidates  runtime.CandidateBuilder
}

type RuntimeSummary struct {
	Decisions        int
	Approved         int
	Rejected         int
	OpenCount        int
	RecentClosed     int
	RecordersEnabled bool
	UniverseMode     string
	Min24hVolumeUSD  float64
	SelectedSymbols  int
	Venues           int
}

type scanCounts struct {
	Decisions int
	Approved  int
	Rejected  int
}

func NewEngine(cfg Config) (*Engine, error) {
	state, err := LoadState(cfg)
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		Config:           cfg,
		State:            state,
		Playbooks:        strategy.DefaultBookTradeRules(),
		Now:              time.Now,
		SelectUniverse:   SelectUniverse,
		FetchOrderBook:   defaultOrderBookFetcher,
		RunTradeRecorder: defaultTradeTapeRecorderRunner,
		RunL2Recorder:    defaultL2RecorderRunner,
		BuildCandidates: runtime.PlaybookCandidateBuilder{
			Playbooks:       strategy.DefaultBookTradeRules(),
			NotionalUSD:     cfg.InitialBalance * cfg.RiskPerTradePct / 100,
			DefaultLeverage: 1,
			MinTP1RR:        cfg.MinTP1RR,
		},
	}
	engine.State.LiveEnabled = false
	engine.State.ExecutionEnabled = false
	engine.State.PaperEnabled = true
	return engine, nil
}

func (e *Engine) Run() (RuntimeSummary, error) {
	summary := RuntimeSummary{RecordersEnabled: e.Config.EnableRecorders}
	if err := EnsureDataFiles(e.Config); err != nil {
		return summary, err
	}
	now := e.Now().UTC().UnixMilli()
	e.State.RecordersStatus = RecorderStatus{Equity: "writing", Events: "writing", Funding: "writing"}
	e.runRecorders()
	e.revalue()
	if err := AppendEquity(e.Config, now, e.State); err != nil {
		return summary, err
	}
	if err := AppendEvent(e.Config, TelemetryEvent{
		Timestamp: now,
		Type:      "paper_runtime_started",
		Decision:  "system",
		Reasons: []string{
			"dry_run",
			"live_disabled",
			"execution_disabled",
		},
	}); err != nil {
		return summary, err
	}
	universe, err := e.SelectUniverse(e.Config)
	if err != nil {
		return summary, err
	}
	if err := writeUniverseCSV(e.Config.UniverseCSVPath, universe.Selected); err != nil {
		return summary, err
	}
	if err := writeUniverseJSON(e.Config.UniverseJSONPath, universe); err != nil {
		return summary, err
	}
	summary.UniverseMode = universe.Mode
	summary.Min24hVolumeUSD = universe.Min24hVolumeUSD
	summary.SelectedSymbols = len(universe.Selected)
	summary.Venues = uniqueUniverseVenues(universe.Selected)
	for _, note := range universe.Notes {
		e.recordSystemEvent("universe_notice", []string{note})
	}
	counts, err := e.scanOnce(universe.Selected)
	if err != nil {
		return summary, err
	}
	summary.Decisions = counts.Decisions
	summary.Approved = counts.Approved
	summary.Rejected = counts.Rejected
	summary.OpenCount = len(e.State.OpenPositions)
	summary.RecentClosed = len(e.State.RecentClosed)
	return summary, SaveState(e.Config, e.State)
}

func (e *Engine) EvaluateCandidate(candidate Candidate, snapshot orderbook.OrderBookSnapshot, thinSession bool) (RiskDecision, *PaperPosition, error) {
	now := e.Now().UTC().UnixMilli()
	decision := CheckRisk(e.State, candidate, e.Config, now)
	e.recordDecision(candidate, decision, now)
	if !decision.Allowed {
		if err := SaveState(e.Config, e.State); err != nil {
			return decision, nil, err
		}
		return decision, nil, nil
	}

	fill, err := SimulateEntryFill(snapshot, candidate.Side, candidate.Quantity, thinSession, e.Config)
	if err != nil {
		return decision, nil, err
	}
	if fill.FilledQty <= 0 {
		return RiskDecision{Allowed: false, Reasons: []string{"no_fill"}}, nil, nil
	}

	position := PaperPosition{
		ID:              fmt.Sprintf("%s-%d", candidate.Symbol, now),
		Venue:           candidate.Venue,
		Symbol:          candidate.Symbol,
		CanonicalSymbol: candidate.CanonicalSymbol,
		Strategy:        candidate.Strategy,
		Playbook:        candidate.Playbook,
		Side:            normalizeSide(candidate.Side),
		Score:           candidate.Score,
		Confidence:      candidate.Confidence,
		EntryPrice:      fill.AveragePrice,
		MarkPrice:       orderbook.Mid(snapshot),
		LastPrice:       orderbook.Mid(snapshot),
		Quantity:        fill.FilledQty,
		InitialQuantity: fill.FilledQty,
		Margin:          fill.AveragePrice * fill.FilledQty / math.Max(candidate.Leverage, 1),
		Leverage:        math.Max(candidate.Leverage, 1),
		Stop:            candidate.StopPrice,
		TP1:             candidate.TP1,
		TP2:             candidate.TP2,
		TP3:             candidate.TP3,
		BreakEvenPrice:  fill.AveragePrice,
		FeesPaid:        fill.FeePaid,
		OpenedTime:      now,
		UpdatedTime:     now,
		EntryReasons:    append([]string(nil), candidate.Reasons...),
		DecisionReasons: append([]string(nil), decision.Reasons...),
		Provenance:      firstNonEmpty(candidate.Provenance, "paper_runtime"),
	}
	e.State.Balance -= position.Margin + fill.FeePaid
	e.State.Reserve += position.Margin
	e.State.OpenPositions = append(e.State.OpenPositions, position)
	e.State.OpenCount = len(e.State.OpenPositions)
	e.State.DailyTradeCount[candidate.Symbol]++
	e.revalue()
	if err := SaveState(e.Config, e.State); err != nil {
		return decision, nil, err
	}
	if err := AppendEvent(e.Config, TelemetryEvent{
		Timestamp:    now,
		Type:         "open",
		Symbol:       position.Symbol,
		Venue:        position.Venue,
		Side:         position.Side,
		Strategy:     position.Strategy,
		Playbook:     position.Playbook,
		Score:        position.Score,
		Confidence:   position.Confidence,
		Decision:     "approved",
		Entry:        position.EntryPrice,
		Mark:         position.MarkPrice,
		Last:         position.LastPrice,
		Reasons:      candidate.Reasons,
		StopDistance: math.Abs(position.EntryPrice - position.Stop),
	}); err != nil {
		return decision, nil, err
	}
	return decision, &position, nil
}

func (e *Engine) ManagePosition(index int, snapshot orderbook.OrderBookSnapshot, noFollowThrough bool, fundingRate float64, endOfDay bool) error {
	if index < 0 || index >= len(e.State.OpenPositions) {
		return nil
	}
	position := e.State.OpenPositions[index]
	position.MarkPrice = orderbook.Mid(snapshot)
	position.LastPrice = orderbook.Mid(snapshot)
	position.OpenPnL = unrealized(position)
	position.MFER = max(position.MFER, position.OpenPnL)
	position.MAER = min(position.MAER, position.OpenPnL)
	exit := EvaluateExit(position, position.MarkPrice, position.LastPrice, noFollowThrough, FundingHazardBlocked(fundingRate, position.Side, e.Config), endOfDay)
	if exit.Action == "" {
		e.State.OpenPositions[index] = position
		e.revalue()
		return SaveState(e.Config, e.State)
	}
	if exit.UpgradeBreakEven {
		position.Stop = position.EntryPrice
		position.BreakEvenPrice = position.EntryPrice
	}
	if exit.NewTrailingStop != 0 {
		position.TrailingStop = exit.NewTrailingStop
	}
	closeQty := position.Quantity * exit.ClosePct
	fill, err := SimulateExitFill(snapshot, exitSide(position.Side), closeQty, false, exit.Reason == "hard_stop", exit.ForceFlat, e.Config)
	if err != nil {
		return err
	}
	realized := realizedPnL(position, fill.AveragePrice, fill.FilledQty) - fill.FeePaid
	position.RealizedPnL += realized
	position.FeesPaid += fill.FeePaid
	position.Quantity -= fill.FilledQty
	position.UpdatedTime = e.Now().UTC().UnixMilli()
	if exit.Reason == "tp1" {
		position.TP1Taken = true
	}
	if exit.Reason == "tp2" {
		position.TP2Taken = true
	}
	if exit.Reason == "tp3" {
		position.TP3Taken = true
	}
	e.State.Balance += position.Margin*exit.ClosePct + realized
	e.State.Reserve -= position.Margin * exit.ClosePct
	e.State.RealizedPnL += realized
	e.State.RealizedToday += realized
	if err := AppendTrade(e.Config, position.UpdatedTime, position, fill.AveragePrice, fill.FilledQty, realized, exit.Reason); err != nil {
		return err
	}
	if err := AppendEvent(e.Config, TelemetryEvent{
		Timestamp:   position.UpdatedTime,
		Type:        exit.Reason,
		Symbol:      position.Symbol,
		Venue:       position.Venue,
		Side:        position.Side,
		Strategy:    position.Strategy,
		Playbook:    position.Playbook,
		Entry:       position.EntryPrice,
		Exit:        fill.AveragePrice,
		Mark:        position.MarkPrice,
		Last:        position.LastPrice,
		RealizedPnL: realized,
		MFE:         position.MFER,
		MAE:         position.MAER,
		HoldTimeSec: (position.UpdatedTime - position.OpenedTime) / 1000,
	}); err != nil {
		return err
	}
	if position.Quantity <= 0.0000001 || exit.Action == "close" {
		e.State.RecentClosed = append([]PaperPosition{position}, e.State.RecentClosed...)
		if len(e.State.RecentClosed) > 10 {
			e.State.RecentClosed = e.State.RecentClosed[:10]
		}
		e.State.OpenPositions = append(e.State.OpenPositions[:index], e.State.OpenPositions[index+1:]...)
		e.State.SymbolCooldowns[position.Symbol] = position.UpdatedTime + int64(time.Hour/time.Millisecond)
	} else {
		e.State.OpenPositions[index] = position
	}
	e.State.OpenCount = len(e.State.OpenPositions)
	e.revalue()
	if e.State.RealizedToday < 0 && math.Abs(e.State.RealizedToday)/math.Max(e.Config.InitialBalance, 1)*100 >= e.Config.MaxDailyLossPct {
		e.State.LossCooldownUntil = e.Now().UTC().Add(4 * time.Hour).UnixMilli()
	}
	return SaveState(e.Config, e.State)
}

func (e *Engine) ApplyFunding(index int, rate float64) error {
	if !e.Config.EnableFunding || index < 0 || index >= len(e.State.OpenPositions) {
		return nil
	}
	position := e.State.OpenPositions[index]
	position.MarkPrice = max(position.MarkPrice, position.EntryPrice)
	charge := FundingCharge(position, rate)
	position.FundingPaid += -charge
	e.State.Balance += charge
	e.State.RealizedPnL += charge
	e.State.RealizedToday += charge
	e.State.OpenPositions[index] = position
	event := FundingEvent{
		Timestamp:  e.Now().UTC().UnixMilli(),
		PositionID: position.ID,
		Symbol:     position.Symbol,
		Side:       position.Side,
		Rate:       rate,
		Amount:     charge,
	}
	if err := AppendFunding(e.Config, event); err != nil {
		return err
	}
	if err := AppendEvent(e.Config, TelemetryEvent{
		Timestamp:   event.Timestamp,
		Type:        "funding",
		Symbol:      position.Symbol,
		Venue:       position.Venue,
		Side:        position.Side,
		Strategy:    position.Strategy,
		Playbook:    position.Playbook,
		RealizedPnL: charge,
		Reasons:     []string{"funding_applied"},
	}); err != nil {
		return err
	}
	e.revalue()
	return SaveState(e.Config, e.State)
}

func (e *Engine) StatusPayload() StatusPayload {
	return BuildStatusPayload(e.State, e.Config)
}

func (e *Engine) recordDecision(candidate Candidate, decision RiskDecision, now int64) {
	decisionStatus := "rejected"
	if decision.Allowed {
		decisionStatus = "approved"
	}
	event := TelemetryEvent{
		Timestamp:    now,
		Type:         "decision",
		Symbol:       candidate.Symbol,
		Venue:        candidate.Venue,
		Side:         candidate.Side,
		Strategy:     candidate.Strategy,
		Playbook:     candidate.Playbook,
		Score:        candidate.Score,
		Confidence:   candidate.Confidence,
		Decision:     decisionStatus,
		Entry:        candidate.EntryPrice,
		Reasons:      append([]string(nil), decision.Reasons...),
		StopDistance: math.Abs(candidate.EntryPrice - candidate.StopPrice),
	}
	e.State.RecentDecisions = append([]TelemetryEvent{event}, e.State.RecentDecisions...)
	if len(e.State.RecentDecisions) > 25 {
		e.State.RecentDecisions = e.State.RecentDecisions[:25]
	}
	_ = AppendEvent(e.Config, event)
}

func (e *Engine) revalue() {
	openPnL := 0.0
	for i := range e.State.OpenPositions {
		e.State.OpenPositions[i].OpenPnL = unrealized(e.State.OpenPositions[i])
		openPnL += e.State.OpenPositions[i].OpenPnL
	}
	e.State.OpenPnL = openPnL
	e.State.Equity = e.State.Balance + e.State.Reserve + openPnL
}

func unrealized(position PaperPosition) float64 {
	return realizedPnL(position, position.MarkPrice, position.Quantity)
}

func realizedPnL(position PaperPosition, exitPrice float64, qty float64) float64 {
	switch normalizeSide(position.Side) {
	case "LONG":
		return (exitPrice - position.EntryPrice) * qty
	case "SHORT":
		return (position.EntryPrice - exitPrice) * qty
	default:
		return 0
	}
}

func normalizeSide(side string) string {
	switch stringsUpperTrim(side) {
	case "BUY", "LONG":
		return "LONG"
	case "SELL", "SHORT":
		return "SHORT"
	default:
		return stringsUpperTrim(side)
	}
}

func exitSide(side string) string {
	if normalizeSide(side) == "LONG" {
		return "SHORT"
	}
	return "LONG"
}

func (e *Engine) decisionCounts() (int, int, int) {
	decisions := 0
	approved := 0
	rejected := 0
	for _, event := range e.State.RecentDecisions {
		if event.Type != "decision" {
			continue
		}
		decisions++
		switch event.Decision {
		case "approved":
			approved++
		case "rejected":
			rejected++
		}
	}
	return decisions, approved, rejected
}

func (e *Engine) scanOnce(universe []UniverseEntry) (scanCounts, error) {
	counts := scanCounts{}
	snapshots := make([]orderbook.OrderBookSnapshot, 0)
	for _, entry := range universe {
		snapshot, err := e.FetchOrderBook(entry.Venue, entry.Symbol)
		if err != nil {
			e.recordSystemEvent("recorder_error", []string{fmt.Sprintf("orderbook_fetch_failed:%s:%s", entry.Venue, entry.Symbol), err.Error()})
			continue
		}
		snapshots = append(snapshots, snapshot)
	}

	for i := len(e.State.OpenPositions) - 1; i >= 0; i-- {
		snapshot, ok := matchingSnapshot(snapshots, e.State.OpenPositions[i].Venue, e.State.OpenPositions[i].Symbol)
		if !ok {
			continue
		}
		if err := e.ManagePosition(i, snapshot, false, 0, false); err != nil {
			e.recordSystemEvent("manage_error", []string{err.Error()})
		}
	}

	syntheticOpened := false
	for _, snapshot := range snapshots {
		candidates := e.BuildCandidates.BuildCandidates(runtime.ContextFromOrderBook(snapshot))
		if len(candidates) == 0 {
			candidate, ok := e.syntheticCandidate(snapshot, syntheticOpened)
			if !ok {
				e.recordDecision(Candidate{
					Venue: snapshot.Venue, Symbol: snapshot.Symbol, CanonicalSymbol: canonicalFromSymbol(snapshot.Symbol),
					Strategy: "runtime_playbooks", Playbook: "runtime_playbooks", Side: "LONG",
					EntryPrice: orderbook.Mid(snapshot),
				}, RiskDecision{Allowed: false, Reasons: []string{"no_runtime_candidate"}}, e.Now().UTC().UnixMilli())
				counts.Decisions++
				counts.Rejected++
				continue
			}
			candidates = []Candidate{candidate}
		}
		for _, candidate := range candidates {
			decision, _, err := e.EvaluateCandidate(candidate, snapshot, false)
			if err != nil {
				return counts, err
			}
			counts.Decisions++
			if decision.Allowed {
				counts.Approved++
				syntheticOpened = true
			} else {
				counts.Rejected++
			}
		}
	}
	return counts, nil
}

func (e *Engine) syntheticCandidate(snapshot orderbook.OrderBookSnapshot, syntheticAlreadyOpened bool) (Candidate, bool) {
	mid := orderbook.Mid(snapshot)
	if mid <= 0 {
		return Candidate{}, false
	}
	if !e.Config.AllowSyntheticCandidate || syntheticAlreadyOpened {
		return Candidate{}, false
	}
	if len(e.Playbooks) == 0 {
		return Candidate{}, false
	}
	playbook := e.Playbooks[0]
	side := inferredSide(playbook.DirectionBias)
	qty := 10.0 / mid
	if qty <= 0 {
		return Candidate{}, false
	}
	stopDistance := mid * 0.003
	tpDistance := stopDistance * max(e.Config.MinTP1RR, 1.0)
	candidate := Candidate{
		Venue:           snapshot.Venue,
		Symbol:          snapshot.Symbol,
		CanonicalSymbol: canonicalFromSymbol(snapshot.Symbol),
		Strategy:        playbook.Name,
		Playbook:        playbook.Name,
		Side:            side,
		Score:           0.55,
		Confidence:      0.60,
		EntryPrice:      mid,
		Quantity:        qty,
		Leverage:        1,
		SpreadPct:       orderbook.SpreadPct(snapshot),
		Liquidity:       orderbook.DepthWithinPct(snapshot, 1),
		FundingRate:     0,
		RequiredRR:      e.Config.MinTP1RR,
		Reasons:         []string{"synthetic_test_candidate", "live_orderbook_priced", playbook.Name},
		Provenance:      "synthetic_test_candidate",
	}
	if side == "LONG" {
		candidate.StopPrice = mid - stopDistance
		candidate.TP1 = mid + tpDistance
		candidate.TP2 = mid + tpDistance*2
		candidate.TP3 = mid + tpDistance*3
	} else {
		candidate.StopPrice = mid + stopDistance
		candidate.TP1 = mid - tpDistance
		candidate.TP2 = mid - tpDistance*2
		candidate.TP3 = mid - tpDistance*3
	}
	return candidate, true
}

func (e *Engine) runRecorders() {
	if !e.Config.EnableRecorders {
		e.State.RecordersStatus = RecorderStatus{
			TradeTape: "disabled",
			L2:        "disabled",
			Equity:    "writing",
			Events:    "writing",
			Funding:   "writing",
		}
		return
	}
	if err := e.RunTradeRecorder(); err != nil {
		e.State.RecordersStatus.TradeTape = "error"
		e.recordSystemEvent("recorder_error", []string{"trade_tape", err.Error()})
	} else {
		e.State.RecordersStatus.TradeTape = "enabled"
	}
	if err := e.RunL2Recorder(); err != nil {
		e.State.RecordersStatus.L2 = "error"
		e.recordSystemEvent("recorder_error", []string{"l2", err.Error()})
	} else {
		e.State.RecordersStatus.L2 = "enabled"
	}
}

func (e *Engine) recordSystemEvent(eventType string, reasons []string) {
	event := TelemetryEvent{
		Timestamp: e.Now().UTC().UnixMilli(),
		Type:      eventType,
		Decision:  "system",
		Reasons:   append([]string(nil), reasons...),
	}
	e.State.RecentDecisions = append([]TelemetryEvent{event}, e.State.RecentDecisions...)
	if len(e.State.RecentDecisions) > 25 {
		e.State.RecentDecisions = e.State.RecentDecisions[:25]
	}
	_ = AppendEvent(e.Config, event)
}

func defaultOrderBookFetcher(venue string, symbol string) (orderbook.OrderBookSnapshot, error) {
	switch strings.ToLower(strings.TrimSpace(venue)) {
	case "aster":
		return aster.GetMainnetOrderBook(normalizePaperAsterSymbol(symbol))
	case "hyperliquid":
		snapshot, err := hyperliquid.GetMainnetL2OrderBookSnapshot(normalizePaperPerpSymbol(symbol))
		if err != nil {
			return orderbook.OrderBookSnapshot{}, err
		}
		return *snapshot, nil
	case "lighter":
		return lighter.GetMainnetOrderBook(normalizePaperPerpSymbol(symbol))
	default:
		return orderbook.OrderBookSnapshot{}, fmt.Errorf("unsupported venue %q", venue)
	}
}

func defaultTradeTapeRecorderRunner() error {
	cfg := tradetape.DefaultRecorderConfig()
	cfg.RunTradeTapeRecorder = true
	cfg.MaxRounds = 1
	cfg.IntervalSeconds = 0
	recorder := tradetape.NewRecorder(cfg, []tradetape.VenueFetcher{
		{Venue: "hyperliquid", Symbol: "BTC", Fetch: func() ([]tradetape.TradeTapePrint, error) { return hyperliquid.GetRecentTrades("BTC") }},
		{Venue: "aster", Symbol: "BTCUSDT", Fetch: func() ([]tradetape.TradeTapePrint, error) { return aster.GetRecentTrades("BTCUSDT", 50) }},
		{Venue: "lighter", Symbol: "BTC", Fetch: func() ([]tradetape.TradeTapePrint, error) { return lighter.GetRecentTrades("BTC") }},
	})
	_, err := recorder.RecordRound()
	return err
}

func defaultL2RecorderRunner() error {
	cfg := l2recorder.DefaultRecorderConfig()
	cfg.MaxSnapshots = 1
	cfg.IntervalSeconds = 0
	recorder := l2recorder.NewRecorder(cfg, nil, nil)
	_, err := recorder.Run()
	return err
}

func matchingSnapshot(snapshots []orderbook.OrderBookSnapshot, venue string, symbol string) (orderbook.OrderBookSnapshot, bool) {
	for _, snapshot := range snapshots {
		if stringsEqualFold(snapshot.Venue, venue) && stringsEqualFold(snapshot.Symbol, symbol) {
			return snapshot, true
		}
	}
	return orderbook.OrderBookSnapshot{}, false
}

func inferredSide(directionBias string) string {
	switch strings.ToLower(strings.TrimSpace(directionBias)) {
	case "short":
		return "SHORT"
	default:
		return "LONG"
	}
}

func canonicalFromSymbol(symbol string) string {
	value := stringsUpperTrim(symbol)
	for _, suffix := range []string{"USDT", "USD", "USDC"} {
		if strings.HasSuffix(value, suffix) {
			return strings.TrimSuffix(value, suffix)
		}
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func normalizePaperAsterSymbol(symbol string) string {
	value := stringsUpperTrim(symbol)
	if value == "BTC" || value == "ETH" || value == "SOL" {
		return value + "USDT"
	}
	if !strings.HasSuffix(value, "USDT") {
		return value + "USDT"
	}
	return value
}

func normalizePaperPerpSymbol(symbol string) string {
	value := stringsUpperTrim(symbol)
	for _, suffix := range []string{"USDT", "USD", "USDC"} {
		if strings.HasSuffix(value, suffix) {
			return strings.TrimSuffix(value, suffix)
		}
	}
	return value
}
