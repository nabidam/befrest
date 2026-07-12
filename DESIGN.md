# Befrest — DESIGN

Living document. The design-system contract for the Befrest client, styling the screens UX.md defines. Implemented as CSS custom properties in `web/src/styles/tokens.css`. Palette, type, shape, and elevation values are adopted from **DESIGN_SYSTEM.md (BegireX)**; this file maps those tokens onto Befrest's semantic names and screens.

**SINGLE-SOURCE RULE:** exact values appear only in §2's token tables. All other prose — including component specs below — refers to tokens by name. `tokens.css` is a verbatim transcription of §2.

---

## 1. Direction

**Adjectives:** friendly, instant, unfussy (SPEC §7) — carried by BegireX's **Minimalist Modern, dark-first** aesthetic: a deep ink-blue ground, hierarchy built from tonal surface layers instead of shadows and blurs, and one pale periwinkle accent that is the only saturated color on screen.
**References:** AirDrop share sheet (single decisive prompt), LocalSend (device grid warmth), PairDrop (one-screen radar feel). Patterns adopted, layouts not copied.
**Visual signature — the join pulse:** when a device joins, its card emits one soft expanding ring in `--color-accent`, fading over `--dur-pulse` with `--ease-decel`. The same accent drives every progress bar and primary action, so "something is happening" always looks like the same living color. Suppressed entirely under `prefers-reduced-motion`.

Two typefaces (BegireX): **Instrument Sans** for all UI, **JetBrains Mono** for system data — URLs, addresses, byte counts, percentages. The mono/UI split makes "machine facts" visually distinct from labels. Both are self-hosted woff2 latin subsets served from the hub itself (embedded in the binary) — no network fetch beyond the LAN, `font-display: swap` over system-stack fallbacks, so the "instant" adjective survives (NFR-4).

**Dark-only in v1.** BegireX defers light mode; there is no OS-driven theme switch and no in-app toggle (C-5). If light mode arrives later it reuses the same token names with new values — components never change.

---

## 2. Tokens

### 2.1 Color — dark (sole theme)

Values are the BegireX roles named in the third column; Befrest keeps its own semantic layer so components read by purpose.

| Token | Value | BegireX role | Role in Befrest |
|---|---|---|---|
| `--color-bg` | `#0b1326` | surface | page background |
| `--color-surface-sunken` | `#060e20` | surface-container-lowest | input bg, progress track, skeletons — "carved in" |
| `--color-surface` | `#171f33` | surface-container | device cards, copy rows |
| `--color-surface-raised` | `#222a3d` | surface-container-high | modals, sheets, toasts, banners |
| `--color-surface-hover` | `#2d3449` | surface-container-highest | hover fill on secondary buttons, rows, list items |
| `--color-text` | `#dae2fd` | on-surface | primary text |
| `--color-text-muted` | `#c7c4d7` | on-surface-variant | secondary text, sizes, labels |
| `--color-icon` | `#b9c8de` | secondary | device-type icons, non-critical glyphs |
| `--color-border` | `#464554` | outline-variant | card/input borders, dividers, layer definition |
| `--color-accent` | `#c0c1ff` | primary | primary actions, progress fill, focus ring, pulse, success edge |
| `--color-accent-hover` | `#e1e0ff` | primary-fixed | hover on accent fills, done-flash |
| `--color-on-accent` | `#1000a9` | on-primary | text/icon on accent fills |
| `--color-danger` | `#ffb4ab` | error | destructive/cancel text, failure toast edge |
| `--color-on-danger` | `#690005` | on-error | text on danger fills |
| `--color-warn-text` | `#ffb783` | tertiary | reachability hint text, B1 banner text |
| `--color-warn-bg` | `#452000` | on-tertiary-container (as bg) | reachability hint / B1 banner background |
| `--color-scrim` | `#060e20b3` | surface-container-lowest @ 70 % | modal/sheet backdrop |

No separate success color: BegireX has none, and NFR-8 already forbids color-only signaling — "done" is carried by copy, the check icon, and the accent edge (`--color-accent`), keeping the one-living-color rule of §1.

Every text/background pairing above meets WCAG AA (≥ 4.5:1 body, ≥ 3:1 large text/icons); re-verify with a contrast checker whenever a value changes.

