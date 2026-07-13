# Releasing Befrest

This repository publishes releases from annotated Git tags. The GitHub Actions release workflow builds the native artifacts on their target operating systems, then creates the GitHub release and its generated notes.

## Before a release

1. Ensure `main` is clean and all intended changes are merged.
2. Update `CHANGELOG.md`: move notable entries from **Unreleased** to a dated version section.
3. Run the same checks used by CI:

   ```sh
   npm --prefix web ci
   npm --prefix e2e install --no-package-lock --no-audit --no-fund
   make test
   make build
   npm --prefix e2e test
   ```

4. Confirm the release version follows Semantic Versioning and has no existing tag.

## Publish

Create and push an annotated version tag. For example:

```sh
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin main
git push origin v0.1.0
```

Pushing a `v*` tag starts the **Release** workflow. It uploads these binaries to the GitHub release:

- `befrest-linux-amd64`
- `befrest-darwin-amd64`
- `befrest-darwin-arm64`
- `befrest-windows-amd64.exe`

It also uploads `SHA256SUMS.txt`. Check the workflow and download each artifact before announcing the release. If a release must be rebuilt, delete the GitHub release and remote tag first, then create a new annotated tag after fixing the cause.
