---
status: ready
---

# Befrest — TASKS

Phase 4 breakdown of `PLAN.md` (status: gate-passed). Each task fits one implementation prompt (~50–300 new lines), has observable acceptance criteria, an interfaces block quoted from ARCHITECTURE.md, and a context pack.

Context packs are **hints predicted before code exists** — the implementation session verifies against real files. Interfaces blocks are **firmer**: they quote the contract; changing a contract routes through ARCHITECTURE.md/PLAN.md, never through a task improvising. An isolated implementer learns neighboring types from the interfaces block, never by reading neighbor code.

Task order is execution order. Walking-skeleton tasks (T0–T14) may not be reordered after feature tasks. Demo gates are tasks; a skipped gate is marked `GATE SKIPPED`, never deleted.

---

## Milestone 1 — Walking skeleton

### T0 — Scaffold: binary serves embedded SPA

- **Objective:** Repo skeleton per FILE_STRUCTURE.md; one Go binary embeds and serves the Svelte SPA; port scan works. No feature logic.
- **Inputs:** FILE_STRUCTURE.md tree; ARCH §1 (stack table), §2 (package layout), §8 (port config).
- **Outputs:** Buildable binary serving a placeholder page; `make web/build/test/dev` targets; one smoke test.
- **Dependencies:** none.
- **Files (create):** `go.mod`, `Makefile`, `.gitignore`, `cmd/befrest/main.go`, `internal/server/server.go`, `internal/server/server_test.go`, `web/embed.go`, `web/package.json`, `web/vite.config.ts`, `web/tsconfig.json`, `web/svelte.config.js`, `web/index.html`, `web/src/main.ts`, `web/src/App.svelte`, `web/src/styles/base.css`
- **Acceptance:**
  - `make build` yields a single binary; running it and opening `http://localhost:5311` shows the placeholder page.
  - With 5311 occupied (`nc -l 5311`), the binary binds 5312 and logs the bound URL (AC-13 groundwork).
- `go test ./...` passes with an `httptest` check that `GET /` returns the SPA HTML.
- `--port` flag and `BEFREST_PORT` env override the default (ARCH §8).
- **Status:** Done — `339b6da`
- **Verification evidence:** `make build && make test` passed; the built binary served the placeholder at `http://localhost:5311`, and a second instance logged `http://localhost:5312` while 5311 was occupied.
- **Difficulty:** medium (build plumbing, embed wiring).
- **NOT:** no WebSocket, no tray, no mDNS, no styling beyond base reset, no CI config.

**Interfaces**
- CONSUMES: nothing (first task).
- PRODUCES (ARCH §1, §2, §8):
  - `web/embed.go`: `package web` exposing an `embed.FS` of `all:dist` for `internal/server` to serve.
  - `internal/server`: constructor taking the embed FS, returning an `http.Handler` on one `http.ServeMux`; `GET /` and static assets → `200` HTML/JS/CSS (ARCH §4.1 row 1).
  - Port contract: "`--port` default `5311`, scan upward to first free (FR-1.4/E-9)"; "`BEFREST_PORT` env var is honored as equivalent to `--port`".
- **Context pack (hints):** FILE_STRUCTURE.md (whole), ARCHITECTURE.md §1 §2 §8, CONVENTIONS.md. Backend+scaffold task: no UX.md, no DESIGN.md.

---

### T1 — Proto types + presence registry (hub)

- **Objective:** Wire message types for presence, and the device registry with dedup and fanout — pure Go, no network code.
- **Inputs:** ARCH §3 Device model, §4.2 presence messages.
- **Outputs:** `internal/proto` presence types; `internal/presence` registry with join/leave/rename and Notifier fanout.
- **Dependencies:** T0.
- **Files (create):** `internal/proto/messages.go`, `internal/presence/registry.go`, `internal/presence/registry_test.go`
- **Acceptance:**
  - Registry unit tests drive: join mints uuidv4 `Device.ID`; name validated per V-1 (trim, 1–32, clamp); duplicate requested name gets ` (2)`-style suffix, `RawName` preserved (FR-2.6, V-2); rename re-dedups (EC-3); leave frees the name for reuse.
  - Every join/leave/rename triggers a full-snapshot fanout through the injected Notifier; test asserts snapshot content excludes the departed device.
  - `internal/presence` has zero imports of `net/http` or any websocket package (compile check in test — ARCH §2 boundary).
