# Befrest — DESIGN

Living document. The design-system contract for the Befrest client, styling the screens UX.md defines. Implemented as CSS custom properties in `web/src/styles/tokens.css`.

**SINGLE-SOURCE RULE:** exact values appear only in §2's token tables. All other prose — including component specs below — refers to tokens by name. `tokens.css` is a verbatim transcription of §2.

---

## 1. Direction

**Adjectives:** friendly, instant, unfussy (SPEC §7).
**References:** AirDrop share sheet (single decisive prompt), LocalSend (device grid warmth), PairDrop (one-screen radar feel). Patterns adopted, layouts not copied.
**Visual signature — the join pulse:** when a device joins, its card emits one soft expanding ring in `--color-accent`, fading over `--dur-pulse` with `--ease-decel`. The same accent drives every progress bar, so "something is happening" always looks like the same living color. Suppressed entirely under `prefers-reduced-motion`.

One typeface family for UI, one mono family for URLs/addresses — both system stacks: zero webfont bytes serves NFR-4 and the "instant" adjective.

---

## 2. Tokens

### 2.1 Color — light (default) and dark (`prefers-color-scheme: dark`)

| Token | Light | Dark | Role |
|---|---|---|---|
| `--color-bg` | `#F6F8FA` | `#10151C` | page background |
| `--color-surface` | `#FFFFFF` | `#1A222D` | cards, sheets, modals |
| `--color-surface-sunken` | `#EDF1F5` | `#141B24` | skeletons, progress track, input bg |
| `--color-text` | `#1A2330` | `#E8EEF5` | primary text |
| `--color-text-muted` | `#5A6B7E` | `#9FB0C3` | secondary text, sizes, labels |
| `--color-border` | `#D6DEE8` | `#2C3A4A` | card/input borders, dividers |
| `--color-accent` | `#0B5FCC` | `#5CA8FF` | primary actions, progress fill, focus ring, pulse |
| `--color-accent-hover` | `#0A54B4` | `#78B7FF` | hover on accent surfaces |
| `--color-accent-active` | `#094A9E` | `#4093F5` | pressed accent |
| `--color-on-accent` | `#FFFFFF` | `#0B1622` | text/icon on accent |
| `--color-danger` | `#C42B2B` | `#F26D6D` | destructive/cancel, failure toasts |
| `--color-on-danger` | `#FFFFFF` | `#240D0D` | text on danger fills |
| `--color-success` | `#1B7F4B` | `#4CC38A` | success toasts, done states |
| `--color-warn-text` | `#8A5A00` | `#E8B95C` | reachability hint text |
| `--color-warn-bg` | `#FFF4DC` | `#2A2415` | reachability hint background |
| `--color-scrim` | `#1A2330B3` | `#03060AB3` | modal/sheet backdrop |

Every text/background pairing above meets WCAG AA (≥ 4.5:1 body, ≥ 3:1 large text); re-verify with a contrast checker whenever a value changes.

### 2.2 Type

| Token | Value |
|---|---|
| `--font-ui` | `system-ui, -apple-system, 'Segoe UI', Roboto, sans-serif` |
| `--font-mono` | `ui-monospace, 'SF Mono', Menlo, Consolas, monospace` |
| `--text-sm` | `0.875rem / 1.25rem` |
| `--text-base` | `1rem / 1.5rem` |
| `--text-lg` | `1.1875rem / 1.625rem` |
| `--text-xl` | `1.5rem / 1.875rem` |
| `--text-display` | `1.875rem / 2.25rem` |
| `--weight-regular` | `400` |
| `--weight-semibold` | `600` |

Two weights only. `--font-mono` is used exclusively for URLs/IP addresses (S3, EmptyInvite).

### 2.3 Spacing (4/8 grid), radii, size

| Token | Value | | Token | Value |
|---|---|---|---|---|
| `--space-1` | `4px` | | `--radius-sm` | `8px` |
| `--space-2` | `8px` | | `--radius-md` | `12px` |
| `--space-3` | `12px` | | `--radius-lg` | `20px` |
| `--space-4` | `16px` | | `--radius-full` | `999px` |
| `--space-5` | `24px` | | `--size-touch` | `56px` |
| `--space-6` | `32px` | | `--size-card-min` | `96px` |
| `--space-7` | `48px` | | `--size-content-max` | `640px` |
| `--space-8` | `64px` | | `--size-bar` | `8px` |
| `--size-card-col` | `240px` | | `--size-qr` | `240px` |
| `--breakpoint` | `40rem` | | `--focus-ring` | `2px` |
| `--lift-hover` | `-2px` | | `--focus-offset` | `2px` |
| `--press-scale` | `0.98` | | `--opacity-disabled` | `0.5` |

`--breakpoint` is the one value components must repeat literally (`40rem`) — CSS custom properties don't work inside media queries. No other raw value may leave this section.

### 2.4 Shadows & motion

| Token | Value |
|---|---|
| `--shadow-card` | `0 1px 3px rgba(10, 20, 35, 0.08)` |
| `--shadow-layer` | `0 8px 32px rgba(10, 20, 35, 0.24)` |
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

