# Releasing Heimdall

Heimdall ships four binaries — `heimdall-hub`, `heimdall-dashboard`, `heimdall-daemon`,
`heimdall-helper` — built for Linux, macOS, and Windows on `amd64` and `arm64`.

## Versioning

- Semantic version tags: `vMAJOR.MINOR.PATCH` (e.g. `v1.2.0`).
- The gRPC contract is versioned independently under `common/proto/monitoring/v1/`.
  Wire changes must stay backwards compatible within a major proto version: only add
  fields, never renumber or repurpose them. A breaking change bumps the proto package
  to `v2` and lives alongside `v1` until daemons have migrated.

## Cutting a release

CI never creates the release — it reacts to one. Publishing a GitHub Release fires the
[`release` workflow](../../.github/workflows/release.yml), which cross-compiles every
binary and attaches them to that release. Re-running is therefore idempotent.

```sh
# from an up-to-date main
gh release create v1.2.0 --target main --title v1.2.0 --generate-notes
```

The workflow then, in parallel:

1. **macOS** (`macos-14`, **CGO enabled**) builds `darwin/arm64` and `darwin/amd64`
   so the Apple Silicon IOReport path — no-sudo GPU utilisation and power — is
   compiled in. CGO-free cross-compiles cannot link IOReport, so these must be
   built on a macOS runner.
2. **Linux + Windows** (`ubuntu-latest`, CGO-free) cross-compiles `linux/{amd64,arm64}`
   and `windows/{amd64,arm64}` via `scripts/release.sh`.
3. a **publish** job downloads both, writes `SHA256SUMS`, and attaches everything
   to the release with `softprops/action-gh-release`.

> `make release` builds Linux + Windows CGO-free and, **when run on a Mac**,
> builds the darwin binaries with CGO so IOReport is compiled in. Run off a Mac it
> skips the darwin targets (a CGO-free darwin build would silently lose macOS
> power) — the official macOS binaries always come from the CI mac runner.

To rebuild assets for an existing release, re-publish it (or re-run the workflow) — the
upload overwrites in place.

## What consumes the assets

- `scripts/install.sh` downloads the per-binary asset for the host's OS/arch and verifies
  it against `SHA256SUMS`.
- `heimdall-<binary> update` self-updates against the latest release using the same assets.
