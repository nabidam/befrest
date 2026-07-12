---
status: gate-passed
---

# Befrest — PLAN

Implementation plan for v1. Ordered chunks; each chunk ≤ ~300 lines of new code, has falsifiable acceptance criteria, and names what NOT to do. References: PRD (FR/AC/E/EC/V/NFR), ARCHITECTURE (§), UX (S/M/B/F), DESIGN, CONVENTIONS.

The first milestone is the **walking skeleton**: the thinnest end-to-end slice that makes the kernel journey (SPEC §2) pass in the real app. Ugly is fine; fake is not — real hub, real browsers, real bytes.

---

## Milestone 1 — Walking skeleton

### Chunk 1 — Scaffold: binary serves embedded SPA

**Files:** `go.mod`, `Makefile`, `.gitignore`, `cmd/befrest/main.go`, `internal/server/server.go`, `internal/server/server_test.go`, `web/embed.go`, `web/package.json`, `web/vite.config.ts`, `web/tsconfig.json`, `web/svelte.config.js`, `web/index.html`, `web/src/main.ts`, `web/src/App.svelte`, `web/src/styles/base.css`

**Requirements:**
- Go 1.23 module; Svelte 5 + TS + Vite scaffold in `web/`; `web/embed.go` embeds `web/dist` (`//go:embed all:dist`), `internal/server` serves it on one `http.ServeMux` (ARCH §1, §2).
- `main.go`: bind port — default `5311`, scan upward to first free (FR-1.4); honor `--port` flag and `BEFREST_PORT` env (ARCH §8); log bound URL via `slog` to stderr.
- `Makefile` targets: `web` (vite build), `build` (web + `go build -ldflags "-s -w"`), `test`, `dev`.
- App.svelte renders a placeholder heading — enough to prove embed works.

**Acceptance (falsifiable):**
- `make build` yields a single binary; running it and opening `http://localhost:5311` shows the placeholder page.
- With 5311 occupied (`nc -l 5311`), the binary binds 5312 and logs it (AC-13 groundwork).
- `go test ./...` passes with a `httptest` check that `/` returns the SPA HTML.

**NOT:** no WebSocket, no tray, no mDNS, no styling beyond base reset, no CI config.

---

### Chunk 2 — Presence: join, name, live device list

**Files:** `internal/proto/messages.go`, `internal/presence/registry.go`, `internal/presence/registry_test.go`, `internal/server/ws.go`, `internal/server/ws_test.go`, `web/src/lib/proto.ts`, `web/src/lib/ws.ts`, `web/src/lib/stores.ts`, `web/src/App.svelte`, `web/src/screens/NameScreen.svelte`, `web/src/screens/MainScreen.svelte`, `web/src/screens/DeviceGrid.svelte`, `web/src/screens/DeviceCard.svelte`

**Requirements:**
- `internal/proto`: `hello`, `set-name`, `welcome`, `need-name`, `devices`, `error` message types exactly per ARCH §4.2. `web/src/lib/proto.ts` mirrors them (CONVENTIONS: same-commit rule).
- `GET /ws` (coder/websocket): `hello` handshake; hub derives name suggestion and `Kind` from the upgrade request's User-Agent (ARCH §9); no stored name → `need-name{suggested}`; `set-name` validates V-1 (trim, 1–32, clamp) and dedup-suffixes (FR-2.6, V-2); `welcome` + full `devices` fanout on every change (FR-3.1–3.3).
- Registry per ARCH §3: `devices` + `namesInUse` maps, mutex-guarded; device exists ⟺ WS open; WS close → fanout without the device.
- Client: `ws.ts` connects, sends `hello` with `localStorage` `befrest.deviceId`/`befrest.name` (FR-2.4); stores per ARCH §5 (`connection`, `self`, `devices`); App switches S1 ↔ S2; S1 with pre-filled name, Join disabled when empty (V-1); S2 renders unstyled device cards filtering self (V-5).

**Acceptance:**
- Two browser tabs: each sees exactly the other, not itself (FR-3.1, V-5).
- Closing a tab removes it from the other's list within ~2 s (AC-10 first half).
- Reload skips S1 and rejoins under the same name (FR-2.4).
- Two clients requesting "Pixel 8" → second listed as "Pixel 8 (2)" (AC-12).
- `presence` package has zero imports of `net/http` or websocket (ARCH §2 boundary; enforced by a compile check in test).

