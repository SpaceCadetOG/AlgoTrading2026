package paper

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
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
	FetchMarketBook  func(venue string, symbol string, marketID int) (orderbook.OrderBookSnapshot, error)
	RunTradeRecorder func() error
	RunL2Recorder    func() error
	BuildCandidates  runtime.CandidateBuilder
}

type RuntimeSummary struct {
	Decisions                int
	Approved                 int
	Rejected                 int
	OpenCount                int
	RecentClosed             int
	RecordersEnabled         bool
	UniverseMode             string
	Min24hVolumeUSD          float64
	SelectedSymbols          int
	Venues                   int
	SelectedUniverse         []UniverseEntry
	QualifiedSkipped         []QualifiedNotSelected
	UniverseBiasDiagnostics  UniverseBiasDiagnostics
	DiscoveryStats           []VenueDiscoveryStats
	QualificationDiagnostics []QualificationDiagnostics
	RejectReasons            map[string]int
	ApprovedBlocked          []CandidateOutcome
	CandidateAdmission       []CandidateAdmissionSummary
	CandidateCoverage        []CandidateCoverageRow
	StateBlockers            StateBlockers
	Mode                     string
	ExecutionMode            string
	LiveEnabled              bool
}

type StateBlockers struct {
	Active              bool `json:"active"`
	CooldownSymbols     int  `json:"cooldownSymbols"`
	SymbolsAtDailyMax   int  `json:"symbolsAtDailyMax"`
	LossCooldownActive  bool `json:"lossCooldownActive"`
	OpenPositionSlots   int  `json:"openPositionSlots"`
	MaxOpenPositions    int  `json:"maxOpenPositions"`
	OpenPositionBlocked bool `json:"openPositionBlocked"`
}

type CandidateOutcome struct {
	Candidate Candidate `json:"candidate"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason"`
}

type CandidateAdmissionSummary struct {
	Venue            string `json:"venue"`
	Created          int    `json:"created"`
	Approved         int    `json:"approved"`
	Deduped          int    `json:"deduped"`
	PortfolioBlocked int    `json:"portfolioBlocked"`
	RiskRejected     int    `json:"riskRejected"`
	StateBlocked     int    `json:"stateBlocked"`
}

type CandidateCoverageRow struct {
	Venue             string         `json:"venue"`
	Symbol            string         `json:"symbol"`
	CanonicalSymbol   string         `json:"canonicalSymbol"`
	Major             bool           `json:"major"`
	SnapshotFetched   bool           `json:"snapshotFetched"`
	Candidates        int            `json:"candidates"`
	Approved          int            `json:"approved"`
	Rejected          int            `json:"rejected"`
	ConfidenceSum     float64        `json:"confidenceSum,omitempty"`
	AverageConfidence float64        `json:"averageConfidence,omitempty"`
	RejectReasons     map[string]int `json:"rejectReasons,omitempty"`
}

type CandidateQualitySummary struct {
	Symbols                   int            `json:"symbols"`
	MajorSymbols              int            `json:"majorSymbols"`
	NonMajorSymbols           int            `json:"nonMajorSymbols"`
	Candidates                int            `json:"candidates"`
	Approved                  int            `json:"approved"`
	Rejected                  int            `json:"rejected"`
	NonMajorCandidates        int            `json:"nonMajorCandidates"`
	NonMajorApproved          int            `json:"nonMajorApproved"`
	AverageConfidence         float64        `json:"averageConfidence,omitempty"`
	NonMajorAverageConfidence float64        `json:"nonMajorAverageConfidence,omitempty"`
	QualifiedToCandidatePct   float64        `json:"qualifiedToCandidatePct"`
	CandidateToApprovedPct    float64        `json:"candidateToApprovedPct"`
	TopNonMajorRejects        map[string]int `json:"topNonMajorRejects,omitempty"`
}

