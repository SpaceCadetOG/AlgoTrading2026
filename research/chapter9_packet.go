package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Chapter9Packet struct {
	ImplementedConcepts      []Chapter9PacketConcept `json:"implementedConcepts"`
	ForLoopBacktesterSummary []string                `json:"forLoopBacktesterSummary"`
	EventDrivenSummary       []string                `json:"eventDrivenBacktesterSummary"`
	SplitStatus              string                  `json:"inSampleOutOfSampleStatus"`
	SimulatedClockStatus     string                  `json:"simulatedClockValueOfTimeStatus"`
	AssumptionGaps           []string                `json:"assumptionGaps"`
	BacktesterComparison     []string                `json:"forLoopVsEventDrivenComparison"`
	StrategyLabUsage         []string                `json:"strategyLabUsage"`
	Chapter10DeferredGaps    []string                `json:"remainingGapsDeferredToChapter10"`
	ReadinessForChapter10    string                  `json:"readinessForChapter10"`
	Conclusion               []string                `json:"conclusion"`
	SourceArtifacts          []string                `json:"sourceArtifacts"`
}

type Chapter9PacketConcept struct {
	BookConcept string `json:"bookConcept"`
	RepoMapping string `json:"repoMapping"`
	Status      string `json:"status"`
}

func BuildChapter9Packet(
	backtestAudit Chapter9BacktestAudit,
	timeReport TimeAssumptionReport,
	eventReport Chapter9EventDrivenReport,
	comparisonReport Chapter9BacktesterComparisonReport,
) Chapter9Packet {
	return Chapter9Packet{
		ImplementedConcepts: []Chapter9PacketConcept{
			{BookConcept: "in-sample vs out-of-sample", RepoMapping: "backtest.InSampleOutOfSampleSplit", Status: "implemented"},
			{BookConcept: "correct backtest assumptions", RepoMapping: "backtest.BacktestAssumptions", Status: "documented"},
			{BookConcept: "for-loop backtester", RepoMapping: "backtest.Engine, backtest.Replay, backtest.RunWithSignals", Status: "implemented"},
			{BookConcept: "event-driven backtester", RepoMapping: "backtest.EventDrivenBacktester with Chapter 7 system queues", Status: "implemented skeleton"},
			{BookConcept: "value of time", RepoMapping: "backtest.SimulatedClock and backtest.CandleTimeIterator", Status: "implemented deterministic clock"},
			{BookConcept: "dual moving average backtest", RepoMapping: "strategies/chapter4.DualMAStrategy", Status: "strategy available; event-driven dual MA comparison deferred"},
			{BookConcept: "paper trading / forward testing", RepoMapping: "not enabled", Status: "deferred"},
			{BookConcept: "data storage", RepoMapping: "research CSV/JSON outputs", Status: "partial"},
		},
		ForLoopBacktesterSummary: []string{
			"The for-loop backtester replays sorted candles directly.",
			"It consumes normalized candles and precomputed signal results.",
			"It simulates fills using candle close plus fee and slippage models.",
			"It exports trades, signals, equity curves, summaries, and research reports.",
			"It is the preferred path for fast signal and strategy research.",
		},
		EventDrivenSummary: []string{
			fmt.Sprintf("The event-driven backtester processed %d candles in the latest report.", eventReport.Result.CandlesProcessed),
			fmt.Sprintf("It created %d orders and filled %d simulated orders.", eventReport.Result.OrdersCreated, eventReport.Result.OrdersFilled),
			"It routes liquidity through LiquidityProvider, OrderBook, TradingStrategy, OrderManager, and MarketSimulator.",
			"It uses the deterministic simulated clock rather than wall-clock time.",
			"It is for OMS, system, and market-simulator validation.",
		},
		SplitStatus: fmt.Sprintf(
			"in-sample=%d out-of-sample=%d ratio=%.2f",
			backtestAudit.InSampleCount,
			backtestAudit.OutOfSampleCount,
			backtestAudit.InSampleRatio,
		),
		SimulatedClockStatus: timeReport.ReadinessForEventBacktest,
		AssumptionGaps:       append([]string(nil), backtestAudit.Assumptions.Gaps...),
		BacktesterComparison: []string{
			fmt.Sprintf("comparison candles=%d", comparisonReport.Result.Candles),
			fmt.Sprintf("for-loop trades=%d", comparisonReport.Result.ForLoopTrades),
			fmt.Sprintf("event-driven orders=%d fills=%d", comparisonReport.Result.EventDrivenOrders, comparisonReport.Result.EventDrivenFills),
			fmt.Sprintf("pnl difference=%.2f", comparisonReport.Result.PnLDifference),
			comparisonReport.Result.AssumptionsDifference,
			comparisonReport.Result.Recommendation,
		},
		StrategyLabUsage: []string{
			"Use for-loop backtests to iterate quickly on indicators, signals, and strategy rules.",
			"Use in-sample/out-of-sample splits before trusting any strategy comparison.",
			"Use event-driven backtests to validate system queue flow, OMS behavior, fills, and market assumptions.",
			"Use research exports to compare strategy candidates before any paper or live execution path is considered.",
		},
		Chapter10DeferredGaps: []string{
			"Backtester versus live-market dislocations.",
			"Latency variance and response-delay modeling.",
			"Market impact and place-in-line estimates.",
			"Historical data accuracy checks.",
			"Slippage/fees realism calibration.",
			"Live strategy analytics and profit decay analysis.",
		},
		ReadinessForChapter10: "Ready to proceed to Chapter 10 market adaptation and realism.",
		Conclusion: []string{
			"Chapter 9 backtesting layer is complete.",
			"For-loop backtester is for fast signal/strategy research.",
			"Event-driven backtester is for OMS/system/market-simulator validation.",
			"Results are not expected to match because assumptions differ.",
			"No live or paper trading is enabled.",
			"Ready to proceed to Chapter 10 market adaptation and realism.",
		},
		SourceArtifacts: []string{
			"research/chapter9_backtest_audit.json",
			"research/chapter9_backtest_audit.md",
			"research/chapter9_time_model.json",
			"research/chapter9_time_model.md",
			"research/chapter9_event_driven.json",
			"research/chapter9_event_driven.md",
			"research/chapter9_backtester_comparison.json",
			"research/chapter9_backtester_comparison.md",
		},
	}
}

