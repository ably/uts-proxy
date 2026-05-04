# Ably UTS Proxy

[![License](https://img.shields.io/github/license/ably/uts-proxy)](LICENSE)
[![CI](https://github.com/ably/uts-proxy/actions/workflows/ci.yml/badge.svg)](https://github.com/ably/uts-proxy/actions/workflows/ci.yml)

A programmable HTTP/WebSocket proxy for [Ably](https://ably.com) SDK integration testing. It sits between the SDK under test and the real Ably sandbox, transparently forwarding traffic by default but injecting faults — dropped connections, modified responses, delayed frames, injected protocol messages — when configured with rules.

---

## Overview

```
┌──────────┐         ┌──────────────────────────┐         ┌───────────────┐
│   SDK    │──WS/HTTP──▶  uts-proxy             │──WSS/HTTPS──▶ Ably Sandbox │
│  under   │◀──────────  (per-session ports)    │◀────────────  (real backend)│
│  test    │         │                          │         └───────────────┘
└──────────┘         │  Control API (:9100)     │
                     └────────────▲─────────────┘
                                  │
                     ┌────────────┴─────────────┐
                     │  Test process             │
                     │  (creates sessions,       │
                     │   configures rules,       │
                     │   triggers actions,       │
                     │   inspects traffic logs)  │
                     └──────────────────────────┘
```

The proxy serves multiple concurrent test sessions, each on its own TCP port. A **control API** on a dedicated port (default `:9100`) manages session lifecycle:

1. **Create a session** — bind a port, set the upstream target, optionally configure rules
2. **SDK connects** through the session port using plain HTTP/WS (the proxy handles TLS upstream)
3. **Rules match** on WebSocket frames (by Ably protocol action, channel, direction) or HTTP requests (by method, path) and fire actions — suppress, delay, inject, replace, disconnect, or return fake responses
4. **Imperative actions** let tests trigger faults on demand (disconnect, close, inject a message)
5. **Traffic logs** capture every frame and request for test assertions
6. **Teardown** closes connections, frees the port, and returns the final log

The proxy is protocol-aware: it decodes Ably protocol messages in both JSON and msgpack for rule matching, but forwards raw bytes unchanged.

---

## Supported Platforms

Prebuilt binaries are available for:

| OS | Architecture |
|----|-------------|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |

Building from source requires **Go 1.22** or later and works on any platform Go supports.

---

## Installation

### Download a prebuilt binary

Download the latest release from [GitHub Releases](https://github.com/ably/uts-proxy/releases), extract it, and place it on your `PATH`.

### Build from source

```bash
go build -o uts-proxy .
```

---

## Quick Start

Start the proxy:

```bash
./uts-proxy --port 9100
```

Check health:

```bash
curl http://localhost:9100/health
# {"ok":true}
```

Create a session that proxies to the Ably sandbox:

```bash
curl -X POST http://localhost:9100/sessions \
  -H 'Content-Type: application/json' \
  -d '{
    "port": 10042,
    "target": {
      "realtimeHost": "sandbox-realtime.ably.io",
      "restHost": "sandbox-rest.ably.io"
    },
    "rules": [{
      "match": {"type": "ws_connect", "count": 1},
      "action": {"type": "refuse_connection"},
      "times": 1,
      "comment": "refuse first connection, then passthrough"
    }],
    "timeoutMs": 30000
  }'
# {"sessionId":"a1b2c3d4","proxy":{"host":"localhost:10042","port":10042}}
```

Point your SDK at the proxy (plain WS/HTTP, no TLS):

```
ws://localhost:10042/?key=YOUR_KEY&heartbeats=true
http://localhost:10042/channels/test/messages
```

Inspect the traffic log:

```bash
curl http://localhost:9100/sessions/a1b2c3d4/log
```

Teardown the session (returns the final log):

```bash
curl -X DELETE http://localhost:9100/sessions/a1b2c3d4
```

---

## API Reference

See [docs/API.md](docs/API.md) for the complete control API reference, including all endpoints, rule match types, action types, and the event log format.

---

## Design

See [docs/DESIGN.md](docs/DESIGN.md) for the original design proposal, including architecture, motivation, and usage patterns.

---

## Releases

This project uses [Semantic Versioning](https://semver.org/).

The [CHANGELOG](CHANGELOG.md) contains details of each release. You can also find prebuilt binaries on the [GitHub Releases](https://github.com/ably/uts-proxy/releases) page.

---

## Contributing

Read the [CONTRIBUTING.md](CONTRIBUTING.md) guidelines to contribute to this project.

---

## Support, feedback and troubleshooting

For help or technical support, visit Ably's [support page](https://ably.com/support) or [open an issue](https://github.com/ably/uts-proxy/issues) on GitHub.
