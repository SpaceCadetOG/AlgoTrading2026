package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Chapter8GatewayConcept struct {
	Concept   string   `json:"concept"`
	BookRole  string   `json:"bookRole"`
	RepoFiles []string `json:"repoFiles"`
	Status    string   `json:"status"`
	Gap       string   `json:"gap"`
}

type Chapter8VenueGatewayAudit struct {
	Venue                  string `json:"venue"`
	CommunicationAPI       string `json:"communicationAPI"`
	PriceUpdates           string `json:"priceUpdates"`
	OrderSending           string `json:"orderSending"`
	MarketResponseHandling string `json:"marketResponseHandling"`
	OtherTradingAPIs       string `json:"otherTradingAPIs"`
	SystemGatewayStatus    string `json:"systemGatewayStatus"`
}

type Chapter8GatewayAudit struct {
	Title      string                      `json:"title"`
	Concepts   []Chapter8GatewayConcept    `json:"concepts"`
	Venues     []Chapter8VenueGatewayAudit `json:"venues"`
	FilesAdded []string                    `json:"filesAdded"`
	Gaps       []string                    `json:"gaps"`
	Conclusion string                      `json:"conclusion"`
	NextPhase  string                      `json:"nextPhase"`
}

func BuildChapter8GatewayAudit() Chapter8GatewayAudit {
	return Chapter8GatewayAudit{
		Title: "Chapter 8B Gateway Abstraction and Adapter Audit",
		Concepts: []Chapter8GatewayConcept{
			{
				Concept:  "communication API",
				BookRole: "Document the messages and request/response shape used to communicate with a market.",
				RepoFiles: []string{
					"exchanges/aster",
					"exchanges/hyperliquid",
					"exchanges/lighter",
					"system/gateway.go",
					"system/gateway_adapter_types.go",
				},
				Status: "partial",
				Gap:    "Real venue adapters are not wrapped into the system gateway path yet.",
			},
			{
				Concept:  "receiving price updates",
				BookRole: "Receive external market data and convert it into internal order-book events.",
				RepoFiles: []string{
					"marketdata",
					"system/order_book.go",
					"system/simulated_gateway.go",
					"exchanges/aster/candles_ws.go",
					"exchanges/hyperliquid/candles_ws.go",
					"exchanges/lighter/candles_ws.go",
				},
				Status: "partial",
				Gap:    "Simulated gateway conversion exists; live venue WS streams remain outside the Chapter 7 gateway path.",
			},
			{
				Concept:  "sending orders",
				BookRole: "Accept internal order intents and forward market-compatible orders.",
				RepoFiles: []string{
					"execution",
					"system/order_manager.go",
					"system/gateway.go",
					"system/simulated_gateway.go",
					"exchanges/aster/orders.go",
					"exchanges/hyperliquid/orders.go",
					"exchanges/lighter/orders.go",
				},
				Status: "partial",
				Gap:    "Simulated order gateway is implemented; real gateway adapter wrappers are intentionally deferred.",
			},
			{
				Concept:  "receiving market responses",
				BookRole: "Convert accepts, fills, rejects, cancels, and amends into internal OMS responses.",
				RepoFiles: []string{
					"system/market_simulator.go",
					"system/order_manager.go",
					"system/simulated_gateway.go",
					"exchanges/aster/user_stream.go",
					"exchanges/hyperliquid/user_stream.go",
					"exchanges/lighter/orders.go",
				},
				Status: "partial",
				Gap:    "System response type exists; venue-specific response translators remain future Chapter 8 work.",
			},
			{
				Concept:  "other trading APIs",
				BookRole: "Account for non-FIX exchange APIs such as REST and WebSocket.",
				RepoFiles: []string{
					"exchanges/aster",
					"exchanges/hyperliquid",
					"exchanges/lighter",
				},
				Status: "ahead",
				Gap:    "Crypto REST/WS/sendTx adapters exist before the formal Chapter 8 adapter wrapping step.",
			},
		},
		Venues: []Chapter8VenueGatewayAudit{
			{
				Venue:                  "aster",
				CommunicationAPI:       "REST plus WebSocket adapters under exchanges/aster",
				PriceUpdates:           "REST candles and WS candles/order-book style market streams",
				OrderSending:           "signed REST order methods in exchanges/aster/orders.go",
				MarketResponseHandling: "order query/cancel responses and user stream events",
				OtherTradingAPIs:       "Aster crypto REST/WS API with EIP-712 signing",
				SystemGatewayStatus:    "audited but not wrapped into system.Gateway",
			},
			{
				Venue:                  "hyperliquid",
				CommunicationAPI:       "REST info/exchange actions plus WebSocket adapters under exchanges/hyperliquid",
				PriceUpdates:           "REST candles and WS candle stream",
				OrderSending:           "signed action payload order methods in exchanges/hyperliquid/orders.go",
				MarketResponseHandling: "open order/account responses and user/order update stream",
				OtherTradingAPIs:       "Hyperliquid crypto REST/WS API",
				SystemGatewayStatus:    "audited but not wrapped into system.Gateway",
			},
			{
				Venue:                  "lighter",
				CommunicationAPI:       "REST plus WebSocket adapters under exchanges/lighter",
				PriceUpdates:           "REST candles and WS stream",
				OrderSending:           "SDK-built transaction submitted through /api/v1/sendTx",
				MarketResponseHandling: "sendTx business responses and account/order status endpoints",
				OtherTradingAPIs:       "Lighter REST/WS/sendTx API",
				SystemGatewayStatus:    "audited but not wrapped into system.Gateway",
			},
		},
		FilesAdded: []string{
			"system/gateway.go",
			"system/simulated_gateway.go",
			"system/gateway_adapter_types.go",
			"research/chapter8_gateway_audit.go",
		},
		Gaps: []string{
			"No FIX session/parser is implemented; crypto REST/WS APIs remain the active adapter style.",
			"Real exchange adapters are not connected to system.Gateway yet.",
			"Gateway heartbeat/reconnect/session recovery is not implemented.",
			"No live or paper gateway is enabled.",
		},
		Conclusion: "Chapter 8B adds the gateway abstraction and simulated gateway path while keeping real venue adapters outside the Chapter 7 system path.",
		NextPhase:  "Chapter 8C should wrap venue adapters behind disabled-by-default gateway translators, still without enabling live or paper trading.",
	}
}