type CandidateQualityGroup struct {
	Key               string         `json:"key"`
	Candidates        int            `json:"candidates"`
	Approved          int            `json:"approved"`
	Rejected          int            `json:"rejected"`
	AverageConfidence float64        `json:"averageConfidence,omitempty"`
	ConfidenceSum     float64        `json:"confidenceSum,omitempty"`
	RejectRate        float64        `json:"rejectRate"`
	Reasons           map[string]int `json:"reasons,omitempty"`
}

type RuntimeLoopSnapshot struct {
	Cycle   int
	Summary RuntimeSummary
}

type scanCounts struct {
	Decisions       int
	Approved        int
	Rejected        int
	RejectReasons   map[string]int
	ApprovedBlocked []CandidateOutcome
	Admission       map[string]*CandidateAdmissionSummary
	Coverage        map[string]*CandidateCoverageRow
}

func NewEngine(cfg Config) (*Engine, error) {
	state, err := LoadState(cfg)
	if err != nil {
		return nil, err
	}
	applyResetFlags(&state, cfg)
	engine := &Engine{
		Config:           cfg,
		State:            state,
		Playbooks:        strategy.DefaultBookTradeRules(),
		Now:              time.Now,
		SelectUniverse:   SelectUniverse,
		FetchOrderBook:   defaultOrderBookFetcher,
		FetchMarketBook:  defaultMarketOrderBookFetcher,
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

func applyResetFlags(state *EngineState, cfg Config) {
	if cfg.ResetAll {
		*state = NewState(cfg)
		return
	}
	if cfg.ResetState || cfg.ResetCooldowns {
		state.SymbolCooldowns = map[string]int64{}
		state.SymbolLocks = map[string]int64{}
		state.LossCooldownUntil = 0
	}
	if cfg.ResetState || cfg.ResetDailyCounts {
		state.DailyTradeCount = map[string]int{}
		state.TradeBudgetUsed = 0
		state.RealizedToday = 0
	}
	if cfg.ResetState {
		state.RecentDecisions = nil
		state.RecentClosed = nil
	}
	if cfg.ResetOpenPositions {
		state.OpenPositions = nil
		state.OpenCount = 0
		state.Reserve = 0
	}
	if cfg.ResetRecentClosed {
		state.RecentClosed = nil
	}
}

func (e *Engine) Run() (RuntimeSummary, error) {
	return e.RunOnce()
}

func (e *Engine) RunOnce() (RuntimeSummary, error) {
	summary := RuntimeSummary{
		RecordersEnabled: e.Config.EnableRecorders,
		Mode:             firstNonEmpty(e.Config.Mode, "paper"),
		ExecutionMode:    "simulated",
		LiveEnabled:      e.State.LiveEnabled,
	}
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
		Type:      "MODE_SELECTED",
		Decision:  "system",
		Reasons: []string{
			"mode=paper",
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
	summary.SelectedUniverse = append([]UniverseEntry(nil), universe.Selected...)
	summary.QualifiedSkipped = append([]QualifiedNotSelected(nil), universe.QualifiedSkipped...)
	summary.UniverseBiasDiagnostics = universe.BiasDiagnostics
	summary.DiscoveryStats = append([]VenueDiscoveryStats(nil), universe.Stats...)
	summary.QualificationDiagnostics = append([]QualificationDiagnostics(nil), universe.Diagnostics...)
	_ = AppendStrategyReviewSelected(e.Config, summary.SelectedUniverse, now)
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
	summary.RejectReasons = counts.RejectReasons
	summary.ApprovedBlocked = append([]CandidateOutcome(nil), counts.ApprovedBlocked...)
	summary.CandidateAdmission = admissionSummaryRows(counts.Admission)
	summary.CandidateCoverage = candidateCoverageRows(counts.Coverage)
	summary.OpenCount = len(e.State.OpenPositions)
	summary.RecentClosed = len(e.State.RecentClosed)
	summary.StateBlockers = BuildStateBlockers(e.State, e.Config, e.Now().UTC().UnixMilli())
	if err := WriteRankingJSON(e.Config, summary, e.State); err != nil {
		return summary, err
	}
	if err := e.recordScannerCycleSummary(summary); err != nil {
		return summary, err
	}
	if err := WriteRuntimeSummaryJSON(e.Config, summary, e.State); err != nil {
		return summary, err
	}
	if err := WriteRejectSummaryJSON(e.Config, summary.RejectReasons); err != nil {
		return summary, err
	}
	if err := WriteCandidateCoverageJSON(e.Config, summary.CandidateCoverage); err != nil {
		return summary, err
	}
	if err := WriteCandidateQualityJSON(e.Config, summary.CandidateCoverage); err != nil {
		return summary, err
	}
	if err := WriteOpportunityCoverageJSON(e.Config, summary); err != nil {
		return summary, err
	}
	if err := AppendStrategyReviewPositions(e.Config, e.State.OpenPositions, e.Now().UTC().UnixMilli()); err != nil {
		return summary, err
	}
	if err := AppendStrategyReviewCycle(e.Config, summary, e.Now().UTC().UnixMilli()); err != nil {
		return summary, err
	}
	if err := WriteOvernightStrategySummary(e.Config, summary); err != nil {
		return summary, err
	}
	return summary, SaveState(e.Config, e.State)
}

func (e *Engine) RunLoop(ctx context.Context, onCycle func(RuntimeLoopSnapshot)) error {
	interval := time.Duration(e.Config.ScanIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	maxCycles := e.Config.MaxRuntimeCycles
	cycle := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		cycle++
		summary, err := e.RunOnce()
		if err != nil {
			return err
		}
		if onCycle != nil {
			onCycle(RuntimeLoopSnapshot{Cycle: cycle, Summary: summary})
		}
		if err := AppendEvent(e.Config, TelemetryEvent{
			Timestamp: e.Now().UTC().UnixMilli(),
			Type:      "HEARTBEAT",
			Decision:  "system",
			Reasons: []string{
				fmt.Sprintf("cycle=%d", cycle),
				fmt.Sprintf("decisions=%d", summary.Decisions),
				fmt.Sprintf("approved=%d", summary.Approved),
				fmt.Sprintf("openPositions=%d", summary.OpenCount),
			},
		}); err != nil {
			return err
		}
		if maxCycles > 0 && cycle >= maxCycles {
			return nil
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (e *Engine) EvaluateCandidate(candidate Candidate, snapshot orderbook.OrderBookSnapshot, thinSession bool) (RiskDecision, *PaperPosition, error) {
	now := e.Now().UTC().UnixMilli()
	decision := CheckRisk(e.State, candidate, e.Config, now)
	e.recordDecision(candidate, decision, now)
	if !decision.Allowed {
		e.recordCandidateEvent("RISK_REJECTED", candidate, now, decision.Reasons)
		if err := SaveState(e.Config, e.State); err != nil {
			return decision, nil, err
		}
		return decision, nil, nil
	}
	e.recordCandidateEvent("RISK_APPROVED", candidate, now, decision.Reasons)

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
	e.State.DailyTradeCount[candidateStateKey(candidate)]++
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
	if err := AppendEvent(e.Config, TelemetryEvent{
		Timestamp:    now,
		Type:         "POSITION_OPENED",
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
		_ = AppendEvent(e.Config, TelemetryEvent{
			Timestamp: e.Now().UTC().UnixMilli(),
			Type:      "POSITION_UPDATED",
			Symbol:    position.Symbol,
			Venue:     position.Venue,
			Side:      position.Side,
			Strategy:  position.Strategy,
			Playbook:  position.Playbook,
			Mark:      position.MarkPrice,
			Last:      position.LastPrice,
			Reasons:   []string{"mark_to_market"},
		})
		return SaveState(e.Config, e.State)
	}
	if exit.UpgradeBreakEven {
		position.Stop = position.EntryPrice
		position.BreakEvenPrice = position.EntryPrice
		_ = AppendEvent(e.Config, TelemetryEvent{
			Timestamp: e.Now().UTC().UnixMilli(),
			Type:      "STOP_MOVED",
			Symbol:    position.Symbol,
			Venue:     position.Venue,
			Side:      position.Side,
			Strategy:  position.Strategy,
			Playbook:  position.Playbook,
			Entry:     position.EntryPrice,
			Reasons:   []string{"break_even"},
		})
	}
	if exit.NewTrailingStop != 0 {
		position.TrailingStop = exit.NewTrailingStop
		_ = AppendEvent(e.Config, TelemetryEvent{
			Timestamp: e.Now().UTC().UnixMilli(),
			Type:      "TRAILING_UPDATED",
			Symbol:    position.Symbol,
			Venue:     position.Venue,
			Side:      position.Side,
			Strategy:  position.Strategy,
			Playbook:  position.Playbook,
			Exit:      position.TrailingStop,
			Reasons:   []string{"tp2_trailing_stop"},
		})
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
		position.ClosedTime = position.UpdatedTime
		position.ExitReason = exit.Reason
		e.State.RecentClosed = append([]PaperPosition{position}, e.State.RecentClosed...)
		if len(e.State.RecentClosed) > 10 {
			e.State.RecentClosed = e.State.RecentClosed[:10]
		}
		e.State.OpenPositions = append(e.State.OpenPositions[:index], e.State.OpenPositions[index+1:]...)
		e.State.SymbolCooldowns[positionStateKey(position)] = position.UpdatedTime + int64(time.Hour/time.Millisecond)
		_ = AppendEvent(e.Config, TelemetryEvent{
			Timestamp:   position.UpdatedTime,
			Type:        "POSITION_CLOSED",
			Symbol:      position.Symbol,
			Venue:       position.Venue,
			Side:        position.Side,
			Strategy:    position.Strategy,
			Playbook:    position.Playbook,
			Exit:        fill.AveragePrice,
			RealizedPnL: realized,
			Reasons:     []string{exit.Reason},
		})
	} else {
		e.State.OpenPositions[index] = position
		_ = AppendEvent(e.Config, TelemetryEvent{
			Timestamp:   position.UpdatedTime,
			Type:        "PARTIAL_EXIT",
			Symbol:      position.Symbol,
			Venue:       position.Venue,
			Side:        position.Side,
			Strategy:    position.Strategy,
			Playbook:    position.Playbook,
			Exit:        fill.AveragePrice,
			RealizedPnL: realized,
			Reasons:     []string{exit.Reason},
		})
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
	if !decision.Allowed {
		_ = AppendStrategyReviewEvent(e.Config, "rejection", event)
	}
}

func (e *Engine) recordCandidateEvent(eventType string, candidate Candidate, now int64, reasons []string) {
	event := TelemetryEvent{
		Timestamp:    now,
		Type:         eventType,
		Symbol:       candidate.Symbol,
		Venue:        candidate.Venue,
		Side:         candidate.Side,
		Strategy:     candidate.Strategy,
		Playbook:     candidate.Playbook,
		Score:        candidate.Score,
		Confidence:   candidate.Confidence,
		Entry:        candidate.EntryPrice,
		Reasons:      append([]string(nil), reasons...),
		StopDistance: math.Abs(candidate.EntryPrice - candidate.StopPrice),
	}
	_ = AppendEvent(e.Config, event)
	if eventType == "CANDIDATE_CREATED" {
		_ = AppendStrategyReviewEvent(e.Config, "candidate", event)
	}
	if strings.Contains(strings.ToLower(eventType), "rejected") {
		_ = AppendStrategyReviewEvent(e.Config, "rejection", event)
	}
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
	counts := scanCounts{RejectReasons: map[string]int{}, Admission: map[string]*CandidateAdmissionSummary{}, Coverage: map[string]*CandidateCoverageRow{}}
	snapshots := make([]orderbook.OrderBookSnapshot, 0)
	entryBySnapshot := map[string]UniverseEntry{}
	for _, entry := range universe {
		coverageFor(counts.Coverage, entry)
		snapshot, err := e.fetchOrderBookForUniverseEntry(entry)
		if err != nil {
			e.recordSystemEvent("recorder_error", []string{fmt.Sprintf("orderbook_fetch_failed:%s:%s", entry.Venue, entry.Symbol), err.Error()})
			coverageFor(counts.Coverage, entry).RejectReasons["orderbook_fetch_failed"]++
			continue
		}
		coverageFor(counts.Coverage, entry).SnapshotFetched = true
		snapshots = append(snapshots, snapshot)
		entryBySnapshot[venueSymbolKey(snapshot.Venue, snapshot.Symbol)] = entry
		_ = AppendEvent(e.Config, TelemetryEvent{
			Timestamp: e.Now().UTC().UnixMilli(),
			Type:      "SCANNER_SNAPSHOT",
			Symbol:    snapshot.Symbol,
			Venue:     snapshot.Venue,
			Mark:      orderbook.Mid(snapshot),
			Reasons: []string{
				fmt.Sprintf("spreadPct=%.6f", orderbook.SpreadPct(snapshot)),
				fmt.Sprintf("liquidity=%.2f", orderbook.DepthWithinPct(snapshot, 1)),
			},
		})
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
		entry := entryBySnapshot[venueSymbolKey(snapshot.Venue, snapshot.Symbol)]
		coverage := coverageFor(counts.Coverage, entry)
		candidates := e.BuildCandidates.BuildCandidates(runtime.ContextFromOrderBook(snapshot))
		if len(candidates) == 0 {
			candidate, ok := e.syntheticCandidate(snapshot, syntheticOpened)
			if !ok {
				e.recordDecision(Candidate{
					Venue: snapshot.Venue, Symbol: snapshot.Symbol, CanonicalSymbol: canonicalFromSymbol(snapshot.Symbol),
					Strategy: "runtime_playbooks", Playbook: "runtime_playbooks", Side: "LONG",
					EntryPrice: orderbook.Mid(snapshot),
				}, RiskDecision{Allowed: false, Reasons: []string{"no_runtime_candidate"}}, e.Now().UTC().UnixMilli())
				e.recordSystemEvent("CANDIDATE_REJECTED", []string{"no_runtime_candidate", snapshot.Venue, snapshot.Symbol})
				counts.Decisions++
				counts.Rejected++
				counts.RejectReasons["no_runtime_candidate"]++
				coverage.Rejected++
				coverage.RejectReasons["no_runtime_candidate"]++
				continue
			}
			e.recordSystemEvent("CANDIDATE_CREATED", []string{"synthetic_test_candidate", snapshot.Venue, snapshot.Symbol})
			candidates = []Candidate{candidate}
		}
		candidates, deduped := e.dedupeCandidates(candidates)
		for _, candidate := range deduped {
			coverage.Candidates++
			coverage.ConfidenceSum += candidate.Confidence
			admissionFor(counts.Admission, candidate.Venue).Created++
			admissionFor(counts.Admission, candidate.Venue).Deduped++
			counts.Decisions++
			counts.Rejected++
			counts.RejectReasons["deduped_out"]++
			coverage.Rejected++
			coverage.RejectReasons["deduped_out"]++
			outcome := CandidateOutcome{Candidate: candidate, Status: "deduped_out", Reason: "same_venue_symbol_side"}
			counts.ApprovedBlocked = append(counts.ApprovedBlocked, outcome)
			e.recordCandidateEvent("CANDIDATE_REJECTED", candidate, e.Now().UTC().UnixMilli(), []string{"deduped_out", "same_venue_symbol_side"})
		}
		for _, candidate := range candidates {
			coverage.Candidates++
			coverage.ConfidenceSum += candidate.Confidence
			admissionFor(counts.Admission, candidate.Venue).Created++
			e.recordCandidateEvent("CANDIDATE_CREATED", candidate, e.Now().UTC().UnixMilli(), candidate.Reasons)
			if reason := e.positionPolicyBlockReason(candidate); reason != "" {
				now := e.Now().UTC().UnixMilli()
				e.recordDecision(candidate, RiskDecision{Allowed: false, Reasons: []string{reason}}, now)
				e.recordCandidateEvent("CANDIDATE_REJECTED", candidate, now, []string{reason})
				counts.Decisions++
				counts.Rejected++
				counts.RejectReasons[reason]++
				coverage.Rejected++
				coverage.RejectReasons[reason]++
				admissionFor(counts.Admission, candidate.Venue).PortfolioBlocked++
				counts.ApprovedBlocked = append(counts.ApprovedBlocked, CandidateOutcome{Candidate: candidate, Status: "portfolio_blocked", Reason: reason})
				continue
			}
			decision, position, err := e.EvaluateCandidate(candidate, snapshot, false)
			if err != nil {
				return counts, err
			}
			counts.Decisions++
			if decision.Allowed {
				counts.Approved++
				coverage.Approved++
				admissionFor(counts.Admission, candidate.Venue).Approved++
				if position == nil {
					counts.ApprovedBlocked = append(counts.ApprovedBlocked, CandidateOutcome{
						Candidate: candidate,
						Status:    "approved_not_entered",
						Reason:    "no_fill",
					})
				}
				syntheticOpened = true
			} else {
				counts.Rejected++
				coverage.Rejected++
				classification := classifyRejectReasons(decision.Reasons)
				switch classification {
				case "state":
					admissionFor(counts.Admission, candidate.Venue).StateBlocked++
				case "portfolio":
					admissionFor(counts.Admission, candidate.Venue).PortfolioBlocked++
				default:
					admissionFor(counts.Admission, candidate.Venue).RiskRejected++
				}
				for _, reason := range decision.Reasons {
					counts.RejectReasons[reason]++
					coverage.RejectReasons[reason]++
				}
			}
		}
	}
	return counts, nil
}

func (e *Engine) fetchOrderBookForUniverseEntry(entry UniverseEntry) (orderbook.OrderBookSnapshot, error) {
	if strings.EqualFold(entry.Venue, "lighter") && strings.TrimSpace(entry.MarketID) != "" {
		marketID, err := strconv.Atoi(strings.TrimSpace(entry.MarketID))
		if err != nil {
			return orderbook.OrderBookSnapshot{}, fmt.Errorf("invalid lighter market id %q for %s: %w", entry.MarketID, entry.Symbol, err)
		}
		return e.FetchMarketBook(entry.Venue, entry.Symbol, marketID)
	}
	return e.FetchOrderBook(entry.Venue, entry.Symbol)
}

func coverageFor(rows map[string]*CandidateCoverageRow, entry UniverseEntry) *CandidateCoverageRow {
	key := venueSymbolKey(entry.Venue, entry.Symbol)
	if rows[key] == nil {
		rows[key] = &CandidateCoverageRow{
			Venue:           strings.ToLower(strings.TrimSpace(entry.Venue)),
			Symbol:          strings.ToUpper(strings.TrimSpace(entry.Symbol)),
			CanonicalSymbol: strings.ToUpper(strings.TrimSpace(entry.CanonicalSymbol)),
			Major:           isMajorCanonical(entry.CanonicalSymbol),
			RejectReasons:   map[string]int{},
		}
	}
	return rows[key]
}

func candidateCoverageRows(rows map[string]*CandidateCoverageRow) []CandidateCoverageRow {
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]CandidateCoverageRow, 0, len(keys))
	for _, key := range keys {
		row := *rows[key]
		if row.Candidates > 0 {
			row.AverageConfidence = row.ConfidenceSum / float64(row.Candidates)
		}
		out = append(out, row)
	}
	return out
}

func (e *Engine) dedupeCandidates(candidates []Candidate) ([]Candidate, []Candidate) {
	if e.Config.AllowStrategyStacking {
		return candidates, nil
	}
	bestByKey := map[string]Candidate{}
	for _, candidate := range candidates {
		key := candidateRouteKey(candidate)
		if current, ok := bestByKey[key]; !ok || candidateBetter(candidate, current) {
			bestByKey[key] = candidate
		}
	}
	kept := make([]Candidate, 0, len(bestByKey))
	deduped := make([]Candidate, 0, len(candidates)-len(bestByKey))
	for _, candidate := range candidates {
		best := bestByKey[candidateRouteKey(candidate)]
		if sameCandidateSelection(candidate, best) {
			kept = append(kept, candidate)
			delete(bestByKey, candidateRouteKey(candidate))
			continue
		}
		deduped = append(deduped, candidate)
	}
	return kept, deduped
}

func candidateBetter(candidate Candidate, current Candidate) bool {
	if candidate.Score != current.Score {
		return candidate.Score > current.Score
	}
	if candidate.Confidence != current.Confidence {
		return candidate.Confidence > current.Confidence
	}
	return candidate.Strategy < current.Strategy
}

func sameCandidateSelection(a Candidate, b Candidate) bool {
	return strings.EqualFold(a.Venue, b.Venue) &&
		strings.EqualFold(a.Symbol, b.Symbol) &&
		normalizeSide(a.Side) == normalizeSide(b.Side) &&
		a.Strategy == b.Strategy &&
		a.Score == b.Score &&
		a.Confidence == b.Confidence
}

func candidateRouteKey(candidate Candidate) string {
	return venueSymbolKey(candidate.Venue, candidate.Symbol) + ":" + normalizeSide(candidate.Side)
}

func (e *Engine) positionPolicyBlockReason(candidate Candidate) string {
	if e.Config.AllowStrategyStacking {
		return ""
	}
	limit := e.Config.MaxPositionsPerSymbolVenue
	if limit <= 0 {
		limit = 1
	}
	count := 0
	for _, position := range e.State.OpenPositions {
		if strings.EqualFold(position.Venue, candidate.Venue) &&
			strings.EqualFold(position.Symbol, candidate.Symbol) &&
			normalizeSide(position.Side) == normalizeSide(candidate.Side) {
			count++
		}
	}
	if count >= limit {
		return "position_policy_duplicate"
	}
	return ""
}

func admissionFor(rows map[string]*CandidateAdmissionSummary, venue string) *CandidateAdmissionSummary {
	venue = strings.ToLower(strings.TrimSpace(venue))
	if venue == "" {
		venue = "unknown"
	}
	if rows[venue] == nil {
		rows[venue] = &CandidateAdmissionSummary{Venue: venue}
	}
	return rows[venue]
}

func admissionSummaryRows(rows map[string]*CandidateAdmissionSummary) []CandidateAdmissionSummary {
	venues := make([]string, 0, len(rows))
	for venue := range rows {
		venues = append(venues, venue)
	}
	sort.Strings(venues)
	out := make([]CandidateAdmissionSummary, 0, len(venues))
	for _, venue := range venues {
		out = append(out, *rows[venue])
	}
	return out
}

func classifyRejectReasons(reasons []string) string {
	for _, reason := range reasons {
		switch reason {
		case "max_open_positions", "position_policy_duplicate":
			return "portfolio"
		case "max_trades_per_symbol_per_day", "trade_budget_exceeded", "symbol_cooldown", "symbol_lock", "loss_cooldown":
			return "state"
		}
	}
	return "risk"
}

func (e *Engine) recordScannerCycleSummary(summary RuntimeSummary) error {
	return AppendEvent(e.Config, TelemetryEvent{
		Timestamp: e.Now().UTC().UnixMilli(),
		Type:      "SCANNER_CYCLE_SUMMARY",
		Decision:  "system",
		Reasons: []string{
			fmt.Sprintf("universeMode=%s", summary.UniverseMode),
			fmt.Sprintf("selectedSymbols=%d", summary.SelectedSymbols),
			fmt.Sprintf("venues=%d", summary.Venues),
			fmt.Sprintf("decisions=%d", summary.Decisions),
			fmt.Sprintf("approved=%d", summary.Approved),
			fmt.Sprintf("rejected=%d", summary.Rejected),
			fmt.Sprintf("openPositions=%d", summary.OpenCount),
		},
	})
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

func defaultMarketOrderBookFetcher(venue string, symbol string, marketID int) (orderbook.OrderBookSnapshot, error) {
	switch strings.ToLower(strings.TrimSpace(venue)) {
	case "lighter":
		return lighter.GetMainnetOrderBookByMarketID(symbol, marketID)
	default:
		return defaultOrderBookFetcher(venue, symbol)
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
