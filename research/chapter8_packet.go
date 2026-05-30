package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Chapter8Packet struct {
	ImplementedConcepts     []Chapter8PacketConcept `json:"implementedConcepts"`
	VenueAPIMapping         []Chapter8PacketVenue   `json:"venueApiMapping"`
	GatewayResponsibilities []string                `json:"gatewayResponsibilities"`
	PriceUpdateHandling     []string                `json:"priceUpdateHandling"`
	OrderRequestHandling    []string                `json:"orderRequestHandling"`
	MarketResponseHandling  []string                `json:"marketResponseHandling"`
	SessionLifecycle        []string                `json:"sessionLifecycleHeartbeatModel"`
	DisabledSafetyModel     []string                `json:"disabledByDefaultSafetyModel"`
	MissingFIXItems         []string                `json:"missingFixItems"`
	RemainingGaps           []string                `json:"remainingChapter8Gaps"`
	ReadinessForChapter9    string                  `json:"readinessForChapter9"`
	Conclusion              []string                `json:"conclusion"`
	SourceArtifacts         []string                `json:"sourceArtifacts"`
}

type Chapter8PacketConcept struct {
	BookConcept string `json:"bookConcept"`
	RepoMapping string `json:"repoMapping"`
	Status      string `json:"status"`
}

type Chapter8PacketVenue struct {
	Venue         string `json:"venue"`
	Communication string `json:"communication"`
	PriceUpdates  string `json:"priceUpdates"`
	OrderRequests string `json:"orderRequests"`
	Responses     string `json:"responses"`
	GatewayStatus string `json:"gatewayStatus"`
}

func BuildChapter8Packet() Chapter8Packet {
	audit := BuildChapter8GatewayAudit()
	venueMap := BuildChapter8VenueGatewayMap()
	sessionLifecycle := BuildChapter8SessionLifecycle()

	concepts := make([]Chapter8PacketConcept, 0, len(audit.Concepts)+3)
	for _, concept := range audit.Concepts {
		concepts = append(concepts, Chapter8PacketConcept{
			BookConcept: concept.Concept,
			RepoMapping: strings.Join(concept.RepoFiles, ", "),
			Status:      concept.Status,
		})
	}
	concepts = append(concepts,
		Chapter8PacketConcept{BookConcept: "gateway abstraction", RepoMapping: "system.Gateway, system.SimulatedGateway", Status: "implemented"},
		Chapter8PacketConcept{BookConcept: "venue gateway translators", RepoMapping: "system.VenueGateway and venue translator functions", Status: "implemented disabled by default"},
		Chapter8PacketConcept{BookConcept: "session lifecycle and heartbeat model", RepoMapping: "system.GatewaySession and system.GatewaySessionManager", Status: "implemented in-memory"},
	)

	venues := make([]Chapter8PacketVenue, 0, len(audit.Venues))
	for _, venue := range audit.Venues {
		venues = append(venues, Chapter8PacketVenue{
			Venue:         venue.Venue,
			Communication: venue.CommunicationAPI,
			PriceUpdates:  venue.PriceUpdates,
			OrderRequests: venue.OrderSending,
			Responses:     venue.MarketResponseHandling,
			GatewayStatus: venue.SystemGatewayStatus,
		})
	}

	return Chapter8Packet{
		ImplementedConcepts: concepts,
		VenueAPIMapping:     venues,
		GatewayResponsibilities: []string{
			"Start, stop, and report gateway status.",
			"Receive price updates and convert external venue data into system liquidity/book DTOs.",
			"Accept internal order intents from the OMS path.",
			"Return normalized market responses to the OMS path.",
			"Keep venue adapter details outside strategy, OMS, and system simulation components.",
		},
		PriceUpdateHandling: []string{
			"system.GatewayPriceUpdate models translated venue market data.",
			"GatewayPriceUpdate.ToLiquidityEvents converts bid/ask snapshots into system.LiquidityEvent values.",
			"Aster, Hyperliquid, and Lighter candle translators map normalized exchanges.Candle data into gateway price updates.",
			"Tests use sample DTOs only; no live REST or WebSocket subscriptions are performed.",
		},
		OrderRequestHandling: []string{
			"system.OrderGateway.SendOrder accepts system.OrderIntent.",
			"system.SimulatedGateway can drive the Chapter 7 OrderManager to MarketSimulator path.",
			"system.VenueGateway is disabled by default and blocks order sends unless future phases explicitly allow them.",
			"No live or paper venue order execution is enabled.",
		},
		MarketResponseHandling: []string{
			"system.GatewayOrderResponse wraps normalized venue responses.",
			"Aster, Hyperliquid, and Lighter order-result translators map execution.OrderResult to system.OrderResponse.",
			"Market response statuses are normalized to ACCEPTED, FILLED, CANCELED, AMENDED, or REJECTED.",
			"Venue-specific user stream/sendTx response handling remains outside the enabled system path.",
		},
		SessionLifecycle: []string{
			"GatewaySession statuses: CREATED, CONNECTING, CONNECTED, HEARTBEAT_OK, DEGRADED, DISCONNECTED, ERROR.",
			"GatewaySession commands: Connect, Heartbeat, Disconnect, MarkError, RecordMessage.",
			"GatewaySessionManager registers sessions by venue and supports connect-all, heartbeat-all, disconnect-all, and summary reporting.",
			fmt.Sprintf("Reference packet models %d sessions with %d heartbeat-ok sessions and %d messages.", sessionLifecycle.Summary.Total, sessionLifecycle.Summary.HeartbeatOK, sessionLifecycle.Summary.Messages),
		},
		DisabledSafetyModel: []string{
			"VenueGatewayConfig.Enabled defaults to false.",
			"VenueGatewayConfig.AllowLiveOrders defaults to false.",
			"Disabled venue gateways enter DISABLED status on Start.",
			"SendOrder returns local rejected responses when disabled or when live orders are not allowed.",
			"Chapter 8 reports are generated without real API calls.",
		},
		MissingFIXItems: []string{
			"FIX session logon/logout is documented but not implemented because current crypto venues use REST, WebSocket, and sendTx APIs.",
			"FIX tag-value parsing, body length, checksum, and sequence-number handling are documented gaps.",
			"FIX market-data request/response acceptor behavior is not implemented yet.",
			"FIX order-entry message mapping is not implemented because real venue adapter bridging remains disabled.",
		},
		RemainingGaps: []string{
			"Real venue adapters are not wrapped into system.Gateway.",
			"Gateway reconnect/backoff logic is not implemented.",
			"Request IDs and response correlation are not modeled yet.",
			"Heartbeat timeouts are not enforced automatically.",
			"FIX protocol implementation is documented but intentionally absent.",
			"Paper/live execution remains disabled.",
		},
		ReadinessForChapter9: "Ready for Chapter 9 backtester audit, with exchange connectivity mapped and safely abstracted but not enabled.",
		Conclusion: []string{
			"Chapter 8 exchange connectivity layer is mapped and safely abstracted.",
			"Real venue adapters remain disabled by default.",
			"No live or paper execution is enabled.",
			"The system is ready for Chapter 9 backtester audit.",
		},
		SourceArtifacts: append(append([]string{}, venueMap.Outputs...), []string{
			"research/chapter8_gateway_audit.json",
			"research/chapter8_gateway_audit.md",
			"research/chapter8_session_lifecycle.json",
			"research/chapter8_session_lifecycle.md",
			"system/gateway.go",
			"system/venue_gateway.go",
			"system/gateway_session.go",
		}...),
	}
}

