---
name: light-simulator-deps
description: >-
  Dependency and architecture tracker for the Light Simulator Go application.
  Use when modifying, extending, or debugging the photography lighting simulator.
  Tracks all Go packages, frontend assets, build tools, CI configuration, and
  inter-module dependencies. Use when adding features, updating dependencies,
  refactoring code, or troubleshooting build/CI issues.
---

# Light Simulator — Dependency & Architecture Guide

## Project Overview

Go-based web application providing an interactive photography lighting simulator
with real-time preview, professional presets, and comprehensive cheatsheets.

## Architecture

```
cmd/server/main.go                → Entry point, server bootstrap, graceful shutdown
internal/
  config/config.go                → Environment-based configuration (PORT, HOST, etc.)
  config/config_test.go           → Config unit tests (11 tests)
  models/lighting.go              → Domain types: Light, Camera, Scene, Preset, EquipmentItem, Position3D
  models/lighting_test.go         → Model JSON serialization & constant tests (10 tests)
  lighting/engine.go              → Lighting computation engine (inverse-square, CSS filters)
  lighting/engine_test.go         → Engine unit tests: contributions, shadows, catchlights, warnings (27 tests)
  cheatsheet/presets.go           → 24 professional lighting presets across 7 categories, each with equipment lists
  cheatsheet/presets_test.go      → Preset validation tests (13 tests)
  cheatsheet/guides.go            → Flash, modifier, and lens guide data
  cheatsheet/guides_test.go       → Guide data validation tests (8 tests)
  handlers/api.go                 → JSON API endpoints (/api/*)
  handlers/api_test.go            → API handler tests (19 tests)
  handlers/pages.go               → HTML template rendering with caching
  handlers/upload.go              → Photo upload with bg removal processing
  handlers/upload_test.go         → Upload handler tests (11 tests)
  imgproc/bgremove.go             → Background removal: edge-aware chroma keying, transparent PNG output
  imgproc/bgremove_test.go        → Image processing tests (13 tests)
  middleware/middleware.go         → Logger, Recovery, CORS, SecurityHeaders
  middleware/middleware_test.go    → Middleware unit tests (7 tests)
  e2e_test.go                     → End-to-end integration tests (19 tests)
web/
  templates/                      → Go html/template files (layout.html + pages)
  static/css/main.css             → Dark-theme responsive CSS
  static/js/simulator.js          → Interactive light placement, drag, preview, URL preset loading, flash settings display, custom preset save/load/rename/delete (localStorage)
  static/js/cheatsheet.js         → Cheatsheet tab loading, clickable preset navigation
  static/images/default-subject.png → Real portrait photo for live preview
  static/images/subject-placeholder.svg → SVG fallback
```

## Dependencies

### Go (Standard Library Only — Zero External Dependencies)
- `net/http` — HTTP server with Go 1.22+ routing patterns
- `html/template` — Server-side template rendering
- `encoding/json` — JSON API serialization
- `log/slog` — Structured logging (Go 1.21+)
- `crypto/rand` — Secure random filename generation
- `math` — Lighting physics calculations
- `image`, `image/color`, `image/png`, `image/jpeg` — Server-side background removal
- `context`, `os/signal` — Graceful shutdown

### Frontend (Zero Dependencies — Vanilla JS)
- No npm packages. Pure CSS + vanilla JavaScript.
- Google Fonts: Inter (UI), JetBrains Mono (code/values)

### Build & CI Tools
- **Go 1.26** — Language runtime (set in `.tool-versions`)
- **golangci-lint v2** — Go linting, v2 config format (CI uses `golangci/golangci-lint-action@v9`)
- **Docker** — Multi-stage build (golang:1.26-alpine → alpine:3.21)
- **Make** — Build automation (`Makefile`), `make update` runs full pre-deploy verification
- **air** — Optional live-reload for development
- **GitHub Actions** — CI/CD pipeline (see CI/CD section below)
  - `actions/checkout@v6`
  - `actions/setup-go@v6`
  - `golangci/golangci-lint-action@v9`
  - `actions/upload-artifact@v4`
  - `dependabot/fetch-metadata@v2`

### Deployment
- **Vercel** — Production deployment using Vercel's native Go framework preset (`"framework": "go"`
  in `vercel.json`). Vercel auto-detects `cmd/server/main.go` and builds Go natively.
  Server must listen on `PORT` env var (already configured in `config.go`).
