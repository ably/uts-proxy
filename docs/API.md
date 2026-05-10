# API Reference

The UTS proxy exposes a control API for managing test sessions and a per-session proxy that handles HTTP and WebSocket traffic.

## Control API

The control API listens on a configurable port (default `9100`, set with `--port`). All requests and responses use JSON.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/sessions` | Create a new test session |
| `GET` | `/sessions/{id}` | Get session metadata |
| `POST` | `/sessions/{id}/rules` | Add rules to an existing session |
| `POST` | `/sessions/{id}/actions` | Trigger an imperative action |
| `GET` | `/sessions/{id}/log` | Get the session's traffic log |
| `DELETE` | `/sessions/{id}` | Teardown session and return final log |

---

### `GET /health`

Returns a simple health check.

**Response `200`:**

```json
{ "ok": true }
```

---

### `POST /sessions`

Creates a new test session, binds the specified port, and starts proxying.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `port` | integer | no | TCP port for the session's proxy listener. If omitted or `0`, a free port is auto-assigned. |
| `target` | object | yes | Upstream hosts (at least one of `realtimeHost` or `restHost`) |
| `target.realtimeHost` | string | no | Upstream WebSocket host (e.g. `sandbox-realtime.ably.io`) |
| `target.restHost` | string | no | Upstream REST host (e.g. `sandbox-rest.ably.io`) |
| `target.insecure` | boolean | no | If `true`, connect to upstream without TLS (default `false`) |
| `rules` | array | no | Initial rules (see [Rule Format](#rule-format) below) |
| `timeoutMs` | integer | no | Auto-cleanup timeout in milliseconds (default `30000`) |

**Response `201`:**

```json
{
  "sessionId": "a1b2c3d4",
  "proxy": {
    "host": "localhost:10042",
    "port": 10042
  }
}
```

**Response `400`:** Missing or invalid fields.

**Response `409`:** Port already in use.

---

### `GET /sessions/{id}`

Returns metadata about an existing session.

**Response `200`:**

```json
{
  "sessionId": "a1b2c3d4",
  "port": 10042,
  "target": {
    "realtimeHost": "sandbox-realtime.ably.io",
    "restHost": "sandbox-rest.ably.io"
  },
  "ruleCount": 3
}
```

**Response `404`:** Session not found.

---

### `POST /sessions/{id}/rules`

Adds rules to an existing session. Rules can be appended (default) or prepended.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `rules` | array | yes | Rules to add (see [Rule Format](#rule-format)) |
| `position` | string | no | `"append"` (default) or `"prepend"` |

**Response `200`:**

```json
{ "ruleCount": 5 }
```

**Response `404`:** Session not found.

---

### `POST /sessions/{id}/actions`

Triggers an imperative action on the session's active WebSocket connection.

**Request body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | yes | Action type (see below) |
| `message` | object | no | Protocol message to inject (for `inject_to_client`, `inject_to_client_and_close`) |
| `closeCode` | integer | no | WebSocket close code (for `close`) |

**Imperative action types:**

| Type | Description |
|------|-------------|
| `disconnect` | Abruptly close the TCP connection (no close frame) |
| `close` | Send a WebSocket close frame with the specified `closeCode` (default `1000`), then close |
| `inject_to_client` | Send a protocol message to the client as if it came from the server |
| `inject_to_client_and_close` | Send a protocol message to the client, then close the WebSocket |

**Response `200`:**

```json
{ "ok": true }
```

**Response `404`:** Session not found.

**Response `409`:** No active WebSocket connection.

---

### `GET /sessions/{id}/log`

Returns the session's complete traffic log.

**Response `200`:**

```json
{
  "events": [ ...events... ]
}
```

See [Event Log Format](#event-log-format) for the event schema.

**Response `404`:** Session not found.

---

### `DELETE /sessions/{id}`

Tears down the session: closes all active connections, stops the listener, frees the port, and returns the final traffic log.

**Response `200`:**

```json
{
  "events": [ ...final events... ]
}
```

**Response `404`:** Session not found.

---

## Rule Format

Each rule has a **match** condition, an **action** to perform, and optional metadata:

```json
{
  "match": { ... },
  "action": { ... },
  "times": 1,
  "comment": "human-readable description"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `match` | object | yes | When this rule fires (see [Match Conditions](#match-conditions)) |
| `action` | object | yes | What happens when it fires (see [Actions](#actions)) |
| `times` | integer | no | Remove the rule after this many firings. `0` or omitted = unlimited. |
| `comment` | string | no | Human-readable label (appears in log as `ruleMatched`) |

Rules are evaluated **in order**. The first matching rule wins. Unmatched traffic passes through unchanged.

### Match Conditions

The `match` object always has a `type` field and optional filters:

| Field | Type | Applies to | Description |
|-------|------|-----------|-------------|
| `type` | string | all | Required. The event type to match. |
| `count` | integer | all | Only match the Nth occurrence (1-based). |
| `action` | string | `ws_frame_to_server`, `ws_frame_to_client` | Ably protocol action name (e.g. `"ATTACH"`, `"CONNECTED"`) or number (e.g. `"15"`, `"4"`). |
| `channel` | string | `ws_frame_to_server`, `ws_frame_to_client` | Channel name must equal this. |
| `method` | string | `http_request` | HTTP method (case-insensitive). |
| `pathContains` | string | `http_request` | Request path must contain this substring. |
| `queryContains` | object | `ws_connect` | Map of query param key→value. Use `"*"` to match any value. |
| `delayMs` | integer | `delay_after_ws_connect` | Milliseconds after WebSocket connection before firing. |

#### Match types

**`ws_connect`** — Matches when a WebSocket connection is established.

```json
{ "type": "ws_connect" }
{ "type": "ws_connect", "count": 2 }
{ "type": "ws_connect", "queryContains": { "resume": "*" } }
```

**`ws_frame_to_server`** — Matches a WebSocket frame sent from client to server.

```json
{ "type": "ws_frame_to_server", "action": "ATTACH", "channel": "my-channel" }
```

**`ws_frame_to_client`** — Matches a WebSocket frame sent from server to client.

```json
{ "type": "ws_frame_to_client", "action": "CONNECTED" }
{ "type": "ws_frame_to_client", "action": "HEARTBEAT" }
```

**`http_request`** — Matches an HTTP request.

```json
{ "type": "http_request", "method": "POST", "pathContains": "/channels/" }
```

**`delay_after_ws_connect`** — A temporal trigger that fires once, `delayMs` milliseconds after the WebSocket connection is established.

```json
{ "type": "delay_after_ws_connect", "delayMs": 5000 }
```

### Count and Times

These two fields control *when* and *how many times* a rule fires:

- **`count`** (on the match): Only fires on the Nth occurrence of a matching event (1-based). The proxy tracks how many times the base condition (type + filters) has matched across the session. Useful for targeting specific connection attempts (e.g., "refuse the 2nd connection").

- **`times`** (on the rule): After the rule has fired this many times, it is removed. Useful for one-shot faults (e.g., "return 401 once, then passthrough").

### Actions

#### Passthrough

Forward the frame/request unchanged. This is the default when no rule matches.

```json
{ "type": "passthrough" }
```

#### Connection-level actions

| Action | Fields | Description |
|--------|--------|-------------|
| `refuse_connection` | — | Refuse the WebSocket upgrade at TCP level. |
| `accept_and_close` | `closeCode` | Accept the WebSocket handshake, then immediately close with the given code. |
| `disconnect` | — | Abruptly close the TCP connection (no close frame). |
| `close` | `closeCode` | Send a WebSocket close frame with the given code. |

#### Frame manipulation actions

| Action | Fields | Description |
|--------|--------|-------------|
| `suppress` | — | Swallow the frame — don't forward it. |
| `delay` | `delayMs` | Wait `delayMs` milliseconds before forwarding. |
| `inject_to_client` | `message` | Send an additional frame to the client (as if from the server), then forward the original. |
| `inject_to_client_and_close` | `message`, `closeCode` | Send a frame to the client, then close the WebSocket. |
| `replace` | `message` | Replace the frame with a different one. |
| `suppress_onwards` | — | Suppress all subsequent frames in the same direction for this connection. |

#### HTTP actions

| Action | Fields | Description |
|--------|--------|-------------|
| `http_respond` | `status`, `body`, `headers` | Return a specified HTTP response without forwarding to upstream. |
| `http_delay` | `delayMs` | Delay `delayMs` milliseconds before forwarding to upstream. |
| `http_drop` | — | Drop the connection without sending any response. |
| `http_replace_response` | `status`, `body`, `headers` | Forward to upstream, discard the real response, and return the specified response instead. |

---

## Event Log Format

Every frame, request, and action is recorded in the session's event log. Each event has the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `timestamp` | string | ISO 8601 timestamp (UTC) |
| `type` | string | Event type: `ws_connect`, `ws_frame`, `ws_disconnect`, `http_request`, `http_response`, `action` |
| `direction` | string | `client_to_server` or `server_to_client` (for frames and requests) |
| `url` | string | WebSocket connection URL (for `ws_connect`) |
| `queryParams` | object | WebSocket connection query parameters (for `ws_connect`) |
| `message` | object | Decoded protocol message (for `ws_frame`) or action payload (for `action`) |
| `method` | string | HTTP method (for `http_request`) |
| `path` | string | HTTP request path (for `http_request`) |
| `status` | integer | HTTP response status code (for `http_response`) |
| `initiator` | string | Who caused the event: `client`, `server`, or `proxy` |
| `closeCode` | integer | WebSocket close code (for `ws_disconnect`) |
| `ruleMatched` | string or null | Label of the rule that matched (comment or `rule-N`), or `null` if no rule matched |
| `headers` | object | HTTP headers (for `http_request`, `http_response`) |

All fields except `timestamp` and `type` are omitted when not applicable.

### Example log

```json
{
  "events": [
    {
      "timestamp": "2026-01-15T10:00:00.123Z",
      "type": "ws_connect",
      "url": "ws://localhost:10042/?key=app.key&heartbeats=true",
      "queryParams": { "key": "app.key", "heartbeats": "true" },
      "ruleMatched": null
    },
    {
      "timestamp": "2026-01-15T10:00:00.200Z",
      "type": "ws_frame",
      "direction": "server_to_client",
      "message": { "action": 4, "connectionId": "abc123" },
      "ruleMatched": null
    },
    {
      "timestamp": "2026-01-15T10:00:01.500Z",
      "type": "ws_disconnect",
      "initiator": "proxy",
      "closeCode": 1006,
      "ruleMatched": null
    },
    {
      "timestamp": "2026-01-15T10:00:02.000Z",
      "type": "ws_connect",
      "url": "ws://localhost:10042/?key=app.key&resume=abc123&heartbeats=true",
      "queryParams": { "key": "app.key", "resume": "abc123", "heartbeats": "true" },
      "ruleMatched": null
    }
  ]
}
```

---

## Usage Patterns

### Imperative disconnect (test connection resume)

Create a passthrough session, let the SDK connect normally, then trigger a disconnect:

```bash
# Create session
curl -X POST http://localhost:9100/sessions \
  -d '{"port": 10042, "target": {"realtimeHost": "sandbox-realtime.ably.io"}}'

# ... SDK connects, reaches CONNECTED state ...

# Trigger disconnect
curl -X POST http://localhost:9100/sessions/{id}/actions \
  -d '{"type": "disconnect"}'

# ... SDK reconnects with resume ...

# Verify: expect 2 ws_connect events, second has resume param
curl http://localhost:9100/sessions/{id}/log
```

### One-shot connection refusal

Refuse the first connection attempt, then let subsequent attempts through:

```json
{
  "port": 10042,
  "target": {"realtimeHost": "sandbox-realtime.ably.io"},
  "rules": [{
    "match": {"type": "ws_connect", "count": 1},
    "action": {"type": "refuse_connection"},
    "times": 1,
    "comment": "refuse first connection"
  }]
}
```

### Injected error with close

Inject a token-expired error 1 second after connection, then close:

```json
{
  "port": 10042,
  "target": {"realtimeHost": "sandbox-realtime.ably.io"},
  "rules": [{
    "match": {"type": "delay_after_ws_connect", "delayMs": 1000},
    "action": {
      "type": "inject_to_client_and_close",
      "message": {
        "action": 6,
        "error": {"code": 40142, "statusCode": 401, "message": "Token expired"}
      }
    },
    "times": 1
  }]
}
```

### REST 401 for token renewal

Return a fake 401 on the first channel request, forcing the SDK to renew its token:

```json
{
  "port": 10042,
  "target": {"restHost": "sandbox-rest.ably.io"},
  "rules": [{
    "match": {"type": "http_request", "pathContains": "/channels/"},
    "action": {
      "type": "http_respond",
      "status": 401,
      "body": {"error": {"code": 40142, "statusCode": 401, "message": "Token expired"}}
    },
    "times": 1
  }]
}
```

### Heartbeat starvation

Suppress all server-to-client frames after 2 seconds, causing the SDK's heartbeat timer to fire:

```json
{
  "port": 10042,
  "target": {"realtimeHost": "sandbox-realtime.ably.io"},
  "rules": [{
    "match": {"type": "delay_after_ws_connect", "delayMs": 2000},
    "action": {"type": "suppress_onwards"},
    "times": 1
  }]
}
```

### Two-phase dynamic rules

Start with a clean passthrough, let the SDK reach a stable state, then add a fault rule:

```bash
# Create passthrough session
curl -X POST http://localhost:9100/sessions \
  -d '{"port": 10042, "target": {"realtimeHost": "sandbox-realtime.ably.io"}}'

# ... SDK connects and attaches channels ...

# Add a rule to suppress the next DETACH response
curl -X POST http://localhost:9100/sessions/{id}/rules \
  -d '{
    "rules": [{
      "match": {"type": "ws_frame_to_client", "action": "DETACHED", "channel": "test"},
      "action": {"type": "suppress"},
      "times": 1
    }],
    "position": "prepend"
  }'

# ... SDK attempts detach, timeout fires ...
```

---

## Protocol Actions

The proxy recognises Ably protocol message actions by name or number:

| Action | Number |
|--------|--------|
| HEARTBEAT | 0 |
| ACK | 1 |
| NACK | 2 |
| CONNECT | 3 |
| CONNECTED | 4 |
| DISCONNECT | 5 |
| DISCONNECTED | 6 |
| CLOSE | 7 |
| CLOSED | 8 |
| ERROR | 9 |
| ATTACH | 10 |
| ATTACHED | 11 |
| DETACH | 12 |
| DETACHED | 13 |
| PRESENCE | 14 |
| MESSAGE | 15 |
| SYNC | 16 |
| AUTH | 17 |

Rules can use either the name (`"action": "ATTACH"`) or number (`"action": "10"`).
