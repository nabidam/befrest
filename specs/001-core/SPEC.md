---
status: gate-passed
---

# Befrest — SPEC

## 1. Core promise

Send a file from any device to any other device on your local network in seconds — nothing to install on phones, nothing to configure.

## 2. Kernel

1. **Hub binary** — single downloadable executable. Double-click → tray icon appears, browser opens showing join page (QR code + `http://befrest.local:port` + raw IP:port fallback). QR encodes the IP URL (most reliable); mDNS name is the human-typeable alternative (port always explicit — mDNS resolves only the address).
2. **Join** — phone scans QR → page opens → device gets a name (auto-suggested from user-agent, editable) → appears in everyone's device list. No account, no pairing.
3. **Live device list** — every open page shows currently connected devices (WebSocket presence). Page closed = device disappears.
4. **Send with accept** — pick file(s), tap target device → receiver sees prompt (sender name, filename, size) → Accept starts transfer, Decline cancels. Progress bar on both sides. GB-scale files must work: streamed end-to-end, never fully buffered in memory.

### Kernel journey

1. Download `befrest` on laptop, double-click → tray icon, browser opens with QR and device name "Laptop".
2. Phone scans QR → page opens, named "Pixel 8" → both devices see each other in the list.
3. On phone: tap "Laptop", pick a 2 GB video → laptop shows accept prompt → Accept → progress on both sides → video lands in laptop's Downloads.
4. Laptop sends a photo back → phone accepts → photo downloads.
5. Phone closes the tab → vanishes from laptop's list. Rescans QR → back in seconds.

## 3. v1 / Backlog

**v1** = kernel + multi-file select (native file picker), drag-and-drop onto a device (desktop browsers).

**Backlog (ranked):**
1. Offline inbox (send to device whose page is closed; hub stores until fetched)
2. WebRTC direct P2P transfer (skip hub relay hop for phone↔phone speed)
3. Resume interrupted transfers
4. Text/clipboard snippets
5. PIN-protected join (untrusted shared wifi)
6. Transfer history
7. Auto-update

## 4. Edge cases

- Multi-NIC host (VPN, docker, ethernet+wifi): hub must pick the real LAN interface; ambiguous → let user choose which to show in QR.
- Phone screen locks mid-upload → browser kills upload; transfer marked failed, sender notified. v1 shows "keep screen on while sending" hint. (Explicit assumption: acceptable for v1; resume is backlog.)
- Receiver declines or disconnects mid-transfer → sender notified, partial data discarded.
- Two devices pick same name → hub suffixes ("Pixel 8 (2)").
- Port busy → hub picks next free port; QR always reflects actual URL.
- Firewall blocks inbound → hub detects unreachability where possible, shows help hint.

## 5. Non-functional requirements

- Transfer at wifi line rate; hub relays with bounded memory (spool to temp disk only if needed, cleaned after transfer).
- Join-to-visible under 5 seconds.
- Binary < 30 MB, zero runtime dependencies, Windows/macOS/Linux.
- Assumption: v1 transfers relay through hub (phone↔phone = two hops) — fine on LAN.
- Plain HTTP on LAN (trusted-network model); no TLS in v1.

## 6. Suggested tech stack

- **Hub:** Go — `net/http`, WebSocket (presence + signaling), `embed.FS` frontend, systray lib, mDNS announce, single static cross-compiled binary.
- **Frontend:** Svelte SPA, no heavy framework; touch-first. (Committed in ARCHITECTURE.md: Svelte 5 + TypeScript + Vite.)

## 7. Design direction

Personality: friendly, instant, unfussy. References: AirDrop share sheet, LocalSend, PairDrop. Touch-first spacious density (phones are primary clients); desktop same layout. Big tap targets, one screen, no navigation. WCAG AA floor.

## 8. Out of scope

Internet/WAN transfers, accounts, cloud storage, chat, encryption-at-rest, native mobile apps, file preview.
