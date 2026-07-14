# Changelog

All notable changes to Befrest are documented here. Releases follow [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.1.1] - 2026-07-14

### Added

- README screenshots covering connection, connected devices, and incoming-file confirmation.

### Fixed

- Preserve files selected on a phone while its native file picker temporarily disconnects the WebSocket; send the offer after reconnecting.
- Update Go networking dependencies for compatibility with Go 1.23 release builds.

## [0.1.0] - 2026-07-14

### Added

- Direct local-network file transfers between browser clients through a small host binary.
- QR-code and mDNS-based joining, device presence, transfer approval, progress, cancellation, and network-interface selection.
- Native release binaries for Linux amd64, macOS amd64/arm64, and Windows amd64.
