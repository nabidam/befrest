# Befrest — ARCHITECTURE

Living document — the current technical truth. Derived from `specs/001-core/PRD.md` and `UX.md`. Every contract named here is buildable; implementers build only what this document names.

---

## 1. System overview

Befrest is a **hub-and-browser-spokes** system. One Go binary (the hub) runs on a host machine; every device — including the hub's own machine — participates as a browser client of the SPA the hub serves. The hub is simultaneously:

1. a **static file server** for the embedded SPA,
2. a **WebSocket switchboard** for presence and transfer control messages,
3. a **streaming relay** that pipes file bytes from the sender's HTTP upload directly into the receiver's HTTP download.

There is no database and no persistence. All state (devices, transfers) lives in hub memory for the lifetime of the process; file bytes flow through a bounded in-memory pipe and are never stored (PRD C-4, NFR-2).

```
┌─────────────┐   WS control    ┌──────────────────────┐   WS control   ┌──────────────┐
│ Sender      │◀───────────────▶│        HUB (Go)      │◀──────────────▶│ Receiver     │
│ browser SPA │                 │  presence · transfer │                │ browser SPA  │
│             │──POST upload───▶│   bounded pipe relay │──GET download─▶│ (→ Downloads)│
└─────────────┘   (stream)      └──────────────────────┘   (stream)     └──────────────┘
```

### Committed stack

| Concern | Decision |
|---|---|
| Hub language/runtime | Go 1.25, single static binary per OS/arch via cross-compilation |
| HTTP server | `net/http` stdlib, one `http.ServeMux` |
| WebSocket | `github.com/coder/websocket` |
| System tray | `fyne.io/systray` |
| mDNS announce | `github.com/grandcat/zeroconf` (announce `befrest.local` A/AAAA + `_http._tcp` service) |
| Browser open | `github.com/pkg/browser` |
| Frontend | Svelte 5 + TypeScript + Vite; built output embedded via `embed.FS` |
| QR rendering | client-side, `qrcode` npm package (SVG output) |
| Client state | Svelte runes stores (no external state library) |
| Styling | hand-written CSS custom properties; no CSS framework |
| Client persistence | `localStorage` (device name + device id only) |
| Testing | Go: stdlib `testing` + `httptest`; frontend: Vitest; e2e: Playwright |

---

## 2. Module responsibilities & boundaries

### Hub (Go) — package layout

```
cmd/befrest/            main: flag parsing, tray, wiring, browser open
internal/netinfo/       interface enumeration + ranking, mDNS announce,
                        reachability self-probe (firewall hint)
internal/presence/      Device registry: join/leave/rename, name dedup,
                        device-list fanout
internal/transfer/      Transfer lifecycle state machine, relay pipes,
                        progress accounting
internal/server/        HTTP mux, WS endpoint, upload/download handlers,
                        embedded SPA serving; translates wire messages
                        to presence/transfer calls
internal/proto/         Wire message types (single source of truth for
                        the WS protocol, mirrored by web/src/lib/proto.ts)
web/                    Svelte app; `web/dist` embedded into the binary
```

**Boundary rules**

- `presence` and `transfer` know nothing about HTTP or WebSocket — they operate on interfaces (`Notifier` callback for pushing messages to a device). Only `server` touches the network.
- `transfer` owns all byte flow; `server`'s upload/download handlers hand their `io.Reader`/`http.ResponseWriter` to `transfer` and get back an error.
- `netinfo` is consulted once at startup (plus on M3 re-pick); nothing else reads network interfaces.
- `proto` types are the only shapes crossing the WS; no ad-hoc maps.

### Client (Svelte) — source layout

```
web/src/
  lib/proto.ts          wire types (mirror of internal/proto)
  lib/ws.ts             WS connection, auto-reconnect w/ backoff, message bus
  lib/stores.ts         self, devices, transfers, invite, connection state
  lib/upload.ts         sequential file POST with per-file streaming
  lib/download.ts       triggers anchor downloads on file-ready messages
  App.svelte            top-level state switch (S1 vs S2) + layer host
  screens/              one component per UX id (see §6)
```

---

## 3. Data model

No persistent store exists by design (PRD C-4); the schema below is the **in-memory model** with its invariants — the moral equivalent of DDL for this system.