### 2.2 Type

| Token | Value |
|---|---|
| `--font-ui` | `'Instrument Sans', system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif` |
| `--font-mono` | `'JetBrains Mono', ui-monospace, 'SF Mono', Menlo, Consolas, monospace` |

Type roles (BegireX scale — size/line-height, weight, tracking):

| Token | Size/Line | Weight | Tracking | Use |
|---|---|---|---|---|
| `--text-display` | `1.5rem / 2rem` | `600` | `-0.02em` | S1 prompt + name field, EmptyInvite heading |
| `--text-headline` | `1.125rem / 1.5rem` | `600` | — | modal/sheet titles, M1 sender name, button labels (weight only) |
| `--text-body-lg` | `1rem / 1.5rem` | `400` | — | default UI text: buttons, file names, copy-row URLs |
| `--text-body-sm` | `0.875rem / 1.25rem` | `400` | — | secondary text, helper/status lines, toasts |
| `--text-label-mono` | `0.75rem / 1rem` | `500` | — | byte counts, percentages, "file N of M", char counter — always `--font-mono` |
| `--text-caption` | `0.75rem / 1rem` | `500` | `0.01em` | section labels ("Send to"), uppercase |

| Token | Value |
|---|---|
| `--weight-regular` | `400` |
| `--weight-medium` | `500` |
| `--weight-semibold` | `600` |

Three weights. `--font-mono` is used exclusively for system data: URLs/IP addresses (S3, EmptyInvite, rendered at `--text-body-lg` size for touch legibility) and everything set in `--text-label-mono`. **Deviation from BegireX noted:** BegireX makes `body-sm` (14 px) the interface default; Befrest keeps `body-lg` (16 px) for primary content because phones are the primary clients and 16 px is the mobile readability floor — `body-sm` covers secondary text only.

Font files: `web/public/fonts/*.woff2` (latin subsets, weights 400/500/600 or variable), declared via `@font-face` in `web/src/styles/fonts.css` with `font-display: swap`. No CDN, ever — the app must work on an offline LAN.

### 2.3 Spacing (4px baseline), radii, size

| Token | Value | | Token | Value |
|---|---|---|---|---|
| `--space-xs` | `4px` | | `--radius-sm` | `0.25rem` |
| `--space-sm` | `8px` | | `--radius-md` | `0.5rem` |
| `--space-md` | `16px` | | `--radius-lg` | `0.75rem` |
| `--space-lg` | `24px` | | `--radius-full` | `9999px` |
| `--space-xl` | `48px` | | `--size-touch` | `56px` |
| `--space-margin` | `32px` | | `--size-card-min` | `96px` |
| `--size-card-col` | `240px` | | `--size-content-max` | `640px` |
| `--size-qr` | `240px` | | `--size-bar` | `8px` |
| `--breakpoint` | `40rem` | | `--focus-ring` | `2px` |
| `--lift-hover` | `-2px` | | `--focus-offset` | `2px` |
| `--press-shift` | `1px` | | `--press-scale` | `0.98` |
| `--opacity-disabled` | `0.5` | | | |

Shape language (BegireX "Soft"): `--radius-sm` for buttons, inputs, and hint blocks; `--radius-md` for cards and toasts; `--radius-lg` for sheets and modals; `--radius-full` for progress bars and the pulse ring. `--size-touch` stays at 56 px — PRD NFR-8 is stricter than BegireX's 44 px minimum and wins.

`--breakpoint` is the one value components must repeat literally (`40rem`) — CSS custom properties don't work inside media queries. No other raw value may leave this section.

### 2.4 Elevation & motion

**Tonal layers, not shadows** (BegireX / WebKitGTK-safe): hierarchy comes from the `--color-surface*` steps plus a 1px solid `--color-border`. Cards cast no shadow. Only floating layers (modals, sheets, toasts) get the single shadow token, for separation from the scrim.

