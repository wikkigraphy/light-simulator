# Light Simulator

Professional-grade interactive photography lighting simulator built in Go. Design, visualize, and learn studio lighting setups with real-time CSS-based preview, drag-and-drop light placement, and comprehensive cheatsheets.

## Features

- **Interactive Simulator** — Top-down diagram with draggable lights, real-time subject preview
- **24 Professional Presets** — Rembrandt, butterfly, clamshell, split, high-key, low-key, beauty ring, cinematic noir, product hero, glassware, food moody, group, sport, and more — each with detailed equipment lists
- **Flash Settings Display** — Per-light flash details (type, modifier, role, power, color temp, CRI, distance, angle, height, grid, feathered) shown automatically when any preset is loaded
- **Custom Presets** — Save, load, rename, update, and delete your own presets stored in browser localStorage; persists across deployments until manually deleted
- **Full Modifier Library** — Softbox, octabox, stripbox, beauty dish, honeycomb grid, snoot, barn doors, parabolic reflector, diffusion panel, umbrella, reflector
- **Camera Controls** — Focal length, aperture, ISO, shutter speed, white balance, sensor size, camera distance and angle
- **Photo Upload** — Upload your own subject or product photo for live lighting preview
- **Lighting Engine** — Inverse-square law falloff, modifier softness modeling, CSS filter computation
- **Flash Guide** — Speedlight, monolight, pack & head, battery strobe, continuous LED, ring light
- **Modifier Guide** — Size ranges, softness ratings, spill control, catchlight shapes, pro tips
- **Lens Guide** — 24mm to 70-200mm with DOF notes, distortion characteristics, and best-use scenarios
- **Product Photography Mode** — Flat lay, hero shot, white background e-commerce setups
- **Scene Analysis** — Key-to-fill ratio, EV calculation, shadow quality, catchlight type, warnings

## Quick Start

```bash
# Clone and run
git clone https://github.com/srivickynesh/light-simulator.git
cd light-simulator
make run
# Open http://localhost:8080
```

## Requirements

| Dependency | Version | Purpose |
|-----------|---------|---------|
| Go | 1.26+ | Server runtime |
| Make | Any | Build automation |
| Docker | 20+ | Container builds (optional) |
| air | Latest | Live reload development (optional) |
| golangci-lint | v2+ (latest) | Linting (CI/local) |

## Project Structure

```
.
├── cmd/server/main.go              # Server entry point
├── internal/
│   ├── config/config.go            # Environment configuration
│   ├── models/lighting.go          # Domain models (Light, Camera, Scene, Preset)
│   ├── lighting/engine.go          # Lighting physics + CSS filter computation
│   ├── cheatsheet/
│   │   ├── presets.go              # 24 professional lighting presets with equipment lists
│   │   └── guides.go              # Flash, modifier, and lens guides
│   ├── handlers/
│   │   ├── api.go                  # JSON API endpoints
│   │   ├── pages.go                # HTML template rendering
│   │   └── upload.go               # Photo upload handler
│   └── middleware/middleware.go     # Logger, Recovery, CORS, Security
├── web/
│   ├── templates/                  # Go HTML templates
│   │   ├── layout.html             # Base layout
│   │   ├── index.html              # Landing page
│   │   ├── simulator.html          # Interactive simulator
│   │   └── cheatsheet.html         # Cheatsheet reference
│   └── static/
│       ├── css/main.css            # Dark-theme responsive styles
│       ├── js/
│       │   ├── simulator.js        # Simulator interaction engine
│       │   └── cheatsheet.js       # Cheatsheet data loader
│       └── images/                 # SVG placeholders
├── .github/
│   ├── dependabot.yml             # Dependabot: Actions, Go, Docker (weekly auto-bump)
│   └── workflows/
│       ├── ci.yml                 # CI pipeline (lint → test → build → docker)
│       ├── deploy.yml             # Vercel production deploy + old deployment cleanup
│       ├── auto-merge-dependabot.yml  # Auto-approve + merge Dependabot PRs
│       └── scheduled.yml          # Weekly cron: verification + security scan
├── Dockerfile                      # Multi-stage production build
├── Makefile                        # Build commands
├── vercel.json                     # Vercel deployment config (Go framework preset)
├── .golangci.yml                   # Linter configuration
├── .air.toml                       # Live reload config
└── .cursor/skills/                 # Cursor AI skill for dependency tracking
```

## Dependencies

### Go Modules

**Zero external dependencies.** The entire server uses Go standard library only:

- `net/http` — HTTP server (Go 1.22+ enhanced routing)
- `html/template` — Server-side rendering
- `encoding/json` — API serialization
- `log/slog` — Structured JSON logging
- `crypto/rand` — Secure filename generation
- `math` — Lighting physics (inverse-square law)
- `context`, `os/signal` — Graceful shutdown

### Frontend

**Zero npm dependencies.** Pure vanilla JavaScript and CSS:

- **Google Fonts** — Inter (UI text), JetBrains Mono (numeric values)
- SVG for diagram rendering and subject placeholder

### Build & CI