```go
// internal/presence
type Device struct {
    ID        string    // uuidv4, minted by hub at first hello, echoed by client thereafter
    Name      string    // display name AFTER dedup suffixing; 1–32 chars (V-1)
    RawName   string    // name as requested by client (pre-suffix)
    Kind      string    // "mobile" | "desktop"  (derived from user-agent, drives card icon)
    IsHost    bool      // true for the hub machine's own page (host token)
    ConnectedAt time.Time
    // invariant: Name is unique among currently connected devices (V-2 / FR-2.6)
    // invariant: a Device exists ⟺ its WS is open (FR-3.3)
}

// internal/transfer
type Transfer struct {
    ID         string        // uuidv4 — doubles as the capability for upload/download URLs
    SenderID   string        // FK → Device.ID, must be connected at offer time (V-6)
    ReceiverID string        // FK → Device.ID, must be connected at offer time (V-6)
    Files      []FileMeta    // len ≥ 1 (V-3)
    State      TransferState
    CreatedAt  time.Time
}

type FileMeta struct {
    Index int    // 0-based position; upload/download URLs address files by Index
    Name  string // filename as reported by sender's browser
    Size  int64  // bytes; ≥ 0 (V-3; 0 allowed per EC-9)
    Sent  int64  // bytes relayed so far (progress accounting)
}

type TransferState int // Offered → Accepted → Streaming → Done
                       //         ↘ Declined   ↘ Failed(reason) / Cancelled(by)
```

**Registry indices** (Go maps, mutex-guarded):
- `devices: map[deviceID]*Device` + `namesInUse: map[string]deviceID` (dedup check O(1))
- `transfers: map[transferID]*Transfer` + `byDevice: map[deviceID][]transferID` (cleanup on disconnect)

**Lifecycle invariants**
- A `Transfer` in `Offered` state is destroyed if sender or receiver disconnects (E-3/E-4/E-5 → notify counterpart).
- Terminal states (`Done`, `Declined`, `Failed`, `Cancelled`) release all pipe resources immediately; the record itself is dropped 60 s after terminal (long enough for late progress queries, no history per C-4).
- Multiple concurrent transfers per device are allowed (FR-4.6); offers to one receiver queue client-side (EC-5 — hub sends all offers; the client displays one M1 at a time).

**Client-side persistence** (`localStorage`): `befrest.deviceId`, `befrest.name`. Nothing else. Presence of `befrest.name` ⇒ skip S1 (FR-2.4).

---

## 4. API contract

Auth model: **none** (PRD C-3, trusted LAN). Capability discipline: transfer upload/download URLs contain the unguessable `transferId` (uuidv4) and are validated against transfer state and the caller's role, so a URL is only usable by the intended party at the intended time. The host page authenticates as host via a one-time `hostToken` query param embedded in the URL the hub opens at launch.

### 4.1 HTTP endpoints

| Method & path | Purpose | Request | Success | Errors |
|---|---|---|---|---|
| `GET /` and static assets | Serve embedded SPA | — | `200` HTML/JS/CSS | — |
| `GET /ws` | Upgrade to control WebSocket | plain upgrade — identity travels in the `hello` frame, not query params | `101` | `400` bad upgrade |
| `POST /api/transfers/{tid}/files/{idx}` | Sender streams one file's bytes | raw body, `Content-Length` required = `FileMeta.Size` | `200 {"ok":true}` after fully relayed | `404` unknown tid/idx · `409` transfer not in `Accepted/Streaming` or file already sent · `411` missing length · `400` length ≠ declared size · `499`* receiver gone mid-stream |
| `GET /api/transfers/{tid}/files/{idx}` | Receiver downloads one file's bytes | — | `200`, `Content-Type: application/octet-stream`, `Content-Length: size`, `Content-Disposition: attachment; filename="…"` | `404` unknown · `409` wrong state / already downloaded · stream abort on sender failure |

*`499`-style: connection reset; sender client treats any non-200/short write as `transfer-failed` and awaits the authoritative WS event.

Upload and download are **one-shot per file index**: the handlers join a `transfer.Pipe(tid, idx)` rendezvous; bytes are copied `upload body → bounded pipe (4 MiB ring) → download response`. Backpressure is natural: a slow receiver blocks the pipe, which blocks the upload read, which TCP-throttles the sender. Hub memory per active file = 4 MiB, independent of file size (NFR-2, AC-7). **No disk spooling** — commitment: blocking relay only.