- **Docker** — Container deployment alternative

## CI/CD & Automation

### Workflow Files

```
.github/
  dependabot.yml                   → Dependabot config (Actions, Go, Docker — weekly Monday 06:00 UTC)
  workflows/
    ci.yml                         → Main CI: make update + Docker build + smoke test (push/PR)
    deploy.yml                     → Vercel production deploy + old deployment cleanup (main push)
    auto-merge-dependabot.yml      → Auto-approve + squash-merge Dependabot PRs after CI passes
    scheduled.yml                  → Weekly cron: full verification + Docker + security scan
```

### Automation Pipeline

1. **Dependabot** opens PR when github-actions / gomod / docker deps have updates
2. **CI** (`ci.yml`) runs `make update` (fmt → vet → lint → test → build) on the PR
3. **Auto-merge** (`auto-merge-dependabot.yml`) re-runs full CI + Docker smoke test,
   then auto-approves and enables squash-merge — only if all checks pass
4. **Deploy** (`deploy.yml`) triggers on main push: verifies, deploys to Vercel production,
   then cleans up old non-production deployments (keeps 3 most recent)
5. **Scheduled** (`scheduled.yml`) runs every Monday: full verification, Docker build
   with health check, and `govulncheck` security scan

### Required GitHub Secrets

| Secret | Purpose |
|--------|---------|
| `VERCEL_TOKEN` | Vercel API token for deploy/cleanup |
| `VERCEL_ORG_ID` | Vercel organization/team ID |
| `VERCEL_PROJECT_ID` | Vercel project ID |
| `ORG_TOKEN` | GitHub PAT for Dependabot auto-merge (approve + merge PRs) |

### Required GitHub Settings

- **Branch protection on `main`**: Require status checks `Verify (make update)` to pass
- **Allow auto-merge**: Enable in repo Settings → General → Pull Requests
- **Dependabot security updates**: Enable in repo Settings → Security

## API Endpoints

