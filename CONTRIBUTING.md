# Contributing to Befrest

Thanks for helping improve Befrest. Please open an issue before making a substantial change so its scope can be discussed.

## Local setup

Install Go 1.25 and Node.js 24, then run:

```sh
npm --prefix web ci
npm --prefix e2e ci
make test
make build
npm --prefix e2e test
```

The end-to-end suite uses Playwright. If Chromium is not installed yet, run:

```sh
npx --prefix e2e playwright install chromium
```

Use conventional commits as described in [CONVENTIONS.md](CONVENTIONS.md), keep the living architecture and UX documents current when behavior changes, and include tests for changed behavior.