### 4.2 WebSocket protocol (`/ws`)

JSON text frames, `{"type": "...", ...}`. Defined in `internal/proto` / `web/src/lib/proto.ts`.

**Client → Hub**

| type | payload | notes |
|---|---|---|
| `hello` | `{deviceId?, name?, hostToken?}` | first frame after connect. `Kind` and the name suggestion are derived hub-side from the upgrade request's User-Agent (§9). No stored name + no `name` ⇒ hub replies `need-name` (client shows S1). Host token ⇒ `IsHost`, name from hub hostname (FR-2.7) |
| `set-name` | `{name}` | S1 Join and header rename (FR-2.5); hub dedups (FR-2.6) |
| `offer` | `{to: deviceId, files: [{name, size}]}` | validated per V-3/V-5/V-6; hub replies `offer-created` or `error` |
| `offer-cancel` | `{transferId}` | sender withdraws pending offer (FR-4.4) |
| `accept` | `{transferId}` | receiver accepts (FR-5.2) |
| `decline` | `{transferId}` | receiver declines (FR-5.2) |
| `transfer-cancel` | `{transferId}` | either side, mid-stream (FR-4.5/5.5) |
| `pick-interface` | `{interfaceId}` | host page only, answers M3 (FR-1.6) |

**Hub → Client**

