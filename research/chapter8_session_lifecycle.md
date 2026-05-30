# Chapter 8D Gateway Session Lifecycle

## Statuses

- CREATED
- CONNECTING
- CONNECTED
- HEARTBEAT_OK
- DEGRADED
- DISCONNECTED
- ERROR

## Commands

- Connect
- Heartbeat
- Disconnect
- MarkError
- RecordMessage

## Sessions

| Venue | Session ID | Status | Messages | Last error | Network behavior |
|---|---|---|---|---|---|
| aster | aster-session | HEARTBEAT_OK | 1 |  | no real sockets, REST calls, WebSockets, or order placement |
| hyperliquid | hyperliquid-session | HEARTBEAT_OK | 1 |  | no real sockets, REST calls, WebSockets, or order placement |
| lighter | lighter-session | HEARTBEAT_OK | 1 |  | no real sockets, REST calls, WebSockets, or order placement |

## Summary

- total: 3
- heartbeat_ok: 3
- messages: 3
- errors: 0

## Conclusion

Chapter 8D models gateway session lifecycle, heartbeats, errors, disconnects, and message accounting without real network calls.

## Next Phase

Chapter 8E can model request IDs and response correlation before any real adapter bridge is enabled.