- **Primary button (Join, Accept, Use this):** fill `--color-accent`, text `--color-on-accent`, radius `--radius-md`, type `--text-lg` `--weight-semibold`. Hover `--color-accent-hover`; active `--color-accent-active`; focus-visible per global rule; disabled per global rule (S1 Join with empty name). Loading: inline spinner + label swap ("Joining…"), still one button.
- **Secondary button (Decline, Close, Copy):** transparent fill, `--color-border` border, text `--color-text`. Hover: bg `--color-surface-sunken`; active: border `--color-accent`; Decline is identical in size to Accept (UX M1 — no dark pattern).
- **Danger/cancel affordance (transfer cancel):** text button in `--color-danger`; hover underline; active darkens via opacity. Never icon-only — always the word.
- **Device card (S2):** surface `--color-surface`, border `--color-border`, radius `--radius-lg`, shadow `--shadow-card`, min-height `--size-card-min`, whole card tappable. Hover: border `--color-accent`, translateY `--lift-hover` over `--dur-fast`. Focus-visible: global ring. Active: scale `--press-scale`. Disabled (B1 shown): global disabled rule, hover effects off. **Drop-target (dragover):** border `--color-accent` dashed, bg tinted toward accent, label "Drop to send to {name}". **Sending:** icon region replaced by progress (see data views); cancel affordance appears. **Join pulse** on entry (§1 signature).
- **Name input (S1, header rename):** bg `--color-surface-sunken`, border `--color-border`, radius `--radius-md`, text `--text-xl`, select-all on focus. Focus: border `--color-accent` + global ring. Invalid (empty): helper text in `--color-danger` under the field; Join disabled. At-limit: live counter in `--color-text-muted`.
- **Copy row (S3 URLs):** `--font-mono` `--text-base` + copy icon; whole row is the target. Hover bg `--color-surface-sunken`; on copy: icon swaps to a check in `--color-success` for `--dur-slow`, plus aria-live "Copied".
- **Rename ✎ (header), Add device (footer), M3 radio rows:** follow secondary-button rules; M3 radio rows are full-width `--size-touch`-tall list items, selected row border `--color-accent`.

### Data views

- **Device grid:** *empty* → EmptyInvite (S3 content inline: heading `--text-xl`, QR at `--size-qr`, copy rows); *loading* → 2 skeleton cards in `--color-surface-sunken` with a shimmer over `--dur-slow` (max ~2 s per UX S2); *error* → none by design: connection problems are B1's job, absent devices simply leave the grid (UX S2).
- **Progress (sender card / receiver banner):** track `--color-surface-sunken`, fill `--color-accent`, height `--size-bar`, radius `--radius-full`; filename `--text-base`, bytes + percent `--text-sm` `--color-text-muted`; multi-file adds "file N of M". Fill animates width over `--dur-base` `--ease-standard`. *Done:* fill switches to `--color-success` for `--dur-slow` before the surface reverts. *Failed/cancelled:* surface reverts immediately; the B2 toast carries the message.
- **Offer modal (M1):** surface `--color-surface`, radius `--radius-lg`, shadow `--shadow-layer`, over `--color-scrim`; sender name `--text-xl` `--weight-semibold`, file rows `--text-base` with sizes in `--color-text-muted`; > 4 files → "and N more" summary row. No outside-tap dismiss; Escape ≡ Decline.
- **Invite sheet (S3):** bottom sheet on mobile / centered card ≥ breakpoint, radius `--radius-lg`, shadow `--shadow-layer`; QR dominant at `--size-qr` (~50 % of sheet); *loading* (no `invite-info` yet): QR-sized skeleton; *warn:* reachability hint block in `--color-warn-bg`/`--color-warn-text`, radius `--radius-sm`.
- **B1 banner:** full-width under header, bg `--color-warn-bg`, text `--color-warn-text`, animated ellipsis dot (off under reduced motion); extended copy after 30 s per UX B1.
- **B2 toasts:** surface `--color-surface`, left edge `--space-1`-wide in `--color-success`/`--color-danger` by outcome, shadow `--shadow-layer`, radius `--radius-md`, `--text-base`; enter/exit over `--dur-base` `--ease-decel`; auto-dismiss after `--dur-toast`; `aria-live=polite`.

---

## 4. Layout

- Single column, max-width `--size-content-max`, centered; page padding `--space-4` below the breakpoint, `--space-6` above.
- One breakpoint: `--breakpoint` (repeated literally in media queries per §2.3 note). Below = touch-first (bottom-sheet S3); at/above = identical layout, denser nothing — spacing stays touch-first per SPEC §7, desktop only gains drag-and-drop and hover states.
- Device grid: `auto-fill` columns, min column `--size-card-col`, gap `--space-4`; > ~12 devices scrolls vertically, no pagination (EC-7).
- Layer stack (z-order low→high): S2 → B1 banner → receive banner → S3 sheet → M1/M3 modal → B2 toasts.

---

## 5. Hard rules

1. **Tokens only in components** — no raw hex, px, ms, or font values outside `tokens.css`. Enforced by grep in PLAN Chunk 11/12 acceptance.
2. **WCAG AA contrast** on every text/background pairing, both themes.
3. **Visible focus** on every interactive element via the global focus rule — never `outline: none` without a replacement.
4. **No template clichés:** no gradient hero, no glassmorphism, no emoji-as-UI (device icons are drawn SVG), no border-radius soup — three radii exist, use them.
5. **State changes are textual** (aria-live) as well as visual — color never signals alone (NFR-8).
6. **Theme follows the OS** (`prefers-color-scheme`); no in-app toggle — C-5 forbids a settings surface.
7. **Motion is subordinate:** nothing animates longer than `--dur-pulse`; everything respects `prefers-reduced-motion`.
