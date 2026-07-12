# Befrest — FILE_STRUCTURE

Living document. Every file that will exist at v1. New files require updating this doc in the same PR (CONVENTIONS folder rules).

```
befrest/
├── ARCHITECTURE.md                      living doc — technical truth
├── CONVENTIONS.md                       living doc — how code is written
├── DESIGN.md                            living doc — design-system contract
├── FILE_STRUCTURE.md                    living doc — this file
├── UX.md                                living doc — screens & flows
├── README.md                            quickstart + flags table (PLAN chunk 13)
├── .gitignore                           ignores web/dist, web/node_modules, dist/
├── Makefile                             web · build · test · e2e · release
├── go.mod                               module github.com/nabidam/befrest, Go 1.23
├── go.sum
│
├── specs/
│   └── 001-core/
│       ├── SPEC.md                      phase 1 — frozen for the cycle
│       ├── PRD.md                       phase 2 — frozen for the cycle
│       ├── PLAN.md                      phase 3 — this cycle's plan
│       └── TASKS.md                     phase 4 output (does not exist yet)
│
├── cmd/
│   └── befrest/
│       ├── main.go                      flags, port scan, netinfo wiring, browser open
│       └── tray.go                      systray menu: open page, open log, quit
│
├── internal/
│   ├── proto/
│   │   └── messages.go                  WS wire types — single source of truth
│   ├── presence/
│   │   ├── registry.go                  Device registry: join/leave/rename, dedup, fanout
│   │   └── registry_test.go
│   ├── transfer/
│   │   ├── manager.go                   Transfer lifecycle state machine, progress
│   │   ├── manager_test.go
│   │   ├── pipe.go                      4 MiB bounded rendezvous pipe, 30 s deadlines
│   │   └── pipe_test.go
│   ├── netinfo/
│   │   ├── netinfo.go                   interface enumeration + ranking, mDNS announce
│   │   ├── netinfo_test.go
│   │   └── probe.go                     reachability self-probe (FR-7.3)
│   └── server/
│       ├── server.go                    mux, static SPA serving
│       ├── server_test.go
│       ├── ws.go                        /ws endpoint; wire ↔ presence/transfer translation
│       ├── ws_test.go
│       ├── files.go                     POST/GET /api/transfers/{tid}/files/{idx}
│       └── files_test.go
│
├── web/
│   ├── embed.go                         package web; //go:embed all:dist
│   ├── package.json                     svelte, vite, typescript, qrcode, vitest
│   ├── package-lock.json
│   ├── vite.config.ts
│   ├── svelte.config.js
│   ├── tsconfig.json
│   ├── index.html
│   ├── dist/                            generated — embedded into the binary, git-ignored
│   └── src/
│       ├── main.ts                      mounts App
│       ├── App.svelte                   S1 ↔ S2 switch; hosts layers
│       ├── styles/
│       │   ├── tokens.css               DESIGN.md §2 verbatim — the only file with raw values
│       │   └── base.css                 reset + element defaults, tokens only
│       ├── lib/
│       │   ├── proto.ts                 wire types — mirror of internal/proto
│       │   ├── ws.ts                    connection, hello, reconnect backoff, store mutation
│       │   ├── stores.ts                connection, self, devices, offers, transfers, invite, toasts
│       │   ├── upload.ts                sequential per-file POST
│       │   ├── download.ts              anchor download on file-ready
│       │   ├── format.ts                byte formatting + reason→copy table
│       │   └── format.test.ts
│       └── screens/                     one component per UX.md id
│           ├── NameScreen.svelte        S1
│           ├── MainScreen.svelte        S2 shell
│           ├── Header.svelte            S2 region 1 — self name, inline rename
│           ├── ConnectionBanner.svelte  B1
│           ├── ReceiveBanner.svelte     M2 receiver side
│           ├── DeviceGrid.svelte        S2 region 3
│           ├── DeviceCard.svelte        card + M2 sender side + drag-drop
│           ├── EmptyInvite.svelte       S2 empty state (S3 content inline)
│           ├── Footer.svelte            S2 region 4 — Add device
│           ├── InviteSheet.svelte       S3
│           ├── OfferModal.svelte        M1
│           ├── InterfacePickerModal.svelte  M3
│           └── ToastHost.svelte         B2
│
└── e2e/
    ├── package.json                     @playwright/test
    ├── package-lock.json
    ├── playwright.config.ts             launches built hub with --no-open --no-mdns
    ├── kernel.spec.ts                   join, presence, offer/accept, transfer, decline, rejoin
    └── failures.spec.ts                 cancels, disconnects, stall verdicts
```

Not in the tree by design: no config files (ARCHITECTURE §8 zero-config), no database, no CI directory (v1 ships from `make release`), no `web/src/lib/utils.ts` (CONVENTIONS).
