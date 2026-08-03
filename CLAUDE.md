# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

A serverless Go function deployed on DigitalOcean's Functions platform (Apache OpenWhisk). It redirects short URLs of the form `/<version>/<fragment>` to LLVM/Clang diagnostic documentation at `releases.llvm.org`.

Example: `/13/wall` → `https://releases.llvm.org/13.0.0/tools/clang/docs/DiagnosticsReference.html#wall`

## Commands

All commands run from `packages/pxy/redirect/`:

```sh
# Build
go build ./...

# Test
go test ./...

# Run a single test
go test -run TestFunctionName ./...

# Static analysis
staticcheck ./...

# Vet
go vet ./...
```

## Deployment

Deployed via `doctl` (DigitalOcean CLI):

```sh
doctl serverless deploy .
```

CI/CD runs on push to `main` via `.github/workflows/main-deploy-production.yml`. The workflow installs `doctl`, connects to the serverless namespace, and deploys.

## Architecture

**Single action**: `packages/pxy/redirect/redirect.go` — the entire function lives here.

**OpenWhisk entry point**: `Main(args map[string]interface{}) Response` — not `main()`. OpenWhisk calls `Main` with request data in the `args` map. The standard `main()` is commented out but can be uncommented for local testing.

**Request flow**:
1. Extract `__ow_path` and `__ow_headers` from `args`
2. `parseRedirectURL` — validates the raw path via `url.Parse`
3. `assembleTargetURL` — splits path into `[version, fragment]`, validates inputs, handles LLVM version quirks, builds redirect URL
4. Emit heartbeat to monitoring service (`emitHeartbeat`)
5. Return `308 Permanent Redirect` with `Location` header, or `400`/`500` on error

**LLVM version quirks** (encoded as hacks in `assembleTargetURL`):
- Version 17 → `17.0.1` (not `17.0.0`)
- Version 18+ → `X.1.0` pattern (e.g., `18.1.0`)

**Fragment normalization**: `+` in fragments is replaced with `-`, and `--` is collapsed to `-` (handles flags like `-Wc++98-c++11-compat-binary-literal`).

## Environment Variables

| Variable | Purpose |
|---|---|
| `LOG_LEVEL` | Set to `debug` to enable debug logging (default: INFO) |
| `LOG_DESTINATIONS` | Log shipping destination (Better Stack / Logtail) |
| `HEARTBEAT_TOKEN` | Auth token for Better Uptime heartbeat |
| `HEARTBEAT_TARGET` | Heartbeat endpoint URL |
| `HEARTBEAT_TARGET_TIMEOUT` | HTTP timeout in seconds for heartbeat (default: 10) |

Configured in `.env` (gitignored) locally and via GitHub Actions secrets/vars in production.

## Static Analysis

`staticcheck.conf` enables all checks except `ST1005` (error string casing). Run `staticcheck ./...` from the package directory.
