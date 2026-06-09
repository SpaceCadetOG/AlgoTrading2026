package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"AlgoTrading2026/backtest"
	"AlgoTrading2026/config"
	"AlgoTrading2026/datasets"
	"AlgoTrading2026/exchanges"
	"AlgoTrading2026/exchanges/aster"
	"AlgoTrading2026/exchanges/hyperliquid"
	"AlgoTrading2026/exchanges/lighter"
	"AlgoTrading2026/features"
	"AlgoTrading2026/indicators"
	runtime "AlgoTrading2026/internal/runtime"
	"AlgoTrading2026/l2recorder"
	"AlgoTrading2026/labels"
	"AlgoTrading2026/orderbook"
	"AlgoTrading2026/orderflow"
	"AlgoTrading2026/pairs"
	"AlgoTrading2026/paper"
	"AlgoTrading2026/realism"
	"AlgoTrading2026/research"
	"AlgoTrading2026/risk"
	"AlgoTrading2026/riskmetrics"
	"AlgoTrading2026/series"
	"AlgoTrading2026/statarb"
	chapter2 "AlgoTrading2026/strategies/chapter2"
	chapter4 "AlgoTrading2026/strategies/chapter4"
	chapter5 "AlgoTrading2026/strategies/chapter5"
	"AlgoTrading2026/strategy"
	"AlgoTrading2026/system"
	"AlgoTrading2026/tradetape"
	"AlgoTrading2026/volatility"
	"AlgoTrading2026/volumeprofile"
	"AlgoTrading2026/vwap"
)

const (
	ResearchDays             = 30
	ResearchInterval         = "15m"
	CandlesPerDay15m         = 96
	ResearchCandleCap        = ResearchDays * CandlesPerDay15m
	VWAPTimeframeCandleLimit = 1000
)

type venueCandles struct {
	Venue   string
	Symbol  string
	Candles []exchanges.Candle
}

type strategyRun struct {
	Name    string
	Signals series.SignalResult
	Regimes []volatility.Regime
}

type strategyBacktestResult struct {
	Name    string
	Signals series.SignalResult
	Report  backtest.Report
	Trades  []strategy.Trade
}

type vwapChapter3Study struct {
	ContextRows   []research.ContextFeatureRow
	TimeframeRows []research.TimeframeFeatureRow
	TrendRows     []research.TrendAlignmentRow
	Summary       research.VWAPChapter3Summary
}

type orderBookResearchRows struct {
	Features        []research.OrderBookFeatureRow
	VWAPInteraction []research.VWAPL2InteractionRow
	Snapshots       []orderbook.OrderBookSnapshot
}

type vwapL2RefreshStudy struct {
	FeatureRows []research.VWAPL2FeatureRow
	ContextRows []research.ContextL2FeatureRow
	Summary     research.VWAPL2BehaviorSummary
}

type priceActionStudy struct {
	Rows    []research.PriceActionFeatureRow
	Summary research.PriceActionSummary
}

type priceActionStrategyStudy struct {
	Rows    []research.PriceActionStrategyStudyRow
	Summary research.PriceActionStrategySummary
}

type priceActionPhase3Study struct {
	Rows    []research.PriceActionPhase3StudyRow
	Summary research.PriceActionPhase3Summary
}

type volumeProfileStudy struct {
	Profile volumeprofile.VolumeProfile
	Rows    []research.VolumeProfileFeatureRow
	Summary research.VolumeProfileSummary
}

type scopedVolumeProfileStudy struct {
	Profiles []volumeprofile.ScopedProfile
	Rows     []research.ScopedVolumeProfileFeatureRow
	Summary  research.ScopedVolumeProfileSummary
}

type volumeProfileShapeStudy struct {
	Rows    []research.VolumeProfileShapeStudyRow
	Summary research.VolumeProfileShapeSummary
}

type flexibleVolumeProfileStudy struct {
	Profiles []volumeprofile.FlexibleProfile
	Rows     []research.FlexibleVolumeProfileRow
	Summary  research.FlexibleVolumeProfileSummary
}

type profileAcceptanceQualityStudy struct {
	Rows    []research.ProfileAcceptanceQualityRow
	Summary research.ProfileAcceptanceQualitySummary
}

type volumeSetupAccumulationStudy struct {
	Rows    []research.VolumeSetupAccumulationRow
	Summary research.VolumeSetupAccumulationSummary
}

type volumeSetupAccumulationQualityStudy struct {
	Rows    []research.VolumeSetupAccumulationQualityRow
	Summary research.VolumeSetupAccumulationQualitySummary
}

type volumeSetupTrendStudy struct {
	Rows    []research.VolumeSetupTrendRow
	Summary research.VolumeSetupTrendSummary
}

type volumeSetupTrendQualityStudy struct {
	Rows    []research.VolumeSetupTrendQualityRow
	Summary research.VolumeSetupTrendQualitySummary
}

type volumeSetupRejectionStudy struct {
	Rows    []research.VolumeSetupRejectionRow
	Summary research.VolumeSetupRejectionSummary
}

type volumeSetupRejectionQualityStudy struct {
	Rows    []research.VolumeSetupRejectionQualityRow
	Summary research.VolumeSetupRejectionQualitySummary
}

type volumeSetupReversalStudy struct {
	Rows    []research.VolumeSetupReversalRow
	Summary research.VolumeSetupReversalSummary
}

type volumeSetupReversalQualityStudy struct {
	Rows    []research.VolumeSetupReversalQualityRow
	Summary research.VolumeSetupReversalQualitySummary
}

type volumeSetupComparisonStudy struct {
	Rows    []research.VolumeSetupComparisonRow
	Summary research.VolumeSetupComparisonSummary
}

type volumeProfileBookCompletionStudy struct {
	Packet research.VolumeProfileBookCompletionPacket
}

type setupFeatureLabelStudy struct {
	Rows    []research.SetupFeatureLabelRow
	Summary research.SetupLabelsSummary
}

type instrumentUniverseStudy struct {
	Rows    []research.InstrumentUniverseRow
	Summary research.InstrumentUniverseSummary
}

type volumeProfileFinalBookStudy struct {
	Packet research.VolumeProfileFinalBookPacket
}

type orderFlowFoundationStudy struct {
	Rows    []research.OrderFlowFoundationRow
	Summary research.OrderFlowFoundationSummary
}

type orderFlowDataCapabilityAuditStudy struct {
	Report research.OrderFlowDataCapabilityAudit
}

type orderFlowRealFootprintsStudy struct {
	Bars    []orderflow.FootprintBar
	Rows    []research.OrderFlowRealFootprintRow
	Summary research.OrderFlowRealFootprintSummary
}

type orderFlowVolumeClusterStudy struct {
	Rows    []research.OrderFlowVolumeClusterSetupRow
	Summary research.OrderFlowVolumeClusterSummary
}

type orderFlowVolumeClusterQualityStudy struct {
	Rows    []research.OrderFlowVolumeClusterQualityRow
	Summary research.OrderFlowVolumeClusterQualitySummary
}

type orderFlowMultipleHVNStudy struct {
	Rows    []research.OrderFlowMultipleHVNSetupRow
	Summary research.OrderFlowMultipleHVNSummary
}

type orderFlowBackfilledFootprintsStudy struct {
	Bars    []orderflow.FootprintBar
	Rows    []research.OrderFlowRealFootprintRow
	Summary research.OrderFlowBackfilledFootprintSummary
}

type orderFlowBackfilledSetupReplayStudy struct {
	Replay     research.OrderFlowBackfilledSetupReplay
	Comparison research.OrderFlowLiveVsBackfillComparison
}

type orderFlowTradesFilterStudy struct {
	Rows    []research.OrderFlowTradesFilterSetupRow
	Summary research.OrderFlowTradesFilterSummary
}

type orderFlowStackedImbalanceStudy struct {
	Rows    []research.OrderFlowStackedImbalanceSetupRow
	Summary research.OrderFlowStackedImbalanceSummary
}

type orderFlowUnfinishedBusinessStudy struct {
	Rows    []research.OrderFlowUnfinishedBusinessSetupRow
	Summary research.OrderFlowUnfinishedBusinessSummary
}

type orderFlowBigLimitOrderStudy struct {
	Rows    []research.OrderFlowBigLimitOrderRow
	Summary research.OrderFlowBigLimitOrderSummary
}

type orderFlowAbsorptionStudy struct {
	Rows    []research.OrderFlowAbsorptionRow
	Summary research.OrderFlowAbsorptionSummary
}

type orderFlowAggressiveDeltaStudy struct {
	Rows    []research.OrderFlowAggressiveDeltaRow
	Summary research.OrderFlowAggressiveDeltaSummary
}

type orderFlowCumulativeDeltaDivergenceStudy struct {
	Rows    []research.OrderFlowCumulativeDeltaDivergenceRow
	Summary research.OrderFlowCumulativeDeltaDivergenceSummary
}

type orderFlowConfirmationComparisonStudy struct {
	Rows    []research.OrderFlowConfirmationComparisonRow
	Summary research.OrderFlowConfirmationComparisonSummary
}

type bookTradeRulesStudy struct {
	Packet        strategy.BookTradeRulesPacket
	ArchivedCount int
}

func asterEnv(name string) string {
	if config.IsTestnet() {
		return os.Getenv("ASTER_TESTNET_" + name)
	}
	return os.Getenv("ASTER_" + name)
}

func runL2Recorder() {
	cfg := l2recorder.DefaultRecorderConfig()
	recorder := l2recorder.NewRecorder(cfg, nil, nil)

	fmt.Println("=== HISTORICAL L2 RECORDER ===")
	fmt.Printf("venues=%d\n", len(cfg.Venues))
	fmt.Printf("symbols=%d\n", len(cfg.Symbols))
	fmt.Printf("intervalSeconds=%d\n", cfg.IntervalSeconds)
	fmt.Printf("maxSnapshots=%d\n", cfg.MaxSnapshots)
	fmt.Printf("output=%s\n", cfg.OutputPath)
	fmt.Println()

	summary, err := recorder.Run()
	if err != nil {
		log.Fatalf("run l2 recorder: %v", err)
	}
	fmt.Printf("rounds=%d\n", summary.Rounds)
	fmt.Printf("rows=%d\n", summary.Rows)
	fmt.Printf("errors=%d\n", summary.Errors)
	fmt.Printf("wrote %s\n", cfg.OutputPath)
}

func main() {
	_ = godotenv.Load()
	log.Println("TRADING_ENV:", config.TradingEnv())

	backfillConfig := tradetape.DefaultBackfillConfig()
	if backfillConfig.RunAsterAggTradesBackfill {
		runAsterAggTradesBackfill(backfillConfig)
		return
	}

	tradeTapeConfig := tradetape.DefaultRecorderConfig()
	if tradeTapeConfig.RunTradeTapeRecorder {
		runTradeTapeRecorder(tradeTapeConfig)
		return
	}

	if strings.EqualFold(os.Getenv("RUN_L2_RECORDER"), "true") {
		runL2Recorder()
		return
	}

	if strings.EqualFold(config.TradingMode(), "paper") {
		runPaperRuntime()
		return
	}
	if strings.EqualFold(config.TradingMode(), "live") {
		runLiveRuntime()
		return
	}

	if shouldRunFullResearchHarness() {
		runResearchHarness()
		return
	}

	runBookTradeRules()
}

func runLiveRuntime() {
	cfg := paper.DefaultConfig()
	cfg.Mode = "live"
	engine, err := paper.NewEngine(cfg)
	if err != nil {
		log.Fatalf("init live runtime: %v", err)
	}

	client := aster.NewClient(asterEnv("USER"), asterEnv("SIGNER"), asterEnv("PRIVATE_KEY"))
	executor := runtime.LiveExecutor{
		Placer: client,
		Gate:   runtime.LiveGateFromEnv(),
	}

	ctx := runtimeContext()
	interval := time.Duration(cfg.ScanIntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	cycle := 0
	fmt.Println("=== LIVE RUNTIME LOOP ===")
	fmt.Printf("mode=%s\n", cfg.Mode)
	fmt.Printf("gateLiveEnabled=%t\n", executor.Gate.LiveEnabled)
	fmt.Printf("gateExecutionEnabled=%t\n", executor.Gate.ExecutionEnabled)
	fmt.Printf("gateVenueHealthy=%t\n", executor.Gate.VenueHealthy)
	fmt.Printf("gateAccountReady=%t\n", executor.Gate.AccountReady)
	fmt.Printf("scanIntervalSeconds=%d\n", int(interval/time.Second))
	for {
		select {
		case <-ctx.Done():
			fmt.Println("status=stopped")
			return
		default:
		}
		cycle++
		summary, err := runLiveCycle(engine, executor, cfg)
		if err != nil {
			log.Fatalf("run live cycle: %v", err)
		}
		fmt.Printf("cycle=%d selectedSymbols=%d decisions=%d approved=%d refused=%d placed=%d\n",
			cycle, summary.SelectedSymbols, summary.Decisions, summary.Approved, summary.Rejected, summary.OpenCount)
		if cfg.MaxRuntimeCycles > 0 && cycle >= cfg.MaxRuntimeCycles {
			fmt.Println("status=max_cycles_reached")
			return
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			fmt.Println("status=stopped")
			return
		case <-timer.C:
		}
	}
}

func runLiveCycle(engine *paper.Engine, executor runtime.LiveExecutor, cfg paper.Config) (paper.RuntimeSummary, error) {
	summary := paper.RuntimeSummary{}
	if err := paper.EnsureDataFiles(cfg); err != nil {
		return summary, err
	}
	now := time.Now().UTC().UnixMilli()
	if err := paper.AppendEvent(cfg, paper.TelemetryEvent{
		Timestamp: now,
		Type:      "MODE_SELECTED",
		Decision:  "system",
		Reasons:   []string{"mode=live"},
	}); err != nil {
		return summary, err
	}
	if err := paper.AppendEvent(cfg, paper.TelemetryEvent{
		Timestamp: now,
		Type:      "VENUE_HEALTH",
		Decision:  "system",
		Reasons: []string{
			fmt.Sprintf("venueHealthy=%t", executor.Gate.VenueHealthy),
			fmt.Sprintf("accountReady=%t", executor.Gate.AccountReady),
			fmt.Sprintf("killSwitch=%t", executor.Gate.KillSwitchActive),
		},
	}); err != nil {
		return summary, err
	}
	decisions := 0
	approved := 0
	refused := 0
	placed := 0
	universe, err := engine.SelectUniverse(cfg)
	if err != nil {
		return summary, err
	}
	summary.UniverseMode = universe.Mode
	summary.Min24hVolumeUSD = universe.Min24hVolumeUSD
	summary.SelectedSymbols = len(universe.Selected)
	summary.Venues = len(cfg.Venues)
	for _, entry := range universe.Selected {
		snapshot, err := engine.FetchOrderBook(entry.Venue, entry.Symbol)
		if err != nil {
			log.Printf("live snapshot failed venue=%s symbol=%s: %v", entry.Venue, entry.Symbol, err)
			continue
		}
		_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
			Timestamp: time.Now().UTC().UnixMilli(),
			Type:      "SCANNER_SNAPSHOT",
			Symbol:    snapshot.Symbol,
			Venue:     snapshot.Venue,
			Mark:      orderbook.Mid(snapshot),
			Reasons: []string{
				fmt.Sprintf("spreadPct=%.6f", orderbook.SpreadPct(snapshot)),
				fmt.Sprintf("liquidity=%.2f", orderbook.DepthWithinPct(snapshot, 1)),
			},
		})
		candidates := engine.BuildCandidates.BuildCandidates(runtime.ContextFromOrderBook(snapshot))
		for _, candidate := range candidates {
			decisions++
			_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
				Timestamp:  time.Now().UTC().UnixMilli(),
				Type:       "CANDIDATE_CREATED",
				Symbol:     candidate.Symbol,
				Venue:      candidate.Venue,
				Side:       candidate.Side,
				Strategy:   candidate.Strategy,
				Playbook:   candidate.Playbook,
				Score:      candidate.Score,
				Confidence: candidate.Confidence,
				Entry:      candidate.EntryPrice,
				Reasons:    candidate.Reasons,
			})
			riskDecision := paper.CheckRisk(engine.State, candidate, cfg, time.Now().UTC().UnixMilli())
			if !riskDecision.Allowed {
				refused++
				_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
					Timestamp:  time.Now().UTC().UnixMilli(),
					Type:       "RISK_REJECTED",
					Symbol:     candidate.Symbol,
					Venue:      candidate.Venue,
					Side:       candidate.Side,
					Strategy:   candidate.Strategy,
					Playbook:   candidate.Playbook,
					Score:      candidate.Score,
					Confidence: candidate.Confidence,
					Reasons:    riskDecision.Reasons,
				})
				continue
			}
			approved++
			_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
				Timestamp:  time.Now().UTC().UnixMilli(),
				Type:       "EXECUTION_REQUESTED",
				Symbol:     candidate.Symbol,
				Venue:      candidate.Venue,
				Side:       candidate.Side,
				Strategy:   candidate.Strategy,
				Playbook:   candidate.Playbook,
				Score:      candidate.Score,
				Confidence: candidate.Confidence,
				Reasons:    candidate.Reasons,
			})
			result, err := executor.Execute(runtime.ExecutionDecision{
				Candidate: candidate,
				Risk:      runtime.RiskDecision{Allowed: riskDecision.Allowed, Reasons: riskDecision.Reasons},
				Mode:      "live",
			})
			if err != nil {
				return summary, err
			}
			if result.Accepted {
				placed++
				_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
					Timestamp: time.Now().UTC().UnixMilli(),
					Type:      "ORDER_SUBMITTED",
					Symbol:    candidate.Symbol,
					Venue:     candidate.Venue,
					Side:      candidate.Side,
					Strategy:  candidate.Strategy,
					Playbook:  candidate.Playbook,
					Decision:  result.Status,
					Reasons:   []string{result.OrderID},
				})
				fmt.Printf("live_order_accepted venue=%s symbol=%s orderID=%s status=%s\n", result.Venue, result.Symbol, result.OrderID, result.Status)
			} else {
				refused++
				_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
					Timestamp: time.Now().UTC().UnixMilli(),
					Type:      "ORDER_REJECTED",
					Symbol:    candidate.Symbol,
					Venue:     candidate.Venue,
					Side:      candidate.Side,
					Strategy:  candidate.Strategy,
					Playbook:  candidate.Playbook,
					Decision:  result.Status,
					Reasons:   []string{result.Message},
				})
				fmt.Printf("live_order_refused venue=%s symbol=%s reason=%s\n", candidate.Venue, candidate.Symbol, result.Message)
			}
			break
		}
		if placed > 0 {
			break
		}
	}
	summary.Decisions = decisions
	summary.Approved = approved
	summary.Rejected = refused
	summary.OpenCount = placed
	_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
		Timestamp: time.Now().UTC().UnixMilli(),
		Type:      "SCANNER_CYCLE_SUMMARY",
		Decision:  "system",
		Reasons: []string{
			fmt.Sprintf("decisions=%d", decisions),
			fmt.Sprintf("approved=%d", approved),
			fmt.Sprintf("refused=%d", refused),
			fmt.Sprintf("placed=%d", placed),
		},
	})
	_ = paper.AppendEvent(cfg, paper.TelemetryEvent{
		Timestamp: time.Now().UTC().UnixMilli(),
		Type:      "RECONCILE_RESULT",
		Decision:  "system",
		Reasons:   []string{"live_order_reconciliation_pending"},
	})
	return summary, nil
}

func shouldRunFullResearchHarness() bool {
	return strings.EqualFold(os.Getenv("RUN_FULL_RESEARCH_HARNESS"), "true")
}

func runAsterAggTradesBackfill(cfg tradetape.BackfillConfig) {
	provider := aster.AggTradeProvider{}
	rows, dedupedRows, dedupe, err := tradetape.RunBackfill(provider, cfg)
	errors := 0
	if err != nil {
		errors = 1
	}
	summary := research.BuildTradeTapeBackfillSummary(
		provider.Venue(),
		cfg.Symbol,
		cfg.StartMS,
		cfg.EndMS,
		len(aster.BackfillAsterAggTradeWindows(cfg.StartMS, cfg.EndMS)),
		len(rows),
		dedupe,
		errors,
	)
	if err := research.WriteTradeTapeBackfillSummaryJSON("research/trade_tape_backfill_summary.json", summary); err != nil {
		log.Fatalf("write trade tape backfill summary json: %v", err)
	}
	if err := research.WriteTradeTapeBackfillSummaryMarkdown("research/trade_tape_backfill_summary.md", summary); err != nil {
		log.Fatalf("write trade tape backfill summary markdown: %v", err)
	}
	if err != nil {
		log.Fatalf("run aster aggTrades backfill: %v", err)
	}
	fmt.Println("=== ASTER AGGTRADES BACKFILL ===")
	fmt.Printf("symbol=%s\n", cfg.Symbol)
	fmt.Printf("start=%d\n", cfg.StartMS)
	fmt.Printf("end=%d\n", cfg.EndMS)
	fmt.Printf("windows=%d\n", summary.Windows)
	fmt.Printf("rows=%d\n", len(rows))
	fmt.Printf("dedupedRows=%d\n", len(dedupedRows))
	fmt.Printf("duplicates=%d\n", dedupe.DuplicateRows)
	fmt.Printf("invalidRows=%d\n", dedupe.InvalidRows)
	fmt.Printf("output=%s\n", cfg.OutputPath)
	fmt.Printf("deduped=%s\n", cfg.DedupedOutputPath)
	fmt.Println("wrote research/trade_tape_backfill_summary.json")
	fmt.Println("wrote research/trade_tape_backfill_summary.md")
}

func runTradeTapeRecorder(cfg tradetape.RecorderConfig) {
	recorder := tradetape.NewRecorder(cfg, buildTradeTapeFetchers())
	fmt.Println("=== TRADE TAPE RECORDER ===")
	fmt.Printf("venues=%d\n", len(recorder.Fetchers))
	fmt.Printf("intervalSeconds=%d\n", cfg.IntervalSeconds)
	fmt.Printf("maxRounds=%d\n", cfg.MaxRounds)
	fmt.Printf("output=%s\n", cfg.OutputPath)
	fmt.Println()

	summary, err := recorder.Run()
	if err != nil {
		log.Fatalf("run trade tape recorder: %v", err)
	}
	fmt.Printf("rounds=%d\n", summary.Rounds)
	fmt.Printf("rows=%d\n", summary.Rows)
	fmt.Printf("errors=%d\n", summary.Errors)
	fmt.Printf("wrote %s\n", cfg.OutputPath)
}

func runBookTradeRules() {
	study := buildBookTradeRules()
	playbooks, entryRules, stopRules, targetRules, managementRules := strategyRuleCounts(study.Packet.Playbooks)

	fmt.Println("=== BOOK TRADE RULES ===")
	fmt.Printf("playbooks=%d\n", len(study.Packet.Playbooks))
	fmt.Printf("entryRules=%d\n", entryRules)
	fmt.Printf("stopRules=%d\n", stopRules)
	fmt.Printf("targetRules=%d\n", targetRules)
	fmt.Printf("managementRules=%d\n", managementRules)
	fmt.Printf("riskRules=%d\n", playbooks)
	fmt.Printf("executionEnabled=%t\n", study.Packet.ExecutionEnabled)
	fmt.Printf("paperTradingEnabled=%t\n", study.Packet.PaperTradingEnabled)
	fmt.Printf("status=%s\n", study.Packet.Status)
	fmt.Printf("archivedResearchArtifacts=%d\n", study.ArchivedCount)
	fmt.Println("wrote docs/book_trade_rules.md")
	fmt.Println("wrote strategy/book_trade_rules_packet.json")
	fmt.Println("wrote strategy/book_trade_rules_packet.md")
}

func runPaperRuntime() {
	cfg := paper.DefaultConfig()
	engine, err := paper.NewEngine(cfg)
	if err != nil {
		log.Fatalf("init paper engine: %v", err)
	}
	fmt.Println("=== PAPER RUNTIME LOOP ===")
	fmt.Printf("mode=%s\n", cfg.Mode)
	fmt.Println("execution=simulated")
	fmt.Printf("liveEnabled=%t\n", engine.State.LiveEnabled)
	fmt.Printf("scanIntervalSeconds=%d\n", cfg.ScanIntervalSeconds)
	ctx := runtimeContext()
	err = engine.RunLoop(ctx, func(snapshot paper.RuntimeLoopSnapshot) {
		status := engine.StatusPayload()
		fmt.Printf("cycle=%d universeMode=%s selectedSymbols=%d venues=%d decisions=%d approved=%d rejected=%d openPositions=%d equity=%.2f\n",
			snapshot.Cycle,
			snapshot.Summary.UniverseMode,
			snapshot.Summary.SelectedSymbols,
			snapshot.Summary.Venues,
			snapshot.Summary.Decisions,
			snapshot.Summary.Approved,
			snapshot.Summary.Rejected,
			status.Paper.OpenCount,
			status.Paper.Equity,
		)
		printVenueDiscovery(snapshot.Summary)
		printQualificationDiagnostics(snapshot.Summary.QualificationDiagnostics)
		printSelectedUniverse(snapshot.Summary.SelectedUniverse, 20)
		printOpenPositions(status.Paper.OpenPositions)
		printRecentClosed(status.Paper.RecentClosed, 10)
		printRejectSummary(snapshot.Summary.RejectReasons, 10)
		printCandidateAdmission(snapshot.Summary.CandidateAdmission)
		printApprovedBlocked(snapshot.Summary.ApprovedBlocked, 10)
		printStateBlockers(snapshot.Summary.StateBlockers)
		printCycleFooter(snapshot.Summary)
	})
	if err != nil && err != context.Canceled {
		log.Fatalf("run paper loop: %v", err)
	}
	fmt.Println("status=stopped")
}

func printQualificationDiagnostics(rows []paper.QualificationDiagnostics) {
	if len(rows) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("=== QUALIFICATION DIAGNOSTICS ===")
	for _, row := range rows {
		if row.Discovered == 0 {
			continue
		}
		fmt.Printf("%s discovered=%d qualified=%d rejected=%d", row.Venue, row.Discovered, row.Qualified, row.Rejected)
		if row.Discovered > 0 && row.Qualified == 0 {
			fmt.Print(" state=all_discovered_assets_filtered")
		}
		fmt.Println()
		type reasonRow struct {
			Reason string
			Count  int
		}
		reasons := make([]reasonRow, 0, len(row.Reasons))
		for reason, count := range row.Reasons {
			if count > 0 {
				reasons = append(reasons, reasonRow{Reason: reason, Count: count})
			}
		}
		sort.Slice(reasons, func(i, j int) bool {
			if reasons[i].Count == reasons[j].Count {
				return reasons[i].Reason < reasons[j].Reason
			}
			return reasons[i].Count > reasons[j].Count
		})
		limit := len(reasons)
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("  %s=%d\n", reasons[i].Reason, reasons[i].Count)
		}
	}
}

func runtimeContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return ctx
}

func printVenueDiscovery(summary paper.RuntimeSummary) {
	fmt.Println()
	fmt.Println("=== VENUE DISCOVERY ===")
	fmt.Printf("universeMode=%s manualSymbols=%t\n", summary.UniverseMode, summary.UniverseMode == "manual")
	if len(summary.DiscoveryStats) == 0 {
		fmt.Println("discovery stats: unavailable")
		return
	}
	for _, stat := range summary.DiscoveryStats {
		if stat.Failed {
			fmt.Printf("%s discovered=%d qualified=%d selected=%d failed=true error=%s\n", stat.Venue, stat.Discovered, stat.Qualified, stat.Selected, stat.Error)
			continue
		}
		fmt.Printf("%s discovered=%d qualified=%d selected=%d\n", stat.Venue, stat.Discovered, stat.Qualified, stat.Selected)
	}
}

func printSelectedUniverse(rows []paper.UniverseEntry, limitPerVenue int) {
	fmt.Println()
	fmt.Println("=== SELECTED UNIVERSE ===")
	if len(rows) == 0 {
		fmt.Println("selected universe: empty")
		return
	}
	byVenue := map[string][]paper.UniverseEntry{}
	var venues []string
	for _, row := range rows {
		venue := strings.ToLower(strings.TrimSpace(row.Venue))
		if _, ok := byVenue[venue]; !ok {
			venues = append(venues, venue)
		}
		byVenue[venue] = append(byVenue[venue], row)
	}
	sort.Strings(venues)
	for _, venue := range venues {
		venueRows := byVenue[venue]
		fmt.Printf("%s:\n", venue)
		limit := limitPerVenue
		if limit <= 0 || limit > len(venueRows) {
			limit = len(venueRows)
		}
		for i := 0; i < limit; i++ {
			row := venueRows[i]
			fmt.Printf("%d. %s", i+1, row.Symbol)
			if row.CanonicalSymbol != "" && row.CanonicalSymbol != row.Symbol {
				fmt.Printf(" canonical=%s", row.CanonicalSymbol)
			}
			if row.MarketID != "" {
				fmt.Printf(" marketId=%s", row.MarketID)
			}
			if row.Rank > 0 {
				fmt.Printf(" rank=%d", row.Rank)
			}
			if row.Volume24hUSD > 0 {
				fmt.Printf(" vol24h=%.0f", row.Volume24hUSD)
			}
			fmt.Println()
		}
		if limit < len(venueRows) {
			fmt.Printf("showing %d of %d selected symbols for %s\n", limit, len(venueRows), venue)
		}
	}
}

func printOpenPositions(positions []paper.PaperPosition) {
	fmt.Println()
	fmt.Println("=== OPEN POSITIONS ===")
	if len(positions) == 0 {
		fmt.Println("open positions: none")
		return
	}
	for i, position := range positions {
		fmt.Printf("%d. venue=%s symbol=%s side=%s strat=%q\n", i+1, position.Venue, position.Symbol, position.Side, firstNonEmpty(position.Strategy, position.Playbook))
		fmt.Printf("   qty=%.8f entry=%.2f mark=%.2f pnl=%+.4f\n", position.Quantity, position.EntryPrice, position.MarkPrice, position.OpenPnL)
		fmt.Printf("   stop=%.2f tp1=%.2f tp2=%.2f tp3=%.2f\n", position.Stop, position.TP1, position.TP2, position.TP3)
		fmt.Printf("   opened=%s provenance=%s\n", formatUnixMillis(position.OpenedTime), firstNonEmpty(position.Provenance, "unknown"))
	}
}

func printRecentClosed(positions []paper.PaperPosition, limit int) {
	fmt.Println()
	fmt.Println("=== RECENT CLOSED ===")
	if len(positions) == 0 {
		fmt.Println("recent closed: none")
		return
	}
	if limit <= 0 || limit > len(positions) {
		limit = len(positions)
	}
	for i := 0; i < limit; i++ {
		position := positions[i]
		hold := time.Duration(0)
		if position.OpenedTime > 0 && position.ClosedTime > position.OpenedTime {
			hold = time.Duration(position.ClosedTime-position.OpenedTime) * time.Millisecond
		}
		fmt.Printf("%d. venue=%s symbol=%s side=%s exit=%s realized=%+.4f hold=%s strat=%q\n",
			i+1,
			position.Venue,
			position.Symbol,
			position.Side,
			firstNonEmpty(position.ExitReason, "unknown"),
			position.RealizedPnL,
			formatDuration(hold),
			firstNonEmpty(position.Strategy, position.Playbook),
		)
	}
}

func printStateBlockers(blockers paper.StateBlockers) {
	fmt.Println()
	fmt.Println("=== STATE BLOCKERS ===")
	fmt.Printf("stateBlockers=%t cooldownSymbols=%d symbolsAtDailyMax=%d lossCooldownActive=%t openSlots=%d maxOpenPositions=%d openPositionBlocked=%t\n",
		blockers.Active,
		blockers.CooldownSymbols,
		blockers.SymbolsAtDailyMax,
		blockers.LossCooldownActive,
		blockers.OpenPositionSlots,
		blockers.MaxOpenPositions,
		blockers.OpenPositionBlocked,
	)
}

func printCycleFooter(summary paper.RuntimeSummary) {
	fmt.Println()
	fmt.Println("=== CYCLE DIAGNOSIS ===")
	if summary.Approved == 0 && summary.Rejected > 0 {
		stateRejects := rejectCount(summary.RejectReasons, "max_open_positions", "max_trades_per_symbol_per_day", "trade_budget_exceeded", "symbol_cooldown", "symbol_lock", "loss_cooldown")
		riskRejects := rejectCount(summary.RejectReasons, "poor_rr", "funding_hazard", "spread_too_wide", "insufficient_liquidity", "invalid_bracket_geometry")
		strategyRejects := rejectCount(summary.RejectReasons, "no_runtime_candidate", "no_live_signal_mapping_yet")
		fmt.Printf("approved=0 rejected=%d stateRejects=%d riskRejects=%d strategyRejects=%d\n", summary.Rejected, stateRejects, riskRejects, strategyRejects)
		if stateRejects > 0 {
			fmt.Println("diagnosis=saved_state_or_limits_blocking_entries")
		} else if riskRejects > 0 {
			fmt.Println("diagnosis=risk_filters_blocking_entries")
		} else if strategyRejects > 0 {
			fmt.Println("diagnosis=strategy_mapping_or_candidate_generation_blocking_entries")
		} else {
			fmt.Println("diagnosis=no_approved_candidates")
		}
		return
	}
	fmt.Printf("approved=%d rejected=%d openPositions=%d\n", summary.Approved, summary.Rejected, summary.OpenCount)
}

func rejectCount(reasons map[string]int, keys ...string) int {
	total := 0
	for _, key := range keys {
		total += reasons[key]
	}
	return total
}

func printRejectSummary(reasons map[string]int, limit int) {
	fmt.Println()
	fmt.Println("=== REJECT SUMMARY ===")
	if len(reasons) == 0 {
		fmt.Println("reject summary: none")
		return
	}
	type row struct {
		Reason string
		Count  int
	}
	rows := make([]row, 0, len(reasons))
	for reason, count := range reasons {
		if reason != "" && count > 0 {
			rows = append(rows, row{Reason: reason, Count: count})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count == rows[j].Count {
			return rows[i].Reason < rows[j].Reason
		}
		return rows[i].Count > rows[j].Count
	})
	if limit <= 0 || limit > len(rows) {
		limit = len(rows)
	}
	for i := 0; i < limit; i++ {
		fmt.Printf("%s: %d\n", rows[i].Reason, rows[i].Count)
	}
}

func printApprovedBlocked(outcomes []paper.CandidateOutcome, limit int) {
	if len(outcomes) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("=== CANDIDATE BLOCKS ===")
	if limit <= 0 || limit > len(outcomes) {
		limit = len(outcomes)
	}
	for i := 0; i < limit; i++ {
		outcome := outcomes[i]
		candidate := outcome.Candidate
		fmt.Printf("%d. venue=%s symbol=%s side=%s strat=%q score=%.2f confidence=%.2f\n", i+1, candidate.Venue, candidate.Symbol, candidate.Side, candidate.Strategy, candidate.Score, candidate.Confidence)
		fmt.Printf("   entry=%.2f stop=%.2f tp1=%.2f status=%s reason=%s\n", candidate.EntryPrice, candidate.StopPrice, candidate.TP1, outcome.Status, outcome.Reason)
	}
}

func printCandidateAdmission(rows []paper.CandidateAdmissionSummary) {
	fmt.Println()
	fmt.Println("=== CANDIDATE ADMISSION ===")
	if len(rows) == 0 {
		fmt.Println("candidate admission: none")
		return
	}
	for _, row := range rows {
		fmt.Printf("%s created=%d approved=%d deduped=%d portfolioBlocked=%d stateBlocked=%d riskRejected=%d\n",
			row.Venue,
			row.Created,
			row.Approved,
			row.Deduped,
			row.PortfolioBlocked,
			row.StateBlocked,
			row.RiskRejected,
		)
	}
}

func formatUnixMillis(value int64) string {
	if value <= 0 {
		return "unknown"
	}
	return time.UnixMilli(value).UTC().Format(time.RFC3339)
}

