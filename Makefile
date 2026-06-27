# Makefile — Unified build/test contract for all agents
.PHONY: build build-tui run-tui run-demo run-daemon run-hub run-helper tui-snapshot dev-certs test-acceptance release test test-integration test-e2e test-contract test-all lint security-scan fmt containers-up containers-down migrate test-features-sync test-features-scaffold sdlc-report sdlc-rotate-logs sdlc-branch-protection sdlc-bootstrap-project-fields help

## Build
build:
	go build ./...

## Build the Heimdall binaries (dashboard + daemon + hub) -> bin/
build-tui:
	@mkdir -p bin
	go build -o bin/heimdall-dashboard ./app/cmd/dashboard
	go build -o bin/heimdall-daemon ./app/cmd/daemon
	go build -o bin/heimdall-hub ./app/cmd/hub
	go build -o bin/heimdall-helper ./app/cmd/helper
	@echo "built bin/heimdall-dashboard, bin/heimdall-daemon, bin/heimdall-hub, bin/heimdall-helper"

## Run the Heimdall TUI dashboard (subscribes to a hub at localhost:9090)
run-tui: build-tui
	./bin/heimdall-dashboard

## Run the dashboard with a simulated multi-host fleet (no hub needed)
run-demo: build-tui
	./bin/heimdall-dashboard --demo

## Run the daemon (samples and prints this machine's metrics)
run-daemon: build-tui
	./bin/heimdall-daemon

## Run the hub (central gRPC server)
run-hub: build-tui
	./bin/heimdall-hub

## Run the privileged helper (power/GPU/thermal over a local socket; needs root for real data)
run-helper: build-tui
	./bin/heimdall-helper

## Acceptance tests (behave) — runnable story scenarios driving the real binaries
test-acceptance: build-tui
	behave tests/features

## Render one dashboard frame to stdout (no TTY required)
tui-snapshot: build-tui
	@./bin/heimdall-dashboard --demo --snapshot

## Generate a self-signed TLS cert for local hub development -> certs/
dev-certs:
	@bash scripts/gen-dev-certs.sh

## Cross-compile release binaries for all platforms -> dist/
release:
	@bash scripts/release.sh

## Unit tests
test:
	go test ./...

## Integration tests (requires containers)
test-integration:
	@echo "Override this target with your project's integration test command"
	@echo "Examples:"
	@echo "  dotnet test --filter 'Category=Integration'"
	@echo "  mvn test -Dgroups=integration"
	@echo "  npx vitest run tests/integration"
	@echo "  pytest tests/integration"

## E2E tests
test-e2e:
	npx playwright test tests/e2e/ --reporter=html

## Contract tests
test-contract:
	@echo "Override this target with your project's contract test command"
	@echo "Example: npx playwright test tests/contract/"

## Run all tests (gate for Phase 8)
test-all: test test-integration test-e2e test-contract
	@echo "All test suites passed."

## Linting
lint:
	@test -z "$$(gofmt -l app common 2>/dev/null)" || { echo "gofmt needed in:"; gofmt -l app common; exit 1; }
	go vet ./...

## Security scanning
security-scan:
	@echo "Override this target with your project's security scan command"
	@echo "Examples:"
	@echo "  dotnet-sonarscanner"
	@echo "  npm audit --audit-level=high"
	@echo "  trivy fs ."

## Format code
fmt:
	gofmt -w app common

## Start test containers
containers-up:
	docker compose -f docker-compose.test.yml up -d --wait

## Stop test containers
containers-down:
	docker compose -f docker-compose.test.yml down -v

## Run database migrations
migrate:
	@echo "Override this target with your project's migration command"
	@echo "Examples:"
	@echo "  dotnet ef database update"
	@echo "  mvn flyway:migrate"
	@echo "  npx prisma migrate deploy"
	@echo "  supabase db push"

## Sync Gherkin from docs/stories/ → tests/features/ (idempotent)
test-features-sync:
	@echo "Extracting Gherkin from story files..."
	@python3 -c " \
import os, re, pathlib; \
stories = pathlib.Path('docs/stories').glob('*.md'); \
out = pathlib.Path('tests/features'); out.mkdir(parents=True, exist_ok=True); \
count = 0; \
for s in stories: \
    text = s.read_text(); \
    blocks = re.findall(r'\`\`\`gherkin\n(.*?)\`\`\`', text, re.DOTALL); \
    if not blocks: continue; \
    name = s.stem + '.feature'; \
    (out / name).write_text('\n\n'.join(blocks)); \
    count += 1; \
print(f'Synced {count} feature file(s) to tests/features/') \
"

## Generate step definition stubs for scenarios not yet covered
test-features-scaffold:
	@echo "Step definition scaffolding — run from Claude Code:"
	@echo "  @qa-cucumber scaffold-steps"
	@echo "Or with the cucumber CLI for your stack:"
	@echo "  npx cucumber-js --dry-run --format usage"
	@echo "  behave --dry-run"

# ─── BEGIN MAPLE MANAGED — updated by `maple update`, do not hand-edit ────────

## Print per-story agent invocation counts and estimated costs
## Reads .claude/logs/skills.jsonl
sdlc-report:
	@if [ ! -f .claude/logs/skills.jsonl ]; then \
		echo "No skills log found. Run some agent workflows first."; exit 0; \
	fi
	@echo "=== SDLC Cost Report ==="
	@python3 -c " \
import json, collections; \
lines = [json.loads(l) for l in open('.claude/logs/skills.jsonl') if l.strip()]; \
by_story = collections.defaultdict(list); \
[by_story[l.get('story','unknown')].append(l) for l in lines]; \
print(f'Stories: {len(by_story)}  Total invocations: {len(lines)}'); \
[print(f'  {s}: {len(v)} invocations') for s,v in sorted(by_story.items())] \
"

## Rotate .claude/logs/ — keep last 5 compressed, delete older
sdlc-rotate-logs:
	@bash scripts/sdlc/rotate-logs.sh

## Apply branch protection rules to main (requires gh admin scope)
sdlc-branch-protection:
	@bash scripts/sdlc/branch-protection.sh

## Show available targets
help:
	@echo "Available make targets:"
	@echo "  build                  - Build the project"
	@echo "  test                   - Run unit tests"
	@echo "  test-integration       - Run integration tests (requires containers)"
	@echo "  test-e2e               - Run E2E tests with Playwright"
	@echo "  test-contract          - Run contract/schema tests"
	@echo "  test-all               - Run all test suites (Phase 8 gate)"
	@echo "  test-features-sync     - Extract Gherkin from stories → tests/features/"
	@echo "  test-features-scaffold - Generate step definition stubs for new scenarios"
	@echo "  lint                   - Run linter"
	@echo "  security-scan          - Run security scanning"
	@echo "  fmt                    - Format code"
	@echo "  containers-up          - Start test containers"
	@echo "  containers-down        - Stop and remove test containers"
	@echo "  migrate                - Run database migrations"
	@echo "  sdlc-report            - Print per-story cost + invocation report"
	@echo "  sdlc-rotate-logs       - Rotate .claude/logs/ (keep last 5 compressed)"
	@echo "  sdlc-branch-protection - Apply branch protection rules to main"

# ─── END MAPLE MANAGED ────────────────────────────────────────────────────────