**NOT:** no reconnect/backoff, no ping/pong sweep, no header rename, no host token, no styling.

### 🔶 DEMO GATE A — live presence
Walk: open two browsers on two devices (manually typed `IP:port`), join both, watch each appear; close one tab, watch it vanish; rejoin, S1 skipped. **Observe:** list updates with no refresh on either side; suffixed name on a duplicate.

---

### Chunk 3 — Transfer engine (hub side)

**Files:** `internal/transfer/manager.go`, `internal/transfer/pipe.go`, `internal/transfer/manager_test.go`, `internal/transfer/pipe_test.go`, `internal/proto/messages.go` (extend), `internal/server/ws.go` (extend), `internal/server/files.go`, `internal/server/files_test.go`

**Requirements:**
- `Transfer`/`FileMeta`/`TransferState` exactly per ARCH §3, uuidv4 ids; state machine Offered → Accepted → Streaming → Done / Declined.
- WS: `offer` (validated V-3/V-5/V-6; unknown or disconnected target → `error{target-gone}`, E-11), `offer-created`, `offer` push to receiver, `accept` → `transfer-accepted`, `decline` → `transfer-declined`.
- HTTP per ARCH §4.1: `POST /api/transfers/{tid}/files/{idx}` (Content-Length required and must equal declared size → `411`/`400`; wrong state → `409`; one-shot per index) and `GET .../files/{idx}` (`Content-Length`, `Content-Disposition: attachment`).
- Bounded pipe: 4 MiB rendezvous buffer between upload reader and download writer; backpressure via blocking (NFR-2). `file-ready{tid, index}` to receiver when the upload for that index opens.
- `progress` events every 500 ms or 1 % delta, whichever first (ARCH §4.2); `transfer-done` when the last index completes; terminal records dropped after 60 s (ARCH §3).

**Acceptance:**
- `httptest` test relays a 100 MiB random stream: received bytes hash-equal sent bytes; peak `runtime.MemStats` heap growth during relay < 16 MiB (NFR-2 proxy for AC-7).
- Upload with wrong Content-Length → `400`; POST before accept → `409`; unknown tid → `404`.
- Progress events observed in test at ≤ 500 ms intervals; final event has `sent == size`.
- Zero-byte file completes with `transfer-done` (EC-9).
- `transfer` package imports neither `net/http` handlers' types beyond `io` interfaces nor websocket (ARCH §2: handlers hand `io.Reader`/`Writer` in).

**NOT:** no cancel paths, no stall deadlines, no disconnect sweep for transfers, no multi-file client sequencing (hub already supports N indexes), no disk spooling — blocking relay only (ARCH §4.1 commitment).

---

### Chunk 4 — Transfer client: pick, accept, stream, progress

**Files:** `web/src/lib/upload.ts`, `web/src/lib/download.ts`, `web/src/lib/format.ts`, `web/src/lib/stores.ts` (extend: `offers`, `transfers`), `web/src/screens/DeviceCard.svelte` (extend), `web/src/screens/OfferModal.svelte`, `web/src/screens/ReceiveBanner.svelte`, `web/src/screens/ToastHost.svelte`, `web/src/lib/format.test.ts`

**Requirements:**
- Tap a device card → hidden `<input type="file" multiple>` opens picker (FR-4.1); on confirm send `offer`; card shows "Waiting for X to accept…" (FR-4.3).
- OfferModal (M1): sender name, per-file name + size (format.ts renders human sizes), Accept/Decline; focus-trapped `role=dialog`; not dismissible by outside tap (UX M1); one at a time off a FIFO `offers` store (EC-5).
- On `transfer-accepted`: `upload.ts` POSTs file 0's `File` blob (browser sets Content-Length); sequential per ARCH §4.2 multi-file rule. On `file-ready`: `download.ts` clicks a temporary anchor to `GET .../files/{idx}` (FR-5.4).
- `progress` events drive sender card bar text and receiver banner text (FR-6.2); `transfer-done` → B2 toast "Sent ✓"/"Saved to Downloads ✓" (FR-6.4); `transfer-declined` → sender toast (E-6). ToastHost is `aria-live=polite`.