func WriteChapter8GatewayAuditJSON(path string, audit Chapter8GatewayAudit) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func WriteChapter8GatewayAuditMarkdown(path string, audit Chapter8GatewayAudit) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", audit.Title)
	fmt.Fprintf(&b, "## Concept Mapping\n\n")
	fmt.Fprintf(&b, "| Concept | Status | Repo files | Gap |\n")
	fmt.Fprintf(&b, "|---|---|---|---|\n")
	for _, concept := range audit.Concepts {
		fmt.Fprintf(
			&b,
			"| %s | %s | %s | %s |\n",
			concept.Concept,
			concept.Status,
			strings.Join(concept.RepoFiles, "<br>"),
			concept.Gap,
		)
	}

	fmt.Fprintf(&b, "\n## Venue Adapter Audit\n\n")
	fmt.Fprintf(&b, "| Venue | Communication API | Price updates | Order sending | Market responses | Other APIs | Gateway status |\n")
	fmt.Fprintf(&b, "|---|---|---|---|---|---|---|\n")
	for _, venue := range audit.Venues {
		fmt.Fprintf(
			&b,
			"| %s | %s | %s | %s | %s | %s | %s |\n",
			venue.Venue,
			venue.CommunicationAPI,
			venue.PriceUpdates,
			venue.OrderSending,
			venue.MarketResponseHandling,
			venue.OtherTradingAPIs,
			venue.SystemGatewayStatus,
		)
	}

	fmt.Fprintf(&b, "\n## Remaining Gaps\n\n")
	for _, gap := range audit.Gaps {
		fmt.Fprintf(&b, "- %s\n", gap)
	}
	fmt.Fprintf(&b, "\n## Conclusion\n\n%s\n\n", audit.Conclusion)
	fmt.Fprintf(&b, "## Next Phase\n\n%s\n", audit.NextPhase)

	return os.WriteFile(path, []byte(b.String()), 0644)
}
