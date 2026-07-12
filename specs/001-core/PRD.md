---
status: draft
---

# Befrest — PRD

Product requirements for Befrest v1: LAN file transfer via a single hub binary and browser clients. Derived from `SPEC.md` and `UX.md` (screen ids S1–S3, M1–M3, B1–B2 refer to UX.md).

---

## 1. Functional requirements

### FR-1 Hub lifecycle
- FR-1.1 The hub ships as a single executable per platform (Windows, macOS, Linux) requiring no installer, runtime, or configuration file.
- FR-1.2 Launching the hub shows a system tray icon and opens the default browser on the hub's own client page.
- FR-1.3 The tray menu offers at minimum: open the app page, quit.
- FR-1.4 The hub binds a listening port; if the preferred port is busy it selects the next free port. All displayed URLs and the QR code reflect the port actually bound.
- FR-1.5 The hub announces itself via mDNS as `befrest.local`.
- FR-1.6 On hosts with multiple network interfaces, the hub ranks candidate interfaces and picks the most likely LAN interface; when ranking is ambiguous, it prompts the user to choose (M3) and updates QR/URLs immediately upon choice.
- FR-1.7 Quitting the hub terminates all connections; open client pages show the connection-lost state (B1).

### FR-2 Join & identity
- FR-2.1 Any device on the same network joins by opening the hub URL in a browser — via QR scan, `http://befrest.local`, or raw `IP:port`. No account, install, or pairing step exists.
- FR-2.2 The QR code encodes the raw `http://IP:port` URL.
- FR-2.3 On first visit, the client shows a name confirmation screen (S1) pre-filled with a name derived from the device's user-agent; the user may edit it before joining.
- FR-2.4 The device name persists on the device across visits; returning devices skip S1.
- FR-2.5 A device may rename itself at any time from the main screen (S2 header); the new name propagates to all connected devices.
- FR-2.6 When a joining or renaming device's name collides with a connected device's name, the hub disambiguates by suffixing (e.g. "Pixel 8 (2)").
- FR-2.7 The hub-launched page auto-joins with a name derived from the host machine's hostname, skipping S1.

### FR-3 Presence
- FR-3.1 Every connected client shows a live list of all *other* connected devices (S2 grid).
- FR-3.2 A newly joined device appears on all other clients' lists without any refresh action.
- FR-3.3 Closing a client page removes that device from all other lists.
- FR-3.4 When no other devices are connected, S2 shows the invite content (QR + URLs) as its empty state.
- FR-3.5 Any client can open the invite panel (S3) at any time via "Add device".

### FR-4 Sending
- FR-4.1 Tapping a device card opens the native file picker with multi-select enabled (v1).
- FR-4.2 On desktop browsers, dropping files onto a device card is equivalent to picking them for that device (v1); the card shows a drop-target highlight during dragover.
- FR-4.3 Confirming a file selection sends a transfer offer to the target device; the sender sees a "waiting for accept" state on the target's card.
- FR-4.4 A sender can cancel an offer while it is pending; the receiver's prompt (M1) closes automatically with a notice.
- FR-4.5 A sender can cancel an in-progress transfer; the receiver is notified and partial data is discarded.
- FR-4.6 A device can send while simultaneously receiving.

### FR-5 Receiving
- FR-5.1 A transfer offer surfaces on the receiver as a modal prompt (M1) showing sender name, file name(s), per-file size, and (for >4 files) a summary with total size.
- FR-5.2 Accept starts the transfer; Decline cancels it with zero file bytes transferred to the receiver.
- FR-5.3 The offer prompt has no timeout; it remains until answered, or closes itself if the sender cancels or disconnects.
- FR-5.4 Accepted files are delivered through the browser's normal download mechanism into the device's default download location.
- FR-5.5 A receiver can cancel an in-progress transfer; the sender is notified and partial data is discarded.