- **Difficulty:** low.
- **NOT:** no WS/HTTP code, no transfer types yet, no User-Agent parsing (that is `server`'s job, ARCH §9).

**Interfaces**
- CONSUMES: nothing from other tasks (leaf package, ARCH §6).
- PRODUCES (ARCH §3, §4.2):
  - `type Device struct { ID string; Name string; RawName string; Kind string; IsHost bool; ConnectedAt time.Time }` — "invariant: Name is unique among currently connected devices"; "a Device exists ⟺ its WS is open".
  - Registry ops used by T2: join (with requested name, kind, isHost) → `*Device`; rename → dedup-suffixed name; leave; snapshot for fanout. Fanout via an injected Notifier callback ("`presence` and `transfer` know nothing about HTTP or WebSocket — they operate on interfaces (`Notifier` callback)").
  - `internal/proto` presence frames, JSON `{"type": "...", ...}` exactly per ARCH §4.2: `hello{deviceId?, name?, hostToken?}`, `set-name{name}`, `welcome{deviceId, self, isHost}`, `need-name{suggested}`, `devices{devices: Device[]}`, `error{code: "target-gone"|"bad-request", message}`.
- **Context pack (hints):** ARCHITECTURE.md §2 §3 §4.2 §6, CONVENTIONS.md, PLAN.md Chunk 2. Backend-only: no UX.md, no DESIGN.md.

---

### T2 — WS presence endpoint (hub)

- **Objective:** `GET /ws` upgrade, `hello`/`set-name` handshake, User-Agent-derived suggestion and Kind, `devices` fanout over real sockets.
- **Inputs:** T1's proto types and registry; ARCH §4.2 handshake, §9 UA table.
- **Outputs:** Working presence over WebSocket; server translates wire frames ↔ registry calls.
- **Dependencies:** T1.
- **Files:** create `internal/server/ws.go`, `internal/server/ws_test.go`; modify `internal/server/server.go` (mount `/ws`)
- **Acceptance (via `httptest` + coder/websocket client):**
  - Connect + `hello` with no stored name → `need-name{suggested}` where suggestion derives from the upgrade request's User-Agent (ARCH §9 table: Android model names, "iPhone"/"iPad", desktop OS names); `Kind` set `"mobile"`/`"desktop"`.
  - `set-name` → `welcome{deviceId, self, isHost}` then `devices` snapshot fanned to **all** connected sockets (FR-3.1–3.3).
  - `hello{deviceId, name}` (returning client) skips `need-name`, goes straight to `welcome` (FR-2.4).
  - Second client requesting "Pixel 8" is listed as "Pixel 8 (2)" in everyone's `devices` (AC-12).
  - Closing a socket → `devices` fanout without that device to survivors (FR-3.3).
- **Difficulty:** medium.
- **NOT:** no ping/pong sweep (T11), no reconnect logic (client-side, T11), no hostToken (T9), no transfer messages (T6).

**Interfaces**
- CONSUMES (from T1, per ARCH §3/§4.2): `Device` struct; registry join/rename/leave/snapshot + Notifier; proto frames `hello`/`set-name`/`welcome`/`need-name`/`devices`/`error` with payloads as quoted in T1.
- PRODUCES (ARCH §4.1, §4.2):
  - `GET /ws` — "Upgrade to control WebSocket; plain upgrade — identity travels in the `hello` frame, not query params; `101`; `400` bad upgrade".
  - Live protocol semantics T3's client codes against: "hello → (need-name{suggested} | welcome) ; set-name → welcome + devices fanout on every change; full snapshot on join and after every change — client filters out self (V-5)".
- **Context pack (hints):** ARCHITECTURE.md §2 §4.1 §4.2 §7 (validation para) §9, `internal/proto/messages.go` + registry API from T1, CONVENTIONS.md, PLAN.md Chunk 2. Backend-only: no UX.md, no DESIGN.md.

---

### T3 — Client join flow + live device list (S1/S2 unstyled)

- **Objective:** Browser client joins the hub, persists identity, renders the live device grid. Structure only, no styling.
- **Inputs:** T2's live WS protocol; UX S1/S2 wireframes.
- **Outputs:** Working S1 ↔ S2 client against a real hub.
- **Dependencies:** T2.
- **Files:** create `web/src/lib/proto.ts`, `web/src/lib/ws.ts`, `web/src/lib/stores.ts`, `web/src/screens/NameScreen.svelte`, `web/src/screens/MainScreen.svelte`, `web/src/screens/DeviceGrid.svelte`, `web/src/screens/DeviceCard.svelte`; modify `web/src/App.svelte`
- **Acceptance (running app, two browser tabs):**
  - Fresh tab shows S1 with the UA-derived name pre-filled; empty field disables Join (V-1, UX S1 validation state).
  - Join → S2; each tab sees exactly the other device, never itself (FR-3.1, V-5).
  - Closing one tab removes its card from the other within ~2 s (AC-10 first half).
  - Reload skips S1 and rejoins under the same name via `localStorage` `befrest.deviceId`/`befrest.name` (FR-2.4).
  - Two clients requesting "Pixel 8" → second renders as "Pixel 8 (2)" (AC-12).
- **Difficulty:** medium.
- **NOT:** no reconnect/backoff (T11), no rename ✎ (T11), no transfer UI (T7), no invite/QR (T10), no styling (T19).

**Interfaces**
- CONSUMES (from T2, ARCH §4.2): the wire protocol quoted in T2 — client sends `hello{deviceId?, name?}` as first frame, then `set-name{name}`; receives `welcome{deviceId, self, isHost}` (persist `deviceId`), `need-name{suggested}` (show S1), `devices{devices: Device[]}` (grid; filter self). `Device` JSON shape: `{id, name, rawName, kind: "mobile"|"desktop", isHost, connectedAt}` mirroring ARCH §3.
- PRODUCES (ARCH §2 client layout, §5):
  - `web/src/lib/proto.ts` — TS mirror of `internal/proto` (CONVENTIONS same-commit rule binds them).
  - `web/src/lib/stores.ts` — `connection`, `self`, `devices` stores; "Components subscribe to stores; only `ws.ts` mutates them" (ARCH §5). Later tasks extend this file with `offers`, `transfers`, `invite`, `toasts`.
  - `ws.ts` message bus later tasks hook transfer frames into.
- **Context pack (hints):** ARCHITECTURE.md §2 (client layout) §4.2 §5, UX.md **S1, S2** wireframes + §5 density notes, `internal/proto/messages.go` from T1 (to mirror), CONVENTIONS.md, PLAN.md Chunk 2. UI task — screens S1, S2; DESIGN.md deferred: chunk NOT rule forbids styling until T19.

---

### T4 — 🔶 DEMO GATE A — live presence

- **Objective:** Human walkthrough of live presence on real devices.
- **Journey:** Open two browsers on two devices (manually typed `IP:port`), join both, watch each appear; close one tab, watch it vanish; rejoin, S1 skipped.
- **Observations required:** List updates with no refresh on either side; suffixed name on a duplicate.
- **Dependencies:** T3.
- **Completion artifact:** The human's walkthrough result recorded on this task (screenshots optional). A skipped gate is marked `GATE SKIPPED` here, never deleted.
- **Difficulty:** n/a (human gate).

---

### T5 — Bounded relay pipe (hub)

- **Objective:** The 4 MiB rendezvous pipe: upload reader → bounded buffer → download writer, with backpressure by blocking. Pure `io`, no HTTP.
- **Inputs:** ARCH §4.1 relay paragraph; NFR-2.
- **Outputs:** `transfer.Pipe` primitive T6 builds on.
- **Dependencies:** T0 (module only).
- **Files (create):** `internal/transfer/pipe.go`, `internal/transfer/pipe_test.go`
- **Acceptance (unit tests):**
  - Relaying a 100 MiB random stream through the pipe: received bytes hash-equal sent bytes; peak `runtime.MemStats` heap growth during relay < 16 MiB (NFR-2 proxy for AC-7).
  - A stalled consumer blocks the producer (backpressure test with a slow reader — producer's write does not complete until reader drains).
  - Zero-byte stream completes cleanly (EC-9 groundwork).
  - Closing either end unblocks the other with an error (teardown groundwork for T12).
- **Difficulty:** medium (concurrency correctness).
- **NOT:** no deadlines yet (T12), no progress accounting (T6), no HTTP types — "handlers hand `io.Reader`/`Writer` in" (ARCH §2).

**Interfaces**
- CONSUMES: nothing beyond stdlib.
- PRODUCES (ARCH §4.1): a rendezvous pipe keyed by `(tid, idx)` semantics for T6 — "the handlers join a `transfer.Pipe(tid, idx)` rendezvous; bytes are copied `upload body → bounded pipe (4 MiB ring) → download response`. Backpressure is natural: a slow receiver blocks the pipe, which blocks the upload read, which TCP-throttles the sender. Hub memory per active file = 4 MiB, independent of file size."
- **Context pack (hints):** ARCHITECTURE.md §2 (boundary rules) §4.1 (relay para), CONVENTIONS.md, PLAN.md Chunk 3. Backend-only: no UX.md, no DESIGN.md.

---

### T6 — Transfer manager, WS control, HTTP file endpoints (hub)

- **Objective:** Full hub-side transfer lifecycle: offer/accept/decline over WS, streaming via T5's pipe over HTTP, progress events.
- **Inputs:** T1 registry (target liveness), T2 WS plumbing, T5 pipe.
- **Outputs:** Hub relays real bytes end-to-end with progress.
- **Dependencies:** T2, T5.
- **Files:** create `internal/transfer/manager.go`, `internal/transfer/manager_test.go`, `internal/server/files.go`, `internal/server/files_test.go`; modify `internal/proto/messages.go` (transfer frames), `internal/server/ws.go` (route transfer frames)
- **Acceptance:**
  - `httptest` end-to-end: offer → accept → POST 100 MiB random body → concurrent GET receives hash-identical bytes; `progress` frames observed at ≤ 500 ms intervals; final `progress` has `sent == size`; `transfer-done` fanned to both parties.
  - `offer` to unknown/disconnected target → `error{target-gone}` (E-11); wrong Content-Length → `400`; missing → `411`; POST before accept → `409`; unknown tid → `404` (ARCH §4.1 table).
  - Zero-byte file completes with `transfer-done` (EC-9).
  - Terminal transfer record dropped after 60 s (ARCH §3 lifecycle invariant — testable with injected clock).
  - `internal/transfer` imports neither `net/http` handler types beyond `io` interfaces nor websocket (ARCH §2; compile check in test).
- **Difficulty:** high.
- **NOT:** no cancel paths, no stall deadlines, no disconnect sweep for transfers (all T12), no multi-file client sequencing (hub already supports N indexes), no disk spooling.

**Interfaces**
- CONSUMES: T1 registry (device connected? lookup for V-6); T2 per-device WS send path; T5 pipe as quoted there.
- PRODUCES (ARCH §3, §4.1, §4.2):
  - `type Transfer struct { ID string; SenderID string; ReceiverID string; Files []FileMeta; State TransferState; CreatedAt time.Time }`; `type FileMeta struct { Index int; Name string; Size int64; Sent int64 }`; states "Offered → Accepted → Streaming → Done / ↘ Declined ↘ Failed(reason) / Cancelled(by)".
  - HTTP (ARCH §4.1): `POST /api/transfers/{tid}/files/{idx}` — "raw body, Content-Length required = FileMeta.Size; `200 {"ok":true}` after fully relayed; `404` unknown tid/idx · `409` wrong state or already sent · `411` missing length · `400` length ≠ declared size". `GET /api/transfers/{tid}/files/{idx}` — "`200`, `Content-Type: application/octet-stream`, `Content-Length: size`, `Content-Disposition: attachment; filename="…"`". One-shot per file index.
  - WS frames (ARCH §4.2) for T7's client: C→H `offer{to, files:[{name,size}]}`, `accept{transferId}`, `decline{transferId}`; H→C `offer-created{transfer}`, `offer{transfer, from}`, `transfer-accepted{transferId}`, `transfer-declined{transferId}`, `file-ready{transferId, index}`, `progress{transferId, index, sent, size, totalSent, totalSize}` "every 500 ms or 1 % delta, whichever first", `transfer-done{transferId}`.
- **Context pack (hints):** ARCHITECTURE.md §2 §3 §4.1 §4.2 §7, `internal/transfer/pipe.go` API from T5, `internal/proto/messages.go`, registry API from T1, CONVENTIONS.md, PLAN.md Chunk 3. Backend-only: no UX.md, no DESIGN.md.

---

### T7 — Client transfer: pick, offer, accept, stream, progress (unstyled)

- **Objective:** Full client transfer journey: tap card → picker → offer → M1 accept → upload/download → progress → done toast.
- **Inputs:** T6's WS frames + HTTP endpoints; UX M1/M2/B2.
- **Outputs:** Real file moves phone → laptop in the browser.
- **Dependencies:** T3, T6.
- **Files:** create `web/src/lib/upload.ts`, `web/src/lib/download.ts`, `web/src/lib/format.ts`, `web/src/lib/format.test.ts`, `web/src/screens/OfferModal.svelte`, `web/src/screens/ReceiveBanner.svelte`, `web/src/screens/ToastHost.svelte`; modify `web/src/lib/proto.ts` (mirror transfer frames), `web/src/lib/stores.ts` (add `offers`, `transfers`, `toasts`), `web/src/screens/DeviceCard.svelte`, `web/src/screens/MainScreen.svelte` (mount banner/modal/toasts), `web/src/App.svelte` (layer host)
- **Acceptance (real devices, real wifi):**
  - Phone → laptop, single 2 GB file: M1 names sender/file/size **before any bytes move** (AC-5); accept → progress advances on both sides (sender in-card, receiver banner); file lands in Downloads checksum-identical (AC-6); B2 toasts "Sent ✓"/"Saved to Downloads ✓" (FR-6.4).
  - Hub process RSS stays < ~100 MB throughout, observed in a process monitor (AC-7).
  - Decline → M1 closes, sender sees "declined" toast, zero bytes in receiver downloads (AC-8).
  - Laptop → phone send works symmetrically via tap + picker (AC-11).
  - M1 is focus-trapped `role=dialog`, not dismissible by outside tap; offers queue FIFO, one M1 at a time (EC-5). ToastHost is `aria-live=polite`.
  - `format.ts` Vitest: 0 B, 1.3 GB, 2.0 GB renderings.
- **Difficulty:** high.
- **NOT:** no cancel buttons (T13), no drag-and-drop (T16), no multi-file "N of M" UI (T15), no keep-screen-on hint (T16), no styling (T19/T20).

**Interfaces**
- CONSUMES (from T6, quoted there): WS frames `offer`/`accept`/`decline`, `offer-created`/`offer`/`transfer-accepted`/`transfer-declined`/`file-ready`/`progress`/`transfer-done`; HTTP `POST`/`GET /api/transfers/{tid}/files/{idx}` contracts. Multi-file rule (ARCH §4.2): "on `transfer-accepted`, sender uploads file 0; hub emits `file-ready(0)`; … One in-flight file per transfer." Browser sets Content-Length from the `File` blob.
- PRODUCES (ARCH §5): `offers` (FIFO, M1 queue), `transfers` (per-id state, M2), `toasts` (B2) stores; `format.ts` human-size renderer — T13 adds the reason→copy table beside it, T15 extends aggregate progress.
- **Context pack (hints):** ARCHITECTURE.md §4.1 §4.2 §5 §7 (client verdict rule), UX.md **M1, M2, B2** (S2 send/receive states) + F1/F2 flows, `web/src/lib/proto.ts` + stores from T3, `internal/proto/messages.go` from T6 (to mirror), CONVENTIONS.md, PLAN.md Chunk 4. UI task — screens M1, M2, B2; DESIGN.md deferred (no styling until T19).

---

### T8 — netinfo: interface enumeration + ranking

- **Objective:** Pick the advertised LAN address: enumerate private-IPv4 interfaces, rank physical > virtual with default-route priority.
- **Inputs:** ARCH §9 ranking heuristic.
- **Outputs:** `netinfo` package returning ranked candidates and the single winner.
- **Dependencies:** T0.
- **Files (create):** `internal/netinfo/netinfo.go`, `internal/netinfo/netinfo_test.go`
- **Acceptance (unit tests over injected interface fixtures):**
  - Candidates = up, non-loopback, private IPv4 only; `wl*` beats `docker*`; default-route interface wins ties (ARCH §9).
  - Single clear winner returned when one exists; ambiguity reported distinctly (consumed in T17 — until then callers take the top-ranked, PLAN Chunk 5).
- **Difficulty:** low.
- **NOT:** no M3/interface-choices (T17), no reachability probe (T17), no mDNS here (announce wired in T9), no continuous watching (ARCH §2: "consulted once at startup plus on M3 re-pick").

**Interfaces**
- CONSUMES: nothing from other tasks.
- PRODUCES (ARCH §9): enumeration returning ranked candidates `{id, kind, address}` (the shape M3's `interface-choices` will carry, ARCH §4.2) plus a clear-winner/ambiguous verdict — "candidate = up, non-loopback, has private IPv4. Score: physical NIC name patterns (`en*`, `eth*`, `wl*`) > virtual (`tun*`, `tap*`, `docker*`, `br-*`, `utun*`); default-route interface gets priority."
- **Context pack (hints):** ARCHITECTURE.md §2 §9, CONVENTIONS.md, PLAN.md Chunk 5. Backend-only: no UX.md, no DESIGN.md.

---

### T9 — Front door hub-side: hostToken, tray, mDNS, browser open, invite-info

- **Objective:** Double-click experience: tray icon, browser auto-open with one-time hostToken, host auto-join, mDNS announce, `invite-info` to every client.
- **Inputs:** T8 winner address; T2 hello path.
- **Outputs:** Hub startup per F1.1; clients receive invite URLs.
- **Dependencies:** T2, T8.
- **Files:** create `cmd/befrest/tray.go`; modify `cmd/befrest/main.go` (token mint, browser open, mDNS, wiring), `internal/server/ws.go` (hostToken handling, `invite-info` after `welcome`), `internal/proto/messages.go` (add `invite-info`)
- **Acceptance:**
  - Fresh machine: double-click binary → tray icon visible (Open befrest / Open log / Quit — FR-1.3); browser opens `http://<ip>:<port>/?hostToken=…`; page joins as OS hostname with **no S1**, `IsHost=true` (FR-2.7; client S1-skip already handles a `welcome` without `need-name`).
  - `http://befrest.local:<port>` typed on another LAN device loads the app (AC-14); `--no-mdns` disables the announce.
  - Tray Quit exits the process and releases the port (FR-1.7). `--no-open` suppresses the browser.
  - Every client receives `invite-info` after `welcome`; both URLs include the bound port; a WS test asserts the frame shape.
  - Second hub instance binds the next port and its `invite-info` carries that actual port (AC-13).
- **Difficulty:** medium.
- **NOT:** no M3/interface-choices, no reachability probe, no `--interface`/`--name` flags (T17); no client UI (T10).

**Interfaces**
- CONSUMES: T8 winner (take top-ranked until T17); T2 hello handshake to extend.
- PRODUCES (ARCH §4.2): `invite-info{urls: {mdns, ip}, port, reachabilityHint?}` — "S3 / S2 empty state content; both URLs include the bound port (browsers don't read mDNS SRV records); re-sent when interface changes". hostToken rule (ARCH §4): "The host page authenticates as host via a one-time `hostToken` query param embedded in the URL the hub opens at launch"; `hello{…, hostToken?}` → `IsHost`, name from hub hostname.
- **Context pack (hints):** ARCHITECTURE.md §1 (stack: systray/zeroconf/browser libs) §4.2 §8 §9, `cmd/befrest/main.go` from T0, ws.go from T2, netinfo API from T8, CONVENTIONS.md, PLAN.md Chunk 5. Backend-only: no UX.md, no DESIGN.md.

---

### T10 — Client invite surfaces: QR, empty state, invite sheet (unstyled)

- **Objective:** Render `invite-info`: QR + copyable URLs as S2 empty state and as the S3 sheet; header self-name; footer Add device.
- **Inputs:** T9's `invite-info` frame; UX S3 + S2 empty state.
- **Outputs:** Scannable onboarding loop closed.
- **Dependencies:** T3, T9.
- **Files:** create `web/src/screens/EmptyInvite.svelte`, `web/src/screens/InviteSheet.svelte`, `web/src/screens/Footer.svelte`, `web/src/screens/Header.svelte`; modify `web/src/lib/proto.ts` (invite-info), `web/src/lib/stores.ts` (add `invite`), `web/src/screens/MainScreen.svelte`, `web/package.json` (add `qrcode`)
- **Acceptance:**
  - Hub page with no other devices: S2 body shows "No other devices yet" + QR + both URLs (S3 content inline, FR-3.4); a phone camera scan of the QR opens the page on the phone (AC-2).
  - QR is client-side `qrcode` → SVG encoding the ip URL (FR-2.2); DevTools network tab shows no external request for it.
  - Footer "Add device" opens InviteSheet with the same content (FR-3.5); tap-to-copy copies each URL.
  - When a second device joins, the empty state collapses into the grid (UX S3 auto-behavior).
  - Header shows self name (static — rename is T11).
- **Difficulty:** low.
- **NOT:** no "Change network" link / M3 (T17), no reachability hint rendering (T17), no rename ✎ (T11), no styling (T19).

**Interfaces**
- CONSUMES (from T9, quoted there): `invite-info{urls: {mdns, ip}, port, reachabilityHint?}` after `welcome`.
- PRODUCES (ARCH §5): `invite` store; `EmptyInvite`/`InviteSheet`/`Footer`/`Header` components in the §5 hierarchy — T11 extends Header (rename), T17 extends InviteSheet (hint + Change network).
- **Context pack (hints):** ARCHITECTURE.md §5, UX.md **S3** wireframe + S2 empty state + §5 density notes, stores/proto from T3, CONVENTIONS.md, PLAN.md Chunk 5. UI task — screens S2(empty), S3; DESIGN.md deferred (no styling until T19).

---

### T11 — Resilience: heartbeat, reconnect, rename

- **Objective:** WS ping/pong liveness hub-side; client auto-reconnect with backoff + B1 banner; inline header rename.
- **Inputs:** T2/T3 presence loop; ARCH §7 client strategy.
- **Outputs:** Hub restarts and network blips self-heal; rename propagates.
- **Dependencies:** T3 (rename also touches Header from T10).
- **Files:** modify `internal/server/ws.go` (ping/pong every 15 s, dead after 2 misses → close semantics), `internal/presence/registry.go` (if sweep hooks needed), `web/src/lib/ws.ts` (backoff 0.5 s → 8 s cap, re-hello, store reset), `web/src/screens/Header.svelte` (✎ rename), `web/src/screens/MainScreen.svelte` (skeleton loading, banner slot); create `web/src/screens/ConnectionBanner.svelte`
- **Acceptance:**
  - Kill the hub with a client open: B1 "Connection lost — reconnecting…" appears, cards dimmed and non-tappable; after 30 s copy extends per UX B1. Restart hub: banner clears with "Reconnected" flash and the grid repopulates with **no manual reload** (AC-19).
  - Reconnect re-`hello`s with stored id/name — same identity, no duplicate card.
  - Rename via header ✎ on one device appears on another within 2 s (AC-20); renaming to an in-use name gets suffixed, not rejected (V-2).
  - S2 shows 2 skeleton cards between page load and first `devices`, max ~2 s before B1 logic (UX S2 loading state).
- **Difficulty:** medium.
- **NOT:** no transfer-failure handling on disconnect (T12), no offline queueing, no styling.

**Interfaces**
- CONSUMES: T2 WS loop; T3 `connection` store + `set-name` frame; T10 Header.
- PRODUCES (ARCH §4.2, §7): "device liveness == WS liveness. WS ping/pong every 15 s, dead after 2 misses; close ⇒ `devices` fanout without the device" — the close semantics T12's disconnect sweep hangs off. Client: "`ws.ts` owns reconnect (backoff, resets stores to 'connecting' state → B1)"; `connection` store states drive B1 and card dimming.
- **Context pack (hints):** ARCHITECTURE.md §4.2 (presence semantics) §7, UX.md **B1** + S2 loading state + S2 header rename, ws.ts/stores from T3, Header from T10, CONVENTIONS.md, PLAN.md Chunk 6. Mixed task — screens B1, S2; DESIGN.md deferred.

---

### T12 — Failure paths hub-side: cancel, disconnect sweep, stall deadlines

- **Objective:** Complete the transfer failure taxonomy on the hub: cancels, disconnect sweep, 30 s relay deadlines, single-verdict fanout.
- **Inputs:** T6 manager + T5 pipe; T11 close semantics; ARCH §7.
- **Outputs:** Every failure maps to exactly one of the five `transfer-failed.reason` values.
- **Dependencies:** T6, T11.
- **Files:** modify `internal/transfer/manager.go`, `internal/transfer/pipe.go` (30 s per-read/write deadlines), `internal/server/ws.go` (route `offer-cancel`/`transfer-cancel`), `internal/proto/messages.go` (add cancel frames, `offer-cancelled`, `transfer-failed`), `internal/transfer/manager_test.go`
- **Acceptance (hub tests):**
  - `offer-cancel` on pending → receiver gets `offer-cancelled{reason: "sender-cancelled"}`; sender disconnect while `Offered` → `offer-cancelled{reason: "sender-disconnected"}` (FR-4.4, E-3); `Offered` transfers destroyed on either party's disconnect (ARCH §3).
  - `transfer-cancel` mid-stream from either side → both HTTP streams closed, `transfer-failed{cancelled-by-sender|cancelled-by-receiver}` fanned to both (FR-4.5/5.5).
  - Stalled upload (test harness stops sending bytes) → teardown at ~30 s, `transfer-failed{stream-error}` both sides, `Sent` frozen (F4.2, ARCH §7).
  - WS close fails/cancels all transfers involving that device with the correct reason per role (E-3/E-4/E-5), driven via `byDevice` index (ARCH §3).
  - Reasons emitted are exactly the five ARCH §4.2 values — a test enumerates them.
- **Difficulty:** high.
- **NOT:** no resume, no retry, no partial-file salvage; no client UI (T13).

**Interfaces**
- CONSUMES: T6 manager/state machine + T5 pipe teardown; T11's "close ⇒ fanout" hook.
- PRODUCES (ARCH §4.2) for T13's client: C→H `offer-cancel{transferId}`, `transfer-cancel{transferId}`; H→C `offer-cancelled{transferId, reason: "sender-cancelled"|"sender-disconnected"}`, `transfer-failed{transferId, reason: "sender-disconnected"|"receiver-disconnected"|"cancelled-by-sender"|"cancelled-by-receiver"|"stream-error"}`. Verdict rule (ARCH §7): "HTTP stream errors are symptoms, the WS `transfer-failed`/`transfer-done` event is the verdict both clients render from."
- **Context pack (hints):** ARCHITECTURE.md §3 (lifecycle invariants) §4.2 §7, manager/pipe from T5/T6, ws.go from T11, CONVENTIONS.md, PLAN.md Chunk 7. Backend-only: no UX.md, no DESIGN.md.

---

### T13 — Failure paths client-side: cancel UI, reason→copy table

- **Objective:** Cancel affordances both sides, offer-cancel handling in M1, and the single reason→copy table for all failure toasts.
- **Inputs:** T12's frames; UX M1/M2/B2 failure copy; E-7 distinct cancelled-vs-failed.
- **Outputs:** Failure drill passes on real devices.
- **Dependencies:** T7, T12.
- **Files:** modify `web/src/lib/proto.ts` (cancel/failed frames), `web/src/lib/stores.ts`, `web/src/lib/format.ts` (reason→copy table), `web/src/screens/OfferModal.svelte` (self-close on `offer-cancelled` + notice), `web/src/screens/DeviceCard.svelte` (sender cancel button, whole-transfer visibility), `web/src/screens/ReceiveBanner.svelte` (receiver cancel), `web/src/screens/ToastHost.svelte`
- **Acceptance (real devices):**
  - Kill the receiver's tab mid-2 GB transfer: sender gets a failure notice within a few seconds; no partial file remains in Downloads (AC-9).
  - Receiver cancels mid-stream: sender toast says cancelled — copy distinct from "failed" (E-7, F4.3); UI resets both sides.
  - Sender cancels a pending offer: receiver's M1 closes itself with a "cancelled" notice (FR-4.4).
  - Cancel affordances visible the whole transfer on both sides: sender in-card, receiver in-banner (FR-6.3).
  - Two simultaneous offers to one receiver: prompts answered one at a time off the FIFO, neither dropped (EC-5).
  - Upload fetch errors are ignored in favor of the WS verdict (ARCH §7) — observable: killing the download mid-stream still yields exactly one failure toast, from the WS event.
- **Difficulty:** medium.
- **NOT:** no reason strings invented outside the copy table; no retry UI; no styling.

**Interfaces**
- CONSUMES (from T12, quoted there): `offer-cancel`/`transfer-cancel` C→H; `offer-cancelled{transferId, reason}` and the five-reason `transfer-failed` H→C; hub-is-verdict rule.
- PRODUCES: the reason→copy table in `format.ts` — "the UI never invents copy beyond the reason→string table" (ARCH §7); T15/T16 reuse it unchanged.
- **Context pack (hints):** ARCHITECTURE.md §4.2 §7, UX.md **M1** (cancel behavior) **M2** (cancel affordances) **B2** (failure copy) + F4, client files from T7, proto frames from T12 (to mirror), CONVENTIONS.md, PLAN.md Chunk 7. UI task — screens M1, M2, B2; DESIGN.md deferred.

---

### T14 — 🔶 DEMO GATE B — kernel journey + failure drill

- **Objective:** Human walkthrough of SPEC §2 kernel journey end-to-end on real hardware.
- **Journey:** Laptop double-click → tray + QR page; phone scans → joins as "Pixel 8"-alike; phone sends 2 GB video → laptop accept → progress both sides → checksum-identical file in Downloads; laptop sends a photo back; phone closes tab → vanishes ≤ 2 s; rescans → back < 5 s with no S1. Then the failure drill: decline, sender offer-cancel, receiver mid-stream cancel, receiver tab-kill mid-transfer.
- **Observations required:** Every AC-1…AC-15 currently in scope (AC-15 deferred to Gate C); hub memory flat during the 2 GB relay; correct distinct copy for declined vs cancelled vs failed.
- **Dependencies:** T13 (and all of T0–T13).
- **Completion artifact:** Human's walkthrough result recorded on this task (screenshots optional). If skipped: mark `GATE SKIPPED`, never delete.
- **Difficulty:** n/a (human gate).

---

## Milestone 2 — v1 features

### T15 — Multi-file transfers

- **Objective:** N-file offers: sequential client upload, per-index downloads, "file N of M" + aggregate progress, summary prompt for > 4 files.
- **Inputs:** T7 upload/download; hub already supports N indexes (T6).
- **Outputs:** 5-file send works end to end.
- **Dependencies:** T13.
- **Files:** modify `web/src/lib/upload.ts` (sequential i → i+1), `web/src/screens/OfferModal.svelte` (list ≤ 4 + "and N more" with total), `web/src/screens/DeviceCard.svelte`, `web/src/screens/ReceiveBanner.svelte` (current filename, "file N of M", aggregate bar), `web/src/lib/format.ts`, `internal/transfer/manager_test.go` (multi-file failure test)
- **Acceptance:**
  - Send 5 files: receiver prompt shows the summary form; all 5 land in Downloads; progress showed "file N of 5" and an aggregate bar from `totalSent/totalSize` (AC-16, FR-6.2).
  - Hub test: 3-file transfer where file 2 fails mid-stream → whole transfer `Failed`, no `file-ready` for file 3.
- **Difficulty:** medium.
- **NOT:** no parallel file streams ("One in-flight file per transfer" — ARCH §4.2), no zip bundling, no folder support.

**Interfaces**
- CONSUMES (ARCH §4.2 multi-file sequencing, quoted in T7): "sender uploads file 0; hub emits `file-ready(0)`; …on completion sender proceeds to file 1; `Transfer` reaches `Done` when the last index completes." `progress.totalSent/totalSize` fields from T6.
- PRODUCES: nothing new on the wire — client-side sequencing + UI only.
- **Context pack (hints):** ARCHITECTURE.md §4.2 (multi-file para), UX.md **M1** (>4 summary) **M2** (multi-file progress), upload.ts/OfferModal/DeviceCard/ReceiveBanner from T7/T13, PLAN.md Chunk 8. UI task — screens M1, M2; DESIGN.md deferred.

---

### T16 — Drag-and-drop, mobile hint, concurrent transfers

- **Objective:** Desktop drop-on-card send; mobile keep-screen-on hint; verified per-transfer UI state.
- **Inputs:** T7/T13 card + transfer stores.
- **Outputs:** AC-17/18/21 pass.
- **Dependencies:** T15.
- **Files:** modify `web/src/screens/DeviceCard.svelte` (dragover highlight + drop → offer flow), `web/src/lib/stores.ts` (per-transfer-id state, not global), `web/src/screens/MainScreen.svelte`
- **Acceptance:**
  - Dragging files from a file manager over a card highlights "Drop to send to X"; dropping starts the same offer flow as the picker (AC-17, FR-4.2).
  - Phone upload shows persistent inline "Keep this screen on until sending finishes." under card progress from upload start to finish (FR-6.6; mobile = client's own `Kind`); locking the screen mid-upload → laptop failure notice within the 30 s stall deadline (AC-18).
  - While receiving, tapping another card and sending works; two progress surfaces update independently (AC-21, FR-4.6).
- **Difficulty:** low.
- **NOT:** no drop-anywhere zone (cards only), no wake-lock API (hint only, C-6), no upload pause/resume.

**Interfaces**
- CONSUMES: T7's offer flow (drop feeds the same `offer` path — ARCH §4.3: "drop event → same `offer`/`accept`/POST/GET path"); `self.kind` from T3's stores; T12's stall verdict.
- PRODUCES: `transfers` store guaranteed keyed per transfer id (FR-4.6) — the shape T19/T20 style against.
- **Context pack (hints):** ARCHITECTURE.md §4.3 (F2 row) §5, UX.md **S2** (drop highlight) **M2** (mobile hint) + F2, DeviceCard/stores from T15, PLAN.md Chunk 9. UI task — screens S2, M2; DESIGN.md deferred.

---

### T17 — Multi-NIC, reachability probe, remaining flags

- **Objective:** Interface ambiguity → M3 picker on the host page; live QR/URL update; reachability self-probe hint; `--interface`/`--name` flags.
- **Inputs:** T8 ranking + ambiguity verdict; T9/T10 invite plumbing.
- **Outputs:** AC-15 passes on a VPN-carrying host.
- **Dependencies:** T9, T10.
- **Files:** create `internal/netinfo/probe.go`, `web/src/screens/InterfacePickerModal.svelte`; modify `internal/netinfo/netinfo.go`, `internal/netinfo/netinfo_test.go`, `cmd/befrest/main.go` (flags), `internal/server/ws.go` (`interface-choices`, `pick-interface`), `internal/proto/messages.go`, `web/src/lib/proto.ts`, `web/src/lib/stores.ts`, `web/src/screens/InviteSheet.svelte` ("Wrong network? Change network" link, hint block)
- **Acceptance:**
  - Host with an active VPN/tun interface: either the QR already encodes the real LAN address or M3 appears with the likely candidate preselected; after choosing, QR/URLs update immediately everywhere and mDNS re-announces (AC-15, FR-1.6).
  - "Change network" link in S3 (host page only) reopens M3.
  - `--interface` pins the interface — M3 never appears even in ambiguity; `--name` sets host display name; precedence flag > auto (ARCH §8).
  - Probe failure (simulated by firewall rule) → hint under the QR naming the port (E-10, FR-7.3).
  - `netinfo` unit tests cover ranking: `wl*` beats `docker*`, default-route wins ties.
- **Difficulty:** medium.
- **NOT:** no IPv6-only support, no continuous interface watching (startup + M3 re-pick only), no persistence of the choice.

**Interfaces**
- CONSUMES: T8 candidates/ambiguity; T9's `invite-info` re-send path; T10's InviteSheet + `invite` store.
- PRODUCES (ARCH §4.2, §9): H→C `interface-choices{choices: [{id, kind, address}], preselected}` (host only); C→H `pick-interface{interfaceId}` → "fresh `invite-info` to everyone". Probe (ARCH §9): "after bind, hub dials `advertisedIP:port` from another local socket; failure ⇒ `invite-info.reachabilityHint` with the port number."
- **Context pack (hints):** ARCHITECTURE.md §4.2 §8 §9, UX.md **M3** wireframe + S3 error state, netinfo from T8, main.go/ws.go from T9, InviteSheet from T10, PLAN.md Chunk 10. Mixed task — screens M3, S3; DESIGN.md deferred.

---

### T18 — 🔶 DEMO GATE C — v1 acceptance sweep

- **Objective:** Human checklist walkthrough of AC-12 through AC-21 on real devices (two phones + laptop, one phone on VPN-free path).
- **Observations required:** Every v1 AC passes; every E-1…E-12 behavior spot-checked; no console errors during the full sweep.
- **Dependencies:** T17 (and all prior).
- **Completion artifact:** Human's checklist result recorded on this task. If skipped: mark `GATE SKIPPED`, never delete.
- **Difficulty:** n/a (human gate).

---

## Milestone 3 — Design & ship

### T19 — Design system: tokens + S1/S2 surfaces

- **Objective:** Token foundation + styled S1, S2, header, grid, cards, empty-invite, footer per DESIGN.md. Dark-only.
- **Inputs:** DESIGN.md token table + component-state specs; DESIGN_SYSTEM.md (dark-only rule).
- **Outputs:** Styled main surfaces; zero raw values in components.
- **Dependencies:** T16 (all S1/S2 behavior in place).
- **Files:** create `web/src/styles/tokens.css`, `web/src/styles/fonts.css`, `web/public/fonts/*.woff2` (Instrument Sans + JetBrains Mono, latin subsets); modify `web/src/styles/base.css` and style `NameScreen`, `MainScreen`, `Header`, `DeviceGrid`, `DeviceCard`, `EmptyInvite`, `Footer`
- **Acceptance:**
  - `grep -rE '#[0-9a-fA-F]{3,8}|[0-9]+px' web/src/screens/` returns no matches — raw values live only in `tokens.css` (DESIGN hard rule).
  - Keyboard-only pass: every interactive element on S1/S2 reachable with a visible focus ring; hover/focus-visible/active/disabled present on every interactive element.
  - No unreadable pairings: spot-checked text/bg contrast ≥ 4.5:1; DevTools network tab shows zero external requests (fonts self-hosted, `font-display: swap`).
  - A joining device's card animates the join pulse once; with `prefers-reduced-motion`, it doesn't.
  - Touch targets ≥ `--size-touch`; card min-height `--size-card-min`; single-column layout capped at `--size-content-max`.
- **Difficulty:** medium.
- **NOT:** no CSS framework, no theme toggle (C-5, dark-only — no `prefers-color-scheme` branching), no layout restructuring of UX.md regions, no styling of layers (T20).

**Interfaces**
- CONSUMES: component structure from T3/T10/T11/T16 (markup exists; this task styles it).
- PRODUCES: `tokens.css` custom properties (DESIGN.md token table verbatim) — T20 consumes the same tokens; components consume tokens only.
- **Context pack (hints):** **DESIGN.md (whole)**, DESIGN_SYSTEM.md, UX.md **S1, S2** + §5 density notes, ARCHITECTURE.md §5, the seven target components, CONVENTIONS.md, PLAN.md Chunk 11. UI task — screens S1, S2.

---

### T20 — Design system: layers + a11y audit

- **Objective:** Style S3/M1/M3/B1/receive-banner/toasts; full NFR-8 a11y audit.
- **Inputs:** T19 tokens; DESIGN component-state specs; UX layer specs.
- **Outputs:** Every surface styled; axe-clean.
- **Dependencies:** T19.
- **Files (style):** `InviteSheet.svelte`, `OfferModal.svelte`, `InterfacePickerModal.svelte`, `ConnectionBanner.svelte`, `ReceiveBanner.svelte`, `ToastHost.svelte`
- **Acceptance:**
  - axe-core scan on S1, S2 (populated + empty), S3, M1 open, B1 shown: zero critical violations (NFR-8).
  - Tab order in M1: file info → Decline → Accept; Escape ≡ Decline; focus returns to the grid on close; M1/M3 focus-trapped `role=dialog`.
  - QR renders on its light tile and scans reliably against the dark theme; S3 QR ≥ ~50 % of sheet (UX §5); M1 Accept filled primary, Decline same-size secondary.
  - Contrast AA on every pairing including the warn hint block; toasts + banners `aria-live=polite`; all state changes textual, never color-only.
  - The raw-hex/px grep from T19 still returns no matches across all of `web/src/screens/`.
- **Difficulty:** medium.
- **NOT:** no new UX surfaces or copy beyond UX.md + the reason→copy table, no animation beyond DESIGN's motion tokens.

**Interfaces**
- CONSUMES: T19's `tokens.css` properties; component structure from T7/T10/T11/T13/T17.
- PRODUCES: nothing new on any contract — styling + a11y attributes only.
- **Context pack (hints):** **DESIGN.md (whole)**, UX.md **S3, M1, M3, B1, B2, M2** + §5, `tokens.css` from T19, the six target components, PLAN.md Chunk 12. UI task — screens S3, M1, M3, B1, B2.

---

### T21 — E2E suite (Playwright)

- **Objective:** Automated kernel + failure specs against a locally launched hub.
- **Inputs:** Built binary (`--no-open --no-mdns`, fixed port); two browser contexts.
- **Outputs:** `make e2e` green from clean checkout.
- **Dependencies:** T20.
- **Files:** create `e2e/package.json`, `e2e/playwright.config.ts`, `e2e/kernel.spec.ts`, `e2e/failures.spec.ts`; modify `Makefile` (add `e2e`)
- **Acceptance:**
  - `kernel.spec.ts` passes: join, presence (two contexts see each other), offer/accept, small-file transfer **with content assertion on the downloaded bytes**, decline, leave/rejoin without S1.
  - `failures.spec.ts` passes: offer-cancel closes receiver's M1; mid-transfer cancel yields "cancelled" copy; context close mid-transfer yields disconnect verdict.
  - `make e2e` green locally from a clean checkout; config launches the built hub itself (no manual server step).
- **Difficulty:** medium.
- **NOT:** no CI pipeline definition, no cross-browser matrix beyond Playwright defaults, no visual regression.

**Interfaces**
- CONSUMES: the full running app — flags contract (ARCH §8: `--no-open`, `--no-mdns`, `--port`); UI surfaces as styled in T19/T20.
- PRODUCES: `make e2e` target consumed by T22's release discipline.
- **Context pack (hints):** ARCHITECTURE.md §8, UX.md F1–F4 flow tables (drive the specs), Makefile from T0, PLAN.md Chunk 13. Test task: no DESIGN.md.

---

### T22 — Packaging + README

- **Objective:** Cross-platform release binaries under the size budget; user-facing quickstart.
- **Inputs:** T0 Makefile; ARCH §8 flags table.
- **Outputs:** 6 release binaries + README.
- **Dependencies:** T21.
- **Files:** modify `Makefile` (add `release`); create `README.md`
- **Acceptance:**
  - `make release` emits 6 binaries (`GOOS` windows/darwin/linux × amd64/arm64, `-ldflags "-s -w"`), all < 30 MB — the target **fails** if any binary ≥ 30 MB (NFR-4).
  - Each binary starts with `--no-open` and serves the SPA (spot-check via curl on the native-arch ones).
  - README covers download → double-click → scan quickstart and the flags table from ARCH §8.
  - Vitest + Go tests + e2e all pass in one `make test e2e` run.
- **Difficulty:** low.
- **NOT:** no auto-update, no code signing, no installers, no CI.

**Interfaces**
- CONSUMES: T0's `build` target; ARCH §8 flags table (README content); T21's `e2e` target.
- PRODUCES: `make release` — the ship artifact Gate D walks.
- **Context pack (hints):** ARCHITECTURE.md §8 §9 (binary-size commitment), Makefile from T21, PLAN.md Chunk 13. Backend-only: no UX.md, no DESIGN.md.

---

### T23 — 🔶 DEMO GATE D — v1 exit

- **Objective:** Human walkthrough of the full kernel journey on each OS binary (Windows, macOS, Linux hosts) with a real phone, then the v1 extras: 5-file drag-and-drop from a desktop, rename mid-session, hub restart recovery.
- **Observations required:** SPEC §2 journey passes on all three platforms; visual result matches DESIGN.md direction (side-by-side with the token table); all 21 ACs checked off; binary sizes recorded.
- **Dependencies:** T22 (and all prior).
- **Completion artifact:** Human's walkthrough result + binary sizes recorded on this task. If skipped: mark `GATE SKIPPED`, never delete.
- **Difficulty:** n/a (human gate).

---

## Dependency graph

```
T0 ─▶ T1 ─▶ T2 ─▶ T3 ─▶ T4 (GATE A)
T0 ─▶ T5 ─┐
     T2 ──┴▶ T6 ─▶ T7 ─┐
T0 ─▶ T8 ─▶ T9 ─▶ T10 ─┤
     T3 ─▶ T11 ────────┤
T6 + T11 ─▶ T12 ─▶ T13 ┴▶ T14 (GATE B)
T13 ─▶ T15 ─▶ T16 ─┐
T9 + T10 ─▶ T17 ───┴▶ T18 (GATE C)
T16 ─▶ T19 ─▶ T20 ─▶ T21 ─▶ T22 ─▶ T23 (GATE D)
```

Out of plan: everything in PRD §9/§10. No task may pull backlog items forward without a PLAN revision.