**Acceptance:**
- Phone → laptop, single 2 GB file over real wifi: prompt names sender/file/size before any bytes move (AC-5); accept → progress advances on both sides; file lands in Downloads, checksum-identical (AC-6).
- Hub process RSS stays bounded (< ~100 MB) throughout, observed in a process monitor (AC-7).
- Decline → prompt closes, sender sees "declined" toast, zero bytes in receiver downloads (AC-8).
- Laptop → phone send works symmetrically (AC-11).
- `format.ts` unit tests: 0 B, 1.3 GB, 2.0 GB renderings.

**NOT:** no cancel buttons yet, no drag-and-drop, no multi-file UI ("file N of M"), no styling beyond structure, no keep-screen-on hint.

---

### Chunk 5 — Front door: QR, invite, tray, mDNS, host auto-join

**Files:** `cmd/befrest/main.go` (extend), `cmd/befrest/tray.go`, `internal/netinfo/netinfo.go`, `internal/netinfo/netinfo_test.go`, `internal/server/ws.go` (extend: hostToken, invite-info), `internal/proto/messages.go` (extend), `web/src/screens/EmptyInvite.svelte`, `web/src/screens/InviteSheet.svelte`, `web/src/screens/Footer.svelte`, `web/src/screens/Header.svelte`, `web/package.json` (add `qrcode`)

**Requirements:**
- `netinfo`: enumerate up, non-loopback, private-IPv4 interfaces; rank per ARCH §9 (physical > virtual, default-route priority); pick the single clear winner (ambiguity handling deferred to Chunk 10 — until then, take the top-ranked).
- Startup: mint one-time `hostToken`; open browser at `http://<ip>:<port>/?hostToken=…` via pkg/browser unless `--no-open`; `hello` with valid hostToken → `IsHost=true`, name from OS hostname, S1 skipped (FR-2.7). Tray (fyne.io/systray): Open befrest, Open log, Quit (FR-1.3); Quit stops the server (FR-1.7).
- mDNS: zeroconf announce so `befrest.local` resolves to the advertised IP (FR-1.5) unless `--no-mdns`.
- `invite-info{urls:{mdns,ip}, port}` sent after `welcome`; both URLs include the bound port (`http://befrest.local:5311`-style — browsers don't read SRV records, so the port is always explicit). EmptyInvite = S2 empty state (FR-3.4): QR (client-side `qrcode` → SVG, encodes the ip URL, FR-2.2), both URLs with tap-to-copy. Footer "Add device" opens InviteSheet (S3, same content, FR-3.5). Header shows self name (static).

**Acceptance:**
- On a fresh machine: double-click the binary → tray icon visible, browser opens, page shows QR + `http://befrest.local:<port>` + `http://<ip>:<port>`, already joined as the hostname with no S1 (AC-1, FR-2.7).
- Phone camera scan of the QR opens the page on the phone (AC-2).
- `http://befrest.local:<port>` typed on another LAN device loads the app (AC-14).
- Tray Quit exits the process; port is released.
- Second hub instance binds the next port and its QR decodes to that actual URL (AC-13).

**NOT:** no M3 interface picker, no reachability probe, no `--interface`/`--name` flags, no "Change network" link, no styling pass.

---

### Chunk 6 — Resilience: heartbeat, reconnect, rename

**Files:** `internal/server/ws.go` (extend), `internal/presence/registry.go` (extend), `web/src/lib/ws.ts` (extend), `web/src/screens/ConnectionBanner.svelte`, `web/src/screens/Header.svelte` (extend), `web/src/screens/MainScreen.svelte` (extend)

**Requirements:**
- WS ping/pong every 15 s, dead after 2 misses → treated as close (ARCH §4.2 presence semantics).
- Client reconnect: exponential backoff 0.5 s → 8 s cap; on reconnect re-`hello` with stored id/name; stores reset to "connecting" (ARCH §7). B1 banner: "Connection lost — reconnecting…", cards dimmed and non-tappable; after 30 s copy extends per UX B1; "Reconnected" flash on success (FR-7.1/7.2, E-2).
- Header ✎ inline rename → `set-name`; propagates via `devices` fanout within 2 s; dedup applies on rename (FR-2.5, EC-3).
- S2 loading state: 2 skeleton cards between page load and first `devices` message, max ~2 s before B1 logic (UX S2).

**Acceptance:**
- Kill the hub with a client open: banner appears, cards untappable. Restart hub: banner clears and the grid repopulates with no manual reload (AC-19).
- Rename on one device appears on another within 2 s (AC-20).
- Renaming to an in-use name gets suffixed, not rejected (V-2).

**NOT:** no transfer-failure handling on disconnect (next chunk), no offline queueing, no styling.

---

### Chunk 7 — Failure paths: cancel, disconnect, stall

**Files:** `internal/transfer/manager.go` (extend), `internal/transfer/pipe.go` (extend: deadlines), `internal/server/ws.go` (extend), `internal/proto/messages.go` (extend), `web/src/lib/stores.ts` (extend), `web/src/screens/OfferModal.svelte` (extend), `web/src/screens/DeviceCard.svelte` (extend: cancel), `web/src/screens/ReceiveBanner.svelte` (extend: cancel), `web/src/screens/ToastHost.svelte` (extend), `internal/transfer/manager_test.go` (extend)

**Requirements:**
- `offer-cancel` (sender withdraws pending; receiver M1 closes with "cancelled" notice — FR-4.4, E-3) and `transfer-cancel` (either side mid-stream — FR-4.5/5.5).
- Failure taxonomy: exactly the five `transfer-failed.reason` values of ARCH §4.2; client renders from a single reason→copy table (ARCH §7); "cancelled" copy distinct from "failed" (E-7).
- Relay deadlines: 30 s per-read/write; stall → teardown, both HTTP streams closed, `transfer-failed{stream-error}` fanned out, `Sent` frozen (ARCH §7, F4.2).
- Disconnect sweep: WS close fails/cancels all transfers involving that device with the right reason (E-3/E-4/E-5); `Offered` transfers destroyed on either party's disconnect (ARCH §3).
- Cancel affordances visible for the whole transfer on both sides (FR-6.3): sender in-card, receiver in-banner.
- Hub is verdict source: upload fetch errors ignored in favor of the WS event (ARCH §7).

**Acceptance:**
- Kill the receiver's tab mid-2 GB transfer: sender gets a failure notice within a few seconds; no partial file remains in downloads (AC-9 — the aborted download's partial is discarded by the browser).
- Receiver cancels mid-stream: sender toast says cancelled (not failed); UI resets both sides (E-7, F4.3).
- Sender cancels a pending offer: receiver's M1 closes itself with a notice (FR-4.4).
- Stalled upload (test harness stops sending bytes): both sides get `stream-error` verdict at ~30 s (F4.2).
- Two simultaneous offers to one receiver: prompts answered one at a time, neither dropped (EC-5).