| Method | Path                   | Handler              | Description                    |
|--------|------------------------|----------------------|--------------------------------|
| GET    | /                      | pages.Index          | Landing page                   |
| GET    | /simulator             | pages.Simulator      | Interactive simulator          |
| GET    | /cheatsheet            | pages.Cheatsheet     | Reference cheatsheets          |
| GET    | /api/presets           | api.ListPresets      | All presets by category        |
| GET    | /api/presets/{id}      | api.GetPreset        | Single preset by ID            |
| POST   | /api/analyze           | api.AnalyzeScene     | Compute lighting analysis      |
| GET    | /api/guides/flash      | api.FlashGuide       | Flash selection guide          |
| GET    | /api/guides/modifiers  | api.ModifierGuide    | Modifier cheatsheet            |
| GET    | /api/guides/lenses     | api.LensGuide        | Lens selection guide           |
| GET    | /api/health            | api.Health           | Health check                   |
| POST   | /api/upload            | upload.HandleUpload  | Photo upload                   |
| GET    | /uploads/*             | FileServer           | Serve uploaded files           |
| GET    | /static/*              | FileServer           | Serve static assets            |

## Lighting Presets (24 total)

| ID | Name | Category |
|----|------|----------|
| rembrandt | Rembrandt Lighting | portrait |
| butterfly | Butterfly / Paramount | portrait |
| split | Split Lighting | portrait |
| loop | Loop Lighting | portrait |
| clamshell | Clamshell Lighting | portrait |
| broad | Broad Lighting | portrait |
| short | Short Lighting | portrait |
| high_key | High-Key Portrait | portrait |
| low_key | Low-Key Portrait | portrait |
| beauty_ring | Beauty Ring Light + Accents | portrait |
| cinematic_noir | Cinematic Film Noir | portrait |
| cross_light | Cross Lighting (Dual Key) | portrait |
| product_topdown | Product Flat Lay / Top-Down | product |
| product_hero | Product Hero Shot | product |
| product_white_bg | Product on White (E-Commerce) | product |
| product_glass | Product Glassware / Bottles | product |
| fashion_editorial | Fashion Editorial | fashion |
| fashion_catalog | Fashion Catalog (Clean) | fashion |
| food_moody | Food Photography (Dark & Moody) | food |
| food_bright | Food Photography (Bright & Airy) | food |
| headshot_corporate | Corporate Headshot | headshot |
| rim_dramatic | Dramatic Rim / Edge Lighting | portrait |
| group_photo | Group / Team Photo | group |
| sport_action | Sport / Action Portrait | sport |

## Test Coverage (151 tests)

| Package | Test File | Tests | Coverage Areas |
|---------|-----------|-------|----------------|
| config | config_test.go | 11 | Defaults, env vars, port validation, Addr(), IsProd() |
| models | lighting_test.go | 10 | Type constants, JSON serialization, Equipment, field presence |
| lighting | engine_test.go | 27 | Inverse-square, modifiers, shadows, catchlights, warnings, CSS filters |
| cheatsheet | presets_test.go | 14 | Count, required fields, unique IDs, categories, equipment lists, specific presets |
| cheatsheet | guides_test.go | 8 | Flash/modifier/lens guides validation |
| handlers | api_test.go | 19 | All API endpoints, error cases, response structure |
| handlers | upload_test.go | 11 | Upload success, bg removal, extensions, serve files |
| imgproc | bgremove_test.go | 13 | BG removal, color distance, edge weight, file/stream I/O, JPEG |
| middleware | middleware_test.go | 7 | Chain order, logger, recovery, CORS, security headers |
| e2e | e2e_test.go | 22 | Full workflow, all endpoints, upload+serve, static files, headers, flash settings, custom preset UI |

## Key Dependency Rules

1. **Zero external Go modules** — Only stdlib. This is intentional for reliability.
2. **Any new Go dependency** must be added to `go.mod` and documented here.
3. **Frontend must remain vanilla JS** — No build step, no npm.
4. **When adding API endpoints**: Update this skill's API table.
5. **When adding models**: Update `internal/models/lighting.go`.
6. **When adding presets**: Add to `internal/cheatsheet/presets.go` with `Equipment` list,
   register in `AllPresets()`, add to the preset table above, and add a test.
7. **When modifying CI/CD**: Update workflows in `.github/workflows/` and this file.
   Dependabot auto-bumps Actions/Go/Docker deps weekly with auto-merge after CI passes.
8. **Cheatsheet → Simulator**: Preset cards link via `?preset={id}` query param.
   The simulator reads `?preset=` on load and applies the preset automatically.
9. **Custom presets**: Saved in `localStorage` under key `light_sim_custom_presets`.
   Users can save, load, rename, update, and delete custom presets.
   Custom presets appear in the "My Presets" optgroup in the preset dropdown.
   They persist across deployments until the user explicitly deletes them.
10. **Flash settings display**: When a preset is loaded, the analysis panel shows a
   detailed Flash & Camera Settings table (per-light: type, modifier, power, temp, CRI,
   distance, angle, height, grid, feathered) plus the equipment list.
11. **Default subject image**: `web/static/images/default-subject.png` (real portrait photo).
10. **Upload flow**: POST `/api/upload` → saves original → `imgproc.RemoveBackground()` →
    outputs transparent PNG → returns processed URL. Client shows loading state during processing.
11. **Background removal**: `internal/imgproc/bgremove.go` uses edge-aware chroma keying:
    samples corners for background color, applies inverse-square distance with soft edge blending.

## Configuration (Environment Variables)

| Variable      | Default      | Description                |
|---------------|-------------|----------------------------|
| PORT          | 8080        | Server port                |
| HOST          | 0.0.0.0     | Listen address             |
| STATIC_DIR    | web/static  | Static file directory      |
| TEMPLATE_DIR  | web/templates | Template directory       |
| UPLOAD_DIR    | uploads     | Upload storage directory   |
| MAX_UPLOAD_MB | 10          | Max upload size in MB      |
| ENVIRONMENT   | development | development / production   |
| LOG_LEVEL     | info        | info / debug               |
| READ_TIMEOUT  | 15          | HTTP read timeout seconds  |
| WRITE_TIMEOUT | 15          | HTTP write timeout seconds |

## Adding New Features

1. Add models to `internal/models/lighting.go`
2. Add business logic to appropriate `internal/` package
3. Add handler in `internal/handlers/`
4. Register route in the handler's `RegisterRoutes` method
5. Update tests in corresponding `*_test.go` files
6. Update this skill file's dependency table, API table, and preset table
7. Run `make update` before committing (runs fmt-check → vet → lint → test → build)
