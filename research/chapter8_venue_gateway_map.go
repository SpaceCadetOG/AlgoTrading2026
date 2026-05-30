package research

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Chapter8VenueGatewayTranslator struct {
	Venue              string   `json:"venue"`
	ConfigDefault      string   `json:"configDefault"`
	PriceTranslator    string   `json:"priceTranslator"`
	ResponseTranslator string   `json:"responseTranslator"`
	SourceTypes        []string `json:"sourceTypes"`
	GuardRails         []string `json:"guardRails"`
	Status             string   `json:"status"`
}

type Chapter8VenueGatewayMap struct {
	Title       string                           `json:"title"`
	Translators []Chapter8VenueGatewayTranslator `json:"translators"`
	FilesAdded  []string                         `json:"filesAdded"`
	Outputs     []string                         `json:"outputs"`
	Conclusion  string                           `json:"conclusion"`
	NextPhase   string                           `json:"nextPhase"`
}

func BuildChapter8VenueGatewayMap() Chapter8VenueGatewayMap {
	commonGuards := []string{
		"enabled defaults to false",
		"allowLiveOrders defaults to false",
		"SendOrder returns a rejected gateway response when live orders are disabled",
		"tests use translator DTOs only and do not call real venue APIs",
	}

	return Chapter8VenueGatewayMap{
		Title: "Chapter 8C Venue Gateway Translators",
		Translators: []Chapter8VenueGatewayTranslator{
			{
				Venue:              "aster",
				ConfigDefault:      "disabled dry-run translator",
				PriceTranslator:    "system.AsterCandleToGatewayPriceUpdate",
				ResponseTranslator: "system.AsterOrderResultToGatewayOrderResponse",
				SourceTypes: []string{
					"exchanges.Candle from exchanges/aster REST or WS candle paths",
					"execution.OrderResult from exchanges/aster order methods",
				},
				GuardRails: commonGuards,
				Status:     "translator implemented, real adapter wrapping disabled",
			},
			{
				Venue:              "hyperliquid",
				ConfigDefault:      "disabled dry-run translator",
				PriceTranslator:    "system.HyperliquidCandleToGatewayPriceUpdate",
				ResponseTranslator: "system.HyperliquidOrderResultToGatewayOrderResponse",
				SourceTypes: []string{
					"exchanges.Candle from exchanges/hyperliquid REST or WS candle paths",
					"execution.OrderResult from exchanges/hyperliquid order methods",
				},
				GuardRails: commonGuards,
				Status:     "translator implemented, real adapter wrapping disabled",
			},
			{
				Venue:              "lighter",
				ConfigDefault:      "disabled dry-run translator",
				PriceTranslator:    "system.LighterCandleToGatewayPriceUpdate",
				ResponseTranslator: "system.LighterOrderResultToGatewayOrderResponse",
				SourceTypes: []string{
					"exchanges.Candle from exchanges/lighter REST or WS candle paths",
					"execution.OrderResult from exchanges/lighter sendTx order path",
				},
				GuardRails: commonGuards,
				Status:     "translator implemented, real adapter wrapping disabled",
			},
		},
		FilesAdded: []string{
			"system/venue_gateway.go",
			"system/venue_gateway_test.go",
			"research/chapter8_venue_gateway_map.go",
		},
		Outputs: []string{
			"research/chapter8_venue_gateway_map.json",
			"research/chapter8_venue_gateway_map.md",
		},
		Conclusion: "Chapter 8C provides disabled-by-default venue gateway translators without enabling real order flow, paper trading, or live WebSocket subscriptions.",
		NextPhase:  "Chapter 8D can audit protocol/session concerns such as heartbeats, reconnects, request IDs, and response correlation.",
	}
}

func WriteChapter8VenueGatewayMapJSON(path string, gatewayMap Chapter8VenueGatewayMap) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(gatewayMap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func WriteChapter8VenueGatewayMapMarkdown(path string, gatewayMap Chapter8VenueGatewayMap) error {
	if err := ensureDir(path); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", gatewayMap.Title)
	fmt.Fprintf(&b, "## Translators\n\n")
	fmt.Fprintf(&b, "| Venue | Default | Price translator | Response translator | Source types | Guard rails | Status |\n")
	fmt.Fprintf(&b, "|---|---|---|---|---|---|---|\n")
	for _, translator := range gatewayMap.Translators {
		fmt.Fprintf(
			&b,
			"| %s | %s | %s | %s | %s | %s | %s |\n",
			translator.Venue,
			translator.ConfigDefault,
			translator.PriceTranslator,
			translator.ResponseTranslator,
			strings.Join(translator.SourceTypes, "<br>"),
			strings.Join(translator.GuardRails, "<br>"),
			translator.Status,
		)
	}

	fmt.Fprintf(&b, "\n## Files Added\n\n")
	for _, file := range gatewayMap.FilesAdded {
		fmt.Fprintf(&b, "- %s\n", file)
	}

	fmt.Fprintf(&b, "\n## Conclusion\n\n%s\n\n", gatewayMap.Conclusion)
	fmt.Fprintf(&b, "## Next Phase\n\n%s\n", gatewayMap.NextPhase)

	return os.WriteFile(path, []byte(b.String()), 0644)
}
