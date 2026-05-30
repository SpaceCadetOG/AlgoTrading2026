package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
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
	"AlgoTrading2026/labels"
	"AlgoTrading2026/pairs"
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
	"AlgoTrading2026/volatility"
)

const (
	ResearchDays      = 30
	ResearchInterval  = "15m"
	CandlesPerDay15m  = 96
	ResearchCandleCap = ResearchDays * CandlesPerDay15m
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

func asterEnv(name string) string {
	if config.IsTestnet() {
		return os.Getenv("ASTER_TESTNET_" + name)
	}
	return os.Getenv("ASTER_" + name)
}

func main() {
	_ = godotenv.Load()
	log.Println("TRADING_ENV:", config.TradingEnv())

	venues := fetchVenueCandles(ResearchInterval, ResearchCandleCap)
	for _, venue := range venues {
		log.Printf("venue=%s symbol=%s candles=%d", venue.Venue, venue.Symbol, len(venue.Candles))
	}

	primary, ok := findVenue(venues, "aster")
	if !ok {
		log.Fatal("aster candles unavailable; cannot run primary chapter research")
	}

	frame := series.FromCandles(primary.Candles)

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
	fmt.Println()
	fmt.Printf("pairCandidates=%d\n", summary.PairCandidates)
	fmt.Println()
	fmt.Println("wrote " + summaryPath)
	fmt.Println()
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
	fmt.Println("conclusion=No strategy is promoted to active/paper execution yet.")
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