| Token | Value |
|---|---|
| `--shadow-layer` | `0 8px 32px rgba(3, 6, 14, 0.5)` |
| `--dur-fast` | `120ms` |
| `--dur-base` | `200ms` |
| `--dur-slow` | `400ms` |
| `--dur-pulse` | `1200ms` |
| `--dur-toast` | `4000ms` |
| `--ease-standard` | `cubic-bezier(0.2, 0, 0, 1)` |
| `--ease-decel` | `cubic-bezier(0, 0, 0, 1)` |

Under `prefers-reduced-motion: reduce`, all durations drop to `0ms` and the join pulse is removed (state changes remain textual per NFR-8, so nothing is lost).

---

## 3. Component states

**Global interaction rules:** every interactive element has a minimum hit area of `--size-touch` square; focus is always visible: `--focus-ring` solid `--color-accent` outline at `--focus-offset` offset on `:focus-visible`; disabled = `--opacity-disabled` opacity + `cursor: not-allowed` + `aria-disabled`; all transitions use `--dur-fast`/`--ease-standard` unless noted.

### Interactive elements

- **Primary button (Join, Accept, Use this):** fill `--color-accent`, text `--color-on-accent`, radius `--radius-sm`, type `--text-body-lg` `--weight-semibold`. Hover `--color-accent-hover`; active: depress by `--press-shift` (translateY) — BegireX press behavior; focus-visible per global rule; disabled per global rule (S1 Join with empty name). Loading: inline spinner + label swap ("Joining…"), still one button.
- **Secondary button (Decline, Close, Copy):** ghost — transparent fill, 1px `--color-border` border, text `--color-text`, radius `--radius-sm`. Hover: bg `--color-surface-hover`; active: depress by `--press-shift` + border `--color-accent`; Decline is identical in size to Accept (UX M1 — no dark pattern).
- **Danger/cancel affordance (transfer cancel):** text button in `--color-danger`; hover underline; active darkens via opacity. Never icon-only — always the word.
- **Device card (S2):** surface `--color-surface`, 1px `--color-border`, radius `--radius-md`, no shadow (tonal layering §2.4), min-height `--size-card-min`, device icon in `--color-icon`, whole card tappable. Hover: border `--color-accent`, translateY `--lift-hover` over `--dur-fast`. Focus-visible: global ring. Active: scale `--press-scale`. Disabled (B1 shown): global disabled rule, hover effects off. **Drop-target (dragover):** border `--color-accent` dashed, bg `--color-surface-hover`, label "Drop to send to {name}". **Sending:** icon region replaced by progress (see data views); cancel affordance appears. **Join pulse** on entry (§1 signature).
- **Name input (S1, header rename):** bg `--color-surface-sunken` ("carved" per BegireX), 1px `--color-border`, radius `--radius-sm`, text `--text-display`, select-all on focus. Focus: border `--color-accent` + global ring. Invalid (empty): helper text `--text-body-sm` in `--color-danger` under the field; Join disabled. At-limit: live counter in `--text-label-mono` `--color-text-muted`.
- **Copy row (S3 URLs):** `--font-mono` at `--text-body-lg` size + copy icon; whole row is the target, bg `--color-surface`, radius `--radius-sm`. Hover bg `--color-surface-hover`; on copy: icon swaps to a check in `--color-accent` for `--dur-slow`, plus aria-live "Copied".
- **Rename ✎ (header), Add device (footer):** follow secondary-button rules.
- **M3 radio rows:** full-width `--size-touch`-tall list items; custom-drawn radio mark (never native controls — BegireX input rule), selected row border `--color-accent`, interface kind in `--text-body-lg`, address in `--font-mono`.
- **Section label ("Send to"):** `--text-caption`, uppercase, `--color-text-muted`.

### Data views

