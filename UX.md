# Befrest — UX

Living document. Screens, wireframes, and flows for the Befrest web client and hub-side surfaces. Adapted patterns (not layouts) from AirDrop share sheet, LocalSend, PairDrop: one screen, device-grid-as-primary-surface, modal accept prompt, zero navigation chrome.

Design stance: **one screen**. Everything else is a modal, sheet, or transient banner layered on S2. Touch-first spacing (phones are primary clients); desktop gets the identical layout with drag-and-drop added.

---

## 1. Screen inventory

| id | Name | Type | Purpose | Entry points |
|----|------|------|---------|--------------|
| S1 | Name confirmation | Full-screen (first visit only) | Confirm/edit auto-suggested device name before joining | First load of the page on a device with no stored name |
| S2 | Device list (main) | Full-screen | See connected devices, pick a target, watch transfers | After S1; every subsequent page load (name remembered); after any modal closes |
| S3 | Invite panel | Sheet/modal over S2 | Show QR code + `http://befrest.local:port` + raw `IP:port` so new devices can join | Shown inline as S2's empty state whenever no other devices are connected (expected first view on the hub-launched page); "Add device" button on S2 (all devices) |
| M1 | Incoming transfer prompt | Modal over S2 | Receiver accepts or declines an offered transfer | Pushed by hub when another device targets this one |
| M2 | Transfer progress | Inline on S2 device card (sender) + banner (receiver) | Show live progress, allow cancel | Automatically when a transfer starts |
| M3 | Network interface picker | Modal over S2, hub page only | Resolve multi-NIC ambiguity: choose which interface's address goes in the QR | Pushed by hub when it cannot pick the LAN interface confidently; "change network" link inside S3 |
| B1 | Connection-lost banner | Persistent banner over S2 | Tell user the WebSocket dropped; auto-reconnect status | WebSocket disconnect |
| B2 | Result toast | Transient toast over S2 | Confirm transfer completed / failed / declined | Transfer terminal states |

No other screens exist. No settings page, no history page, no navigation.

---

## 2. Navigation map

```
first visit ──▶ S1 (name) ──confirm──▶ S2 (device list)
                                        │
returning visit ────────────────────────┤  (name remembered, straight to S2)
                                        │
        S2 ──"Add device"──────────────▶ S3 (invite sheet) ──close──▶ S2
        S2 ◀──hub push (offer)────────── M1 (accept prompt) ──accept/decline──▶ S2
        S2 ──tap device + pick files───▶ M2 (progress, inline) ──done/cancel──▶ S2
        S2 ◀──hub push (ambiguous NIC)── M3 (interface picker, hub page only) ──▶ S2
        S2 ◀──ws drop──────────────────  B1 (banner, non-blocking)
```

There is no back button and no route history: S3/M1/M3 are dismissible layers; closing any of them always reveals S2.

---

## 3. Wireframes per screen

### S1 — Name confirmation

```
┌──────────────────────────────────┐
│                                  │
│            [befrest logo]        │   region 1: brand mark, small
│                                  │
│   You'll appear to others as:    │   region 2: prompt line
│                                  │
│   ┌──────────────────────────┐   │   region 3: name field, pre-filled
│   │  Pixel 8            [✕]  │   │   from user-agent, focused,
│   └──────────────────────────┘   │   select-all on focus
│                                  │
│   ┌──────────────────────────┐   │   region 4: primary action,
│   │         Join  →          │   │   full-width, min 56px tall
│   └──────────────────────────┘   │
│                                  │
└──────────────────────────────────┘
```

- **Eye goes first to:** the pre-filled name field (largest text on screen, focused).
- **Primary action:** Join. Enter key submits.
- **Validation state:** empty name → Join disabled, helper text "Give this device a name". Name over 32 chars → truncated with counter.
- **Loading state:** Join pressed → button shows spinner, "Joining…"; on success S2 appears.
- **Error state:** hub unreachable → inline message under button: "Can't reach the hub. Are you on the same wifi?" with Retry.
- Shown **once per device** (name persisted in localStorage). Name is editable later by tapping own device chip on S2.

### S2 — Device list (main screen)

```
┌──────────────────────────────────┐
│  befrest        [You: Pixel 8 ✎] │  region 1: header — own identity,
│                                  │  tap ✎ to rename inline
├──────────────────────────────────┤
│                                  │
│   Send to                        │  region 2: section label
│                                  │
│   ┌───────────┐  ┌───────────┐   │  region 3: device grid —
│   │    💻     │  │    💻     │   │  one card per OTHER device.
│   │  Laptop   │  │ Dana's PC │   │  Card: device-type icon,
│   │           │  │           │   │  name. Min 96px tall,
│   └───────────┘  └───────────┘   │  full card is the tap target.
│                                  │
│   (card during send:)            │
│   ┌───────────┐                  │
│   │  Laptop   │                  │
│   │ ▓▓▓▓░░ 64%│  video.mp4      │  progress replaces icon,
│   │ [cancel]  │  1.3 / 2.0 GB   │  cancel appears
│   └───────────┘                  │
│                                  │
├──────────────────────────────────┤
│  [＋ Add device]                 │  region 4: footer — opens S3
└──────────────────────────────────┘
```