**NOT:** no resume, no retry logic, no partial-file salvage, no reason strings invented outside the copy table.

### 🔶 DEMO GATE B — kernel journey + failure drill
Walk SPEC §2 kernel journey end-to-end on real hardware: laptop double-click → tray + QR page; phone scans → joins as "Pixel 8"-alike; phone sends 2 GB video → laptop accept → progress both sides → checksum-identical file in Downloads; laptop sends a photo back; phone closes tab → vanishes ≤ 2 s; rescans → back < 5 s with no S1. Then the failure drill: decline, sender offer-cancel, receiver mid-stream cancel, receiver tab-kill mid-transfer. **Observe:** every AC-1…AC-15 currently in scope (AC-15 deferred to Gate C), hub memory flat during the 2 GB relay, correct distinct copy for declined vs cancelled vs failed.

---

## Milestone 2 — v1 features

### Chunk 8 — Multi-file transfers

**Files:** `web/src/lib/upload.ts` (extend), `web/src/screens/OfferModal.svelte` (extend), `web/src/screens/DeviceCard.svelte` (extend), `web/src/screens/ReceiveBanner.svelte` (extend), `internal/transfer/manager_test.go` (extend), `web/src/lib/format.ts` (extend)

**Requirements:**
- Sequential upload of N files per ARCH §4.2: file i completes → POST file i+1; `file-ready` per index triggers each receiver download; `Done` after last index.
- M1 lists up to 4 files with per-file sizes; > 4 → first files + "and N more" summary with total size (FR-5.1, UX M1).
- Progress UI: current filename, "file N of M", aggregate bar from `totalSent/totalSize` (FR-6.2).

**Acceptance:**
- Send 5 files: receiver prompt shows the summary form; all 5 land in downloads; progress showed "file N of 5" and an aggregate bar (AC-16).
- Hub test: 3-file transfer where file 2 fails mid-stream → whole transfer `Failed`, no `file-ready` for file 3.

