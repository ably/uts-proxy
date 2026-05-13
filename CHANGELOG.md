# Changelog

## [0.2.0](https://github.com/ably/uts-proxy/tree/v0.2.0)

### Features

- Allow proxy to auto-assign a free port when creating sessions: the `port` field in `POST /sessions` is now optional — omit it or pass `0` to let the OS pick a free port. The actual bound port is returned in the response.

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
