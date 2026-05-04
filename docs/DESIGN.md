# Ably UTS Proxy — Design

> **Note:** This is the original design proposal for the test proxy. For the authoritative API reference, see [API.md](API.md).

## Overview

A programmable HTTP/WebSocket proxy that sits between an Ably SDK under test and the real Ably sandbox backend. The proxy transparently forwards traffic by default, but can be configured with **rules** to inject faults — dropped connections, modified responses, injected protocol messages, delayed frames, etc.

This enables **integration tests for fault behaviour** that would otherwise require mocking. The proxy gives tests the realism of talking to the actual Ably sandbox while retaining the ability to simulate network and protocol faults.

## Motivation

The existing UTS unit tests use mock HTTP/WebSocket clients to test fault handling (connection failures, token expiry, heartbeat starvation, channel errors, etc.). These are valuable but have limitations:

- They test against synthetic responses, not the real server protocol
- They cannot verify that resume actually works end-to-end with a real server
- They require the test to script every server response, including the "happy path" ones

A proxy-based approach lets tests rely on the real sandbox for normal behaviour and only inject specific faults. This increases confidence that the SDK handles real-world failure modes correctly.

## Architecture

```
                     ┌────────────────────────────────────────────┐
                     │          Ably Test Proxy (single process)  │
                     │                                            │
┌──────────┐         │  ┌──────────────────┐                      │      ┌───────────────┐
│  SDK     │────WS──▶│  │ :10042 (session1)│───wss──────────────▶│──────│ Ably Sandbox  │
│  under   │◀───────▶│  │ :10043 (session2)│◀──────────────────▶│  │◀────│ (real backend) │
│  test    │──HTTP──▶│  │ ...              │───https─────────────│──────│               │
└──────────┘         │  └──────────────────┘                      │      └───────────────┘
                     │                                            │
                     │  ┌──────────────────┐                      │
                     │  │ :9100 control API │                      │
                     │  └──────────────────┘                      │
                     └────────────────────────────────────────────┘
                              ▲
                              │ HTTP control API
                     ┌────────┴────────────┐
                     │  Test process        │
                     │  (creates sessions,  │
                     │   assigns ports,     │
                     │   adds rules,        │
                     │   triggers actions)  │
                     └─────────────────────┘
```

- **Single proxy process** serves multiple concurrent test sessions
- **Control API** (HTTP on a dedicated port, e.g. `:9100`) manages sessions and rules
- **Per-session ports** (assigned by the test process from a port pool) handle proxied WS and HTTP traffic. Each session binds its own TCP listener so the SDK can connect with standard URL paths.
- **No TLS between client and proxy.** The proxy serves plain HTTP/WS to the SDK. Upstream connections to the Ably sandbox use TLS (`wss://`, `https://`).
- **Default behaviour** is transparent passthrough to the real Ably sandbox
- **Protocol-aware for both JSON and msgpack.** The proxy decodes frames in both formats for rule matching. Raw bytes are forwarded unchanged (no re-encoding).

## Key Design Decisions

### Port-per-session routing

Each test session gets its own TCP port rather than routing by URL prefix. This allows the SDK to connect with standard Ably URL paths (e.g., `/channels/test/messages`) without any proxy-specific URL modifications.

### No TLS between client and proxy

The proxy accepts plain HTTP/WS from the SDK and handles TLS upstream. This avoids the complexity of certificate management for local testing and simplifies SDK configuration (just set `tls: false` and point at `localhost`).

### Protocol-aware but non-modifying

The proxy decodes Ably protocol messages (both JSON and msgpack) for rule matching, but forwards the original raw bytes unchanged. This avoids introducing encoding artifacts while enabling fine-grained matching by action, channel, and other protocol fields.

### Declarative rules with imperative escape hatch

Most faults are expressed as declarative rules (match condition → action), but tests can also trigger actions imperatively via the control API. This covers both timed faults (rules with temporal triggers) and on-demand faults (imperative disconnect after the test reaches a specific state).

## Scope and Non-Goals

### In scope

- WebSocket proxying with Ably protocol message awareness (JSON and msgpack)
- HTTP proxying for REST API calls
- Rule-based fault injection (connection, frame, and HTTP levels)
- Imperative actions (disconnect, close, inject)
- Traffic capture and logging
- Concurrent sessions on separate ports for parallel tests

### Not in scope

- Fake timers / time advancement (integration tests use real time with short configured timeouts)
- Mock authUrl server (tests can spin up their own if needed)
- TLS between client and proxy (proxy serves plain HTTP/WS; TLS is used only upstream to sandbox)
- Modifying the SDK's internal state