| type | payload | serves |
|---|---|---|
| `welcome` | `{deviceId, self: Device, isHost}` | join handshake; client persists `deviceId` |
| `need-name` | `{suggested}` | S1 with pre-fill (FR-2.3; suggestion derived hub-side from User-Agent) |
| `devices` | `{devices: Device[]}` | full snapshot on join and after every change — S2 grid (FR-3.1–3.3). Client filters out self (V-5) |
| `invite-info` | `{urls: {mdns, ip}, port, reachabilityHint?}` | S3 / S2 empty state content; **both URLs include the bound port** (browsers don't read mDNS SRV records); re-sent when interface changes; `reachabilityHint` set when self-probe fails (FR-7.3) |
| `interface-choices` | `{choices: [{id, kind, address}], preselected}` | M3 (FR-1.6) |
| `offer-created` | `{transfer: Transfer}` | sender's "waiting for accept" card state (FR-4.3) |
| `offer` | `{transfer: Transfer, from: Device}` | receiver's M1 (FR-5.1) |
| `offer-cancelled` | `{transferId, reason: "sender-cancelled" \| "sender-disconnected"}` | closes M1 (FR-4.4, E-3) |
| `transfer-accepted` | `{transferId}` | sender: begin sequential uploads via 4.1 POST |
| `transfer-declined` | `{transferId}` | sender toast (E-6) |
| `file-ready` | `{transferId, index}` | receiver: trigger anchor download `GET …/files/{index}` — sent per file when the sender's upload for that index opens |
| `progress` | `{transferId, index, sent, size, totalSent, totalSize}` | M2 both sides (FR-6.2); emitted by relay every 500 ms or 1 % delta, whichever first |
| `transfer-done` | `{transferId}` | B2 success both sides (FR-6.4) |
| `transfer-failed` | `{transferId, reason: "sender-disconnected" \| "receiver-disconnected" \| "cancelled-by-sender" \| "cancelled-by-receiver" \| "stream-error"}` | B2 failure with cause category (FR-6.5, E-4/5/7/8) |
| `error` | `{code: "target-gone" \| "bad-request", message}` | e.g. E-11 immediate offer failure |

**Presence semantics:** device liveness == WS liveness. WS ping/pong every 15 s, dead after 2 misses; close ⇒ `devices` fanout without the device (FR-3.3, AC-10's ~2 s window) and E-3/4/5 handling for its transfers.

**Multi-file sequencing:** on `transfer-accepted`, sender uploads file 0; hub emits `file-ready(0)` to receiver; receiver's anchor download attaches; bytes flow; on completion sender proceeds to file 1, etc. `Transfer` reaches `Done` when the last index completes. One in-flight file per transfer — simple, and wifi is the bottleneck anyway (NFR-1).

### 4.3 Kernel-journey traceability (UX.md F1–F4 → contract)

| Flow step | Serving contract |
|---|---|
| F1.1 hub launch → tray, browser, QR page | process start; `GET /` (+`hostToken`); `hello`→`welcome`; `invite-info` renders QR/URLs |
| F1.2 phone scans QR → S1 pre-filled | QR encodes `http://ip:port` (from `invite-info.urls.ip`); `GET /`; `hello`(no name)→`need-name{suggested}` |
| F1.3 Join → both lists update | `set-name` → `welcome` + `devices` fanout to all (AC-3) |
| F1.4 tap Laptop → picker | client-local (`<input type=file multiple>`); no contract needed |
| F1.5 pick 2 GB video → laptop prompt | `offer` → `offer-created` (sender) + `offer` (receiver M1) |
| F1.6 Accept → transfer starts | `accept` → `transfer-accepted` (sender) → `POST /api/transfers/{tid}/files/0` + `file-ready` → receiver `GET …/files/0` |
| F1.7 progress → saved to Downloads | `progress` events both sides; download completes via `Content-Disposition: attachment`; `transfer-done` |
| F2 laptop → phone (drag-drop) | drop event → same `offer`/`accept`/POST/GET path |
| F3 close tab → vanish; rescan → back <5 s | WS close → `devices` fanout; rejoin: `GET /`, `hello{deviceId, name}` (localStorage) → `welcome`+`devices`, S1 skipped |
| F4.1 decline | `decline` → `transfer-declined` |
| F4.2 screen-lock kills upload | POST body stalls → relay read timeout (30 s no bytes) → `transfer-failed{stream-error}` both sides; receiver's download stream aborted (partial discarded by browser) |
| F4.3 receiver cancels | `transfer-cancel` → abort both streams → `transfer-failed{cancelled-by-receiver}` |
| B1 reconnect | `ws.ts` exponential backoff (0.5 s → 8 s cap); re-`hello` with stored id/name (AC-19) |
| M3 interface pick | `interface-choices` → `pick-interface` → fresh `invite-info` to host (FR-1.6, AC-15) |
| AC-12 duplicate names | dedup in `set-name`/`hello` handling → suffixed `Name` in `devices` |
| AC-13 port busy | startup port scan (§8 config); bound port flows into `invite-info` |
| AC-14 `befrest.local` | zeroconf announce at startup |

Every step names its contract; no UX step is unserved.

---

## 5. Component hierarchy (client)

Mapped to UX.md screen ids:

```
App.svelte                     state switch: need-name? → S1 : S2; hosts layers
├── NameScreen.svelte          S1  (need-name → set-name)
├── MainScreen.svelte          S2
│   ├── Header.svelte          S2 region 1 — self name + inline rename (set-name)
│   ├── ConnectionBanner.svelte B1 — driven by ws.ts status store
│   ├── ReceiveBanner.svelte   M2 (receiver) — active incoming transfers
│   ├── DeviceGrid.svelte      S2 region 3
│   │   └── DeviceCard.svelte  tap→file picker · dragover/drop (v1) ·
│   │                          embeds M2 (sender) progress state per card
│   ├── EmptyInvite.svelte     S2 empty state = S3 content inline (QR, URLs)
│   └── Footer.svelte          S2 region 4 — "Add device" → InviteSheet
├── InviteSheet.svelte         S3 — QR (qrcode→SVG), copyable URLs,
│   │                          reachability hint, "Change network" → M3
├── OfferModal.svelte          M1 — one at a time off a FIFO offers store (EC-5)
├── InterfacePickerModal.svelte M3 — host only
└── ToastHost.svelte           B2 — success/failure/join toasts, aria-live
```

Stores (`lib/stores.ts`): `connection` (B1), `self` (S1/S2 header), `devices` (grid), `offers` (M1 queue), `transfers` (M2 states per transfer), `invite` (S3/EmptyInvite/M3), `toasts` (B2). Components subscribe to stores; only `ws.ts` mutates them.

---

## 6. Dependency graph

```
cmd/befrest ──▶ internal/server ──▶ internal/presence ──▶ internal/proto
     │                │────────────▶ internal/transfer ──▶ internal/proto
     │                └────────────▶ web (embed.FS)
     └────────▶ internal/netinfo   (also called by server for M3 re-pick)

External (hub): coder/websocket · fyne.io/systray · grandcat/zeroconf · pkg/browser
External (web): svelte · vite · typescript · qrcode
Client: App ─▶ screens ─▶ stores ◀─ ws.ts ─▶ proto.ts ; upload.ts/download.ts ─▶ stores
```

No cycles: `proto` is the leaf; `presence`/`transfer` never import `server`; UI components never import `ws.ts` directly (stores are the seam).

---

## 7. Error handling strategy

**Principles:** the hub is the single source of truth for transfer outcomes; HTTP stream errors are *symptoms*, the WS `transfer-failed`/`transfer-done` event is the *verdict* both clients render from. Every failure maps to one of the five `transfer-failed.reason` categories — the UI never invents copy beyond the reason→string table.

- **Hub internal:** handlers return errors upward; `server` translates to HTTP status (4.1 table) or WS `error` frames. Panics in per-connection goroutines are recovered, logged, and treated as that connection's disconnect. Structured logging via `log/slog` to stderr (and tray "Open log" reveals the file on disk under the OS temp dir).
- **Relay:** `io.Copy` loop with 30 s per-read/write deadlines. Any error or deadline → pipe torn down, both HTTP streams closed, transfer → `Failed`, WS verdict fanned out, `Sent` counters frozen. Partial receiver data is discarded by the browser when the download stream aborts before `Content-Length` (E-4/5/7/8; AC-9).
- **Disconnect sweep:** on WS close, presence removes the device, fans out `devices`, and asks transfer manager to fail/cancel everything involving that device (E-3/4/5).
- **Client:** `ws.ts` owns reconnect (backoff, resets stores to "connecting" state → B1); upload fetch failures are ignored in favor of the WS verdict; join failure (initial WS/`hello` unreachable) renders S1's inline error (E-1).
- **Validation:** V-1 enforced client-side (UX) *and* hub-side (trim, length clamp); V-3/V-5/V-6 hub-side at `offer` time → `error{target-gone|bad-request}` (E-11).

---

## 8. Configuration strategy

Zero-config by default (SPEC core promise); flags exist for edge cases only. No config file, no env-var sprawl.

| Flag | Default | Purpose |
|---|---|---|
| `--port` | `5311`, scan upward to first free (FR-1.4/E-9) | fixed port for firewall allow-listing |
| `--interface` | auto-rank (private IPv4, physical > virtual, wifi/ethernet > tun/bridge); ambiguity ⇒ M3 prompt | pin the advertised interface (skips M3) |
| `--name` | OS hostname | host device display name |
| `--no-open` | off | suppress auto-opening the browser (headless/CI) |
| `--no-mdns` | off | disable zeroconf announce |

Precedence: flag > auto-detection. `BEFREST_PORT` env var is honored as equivalent to `--port` (container friendliness); no other env vars. All defaults are compiled in — the binary remains dependency- and file-free (NFR-4). Frontend has no build-time configuration; it derives everything (host, port) from `window.location` and `invite-info`.

---

## 9. Cross-cutting commitments

- **Interface ranking heuristic** (`netinfo`): candidate = up, non-loopback, has private IPv4. Score: physical NIC name patterns (`en*`, `eth*`, `wl*`) > virtual (`tun*`, `tap*`, `docker*`, `br-*`, `utun*`); default-route interface gets priority. One clear winner ⇒ use it; else ⇒ `interface-choices` to host page (M3, AC-15).
- **Reachability self-probe** (FR-7.3): after bind, hub dials `advertisedIP:port` from another local socket; failure ⇒ `invite-info.reachabilityHint` with the port number.
- **Name suggestion & device kind**: hub-side User-Agent parsing table (model names for Android, "iPhone"/"iPad", OS names for desktop) yields both the `need-name` suggestion and `Device.Kind` — keeps the logic in one place and out of the bundle.
- **Binary size** (NFR-4): Svelte output is the only embedded asset; build with `-ldflags "-s -w"`; expected well under 30 MB.
- **A11y** (NFR-8): toasts and banners are `aria-live=polite`; M1 is a focus-trapped `role=dialog`; all interactive targets ≥ 56 px; state changes always textual (PRD NFR-8).
