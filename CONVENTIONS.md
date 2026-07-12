# Befrest — CONVENTIONS

Living document. Applies to all code in this repo. ARCHITECTURE.md owns *what* exists; this owns *how* it's written.

## Naming

- **Go:** package names short, lowercase, no underscores (`presence`, `netinfo`). Exported identifiers only when crossed by another package. Message type constants mirror the wire string: `MsgOffer` ↔ `"offer"`. Mutex fields named `mu`, guarding the fields declared below them.
- **TypeScript:** `camelCase` values, `PascalCase` types/components, `SCREAMING_SNAKE` never. Svelte components named after their UX id's element (`OfferModal` = M1, `ConnectionBanner` = B1); one component per file, filename = component name.
- **Wire protocol:** message `type` strings are `kebab-case` (`offer-cancelled`), payload fields `camelCase` (`transferId`). `internal/proto` is the single source of truth; **any change to it and `web/src/lib/proto.ts` lands in the same commit** — a PR touching one without the other is rejected.
- **Files/URLs:** HTTP paths `kebab-case` plural (`/api/transfers/{tid}/files/{idx}`). localStorage keys prefixed `befrest.` (`befrest.deviceId`).
- **CSS:** custom properties only from `tokens.css`/`fonts.css`, named `--{category}-{name}` (`--color-accent`, `--space-md`). Component classes `kebab-case`, scoped by Svelte — no global class taxonomy, no BEM.

## Error handling

- **Hub:** errors flow up; only `internal/server` translates them to HTTP statuses (ARCHITECTURE §4.1 table) or WS `error` frames. Wrap with context: `fmt.Errorf("accept transfer %s: %w", id, err)`. No `panic` across package boundaries; per-connection goroutines `recover`, log, and treat as disconnect. Sentinel errors (`ErrNotFound`, `ErrWrongState`) exported from `transfer`/`presence` for `errors.Is` in `server`.
- **Verdict rule:** the WS `transfer-done`/`transfer-failed` event is the only truth clients render from; HTTP stream errors are symptoms and are swallowed client-side (ARCHITECTURE §7).
- **Copy rule:** user-facing failure text comes only from the reason→copy table in `web/src/lib/format.ts`; no component invents error strings.
- **Logging:** `log/slog` to stderr, key-value style (`slog.Error("relay failed", "tid", id, "err", err)`). No log package in `web/` — `console.error` only inside `ws.ts`/`upload.ts` seams.

## Folder rules

- Dependency direction is ARCHITECTURE §6 and is law: `proto` is the leaf; `presence`/`transfer` never import `server` or anything network-shaped; UI components never import `ws.ts` — stores are the seam.
- New hub code goes in an existing `internal/` package; a new package requires an ARCHITECTURE.md update in the same PR.
- `web/src/screens/` holds only components mapped to UX.md ids; shared logic lives in `web/src/lib/`. No `utils.ts` grab-bag — name files by what they do (`format.ts`, `upload.ts`).
- Generated output (`web/dist/`) is never edited and never imported from source.

## Test style

- **Go:** stdlib `testing` only. Table-driven where there are ≥ 2 cases; subtests named for the behavior (`t.Run("suffixes duplicate name", …)`). Network-touching tests use `httptest`; `presence`/`transfer` tests construct structs directly, no server. Test files live beside the code.
- **Vitest:** colocated `*.test.ts` next to the module; test pure logic (`format.ts`, store transitions), not Svelte rendering.
- **Playwright:** `e2e/` only, two-browser-context pattern; assert on user-visible text and downloaded file content, never on internal store state.
- Every PRD error case (E-n) that gets a fix must gain a test reproducing it first.

## Commit style

- Conventional commits: `feat(transfer): relay with 4MiB bounded pipe`, `fix(presence): dedup on rename`. Scopes = package or screen name (`transfer`, `presence`, `server`, `netinfo`, `web`, `e2e`, `docs`).
- One chunk (PLAN.md) may span several commits, but each commit compiles and passes `make test`.
- Body references contract ids when relevant: `Implements FR-2.6, AC-12.`

## Misc

- Go formatted by `gofmt`; TS/Svelte by Prettier defaults (2-space). No lint suppressions without a comment naming why.
- No TODOs in merged code — unfinished work is a PLAN.md chunk or a backlog entry, not a comment.
- Dependencies: adding any new Go module or npm package requires an ARCHITECTURE.md "Committed stack" table update in the same PR.
