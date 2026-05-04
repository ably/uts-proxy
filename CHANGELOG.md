# Changelog

## [0.1.0](https://github.com/ably/uts-proxy/tree/v0.1.0)

Initial release. Extracted from the [Ably specification](https://github.com/ably/specification) repository.

### Features

- Programmable HTTP and WebSocket proxy for SDK integration testing
- Control API for session management, rule configuration, and fault injection
- Declarative rule-based matching on WebSocket frames and HTTP requests
- Imperative actions for ad-hoc fault injection mid-test
- Temporal triggers for time-delayed fault injection
- Ably protocol-aware matching (JSON and msgpack)
- Complete traffic logging for test assertions
- Multiple concurrent test sessions on separate ports
