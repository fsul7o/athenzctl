---
description: Repository-wide development rules for athenzctl
---

- Treat `.apm/` as the source of truth for AI instructions and agents. Do not hand-edit generated `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, or generated client rule files; update the source and run the documented APM generation flow.
- Read `README.md`, `CONTRIBUTING.md`, `Makefile`, and the relevant package tests before making a change. Preserve unrelated working-tree changes and keep edits focused.
- This repository is a Go CLI. The executable entry point is `cmd/athenzctl`; Cobra command wiring is in `internal/cli`; Athenz transport clients are in `internal/client`; configuration is in `internal/config`; output rendering is in `internal/printer`; canonical resource names are in `internal/resource`.
- Keep the verb-resource CLI model consistent. Shared behavior belongs in the appropriate shared package rather than being duplicated across resource handlers.
- When changing command flags, resource aliases, config YAML, output formats, authentication, or TLS/proxy behavior, update the relevant unit tests and user-facing documentation. Preserve compatibility with existing command forms unless the change explicitly requires a breaking change.
- Use the existing Athenz client and configuration abstractions. Do not bypass mTLS, exec-based credential handling, context selection, TLS verification, or proxy configuration with ad-hoc HTTP code.
- Treat certificates, private keys, access tokens, service credentials, and Athenz configuration files as sensitive. Never commit real credentials, print secret material in tests, or weaken TLS verification by default.
- Prefer small, table-driven unit tests for parsing, configuration, flag validation, client setup, and rendering behavior. Use the Gherkin suite under `test/e2e/features` for behavior that requires a real Athenz stack.
- Run `make test`, `make lint`, and the relevant E2E scenarios before handing off a code change. E2E tests require Docker, KinD, and the local Athenz distribution; do not replace them with fabricated success when the environment is unavailable.
- For APM changes, use `apm install` followed by `apm compile --all --clean`. Run `apm compile --validate --all` and `apm audit --ci` before committing generated context. Do not add external APM packages or MCP servers without an explicit requirement.
- Review the final diff for accidental generated files, credentials, stale documentation, and unrelated changes. Use `git diff --check` as a final hygiene check.