func WriteChapter9PacketJSON(path string, packet Chapter9Packet) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter9PacketMarkdown(path string, packet Chapter9Packet) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter9PacketMarkdown(packet)), 0644)
}

func Chapter9PacketMarkdown(packet Chapter9Packet) string {
	var b strings.Builder
	b.WriteString("# Chapter 9 Final Backtester Packet\n\n")

	b.WriteString("## Concepts Implemented\n\n")
	b.WriteString("| Book Concept | Repo Mapping | Status |\n")
	b.WriteString("|---|---|---|\n")
	for _, concept := range packet.ImplementedConcepts {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", concept.BookConcept, concept.RepoMapping, concept.Status)
	}

	writeStringList(&b, "For-Loop Backtester Summary", packet.ForLoopBacktesterSummary)
	writeStringList(&b, "Event-Driven Backtester Summary", packet.EventDrivenSummary)

	b.WriteString("\n## In-Sample / Out-of-Sample Status\n\n")
	b.WriteString(packet.SplitStatus + "\n")

	b.WriteString("\n## Simulated Clock / Value-of-Time Status\n\n")
	b.WriteString(packet.SimulatedClockStatus + "\n")

	writeStringList(&b, "Assumption Gaps", packet.AssumptionGaps)
	writeStringList(&b, "For-Loop vs Event-Driven Comparison", packet.BacktesterComparison)
	writeStringList(&b, "How To Use The Backtester As Strategy Lab", packet.StrategyLabUsage)
	writeStringList(&b, "Remaining Gaps Deferred To Chapter 10", packet.Chapter10DeferredGaps)

	b.WriteString("\n## Readiness For Chapter 10\n\n")
	b.WriteString(packet.ReadinessForChapter10 + "\n")

	writeStringList(&b, "Conclusion", packet.Conclusion)
	return b.String()
}
