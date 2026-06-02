package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type VolumeProfileBookCompletionPacket struct {
	Status              string              `json:"status"`
	Setups              int                 `json:"setups"`
	Docs                int                 `json:"docs"`
	TradingStyleMap     []BookPacketSection `json:"tradingStyleMap"`
	InstrumentSelection []string            `json:"instrumentSelection"`
	MarketAnalysisAToZ  []string            `json:"marketAnalysisAToZ"`
	PositionManagement  []string            `json:"positionManagement"`
	MoneyManagement     []string            `json:"moneyManagement"`
	PsychologyNotes     []string            `json:"psychologyNotes"`
	BacktestingProgress []string            `json:"backtestingProgression"`
	RepoMapping         []BookPacketSection `json:"repoMapping"`
	Gaps                []string            `json:"gaps"`
	FinalRecommendation []string            `json:"finalRecommendation"`
	GeneratedDocs       []string            `json:"generatedDocs"`
	GeneratedResearch   []string            `json:"generatedResearch"`
}

type BookPacketSection struct {
	Name    string   `json:"name"`
	Details []string `json:"details"`
}

func BuildVolumeProfileBookCompletionPacket() VolumeProfileBookCompletionPacket {
	return VolumeProfileBookCompletionPacket{
		Status: "complete",
		Setups: 4,
		Docs:   3,
		TradingStyleMap: []BookPacketSection{
			{Name: "intraday", Details: []string{"daily_session profile", "rolling_3d profile", "VWAP session context", "daily open and failed-auction studies", "latest L2 context for spread and imbalance"}},
			{Name: "swing", Details: []string{"rolling_7d profile", "composite_30d profile", "accumulation and trend setup studies", "setup comparison and quality reviews"}},
			{Name: "long-term", Details: []string{"composite_30d profile", "profile shape study", "major POC/VAH/VAL context", "research-only, no investing strategy implemented"}},
		},
		InstrumentSelection: []string{
			"Liquidity: depth and liquidity near price/VWAP from orderbook and L2 exports.",
			"Spread: normalized spread percent from the orderbook package.",
			"Volatility: Chapter 5 volatility metrics and regime reports.",
			"Data availability: candles, selected trade feeds, L2 snapshots, and L2 recorder outputs.",
			"Venue reliability: Aster, Hyperliquid, and Lighter market-data availability.",
		},
		MarketAnalysisAToZ: []string{
			"Price action features identify aggression, initiation, rejection, failed auction, opens, highs, and lows.",
			"Volume Profile features identify POC, VAH, VAL, HVN, LVN, shape, scoped profiles, and flexible event profiles.",
			"VWAP features provide session/anchored VWAP, distance, slope, interaction, and context studies.",
			"L2 context provides spread, depth, imbalance, and liquidity-near-VWAP research.",
			"Risk context comes from Chapter 6 metrics, filters, ranking, and research-only enforcement.",
			"Setup comparison ranks accumulation, trend, rejection, and reversal studies side by side.",
		},
		PositionManagement: []string{
			"Profit targets can be researched around POC, VAH, VAL, HVN, LVN, VWAP, and prior highs/lows.",
			"Stops can be researched beyond rejection high/low, VAH/VAL, failed-auction level, or accumulation range.",
			"Volume-based target/stop notes are documented only; no executable rules are added.",
			"Early exits can be researched when acceptance fails, VWAP alignment breaks, or L2 quality deteriorates.",
		},
		MoneyManagement: []string{
			"Risk per trade must be defined before any future strategy execution work.",
			"R:R should be measured from profile-defined target and invalidation levels.",
			"Position sizing should account for volatility, spread, and liquidity.",
			"Correlation and excessive exposure should be considered before combining setups or venues.",
		},
		PsychologyNotes: []string{
			"Good winner: follows rules and planned target logic.",
			"Bad winner: profitable only because rules were ignored.",
			"Good loser: follows invalidation and preserves risk discipline.",
			"Bad loser: ignores stop, widens risk, or adds outside plan.",
			"Rule-breaking prevention: keep setup research separate from execution.",
		},
		BacktestingProgress: []string{
			"Rough backtest: inspect setup exports, quality reports, and passive follow-through.",
			"Thorough backtest: only after manual review, define explicit signal rules and use the existing backtester.",
			"Micro trading: not enabled.",
			"Half positions: not enabled.",
			"Full positions: not enabled.",
		},
		RepoMapping: []BookPacketSection{
			{Name: "priceaction/", Details: []string{"aggression", "sideways action", "initiation", "rejection", "opens", "high/low studies", "failed auctions"}},
			{Name: "volumeprofile/", Details: []string{"price bins", "POC", "VAH/VAL", "HVN/LVN", "profile shapes", "scoped/flexible profiles", "setup studies"}},
			{Name: "vwap/", Details: []string{"session VWAP", "anchored VWAP", "distance", "slope", "regime", "interaction", "context"}},
			{Name: "orderbook/", Details: []string{"normalized L2 schema", "spread", "mid", "depth", "imbalance", "liquidity near price"}},
			{Name: "l2recorder/", Details: []string{"snapshot recorder", "CSV storage", "dataset analysis"}},
			{Name: "riskmetrics/", Details: []string{"Sharpe", "Sortino", "variance", "expectancy", "execution stats", "risk grades"}},
			{Name: "backtest/", Details: []string{"for-loop backtester", "event-driven skeleton", "clock", "splits", "assumptions", "risk-enforced research path"}},
			{Name: "research/", Details: []string{"CSV/JSON/Markdown reports for all book-derived research layers"}},
		},
		Gaps: []string{
			"Tick-accurate Volume Profile.",
			"Historical L2 replay.",
			"Real trade tape volume-at-price.",
			"Manual chart review of strongest setup filters.",
			"Pacifica adapter later.",
			"ML later.",
		},
		FinalRecommendation: []string{
			"Volume Profile book research layer is complete.",
			"Do not add live or paper trading yet.",
			"Next best step is manual review of strongest setup filters.",
			"After manual review, choose either Pacifica adapter work or refined strategy research.",
		},
		GeneratedDocs: []string{
			"docs/volume_profile_playbook.md",
			"docs/volume_profile_risk_playbook.md",
			"docs/volume_profile_backtesting_playbook.md",
		},
		GeneratedResearch: []string{
			"research/volume_setup_comparison.csv",
			"research/volume_setup_comparison.json",
			"research/volume_setup_comparison.md",
			"research/volume_profile_book_completion_packet.json",
			"research/volume_profile_book_completion_packet.md",
		},
	}
}

