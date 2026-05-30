package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Chapter7Packet struct {
	ImplementedComponents []Chapter7ComponentMapping `json:"implementedComponents"`
	CriticalComponents    []string                   `json:"criticalComponents"`
	NonCriticalComponents []string                   `json:"nonCriticalComponents"`
	QueueDataFlow         []string                   `json:"queueDataFlow"`
	TradingLifecycle      []string                   `json:"tradingSimulationLifecycle"`
	OrderLifecycle        []string                   `json:"orderLifecycle"`
	ServiceLifecycle      []string                   `json:"serviceSupervisorLifecycle"`
	RemainingGaps         []string                   `json:"remainingChapter7Gaps"`
	ReadinessForChapter8  string                     `json:"readinessForChapter8"`
	Conclusion            []string                   `json:"conclusion"`
}

type Chapter7ComponentMapping struct {
	BookComponent string `json:"bookComponent"`
	RepoComponent string `json:"repoComponent"`
	Status        string `json:"status"`
}

func BuildChapter7Packet() Chapter7Packet {
	return Chapter7Packet{
		ImplementedComponents: []Chapter7ComponentMapping{
			{BookComponent: "LiquidityProvider", RepoComponent: "system.LiquidityProvider", Status: "implemented"},
			{BookComponent: "OrderBook", RepoComponent: "system.OrderBook", Status: "implemented"},
			{BookComponent: "TradingStrategy", RepoComponent: "system.TradingStrategy", Status: "implemented"},
			{BookComponent: "OrderManager", RepoComponent: "system.OrderManager", Status: "implemented"},
			{BookComponent: "MarketSimulator", RepoComponent: "system.MarketSimulator", Status: "implemented"},
			{BookComponent: "TestTradingSimulation", RepoComponent: "system.TradingSimulation", Status: "implemented"},
			{BookComponent: "Command and control", RepoComponent: "system.CommandControl", Status: "implemented"},
			{BookComponent: "Services", RepoComponent: "system.Service/SystemSupervisor", Status: "implemented"},
			{BookComponent: "Risk service", RepoComponent: "system.RiskService", Status: "placeholder only"},
		},
		CriticalComponents: []string{
			"LiquidityProvider",
			"OrderBook",
			"TradingStrategy",
			"OrderManager",
			"MarketSimulator",
			"TestTradingSimulation",
		},
		NonCriticalComponents: []string{
			"CommandControl",
			"SystemSupervisor",
			"LoggingService",
			"PositionService",
			"MarketDataService",
			"OrderService",
			"RiskService placeholder",
		},
		QueueDataFlow: []string{
			"lp_2_gateway: LiquidityProvider -> OrderBook",
			"ob_2_ts: OrderBook -> TradingStrategy",
			"ts_2_om: TradingStrategy -> OrderManager",
			"om_2_gw: OrderManager -> MarketSimulator",
			"gw_2_om: MarketSimulator -> OrderManager",
			"om_2_ts: OrderManager -> TradingStrategy",
			"ms_2_om: MarketSimulator -> OrderManager audit stream",
		},
		TradingLifecycle: []string{
			"LiquidityProvider inserts simulated bid/ask liquidity.",
			"OrderBook maintains sorted bid/ask books and emits top-of-book changes.",
			"TradingStrategy detects crossed-book arbitrage and creates buy/sell simulated intents.",
			"OrderManager validates and assigns internal order IDs.",
			"MarketSimulator accepts valid orders first.",
			"MarketSimulator FillAllOrders emits fills.",
			"OrderManager forwards fills back to TradingStrategy.",
			"TradingStrategy updates position, cash, and realized PnL.",
		},
		OrderLifecycle: []string{
			"NEW",
			"ACCEPTED",
			"FILLED",
			"CANCELED",
			"AMENDED",
			"REJECTED",
		},
		ServiceLifecycle: []string{
			"Services start in CREATED state.",
			"SystemSupervisor starts all registered services.",
			"CommandControl handles START, STOP, STATUS, PAUSE, and RESUME.",
			"CommandControl records every command in an audit log.",
			"SystemSupervisor reports service status summaries.",
		},
		RemainingGaps: []string{
			"Exchange gateway connectivity belongs to Chapter 8 and is intentionally outside the system path.",
			"RiskService is lifecycle-only and does not enforce new risk logic.",
			"OrderBook is simulated and insert-focused; richer exchange-style deltas can come later.",
			"Command/control has no network UI or operator console yet.",
			"Services are lifecycle shells, not long-running goroutines.",
		},
		ReadinessForChapter8: "Ready for Chapter 8 exchange connectivity audit, while keeping exchange adapters outside this system path until explicitly mapped.",
		Conclusion: []string{
			"Chapter 7 trading-system skeleton is complete.",
			"No live/paper trading is enabled.",
			"Exchange adapters remain outside this system path until Chapter 8.",
			"RiskService is placeholder only.",
			"The system is ready to proceed to Chapter 8 exchange connectivity audit.",
		},
	}
}

func WriteChapter7PacketJSON(path string, packet Chapter7Packet) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter7PacketMarkdown(path string, packet Chapter7Packet) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter7PacketMarkdown(packet)), 0644)
}

func Chapter7PacketMarkdown(packet Chapter7Packet) string {
	var b strings.Builder
	b.WriteString("# Chapter 7 Final Trading System Packet\n\n")

	b.WriteString("## Components Implemented\n\n")
	b.WriteString("| Book Component | Repo Component | Status |\n")
	b.WriteString("|---|---|---|\n")
	for _, row := range packet.ImplementedComponents {
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", row.BookComponent, row.RepoComponent, row.Status))
	}

	writeStringList(&b, "Critical Components", packet.CriticalComponents)
	writeStringList(&b, "Non-Critical Components", packet.NonCriticalComponents)
	writeStringList(&b, "Queue/Channel Data Flow", packet.QueueDataFlow)
	writeStringList(&b, "Trading Simulation Lifecycle", packet.TradingLifecycle)
	writeStringList(&b, "Order Lifecycle", packet.OrderLifecycle)
	writeStringList(&b, "Service/Supervisor Lifecycle", packet.ServiceLifecycle)
	writeStringList(&b, "Remaining Chapter 7 Gaps", packet.RemainingGaps)

	b.WriteString("\n## Readiness For Chapter 8\n\n")
	b.WriteString(packet.ReadinessForChapter8 + "\n")

	writeStringList(&b, "Conclusion", packet.Conclusion)
	return b.String()
}

func writeStringList(b *strings.Builder, title string, rows []string) {
	b.WriteString("\n## " + title + "\n\n")
	if len(rows) == 0 {
		b.WriteString("_None._\n")
		return
	}
	for _, row := range rows {
		b.WriteString("- " + row + "\n")
	}
}