### FR-6 Transfer execution & progress
- FR-6.1 Transfers are streamed end-to-end; a file is never fully buffered in hub or client memory. Files of at least 2 GB must transfer successfully.
- FR-6.2 Both sender and receiver see live progress: filename, bytes transferred / total, percentage; multi-file transfers additionally show "file N of M" and an aggregate bar.
- FR-6.3 Progress UI offers a cancel affordance on both sides for the entire duration.
- FR-6.4 On completion, both sides get a success notice (B2); the receiver's notice confirms the file was saved.
- FR-6.5 On failure from any cause (sender disconnect, receiver disconnect, stream stall, cancel), both reachable sides get a failure notice naming the cause category, the UI resets, and partial data is discarded.
- FR-6.6 When an upload starts on a mobile browser, the sender UI shows a persistent hint to keep the screen on until sending finishes.

### FR-7 Connection resilience
- FR-7.1 If a client's connection to the hub drops, the client shows a reconnecting banner (B1), dims/disables send actions, and retries automatically.
- FR-7.2 On successful reconnect, the device list refreshes and the banner clears; the device reappears on other clients under its persisted name.
- FR-7.3 Where the hub can detect that its advertised address is unreachable (e.g. firewall heuristic), the invite panel shows a help hint naming the port.

---

## 2. Non-functional requirements