func WriteVolumeProfileBookCompletionPacketJSON(path string, packet VolumeProfileBookCompletionPacket) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteVolumeProfileBookCompletionPacketMarkdown(path string, packet VolumeProfileBookCompletionPacket) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body := strings.Builder{}
	body.WriteString("# Volume Profile Book Completion Packet\n\n")
	body.WriteString(fmt.Sprintf("status=%s\n\nsetups=%d\n\ndocs=%d\n\n", packet.Status, packet.Setups, packet.Docs))
	writeBookPacketSections(&body, "Trading Style Map", packet.TradingStyleMap)
	writeBookPacketList(&body, "Instrument Selection", packet.InstrumentSelection)
	writeBookPacketList(&body, "Market Analysis A To Z", packet.MarketAnalysisAToZ)
	writeBookPacketList(&body, "Position Management", packet.PositionManagement)
	writeBookPacketList(&body, "Money Management", packet.MoneyManagement)
	writeBookPacketList(&body, "Psychology Notes", packet.PsychologyNotes)
	writeBookPacketList(&body, "Backtesting Progression", packet.BacktestingProgress)
	writeBookPacketSections(&body, "Repo Mapping", packet.RepoMapping)
	writeBookPacketList(&body, "Gaps", packet.Gaps)
	writeBookPacketList(&body, "Final Recommendation", packet.FinalRecommendation)
	writeBookPacketList(&body, "Generated Docs", packet.GeneratedDocs)
	writeBookPacketList(&body, "Generated Research", packet.GeneratedResearch)
	return os.WriteFile(path, []byte(body.String()), 0644)
}

func writeBookPacketList(body *strings.Builder, title string, values []string) {
	body.WriteString("## " + title + "\n\n")
	for _, value := range values {
		body.WriteString("- " + value + "\n")
	}
	body.WriteString("\n")
}

func writeBookPacketSections(body *strings.Builder, title string, sections []BookPacketSection) {
	body.WriteString("## " + title + "\n\n")
	for _, section := range sections {
		body.WriteString("### " + section.Name + "\n\n")
		for _, detail := range section.Details {
			body.WriteString("- " + detail + "\n")
		}
		body.WriteString("\n")
	}
}