**NOT:** no parallel file streams (one in-flight file per transfer — ARCH §4.2), no zip bundling, no folder support.

---

### Chunk 9 — Drag-and-drop, mobile hint, concurrent transfers

**Files:** `web/src/screens/DeviceCard.svelte` (extend), `web/src/lib/stores.ts` (extend), `web/src/screens/MainScreen.svelte` (extend)

**Requirements:**
- Desktop: `dragover` on a card → highlight "Drop to send to X"; drop → same offer flow as the picker (FR-4.2, F2).
- Mobile senders: from upload start until finish, persistent inline hint "Keep this screen on until sending finishes." under the card progress (FR-6.6; mobile = the client's own `Kind`).
- Verify independent state per transfer id so a device can send while receiving with separate progress surfaces (FR-4.6 — engine already supports it; this chunk makes the UI state per-transfer, not global).

**Acceptance:**
- Dragging files from a file manager over a card highlights it; dropping starts the offer flow (AC-17).
- Phone upload shows the keep-screen-on hint; locking the screen mid-upload → laptop failure notice within the 30 s stall deadline (AC-18).
- While receiving, tapping another card and sending works; two progress surfaces update independently (AC-21).

**NOT:** no drop-anywhere zone (cards only), no wake-lock API usage (hint only, per C-6), no upload pause/resume.

---

### Chunk 10 — Multi-NIC, reachability, flags

**Files:** `internal/netinfo/netinfo.go` (extend), `internal/netinfo/probe.go`, `internal/netinfo/netinfo_test.go` (extend), `cmd/befrest/main.go` (extend), `internal/server/ws.go` (extend), `internal/proto/messages.go` (extend), `web/src/screens/InterfacePickerModal.svelte`, `web/src/screens/InviteSheet.svelte` (extend)

**Requirements:**
- Ambiguity rule per ARCH §9: no single clear winner → `interface-choices{choices, preselected}` to the host page → M3 modal (host only); `pick-interface` → fresh `invite-info` to everyone, QR/URLs update immediately (FR-1.6); mDNS re-announce on change. S3 "Wrong network? Change network" link (host page only) reopens M3.
- Reachability self-probe after bind: dial `advertisedIP:port`; failure → `invite-info.reachabilityHint`; S3 warn hint naming the port (FR-7.3, E-10).
- Flags: `--interface` (pins, skips M3), `--name` (host display name), completing ARCH §8's table; precedence flag > auto.

**Acceptance:**
- Host with an active VPN/tun interface: either the QR already encodes the real LAN address or M3 appears; after choosing, the QR encodes the chosen interface (AC-15).
- `--interface` set → M3 never appears even in ambiguity.
- Probe failure (simulated by firewall rule) → hint under the QR names the port (E-10).
- `netinfo` unit tests cover ranking: `wl*` beats `docker*`, default-route interface wins ties.

**NOT:** no IPv6-only support, no continuous interface watching (startup + M3 re-pick only — ARCH §2), no persistence of the choice.

### 🔶 DEMO GATE C — v1 acceptance sweep
Walk AC-12 through AC-21 as a checklist on real devices (two phones + laptop, one phone on VPN-free path). **Observe:** every v1 AC passes; every E-1…E-12 behavior spot-checked; no console errors during the full sweep.

---

## Milestone 3 — Design & ship

### Chunk 11 — Design system: tokens + S1/S2 surfaces

**Files:** `web/src/styles/tokens.css`, `web/src/styles/fonts.css`, `web/public/fonts/*.woff2` (Instrument Sans + JetBrains Mono, latin subsets), `web/src/styles/base.css` (extend), `web/src/screens/NameScreen.svelte` (style), `web/src/screens/MainScreen.svelte` (style), `web/src/screens/Header.svelte` (style), `web/src/screens/DeviceGrid.svelte` (style), `web/src/screens/DeviceCard.svelte` (style), `web/src/screens/EmptyInvite.svelte` (style), `web/src/screens/Footer.svelte` (style)

**Requirements:**
- `tokens.css` implements DESIGN.md's token table verbatim (dark-only per DESIGN_SYSTEM.md — no `prefers-color-scheme` branching); `fonts.css` declares self-hosted `@font-face` for Instrument Sans + JetBrains Mono with `font-display: swap` and system fallbacks; components consume tokens only — zero raw hex/px/font values in component styles (DESIGN hard rule).
- S1, S2, header, grid, cards, empty-invite, footer styled to DESIGN component-state specs: hover/focus-visible/active/disabled on every interactive element; skeleton loading; join pulse signature on card entry.
- Touch targets ≥ `--size-touch`; card min-height `--size-card-min`; layout per DESIGN (single column, `--size-content-max`).

**Acceptance:**
- `grep -rE '#[0-9a-fA-F]{3,8}|[0-9]+px' web/src/screens/` returns no matches (values live only in tokens.css).
- Keyboard-only pass: every interactive element on S1/S2 reachable with a visible focus ring.
- No unreadable pairings in the dark palette (spot-check text/bg with a contrast checker ≥ 4.5:1); fonts load from the hub itself — DevTools network tab shows zero external requests.
- A joining device's card animates the pulse once; with `prefers-reduced-motion`, it doesn't.

**NOT:** no CSS framework, no in-app theme toggle (C-5), no layout restructuring of UX.md's regions, no styling of layers (next chunk).

---

### Chunk 12 — Design system: layers + a11y audit

**Files:** `web/src/screens/InviteSheet.svelte` (style), `web/src/screens/OfferModal.svelte` (style), `web/src/screens/InterfacePickerModal.svelte` (style), `web/src/screens/ConnectionBanner.svelte` (style), `web/src/screens/ReceiveBanner.svelte` (style), `web/src/screens/ToastHost.svelte` (style)

**Requirements:**
- S3 sheet (QR ≥ ~50 % of sheet per UX §5), M1 modal (Accept filled primary, Decline same-size secondary — UX M1), M3, B1 banner, receive banner, toasts — all to DESIGN component-state specs.
- A11y audit vs NFR-8: M1/M3 focus-trapped `role=dialog` with Escape≡Decline/close; toasts + banners `aria-live=polite`; all state changes textual, never color-only; contrast AA on every pairing incl. the warn hint block; QR renders on its light tile and scans reliably against the dark theme.

**Acceptance:**
- axe-core scan on S1, S2 (populated + empty), S3, M1 open, B1 shown: zero critical violations.
- Tab order in M1: file info → Decline → Accept; focus returns to the grid on close.
- The raw-hex/px grep from Chunk 11 still returns no matches across all of `web/src/screens/`.

**NOT:** no new UX surfaces or copy beyond UX.md + the reason→copy table, no animation beyond DESIGN's motion tokens.

---

### Chunk 13 — E2E suite + packaging

**Files:** `e2e/package.json`, `e2e/playwright.config.ts`, `e2e/kernel.spec.ts`, `e2e/failures.spec.ts`, `Makefile` (extend: `release`, `e2e`), `README.md`

**Requirements:**
- Playwright against a locally launched hub (`--no-open --no-mdns`, fixed port): kernel spec (join, presence, offer/accept, small-file transfer with content assertion, decline, leave/rejoin) and failures spec (offer-cancel, mid-transfer cancel, disconnect) using two browser contexts.
- `make release`: cross-compile `GOOS` windows/darwin/linux × amd64/arm64 with `-ldflags "-s -w"`; fail the target if any binary ≥ 30 MB (NFR-4).
- README: download → double-click → scan quickstart; flags table from ARCH §8.

**Acceptance:**
- `make e2e` green locally from a clean checkout.
- `make release` emits 6 binaries, all < 30 MB, each starts with `--no-open` and serves the SPA.
- Vitest + Go tests + e2e all pass in one `make test e2e` run.

**NOT:** no auto-update, no code signing, no installers, no CI pipeline definition (repo-local tooling only for v1).

### 🔶 DEMO GATE D — v1 exit
Walk the full kernel journey on each OS binary (Windows, macOS, Linux hosts) with a real phone, then the v1 extras: 5-file drag-and-drop from a desktop, rename mid-session, hub restart recovery. **Observe:** SPEC §2 journey passes on all three platforms; visual result matches DESIGN.md direction (side-by-side with the token table); all 21 ACs checked off; binary sizes recorded.

---

## Out of plan

Everything in PRD §9/§10 (backlog). No chunk may pull backlog items forward without a PLAN revision.