- NFR-1 **Throughput:** transfers proceed at wifi line rate; the hub relay is not the bottleneck on a typical 802.11ac LAN.
- NFR-2 **Memory:** hub memory use is bounded and independent of file size; disk spooling, if used, is temporary and cleaned after the transfer ends (success or failure).
- NFR-3 **Latency:** join-to-visible (QR scan to appearing in other devices' lists) under 5 seconds.
- NFR-4 **Footprint:** hub binary under 30 MB; zero runtime dependencies.
- NFR-5 **Platforms (hub):** Windows, macOS, Linux.
- NFR-6 **Platforms (client):** current Chrome, Safari, Firefox, Edge on desktop; current Safari (iOS) and Chrome (Android) on mobile.
- NFR-7 **Transport model:** v1 relays all transfers through the hub (device↔device = two hops); plain HTTP on the LAN, no TLS (trusted-network model per SPEC).
- NFR-8 **Accessibility:** WCAG 2.1 AA floor; all touch targets ≥ 56 px; every state change signaled by text (aria-live), never color alone.

---

## 3. User stories

- US-1 As a laptop owner, I download one file, double-click it, and my browser shows me a QR code — so that setup takes seconds and nothing is installed on my phone.
- US-2 As a phone user, I scan a QR with my camera app and land on a page already knowing a sensible name for my phone — so that joining is one tap.
- US-3 As a sender, I tap the device I can see in front of me and pick files — so that sending feels like handing something over, not configuring a connection.
- US-4 As a receiver, I see who is sending me what and how big it is before anything transfers — so that nothing lands on my device without consent.
- US-5 As a sender of a huge video, I watch a progress bar and get told clearly if it fails — so that I never wonder whether the file arrived.
- US-6 As a returning user, I rescan the QR after closing the tab and I'm back in seconds under the same name — so that leaving is never costly.
- US-7 As a desktop user, I drag files straight onto a device card — so that sending matches my desktop habits.
- US-8 As a user on a VPN-laden laptop, I'm asked which network my devices are on when it's ambiguous — so that the QR always points somewhere my phone can reach.

---

## 4. Acceptance criteria

All criteria are observable in the running app. Kernel criteria gate the kernel; v1 criteria gate v1.

### Kernel

- AC-1 On a fresh machine, download the binary and double-click it: within a few seconds a tray icon is visible and a browser tab opens showing a QR code, `http://befrest.local`, and an `IP:port` URL. *(FR-1.1, 1.2, FR-3.4)*
- AC-2 Scan the QR with a phone camera: the page opens on the phone showing a name field pre-filled with a recognizable device name (e.g. "Pixel 8" on a Pixel 8). *(FR-2.2, 2.3)*
- AC-3 Tap Join on the phone: within 5 seconds the phone appears by name on the laptop page, and the laptop appears on the phone, with no refresh on either side. *(FR-3.1, 3.2, NFR-3)*
- AC-4 On the phone, tap the laptop card: the native file picker opens and allows selecting multiple files. *(FR-4.1)*
- AC-5 Select a single file ≥ 2 GB and confirm: the laptop shows a prompt naming the phone, the filename, and a size ≥ 2 GB, before any transfer starts. *(FR-4.3, FR-5.1)*
- AC-6 Click Accept on the laptop: progress bars advance on both devices; when done, the file exists in the laptop's downloads folder, byte-identical to the original (verify with a checksum). *(FR-5.4, FR-6.1, 6.2, 6.4)*
- AC-7 While the 2 GB transfer runs, watch hub memory in the OS process monitor: it stays bounded (does not grow toward the file size). *(NFR-2)*
- AC-8 Click Decline instead of Accept: the prompt closes, the sender sees a "declined" notice, and no partial file appears in the receiver's downloads. *(FR-5.2, FR-6.5)*
- AC-9 Kill the phone's browser mid-transfer: within a few seconds the laptop shows a failure notice and no partial file remains available. *(FR-6.5)*
- AC-10 Close the phone's tab: within ~2 seconds the phone disappears from the laptop's list. Rescan the QR: the phone is back in the list within 5 seconds, under the same name, with no name re-entry screen. *(FR-3.3, FR-2.4, F3)*
- AC-11 Send a file from the laptop to the phone: the phone shows the accept prompt, and after accepting, the file lands in the phone browser's downloads. *(kernel journey step 4)*
- AC-12 Join a second phone with the same model (same suggested name): the device list shows it with a distinct suffixed name. *(FR-2.6)*
- AC-13 Start the hub twice / with the preferred port occupied: the second context binds a different port and the displayed QR decodes to the actually reachable URL. *(FR-1.4)*
- AC-14 Type `http://befrest.local` (with actual bound port if non-default) into a browser on the LAN: the app page loads. *(FR-1.5)*
- AC-15 On a host with an active VPN interface, launch the hub: either the QR encodes the real LAN address, or an interface-picker prompt appears; after choosing, the QR encodes the chosen interface's address. *(FR-1.6)*

### v1

- AC-16 Tap a device card and select 5 files: the receiver's prompt lists them (or shows a "+N more" summary with total size); after accept, all 5 land in downloads and progress showed "file N of 5". *(FR-4.1, FR-5.1, FR-6.2)*
- AC-17 On a desktop browser, drag files from the file manager over a device card: the card highlights; dropping starts the same offer flow as the picker. *(FR-4.2)*
- AC-18 Start an upload from a phone: a visible hint says to keep the screen on. Lock the phone screen mid-upload: the laptop shows a failure notice within a few seconds. *(FR-6.6, FR-6.5)*
- AC-19 Stop the hub process while a client page is open: the page shows a reconnecting banner and device cards are not tappable. Restart the hub: the banner clears without a manual page reload and the device list repopulates. *(FR-7.1, 7.2)*
- AC-20 Rename your device from the header: the new name appears on other devices' lists within 2 seconds. *(FR-2.5)*
- AC-21 While receiving a file, tap another device card and send it a file: both transfers proceed and show independent progress. *(FR-4.6)*

---

## 5. Validation rules

- V-1 Device name: 1–32 characters after trimming whitespace; empty name blocks Join with helper text; over-length input is truncated live with a counter.
- V-2 Device name uniqueness is enforced hub-side by suffixing, never by rejecting the join.
- V-3 Transfer offer: must contain ≥ 1 file; each file must report a size; offers to a device that has disconnected since the picker opened fail immediately with a notice ("Laptop is no longer connected").
- V-4 File count and size: no artificial upper limit imposed by the app; practical limits are the receiver's disk and browser.
- V-5 A device cannot send to itself: the sender's own device never appears in the S2 grid.
- V-6 Only currently-connected devices are valid transfer targets; the offer is validated against live presence at send time.

## 6. Error cases

| # | Condition | Required behavior |
|---|-----------|-------------------|
| E-1 | Hub unreachable on first page load / join | S1 shows inline error with Retry ("Can't reach the hub. Are you on the same wifi?") |
| E-2 | WebSocket drops on a joined client | B1 banner, send actions disabled, auto-reconnect; extended copy after 30 s |
| E-3 | Sender disconnects while offer pending | Receiver's M1 closes with "cancelled" notice |
| E-4 | Sender disconnects mid-transfer | Receiver notified, transfer marked failed, partial data discarded |
| E-5 | Receiver disconnects mid-transfer | Sender notified, transfer marked failed, hub discards buffered/spooled data |
| E-6 | Receiver declines | Sender notified; zero file bytes delivered |
| E-7 | Either side cancels mid-transfer | Other side notified with "cancelled" (distinct from "failed"); partial data discarded |
| E-8 | Mobile screen lock kills upload | Treated as E-4; sender-side hint (FR-6.6) mitigates proactively |
| E-9 | Preferred port busy | Next free port; all URLs/QR reflect actual port |
| E-10 | Inbound connections blocked (firewall) | Where detectable, invite panel shows help hint naming the port |
| E-11 | Target device vanished between picker open and confirm | Offer fails immediately with notice; no hung "waiting" state |
| E-12 | Hub host sleeps / hub quits | All clients enter E-2 state |

## 7. Edge cases

- EC-1 **Empty state:** single device connected → S2 shows invite content inline; no dead "empty list" moment. First joiner on any device sees this until a second device arrives.
- EC-2 **Offline / wrong network:** phone on cellular or a different SSID scanning the QR → page never loads (browser-level error); mitigation is the S3 instruction copy ("Are you on the same wifi?" appears on any in-app reachability failure, E-1).
- EC-3 **Same name devices:** handled by suffixing (FR-2.6), including the rename path.
- EC-4 **Multi-NIC host:** handled by ranking + M3 picker (FR-1.6).
- EC-5 **Simultaneous offers to one receiver:** prompts queue — receiver answers one at a time (second prompt appears after the first is resolved); neither offer is dropped.
- EC-6 **Send while receiving:** explicitly supported (FR-4.6, AC-21).
- EC-7 **Very many devices (> ~12):** grid scrolls vertically; no pagination.
- EC-8 **Hub page closed on hub machine:** hub keeps running (tray is the process owner); reopening from the tray menu restores the page. Closing the page only removes the hub *device* from lists, per FR-3.3.
- EC-9 **Zero-byte file:** transfers and completes normally (progress goes straight to done).
- EC-10 **Filename collisions in receiver's downloads:** deferred to the browser's native download dedup behavior (" (1)" suffixing); app does not manage it.

## 8. Constraints

- C-1 Clients are stock mobile/desktop browsers — no native apps, no browser extensions, no installs on client devices.
- C-2 Hub relays all transfer traffic in v1; no client↔client direct connections.
- C-3 Plain HTTP; the product's security model is "trusted local network" (SPEC §5). No TLS, no auth in v1.
- C-4 No persistence of files, device registry, or history on the hub beyond a running session, except temporary transfer spooling (cleaned per NFR-2).
- C-5 The entire client UI is one screen plus layers (UX.md); no additional routes/screens may be introduced in v1.
- C-6 Mobile browsers may kill background uploads on screen lock; v1 accepts this (SPEC explicit assumption) with the FR-6.6 hint as sole mitigation.

## 9. Out of scope (v1)

Internet/WAN transfers, accounts, cloud storage, chat, encryption-at-rest, native mobile apps, file preview, TLS, transfer resume, offline delivery, clipboard/text snippets, PIN-protected join, transfer history, auto-update.

## 10. Future improvements (Backlog, ranked — not promoted)

1. Offline inbox (send to a device whose page is closed; hub stores until fetched)
2. WebRTC direct P2P transfer (skip the hub relay hop for phone↔phone speed)
3. Resume interrupted transfers
4. Text/clipboard snippets
5. PIN-protected join (untrusted shared wifi)
6. Transfer history
7. Auto-update