- **Eye goes first to:** the device grid (biggest, most colorful region).
- **Primary action:** tap a device card → native file picker opens (multi-select). On desktop, dragging files onto a card is equivalent; card shows a drop highlight ("Drop to send to Laptop") on dragover.
- **Empty state (no other devices):** grid replaced by centered invite content — "No other devices yet" + inline QR + `http://befrest.local:port` + `IP:port`, i.e. S3's content promoted into the page body. On the hub-launched page this is the expected first view.
- **Loading state:** between page load and first WebSocket device-list message: skeleton cards (2 gray placeholders), max ~2s before B1 logic kicks in.
- **Error state:** covered by B1 (connection lost) and B2 (transfer failures) — the grid itself never shows an error; unreachable devices simply disappear.
- **Receiving state:** while receiving, a slim progress banner pins under the header: "Receiving video.mp4 from Pixel 8 — ▓▓░░ 41% [cancel]". Grid stays usable (can still send while receiving).

### S3 — Invite panel (sheet over S2)

```
┌──────────────────────────────────┐
│  Add a device                [✕] │  region 1: title + close
├──────────────────────────────────┤
│                                  │
│         ┌────────────┐           │  region 2: QR code, dominant,
│         │  ▓▓ QR ▓▓  │           │  encodes http://<ip>:<port>
│         │  ▓▓▓▓▓▓▓▓  │           │
│         └────────────┘           │
│                                  │
│   Scan with a phone camera,      │  region 3: instructions +
│   or type either of these:       │  human-typeable fallbacks
│                                  │
│   http://befrest.local:5311 [⧉] │  tap to copy
│   http://192.168.1.7:5311  [⧉]  │  tap to copy
│   (mDNS resolves the name only;  │
│    the port is always shown)     │
│                                  │
│   (hub page only:)               │
│   Wrong network? Change network  │  region 4: link → M3
└──────────────────────────────────┘
```

- **Eye goes first to:** the QR code.
- **Primary action:** none on-screen — the action happens on the *other* device (scanning). Close is secondary.
- **Auto-behavior:** on the hub-launched page, S3 content is shown inline (S2 empty state) until a second device joins; the moment one joins, it collapses into the grid with a subtle "Pixel 8 joined" toast.
- **Error state:** if hub reports the QR address may be unreachable (firewall heuristic), a warm-toned hint (DESIGN warn tokens) appears under the QR: "If scanning doesn't work, check your firewall allows befrest on port 5311."

### M1 — Incoming transfer prompt

```
┌──────────────────────────────────┐
│                                  │
│   📁  Incoming files             │  region 1: type icon + title
│                                  │
│   Pixel 8 wants to send you      │  region 2: sender name (bold),
│                                  │
│   video.mp4                      │  region 3: file list — name +
│   2.0 GB                         │  size per file; >4 files →
│                                  │  "and 3 more" summary line
│                                  │  with total size
│   ┌────────────┐ ┌────────────┐  │
│   │   Decline  │ │   Accept   │  │  region 4: actions — Accept
│   └────────────┘ └────────────┘  │  is the filled/primary button
│                                  │
└──────────────────────────────────┘
```

- **Eye goes first to:** sender name + filename (answers "who, what" before any button press).
- **Primary action:** Accept. Decline is visually secondary but same size (no dark pattern — declining is legitimate).
- **No timeout dismissal in v1:** prompt stays until answered or sender cancels; if sender cancels/disconnects, modal closes itself with toast "Pixel 8 cancelled".
- Modal is **not** dismissible by tapping outside (accident-prone mid-decision); Decline is the escape.

### M2 — Transfer progress (states, not a separate screen)

Sender side: rendered **inside the target device's card** on S2 (see S2 wireframe) — progress bar, filename, `sent / total`, cancel. Multiple files: current file name + "file 2 of 5" + aggregate bar.
Receiver side: pinned banner under S2 header — same info + cancel.

- **Completion:** bar fills → card/banner reverts, B2 toast "Sent video.mp4 to Laptop ✓" / "Saved to Downloads ✓".
- **Failure/cancel:** B2 toast with reason — "Laptop declined", "Transfer failed — Pixel 8 disconnected", "Cancelled".
- **Mobile hint (v1):** the moment an upload starts on a mobile browser, a persistent inline hint under the progress: "Keep this screen on until sending finishes."

### M3 — Network interface picker (hub page only)

```
┌──────────────────────────────────┐
│  Which network are your devices  │
│  on?                             │
├──────────────────────────────────┤
│  ○ wifi     192.168.1.7          │  one row per candidate
│  ○ ethernet 10.0.0.4             │  interface: kind + address
│  ○ vpn      100.64.2.11          │
│                                  │
│  ┌──────────────────────────┐    │
│  │        Use this          │    │
│  └──────────────────────────┘    │
└──────────────────────────────────┘
```

- Appears only when the hub cannot rank interfaces confidently; most-likely candidate pre-selected. Choice updates the QR/URLs in S3 immediately.