func WriteChapter8PacketJSON(path string, packet Chapter8Packet) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	body, err := json.MarshalIndent(packet, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(body, '\n'), 0644)
}

func WriteChapter8PacketMarkdown(path string, packet Chapter8Packet) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Chapter8PacketMarkdown(packet)), 0644)
}

func Chapter8PacketMarkdown(packet Chapter8Packet) string {
	var b strings.Builder
	b.WriteString("# Chapter 8 Final Exchange Connectivity Packet\n\n")

	b.WriteString("## Concepts Implemented\n\n")
	b.WriteString("| Book Concept | Repo Mapping | Status |\n")
	b.WriteString("|---|---|---|\n")
	for _, row := range packet.ImplementedConcepts {
		b.WriteString(fmt.Sprintf("| %s | %s | %s |\n", row.BookConcept, row.RepoMapping, row.Status))
	}

	b.WriteString("\n## Venue API Mapping\n\n")
	b.WriteString("| Venue | Communication | Price Updates | Order Requests | Responses | Gateway Status |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, row := range packet.VenueAPIMapping {
		b.WriteString(fmt.Sprintf(
			"| %s | %s | %s | %s | %s | %s |\n",
			row.Venue,
			row.Communication,
			row.PriceUpdates,
			row.OrderRequests,
			row.Responses,
			row.GatewayStatus,
		))
	}

	writeStringList(&b, "Gateway Responsibilities", packet.GatewayResponsibilities)
	writeStringList(&b, "Price Update Handling", packet.PriceUpdateHandling)
	writeStringList(&b, "Order Request Handling", packet.OrderRequestHandling)
	writeStringList(&b, "Market Response Handling", packet.MarketResponseHandling)
	writeStringList(&b, "Session Lifecycle / Heartbeat Model", packet.SessionLifecycle)
	writeStringList(&b, "Disabled-By-Default Safety Model", packet.DisabledSafetyModel)
	writeStringList(&b, "Missing FIX Items", packet.MissingFIXItems)
	writeStringList(&b, "Remaining Chapter 8 Gaps", packet.RemainingGaps)

	b.WriteString("\n## Readiness For Chapter 9\n\n")
	b.WriteString(packet.ReadinessForChapter9 + "\n")

	writeStringList(&b, "Conclusion", packet.Conclusion)
	return b.String()
}