| Tool | Version | Config File | Purpose |
|------|---------|------------|---------|
| Go | 1.26 | `.tool-versions`, `go.mod` | Language runtime |
| golangci-lint | v2 (latest) | `.golangci.yml` | Static analysis (v2 config format) |
| GitHub Actions | — | `.github/workflows/*.yml` | CI/CD, deploy, auto-merge, scheduled |
| Docker | — | `Dockerfile` (golang:1.26-alpine) | Container builds |
| Make | — | `Makefile` | Task automation |
| air | Latest | `.air.toml` | Dev live reload |

**GitHub Actions versions:**

| Action | Version |
|--------|---------|
| `actions/checkout` | `v6` |
| `actions/setup-go` | `v6` |
| `golangci/golangci-lint-action` | `v9` |
| `actions/upload-artifact` | `v4` |
| `dependabot/fetch-metadata` | `v2` |

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/health` | Health check |
| `GET` | `/api/presets` | All presets by category |
| `GET` | `/api/presets/{id}` | Single preset |
| `POST` | `/api/analyze` | Analyze a lighting scene (JSON body) |
| `GET` | `/api/guides/flash` | Flash type guide |
| `GET` | `/api/guides/modifiers` | Modifier guide |
| `GET` | `/api/guides/lenses` | Lens guide |
| `POST` | `/api/upload` | Upload subject photo (multipart) |

## Configuration

All configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `HOST` | `0.0.0.0` | Listen address |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `LOG_LEVEL` | `info` | `info` or `debug` |
| `MAX_UPLOAD_MB` | `10` | Maximum upload file size |
| `STATIC_DIR` | `web/static` | Static files directory |
| `TEMPLATE_DIR` | `web/templates` | HTML templates directory |
| `UPLOAD_DIR` | `uploads` | Upload storage directory |
| `READ_TIMEOUT` | `15` | HTTP read timeout (seconds) |
| `WRITE_TIMEOUT` | `15` | HTTP write timeout (seconds) |

## Development

```bash
# Pre-deploy verification (format check → vet → lint → test → build)
make update

# Run with live reload
make dev

# Run tests
make test

# Lint
make lint

# Format code
make fmt

# Build binary
make build

# Docker build + run
make docker-run
```

**Always run `make update` before deploying.** This is the same command used in CI.

## Deployment

### Vercel (Native Go Support)

The project uses Vercel's native Go framework preset (`"framework": "go"` in `vercel.json`).
Vercel automatically detects `cmd/server/main.go`, installs the Go version from `go.mod`,
builds the binary, and deploys it. The server listens on the `PORT` environment variable
set by Vercel at runtime.

```bash
vercel deploy
```

### Docker

```bash
docker build -t light-simulator .
docker run -p 8080:8080 -e ENVIRONMENT=production light-simulator
```

## CI/CD & Automation

### Fully Automated Infrastructure

The project runs with zero-touch dependency management, deployment, and cleanup:

```
Dependabot opens PR → CI runs make update → Auto-approve → Squash merge → Deploy to Vercel → Cleanup old deployments
```

### Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push/PR to main | `make update` + Docker build + smoke test |
| `deploy.yml` | Push to main | Vercel production deploy + old deployment cleanup |
| `auto-merge-dependabot.yml` | Dependabot PR | Full CI + Docker smoke → auto-approve → squash merge |
| `scheduled.yml` | Weekly (Monday 06:00 UTC) | Full verification + Docker health check + `govulncheck` |

### Dependabot

Configured in `.github/dependabot.yml` to automatically bump:

- **GitHub Actions** — All action versions (checkout, setup-go, etc.)
- **Go modules** — Any future external Go dependencies
- **Docker** — Base image versions (golang, alpine)

Schedule: Weekly on Monday at 06:00 UTC. PRs are auto-merged after all CI passes.

### Auto-Merge Flow

1. Dependabot opens a PR with the version bump
2. `auto-merge-dependabot.yml` runs the full CI suite (`make update` + Docker smoke test)
3. If all checks pass, the PR is auto-approved and squash-merged
4. The merge to `main` triggers `deploy.yml` which deploys to Vercel production
5. After successful deploy, old non-production deployments are cleaned up (keeps 3 most recent)

### Required GitHub Secrets

| Secret | Purpose |
|--------|---------|
| `VERCEL_TOKEN` | Vercel API token for deploy and cleanup |
| `VERCEL_ORG_ID` | Vercel organization/team ID |
| `VERCEL_PROJECT_ID` | Vercel project ID |

### Required GitHub Settings

1. **Branch protection on `main`**: Require `Verify (make update)` status check
2. **Allow auto-merge**: Settings → General → Pull Requests → Enable auto-merge
3. **Dependabot security updates**: Settings → Code security → Enable

### Weekly Cron

Runs every Monday even without code changes:

- Full `make update` verification (fmt → vet → lint → 148 tests → build)
- Docker image build + health check (starts container, curls `/api/health`)
- `govulncheck` security scan for known Go vulnerabilities
- Secret file detection in git history

## Adding Features

1. Add domain types to `internal/models/lighting.go`
2. Implement logic in the appropriate `internal/` package
3. Create handler in `internal/handlers/`
4. Register routes via `RegisterRoutes(mux)`
5. Write tests in `*_test.go`
6. Update `.cursor/skills/light-simulator-deps/SKILL.md`
7. Update this README's dependency and API tables
8. Run `make update` to verify everything passes

## License

MIT