### B1 — Connection-lost banner

Pinned full-width under header on S2: "Connection lost — reconnecting…" with animated dot. Auto-retries; on success it disappears with a brief "Reconnected" flash. While shown, device cards are dimmed and non-tappable (can't offer a transfer with no hub). After 30s of failure the copy extends: "…Still trying. Is the hub running?"

---

## 4. Key flows

### F1 — Kernel journey (first send, phone → laptop)

| # | User sees | User does | System responds |
|---|-----------|-----------|-----------------|
| 1 | (Laptop) Downloads `befrest`, double-clicks | — | Tray icon appears; default browser opens the app; name "Laptop" auto-confirmed for hub-launched page (S1 skipped, hub knows its own hostname); S2 shows empty state with inline QR (S3 content) |
| 2 | (Phone) Camera app over the QR | Taps the URL notification | Browser opens S1, name field pre-filled "Pixel 8" |
| 3 | (Phone) S1 with suggested name | Taps **Join** | S2 appears; grid shows "Laptop". (Laptop) QR collapses, grid shows "Pixel 8", toast "Pixel 8 joined" |
| 4 | (Phone) S2 grid | Taps **Laptop** card | Native file picker opens (multi-select enabled) |
| 5 | (Phone) File picker | Picks 2 GB video, confirms | Card shows "Waiting for Laptop to accept…"; (Laptop) M1 appears: "Pixel 8 wants to send you video.mp4, 2.0 GB" |
| 6 | (Laptop) M1 | Clicks **Accept** | Transfer starts. (Phone) card shows progress + keep-screen-on hint; (Laptop) receive banner shows progress |
| 7 | Both: progress bars fill | Waits | (Laptop) browser saves file to Downloads; B2 both sides: "Saved ✓" / "Sent ✓" |

### F2 — Reply send (laptop → phone)

| # | User sees | User does | System responds |
|---|-----------|-----------|-----------------|
| 1 | (Laptop) S2 grid with "Pixel 8" | Drags a photo from desktop onto the Pixel 8 card (or taps card → picker) | Card highlights "Drop to send to Pixel 8"; on drop: "Waiting for Pixel 8 to accept…" |
| 2 | (Phone) M1 prompt | Taps **Accept** | Progress both sides (fast — small file) |
| 3 | (Phone) B2 toast "Saved to Downloads ✓" | Opens photo from browser downloads | — |

### F3 — Leave and rejoin

| # | User sees | User does | System responds |
|---|-----------|-----------|-----------------|
| 1 | (Phone) S2 | Closes the tab | (Laptop) "Pixel 8" card fades out of grid within ~2s (WebSocket close) |
| 2 | (Phone) Camera over QR again (or taps `befrest.local` from history) | Opens page | S1 is **skipped** — name remembered; S2 loads directly; (Laptop) "Pixel 8" reappears. Join-to-visible under 5s |

### F4 — Decline and mid-transfer failure

| # | User sees | User does | System responds |
|---|-----------|-----------|-----------------|
| 1 | (Receiver) M1 prompt | Taps **Decline** | Modal closes; (Sender) B2 toast "Laptop declined"; card resets; no bytes transferred |
| 2 | (Sender, retry) sends again, transfer running | Phone screen locks mid-upload | Browser kills upload; hub detects stalled stream; both sides get B2 "Transfer failed"; receiver's partial data discarded; sender's card resets |
| 3 | (Receiver, alt) transfer running | Taps **cancel** on receive banner | Same failure path: sender toast "Laptop cancelled the transfer", partial data discarded |

---

## 5. Density & hierarchy notes

### S1
- Exactly one decision (the name), one action (Join). Nothing else exists — no links, no settings.
- One tap away: joining with suggested name (zero typing needed).

### S2
- **One tap away:** sending to any device (tap card → picker), adding a device (footer button), renaming self (header ✎).
- **Deliberately buried:** nothing on this screen has a second level. Rename is the only disclosure (tap ✎ → inline edit). There is intentionally no per-device menu, no long-press actions, no settings.
- Grid cards are the whole interaction surface — the entire card is tappable/droppable, never just an inner button.
- Touch targets ≥ 56px everywhere; grid gaps generous (touch-first spacious density per SPEC design direction).

### S3
- QR gets ~50% of the sheet; text fallbacks are one visual level down (people scan first, type as fallback).
- Copy buttons one tap away; interface change (M3) buried behind a text link — rare, expert-only.

### M1
- Identity ("who") and payload ("what, how big") carry the visual weight; buttons second.
- Nothing to disclose or configure — binary decision by design.

### M2
- Progress is glanceable from across the room: big bar, percentage, no small text needed to know state.
- Cancel one tap away on both sides throughout the transfer.

### Global
- No navigation chrome anywhere. The app is S2 plus layers; a user who has seen S2 once has seen the whole app.
- All states (empty, loading, receiving, disconnected) are expressed *on* S2, never on separate routes — supports the "one screen, no navigation" SPEC direction.
- WCAG AA floor: all state changes (join/leave, transfer start/end) get both a visual change and a toast/text announcement (aria-live) — never color-only signaling.