func formatDuration(value time.Duration) string {
	if value <= 0 {
		return "unknown"
	}
	if value < time.Hour {
		return value.Round(time.Second).String()
	}
	return value.Round(time.Minute).String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func buildBookTradeRules() bookTradeRulesStudy {
	archivedCount, err := research.ArchiveGeneratedResearchOutputs("research")
	if err != nil {
		log.Fatalf("archive generated research outputs: %v", err)
	}

	packet := strategy.DefaultBookTradeRulesPacket()
	if err := strategy.WriteBookTradeRulesMarkdown("docs/book_trade_rules.md", packet.Playbooks); err != nil {
		log.Fatalf("write book trade rules docs: %v", err)
	}
	if err := strategy.WriteBookTradeRulesPacketJSON("strategy/book_trade_rules_packet.json", packet); err != nil {
		log.Fatalf("write book trade rules packet json: %v", err)
	}
	if err := strategy.WriteBookTradeRulesPacketMarkdown("strategy/book_trade_rules_packet.md", packet); err != nil {
		log.Fatalf("write book trade rules packet markdown: %v", err)
	}

	return bookTradeRulesStudy{
		Packet:        packet,
		ArchivedCount: archivedCount,
	}
}

func strategyRuleCounts(playbooks []strategy.ExecutablePlaybook) (int, int, int, int, int) {
	playbookCount := len(playbooks)
	return playbookCount, playbookCount, playbookCount, playbookCount, playbookCount
}

func buildTradeTapeFetchers() []tradetape.VenueFetcher {
	return []tradetape.VenueFetcher{
		{
			Venue: "hyperliquid", Symbol: "BTC",
			Fetch: func() ([]tradetape.TradeTapePrint, error) {
				return hyperliquid.GetRecentTrades("BTC")
			},
		},
		{
			Venue: "aster", Symbol: "BTCUSDT",
			Fetch: func() ([]tradetape.TradeTapePrint, error) {
				return aster.GetRecentTrades("BTCUSDT", 100)
			},
		},
		{
			Venue: "lighter", Symbol: "BTC",
			Fetch: func() ([]tradetape.TradeTapePrint, error) {
				return lighter.GetRecentTrades("BTC")
			},
		},
	}
}

func runResearchHarness() {
	venues := fetchVenueCandles(ResearchInterval, ResearchCandleCap)
	for _, venue := range venues {
		log.Printf("venue=%s symbol=%s candles=%d", venue.Venue, venue.Symbol, len(venue.Candles))
	}

	primary, ok := findVenue(venues, "aster")
	if !ok {
		log.Fatal("aster candles unavailable; cannot run primary chapter research")
	}

	frame := series.FromCandles(primary.Candles)

	vwapRows := buildVWAPFeatures(primary.Candles)
	vwapInteractions, vwapReactions, vwapBehaviorSummary := buildVWAPBehaviorStudy(primary.Candles)
	vwapChapter3 := buildVWAPChapter3Study(primary.Candles, fetchVWAPTimeframeCandles())
	priceAction := buildPriceActionStudy(primary.Candles)
	priceActionStrategies := buildPriceActionStrategyStudy(primary.Candles)
	priceActionPhase3 := buildPriceActionPhase3Study(primary.Candles)
	volumeProfile := buildVolumeProfileStudy(primary.Candles)
	scopedVolumeProfile := buildScopedVolumeProfileStudy(primary.Candles)
	volumeProfileShapes := buildVolumeProfileShapeStudy(scopedVolumeProfile.Profiles)
	flexibleVolumeProfiles := buildFlexibleVolumeProfileStudy(primary.Candles, priceAction.Rows, priceActionStrategies.Rows, priceActionPhase3.Rows)
	profileAcceptanceQuality := buildProfileAcceptanceQualityStudy()
	volumeSetupAccumulation := buildVolumeSetupAccumulationStudy()
	volumeSetupAccumulationQuality := buildVolumeSetupAccumulationQualityStudy()
	volumeSetupTrend := buildVolumeSetupTrendStudy()
	volumeSetupTrendQuality := buildVolumeSetupTrendQualityStudy()
	volumeSetupRejection := buildVolumeSetupRejectionStudy()
	volumeSetupRejectionQuality := buildVolumeSetupRejectionQualityStudy()
	volumeSetupReversal := buildVolumeSetupReversalStudy()
	volumeSetupReversalQuality := buildVolumeSetupReversalQualityStudy()
	volumeSetupComparison := buildVolumeSetupComparisonStudy()
	setupFeatureLabels := buildSetupFeatureLabelExport(primary.Candles)
	orderBookSnapshots := fetchMultiVenueL2Snapshots()
	hyperliquidL2Snapshot := snapshotByVenue(orderBookSnapshots, "hyperliquid")
	orderBookRows := buildOrderBookResearchRows(orderBookSnapshots, candlesByVenue(venues))
	instrumentUniverse := buildInstrumentUniverseStudy(orderBookRows.Features, venues)
	volumeProfileFinalBook := buildVolumeProfileFinalBookPacket()
	orderFlowFoundation := buildOrderFlowFoundationStudy()
	orderFlowDataCapabilityAudit := buildOrderFlowDataCapabilityAuditStudy()
	tradeTapeAnalysis, tradeTapeAnalysisOK := buildTradeTapeAnalysis()
	orderFlowRealFootprints, orderFlowRealFootprintsOK := buildOrderFlowRealFootprints()
	orderFlowVolumeClusters, orderFlowVolumeClustersOK := buildOrderFlowVolumeClusters(orderFlowRealFootprints.Rows, orderFlowRealFootprintsOK)
	orderFlowVolumeClusterQuality, orderFlowVolumeClusterQualityOK := buildOrderFlowVolumeClusterQuality(orderFlowVolumeClustersOK)
	orderFlowMultipleHVNs, orderFlowMultipleHVNsOK := buildOrderFlowMultipleHVNs(orderFlowRealFootprints.Rows, orderFlowRealFootprintsOK)
	orderFlowBackfilledFootprints, orderFlowBackfilledFootprintsOK := buildOrderFlowBackfilledFootprints()
	orderFlowBackfilledReplay, orderFlowBackfilledReplayOK := buildOrderFlowBackfilledSetupReplay(
		orderFlowBackfilledFootprints.Rows,
		orderFlowBackfilledFootprints.Summary,
		orderFlowBackfilledFootprintsOK,
		tradeTapeAnalysis.Rows,
		orderFlowRealFootprints.Summary,
		orderFlowVolumeClusters.Summary,
		orderFlowMultipleHVNs.Summary,
	)
	orderFlowTradesFilter, orderFlowTradesFilterOK := buildOrderFlowTradesFilter(
		orderFlowBackfilledFootprints.Rows,
		orderFlowBackfilledReplay.Replay.MultipleHVNRows,
		orderFlowBackfilledFootprintsOK && orderFlowBackfilledReplayOK,
		orderFlowBackfilledReplay.Replay.VolumeClusterSummary,
		orderFlowBackfilledReplay.Replay.MultipleHVNSummary,
	)
	orderFlowStackedImbalances, orderFlowStackedImbalancesOK := buildOrderFlowStackedImbalances(
		orderFlowBackfilledFootprints.Bars,
		orderFlowBackfilledFootprints.Rows,
		orderFlowBackfilledReplay.Replay.MultipleHVNRows,
		orderFlowBackfilledFootprintsOK && orderFlowBackfilledReplayOK && orderFlowTradesFilterOK,
		orderFlowBackfilledReplay.Replay.VolumeClusterSummary,
		orderFlowBackfilledReplay.Replay.MultipleHVNSummary,
		orderFlowTradesFilter.Summary,
	)
	orderFlowUnfinishedBusiness, orderFlowUnfinishedBusinessOK := buildOrderFlowUnfinishedBusiness(
		orderFlowBackfilledFootprints.Bars,
		orderFlowBackfilledFootprints.Rows,
		orderFlowStackedImbalancesOK,
		orderFlowBackfilledReplay.Replay.VolumeClusterSummary,
		orderFlowBackfilledReplay.Replay.MultipleHVNSummary,
		orderFlowTradesFilter.Summary,
		orderFlowStackedImbalances.Summary,
	)
	orderFlowBigLimitOrders, orderFlowBigLimitOrdersOK := buildOrderFlowBigLimitOrders(
		orderFlowRealFootprints.Rows,
		orderFlowBackfilledFootprints.Rows,
		orderFlowTradesFilter.Rows,
		orderFlowRealFootprintsOK || orderFlowUnfinishedBusinessOK,
		orderFlowTradesFilter.Summary,
		orderFlowStackedImbalances.Summary,
		orderFlowUnfinishedBusiness.Summary,
	)
	orderFlowAbsorption, orderFlowAbsorptionOK := buildOrderFlowAbsorption(
		orderFlowRealFootprints.Rows,
		orderFlowBackfilledFootprints.Rows,
		orderFlowTradesFilter.Rows,
		orderFlowRealFootprintsOK || orderFlowBackfilledFootprintsOK,
		orderFlowBigLimitOrders.Summary,
		orderFlowTradesFilter.Summary,
		orderFlowStackedImbalances.Summary,
	)
	orderFlowAggressiveDelta, orderFlowAggressiveDeltaOK := buildOrderFlowAggressiveDelta(
		orderFlowRealFootprints.Rows,
		orderFlowBackfilledFootprints.Rows,
		orderFlowTradesFilter.Rows,
		orderFlowRealFootprintsOK || orderFlowBackfilledFootprintsOK,
	)
	orderFlowCumulativeDeltaDivergence, orderFlowCumulativeDeltaDivergenceOK := buildOrderFlowCumulativeDeltaDivergence(
		orderFlowRealFootprints.Rows,
		orderFlowBackfilledFootprints.Rows,
		orderFlowTradesFilter.Rows,
		orderFlowRealFootprintsOK || orderFlowBackfilledFootprintsOK,
	)
	orderFlowConfirmationComparison, orderFlowConfirmationComparisonOK := buildOrderFlowConfirmationComparison(
		orderFlowBigLimitOrders.Summary,
		orderFlowAbsorption.Summary,
		orderFlowAggressiveDelta.Summary,
		orderFlowCumulativeDeltaDivergence.Summary,
		orderFlowBigLimitOrdersOK && orderFlowAbsorptionOK && orderFlowAggressiveDeltaOK && orderFlowCumulativeDeltaDivergenceOK,
	)
	vwapL2Refresh := buildVWAPL2RefreshStudy(venues, orderBookRows.Snapshots)
	l2SnapshotAnalysis, l2SnapshotAnalysisOK := buildL2SnapshotAnalysis()
	chapter2Results := runChapter2(primary, frame)
	featureRows, labelRows, trainingRows := runChapter3(primary, frame)
	chapter4SingleResults, chapter4VenueRows, chapter4VenueAnalyses, pairCandidates := runChapter4(venues, primary, frame)
	chapter5Results, statarbCandidates := runChapter5(venues, primary, frame)
	riskRows := buildChapter6RiskMetrics(chapter2Results, chapter4SingleResults, chapter5Results)
	riskSummary := buildChapter6RiskSummary(riskRows)
	riskMetricsPath := "research/chapter6_risk_metrics.csv"
	if err := research.WriteChapter6RiskMetricsCSV(riskMetricsPath, riskRows); err != nil {
		log.Fatalf("write chapter 6 risk metrics: %v", err)
	}
	riskSummaryPath := "research/chapter6_risk_summary.json"
	if err := research.WriteChapter6RiskSummaryJSON(riskSummaryPath, riskSummary); err != nil {
		log.Fatalf("write chapter 6 risk summary: %v", err)
	}
	filterDecisions := research.EvaluateChapter6RiskFilters(riskRows, riskmetrics.DefaultRiskFilterConfig())
	filterCSVPath := "research/chapter6_risk_filters.csv"
	if err := research.WriteChapter6RiskFiltersCSV(filterCSVPath, filterDecisions); err != nil {
		log.Fatalf("write chapter 6 risk filters: %v", err)
	}
	filterSummary := research.NewChapter6RiskFilterSummary(filterDecisions)
	filterSummaryPath := "research/chapter6_risk_filters_summary.json"
	if err := research.WriteChapter6RiskFilterSummaryJSON(filterSummaryPath, filterSummary); err != nil {
		log.Fatalf("write chapter 6 risk filter summary: %v", err)
	}
	riskRankings := research.BuildChapter6RiskRanking(riskRows, filterDecisions)
	rankingCSVPath := "research/chapter6_risk_ranking.csv"
	if err := research.WriteChapter6RiskRankingCSV(rankingCSVPath, riskRankings); err != nil {
		log.Fatalf("write chapter 6 risk ranking: %v", err)
	}
	rankingSummary := research.NewChapter6RiskRankingSummary(riskRankings)
	rankingSummaryPath := "research/chapter6_risk_ranking_summary.json"
	if err := research.WriteChapter6RiskRankingSummaryJSON(rankingSummaryPath, rankingSummary); err != nil {
		log.Fatalf("write chapter 6 risk ranking summary: %v", err)
	}
	promotionGates := research.BuildChapter6PromotionGates(riskRankings, riskmetrics.DefaultPromotionConfig())
	promotionCSVPath := "research/chapter6_promotion_gates.csv"
	if err := research.WriteChapter6PromotionGatesCSV(promotionCSVPath, promotionGates); err != nil {
		log.Fatalf("write chapter 6 promotion gates: %v", err)
	}
	promotionSummary := research.NewChapter6PromotionGatesSummary(promotionGates)
	promotionSummaryPath := "research/chapter6_promotion_gates_summary.json"
	if err := research.WriteChapter6PromotionGatesSummaryJSON(promotionSummaryPath, promotionSummary); err != nil {
		log.Fatalf("write chapter 6 promotion gates summary: %v", err)
	}
	chapter6Packet := research.BuildChapter6RiskPacket(riskRankings, promotionGates)
	packetJSONPath := "research/chapter6_packet.json"
	if err := research.WriteChapter6RiskPacketJSON(packetJSONPath, chapter6Packet); err != nil {
		log.Fatalf("write chapter 6 packet json: %v", err)
	}
	packetMarkdownPath := "research/chapter6_packet.md"
	if err := research.WriteChapter6RiskPacketMarkdown(packetMarkdownPath, chapter6Packet); err != nil {
		log.Fatalf("write chapter 6 packet markdown: %v", err)
	}

	summary := buildSummary(
		len(primary.Candles),
		chapter2Results,
		chapter4SingleResults,
		chapter5Results,
		len(featureRows),
		len(labelRows),
		len(trainingRows),
		len(pairCandidates)+len(statarbCandidates),
	)
	summaryPath := "research/research_summary.json"
	if err := research.WriteResearchSummaryJSON(summaryPath, summary); err != nil {
		log.Fatalf("write research summary: %v", err)
	}

	fmt.Println("=== 30 DAY RESEARCH VALIDATION ===")
	fmt.Println()
	fmt.Printf("candles=%d\n", len(primary.Candles))
	fmt.Println()
	fmt.Println("CHAPTER 2")
	fmt.Printf("bestStrategy=%s\n", summary.BestChapter2Strategy)
	fmt.Printf("netPnL=%.2f\n", summary.BestChapter2NetPnL)
	fmt.Println()
	fmt.Println("CHAPTER 4")
	fmt.Printf("bestStrategy=%s\n", summary.BestChapter4Strategy)
	fmt.Printf("netPnL=%.2f\n", summary.BestChapter4NetPnL)
	fmt.Println()
	fmt.Println("CHAPTER 5")
	fmt.Printf("bestStrategy=%s\n", summary.BestChapter5Strategy)
	fmt.Printf("netPnL=%.2f\n", summary.BestChapter5NetPnL)
	fmt.Println()
	fmt.Printf("worstDrawdown=%s %.2f\n", summary.WorstDrawdownStrategy, summary.WorstDrawdown)
	fmt.Printf("mostTrades=%s %d\n", summary.MostActiveStrategy, summary.MostTrades)
	fmt.Println()
	fmt.Printf("featureRows=%d\n", summary.FeatureRows)
	fmt.Printf("labelRows=%d\n", summary.LabelRows)
	fmt.Printf("trainingRows=%d\n", summary.TrainingRows)
	fmt.Printf("vwapRows=%d\n", len(vwapRows))
	fmt.Println()
	fmt.Printf("pairCandidates=%d\n", summary.PairCandidates)
	fmt.Println()
	fmt.Println("wrote " + summaryPath)
	fmt.Println()
	fmt.Println("=== VWAP CHAPTER 2 BEHAVIOR STUDY ===")
	fmt.Printf("candles=%d\n", vwapBehaviorSummary.Candles)
	fmt.Printf("touches=%d\n", vwapBehaviorSummary.Touches)
	fmt.Printf("crosses=%d\n", vwapBehaviorSummary.Crosses)
	fmt.Printf("bounces=%d\n", vwapBehaviorSummary.Bounces)
	fmt.Printf("breaks=%d\n", vwapBehaviorSummary.Breaks)
	fmt.Printf("chops=%d\n", vwapBehaviorSummary.Chops)
	fmt.Printf("bounceRate=%.2f\n", vwapBehaviorSummary.BounceRate)
	fmt.Printf("breakRate=%.2f\n", vwapBehaviorSummary.BreakRate)
	fmt.Printf("interactionRows=%d\n", len(vwapInteractions))
	fmt.Printf("reactionRows=%d\n", len(vwapReactions))
	fmt.Println("wrote research/vwap_interaction.csv")
	fmt.Println("wrote research/vwap_reactions.csv")
	fmt.Println("wrote research/vwap_behavior_summary.json")
	fmt.Println()
	fmt.Println("=== VWAP CHAPTER 3 CONTEXT STUDY ===")
	fmt.Printf("candles=%d\n", vwapChapter3.Summary.TotalCandles)
	fmt.Printf("bullAlignments=%d\n", vwapChapter3.Summary.BullAlignmentCount)
	fmt.Printf("bearAlignments=%d\n", vwapChapter3.Summary.BearAlignmentCount)
	fmt.Printf("mixedAlignments=%d\n", vwapChapter3.Summary.MixedCount)
	fmt.Printf("strongBull=%d\n", vwapChapter3.Summary.StrongBullCount)
	fmt.Printf("strongBear=%d\n", vwapChapter3.Summary.StrongBearCount)
	fmt.Printf("contextRows=%d\n", len(vwapChapter3.ContextRows))
	fmt.Printf("timeframeRows=%d\n", len(vwapChapter3.TimeframeRows))
	fmt.Printf("trendRows=%d\n", len(vwapChapter3.TrendRows))
	fmt.Println("wrote research/context_features.csv")
	fmt.Println("wrote research/timeframe_features.csv")
	fmt.Println("wrote research/trend_alignment.csv")
	fmt.Println("wrote research/chapter3_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE BOOK PRICE ACTION FOUNDATION ===")
	fmt.Printf("candles=%d\n", priceAction.Summary.Candles)
	fmt.Printf("sidewaysAreas=%d\n", priceAction.Summary.SidewaysAreas)
	fmt.Printf("bullishInitiations=%d\n", priceAction.Summary.BullishInitiations)
	fmt.Printf("bearishInitiations=%d\n", priceAction.Summary.BearishInitiations)
	fmt.Printf("bullishRejections=%d\n", priceAction.Summary.BullishRejections)
	fmt.Printf("bearishRejections=%d\n", priceAction.Summary.BearishRejections)
	fmt.Printf("failedAuctions=%d\n", priceAction.Summary.FailedAuctions)
	fmt.Println("wrote research/price_action_features.csv")
	fmt.Println("wrote research/price_action_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE BOOK PRICE ACTION STRATEGY STUDY ===")
	fmt.Printf("candles=%d\n", priceActionStrategies.Summary.Candles)
	fmt.Printf("srFlipSetups=%d\n", priceActionStrategies.Summary.SRFlipSetups)
	fmt.Printf("openDriveSetups=%d\n", priceActionStrategies.Summary.OpenDriveSetups)
	fmt.Printf("abcdSetups=%d\n", priceActionStrategies.Summary.ABCDSetups)
	fmt.Printf("sessionOpenSetups=%d\n", priceActionStrategies.Summary.SessionOpenSetups)
	fmt.Printf("dailyOpenSetups=%d\n", priceActionStrategies.Summary.DailyOpenSetups)
	fmt.Println("wrote research/price_action_strategy_study.csv")
	fmt.Println("wrote research/price_action_strategy_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE BOOK PRICE ACTION PHASE 3 ===")
	fmt.Printf("candles=%d\n", priceActionPhase3.Summary.Candles)
	fmt.Printf("dailyHighRetests=%d\n", priceActionPhase3.Summary.DailyHighRetests)
	fmt.Printf("dailyLowRetests=%d\n", priceActionPhase3.Summary.DailyLowRetests)
	fmt.Printf("weeklyHighRetests=%d\n", priceActionPhase3.Summary.WeeklyHighRetests)
	fmt.Printf("weeklyLowRetests=%d\n", priceActionPhase3.Summary.WeeklyLowRetests)
	fmt.Printf("strongHighs=%d\n", priceActionPhase3.Summary.StrongHighs)
	fmt.Printf("strongLows=%d\n", priceActionPhase3.Summary.StrongLows)
	fmt.Printf("weakHighs=%d\n", priceActionPhase3.Summary.WeakHighs)
	fmt.Printf("weakLows=%d\n", priceActionPhase3.Summary.WeakLows)
	fmt.Printf("failedHighAuctions=%d\n", priceActionPhase3.Summary.FailedHighAuctions)
	fmt.Printf("failedLowAuctions=%d\n", priceActionPhase3.Summary.FailedLowAuctions)
	fmt.Println("wrote research/price_action_phase3_study.csv")
	fmt.Println("wrote research/price_action_phase3_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE FOUNDATION ===")
	fmt.Printf("profiles=%d\n", volumeProfile.Summary.Profiles)
	fmt.Printf("bins=%d\n", volumeProfile.Summary.Bins)
	fmt.Printf("poc=%.2f\n", volumeProfile.Summary.POC)
	fmt.Printf("vah=%.2f\n", volumeProfile.Summary.VAH)
	fmt.Printf("val=%.2f\n", volumeProfile.Summary.VAL)
	fmt.Printf("hvns=%d\n", volumeProfile.Summary.HVNs)
	fmt.Printf("lvns=%d\n", volumeProfile.Summary.LVNs)
	fmt.Printf("shape=%s\n", volumeProfile.Profile.Shape)
	fmt.Println("wrote research/volume_profile_features.csv")
	fmt.Println("wrote research/volume_profile_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE SCOPED PROFILES ===")
	fmt.Printf("dailyPOC=%.2f\n", scopedPOC(scopedVolumeProfile.Profiles, volumeprofile.ScopeDailySession))
	fmt.Printf("rolling3dPOC=%.2f\n", scopedPOC(scopedVolumeProfile.Profiles, volumeprofile.ScopeRolling3D))
	fmt.Printf("rolling7dPOC=%.2f\n", scopedPOC(scopedVolumeProfile.Profiles, volumeprofile.ScopeRolling7D))
	fmt.Printf("composite30dPOC=%.2f\n", scopedPOC(scopedVolumeProfile.Profiles, volumeprofile.ScopeComposite30D))
	fmt.Println("wrote research/volume_profile_scoped_features.csv")
	fmt.Println("wrote research/volume_profile_scoped_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE SHAPE STUDY ===")
	fmt.Printf("dailyShape=%s\n", volumeProfileShapes.Summary.Daily.Shape)
	fmt.Printf("rolling3dShape=%s\n", volumeProfileShapes.Summary.Rolling3D.Shape)
	fmt.Printf("rolling7dShape=%s\n", volumeProfileShapes.Summary.Rolling7D.Shape)
	fmt.Printf("composite30dShape=%s\n", volumeProfileShapes.Summary.Composite30D.Shape)
	fmt.Println("wrote research/volume_profile_shape_study.csv")
	fmt.Println("wrote research/volume_profile_shape_summary.json")
	fmt.Println()
	fmt.Println("=== FLEXIBLE VOLUME PROFILE ===")
	fmt.Printf("sidewaysProfiles=%d\n", flexibleVolumeProfiles.Summary.SidewaysAccumulationProfiles)
	fmt.Printf("openDriveProfiles=%d\n", flexibleVolumeProfiles.Summary.OpenDriveProfiles)
	fmt.Printf("failedAuctionProfiles=%d\n", flexibleVolumeProfiles.Summary.FailedAuctionProfiles)
	fmt.Printf("highLowRetestProfiles=%d\n", flexibleVolumeProfiles.Summary.HighLowRetestProfiles)
	fmt.Printf("acceptedProfiles=%d\n", flexibleVolumeProfiles.Summary.AcceptedProfiles)
	fmt.Printf("rejectedProfiles=%d\n", flexibleVolumeProfiles.Summary.RejectedProfiles)
	fmt.Println("wrote research/flexible_volume_profile.csv")
	fmt.Println("wrote research/flexible_volume_profile_summary.json")
	fmt.Println()
	fmt.Println("=== PROFILE ACCEPTANCE QUALITY ===")
	fmt.Printf("bestProfileType=%s\n", profileAcceptanceQuality.Summary.BestProfileType)
	fmt.Printf("bestProfileAcceptanceRate=%.2f\n", profileAcceptanceQuality.Summary.BestProfileAcceptanceRate)
	fmt.Printf("bestShape=%s\n", profileAcceptanceQuality.Summary.BestShape)
	fmt.Printf("bestShapeAcceptanceRate=%.2f\n", profileAcceptanceQuality.Summary.BestShapeAcceptanceRate)
	fmt.Printf("acceptedProfiles=%d\n", profileAcceptanceQuality.Summary.AcceptedProfiles)
	fmt.Printf("rejectedProfiles=%d\n", profileAcceptanceQuality.Summary.RejectedProfiles)
	fmt.Println("wrote research/profile_acceptance_quality.csv")
	fmt.Println("wrote research/profile_acceptance_quality_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME SETUP #1 ACCUMULATION STUDY ===")
	fmt.Printf("setups=%d\n", volumeSetupAccumulation.Summary.Setups)
	fmt.Printf("longContexts=%d\n", volumeSetupAccumulation.Summary.LongContexts)
	fmt.Printf("shortContexts=%d\n", volumeSetupAccumulation.Summary.ShortContexts)
	fmt.Printf("accepted=%d\n", volumeSetupAccumulation.Summary.Accepted)
	fmt.Printf("rejected=%d\n", volumeSetupAccumulation.Summary.Rejected)
	fmt.Printf("retests=%d\n", volumeSetupAccumulation.Summary.Retests)
	fmt.Printf("dailyPocConfluence=%d\n", volumeSetupAccumulation.Summary.DailyPOCConfluence)
	fmt.Printf("rolling3dPocConfluence=%d\n", volumeSetupAccumulation.Summary.Rolling3DPOCConfluence)
	fmt.Printf("rolling7dPocConfluence=%d\n", volumeSetupAccumulation.Summary.Rolling7DPOCConfluence)
	fmt.Printf("composite30dPocConfluence=%d\n", volumeSetupAccumulation.Summary.Composite30DPOCConfluence)
	fmt.Println("wrote research/volume_setup_accumulation.csv")
	fmt.Println("wrote research/volume_setup_accumulation_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME SETUP #1 QUALITY ===")
	fmt.Printf("totalSetups=%d\n", volumeSetupAccumulationQuality.Summary.TotalSetups)
	fmt.Printf("bestConfluenceGroup=%s\n", volumeSetupAccumulationQuality.Summary.BestConfluenceGroup)
	fmt.Printf("bestConfluenceAcceptanceRate=%.2f\n", volumeSetupAccumulationQuality.Summary.BestConfluenceAcceptanceRate)
	fmt.Printf("bestShape=%s\n", volumeSetupAccumulationQuality.Summary.BestShape)
	fmt.Printf("bestShapeAcceptanceRate=%.2f\n", volumeSetupAccumulationQuality.Summary.BestShapeAcceptanceRate)
	fmt.Printf("bestDirection=%s\n", volumeSetupAccumulationQuality.Summary.BestDirection)
	fmt.Printf("bestDirectionFollowThrough20=%.2f\n", volumeSetupAccumulationQuality.Summary.BestDirectionAverageFollowThrough20)
	fmt.Println("wrote research/volume_setup_accumulation_quality.csv")
	fmt.Println("wrote research/volume_setup_accumulation_quality_summary.json")
	fmt.Println("wrote research/volume_setup_accumulation_quality.md")
	fmt.Println()
	fmt.Println("=== VOLUME SETUP #2 TREND STUDY ===")
	fmt.Printf("setups=%d\n", volumeSetupTrend.Summary.Setups)
	fmt.Printf("longContexts=%d\n", volumeSetupTrend.Summary.LongContexts)
	fmt.Printf("shortContexts=%d\n", volumeSetupTrend.Summary.ShortContexts)
	fmt.Printf("accepted=%d\n", volumeSetupTrend.Summary.Accepted)
	fmt.Printf("rejected=%d\n", volumeSetupTrend.Summary.Rejected)
	fmt.Printf("pocRetests=%d\n", volumeSetupTrend.Summary.POCRetests)
	fmt.Printf("hvnRetests=%d\n", volumeSetupTrend.Summary.HVNRetests)
	fmt.Printf("vwapAligned=%d\n", volumeSetupTrend.Summary.VWAPAligned)
	fmt.Printf("bidPressureAligned=%d\n", volumeSetupTrend.Summary.BidPressureAligned)
	fmt.Printf("askPressureAligned=%d\n", volumeSetupTrend.Summary.AskPressureAligned)
	fmt.Printf("averageTrendStrength=%.2f\n", volumeSetupTrend.Summary.AverageTrendStrength)
	fmt.Println("wrote research/volume_setup_trend.csv")
	fmt.Println("wrote research/volume_setup_trend_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME SETUP #2 QUALITY ===")
	fmt.Printf("totalSetups=%d\n", volumeSetupTrendQuality.Summary.TotalSetups)
	fmt.Printf("bestDirection=%s\n", volumeSetupTrendQuality.Summary.BestDirection)
	fmt.Printf("bestDirectionFollowThrough20=%.2f\n", volumeSetupTrendQuality.Summary.BestDirectionFollowThrough20)
	fmt.Printf("bestShape=%s\n", volumeSetupTrendQuality.Summary.BestShape)
	fmt.Printf("bestShapeAcceptanceRate=%.2f\n", volumeSetupTrendQuality.Summary.BestShapeAcceptanceRate)
	fmt.Printf("bestTrendStrengthBucket=%s\n", volumeSetupTrendQuality.Summary.BestTrendStrengthBucket)
	fmt.Printf("bestTrendStrengthAcceptanceRate=%.2f\n", volumeSetupTrendQuality.Summary.BestTrendStrengthAcceptanceRate)
	fmt.Printf("bestConfluence=%s\n", volumeSetupTrendQuality.Summary.BestConfluence)
	fmt.Printf("bestConfluenceAcceptanceRate=%.2f\n", volumeSetupTrendQuality.Summary.BestConfluenceAcceptanceRate)
	fmt.Println("wrote research/volume_setup_trend_quality.csv")
	fmt.Println("wrote research/volume_setup_trend_quality_summary.json")
	fmt.Println("wrote research/volume_setup_trend_quality.md")
	fmt.Println()
	fmt.Println("=== VOLUME SETUP #3 REJECTION STUDY ===")
	fmt.Printf("setups=%d\n", volumeSetupRejection.Summary.Setups)
	fmt.Printf("longContexts=%d\n", volumeSetupRejection.Summary.LongContexts)
	fmt.Printf("shortContexts=%d\n", volumeSetupRejection.Summary.ShortContexts)
	fmt.Printf("accepted=%d\n", volumeSetupRejection.Summary.Accepted)
	fmt.Printf("rejected=%d\n", volumeSetupRejection.Summary.Rejected)
	fmt.Printf("pocRetests=%d\n", volumeSetupRejection.Summary.POCRetests)
	fmt.Printf("hvnRetests=%d\n", volumeSetupRejection.Summary.HVNRetests)
	fmt.Printf("vwapAligned=%d\n", volumeSetupRejection.Summary.VWAPAligned)
	fmt.Printf("bidPressureAligned=%d\n", volumeSetupRejection.Summary.BidPressureAligned)
	fmt.Printf("askPressureAligned=%d\n", volumeSetupRejection.Summary.AskPressureAligned)
	fmt.Printf("dailyPocConfluence=%d\n", volumeSetupRejection.Summary.DailyPOCConfluence)
	fmt.Printf("rolling3dPocConfluence=%d\n", volumeSetupRejection.Summary.Rolling3DPOCConfluence)
	fmt.Printf("rolling7dPocConfluence=%d\n", volumeSetupRejection.Summary.Rolling7DPOCConfluence)
	fmt.Printf("composite30dPocConfluence=%d\n", volumeSetupRejection.Summary.Composite30DPOCConfluence)
	fmt.Println("wrote research/volume_setup_rejection.csv")
	fmt.Println("wrote research/volume_setup_rejection_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME SETUP #3 QUALITY ===")
	fmt.Printf("totalSetups=%d\n", volumeSetupRejectionQuality.Summary.TotalSetups)
	fmt.Printf("bestDirection=%s\n", volumeSetupRejectionQuality.Summary.BestDirection)
	fmt.Printf("bestDirectionFollowThrough20=%.2f\n", volumeSetupRejectionQuality.Summary.BestDirectionFollowThrough20)
	fmt.Printf("bestShape=%s\n", volumeSetupRejectionQuality.Summary.BestShape)
	fmt.Printf("bestShapeAcceptanceRate=%.2f\n", volumeSetupRejectionQuality.Summary.BestShapeAcceptanceRate)
	fmt.Printf("bestConfluence=%s\n", volumeSetupRejectionQuality.Summary.BestConfluence)
	fmt.Printf("bestConfluenceAcceptanceRate=%.2f\n", volumeSetupRejectionQuality.Summary.BestConfluenceAcceptanceRate)
	fmt.Printf("bestFilter=%s\n", volumeSetupRejectionQuality.Summary.BestFilter)
	fmt.Printf("bestFilterFollowThrough20=%.2f\n", volumeSetupRejectionQuality.Summary.BestFilterFollowThrough20)
	fmt.Println("wrote research/volume_setup_rejection_quality.csv")
	fmt.Println("wrote research/volume_setup_rejection_quality_summary.json")
	fmt.Println("wrote research/volume_setup_rejection_quality.md")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE REVERSAL TRADE STUDY ===")
	fmt.Printf("setups=%d\n", volumeSetupReversal.Summary.Setups)
	fmt.Printf("longContexts=%d\n", volumeSetupReversal.Summary.LongContexts)
	fmt.Printf("shortContexts=%d\n", volumeSetupReversal.Summary.ShortContexts)
	fmt.Printf("accepted=%d\n", volumeSetupReversal.Summary.Accepted)
	fmt.Printf("rejected=%d\n", volumeSetupReversal.Summary.Rejected)
	fmt.Printf("neutral=%d\n", volumeSetupReversal.Summary.Neutral)
	fmt.Printf("vwapAligned=%d\n", volumeSetupReversal.Summary.VWAPAligned)
	fmt.Printf("pocConfluence=%d\n", volumeSetupReversal.Summary.POCConfluence)
	fmt.Printf("hvnConfluence=%d\n", volumeSetupReversal.Summary.HVNConfluence)
	fmt.Printf("vahValRejections=%d\n", volumeSetupReversal.Summary.VAHVALRejections)
	fmt.Printf("averageFollowThrough20=%.2f\n", volumeSetupReversal.Summary.AverageFollowThrough20)
	fmt.Println("wrote research/volume_setup_reversal.csv")
	fmt.Println("wrote research/volume_setup_reversal_summary.json")
	fmt.Println()
	fmt.Println("=== REVERSAL LABEL AUDIT ===")
	fmt.Println("rootCause=Reversal setup rows did not expose accepted/rejected outcome fields, so comparison counted zero outcomes.")
	fmt.Printf("accepted=%d\n", volumeSetupReversal.Summary.Accepted)
	fmt.Printf("rejected=%d\n", volumeSetupReversal.Summary.Rejected)
	fmt.Printf("neutral=%d\n", volumeSetupReversal.Summary.Neutral)
	fmt.Println()
	fmt.Println("=== REVERSAL QUALITY ===")
	fmt.Printf("bestDirection=%s\n", volumeSetupReversalQuality.Summary.BestDirection)
	fmt.Printf("bestConfluence=%s\n", volumeSetupReversalQuality.Summary.BestConfluence)
	fmt.Printf("bestFT20=%.2f\n", volumeSetupReversalQuality.Summary.BestFilterFollowThrough20)
	fmt.Println("wrote research/volume_setup_reversal_quality.csv")
	fmt.Println("wrote research/volume_setup_reversal_quality_summary.json")
	fmt.Println("wrote research/volume_setup_reversal_quality.md")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE SETUP COMPARISON ===")
	fmt.Printf("bestAcceptanceSetup=%s\n", volumeSetupComparison.Summary.BestAcceptanceSetup)
	fmt.Printf("bestAcceptanceRate=%.2f\n", volumeSetupComparison.Summary.BestAcceptanceRate)
	fmt.Printf("bestFT20Setup=%s\n", volumeSetupComparison.Summary.BestFT20Setup)
	fmt.Printf("bestVWAPSetup=%s\n", volumeSetupComparison.Summary.BestVWAPSetup)
	fmt.Printf("bestPOCSetup=%s\n", volumeSetupComparison.Summary.BestPOCSetup)
	fmt.Printf("bestHVNSetup=%s\n", volumeSetupComparison.Summary.BestHVNSetup)
	fmt.Printf("largestSampleSetup=%s\n", volumeSetupComparison.Summary.LargestSampleSetup)
	fmt.Println("wrote research/volume_setup_comparison.csv")
	fmt.Println("wrote research/volume_setup_comparison.json")
	fmt.Println("wrote research/volume_setup_comparison.md")
	fmt.Println()
	fmt.Println("=== UPDATED SETUP COMPARISON ===")
	fmt.Printf("bestAcceptanceSetup=%s\n", volumeSetupComparison.Summary.BestAcceptanceSetup)
	fmt.Printf("bestFT20Setup=%s\n", volumeSetupComparison.Summary.BestFT20Setup)
	fmt.Printf("bestVWAPSetup=%s\n", volumeSetupComparison.Summary.BestVWAPSetup)
	fmt.Println()
	fmt.Println("=== SETUP FEATURE/LABEL EXPORT ===")
	fmt.Printf("rows=%d\n", setupFeatureLabels.Summary.Rows)
	fmt.Printf("reversalRows=%d\n", setupFeatureLabels.Summary.BySetupType["REVERSAL"])
	fmt.Printf("accumulationRows=%d\n", setupFeatureLabels.Summary.BySetupType["ACCUMULATION"])
	fmt.Printf("trendRows=%d\n", setupFeatureLabels.Summary.BySetupType["TREND"])
	fmt.Printf("rejectionRows=%d\n", setupFeatureLabels.Summary.BySetupType["REJECTION"])
	fmt.Printf("upperHits=%d\n", setupFeatureLabels.Summary.TripleBarrierLabels["upper_hit"])
	fmt.Printf("lowerHits=%d\n", setupFeatureLabels.Summary.TripleBarrierLabels["lower_hit"])
	fmt.Printf("timeExpired=%d\n", setupFeatureLabels.Summary.TripleBarrierLabels["time_expired"])
	fmt.Println("wrote research/setup_features_labels.csv")
	fmt.Println("wrote research/setup_labels_summary.json")
	fmt.Println()
	fmt.Println("=== INSTRUMENT UNIVERSE RESEARCH SCANNER ===")
	fmt.Printf("venues=%d\n", instrumentUniverse.Summary.Venues)
	fmt.Printf("symbols=%d\n", instrumentUniverse.Summary.Symbols)
	fmt.Printf("crypto=%d\n", instrumentUniverse.Summary.Crypto)
	fmt.Printf("rwa=%d\n", instrumentUniverse.Summary.RWA)
	fmt.Printf("topInstrument=%s\n", instrumentUniverse.Summary.TopInstrument)
	fmt.Println("wrote research/instrument_universe.csv")
	fmt.Println("wrote research/instrument_universe_summary.json")
	fmt.Println()
	fmt.Println("=== VOLUME PROFILE FINAL BOOK PACKET ===")
	fmt.Println("status=complete")
	fmt.Printf("docs=%d\n", len(volumeProfileFinalBook.Packet.CompletedDocs))
	fmt.Printf("archivedAudits=%d\n", volumeProfileFinalBook.Packet.ArchivedAudits)
	fmt.Println("packet=research/volume_profile_final_book_packet.md")
	fmt.Println()
	fmt.Println("=== ORDER FLOW PHASE 1 FOUNDATION ===")
	fmt.Printf("bars=%d\n", orderFlowFoundation.Summary.Bars)
	fmt.Printf("validBars=%d\n", orderFlowFoundation.Summary.ValidBars)
	fmt.Printf("totalDelta=%.2f\n", orderFlowFoundation.Summary.TotalDelta)
	fmt.Printf("finalCumulativeDelta=%.2f\n", orderFlowFoundation.Summary.FinalCumulativeDelta)
	fmt.Printf("volumeClusters=%d\n", orderFlowFoundation.Summary.VolumeClusters)
	fmt.Printf("imbalances=%d\n", orderFlowFoundation.Summary.Imbalances)
	fmt.Printf("stackedImbalances=%d\n", orderFlowFoundation.Summary.StackedImbalances)
	fmt.Printf("unfinishedBusinessHigh=%d\n", orderFlowFoundation.Summary.UnfinishedBusinessHigh)
	fmt.Printf("unfinishedBusinessLow=%d\n", orderFlowFoundation.Summary.UnfinishedBusinessLow)
	fmt.Printf("largeTrades=%d\n", orderFlowFoundation.Summary.LargeTrades)
	fmt.Println("wrote research/orderflow_foundation.csv")
	fmt.Println("wrote research/orderflow_foundation_summary.json")
	fmt.Println()
	fmt.Println("=== ORDER FLOW DATA CAPABILITY AUDIT ===")
	fmt.Printf("venues=%d\n", len(orderFlowDataCapabilityAudit.Report.Venues))
	fmt.Printf("footprintReady=%d\n", research.OrderFlowFootprintReadyCount(orderFlowDataCapabilityAudit.Report))
	fmt.Printf("tradeTapeReady=%d\n", research.OrderFlowTradeTapeReadyCount(orderFlowDataCapabilityAudit.Report))
	fmt.Printf("recommendedNextPhase=%s\n", orderFlowDataCapabilityAudit.Report.NextPhase)
	fmt.Printf("lighterTradeTapeStatus=%s\n", research.LighterTradeTapeStatus(orderFlowDataCapabilityAudit.Report))
	fmt.Println("wrote research/orderflow_data_capability_audit.md")
	fmt.Println("wrote research/orderflow_data_capability_audit.json")
	fmt.Println("wrote research/lighter_trade_tape_adapter_plan.md")
	fmt.Println()
	if tradeTapeAnalysisOK {
		fmt.Println("=== TRADE TAPE ANALYSIS ===")
		fmt.Printf("rows=%d\n", tradeTapeAnalysis.Rows)
		fmt.Printf("validRows=%d\n", tradeTapeAnalysis.ValidRows)
		fmt.Printf("venues=%d\n", tradeTapeAnalysis.Venues)
		fmt.Printf("venueSymbols=%d\n", tradeTapeAnalysis.VenueSymbols)
		fmt.Printf("canonicalSymbols=%d\n", tradeTapeAnalysis.CanonicalSymbols)
		fmt.Printf("totalVolume=%.8f\n", tradeTapeAnalysis.TotalVolume)
		fmt.Printf("largestTrade=%.8f\n", tradeTapeAnalysis.LargestTrade)
		fmt.Println("wrote research/trade_tape_analysis.json")
		fmt.Println("wrote research/trade_tape_analysis.md")
		fmt.Println()
	}
	if orderFlowRealFootprintsOK {
		fmt.Println("=== ORDER FLOW PHASE 2B REAL FOOTPRINTS ===")
		fmt.Printf("bars=%d\n", orderFlowRealFootprints.Summary.Bars)
		fmt.Printf("venues=%d\n", orderFlowRealFootprints.Summary.Venues)
		fmt.Printf("venueSymbols=%d\n", orderFlowRealFootprints.Summary.VenueSymbols)
		fmt.Printf("canonicalSymbols=%d\n", orderFlowRealFootprints.Summary.CanonicalSymbols)
		fmt.Printf("volumeClusters=%d\n", orderFlowRealFootprints.Summary.VolumeClusters)
		fmt.Printf("imbalances=%d\n", orderFlowRealFootprints.Summary.Imbalances)
		fmt.Printf("stackedImbalances=%d\n", orderFlowRealFootprints.Summary.StackedImbalances)
		fmt.Printf("unfinishedBusinessHigh=%d\n", orderFlowRealFootprints.Summary.UnfinishedBusinessHigh)
		fmt.Printf("unfinishedBusinessLow=%d\n", orderFlowRealFootprints.Summary.UnfinishedBusinessLow)
		fmt.Println("wrote research/orderflow_real_footprints.csv")
		fmt.Println("wrote research/orderflow_real_footprints_summary.json")
		fmt.Println("wrote research/orderflow_real_footprints.md")
		fmt.Println()
	}
	if orderFlowVolumeClustersOK {
		fmt.Println("=== ORDER FLOW SETUP #1 VOLUME CLUSTERS ===")
		fmt.Printf("setups=%d\n", orderFlowVolumeClusters.Summary.Setups)
		fmt.Printf("longContexts=%d\n", orderFlowVolumeClusters.Summary.LongContexts)
		fmt.Printf("shortContexts=%d\n", orderFlowVolumeClusters.Summary.ShortContexts)
		fmt.Printf("accepted=%d\n", orderFlowVolumeClusters.Summary.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowVolumeClusters.Summary.Rejected)
		fmt.Printf("retests=%d\n", orderFlowVolumeClusters.Summary.Retests)
		fmt.Printf("bestDirection=%s\n", orderFlowVolumeClusters.Summary.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowVolumeClusters.Summary.BestDirectionFT20)
		fmt.Println("wrote research/orderflow_setup_volume_cluster.csv")
		fmt.Println("wrote research/orderflow_setup_volume_cluster_summary.json")
		fmt.Println("wrote research/orderflow_setup_volume_cluster.md")
		fmt.Println()
	}
	if orderFlowVolumeClusterQualityOK {
		fmt.Println("=== ORDER FLOW VOLUME CLUSTER QUALITY ===")
		fmt.Printf("bestDirection=%s\n", orderFlowVolumeClusterQuality.Summary.BestDirection)
		fmt.Printf("bestVenue=%s\n", orderFlowVolumeClusterQuality.Summary.BestVenue)
		fmt.Printf("bestRetestGroup=%s\n", orderFlowVolumeClusterQuality.Summary.BestRetestGroup)
		fmt.Printf("acceptedClusterVolume=%.8f\n", orderFlowVolumeClusterQuality.Summary.AcceptedClusterVolume)
		fmt.Printf("acceptedDelta=%.8f\n", orderFlowVolumeClusterQuality.Summary.AcceptedDelta)
		fmt.Println("wrote research/orderflow_volume_cluster_quality.csv")
		fmt.Println("wrote research/orderflow_volume_cluster_quality.json")
		fmt.Println("wrote research/orderflow_volume_cluster_quality.md")
		fmt.Println()
	}
	if orderFlowMultipleHVNsOK {
		fmt.Println("=== ORDER FLOW SETUP #2 MULTIPLE HVNS ===")
		fmt.Printf("setups=%d\n", orderFlowMultipleHVNs.Summary.Setups)
		fmt.Printf("longContexts=%d\n", orderFlowMultipleHVNs.Summary.LongContexts)
		fmt.Printf("shortContexts=%d\n", orderFlowMultipleHVNs.Summary.ShortContexts)
		fmt.Printf("accepted=%d\n", orderFlowMultipleHVNs.Summary.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowMultipleHVNs.Summary.Rejected)
		fmt.Printf("retests=%d\n", orderFlowMultipleHVNs.Summary.Retests)
		fmt.Printf("averageHVNCount=%.2f\n", orderFlowMultipleHVNs.Summary.AverageHVNCount)
		fmt.Printf("bestDirection=%s\n", orderFlowMultipleHVNs.Summary.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowMultipleHVNs.Summary.BestDirectionFT20)
		fmt.Println("wrote research/orderflow_setup_multiple_hvn.csv")
		fmt.Println("wrote research/orderflow_setup_multiple_hvn_summary.json")
		fmt.Println("wrote research/orderflow_setup_multiple_hvn.md")
		fmt.Println()
	}
	if orderFlowBackfilledFootprintsOK {
		fmt.Println("=== ORDER FLOW PHASE 2D BACKFILLED FOOTPRINTS ===")
		fmt.Printf("source=%s\n", orderFlowBackfilledFootprints.Summary.Source)
		fmt.Printf("tradeRows=%d\n", orderFlowBackfilledFootprints.Summary.TradeRows)
		fmt.Printf("footprintBars=%d\n", orderFlowBackfilledFootprints.Summary.FootprintBars)
		fmt.Printf("volumeClusters=%d\n", orderFlowBackfilledFootprints.Summary.VolumeClusters)
		fmt.Printf("imbalances=%d\n", orderFlowBackfilledFootprints.Summary.Imbalances)
		fmt.Printf("stackedImbalances=%d\n", orderFlowBackfilledFootprints.Summary.StackedImbalances)
		fmt.Println("wrote research/orderflow_backfilled_footprints.csv")
		fmt.Println("wrote research/orderflow_backfilled_footprints_summary.json")
		fmt.Println("wrote research/orderflow_backfilled_footprints.md")
		fmt.Println()
	}
	if orderFlowBackfilledReplayOK {
		fmt.Println("=== ORDER FLOW BACKFILLED SETUP REPLAY ===")
		fmt.Printf("volumeClusterSetups=%d\n", orderFlowBackfilledReplay.Replay.VolumeClusterSummary.Setups)
		fmt.Printf("multipleHVNSetups=%d\n", orderFlowBackfilledReplay.Replay.MultipleHVNSummary.Setups)
		fmt.Printf("volumeClusterAccepted=%d\n", orderFlowBackfilledReplay.Replay.VolumeClusterSummary.Accepted)
		fmt.Printf("multipleHVNAccepted=%d\n", orderFlowBackfilledReplay.Replay.MultipleHVNSummary.Accepted)
		fmt.Println("wrote backfilled setup reports")
		fmt.Println()
		fmt.Println("=== ORDER FLOW LIVE VS BACKFILL COMPARISON ===")
		fmt.Printf("liveFootprintBars=%d\n", orderFlowBackfilledReplay.Comparison.LiveFootprintBars)
		fmt.Printf("backfilledFootprintBars=%d\n", orderFlowBackfilledReplay.Comparison.BackfilledFootprintBars)
		fmt.Printf("liveVolumeClusterSetups=%d\n", orderFlowBackfilledReplay.Comparison.LiveVolumeClusterSetups)
		fmt.Printf("backfilledVolumeClusterSetups=%d\n", orderFlowBackfilledReplay.Comparison.BackfilledVolumeClusterSetups)
		fmt.Printf("liveMultipleHVNSetups=%d\n", orderFlowBackfilledReplay.Comparison.LiveMultipleHVNSetups)
		fmt.Printf("backfilledMultipleHVNSetups=%d\n", orderFlowBackfilledReplay.Comparison.BackfilledMultipleHVNSetups)
		fmt.Println("wrote research/orderflow_live_vs_backfill_comparison.md")
		fmt.Println()
	}
	if orderFlowTradesFilterOK {
		fmt.Println("=== ORDER FLOW SETUP #3 TRADES FILTER ===")
		fmt.Printf("setups=%d\n", orderFlowTradesFilter.Summary.Setups)
		fmt.Printf("largeBuyTrades=%d\n", orderFlowTradesFilter.Summary.LargeBuyTrades)
		fmt.Printf("largeSellTrades=%d\n", orderFlowTradesFilter.Summary.LargeSellTrades)
		fmt.Printf("accepted=%d\n", orderFlowTradesFilter.Summary.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowTradesFilter.Summary.Rejected)
		fmt.Printf("nearHVN=%d\n", orderFlowTradesFilter.Summary.NearHVN)
		fmt.Printf("nearVolumeCluster=%d\n", orderFlowTradesFilter.Summary.NearVolumeCluster)
		fmt.Printf("nearMultipleHVN=%d\n", orderFlowTradesFilter.Summary.NearMultipleHVN)
		fmt.Printf("bestDirection=%s\n", orderFlowTradesFilter.Summary.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowTradesFilter.Summary.BestDirectionFT20)
		fmt.Println("wrote research/orderflow_setup_trades_filter.csv")
		fmt.Println("wrote research/orderflow_setup_trades_filter_summary.json")
		fmt.Println("wrote research/orderflow_setup_trades_filter.md")
		fmt.Println()
	}
	if orderFlowStackedImbalancesOK {
		fmt.Println("=== ORDER FLOW SETUP #4 STACKED IMBALANCES ===")
		fmt.Printf("setups=%d\n", orderFlowStackedImbalances.Summary.Setups)
		fmt.Printf("buyStackedImbalances=%d\n", orderFlowStackedImbalances.Summary.BuyStackedImbalances)
		fmt.Printf("sellStackedImbalances=%d\n", orderFlowStackedImbalances.Summary.SellStackedImbalances)
		fmt.Printf("accepted=%d\n", orderFlowStackedImbalances.Summary.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowStackedImbalances.Summary.Rejected)
		fmt.Printf("neutral=%d\n", orderFlowStackedImbalances.Summary.Neutral)
		fmt.Printf("retests=%d\n", orderFlowStackedImbalances.Summary.Retests)
		fmt.Printf("averageStackedLevels=%.2f\n", orderFlowStackedImbalances.Summary.AverageStackedLevels)
		fmt.Printf("bestDirection=%s\n", orderFlowStackedImbalances.Summary.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowStackedImbalances.Summary.BestDirectionFT20)
		fmt.Printf("nearHVN=%d\n", orderFlowStackedImbalances.Summary.NearHVN)
		fmt.Printf("nearVolumeCluster=%d\n", orderFlowStackedImbalances.Summary.NearVolumeCluster)
		fmt.Printf("nearMultipleHVN=%d\n", orderFlowStackedImbalances.Summary.NearMultipleHVN)
		fmt.Println("wrote research/orderflow_setup_stacked_imbalance.csv")
		fmt.Println("wrote research/orderflow_setup_stacked_imbalance_summary.json")
		fmt.Println("wrote research/orderflow_setup_stacked_imbalance.md")
		fmt.Println()
	}
	if orderFlowUnfinishedBusinessOK {
		fmt.Println("=== ORDER FLOW SETUP #5 UNFINISHED BUSINESS ===")
		fmt.Printf("setups=%d\n", orderFlowUnfinishedBusiness.Summary.Setups)
		fmt.Printf("unfinishedHigh=%d\n", orderFlowUnfinishedBusiness.Summary.UnfinishedHigh)
		fmt.Printf("unfinishedLow=%d\n", orderFlowUnfinishedBusiness.Summary.UnfinishedLow)
		fmt.Printf("revisited=%d\n", orderFlowUnfinishedBusiness.Summary.Revisited)
		fmt.Printf("accepted=%d\n", orderFlowUnfinishedBusiness.Summary.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowUnfinishedBusiness.Summary.Rejected)
		fmt.Printf("neutral=%d\n", orderFlowUnfinishedBusiness.Summary.Neutral)
		fmt.Printf("magnetContexts=%d\n", orderFlowUnfinishedBusiness.Summary.MagnetContexts)
		fmt.Printf("bestLocation=%s\n", orderFlowUnfinishedBusiness.Summary.BestLocation)
		fmt.Printf("bestLocationRevisitRate=%.2f\n", orderFlowUnfinishedBusiness.Summary.BestLocationRevisitRate)
		fmt.Println("wrote research/orderflow_setup_unfinished_business.csv")
		fmt.Println("wrote research/orderflow_setup_unfinished_business_summary.json")
		fmt.Println("wrote research/orderflow_setup_unfinished_business.md")
		fmt.Println()
	}
	if orderFlowBigLimitOrdersOK {
		fmt.Println("=== ORDER FLOW CONFIRMATION #1 BIG LIMIT ORDERS ===")
		fmt.Printf("combinedConfirmations=%d\n", orderFlowBigLimitOrders.Summary.Combined.Confirmations)
		fmt.Printf("liveConfirmations=%d\n", orderFlowBigLimitOrders.Summary.Sources[research.LiveMultiVenueFootprintSource].Confirmations)
		fmt.Printf("backfillConfirmations=%d\n", orderFlowBigLimitOrders.Summary.Sources[research.BackfilledAsterAggTradesSource].Confirmations)
		fmt.Printf("venues=%d\n", orderFlowBigLimitOrders.Summary.Combined.Venues)
		fmt.Printf("canonicalSymbols=%d\n", orderFlowBigLimitOrders.Summary.Combined.CanonicalSymbols)
		fmt.Printf("bullishAbsorption=%d\n", orderFlowBigLimitOrders.Summary.Combined.BullishAbsorption)
		fmt.Printf("bearishAbsorption=%d\n", orderFlowBigLimitOrders.Summary.Combined.BearishAbsorption)
		fmt.Printf("accepted=%d\n", orderFlowBigLimitOrders.Summary.Combined.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowBigLimitOrders.Summary.Combined.Rejected)
		fmt.Printf("bestDirection=%s\n", orderFlowBigLimitOrders.Summary.Combined.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowBigLimitOrders.Summary.Combined.BestDirectionFT20)
		fmt.Printf("venueCoverage=%s\n", formatOrderFlowVenueCoverage(orderFlowBigLimitOrders.Summary.VenueCoverage))
		fmt.Println("wrote research/orderflow_confirmation_big_limit_orders.csv")
		fmt.Println("wrote research/orderflow_confirmation_big_limit_orders_summary.json")
		fmt.Println("wrote research/orderflow_confirmation_big_limit_orders.md")
		fmt.Println()
	}
	if orderFlowAbsorptionOK {
		fmt.Println("=== ORDER FLOW CONFIRMATION #2 ABSORPTION ===")
		fmt.Printf("combinedConfirmations=%d\n", orderFlowAbsorption.Summary.Combined.Confirmations)
		fmt.Printf("liveConfirmations=%d\n", orderFlowAbsorption.Summary.SourceCoverage[research.LiveMultiVenueFootprintSource].Confirmations)
		fmt.Printf("backfillConfirmations=%d\n", orderFlowAbsorption.Summary.SourceCoverage[research.BackfilledAsterAggTradesSource].Confirmations)
		fmt.Printf("venues=%d\n", orderFlowAbsorption.Summary.Combined.Venues)
		fmt.Printf("canonicalSymbols=%d\n", orderFlowAbsorption.Summary.Combined.CanonicalSymbols)
		fmt.Printf("bullishAbsorption=%d\n", orderFlowAbsorption.Summary.Combined.BullishAbsorption)
		fmt.Printf("bearishAbsorption=%d\n", orderFlowAbsorption.Summary.Combined.BearishAbsorption)
		fmt.Printf("accepted=%d\n", orderFlowAbsorption.Summary.Combined.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowAbsorption.Summary.Combined.Rejected)
		fmt.Printf("neutral=%d\n", orderFlowAbsorption.Summary.Combined.Neutral)
		fmt.Printf("bestDirection=%s\n", orderFlowAbsorption.Summary.Combined.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowAbsorption.Summary.Combined.BestDirectionFT20)
		fmt.Printf("venueCoverage=%s\n", formatOrderFlowVenueCoverage(orderFlowAbsorption.Summary.VenueCoverage))
		fmt.Println("wrote research/orderflow_confirmation_absorption.csv")
		fmt.Println("wrote research/orderflow_confirmation_absorption_summary.json")
		fmt.Println("wrote research/orderflow_confirmation_absorption.md")
		fmt.Println()
	}
	if orderFlowAggressiveDeltaOK {
		fmt.Println("=== ORDER FLOW CONFIRMATION #3 AGGRESSIVE DELTA ===")
		fmt.Printf("combinedConfirmations=%d\n", orderFlowAggressiveDelta.Summary.Combined.Confirmations)
		fmt.Printf("liveConfirmations=%d\n", orderFlowAggressiveDelta.Summary.SourceCoverage[research.LiveMultiVenueFootprintSource].Confirmations)
		fmt.Printf("backfillConfirmations=%d\n", orderFlowAggressiveDelta.Summary.SourceCoverage[research.BackfilledAsterAggTradesSource].Confirmations)
		fmt.Printf("venues=%d\n", orderFlowAggressiveDelta.Summary.Combined.Venues)
		fmt.Printf("canonicalSymbols=%d\n", orderFlowAggressiveDelta.Summary.Combined.CanonicalSymbols)
		fmt.Printf("aggressiveBuy=%d\n", orderFlowAggressiveDelta.Summary.Combined.AggressiveBuy)
		fmt.Printf("aggressiveSell=%d\n", orderFlowAggressiveDelta.Summary.Combined.AggressiveSell)
		fmt.Printf("accepted=%d\n", orderFlowAggressiveDelta.Summary.Combined.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowAggressiveDelta.Summary.Combined.Rejected)
		fmt.Printf("neutral=%d\n", orderFlowAggressiveDelta.Summary.Combined.Neutral)
		fmt.Printf("bestDirection=%s\n", orderFlowAggressiveDelta.Summary.Combined.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowAggressiveDelta.Summary.Combined.BestDirectionFT20)
		fmt.Printf("bestDeltaBucket=%s\n", orderFlowAggressiveDelta.Summary.Combined.BestDeltaBucket)
		fmt.Printf("bestDeltaBucketFT20=%.8f\n", orderFlowAggressiveDelta.Summary.Combined.BestDeltaBucketFT20)
		fmt.Printf("venueCoverage=%s\n", formatOrderFlowVenueCoverage(orderFlowAggressiveDelta.Summary.VenueCoverage))
		fmt.Println("wrote research/orderflow_confirmation_aggressive_delta.csv")
		fmt.Println("wrote research/orderflow_confirmation_aggressive_delta_summary.json")
		fmt.Println("wrote research/orderflow_confirmation_aggressive_delta.md")
		fmt.Println()
	}
	if orderFlowCumulativeDeltaDivergenceOK {
		fmt.Println("=== ORDER FLOW CONFIRMATION #4 CUMULATIVE DELTA DIVERGENCE ===")
		fmt.Printf("combinedConfirmations=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.Confirmations)
		fmt.Printf("liveConfirmations=%d\n", orderFlowCumulativeDeltaDivergence.Summary.SourceCoverage[research.LiveMultiVenueFootprintSource].Confirmations)
		fmt.Printf("backfillConfirmations=%d\n", orderFlowCumulativeDeltaDivergence.Summary.SourceCoverage[research.BackfilledAsterAggTradesSource].Confirmations)
		fmt.Printf("venues=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.Venues)
		fmt.Printf("canonicalSymbols=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.CanonicalSymbols)
		fmt.Printf("bullishDivergence=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.BullishDivergence)
		fmt.Printf("bearishDivergence=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.BearishDivergence)
		fmt.Printf("accepted=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.Accepted)
		fmt.Printf("rejected=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.Rejected)
		fmt.Printf("neutral=%d\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.Neutral)
		fmt.Printf("bestDirection=%s\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.BestDirection)
		fmt.Printf("bestDirectionFT20=%.8f\n", orderFlowCumulativeDeltaDivergence.Summary.Combined.BestDirectionFT20)
		fmt.Printf("venueCoverage=%s\n", formatOrderFlowVenueCoverage(orderFlowCumulativeDeltaDivergence.Summary.VenueCoverage))
		fmt.Println("wrote research/orderflow_confirmation_cumulative_delta_divergence.csv")
		fmt.Println("wrote research/orderflow_confirmation_cumulative_delta_divergence_summary.json")
		fmt.Println("wrote research/orderflow_confirmation_cumulative_delta_divergence.md")
		fmt.Println()
	}
	if orderFlowConfirmationComparisonOK {
		fmt.Println("=== ORDER FLOW CONFIRMATION COMPARISON ===")
		fmt.Printf("bestAcceptanceConfirmation=%s\n", orderFlowConfirmationComparison.Summary.BestAcceptanceConfirmation)
		fmt.Printf("bestFT5Confirmation=%s\n", orderFlowConfirmationComparison.Summary.BestFT5Confirmation)
		fmt.Printf("bestFT10Confirmation=%s\n", orderFlowConfirmationComparison.Summary.BestFT10Confirmation)
		fmt.Printf("bestFT20Confirmation=%s\n", orderFlowConfirmationComparison.Summary.BestFT20Confirmation)
		fmt.Printf("largestSampleConfirmation=%s\n", orderFlowConfirmationComparison.Summary.LargestSampleConfirmation)
		fmt.Printf("broadestVenueCoverage=%s\n", orderFlowConfirmationComparison.Summary.BroadestVenueCoverage)
		fmt.Println("wrote research/orderflow_confirmation_comparison.csv")
		fmt.Println("wrote research/orderflow_confirmation_comparison.json")
		fmt.Println("wrote research/orderflow_confirmation_comparison.md")
		fmt.Println()
	}
	printHyperliquidL2Snapshot(hyperliquidL2Snapshot)
	printMultiVenueL2Summary(orderBookRows.Snapshots)
	fmt.Println("=== ORDER BOOK RESEARCH EXPORTS ===")
	fmt.Printf("orderbookFeatureRows=%d\n", len(orderBookRows.Features))
	fmt.Printf("vwapL2InteractionRows=%d\n", len(orderBookRows.VWAPInteraction))
	fmt.Println("wrote research/orderbook_features.csv")
	fmt.Println("wrote research/vwap_l2_interaction.csv")
	fmt.Println("wrote research/multi_venue_orderbook_features.csv")
	fmt.Println("wrote research/multi_venue_vwap_l2_interaction.csv")
	fmt.Println()
	fmt.Println("=== VWAP CHAPTER 1-3 L2 REFRESH ===")
	fmt.Printf("venues=%d\n", len(orderBookRows.Snapshots))
	fmt.Printf("featureRows=%d\n", len(vwapL2Refresh.FeatureRows))
	fmt.Printf("contextRows=%d\n", len(vwapL2Refresh.ContextRows))
	fmt.Printf("bounceBidSupport=%d\n", vwapL2Refresh.Summary.BounceBidSupport)
	fmt.Printf("bounceAskWeakness=%d\n", vwapL2Refresh.Summary.BounceAskWeakness)
	fmt.Printf("breakAskPressure=%d\n", vwapL2Refresh.Summary.BreakAskPressure)
	fmt.Printf("breakBidCollapse=%d\n", vwapL2Refresh.Summary.BreakBidCollapse)
	fmt.Println("wrote research/vwap_features_l2.csv")
	fmt.Println("wrote research/context_features_l2.csv")
	fmt.Println("wrote research/vwap_behavior_l2_summary.json")
	fmt.Println()
	if l2SnapshotAnalysisOK {
		fmt.Println("=== HISTORICAL L2 ANALYSIS ===")
		fmt.Printf("rows=%d\n", l2SnapshotAnalysis.DatasetQuality.TotalRows)
		fmt.Printf("validRows=%d\n", l2SnapshotAnalysis.DatasetQuality.ValidRows)
		fmt.Printf("venues=%d\n", len(l2SnapshotAnalysis.VenueSummaries))
		fmt.Println("wrote research/l2_snapshot_analysis.json")
		fmt.Println("wrote research/l2_snapshot_analysis.md")
		fmt.Println()
	}
	fmt.Println("=== CHAPTER 6 RISK ANALYTICS ===")
	fmt.Println()
	fmt.Println("strategy sharpe sortino expectancy drawdown grade")
	for _, row := range riskRows {
		fmt.Printf(
			"%s %.2f %.2f %.2f %.2f %s\n",
			row.Strategy,
			row.Sharpe,
			row.Sortino,
			row.Expectancy,
			row.MaxDrawdown,
			row.RiskGrade,
		)
	}
	fmt.Println()
	fmt.Println("wrote " + riskMetricsPath)
	fmt.Println("wrote " + riskSummaryPath)
	fmt.Println()
	fmt.Println("=== CHAPTER 6 RISK FILTERS ===")
	fmt.Println()
	fmt.Println("strategy decision reasons")
	for _, decision := range filterDecisions {
		fmt.Printf(
			"%s %s %s\n",
			decision.Strategy,
			decision.Decision,
			strings.Join(decision.Reasons, ","),
		)
	}
	fmt.Println()
	fmt.Println("wrote " + filterCSVPath)
	fmt.Println("wrote " + filterSummaryPath)
	fmt.Println()
	fmt.Println("=== CHAPTER 6 RISK RANKING ===")
	fmt.Println()
	fmt.Println("rank strategy decision score reasons")
	for _, row := range riskRankings {
		fmt.Printf(
			"%d %s %s %.2f %s\n",
			row.Rank,
			row.Strategy,
			row.Decision,
			row.Score,
			strings.Join(row.Reasons, ","),
		)
	}
	fmt.Println()
	fmt.Println("wrote " + rankingCSVPath)
	fmt.Println("wrote " + rankingSummaryPath)
	fmt.Println()
	fmt.Println("=== CHAPTER 6 PROMOTION GATES ===")
	fmt.Println()
	fmt.Println("rank strategy gate reasons")
	for _, gate := range promotionGates {
		fmt.Printf(
			"%d %s %s %s\n",
			gate.Rank,
			gate.Strategy,
			gate.Gate,
			strings.Join(gate.Reasons, ","),
		)
	}
	fmt.Println()
	fmt.Println("wrote " + promotionCSVPath)
	fmt.Println("wrote " + promotionSummaryPath)
	fmt.Println()
	fmt.Println("=== CHAPTER 6 FINAL RISK PACKET ===")
	fmt.Println()
	fmt.Printf("bestRiskAdjusted=%d\n", len(chapter6Packet.BestRiskAdjustedStrategies))
	fmt.Printf("keptThrottled=%d\n", len(chapter6Packet.KeptThrottled))
	fmt.Printf("rewriteRequired=%d\n", len(chapter6Packet.RewriteRequired))
	fmt.Printf("removedFromCandidates=%d\n", len(chapter6Packet.RemovedFromCandidates))
	fmt.Println("conclusion=Chapter 6 remains a research-only risk analysis packet.")
	fmt.Println("next=" + chapter6Packet.NextRecommendedPhase)
	fmt.Println()
	fmt.Println("wrote " + packetJSONPath)
	fmt.Println("wrote " + packetMarkdownPath)
	fmt.Println()
	fmt.Println("=== CHAPTER 7B TRADING SYSTEM SIMULATION ===")
	chapter7Summary := system.NewTestTradingSimulation(16).RunArbitrageExample()
	fmt.Printf(
		"order=%s status=%s pnl=%.2f bid=%.2f ask=%.2f auditEvents=%d\n",
		chapter7Summary.OrderID,
		chapter7Summary.OrderStatus,
		chapter7Summary.RealizedPnL,
		chapter7Summary.BookBestBid,
		chapter7Summary.BookBestAsk,
		chapter7Summary.AuditEvents,
	)
	fmt.Println("wrote research/chapter7_architecture.md")
	fmt.Println()
	fmt.Println("=== CHAPTER 7C COMMAND CONTROL SERVICES ===")
	supervisor := system.NewChapter7Supervisor()
	control := system.NewCommandControl(supervisor)
	control.Handle(system.CommandStart)
	control.Handle(system.CommandPause)
	control.Handle(system.CommandResume)
	control.Handle(system.CommandStatus)
	fmt.Printf(
		"state=%s services=%d running=%d auditEntries=%d\n",
		control.State(),
		supervisor.ServiceCount(),
		supervisor.CountByStatus(system.ServiceRunning),
		len(control.AuditLog()),
	)
	chapter7Packet := research.BuildChapter7Packet()
	chapter7PacketJSONPath := "research/chapter7_packet.json"
	if err := research.WriteChapter7PacketJSON(chapter7PacketJSONPath, chapter7Packet); err != nil {
		log.Fatalf("write chapter 7 packet json: %v", err)
	}
	chapter7PacketMarkdownPath := "research/chapter7_packet.md"
	if err := research.WriteChapter7PacketMarkdown(chapter7PacketMarkdownPath, chapter7Packet); err != nil {
		log.Fatalf("write chapter 7 packet markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 7D FINAL TRADING SYSTEM PACKET ===")
	fmt.Printf("components=%d critical=%d nonCritical=%d gaps=%d\n",
		len(chapter7Packet.ImplementedComponents),
		len(chapter7Packet.CriticalComponents),
		len(chapter7Packet.NonCriticalComponents),
		len(chapter7Packet.RemainingGaps),
	)
	fmt.Println("conclusion=Chapter 7 trading-system skeleton is complete.")
	fmt.Println("readiness=" + chapter7Packet.ReadinessForChapter8)
	fmt.Println("wrote " + chapter7PacketJSONPath)
	fmt.Println("wrote " + chapter7PacketMarkdownPath)
	chapter8Audit := research.BuildChapter8GatewayAudit()
	chapter8AuditJSONPath := "research/chapter8_gateway_audit.json"
	if err := research.WriteChapter8GatewayAuditJSON(chapter8AuditJSONPath, chapter8Audit); err != nil {
		log.Fatalf("write chapter 8 gateway audit json: %v", err)
	}
	chapter8AuditMarkdownPath := "research/chapter8_gateway_audit.md"
	if err := research.WriteChapter8GatewayAuditMarkdown(chapter8AuditMarkdownPath, chapter8Audit); err != nil {
		log.Fatalf("write chapter 8 gateway audit markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 8B GATEWAY AUDIT ===")
	fmt.Printf("concepts=%d venues=%d gaps=%d\n", len(chapter8Audit.Concepts), len(chapter8Audit.Venues), len(chapter8Audit.Gaps))
	fmt.Println("conclusion=" + chapter8Audit.Conclusion)
	fmt.Println("next=" + chapter8Audit.NextPhase)
	fmt.Println("wrote " + chapter8AuditJSONPath)
	fmt.Println("wrote " + chapter8AuditMarkdownPath)
	chapter8VenueMap := research.BuildChapter8VenueGatewayMap()
	chapter8VenueMapJSONPath := "research/chapter8_venue_gateway_map.json"
	if err := research.WriteChapter8VenueGatewayMapJSON(chapter8VenueMapJSONPath, chapter8VenueMap); err != nil {
		log.Fatalf("write chapter 8 venue gateway map json: %v", err)
	}
	chapter8VenueMapMarkdownPath := "research/chapter8_venue_gateway_map.md"
	if err := research.WriteChapter8VenueGatewayMapMarkdown(chapter8VenueMapMarkdownPath, chapter8VenueMap); err != nil {
		log.Fatalf("write chapter 8 venue gateway map markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 8C VENUE GATEWAY TRANSLATORS ===")
	fmt.Printf("translators=%d enabledByDefault=false allowLiveOrders=false\n", len(chapter8VenueMap.Translators))
	fmt.Println("conclusion=" + chapter8VenueMap.Conclusion)
	fmt.Println("next=" + chapter8VenueMap.NextPhase)
	fmt.Println("wrote " + chapter8VenueMapJSONPath)
	fmt.Println("wrote " + chapter8VenueMapMarkdownPath)
	chapter8SessionLifecycle := research.BuildChapter8SessionLifecycle()
	chapter8SessionJSONPath := "research/chapter8_session_lifecycle.json"
	if err := research.WriteChapter8SessionLifecycleJSON(chapter8SessionJSONPath, chapter8SessionLifecycle); err != nil {
		log.Fatalf("write chapter 8 session lifecycle json: %v", err)
	}
	chapter8SessionMarkdownPath := "research/chapter8_session_lifecycle.md"
	if err := research.WriteChapter8SessionLifecycleMarkdown(chapter8SessionMarkdownPath, chapter8SessionLifecycle); err != nil {
		log.Fatalf("write chapter 8 session lifecycle markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 8D GATEWAY SESSION LIFECYCLE ===")
	fmt.Printf(
		"sessions=%d heartbeatOK=%d messages=%d errors=%d\n",
		chapter8SessionLifecycle.Summary.Total,
		chapter8SessionLifecycle.Summary.HeartbeatOK,
		chapter8SessionLifecycle.Summary.Messages,
		chapter8SessionLifecycle.Summary.Errors,
	)
	fmt.Println("conclusion=" + chapter8SessionLifecycle.Conclusion)
	fmt.Println("next=" + chapter8SessionLifecycle.NextPhase)
	fmt.Println("wrote " + chapter8SessionJSONPath)
	fmt.Println("wrote " + chapter8SessionMarkdownPath)
	chapter8Packet := research.BuildChapter8Packet()
	chapter8PacketJSONPath := "research/chapter8_packet.json"
	if err := research.WriteChapter8PacketJSON(chapter8PacketJSONPath, chapter8Packet); err != nil {
		log.Fatalf("write chapter 8 packet json: %v", err)
	}
	chapter8PacketMarkdownPath := "research/chapter8_packet.md"
	if err := research.WriteChapter8PacketMarkdown(chapter8PacketMarkdownPath, chapter8Packet); err != nil {
		log.Fatalf("write chapter 8 packet markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 8E FINAL EXCHANGE CONNECTIVITY PACKET ===")
	fmt.Printf(
		"concepts=%d venues=%d gaps=%d fixItems=%d\n",
		len(chapter8Packet.ImplementedConcepts),
		len(chapter8Packet.VenueAPIMapping),
		len(chapter8Packet.RemainingGaps),
		len(chapter8Packet.MissingFIXItems),
	)
	fmt.Println("conclusion=Chapter 8 exchange connectivity layer is mapped and safely abstracted.")
	fmt.Println("readiness=" + chapter8Packet.ReadinessForChapter9)
	fmt.Println("wrote " + chapter8PacketJSONPath)
	fmt.Println("wrote " + chapter8PacketMarkdownPath)
	inSample, outOfSample := backtest.InSampleOutOfSampleSplit(primary.Candles, backtest.DefaultInSampleRatio)
	chapter9Assumptions := backtest.DefaultBacktestAssumptions()
	chapter9Audit := research.BuildChapter9BacktestAudit(len(inSample), len(outOfSample), chapter9Assumptions)
	chapter9AuditJSONPath := "research/chapter9_backtest_audit.json"
	if err := research.WriteChapter9BacktestAuditJSON(chapter9AuditJSONPath, chapter9Audit); err != nil {
		log.Fatalf("write chapter 9 backtest audit json: %v", err)
	}
	chapter9AuditMarkdownPath := "research/chapter9_backtest_audit.md"
	if err := research.WriteChapter9BacktestAuditMarkdown(chapter9AuditMarkdownPath, chapter9Audit); err != nil {
		log.Fatalf("write chapter 9 backtest audit markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 9B BACKTEST SPLITS AND ASSUMPTIONS ===")
	fmt.Printf("inSample=%d outOfSample=%d assumptionGaps=%d\n", len(inSample), len(outOfSample), len(chapter9Assumptions.Gaps))
	fmt.Println("engine=" + chapter9Assumptions.EngineModel)
	fmt.Println("next=" + chapter9Audit.NextPhase)
	fmt.Println("wrote " + chapter9AuditJSONPath)
	fmt.Println("wrote " + chapter9AuditMarkdownPath)
	chapter9TimeReport := research.BuildChapter9TimeModelReport(len(primary.Candles))
	chapter9TimeJSONPath := "research/chapter9_time_model.json"
	if err := research.WriteChapter9TimeModelJSON(chapter9TimeJSONPath, chapter9TimeReport); err != nil {
		log.Fatalf("write chapter 9 time model json: %v", err)
	}
	chapter9TimeMarkdownPath := "research/chapter9_time_model.md"
	if err := research.WriteChapter9TimeModelMarkdown(chapter9TimeMarkdownPath, chapter9TimeReport); err != nil {
		log.Fatalf("write chapter 9 time model markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 9C SIMULATED CLOCK ===")
	fmt.Printf("candles=%d timeGaps=%d\n", len(primary.Candles), len(chapter9TimeReport.RemainingGaps))
	fmt.Println("clock=deterministic simulated time source")
	fmt.Println("next=" + chapter9TimeReport.ReadinessForEventBacktest)
	fmt.Println("wrote " + chapter9TimeJSONPath)
	fmt.Println("wrote " + chapter9TimeMarkdownPath)
	eventBacktester := backtest.NewEventDrivenBacktester(backtest.EventDrivenBacktestConfig{
		Symbol:              primary.Symbol,
		StartingCash:        10000,
		MaxCandles:          25,
		UseSimulatedGateway: true,
		AllowLiveOrders:     false,
	})
	eventResult, err := eventBacktester.Run(primary.Candles)
	if err != nil {
		log.Fatalf("run chapter 9 event-driven backtest: %v", err)
	}
	chapter9EventReport := research.BuildChapter9EventDrivenReport(eventResult)
	chapter9EventJSONPath := "research/chapter9_event_driven.json"
	if err := research.WriteChapter9EventDrivenJSON(chapter9EventJSONPath, chapter9EventReport); err != nil {
		log.Fatalf("write chapter 9 event-driven json: %v", err)
	}
	chapter9EventMarkdownPath := "research/chapter9_event_driven.md"
	if err := research.WriteChapter9EventDrivenMarkdown(chapter9EventMarkdownPath, chapter9EventReport); err != nil {
		log.Fatalf("write chapter 9 event-driven markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 9D EVENT-DRIVEN BACKTESTER ===")
	fmt.Printf(
		"candles=%d orders=%d fills=%d pnl=%.2f auditEvents=%d\n",
		eventResult.CandlesProcessed,
		eventResult.OrdersCreated,
		eventResult.OrdersFilled,
		eventResult.FinalPnL,
		eventResult.AuditEvents,
	)
	fmt.Println("clockStart=" + eventResult.ClockStart.Format(time.RFC3339))
	fmt.Println("clockEnd=" + eventResult.ClockEnd.Format(time.RFC3339))
	fmt.Println("next=" + chapter9EventReport.NextPhase)
	fmt.Println("wrote " + chapter9EventJSONPath)
	fmt.Println("wrote " + chapter9EventMarkdownPath)
	comparisonResult, err := backtest.CompareForLoopVsEventDriven(
		primary.Candles,
		backtest.Config{
			Venue:           primary.Venue,
			Symbol:          primary.Symbol,
			Interval:        ResearchInterval,
			StartingBalance: 10000,
			FixedNotional:   100,
			FeeModel:        backtest.DefaultFeeModel(),
			SlippageModel:   backtest.DefaultSlippageModel(),
		},
		backtest.EventDrivenBacktestConfig{
			Symbol:              primary.Symbol,
			StartingCash:        10000,
			MaxCandles:          25,
			UseSimulatedGateway: true,
			AllowLiveOrders:     false,
		},
	)
	if err != nil {
		log.Fatalf("compare for-loop and event-driven backtesters: %v", err)
	}
	chapter9ComparisonReport := research.BuildChapter9BacktesterComparisonReport(comparisonResult)
	chapter9ComparisonJSONPath := "research/chapter9_backtester_comparison.json"
	if err := research.WriteChapter9BacktesterComparisonJSON(chapter9ComparisonJSONPath, chapter9ComparisonReport); err != nil {
		log.Fatalf("write chapter 9 backtester comparison json: %v", err)
	}
	chapter9ComparisonMarkdownPath := "research/chapter9_backtester_comparison.md"
	if err := research.WriteChapter9BacktesterComparisonMarkdown(chapter9ComparisonMarkdownPath, chapter9ComparisonReport); err != nil {
		log.Fatalf("write chapter 9 backtester comparison markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 9E BACKTESTER COMPARISON ===")
	fmt.Printf(
		"candles=%d forLoopTrades=%d eventOrders=%d eventFills=%d pnlDiff=%.2f\n",
		comparisonResult.Candles,
		comparisonResult.ForLoopTrades,
		comparisonResult.EventDrivenOrders,
		comparisonResult.EventDrivenFills,
		comparisonResult.PnLDifference,
	)
	fmt.Println("recommendation=" + comparisonResult.Recommendation)
	fmt.Println("wrote " + chapter9ComparisonJSONPath)
	fmt.Println("wrote " + chapter9ComparisonMarkdownPath)
	chapter9Packet := research.BuildChapter9Packet(chapter9Audit, chapter9TimeReport, chapter9EventReport, chapter9ComparisonReport)
	chapter9PacketJSONPath := "research/chapter9_packet.json"
	if err := research.WriteChapter9PacketJSON(chapter9PacketJSONPath, chapter9Packet); err != nil {
		log.Fatalf("write chapter 9 packet json: %v", err)
	}
	chapter9PacketMarkdownPath := "research/chapter9_packet.md"
	if err := research.WriteChapter9PacketMarkdown(chapter9PacketMarkdownPath, chapter9Packet); err != nil {
		log.Fatalf("write chapter 9 packet markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 9F FINAL BACKTESTER PACKET ===")
	fmt.Printf(
		"concepts=%d assumptionGaps=%d chapter10Deferred=%d\n",
		len(chapter9Packet.ImplementedConcepts),
		len(chapter9Packet.AssumptionGaps),
		len(chapter9Packet.Chapter10DeferredGaps),
	)
	fmt.Println("conclusion=Chapter 9 backtesting layer is complete.")
	fmt.Println("readiness=" + chapter9Packet.ReadinessForChapter10)
	fmt.Println("wrote " + chapter9PacketJSONPath)
	fmt.Println("wrote " + chapter9PacketMarkdownPath)
	chapter10Audit := research.BuildChapter10RealismAudit(comparisonResult)
	chapter10AuditJSONPath := "research/chapter10_realism_audit.json"
	if err := research.WriteChapter10RealismAuditJSON(chapter10AuditJSONPath, chapter10Audit); err != nil {
		log.Fatalf("write chapter 10 realism audit json: %v", err)
	}
	chapter10AuditMarkdownPath := "research/chapter10_realism_audit.md"
	if err := research.WriteChapter10RealismAuditMarkdown(chapter10AuditMarkdownPath, chapter10Audit); err != nil {
		log.Fatalf("write chapter 10 realism audit markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 10A REALISM AUDIT ===")
	fmt.Printf(
		"bias=%s severity=%s grade=%s gaps=%d pnlDiff=%.2f\n",
		chapter10Audit.Dislocation.BiasDirection,
		chapter10Audit.Dislocation.Severity,
		chapter10Audit.Grade,
		len(chapter10Audit.RealismGaps),
		chapter10Audit.Dislocation.PnLDifference,
	)
	fmt.Println("eventDriven=" + chapter10Audit.EventDrivenClassification)
	fmt.Println("next=" + chapter10Audit.NextPhase)
	fmt.Println("wrote " + chapter10AuditJSONPath)
	fmt.Println("wrote " + chapter10AuditMarkdownPath)
	chapter10Models := realism.DefaultAggregateRealismModel()
	chapter10ModelsReport := research.BuildChapter10RealismModelsReport(chapter10Models)
	chapter10ModelsJSONPath := "research/chapter10_realism_models.json"
	if err := research.WriteChapter10RealismModelsJSON(chapter10ModelsJSONPath, chapter10ModelsReport); err != nil {
		log.Fatalf("write chapter 10 realism models json: %v", err)
	}
	chapter10ModelsMarkdownPath := "research/chapter10_realism_models.md"
	if err := research.WriteChapter10RealismModelsMarkdown(chapter10ModelsMarkdownPath, chapter10ModelsReport); err != nil {
		log.Fatalf("write chapter 10 realism models markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 10B SIMULATION REALISM MODELS ===")
	fmt.Printf(
		"overallRisk=%s latency=%s placeInLine=%s marketImpact=%s fill=%s warnings=%d\n",
		chapter10Models.OverallRisk,
		chapter10Models.Latency.Risk,
		chapter10Models.PlaceInLine.Risk,
		chapter10Models.MarketImpact.Risk,
		chapter10Models.FillAssumption.Risk,
		len(chapter10Models.Warnings),
	)
	fmt.Println("conclusion=Chapter 10B models realism assumptions only.")
	fmt.Println("next=" + chapter10ModelsReport.NextPhase)
	fmt.Println("wrote " + chapter10ModelsJSONPath)
	fmt.Println("wrote " + chapter10ModelsMarkdownPath)
	dataQuality := realism.CheckMarketDataQuality(primary.Candles, intervalToMillis(ResearchInterval))
	dataParity := realism.DefaultHistoricalLiveParityCheck(primary.Venue, primary.Symbol, ResearchInterval)
	dataQualityReport := research.BuildChapter10DataQualityReport(
		primary.Venue,
		primary.Symbol,
		ResearchInterval,
		realism.BuildDataQualityReport(dataQuality, dataParity),
	)
	chapter10DataQualityJSONPath := "research/chapter10_data_quality.json"
	if err := research.WriteChapter10DataQualityJSON(chapter10DataQualityJSONPath, dataQualityReport); err != nil {
		log.Fatalf("write chapter 10 data quality json: %v", err)
	}
	chapter10DataQualityMarkdownPath := "research/chapter10_data_quality.md"
	if err := research.WriteChapter10DataQualityMarkdown(chapter10DataQualityMarkdownPath, dataQualityReport); err != nil {
		log.Fatalf("write chapter 10 data quality markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 10C DATA QUALITY AUDIT ===")
	fmt.Printf(
		"candles=%d qualityRisk=%s parityRisk=%s warnings=%d recommendations=%d\n",
		dataQualityReport.DataQuality.TotalCandlesChecked,
		dataQualityReport.DataQuality.QualityRisk,
		dataQualityReport.DataQuality.ParityRisk,
		len(dataQualityReport.DataQuality.Warnings),
		len(dataQualityReport.DataQuality.Recommendations),
	)
	fmt.Println("conclusion=Chapter 10C audits market data quality and historical/live parity only.")
	fmt.Println("next=" + dataQualityReport.NextPhase)
	fmt.Println("wrote " + chapter10DataQualityJSONPath)
	fmt.Println("wrote " + chapter10DataQualityMarkdownPath)
	riskEnforcementResults := runChapter6RiskEnforcement(primary, append(append(chapter2Results, chapter4SingleResults...), chapter5Results...))
	riskEnforcementSummary := research.BuildChapter6RiskEnforcementSummary(riskEnforcementResults, removedCandidateSetFromPacket(chapter6Packet))
	riskEnforcementCSVPath := "research/chapter6_risk_enforcement.csv"
	if err := research.WriteChapter6RiskEnforcementCSV(riskEnforcementCSVPath, riskEnforcementSummary.Rows); err != nil {
		log.Fatalf("write chapter 6 risk enforcement csv: %v", err)
	}
	riskEnforcementJSONPath := "research/chapter6_risk_enforcement_summary.json"
	if err := research.WriteChapter6RiskEnforcementSummaryJSON(riskEnforcementJSONPath, riskEnforcementSummary); err != nil {
		log.Fatalf("write chapter 6 risk enforcement summary json: %v", err)
	}
	riskEnforcementMarkdownPath := "research/chapter6_risk_enforcement.md"
	if err := research.WriteChapter6RiskEnforcementMarkdown(riskEnforcementMarkdownPath, riskEnforcementSummary); err != nil {
		log.Fatalf("write chapter 6 risk enforcement markdown: %v", err)
	}
	fmt.Println()
	fmt.Println("=== CHAPTER 6B/6C RISK CONTROLS ENFORCEMENT ===")
	fmt.Printf(
		"strategies=%d tradesBefore=%d tradesAfter=%d violations=%d\n",
		riskEnforcementSummary.Strategies,
		riskEnforcementSummary.TradesBefore,
		riskEnforcementSummary.TradesAfter,
		riskEnforcementSummary.Violations,
	)
	fmt.Printf("improved=%d stillRejected=%d\n", len(riskEnforcementSummary.StrategiesImproved), len(riskEnforcementSummary.StrategiesRejected))
	fmt.Println("conclusion=Risk controls are enforced in backtest/research only.")
	fmt.Println("wrote " + riskEnforcementCSVPath)
	fmt.Println("wrote " + riskEnforcementJSONPath)
	fmt.Println("wrote " + riskEnforcementMarkdownPath)

	_ = chapter4VenueRows
	_ = chapter4VenueAnalyses
}

func buildChapter6RiskMetrics(groups ...[]strategyBacktestResult) []research.StrategyRiskMetrics {
	var rows []research.StrategyRiskMetrics
	for _, group := range groups {
		for _, result := range group {
			pnls := tradePnLs(result.Trades)
			avgWin, avgLoss := averageWinLoss(result.Trades)
			winRate := result.Report.Metrics.WinRate / 100
			hold := riskmetrics.CalculateHoldStats(result.Trades)
			executions := riskmetrics.CalculateExecutionStats(result.Report.Metrics.TotalTrades, ResearchDays)
			expectancy := riskmetrics.Expectancy(winRate, avgWin, avgLoss)
			sharpe := riskmetrics.Sharpe(pnls)

			rows = append(rows, research.StrategyRiskMetrics{
				Strategy:           result.Name,
				Trades:             result.Report.Metrics.TotalTrades,
				WinRate:            result.Report.Metrics.WinRate,
				NetPnL:             result.Report.Metrics.NetPnL,
				Sharpe:             sharpe,
				Sortino:            riskmetrics.Sortino(pnls),
				Expectancy:         expectancy,
				Variance:           riskmetrics.Variance(pnls),
				StdDev:             riskmetrics.StdDev(pnls),
				MaxDrawdown:        result.Report.Metrics.MaxDrawdown,
				MaxDrawdownPct:     result.Report.Metrics.MaxDrawdownPct,
				AverageHoldCandles: hold.AverageHoldCandles,
				AverageHoldHours:   hold.AverageHoldHours,
				TradesPerDay:       executions.TradesPerDay,
				TradesPerWeek:      executions.TradesPerWeek,
				TradesPerMonth:     executions.TradesPerMonth,
				RiskGrade:          riskmetrics.Grade(sharpe, expectancy, result.Report.Metrics.MaxDrawdownPct),
			})
		}
	}
	return rows
}

func buildChapter6RiskSummary(rows []research.StrategyRiskMetrics) research.Chapter6RiskSummary {
	summary := research.Chapter6RiskSummary{Metrics: rows}
	if len(rows) == 0 {
		return summary
	}

	topSharpe := rows[0]
	worstGrade := rows[0]
	largestExecution := rows[0]
	for _, row := range rows[1:] {
		if row.Sharpe > topSharpe.Sharpe {
			topSharpe = row
		}
		if gradeRank(row.RiskGrade) > gradeRank(worstGrade.RiskGrade) {
			worstGrade = row
		}
		if row.TradesPerDay > largestExecution.TradesPerDay {
			largestExecution = row
		}
	}

	summary.TopSharpeStrategy = topSharpe.Strategy
	summary.TopSharpe = topSharpe.Sharpe
	summary.WorstRiskGradeStrategy = worstGrade.Strategy
	summary.WorstRiskGrade = worstGrade.RiskGrade
	summary.LargestExecutionRate = largestExecution.TradesPerDay
	summary.LargestExecutionStrategy = largestExecution.Strategy
	return summary
}

func gradeRank(grade riskmetrics.RiskGrade) int {
	switch grade {
	case riskmetrics.GradeA:
		return 1
	case riskmetrics.GradeB:
		return 2
	case riskmetrics.GradeC:
		return 3
	case riskmetrics.GradeD:
		return 4
	default:
		return 5
	}
}

func tradePnLs(trades []strategy.Trade) []float64 {
	out := make([]float64, len(trades))
	for i, trade := range trades {
		out[i] = trade.RealizedPnL
	}
	return out
}

func averageWinLoss(trades []strategy.Trade) (float64, float64) {
	winTotal := 0.0
	lossTotal := 0.0
	wins := 0
	losses := 0
	for _, trade := range trades {
		switch {
		case trade.RealizedPnL > 0:
			winTotal += trade.RealizedPnL
			wins++
		case trade.RealizedPnL < 0:
			lossTotal += trade.RealizedPnL
			losses++
		}
	}

	avgWin := 0.0
	if wins > 0 {
		avgWin = winTotal / float64(wins)
	}
	avgLoss := 0.0
	if losses > 0 {
		avgLoss = lossTotal / float64(losses)
	}
	return avgWin, avgLoss
}

func fetchVenueCandles(interval string, limit int) []venueCandles {
	asterClient := aster.NewClient(asterEnv("USER"), asterEnv("SIGNER"), asterEnv("PRIVATE_KEY"))
	inputs := []struct {
		venue  string
		symbol string
		reader exchanges.CandleReader
	}{
		{venue: "aster", symbol: "BTCUSDT", reader: asterClient},
		{venue: "hyperliquid", symbol: "BTC", reader: hyperliquid.NewClient("")},
		{venue: "lighter", symbol: "BTC", reader: lighter.NewClient("")},
	}

	out := make([]venueCandles, 0, len(inputs))
	for _, input := range inputs {
		candles, err := input.reader.GetCandles(input.symbol, interval, limit)
		if err != nil {
			log.Printf("skip venue %s %s: %v", input.venue, input.symbol, err)
			continue
		}
		if len(candles) == 0 {
			log.Printf("skip venue %s %s: no candles returned", input.venue, input.symbol)
			continue
		}
		out = append(out, venueCandles{Venue: input.venue, Symbol: input.symbol, Candles: candles})
	}
	return out
}

func buildVWAPFeatures(candles []exchanges.Candle) []research.VWAPFeatureRow {
	anchorTime := int64(0)
	if len(candles) > 0 {
		anchorTime = candles[0].StartTime
	}
	rows := research.BuildVWAPFeatureRows(candles, anchorTime, 5)
	if err := research.WriteVWAPFeaturesCSV("research/vwap_features.csv", rows); err != nil {
		log.Fatalf("write vwap features: %v", err)
	}
	return rows
}

func buildVWAPBehaviorStudy(candles []exchanges.Candle) ([]vwap.Interaction, []vwap.ReactionStudy, research.VWAPBehaviorSummary) {
	interactions, reactions, magnetStats := research.BuildVWAPInteractions(candles)
	if err := research.WriteVWAPInteractionCSV("research/vwap_interaction.csv", interactions); err != nil {
		log.Fatalf("write vwap interaction: %v", err)
	}
	if err := research.WriteVWAPReactionsCSV("research/vwap_reactions.csv", reactions); err != nil {
		log.Fatalf("write vwap reactions: %v", err)
	}
	summary := research.BuildVWAPBehaviorSummary(interactions, magnetStats)
	if err := research.WriteVWAPBehaviorSummaryJSON("research/vwap_behavior_summary.json", summary); err != nil {
		log.Fatalf("write vwap behavior summary: %v", err)
	}
	return interactions, reactions, summary
}

func fetchVWAPTimeframeCandles() map[string][]exchanges.Candle {
	client := aster.NewClient(asterEnv("USER"), asterEnv("SIGNER"), asterEnv("PRIVATE_KEY"))
	out := make(map[string][]exchanges.Candle)
	for _, interval := range []string{"1m", "3m", "5m"} {
		candles, err := client.GetCandles("BTCUSDT", interval, VWAPTimeframeCandleLimit)
		if err != nil {
			log.Printf("skip vwap timeframe %s: %v", interval, err)
			continue
		}
		if len(candles) == 0 {
			log.Printf("skip vwap timeframe %s: no candles returned", interval)
			continue
		}
		out[interval] = candles
	}
	return out
}

func buildVWAPChapter3Study(candles []exchanges.Candle, candlesByTimeframe map[string][]exchanges.Candle) vwapChapter3Study {
	contextRows := research.BuildContextFeatureRows(candles)
	if err := research.WriteContextFeaturesCSV("research/context_features.csv", contextRows); err != nil {
		log.Fatalf("write context features: %v", err)
	}

	timeframeRows := research.BuildTimeframeFeatureRows(candlesByTimeframe)
	if err := research.WriteTimeframeFeaturesCSV("research/timeframe_features.csv", timeframeRows); err != nil {
		log.Fatalf("write timeframe features: %v", err)
	}

	trendRows := research.BuildTrendAlignmentRows(candles)
	if err := research.WriteTrendAlignmentCSV("research/trend_alignment.csv", trendRows); err != nil {
		log.Fatalf("write trend alignment: %v", err)
	}

	summary := research.BuildVWAPChapter3Summary(contextRows)
	if err := research.WriteVWAPChapter3SummaryJSON("research/chapter3_summary.json", summary); err != nil {
		log.Fatalf("write vwap chapter 3 summary: %v", err)
	}

	return vwapChapter3Study{
		ContextRows:   contextRows,
		TimeframeRows: timeframeRows,
		TrendRows:     trendRows,
		Summary:       summary,
	}
}

func buildPriceActionStudy(candles []exchanges.Candle) priceActionStudy {
	rows := research.BuildPriceActionFeatureRows(candles)
	if err := research.WritePriceActionFeaturesCSV("research/price_action_features.csv", rows); err != nil {
		log.Fatalf("write price action features: %v", err)
	}
	summary := research.BuildPriceActionSummary(rows)
	if err := research.WritePriceActionSummaryJSON("research/price_action_summary.json", summary); err != nil {
		log.Fatalf("write price action summary: %v", err)
	}
	return priceActionStudy{Rows: rows, Summary: summary}
}

func buildPriceActionStrategyStudy(candles []exchanges.Candle) priceActionStrategyStudy {
	rows := research.BuildPriceActionStrategyStudyRows(candles)
	if err := research.WritePriceActionStrategyStudyCSV("research/price_action_strategy_study.csv", rows); err != nil {
		log.Fatalf("write price action strategy study: %v", err)
	}
	summary := research.BuildPriceActionStrategySummary(len(candles), rows)
	if err := research.WritePriceActionStrategySummaryJSON("research/price_action_strategy_summary.json", summary); err != nil {
		log.Fatalf("write price action strategy summary: %v", err)
	}
	return priceActionStrategyStudy{Rows: rows, Summary: summary}
}

func buildPriceActionPhase3Study(candles []exchanges.Candle) priceActionPhase3Study {
	rows := research.BuildPriceActionPhase3StudyRows(candles)
	if err := research.WritePriceActionPhase3StudyCSV("research/price_action_phase3_study.csv", rows); err != nil {
		log.Fatalf("write price action phase 3 study: %v", err)
	}
	summary := research.BuildPriceActionPhase3Summary(len(candles), rows)
	if err := research.WritePriceActionPhase3SummaryJSON("research/price_action_phase3_summary.json", summary); err != nil {
		log.Fatalf("write price action phase 3 summary: %v", err)
	}
	return priceActionPhase3Study{Rows: rows, Summary: summary}
}

func buildVolumeProfileStudy(candles []exchanges.Candle) volumeProfileStudy {
	profile := volumeprofile.BuildProfile(candles, volumeprofile.DefaultBinConfig())
	rows := research.BuildVolumeProfileFeatureRows(profile)
	if err := research.WriteVolumeProfileFeaturesCSV("research/volume_profile_features.csv", rows); err != nil {
		log.Fatalf("write volume profile features: %v", err)
	}
	summary := research.BuildVolumeProfileSummary([]volumeprofile.VolumeProfile{profile})
	if err := research.WriteVolumeProfileSummaryJSON("research/volume_profile_summary.json", summary); err != nil {
		log.Fatalf("write volume profile summary: %v", err)
	}
	return volumeProfileStudy{Profile: profile, Rows: rows, Summary: summary}
}

func buildScopedVolumeProfileStudy(candles []exchanges.Candle) scopedVolumeProfileStudy {
	profiles := volumeprofile.BuildScopedProfiles(candles, volumeprofile.DefaultBinConfig())
	rows := research.BuildScopedVolumeProfileFeatureRows(profiles)
	if err := research.WriteScopedVolumeProfileFeaturesCSV("research/volume_profile_scoped_features.csv", rows); err != nil {
		log.Fatalf("write scoped volume profile features: %v", err)
	}
	summary := research.BuildScopedVolumeProfileSummary(profiles)
	if err := research.WriteScopedVolumeProfileSummaryJSON("research/volume_profile_scoped_summary.json", summary); err != nil {
		log.Fatalf("write scoped volume profile summary: %v", err)
	}
	return scopedVolumeProfileStudy{Profiles: profiles, Rows: rows, Summary: summary}
}

func scopedPOC(profiles []volumeprofile.ScopedProfile, scope string) float64 {
	profile, ok := volumeprofile.ScopedProfileByName(profiles, scope)
	if !ok {
		return 0
	}
	return profile.Profile.POC
}

func buildVolumeProfileShapeStudy(profiles []volumeprofile.ScopedProfile) volumeProfileShapeStudy {
	rows := research.BuildVolumeProfileShapeStudyRows(profiles)
	if err := research.WriteVolumeProfileShapeStudyCSV("research/volume_profile_shape_study.csv", rows); err != nil {
		log.Fatalf("write volume profile shape study: %v", err)
	}
	summary := research.BuildVolumeProfileShapeSummary(rows)
	if err := research.WriteVolumeProfileShapeSummaryJSON("research/volume_profile_shape_summary.json", summary); err != nil {
		log.Fatalf("write volume profile shape summary: %v", err)
	}
	return volumeProfileShapeStudy{Rows: rows, Summary: summary}
}

func buildFlexibleVolumeProfileStudy(candles []exchanges.Candle, priceRows []research.PriceActionFeatureRow, strategyRows []research.PriceActionStrategyStudyRow, phase3Rows []research.PriceActionPhase3StudyRow) flexibleVolumeProfileStudy {
	profiles := research.BuildFlexibleVolumeProfiles(candles, priceRows, strategyRows, phase3Rows)
	rows := research.BuildFlexibleVolumeProfileRows(profiles)
	if err := research.WriteFlexibleVolumeProfileCSV("research/flexible_volume_profile.csv", rows); err != nil {
		log.Fatalf("write flexible volume profile csv: %v", err)
	}
	summary := research.BuildFlexibleVolumeProfileSummary(rows)
	if err := research.WriteFlexibleVolumeProfileSummaryJSON("research/flexible_volume_profile_summary.json", summary); err != nil {
		log.Fatalf("write flexible volume profile summary: %v", err)
	}
	return flexibleVolumeProfileStudy{Profiles: profiles, Rows: rows, Summary: summary}
}

func buildProfileAcceptanceQualityStudy() profileAcceptanceQualityStudy {
	rows, summary, err := research.BuildProfileAcceptanceQualityFromFiles(
		"research/flexible_volume_profile.csv",
		"research/volume_profile_shape_study.csv",
		"research/volume_profile_scoped_features.csv",
	)
	if err != nil {
		log.Fatalf("build profile acceptance quality: %v", err)
	}
	if err := research.WriteProfileAcceptanceQualityCSV("research/profile_acceptance_quality.csv", rows); err != nil {
		log.Fatalf("write profile acceptance quality csv: %v", err)
	}
	if err := research.WriteProfileAcceptanceQualitySummaryJSON("research/profile_acceptance_quality_summary.json", summary); err != nil {
		log.Fatalf("write profile acceptance quality summary: %v", err)
	}
	return profileAcceptanceQualityStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupAccumulationStudy() volumeSetupAccumulationStudy {
	rows, summary, err := research.BuildVolumeSetupAccumulationFromFiles(
		"research/flexible_volume_profile.csv",
		"research/price_action_features.csv",
		"research/price_action_strategy_study.csv",
		"research/volume_profile_scoped_features.csv",
		"research/profile_acceptance_quality.csv",
		volumeprofile.DefaultAccumulationSetupConfig(),
	)
	if err != nil {
		log.Fatalf("build volume setup accumulation: %v", err)
	}
	if err := research.WriteVolumeSetupAccumulationCSV("research/volume_setup_accumulation.csv", rows); err != nil {
		log.Fatalf("write volume setup accumulation csv: %v", err)
	}
	if err := research.WriteVolumeSetupAccumulationSummaryJSON("research/volume_setup_accumulation_summary.json", summary); err != nil {
		log.Fatalf("write volume setup accumulation summary: %v", err)
	}
	return volumeSetupAccumulationStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupAccumulationQualityStudy() volumeSetupAccumulationQualityStudy {
	rows, summary, err := research.BuildVolumeSetupAccumulationQualityFromFile("research/volume_setup_accumulation.csv")
	if err != nil {
		log.Fatalf("build volume setup accumulation quality: %v", err)
	}
	if err := research.WriteVolumeSetupAccumulationQualityCSV("research/volume_setup_accumulation_quality.csv", rows); err != nil {
		log.Fatalf("write volume setup accumulation quality csv: %v", err)
	}
	if err := research.WriteVolumeSetupAccumulationQualitySummaryJSON("research/volume_setup_accumulation_quality_summary.json", summary); err != nil {
		log.Fatalf("write volume setup accumulation quality summary: %v", err)
	}
	if err := research.WriteVolumeSetupAccumulationQualityMarkdown("research/volume_setup_accumulation_quality.md", rows, summary); err != nil {
		log.Fatalf("write volume setup accumulation quality markdown: %v", err)
	}
	return volumeSetupAccumulationQualityStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupTrendStudy() volumeSetupTrendStudy {
	rows, summary, err := research.BuildVolumeSetupTrendFromFiles(
		"research/price_action_features.csv",
		"research/price_action_strategy_study.csv",
		"research/price_action_phase3_study.csv",
		"research/flexible_volume_profile.csv",
		"research/volume_profile_scoped_features.csv",
		"research/profile_acceptance_quality.csv",
		volumeprofile.DefaultTrendSetupConfig(),
	)
	if err != nil {
		log.Fatalf("build volume setup trend: %v", err)
	}
	if err := research.WriteVolumeSetupTrendCSV("research/volume_setup_trend.csv", rows); err != nil {
		log.Fatalf("write volume setup trend csv: %v", err)
	}
	if err := research.WriteVolumeSetupTrendSummaryJSON("research/volume_setup_trend_summary.json", summary); err != nil {
		log.Fatalf("write volume setup trend summary: %v", err)
	}
	return volumeSetupTrendStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupTrendQualityStudy() volumeSetupTrendQualityStudy {
	rows, summary, err := research.BuildVolumeSetupTrendQualityFromFile("research/volume_setup_trend.csv")
	if err != nil {
		log.Fatalf("build volume setup trend quality: %v", err)
	}
	if err := research.WriteVolumeSetupTrendQualityCSV("research/volume_setup_trend_quality.csv", rows); err != nil {
		log.Fatalf("write volume setup trend quality csv: %v", err)
	}
	if err := research.WriteVolumeSetupTrendQualitySummaryJSON("research/volume_setup_trend_quality_summary.json", summary); err != nil {
		log.Fatalf("write volume setup trend quality summary: %v", err)
	}
	if err := research.WriteVolumeSetupTrendQualityMarkdown("research/volume_setup_trend_quality.md", rows, summary); err != nil {
		log.Fatalf("write volume setup trend quality markdown: %v", err)
	}
	return volumeSetupTrendQualityStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupRejectionStudy() volumeSetupRejectionStudy {
	rows, summary, err := research.BuildVolumeSetupRejectionFromFiles(
		"research/price_action_features.csv",
		"research/price_action_phase3_study.csv",
		"research/flexible_volume_profile.csv",
		"research/volume_profile_scoped_features.csv",
		"research/profile_acceptance_quality.csv",
		"research/vwap_features.csv",
		"research/context_features.csv",
		"research/context_features_l2.csv",
		volumeprofile.DefaultRejectionSetupConfig(),
	)
	if err != nil {
		log.Fatalf("build volume setup rejection: %v", err)
	}
	if err := research.WriteVolumeSetupRejectionCSV("research/volume_setup_rejection.csv", rows); err != nil {
		log.Fatalf("write volume setup rejection csv: %v", err)
	}
	if err := research.WriteVolumeSetupRejectionSummaryJSON("research/volume_setup_rejection_summary.json", summary); err != nil {
		log.Fatalf("write volume setup rejection summary: %v", err)
	}
	return volumeSetupRejectionStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupRejectionQualityStudy() volumeSetupRejectionQualityStudy {
	rows, summary, err := research.BuildVolumeSetupRejectionQualityFromFile("research/volume_setup_rejection.csv")
	if err != nil {
		log.Fatalf("build volume setup rejection quality: %v", err)
	}
	if err := research.WriteVolumeSetupRejectionQualityCSV("research/volume_setup_rejection_quality.csv", rows); err != nil {
		log.Fatalf("write volume setup rejection quality csv: %v", err)
	}
	if err := research.WriteVolumeSetupRejectionQualitySummaryJSON("research/volume_setup_rejection_quality_summary.json", summary); err != nil {
		log.Fatalf("write volume setup rejection quality summary: %v", err)
	}
	if err := research.WriteVolumeSetupRejectionQualityMarkdown("research/volume_setup_rejection_quality.md", rows, summary); err != nil {
		log.Fatalf("write volume setup rejection quality markdown: %v", err)
	}
	return volumeSetupRejectionQualityStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupReversalStudy() volumeSetupReversalStudy {
	rows, summary, err := research.BuildVolumeSetupReversalFromFiles(
		"research/volume_setup_rejection.csv",
		"research/volume_setup_rejection_quality.csv",
		"research/price_action_phase3_study.csv",
		"research/flexible_volume_profile.csv",
		"research/volume_profile_scoped_features.csv",
		"research/vwap_features.csv",
		"research/context_features.csv",
		volumeprofile.DefaultReversalSetupConfig(),
	)
	if err != nil {
		log.Fatalf("build volume setup reversal: %v", err)
	}
	if err := research.WriteVolumeSetupReversalCSV("research/volume_setup_reversal.csv", rows); err != nil {
		log.Fatalf("write volume setup reversal csv: %v", err)
	}
	if err := research.WriteVolumeSetupReversalSummaryJSON("research/volume_setup_reversal_summary.json", summary); err != nil {
		log.Fatalf("write volume setup reversal summary: %v", err)
	}
	return volumeSetupReversalStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupReversalQualityStudy() volumeSetupReversalQualityStudy {
	rows, summary, err := research.BuildVolumeSetupReversalQualityFromFile("research/volume_setup_reversal.csv")
	if err != nil {
		log.Fatalf("build volume setup reversal quality: %v", err)
	}
	if err := research.WriteVolumeSetupReversalQualityCSV("research/volume_setup_reversal_quality.csv", rows); err != nil {
		log.Fatalf("write volume setup reversal quality csv: %v", err)
	}
	if err := research.WriteVolumeSetupReversalQualitySummaryJSON("research/volume_setup_reversal_quality_summary.json", summary); err != nil {
		log.Fatalf("write volume setup reversal quality summary: %v", err)
	}
	if err := research.WriteVolumeSetupReversalQualityMarkdown("research/volume_setup_reversal_quality.md", rows, summary); err != nil {
		log.Fatalf("write volume setup reversal quality markdown: %v", err)
	}
	return volumeSetupReversalQualityStudy{Rows: rows, Summary: summary}
}

func buildVolumeSetupComparisonStudy() volumeSetupComparisonStudy {
	rows, summary, err := research.BuildVolumeSetupComparisonFromFiles(
		"research/volume_setup_accumulation.csv",
		"research/volume_setup_trend.csv",
		"research/volume_setup_rejection.csv",
		"research/volume_setup_reversal.csv",
	)
	if err != nil {
		log.Fatalf("build volume setup comparison: %v", err)
	}
	if err := research.WriteVolumeSetupComparisonCSV("research/volume_setup_comparison.csv", rows); err != nil {
		log.Fatalf("write volume setup comparison csv: %v", err)
	}
	if err := research.WriteVolumeSetupComparisonJSON("research/volume_setup_comparison.json", summary); err != nil {
		log.Fatalf("write volume setup comparison json: %v", err)
	}
	if err := research.WriteVolumeSetupComparisonMarkdown("research/volume_setup_comparison.md", rows, summary); err != nil {
		log.Fatalf("write volume setup comparison markdown: %v", err)
	}
	return volumeSetupComparisonStudy{Rows: rows, Summary: summary}
}

func buildSetupFeatureLabelExport(candles []exchanges.Candle) setupFeatureLabelStudy {
	rows, summary, err := research.BuildSetupFeatureLabelExport(research.SetupFeatureLabelPaths{
		ReversalPath:     "research/volume_setup_reversal.csv",
		AccumulationPath: "research/volume_setup_accumulation.csv",
		TrendPath:        "research/volume_setup_trend.csv",
		RejectionPath:    "research/volume_setup_rejection.csv",
		VWAPPath:         "research/vwap_features.csv",
		ContextPath:      "research/context_features.csv",
		ScopedPath:       "research/volume_profile_scoped_features.csv",
		PriceActionPath:  "research/price_action_features.csv",
	}, candles)
	if err != nil {
		log.Fatalf("build setup feature label export: %v", err)
	}
	if err := research.WriteSetupFeatureLabelCSV("research/setup_features_labels.csv", rows); err != nil {
		log.Fatalf("write setup feature label csv: %v", err)
	}
	if err := research.WriteSetupLabelsSummaryJSON("research/setup_labels_summary.json", summary); err != nil {
		log.Fatalf("write setup labels summary: %v", err)
	}
	return setupFeatureLabelStudy{Rows: rows, Summary: summary}
}

func buildInstrumentUniverseStudy(orderBookRows []research.OrderBookFeatureRow, venues []venueCandles) instrumentUniverseStudy {
	candleCounts := map[string]int{}
	for _, venue := range venues {
		candleCounts[venue.Venue+":"+venue.Symbol] = len(venue.Candles)
	}
	rows := research.BuildInstrumentUniverse(orderBookRows, candleCounts)
	summary := research.BuildInstrumentUniverseSummary(rows)
	if err := research.WriteInstrumentUniverseCSV("research/instrument_universe.csv", rows); err != nil {
		log.Fatalf("write instrument universe csv: %v", err)
	}
	if err := research.WriteInstrumentUniverseSummaryJSON("research/instrument_universe_summary.json", summary); err != nil {
		log.Fatalf("write instrument universe summary: %v", err)
	}
	return instrumentUniverseStudy{Rows: rows, Summary: summary}
}

func buildVolumeProfileFinalBookPacket() volumeProfileFinalBookStudy {
	archived := archivedResearchFiles()
	packet := research.BuildVolumeProfileFinalBookPacket(archived)
	if err := research.WriteVolumeProfileFinalBookPacketJSON("research/volume_profile_final_book_packet.json", packet); err != nil {
		log.Fatalf("write volume profile final packet json: %v", err)
	}
	if err := research.WriteVolumeProfileFinalBookPacketMarkdown("research/volume_profile_final_book_packet.md", packet); err != nil {
		log.Fatalf("write volume profile final packet markdown: %v", err)
	}
	return volumeProfileFinalBookStudy{Packet: packet}
}

func buildOrderFlowFoundationStudy() orderFlowFoundationStudy {
	rows, summary := research.BuildOrderFlowFoundationRows(research.SyntheticOrderFlowFootprints())
	if err := research.WriteOrderFlowFoundationCSV("research/orderflow_foundation.csv", rows); err != nil {
		log.Fatalf("write orderflow foundation csv: %v", err)
	}
	if err := research.WriteOrderFlowFoundationSummaryJSON("research/orderflow_foundation_summary.json", summary); err != nil {
		log.Fatalf("write orderflow foundation summary: %v", err)
	}
	return orderFlowFoundationStudy{Rows: rows, Summary: summary}
}

func buildOrderFlowDataCapabilityAuditStudy() orderFlowDataCapabilityAuditStudy {
	report := research.BuildOrderFlowDataCapabilityAudit()
	if err := research.WriteOrderFlowDataCapabilityAuditMarkdown("research/orderflow_data_capability_audit.md", report); err != nil {
		log.Fatalf("write orderflow data capability audit markdown: %v", err)
	}
	if err := research.WriteOrderFlowDataCapabilityAuditJSON("research/orderflow_data_capability_audit.json", report); err != nil {
		log.Fatalf("write orderflow data capability audit json: %v", err)
	}
	if err := research.WriteLighterTradeTapeAdapterPlanMarkdown("research/lighter_trade_tape_adapter_plan.md"); err != nil {
		log.Fatalf("write lighter trade tape adapter plan: %v", err)
	}
	return orderFlowDataCapabilityAuditStudy{Report: report}
}

func buildTradeTapeAnalysis() (tradetape.TradeTapeAnalysis, bool) {
	cfg := tradetape.DefaultRecorderConfig()
	if _, err := os.Stat(cfg.OutputPath); err != nil {
		log.Printf("skip trade tape analysis: %v", err)
		return tradetape.TradeTapeAnalysis{}, false
	}
	analysis, err := tradetape.AnalyzeCSV(cfg.OutputPath)
	if err != nil {
		log.Printf("skip trade tape analysis: %v", err)
		return tradetape.TradeTapeAnalysis{}, false
	}
	if err := tradetape.WriteAnalysisJSON("research/trade_tape_analysis.json", analysis); err != nil {
		log.Fatalf("write trade tape analysis json: %v", err)
	}
	if err := tradetape.WriteAnalysisMarkdown("research/trade_tape_analysis.md", analysis); err != nil {
		log.Fatalf("write trade tape analysis markdown: %v", err)
	}
	return analysis, true
}

func buildOrderFlowRealFootprints() (orderFlowRealFootprintsStudy, bool) {
	cfg := tradetape.DefaultRecorderConfig()
	if _, err := os.Stat(cfg.OutputPath); err != nil {
		log.Printf("skip real footprint build: %v", err)
		return orderFlowRealFootprintsStudy{}, false
	}
	bars, rows, summary, err := research.BuildRealFootprintReport(cfg.OutputPath)
	if err != nil {
		log.Printf("skip real footprint build: %v", err)
		return orderFlowRealFootprintsStudy{}, false
	}
	if err := research.WriteOrderFlowRealFootprintsCSV("research/orderflow_real_footprints.csv", rows); err != nil {
		log.Fatalf("write orderflow real footprints csv: %v", err)
	}
	if err := research.WriteOrderFlowRealFootprintsSummaryJSON("research/orderflow_real_footprints_summary.json", summary); err != nil {
		log.Fatalf("write orderflow real footprints summary: %v", err)
	}
	if err := research.WriteOrderFlowRealFootprintsMarkdown("research/orderflow_real_footprints.md", summary); err != nil {
		log.Fatalf("write orderflow real footprints markdown: %v", err)
	}
	return orderFlowRealFootprintsStudy{Bars: bars, Rows: rows, Summary: summary}, true
}

func buildOrderFlowVolumeClusters(footprintRows []research.OrderFlowRealFootprintRow, ok bool) (orderFlowVolumeClusterStudy, bool) {
	if !ok {
		return orderFlowVolumeClusterStudy{}, false
	}
	rows, summary := research.BuildOrderFlowVolumeClusterStudy(footprintRows)
	if err := research.WriteOrderFlowVolumeClusterCSV("research/orderflow_setup_volume_cluster.csv", rows); err != nil {
		log.Fatalf("write orderflow volume cluster csv: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterSummaryJSON("research/orderflow_setup_volume_cluster_summary.json", summary); err != nil {
		log.Fatalf("write orderflow volume cluster summary: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterMarkdown("research/orderflow_setup_volume_cluster.md", rows, summary); err != nil {
		log.Fatalf("write orderflow volume cluster markdown: %v", err)
	}
	return orderFlowVolumeClusterStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowVolumeClusterQuality(ok bool) (orderFlowVolumeClusterQualityStudy, bool) {
	if !ok {
		return orderFlowVolumeClusterQualityStudy{}, false
	}
	rows, summary, err := research.BuildOrderFlowVolumeClusterQualityFromFile("research/orderflow_setup_volume_cluster.csv")
	if err != nil {
		log.Printf("skip orderflow volume cluster quality: %v", err)
		return orderFlowVolumeClusterQualityStudy{}, false
	}
	if err := research.WriteOrderFlowVolumeClusterQualityCSV("research/orderflow_volume_cluster_quality.csv", rows); err != nil {
		log.Fatalf("write orderflow volume cluster quality csv: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterQualityJSON("research/orderflow_volume_cluster_quality.json", summary); err != nil {
		log.Fatalf("write orderflow volume cluster quality json: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterQualityMarkdown("research/orderflow_volume_cluster_quality.md", rows, summary); err != nil {
		log.Fatalf("write orderflow volume cluster quality markdown: %v", err)
	}
	return orderFlowVolumeClusterQualityStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowMultipleHVNs(footprintRows []research.OrderFlowRealFootprintRow, ok bool) (orderFlowMultipleHVNStudy, bool) {
	if !ok {
		return orderFlowMultipleHVNStudy{}, false
	}
	rows, summary := research.BuildOrderFlowMultipleHVNStudy(footprintRows)
	if err := research.WriteOrderFlowMultipleHVNCSV("research/orderflow_setup_multiple_hvn.csv", rows); err != nil {
		log.Fatalf("write orderflow multiple hvn csv: %v", err)
	}
	if err := research.WriteOrderFlowMultipleHVNSummaryJSON("research/orderflow_setup_multiple_hvn_summary.json", summary); err != nil {
		log.Fatalf("write orderflow multiple hvn summary: %v", err)
	}
	clusterSummary, err := research.ReadOrderFlowVolumeClusterSummaryJSON("research/orderflow_setup_volume_cluster_summary.json")
	if err != nil {
		clusterSummary = research.OrderFlowVolumeClusterSummary{}
	}
	if err := research.WriteOrderFlowMultipleHVNMarkdown("research/orderflow_setup_multiple_hvn.md", rows, summary, clusterSummary); err != nil {
		log.Fatalf("write orderflow multiple hvn markdown: %v", err)
	}
	return orderFlowMultipleHVNStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowBackfilledFootprints() (orderFlowBackfilledFootprintsStudy, bool) {
	path := config.DataPath("trade_tape", "backfill", "aster_btcusdt_deduped.csv")
	if _, err := os.Stat(path); err != nil {
		log.Printf("skip backfilled footprint replay: %v", err)
		return orderFlowBackfilledFootprintsStudy{}, false
	}
	bars, rows, summary, err := research.BuildOrderFlowBackfilledFootprints(path)
	if err != nil {
		log.Printf("skip backfilled footprint replay: %v", err)
		return orderFlowBackfilledFootprintsStudy{}, false
	}
	if err := research.WriteOrderFlowBackfilledFootprintsCSV("research/orderflow_backfilled_footprints.csv", rows); err != nil {
		log.Fatalf("write orderflow backfilled footprints csv: %v", err)
	}
	if err := research.WriteOrderFlowBackfilledFootprintsSummaryJSON("research/orderflow_backfilled_footprints_summary.json", summary); err != nil {
		log.Fatalf("write orderflow backfilled footprints summary: %v", err)
	}
	if err := research.WriteOrderFlowBackfilledFootprintsMarkdown("research/orderflow_backfilled_footprints.md", summary); err != nil {
		log.Fatalf("write orderflow backfilled footprints markdown: %v", err)
	}
	return orderFlowBackfilledFootprintsStudy{Bars: bars, Rows: rows, Summary: summary}, true
}

func buildOrderFlowBackfilledSetupReplay(
	footprintRows []research.OrderFlowRealFootprintRow,
	footprintSummary research.OrderFlowBackfilledFootprintSummary,
	ok bool,
	liveTradeRows int,
	liveFootprintSummary research.OrderFlowRealFootprintSummary,
	liveVolumeClusterSummary research.OrderFlowVolumeClusterSummary,
	liveMultipleHVNSummary research.OrderFlowMultipleHVNSummary,
) (orderFlowBackfilledSetupReplayStudy, bool) {
	if !ok {
		return orderFlowBackfilledSetupReplayStudy{}, false
	}
	replay := research.BuildOrderFlowBackfilledSetupReplay(footprintRows)
	if err := research.WriteOrderFlowVolumeClusterCSV("research/orderflow_backfilled_volume_cluster.csv", replay.VolumeClusterRows); err != nil {
		log.Fatalf("write orderflow backfilled volume cluster csv: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterSummaryJSON("research/orderflow_backfilled_volume_cluster_summary.json", replay.VolumeClusterSummary); err != nil {
		log.Fatalf("write orderflow backfilled volume cluster summary: %v", err)
	}
	if err := research.WriteOrderFlowBackfilledVolumeClusterMarkdown("research/orderflow_backfilled_volume_cluster.md", replay.VolumeClusterRows, replay.VolumeClusterSummary); err != nil {
		log.Fatalf("write orderflow backfilled volume cluster markdown: %v", err)
	}
	if err := research.WriteOrderFlowMultipleHVNCSV("research/orderflow_backfilled_multiple_hvn.csv", replay.MultipleHVNRows); err != nil {
		log.Fatalf("write orderflow backfilled multiple hvn csv: %v", err)
	}
	if err := research.WriteOrderFlowMultipleHVNSummaryJSON("research/orderflow_backfilled_multiple_hvn_summary.json", replay.MultipleHVNSummary); err != nil {
		log.Fatalf("write orderflow backfilled multiple hvn summary: %v", err)
	}
	if err := research.WriteOrderFlowBackfilledMultipleHVNMarkdown("research/orderflow_backfilled_multiple_hvn.md", replay.MultipleHVNRows, replay.MultipleHVNSummary, replay.VolumeClusterSummary); err != nil {
		log.Fatalf("write orderflow backfilled multiple hvn markdown: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterQualityCSV("research/orderflow_backfilled_volume_cluster_quality.csv", replay.QualityRows); err != nil {
		log.Fatalf("write orderflow backfilled volume cluster quality csv: %v", err)
	}
	if err := research.WriteOrderFlowVolumeClusterQualityJSON("research/orderflow_backfilled_volume_cluster_quality.json", replay.QualitySummary); err != nil {
		log.Fatalf("write orderflow backfilled volume cluster quality json: %v", err)
	}
	if err := research.WriteOrderFlowBackfilledVolumeClusterQualityMarkdown("research/orderflow_backfilled_volume_cluster_quality.md", replay.QualityRows, replay.QualitySummary); err != nil {
		log.Fatalf("write orderflow backfilled volume cluster quality markdown: %v", err)
	}
	comparison := research.BuildOrderFlowLiveVsBackfillComparison(
		liveTradeRows,
		liveFootprintSummary,
		liveVolumeClusterSummary,
		liveMultipleHVNSummary,
		footprintSummary,
		replay.VolumeClusterSummary,
		replay.MultipleHVNSummary,
	)
	if err := research.WriteOrderFlowLiveVsBackfillComparisonJSON("research/orderflow_live_vs_backfill_comparison.json", comparison); err != nil {
		log.Fatalf("write orderflow live vs backfill comparison json: %v", err)
	}
	if err := research.WriteOrderFlowLiveVsBackfillComparisonMarkdown("research/orderflow_live_vs_backfill_comparison.md", comparison); err != nil {
		log.Fatalf("write orderflow live vs backfill comparison markdown: %v", err)
	}
	return orderFlowBackfilledSetupReplayStudy{Replay: replay, Comparison: comparison}, true
}

func buildOrderFlowTradesFilter(
	footprintRows []research.OrderFlowRealFootprintRow,
	multipleHVNRows []research.OrderFlowMultipleHVNSetupRow,
	ok bool,
	volumeClusterSummary research.OrderFlowVolumeClusterSummary,
	multipleHVNSummary research.OrderFlowMultipleHVNSummary,
) (orderFlowTradesFilterStudy, bool) {
	if !ok {
		return orderFlowTradesFilterStudy{}, false
	}
	rows, summary, err := research.BuildOrderFlowTradesFilterStudyFromFile(
		config.DataPath("trade_tape", "backfill", "aster_btcusdt_deduped.csv"),
		footprintRows,
		multipleHVNRows,
	)
	if err != nil {
		log.Printf("skip orderflow trades filter: %v", err)
		return orderFlowTradesFilterStudy{}, false
	}
	if err := research.WriteOrderFlowTradesFilterCSV("research/orderflow_setup_trades_filter.csv", rows); err != nil {
		log.Fatalf("write orderflow trades filter csv: %v", err)
	}
	if err := research.WriteOrderFlowTradesFilterSummaryJSON("research/orderflow_setup_trades_filter_summary.json", summary); err != nil {
		log.Fatalf("write orderflow trades filter summary: %v", err)
	}
	if err := research.WriteOrderFlowTradesFilterMarkdown("research/orderflow_setup_trades_filter.md", rows, summary, volumeClusterSummary, multipleHVNSummary); err != nil {
		log.Fatalf("write orderflow trades filter markdown: %v", err)
	}
	return orderFlowTradesFilterStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowStackedImbalances(
	bars []orderflow.FootprintBar,
	footprintRows []research.OrderFlowRealFootprintRow,
	multipleHVNRows []research.OrderFlowMultipleHVNSetupRow,
	ok bool,
	volumeClusterSummary research.OrderFlowVolumeClusterSummary,
	multipleHVNSummary research.OrderFlowMultipleHVNSummary,
	tradesFilterSummary research.OrderFlowTradesFilterSummary,
) (orderFlowStackedImbalanceStudy, bool) {
	if !ok {
		return orderFlowStackedImbalanceStudy{}, false
	}
	rows, summary := research.BuildOrderFlowStackedImbalanceStudy(bars, footprintRows, multipleHVNRows)
	if err := research.WriteOrderFlowStackedImbalanceCSV("research/orderflow_setup_stacked_imbalance.csv", rows); err != nil {
		log.Fatalf("write orderflow stacked imbalance csv: %v", err)
	}
	if err := research.WriteOrderFlowStackedImbalanceSummaryJSON("research/orderflow_setup_stacked_imbalance_summary.json", summary); err != nil {
		log.Fatalf("write orderflow stacked imbalance summary: %v", err)
	}
	if err := research.WriteOrderFlowStackedImbalanceMarkdown("research/orderflow_setup_stacked_imbalance.md", rows, summary, volumeClusterSummary, multipleHVNSummary, tradesFilterSummary); err != nil {
		log.Fatalf("write orderflow stacked imbalance markdown: %v", err)
	}
	return orderFlowStackedImbalanceStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowUnfinishedBusiness(
	bars []orderflow.FootprintBar,
	footprintRows []research.OrderFlowRealFootprintRow,
	ok bool,
	volumeClusterSummary research.OrderFlowVolumeClusterSummary,
	multipleHVNSummary research.OrderFlowMultipleHVNSummary,
	tradesFilterSummary research.OrderFlowTradesFilterSummary,
	stackedSummary research.OrderFlowStackedImbalanceSummary,
) (orderFlowUnfinishedBusinessStudy, bool) {
	if !ok {
		return orderFlowUnfinishedBusinessStudy{}, false
	}
	rows, summary := research.BuildOrderFlowUnfinishedBusinessStudy(bars, footprintRows)
	if err := research.WriteOrderFlowUnfinishedBusinessCSV("research/orderflow_setup_unfinished_business.csv", rows); err != nil {
		log.Fatalf("write orderflow unfinished business csv: %v", err)
	}
	if err := research.WriteOrderFlowUnfinishedBusinessSummaryJSON("research/orderflow_setup_unfinished_business_summary.json", summary); err != nil {
		log.Fatalf("write orderflow unfinished business summary: %v", err)
	}
	if err := research.WriteOrderFlowUnfinishedBusinessMarkdown("research/orderflow_setup_unfinished_business.md", rows, summary, volumeClusterSummary, multipleHVNSummary, tradesFilterSummary, stackedSummary); err != nil {
		log.Fatalf("write orderflow unfinished business markdown: %v", err)
	}
	return orderFlowUnfinishedBusinessStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowBigLimitOrders(
	liveFootprintRows []research.OrderFlowRealFootprintRow,
	footprintRows []research.OrderFlowRealFootprintRow,
	tradesFilterRows []research.OrderFlowTradesFilterSetupRow,
	ok bool,
	tradesFilterSummary research.OrderFlowTradesFilterSummary,
	stackedSummary research.OrderFlowStackedImbalanceSummary,
	unfinishedSummary research.OrderFlowUnfinishedBusinessSummary,
) (orderFlowBigLimitOrderStudy, bool) {
	if !ok {
		return orderFlowBigLimitOrderStudy{}, false
	}
	sources := []research.OrderFlowBigLimitOrderSourceInput{
		{Source: research.LiveMultiVenueFootprintSource, FootprintRows: liveFootprintRows},
		{Source: research.BackfilledAsterAggTradesSource, FootprintRows: footprintRows},
	}
	rows, summary := research.BuildOrderFlowBigLimitOrderConfirmationsForSources(sources, tradesFilterRows)
	if err := research.WriteOrderFlowBigLimitOrderCSV("research/orderflow_confirmation_big_limit_orders.csv", rows); err != nil {
		log.Fatalf("write orderflow big limit order csv: %v", err)
	}
	if err := research.WriteOrderFlowBigLimitOrderSummaryJSON("research/orderflow_confirmation_big_limit_orders_summary.json", summary); err != nil {
		log.Fatalf("write orderflow big limit order summary: %v", err)
	}
	if err := research.WriteOrderFlowBigLimitOrderMarkdown("research/orderflow_confirmation_big_limit_orders.md", rows, summary, tradesFilterSummary, stackedSummary, unfinishedSummary); err != nil {
		log.Fatalf("write orderflow big limit order markdown: %v", err)
	}
	return orderFlowBigLimitOrderStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowAbsorption(
	liveFootprintRows []research.OrderFlowRealFootprintRow,
	footprintRows []research.OrderFlowRealFootprintRow,
	tradesFilterRows []research.OrderFlowTradesFilterSetupRow,
	ok bool,
	bigLimitSummary research.OrderFlowBigLimitOrderSummary,
	tradesFilterSummary research.OrderFlowTradesFilterSummary,
	stackedSummary research.OrderFlowStackedImbalanceSummary,
) (orderFlowAbsorptionStudy, bool) {
	if !ok {
		return orderFlowAbsorptionStudy{}, false
	}
	sources := []research.OrderFlowBigLimitOrderSourceInput{
		{Source: research.LiveMultiVenueFootprintSource, FootprintRows: liveFootprintRows},
		{Source: research.BackfilledAsterAggTradesSource, FootprintRows: footprintRows},
	}
	rows, summary := research.BuildOrderFlowAbsorptionConfirmationsForSources(sources, tradesFilterRows)
	if err := research.WriteOrderFlowAbsorptionCSV("research/orderflow_confirmation_absorption.csv", rows); err != nil {
		log.Fatalf("write orderflow absorption csv: %v", err)
	}
	if err := research.WriteOrderFlowAbsorptionSummaryJSON("research/orderflow_confirmation_absorption_summary.json", summary); err != nil {
		log.Fatalf("write orderflow absorption summary: %v", err)
	}
	if err := research.WriteOrderFlowAbsorptionMarkdown("research/orderflow_confirmation_absorption.md", rows, summary, bigLimitSummary, tradesFilterSummary, stackedSummary); err != nil {
		log.Fatalf("write orderflow absorption markdown: %v", err)
	}
	return orderFlowAbsorptionStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowAggressiveDelta(
	liveFootprintRows []research.OrderFlowRealFootprintRow,
	footprintRows []research.OrderFlowRealFootprintRow,
	tradesFilterRows []research.OrderFlowTradesFilterSetupRow,
	ok bool,
) (orderFlowAggressiveDeltaStudy, bool) {
	if !ok {
		return orderFlowAggressiveDeltaStudy{}, false
	}
	sources := []research.OrderFlowBigLimitOrderSourceInput{
		{Source: research.LiveMultiVenueFootprintSource, FootprintRows: liveFootprintRows},
		{Source: research.BackfilledAsterAggTradesSource, FootprintRows: footprintRows},
	}
	rows, summary := research.BuildOrderFlowAggressiveDeltaForSources(sources, tradesFilterRows)
	if err := research.WriteOrderFlowAggressiveDeltaCSV("research/orderflow_confirmation_aggressive_delta.csv", rows); err != nil {
		log.Fatalf("write orderflow aggressive delta csv: %v", err)
	}
	if err := research.WriteOrderFlowAggressiveDeltaSummaryJSON("research/orderflow_confirmation_aggressive_delta_summary.json", summary); err != nil {
		log.Fatalf("write orderflow aggressive delta summary: %v", err)
	}
	if err := research.WriteOrderFlowAggressiveDeltaMarkdown("research/orderflow_confirmation_aggressive_delta.md", rows, summary); err != nil {
		log.Fatalf("write orderflow aggressive delta markdown: %v", err)
	}
	return orderFlowAggressiveDeltaStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowCumulativeDeltaDivergence(
	liveFootprintRows []research.OrderFlowRealFootprintRow,
	footprintRows []research.OrderFlowRealFootprintRow,
	tradesFilterRows []research.OrderFlowTradesFilterSetupRow,
	ok bool,
) (orderFlowCumulativeDeltaDivergenceStudy, bool) {
	if !ok {
		return orderFlowCumulativeDeltaDivergenceStudy{}, false
	}
	sources := []research.OrderFlowBigLimitOrderSourceInput{
		{Source: research.LiveMultiVenueFootprintSource, FootprintRows: liveFootprintRows},
		{Source: research.BackfilledAsterAggTradesSource, FootprintRows: footprintRows},
	}
	rows, summary := research.BuildOrderFlowCumulativeDeltaDivergenceForSources(sources, tradesFilterRows)
	if err := research.WriteOrderFlowCumulativeDeltaDivergenceCSV("research/orderflow_confirmation_cumulative_delta_divergence.csv", rows); err != nil {
		log.Fatalf("write orderflow cumulative delta divergence csv: %v", err)
	}
	if err := research.WriteOrderFlowCumulativeDeltaDivergenceSummaryJSON("research/orderflow_confirmation_cumulative_delta_divergence_summary.json", summary); err != nil {
		log.Fatalf("write orderflow cumulative delta divergence summary: %v", err)
	}
	if err := research.WriteOrderFlowCumulativeDeltaDivergenceMarkdown("research/orderflow_confirmation_cumulative_delta_divergence.md", rows, summary); err != nil {
		log.Fatalf("write orderflow cumulative delta divergence markdown: %v", err)
	}
	return orderFlowCumulativeDeltaDivergenceStudy{Rows: rows, Summary: summary}, true
}

func buildOrderFlowConfirmationComparison(
	bigLimitSummary research.OrderFlowBigLimitOrderSummary,
	absorptionSummary research.OrderFlowAbsorptionSummary,
	aggressiveSummary research.OrderFlowAggressiveDeltaSummary,
	divergenceSummary research.OrderFlowCumulativeDeltaDivergenceSummary,
	ok bool,
) (orderFlowConfirmationComparisonStudy, bool) {
	if !ok {
		return orderFlowConfirmationComparisonStudy{}, false
	}
	rows, summary := research.BuildOrderFlowConfirmationComparison(bigLimitSummary, absorptionSummary, aggressiveSummary, divergenceSummary)
	if err := research.WriteOrderFlowConfirmationComparisonCSV("research/orderflow_confirmation_comparison.csv", rows); err != nil {
		log.Fatalf("write orderflow confirmation comparison csv: %v", err)
	}
	if err := research.WriteOrderFlowConfirmationComparisonJSON("research/orderflow_confirmation_comparison.json", summary); err != nil {
		log.Fatalf("write orderflow confirmation comparison json: %v", err)
	}
	if err := research.WriteOrderFlowConfirmationComparisonMarkdown("research/orderflow_confirmation_comparison.md", rows, summary); err != nil {
		log.Fatalf("write orderflow confirmation comparison markdown: %v", err)
	}
	return orderFlowConfirmationComparisonStudy{Rows: rows, Summary: summary}, true
}

func formatOrderFlowVenueCoverage(coverage map[string]research.OrderFlowVenueCoverage) string {
	parts := make([]string, 0, 3)
	for _, venue := range []string{"aster", "hyperliquid", "lighter"} {
		row := coverage[venue]
		count := row.Setups
		if row.Confirmations > 0 || row.Setups == 0 {
			count = row.Confirmations
		}
		parts = append(parts, fmt.Sprintf("%s:%d", venue, count))
	}
	return strings.Join(parts, ",")
}

func archivedResearchFiles() []string {
	entries, err := os.ReadDir("research/archive")
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		out = append(out, "research/archive/"+entry.Name())
	}
	sort.Strings(out)
	return out
}

func buildVolumeProfileBookCompletionPacket() volumeProfileBookCompletionStudy {
	packet := research.BuildVolumeProfileBookCompletionPacket()
	if err := research.WriteVolumeProfileBookCompletionPacketJSON("research/volume_profile_book_completion_packet.json", packet); err != nil {
		log.Fatalf("write volume profile book completion packet json: %v", err)
	}
	if err := research.WriteVolumeProfileBookCompletionPacketMarkdown("research/volume_profile_book_completion_packet.md", packet); err != nil {
		log.Fatalf("write volume profile book completion packet markdown: %v", err)
	}
	return volumeProfileBookCompletionStudy{Packet: packet}
}

func fetchHyperliquidL2Snapshot(symbol string) *orderbook.OrderBookSnapshot {
	snapshot, err := hyperliquid.GetMainnetL2OrderBookSnapshot(symbol)
	if err != nil {
		log.Printf("skip hyperliquid mainnet l2 snapshot %s: %v", symbol, err)
		return nil
	}
	return snapshot
}

func fetchMultiVenueL2Snapshots() []orderbook.OrderBookSnapshot {
	out := make([]orderbook.OrderBookSnapshot, 0, 3)

	if snapshot := fetchHyperliquidL2Snapshot("BTC"); snapshot != nil {
		out = append(out, *snapshot)
	}

	asterSnapshot, err := aster.GetMainnetOrderBook("BTCUSDT")
	if err != nil {
		log.Printf("skip aster mainnet l2 snapshot BTCUSDT: %v", err)
	} else {
		out = append(out, asterSnapshot)
	}

	lighterSnapshot, err := lighter.GetMainnetOrderBook("BTC")
	if err != nil {
		log.Printf("skip lighter mainnet l2 snapshot BTC: %v", err)
	} else {
		out = append(out, lighterSnapshot)
	}

	return out
}

func snapshotByVenue(snapshots []orderbook.OrderBookSnapshot, venue string) *orderbook.OrderBookSnapshot {
	for i := range snapshots {
		if snapshots[i].Venue == venue {
			return &snapshots[i]
		}
	}
	return nil
}

func printHyperliquidL2Snapshot(snapshot *orderbook.OrderBookSnapshot) {
	fmt.Println("=== HYPERLIQUID L2 SNAPSHOT ===")
	if snapshot == nil {
		fmt.Println("valid=false")
		fmt.Println()
		return
	}

	valid := orderbook.ValidateSnapshot(*snapshot) == nil
	fmt.Printf("symbol=%s\n", snapshot.Symbol)
	fmt.Printf("bestBid=%s\n", orderbook.BestBid(*snapshot).Price)
	fmt.Printf("bestAsk=%s\n", orderbook.BestAsk(*snapshot).Price)
	fmt.Printf("spread=%.8f\n", orderbook.Spread(*snapshot))
	fmt.Printf("spreadPct=%.8f\n", orderbook.SpreadPct(*snapshot))
	fmt.Printf("mid=%.8f\n", orderbook.Mid(*snapshot))
	fmt.Printf("bidDepth1Pct=%.8f\n", orderbook.BidDepthWithinPct(*snapshot, 1))
	fmt.Printf("askDepth1Pct=%.8f\n", orderbook.AskDepthWithinPct(*snapshot, 1))
	fmt.Printf("imbalance1Pct=%.8f\n", orderbook.Imbalance(*snapshot, 1))
	fmt.Printf("valid=%t\n", valid)
	fmt.Println()
}

func printMultiVenueL2Summary(snapshots []orderbook.OrderBookSnapshot) {
	fmt.Println("=== MULTI VENUE L2 SUMMARY ===")
	if len(snapshots) == 0 {
		fmt.Println("snapshots=0")
		fmt.Println()
		return
	}
	for _, snapshot := range snapshots {
		fmt.Printf("venue=%s\n", snapshot.Venue)
		fmt.Printf("spreadPct=%.8f\n", orderbook.SpreadPct(snapshot))
		fmt.Printf("imbalance=%.8f\n", orderbook.Imbalance(snapshot, 1))
		fmt.Printf("valid=%t\n", orderbook.ValidateSnapshot(snapshot) == nil)
		fmt.Println()
	}
}

func candlesByVenue(venues []venueCandles) map[string][]exchanges.Candle {
	out := make(map[string][]exchanges.Candle, len(venues))
	for _, venue := range venues {
		out[venue.Venue] = venue.Candles
	}
	return out
}

func buildOrderBookResearchRows(snapshots []orderbook.OrderBookSnapshot, venueCandles map[string][]exchanges.Candle) orderBookResearchRows {
	rows := orderBookResearchRows{}
	for _, snapshot := range snapshots {
		rows.Snapshots = append(rows.Snapshots, snapshot)
		rows.Features = append(rows.Features, research.BuildOrderBookFeatureRow(snapshot))
		rows.VWAPInteraction = append(rows.VWAPInteraction, research.BuildVWAPL2InteractionRow(snapshot, venueCandles[snapshot.Venue]))
	}
	if err := research.WriteOrderBookFeaturesCSV("research/orderbook_features.csv", rows.Features); err != nil {
		log.Fatalf("write orderbook features: %v", err)
	}
	if err := research.WriteVWAPL2InteractionCSV("research/vwap_l2_interaction.csv", rows.VWAPInteraction); err != nil {
		log.Fatalf("write vwap l2 interaction: %v", err)
	}
	if err := research.WriteOrderBookFeaturesCSV("research/multi_venue_orderbook_features.csv", rows.Features); err != nil {
		log.Fatalf("write multi venue orderbook features: %v", err)
	}
	if err := research.WriteVWAPL2InteractionCSV("research/multi_venue_vwap_l2_interaction.csv", rows.VWAPInteraction); err != nil {
		log.Fatalf("write multi venue vwap l2 interaction: %v", err)
	}
	return rows
}

func buildVWAPL2RefreshStudy(venues []venueCandles, snapshots []orderbook.OrderBookSnapshot) vwapL2RefreshStudy {
	candles := candlesByVenue(venues)
	inputs := make([]research.VWAPL2RefreshInput, 0, len(snapshots))
	for _, snapshot := range snapshots {
		inputs = append(inputs, research.VWAPL2RefreshInput{
			Venue:    snapshot.Venue,
			Symbol:   snapshot.Symbol,
			Candles:  candles[snapshot.Venue],
			Snapshot: snapshot,
		})
	}

	result := research.BuildVWAPL2Refresh(inputs)
	if err := research.WriteVWAPL2FeaturesCSV("research/vwap_features_l2.csv", result.FeatureRows); err != nil {
		log.Fatalf("write vwap l2 features: %v", err)
	}
	if err := research.WriteContextL2FeaturesCSV("research/context_features_l2.csv", result.ContextRows); err != nil {
		log.Fatalf("write context l2 features: %v", err)
	}
	if err := research.WriteVWAPL2BehaviorSummaryJSON("research/vwap_behavior_l2_summary.json", result.Summary); err != nil {
		log.Fatalf("write vwap l2 behavior summary: %v", err)
	}

	return vwapL2RefreshStudy{
		FeatureRows: result.FeatureRows,
		ContextRows: result.ContextRows,
		Summary:     result.Summary,
	}
}

func buildL2SnapshotAnalysis() (l2recorder.SnapshotAnalysis, bool) {
	analysis, err := l2recorder.AnalyzeSnapshotCSV(config.DataPath("l2", "l2_snapshots.csv"))
	if err != nil {
		log.Printf("skip l2 snapshot analysis: %v", err)
		return l2recorder.SnapshotAnalysis{}, false
	}
	if err := research.WriteL2SnapshotAnalysisJSON("research/l2_snapshot_analysis.json", analysis); err != nil {
		log.Fatalf("write l2 snapshot analysis json: %v", err)
	}
	if err := research.WriteL2SnapshotAnalysisMarkdown("research/l2_snapshot_analysis.md", analysis); err != nil {
		log.Fatalf("write l2 snapshot analysis markdown: %v", err)
	}
	return analysis, true
}

func runChapter2(primary venueCandles, frame series.Frame) []strategyBacktestResult {
	values := buildChapter2Indicators(frame)
	if err := research.WriteChapter2IndicatorsCSV("research/chapter2_indicators.csv", values); err != nil {
		log.Fatalf("write chapter 2 indicators: %v", err)
	}

	results := runStrategies(primary.Candles, buildChapter2Strategies(frame), primary.Venue, primary.Symbol, ResearchInterval)
	rows := make([]research.StrategyComparisonRow, 0, len(results))
	analyses := make([]research.Chapter2Analysis, 0, len(results))
	for _, result := range results {
		rows = append(rows, research.StrategyComparisonRow{Strategy: result.Name, Report: result.Report})
		analyses = append(analyses, research.AnalyzeChapter2Strategy(result.Name, result.Report, result.Trades, result.Signals))
	}

	if err := research.WriteChapter2StrategyComparisonCSV("research/chapter2_strategy_comparison.csv", rows); err != nil {
		log.Fatalf("write chapter 2 strategy comparison: %v", err)
	}
	if err := research.WriteChapter2AnalysisCSV("research/chapter2_analysis.csv", analyses); err != nil {
		log.Fatalf("write chapter 2 analysis: %v", err)
	}
	if err := research.WriteChapter2AnalysisJSON("research/chapter2_analysis.json", analyses); err != nil {
		log.Fatalf("write chapter 2 analysis json: %v", err)
	}
	return results
}

func runChapter3(primary venueCandles, frame series.Frame) ([]features.FeatureRow, []labels.LabelRow, []datasets.TrainingRow) {
	featureRows := features.BuildFeatures(primary.Venue, primary.Symbol, ResearchInterval, frame)
	close := frame.CloseColumn()
	times := frame.TimeColumn()
	futureReturns := labels.FutureReturn(close, 10)
	directions := labels.DirectionLabel(futureReturns, 0.002)
	tpsl := labels.TPSLLabel(close, labels.TPSLConfig{Horizon: 10, TakeProfitPct: 0.005, StopLossPct: 0.003})
	labelRows := labels.BuildRows(times, futureReturns, directions, tpsl)
	trainingRows := datasets.BuildTrainingDataset(featureRows, labelRows)

	if err := features.WriteCSV("research/features.csv", featureRows); err != nil {
		log.Fatalf("write features: %v", err)
	}
	if err := labels.WriteCSV("research/labels.csv", labelRows); err != nil {
		log.Fatalf("write labels: %v", err)
	}
	if err := datasets.WriteCSV("research/training_dataset.csv", trainingRows); err != nil {
		log.Fatalf("write training dataset: %v", err)
	}
	if len(featureRows) != len(labelRows) || len(featureRows) != len(trainingRows) {
		log.Printf("dataset row mismatch features=%d labels=%d training=%d", len(featureRows), len(labelRows), len(trainingRows))
	}
	return featureRows, labelRows, trainingRows
}

func runChapter4(venues []venueCandles, primary venueCandles, frame series.Frame) ([]strategyBacktestResult, []research.StrategyComparisonRow, []research.Chapter4Analysis, []pairs.PairCandidate) {
	singleResults := runStrategies(primary.Candles, buildChapter4Strategies(frame), primary.Venue, primary.Symbol, ResearchInterval)
	singleRows := make([]research.StrategyComparisonRow, 0, len(singleResults))
	singleAnalyses := make([]research.Chapter4Analysis, 0, len(singleResults))
	for _, result := range singleResults {
		singleRows = append(singleRows, research.StrategyComparisonRow{Strategy: result.Name, Report: result.Report})
		singleAnalyses = append(singleAnalyses, research.AnalyzeChapter4Strategy(result.Name, result.Report, result.Trades, result.Signals))
	}
	if err := research.WriteChapter4StrategyComparisonCSV("research/chapter4_strategy_comparison.csv", singleRows); err != nil {
		log.Fatalf("write chapter 4 strategy comparison: %v", err)
	}
	if err := research.WriteChapter4AnalysisCSV("research/chapter4_analysis.csv", singleAnalyses); err != nil {
		log.Fatalf("write chapter 4 analysis: %v", err)
	}

	var venueRows []research.StrategyComparisonRow
	var venueAnalyses []research.Chapter4Analysis
	for _, venue := range venues {
		venueFrame := series.FromCandles(venue.Candles)
		for _, result := range runStrategies(venue.Candles, buildChapter4Strategies(venueFrame), venue.Venue, venue.Symbol, ResearchInterval) {
			venueRows = append(venueRows, research.StrategyComparisonRow{Strategy: result.Name, Report: result.Report})
			venueAnalyses = append(venueAnalyses, research.AnalyzeChapter4Strategy(result.Name, result.Report, result.Trades, result.Signals))
		}
	}
	if err := research.WriteChapter4VenueComparisonCSV("research/chapter4_venue_comparison.csv", venueRows); err != nil {
		log.Fatalf("write chapter 4 venue comparison: %v", err)
	}
	if err := research.WriteChapter4VenueAnalysisCSV("research/chapter4_venue_analysis.csv", venueAnalyses); err != nil {
		log.Fatalf("write chapter 4 venue analysis: %v", err)
	}

	pairCandidates := buildPairCandidates(venues)
	if err := pairs.WriteCandidatesCSV("research/chapter4_pairs_candidates.csv", pairCandidates); err != nil {
		log.Fatalf("write chapter 4 pairs candidates: %v", err)
	}
	return singleResults, venueRows, venueAnalyses, pairCandidates
}

func runChapter5(venues []venueCandles, primary venueCandles, frame series.Frame) ([]strategyBacktestResult, []statarb.Candidate) {
	volRows := buildVolatilityRows(frame)
	if err := research.WriteChapter5VolatilityCSV("research/chapter5_volatility.csv", volRows); err != nil {
		log.Fatalf("write chapter 5 volatility: %v", err)
	}

	strategies := buildChapter5Strategies(frame)
	results := runStrategies(primary.Candles, strategies, primary.Venue, primary.Symbol, ResearchInterval)
	rows := make([]research.Chapter5Analysis, 0, len(results))
	for i, result := range results {
		low, normal, high := chapter5.CountRegimes(strategies[i].Regimes)
		rows = append(rows, buildChapter5Analysis(result, low, normal, high))
	}
	if err := research.WriteChapter5StrategyComparisonCSV("research/chapter5_strategy_comparison.csv", rows); err != nil {
		log.Fatalf("write chapter 5 strategy comparison: %v", err)
	}
	if err := research.WriteChapter5AnalysisCSV("research/chapter5_analysis.csv", rows); err != nil {
		log.Fatalf("write chapter 5 analysis: %v", err)
	}

	statarbCandidates := buildStatArbCandidates(venues)
	if err := statarb.WriteCandidatesCSV("research/chapter5_statarb_candidates.csv", statarbCandidates); err != nil {
		log.Fatalf("write chapter 5 statarb candidates: %v", err)
	}
	return results, statarbCandidates
}

func buildChapter2Indicators(frame series.Frame) research.Chapter2Indicators {
	cfg := chapter2.DefaultConfig()
	times := frame.TimeColumn()
	close := frame.CloseColumn()
	high := frame.HighColumn()
	low := frame.LowColumn()
	macd := indicators.MACD(close, cfg.FastPeriod, cfg.SlowPeriod, cfg.SignalPeriod)
	bands := indicators.BollingerBands(close, cfg.Period, cfg.StdDevFactor)
	sr := indicators.RollingSupportResistance(high, low, cfg.Period)
	return research.Chapter2Indicators{
		Time:       times,
		Close:      close,
		SMA:        indicators.SMA(close, cfg.Period),
		EMA:        indicators.EMA(close, cfg.Period),
		APO:        indicators.APO(close, cfg.FastPeriod, cfg.SlowPeriod),
		MACD:       macd.MACD,
		MACDSignal: macd.Signal,
		MACDHist:   macd.Histogram,
		BBMiddle:   bands.Middle,
		BBUpper:    bands.Upper,
		BBLower:    bands.Lower,
		RSI:        indicators.RSI(close, 14),
		StdDev:     indicators.StdDev(close, cfg.Period),
		Momentum:   indicators.Momentum(close, 10),
		Support:    sr.Support,
		Resistance: sr.Resistance,
		Hour:       indicators.HourOfDay(times),
		DayOfWeek:  indicators.DayOfWeek(times),
	}
}

func buildChapter2Strategies(frame series.Frame) []strategyRun {
	cfg := chapter2.DefaultConfig()
	return []strategyRun{
		{Name: "sma", Signals: chapter2.SMAStrategy(frame, cfg.Period)},
		{Name: "ema", Signals: chapter2.EMAStrategy(frame, cfg.Period)},
		{Name: "apo", Signals: chapter2.APOStrategy(frame, cfg.FastPeriod, cfg.SlowPeriod)},
		{Name: "macd", Signals: chapter2.MACDStrategy(frame, cfg.FastPeriod, cfg.SlowPeriod, cfg.SignalPeriod)},
		{Name: "bollinger", Signals: chapter2.BollingerStrategy(frame, cfg.Period, cfg.StdDevFactor)},
		{Name: "rsi", Signals: chapter2.RSIStrategy(frame, 14, cfg.RSILow, cfg.RSIHigh)},
		{Name: "momentum", Signals: chapter2.MomentumStrategy(frame, 10)},
		{Name: "support_resistance", Signals: chapter2.SupportResistanceStrategy(frame, cfg.Period)},
	}
}

func buildChapter4Strategies(frame series.Frame) []strategyRun {
	return []strategyRun{
		{Name: "momentum", Signals: chapter4.MomentumStrategy(frame, 20)},
		{Name: "dual_ma", Signals: chapter4.DualMAStrategy(frame, 20, 50)},
		{Name: "turtle", Signals: chapter4.TurtleBreakoutStrategy(frame, 20, 10)},
		{Name: "mean_reversion", Signals: chapter4.MeanReversionStrategy(frame, 20, 2, 14, 30, 50)},
	}
}

func buildChapter5Strategies(frame series.Frame) []strategyRun {
	meanCfg := chapter5.DefaultVolatilityMeanReversionConfig()
	trendCfg := chapter5.DefaultVolatilityTrendConfig()
	return []strategyRun{
		{Name: "vol_mean_reversion", Signals: chapter5.VolatilityMeanReversion(frame, meanCfg), Regimes: chapter5.RegimesForMeanReversion(frame)},
		{Name: "vol_trend_following", Signals: chapter5.VolatilityTrendFollowing(frame, trendCfg), Regimes: chapter5.RegimesForTrendFollowing(frame)},
	}
}

func runStrategies(candles []exchanges.Candle, strategies []strategyRun, venue string, symbol string, interval string) []strategyBacktestResult {
	results := make([]strategyBacktestResult, 0, len(strategies))
	for _, run := range strategies {
		engine := backtest.NewEngine(backtest.Config{
			Venue:           venue,
			Symbol:          symbol,
			Interval:        interval,
			StartingBalance: 10000,
			FixedNotional:   100,
		}, risk.NewEngine(researchLimits()))
		report := engine.RunWithSignals(candles, run.Signals)
		results = append(results, strategyBacktestResult{
			Name:    run.Name,
			Signals: run.Signals,
			Report:  report,
			Trades:  append([]strategy.Trade(nil), engine.ClosedTrades...),
		})
	}
	return results
}

func runChapter6RiskEnforcement(primary venueCandles, results []strategyBacktestResult) []backtest.RiskEnforcedResult {
	controlConfig := risk.RiskControlConfig{
		MaxTradesPerDay:           4,
		MaxTradeSize:              1,
		MaxNotional:               100,
		MaxHoldBars:               48,
		StopLossPct:               0.01,
		MaxVolumeParticipationPct: 1,
		EnableViolationLogging:    true,
	}
	config := backtest.Config{
		Venue:           primary.Venue,
		Symbol:          primary.Symbol,
		Interval:        ResearchInterval,
		StartingBalance: 10000,
		FixedNotional:   100,
	}

	out := make([]backtest.RiskEnforcedResult, 0, len(results))
	for _, result := range results {
		out = append(out, backtest.RunRiskEnforcedWithSignals(primary.Candles, result.Signals, config, controlConfig, result.Name))
	}
	return out
}

func removedCandidateSetFromPacket(packet research.Chapter6RiskPacket) map[string]bool {
	out := make(map[string]bool, len(packet.RemovedFromCandidates))
	for _, row := range packet.RemovedFromCandidates {
		out[row.Strategy] = true
	}
	return out
}

func researchLimits() risk.Limits {
	limits := risk.DefaultLimits()
	limits.LiveTradingEnabled = true
	limits.MaxOpenPositions = 5
	limits.MaxTotalExposureUSD = 500
	limits.MaxSymbolExposureUSD = 500
	limits.MaxLeverage = 5
	limits.MinAvailableUSD = 100
	return limits
}

func buildVolatilityRows(frame series.Frame) []research.Chapter5VolatilityRow {
	times := frame.TimeColumn()
	close := frame.CloseColumn()
	high := frame.HighColumn()
	low := frame.LowColumn()
	trueRange := volatility.TrueRange(high, low, close)
	atr := volatility.ATR(high, low, close, 14)
	logReturns := volatility.LogReturns(close)
	realizedVol := volatility.RealizedVolatility(close, 20)
	rollingStdDev := volatility.RollingStdDev(close, 20)
	regime := volatility.DefaultRegime(realizedVol)
	rows := make([]research.Chapter5VolatilityRow, 0, len(frame.Rows))
	for i := range frame.Rows {
		rows = append(rows, research.Chapter5VolatilityRow{
			Time:          times[i],
			Close:         close[i],
			High:          high[i],
			Low:           low[i],
			TrueRange:     trueRange[i],
			ATR:           atr[i],
			LogReturn:     logReturns[i],
			RealizedVol:   realizedVol[i],
			RollingStdDev: rollingStdDev[i],
			Regime:        regime[i],
		})
	}
	return rows
}

func buildChapter5Analysis(result strategyBacktestResult, low int, normal int, high int) research.Chapter5Analysis {
	report := result.Report
	return research.Chapter5Analysis{
		Strategy:          result.Name,
		Venue:             report.Venue,
		Symbol:            report.Symbol,
		Interval:          report.Interval,
		Candles:           report.CandleCount,
		Trades:            report.Metrics.TotalTrades,
		Wins:              report.Metrics.Wins,
		Losses:            report.Metrics.Losses,
		WinRate:           report.Metrics.WinRate,
		GrossPnL:          report.Metrics.GrossPnL,
		Fees:              report.Metrics.Fees,
		NetPnL:            report.Metrics.NetPnL,
		EndingBalance:     report.EndingBalance,
		MaxDrawdown:       report.Metrics.MaxDrawdown,
		MaxDrawdownPct:    report.Metrics.MaxDrawdownPct,
		AverageTrade:      averageTrade(result.Trades),
		AverageHold:       averageHold(result.Trades, result.Signals.Times),
		SignalCount:       signalCount(result.Signals),
		OvertradeScore:    overtradeScore(report),
		RegimeLowCount:    low,
		RegimeNormalCount: normal,
		RegimeHighCount:   high,
	}
}

func buildPairCandidates(venues []venueCandles) []pairs.PairCandidate {
	byVenue := venueMap(venues)
	inputs := [][2]string{{"aster", "hyperliquid"}, {"aster", "lighter"}, {"hyperliquid", "lighter"}}
	var out []pairs.PairCandidate
	for _, input := range inputs {
		a, okA := byVenue[input[0]]
		b, okB := byVenue[input[1]]
		if !okA || !okB {
			log.Printf("skip chapter 4 pair %s/%s: missing venue candles", input[0], input[1])
			continue
		}
		candidate, err := pairs.BuildPairCandidate(a.Venue, a.Symbol, a.Candles, b.Venue, b.Symbol, b.Candles, 20)
		if err != nil {
			log.Printf("skip chapter 4 pair %s/%s: %v", input[0], input[1], err)
			continue
		}
		out = append(out, candidate)
	}
	return out
}

func buildStatArbCandidates(venues []venueCandles) []statarb.Candidate {
	byVenue := venueMap(venues)
	inputs := [][2]string{{"aster", "hyperliquid"}, {"aster", "lighter"}, {"hyperliquid", "lighter"}}
	var out []statarb.Candidate
	for _, input := range inputs {
		a, okA := byVenue[input[0]]
		b, okB := byVenue[input[1]]
		if !okA || !okB {
			log.Printf("skip chapter 5 statarb %s/%s: missing venue candles", input[0], input[1])
			continue
		}
		candidate, err := statarb.BuildCandidate(a.Venue, a.Symbol, a.Candles, b.Venue, b.Symbol, b.Candles, 20)
		if err != nil {
			log.Printf("skip chapter 5 statarb %s/%s: %v", input[0], input[1], err)
			continue
		}
		out = append(out, candidate)
	}
	return out
}

func venueMap(venues []venueCandles) map[string]venueCandles {
	out := make(map[string]venueCandles, len(venues))
	for _, venue := range venues {
		out[venue.Venue] = venue
	}
	return out
}

func findVenue(venues []venueCandles, name string) (venueCandles, bool) {
	for _, venue := range venues {
		if venue.Venue == name {
			return venue, true
		}
	}
	return venueCandles{}, false
}

func buildSummary(candles int, chapter2Results []strategyBacktestResult, chapter4Results []strategyBacktestResult, chapter5Results []strategyBacktestResult, featureRows int, labelRows int, trainingRows int, pairCandidates int) research.ResearchSummary {
	all := append(append([]strategyBacktestResult{}, chapter2Results...), chapter4Results...)
	all = append(all, chapter5Results...)
	worst := worstDrawdown(all)
	active := mostActive(all)
	best2 := bestNetPnL(chapter2Results)
	best4 := bestNetPnL(chapter4Results)
	best5 := bestNetPnL(chapter5Results)
	return research.ResearchSummary{
		Candles:               candles,
		BestChapter2Strategy:  best2.Name,
		BestChapter2NetPnL:    best2.Report.Metrics.NetPnL,
		BestChapter4Strategy:  best4.Name,
		BestChapter4NetPnL:    best4.Report.Metrics.NetPnL,
		BestChapter5Strategy:  best5.Name,
		BestChapter5NetPnL:    best5.Report.Metrics.NetPnL,
		WorstDrawdownStrategy: worst.Name,
		WorstDrawdown:         worst.Report.Metrics.MaxDrawdown,
		MostActiveStrategy:    active.Name,
		MostTrades:            active.Report.Metrics.TotalTrades,
		FeatureRows:           featureRows,
		LabelRows:             labelRows,
		TrainingRows:          trainingRows,
		PairCandidates:        pairCandidates,
	}
}

func bestNetPnL(results []strategyBacktestResult) strategyBacktestResult {
	if len(results) == 0 {
		return strategyBacktestResult{}
	}
	best := results[0]
	for _, result := range results[1:] {
		if result.Report.Metrics.NetPnL > best.Report.Metrics.NetPnL {
			best = result
		}
	}
	return best
}

func worstDrawdown(results []strategyBacktestResult) strategyBacktestResult {
	if len(results) == 0 {
		return strategyBacktestResult{}
	}
	worst := results[0]
	for _, result := range results[1:] {
		if result.Report.Metrics.MaxDrawdown > worst.Report.Metrics.MaxDrawdown {
			worst = result
		}
	}
	return worst
}

func mostActive(results []strategyBacktestResult) strategyBacktestResult {
	if len(results) == 0 {
		return strategyBacktestResult{}
	}
	most := results[0]
	for _, result := range results[1:] {
		if result.Report.Metrics.TotalTrades > most.Report.Metrics.TotalTrades {
			most = result
		}
	}
	return most
}

func averageTrade(trades []strategy.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}
	total := 0.0
	for _, trade := range trades {
		total += trade.RealizedPnL
	}
	return total / float64(len(trades))
}

func averageHold(trades []strategy.Trade, times []int64) float64 {
	if len(trades) == 0 {
		return 0
	}
	interval := candleInterval(times)
	if interval <= 0 {
		return 0
	}
	total := 0
	for _, trade := range trades {
		total += holdCandles(trade.OpenedAt, trade.ClosedAt, interval)
	}
	return float64(total) / float64(len(trades))
}

func signalCount(signals series.SignalResult) int {
	count := 0
	for _, position := range signals.Positions {
		if position != 0 {
			count++
		}
	}
	return count
}

func overtradeScore(report backtest.Report) float64 {
	if report.CandleCount == 0 {
		return 0
	}
	return float64(report.Metrics.TotalTrades) / float64(report.CandleCount)
}

func candleInterval(times []int64) time.Duration {
	for i := 1; i < len(times); i++ {
		if times[i] > times[i-1] {
			return time.Duration(times[i]-times[i-1]) * time.Millisecond
		}
	}
	return 0
}

func intervalToMillis(interval string) int64 {
	switch interval {
	case "1m":
		return int64(time.Minute / time.Millisecond)
	case "5m":
		return int64(5 * time.Minute / time.Millisecond)
	case "15m":
		return int64(15 * time.Minute / time.Millisecond)
	case "1h":
		return int64(time.Hour / time.Millisecond)
	case "1d":
		return int64(24 * time.Hour / time.Millisecond)
	default:
		return 0
	}
}

func holdCandles(openedAt time.Time, closedAt time.Time, interval time.Duration) int {
	if interval <= 0 || closedAt.Before(openedAt) {
		return 0
	}
	count := int(closedAt.Sub(openedAt) / interval)
	if count < 1 {
		return 1
	}
	return count
}

func topStrategies(results []strategyBacktestResult, n int) []strategyBacktestResult {
	out := append([]strategyBacktestResult(nil), results...)
	sort.Slice(out, func(i int, j int) bool {
		return out[i].Report.Metrics.NetPnL > out[j].Report.Metrics.NetPnL
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