- **Device grid:** *empty* → EmptyInvite (S3 content inline: heading `--text-display`, QR at `--size-qr`, copy rows); *loading* → 2 skeleton cards in `--color-surface-sunken` with a shimmer over `--dur-slow` (max ~2 s per UX S2); *error* → none by design: connection problems are B1's job, absent devices simply leave the grid (UX S2).
- **Progress (sender card / receiver banner):** track `--color-surface-sunken`, fill `--color-accent`, height `--size-bar`, radius `--radius-full` (pill — BegireX). Filename `--text-body-lg`; bytes + percent `--text-label-mono` `--color-text-muted`; status line ("Waiting for Laptop to accept…") `--text-body-sm`; multi-file adds "file N of M" in `--text-label-mono`. Fill animates width over `--dur-base` `--ease-standard`. *Done:* fill flashes `--color-accent-hover` for `--dur-slow` before the surface reverts; the B2 toast carries the confirmation. *Failed/cancelled:* surface reverts immediately; the B2 toast carries the message. (BegireX's 4 px "standard" bar height has no use here — every Befrest bar is an active transfer, so all bars are `--size-bar`.)
- **Offer modal (M1):** surface `--color-surface-raised`, 1px `--color-border`, radius `--radius-lg`, shadow `--shadow-layer`, over `--color-scrim`; sender name `--text-headline`; file rows `--text-body-lg` with sizes in `--text-label-mono` `--color-text-muted`; > 4 files → "and N more" summary row. No outside-tap dismiss; Escape ≡ Decline.
- **Invite sheet (S3):** bottom sheet on mobile / centered card ≥ breakpoint, surface `--color-surface-raised`, 1px `--color-border`, radius `--radius-lg`, shadow `--shadow-layer`; title `--text-headline`; QR dominant at `--size-qr` (~50 % of sheet) on a `--color-accent-hover` quiet-light tile so it scans reliably against the dark theme; *loading* (no `invite-info` yet): QR-sized skeleton; *warn:* reachability hint block in `--color-warn-bg`/`--color-warn-text`, radius `--radius-sm`.
- **B1 banner:** full-width under header, bg `--color-warn-bg`, text `--color-warn-text` `--text-body-sm`, animated ellipsis dot (off under reduced motion); extended copy after 30 s per UX B1.
- **B2 toasts:** surface `--color-surface-raised`, 1px `--color-border`, left edge `--space-xs`-wide in `--color-accent` (success) / `--color-danger` (failure) by outcome, shadow `--shadow-layer`, radius `--radius-md`, `--text-body-sm`; enter/exit over `--dur-base` `--ease-decel`; auto-dismiss after `--dur-toast`; `aria-live=polite`.

---

## 4. Layout

- Single column, max-width `--size-content-max`, centered; page padding `--space-md` below the breakpoint, `--space-margin` above (BegireX safe zone on large displays).
- One breakpoint: `--breakpoint` (repeated literally in media queries per §2.3 note). Below = touch-first (bottom-sheet S3); at/above = identical layout, denser nothing — spacing stays touch-first per SPEC §7, desktop only gains drag-and-drop and hover states. BegireX's fixed sidebar/rail pattern does not apply — Befrest is one screen with no navigation (C-5).
- Interactive elements separated by `--space-md` (BegireX rhythm); device grid: `auto-fill` columns, min column `--size-card-col`, gap `--space-md`; > ~12 devices scrolls vertically, no pagination (EC-7).
- Layer stack (z-order low→high): S2 → B1 banner → receive banner → S3 sheet → M1/M3 modal → B2 toasts.
- Logical properties (`margin-inline`, `padding-block`) throughout — BegireX RTL-readiness for free.

---

## 5. Hard rules

1. **Tokens only in components** — no raw hex, px, ms, or font values outside `tokens.css`/`fonts.css`. Enforced by grep in PLAN Chunk 11/12 acceptance.
2. **WCAG AA contrast** on every text/background pairing.
3. **Visible focus** on every interactive element via the global focus rule — never `outline: none` without a replacement.
4. **No template clichés:** no gradient hero, no glassmorphism or backdrop blurs (WebKitGTK-safe per BegireX), no emoji-as-UI (device icons are drawn SVG in `--color-icon`), no border-radius soup — three radii + pill exist, use them.
5. **State changes are textual** (aria-live) as well as visual — color never signals alone (NFR-8).
6. **Dark-only v1** (BegireX defers light mode); no `prefers-color-scheme` branching, no in-app toggle — C-5 forbids a settings surface. A future light theme swaps token values, never token names.
7. **Motion is subordinate:** nothing animates longer than `--dur-pulse`; everything respects `prefers-reduced-motion`.
8. **Fonts are self-hosted** (embedded in the hub binary) with system-stack fallbacks — no external font requests.
