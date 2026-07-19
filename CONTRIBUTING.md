# Contributing to athenzctl

## Development prerequisites

- Go version declared in `go.mod`
- Docker and Docker Compose for the end-to-end environment
- KinD and `kubectl` for the local Athenz distribution used by E2E tests
- Microsoft APM CLI for the repository's AI development context

The repository is a Go CLI. Start with `README.md` for user-facing behavior and use the package layout and test files as the implementation reference.

## AI context setup

APM source files under `.apm/` are the canonical source for repository-wide AI instructions and the review agent. Generated files must not be edited by hand. The manifest intentionally does not pin a default target; choose the AI client(s) needed for your environment.

Install APM using its official installation method, then run the target-specific flow. For example, for Codex:

```sh
apm install --target codex
apm compile --target codex --clean
```

Choose several clients with a comma-separated target list:

```sh
apm install --target codex,copilot
apm compile --target codex,copilot --clean
```

Use `--target all` only when every supported client is required:

```sh
apm install --target all
apm compile --all --clean
```

`apm install` deploys the local APM primitives and dependencies for the selected targets. `apm compile` generates aggregate context files such as `AGENTS.md`, `CLAUDE.md`, and `GEMINI.md`, plus client-specific rule files. Running bare `apm install` leaves target resolution to APM auto-detection and is not the recommended setup command for a clean checkout.

For a clean, reproducible checkout where `apm.lock.yaml` is present, retain the same target selection:

```sh
apm install --target codex --frozen
apm compile --target codex --clean
```

When changing `.apm/` content, validate it before committing:

```sh
apm compile --validate --all
apm audit --ci
```

Commit `apm.yml`, `apm.lock.yaml` when present, and `.apm/`. Generated client context is target-specific, is covered by `.gitignore`, and should remain local; do not hand-edit it. Change the `.apm/` source and regenerate the selected target instead. Do not commit `apm_modules/`.

## Go development

Run the fast checks from the repository root:

```sh
make test
make lint
```

Use `go test ./...` for the unit-test suite and `go vet ./...` when a package-level check is useful. Keep tests close to the package they exercise and prefer table-driven tests for parsing and configuration behavior.

The main code areas are:

- `cmd/athenzctl`: executable entry point
- `internal/cli`: Cobra command groups and resource handlers
- `internal/client`: authenticated ZMS/ZTS HTTP clients
- `internal/config`: kubeconfig-style context persistence
- `internal/cliopts`: shared flags and context resolution
- `internal/printer`: table, JSON, YAML, and pretty output
- `internal/resource`: canonical resource kinds and reference parsing
- `test/e2e`: Gherkin scenarios and step definitions

When changing CLI flags, config fields, authentication, TLS/proxy behavior, output formats, or resource aliases, update the relevant unit tests and documentation. Keep the existing mTLS and exec-based credential flows intact and never commit real certificates, keys, tokens, or local Athenz configuration.

## End-to-end tests

E2E tests use a real local Athenz stack from `ctyano/athenz-distribution`:

```sh
make e2e-up
make e2e
make e2e-down
```

The stack can be reused between runs. Filter scenarios with a Godog tag, for example:

```sh
GODOG_TAGS=@issue make e2e-focus
```

Add behavior scenarios under `test/e2e/features/*.feature`; add missing step definitions in `test/e2e/steps.go`. The suite creates unique `e2e-*` domains and cleans them up after scenarios. Use `make e2e-sweep` to remove leaked test domains after an interrupted run.

## Change hygiene

- Inspect the existing implementation and tests before editing.
- Preserve unrelated user changes in the working tree.
- Keep changes focused and avoid unrelated formatting or dependency updates.
- Run the relevant tests and review the final diff.
- Run `git diff --check` before handoff.
