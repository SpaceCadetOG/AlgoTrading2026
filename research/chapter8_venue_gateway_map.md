# Chapter 8C Venue Gateway Translators

## Translators

| Venue | Default | Price translator | Response translator | Source types | Guard rails | Status |
|---|---|---|---|---|---|---|
| aster | disabled dry-run translator | system.AsterCandleToGatewayPriceUpdate | system.AsterOrderResultToGatewayOrderResponse | exchanges.Candle from exchanges/aster REST or WS candle paths<br>execution.OrderResult from exchanges/aster order methods | enabled defaults to false<br>allowLiveOrders defaults to false<br>SendOrder returns a rejected gateway response when live orders are disabled<br>tests use translator DTOs only and do not call real venue APIs | translator implemented, real adapter wrapping disabled |
| hyperliquid | disabled dry-run translator | system.HyperliquidCandleToGatewayPriceUpdate | system.HyperliquidOrderResultToGatewayOrderResponse | exchanges.Candle from exchanges/hyperliquid REST or WS candle paths<br>execution.OrderResult from exchanges/hyperliquid order methods | enabled defaults to false<br>allowLiveOrders defaults to false<br>SendOrder returns a rejected gateway response when live orders are disabled<br>tests use translator DTOs only and do not call real venue APIs | translator implemented, real adapter wrapping disabled |
| lighter | disabled dry-run translator | system.LighterCandleToGatewayPriceUpdate | system.LighterOrderResultToGatewayOrderResponse | exchanges.Candle from exchanges/lighter REST or WS candle paths<br>execution.OrderResult from exchanges/lighter sendTx order path | enabled defaults to false<br>allowLiveOrders defaults to false<br>SendOrder returns a rejected gateway response when live orders are disabled<br>tests use translator DTOs only and do not call real venue APIs | translator implemented, real adapter wrapping disabled |

## Files Added

- system/venue_gateway.go
- system/venue_gateway_test.go
- research/chapter8_venue_gateway_map.go

## Conclusion

Chapter 8C provides disabled-by-default venue gateway translators without enabling real order flow, paper trading, or live WebSocket subscriptions.

## Next Phase

Chapter 8D can audit protocol/session concerns such as heartbeats, reconnects, request IDs, and response correlation.
