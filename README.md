# athenzctl

> **Disclaimer:** This is **not** an official Athenz project.

Unified, kubectl-style command-line client for [Athenz](https://www.athenz.io/).

`athenzctl` consolidates the functionality of the individual Athenz command-line tools (`zms-cli`, `zts-accesstoken`, `zts-svccert`, `zts-rolecert`, `zpe-updater`) behind a single verb-resource interface and a kubeconfig-style context file.

> Status: **early development (pre-v0.1)**. The command surface is being built incrementally; see the plan in this repository for scope.

## Install

```
go install github.com/fsul7o/athenzctl/cmd/athenzctl@latest
```

Homebrew and pre-built binaries via GitHub Releases will follow with the first tagged release.

## Quick start

```sh
# Register a context (mTLS: service identity cert + key)
athenzctl config set-context prod \
    --zms-url https://athenz-zms-server:4443/zms/v1 \
    --zts-url https://athenz-zts-server:4443/zts/v1 \
    --cert /path/to/.athenz/service.cert \
    --key  /path/to/.athenz/private.key

athenzctl config use-context prod
athenzctl config get-contexts

# Later
athenzctl get domains
athenzctl get roles -d my.domain
athenzctl issue accesstoken -d my.domain -r admin
athenzctl fetch signedpolicy my.domain --output-dir /var/lib/zpe/
```

## Design

- **Verbs:** `get`, `describe`, `create`, `delete`, `edit`, `issue`, `fetch`, `config`.
- **Auth:** mTLS (service identity certificate + private key, the default) or `auth-mode: exec` to obtain the client cert from an external command (kubectl exec-credential style — see below). NToken and legacy role-token issuance are out of scope.
- **Config file:** `~/.athenzctl/config.yaml` (kubeconfig-style; override with `$ATHENZCTL_CONFIG` or `--config`).

## Auth modes

**mTLS (default)** — as shown in Quick start: `--cert`/`--key` point at a static service identity certificate and key.

**exec** — athenzctl never implements credential-issuance logic (OIDC login flows, browser prompts, etc.) itself. Instead, `auth-mode: exec` names an external command that's expected to place a fresh client certificate and key at two known file paths; athenzctl runs that command before every ZMS/ZTS call and reads the cert/key back from those paths. This mirrors the common Athenz-ecosystem pattern of standalone credential tools (e.g. [ctyano/athenz-user-cert](https://github.com/ctyano/athenz-user-cert)) that write `~/.athenz/user.{cert,key}.pem` and exit — no JSON-on-stdout contract to implement, just point `--exec-cert-path`/`--exec-key-path` at wherever your tool already writes:

```sh
athenzctl config set-context prod-usercert \
    --zms-url https://athenz-zms-server:4443/zms/v1 \
    --zts-url https://athenz-zts-server:4443/zts/v1 \
    --auth-mode exec \
    --exec-command athenzusercert \
    --exec-arg -oidc-issuer --exec-arg https://login.example.com/dex \
    --exec-arg -endpoint --exec-arg https://athenz-zts-server:4443/zts/v1/usercert \
    --exec-cert-path "$HOME/.athenz/user.cert.pem" \
    --exec-key-path "$HOME/.athenz/user.key.pem"

athenzctl config use-context prod-usercert
athenzctl get domains   # execs athenzusercert first, then reads the cert/key it wrote
```

Which writes this to `~/.athenzctl/config.yaml`:

```yaml
contexts:
  - name: prod-usercert
    zms-url: https://athenz-zms-server:4443/zms/v1
    zts-url: https://athenz-zts-server:4443/zts/v1
    auth-mode: exec
    exec:
      command: athenzusercert
      args:
        - -oidc-issuer
        - https://login.example.com/dex
        - -endpoint
        - https://athenz-zts-server:4443/zts/v1/usercert
      cert-path: "$HOME/.athenz/user.cert.pem"
      key-path: "$HOME/.athenz/user.key.pem"
```

- `--exec-arg` is repeatable and order-preserving (one flag per argument — don't try to pack multiple args into one `--exec-arg` value).
- `--exec-env KEY=VALUE` (also repeatable) sets extra environment variables for the command — handy for secrets you don't want on the process argv (e.g. `--exec-env OIDC_CLIENT_SECRET=...`).
- The command's stdout/stderr are passed straight through to your terminal, so interactive login prompts still work.
- athenzctl re-runs the command on every invocation; it does not cache or refresh anything itself — that's entirely the external tool's job.

## Local e2e (Gherkin / godog)

End-to-end tests run each subcommand in-process against a real Athenz stack spun up locally by [ctyano/athenz-distribution](https://github.com/ctyano/athenz-distribution) via Docker Compose. Scenarios are written in Gherkin (`test/e2e/features/*.feature`) and executed by [godog](https://github.com/cucumber/godog).

Prerequisites: Docker + Docker Compose. No `/etc/hosts` edit needed — the generated context uses `zms-server-name` / `zts-server-name` to override TLS SNI/verification so `https://localhost:{4443,8443}` validates against the stack's server certs (SAN `athenz-{zms,zts}-server`).

Run:

```sh
make e2e-up           # clones athenz-distribution to .local/, brings the stack up,
                      # writes .local/e2e/config.yaml with admin mTLS material
make e2e              # runs all .feature scenarios (skips @skip by default)
make e2e-down         # tears the stack down (optional; the stack can be reused)
```

The same running stack can be reused across many `make e2e` invocations:

- Each scenario allocates a unique `e2e-<slug>-<ns>` domain, so parallel or repeated runs never collide.
- The `After` hook cascade-deletes that domain (roles / services / policies / groups first, then the top-level domain) so state does not accumulate.
- The `BeforeSuite` hook sweeps any `e2e-*` domains leaked by a prior interrupted run, so a fresh `make e2e` always starts from a clean slate without needing `make e2e-down && make e2e-up`.
- Manual sweep at any time: `make e2e-sweep` (fires the BeforeSuite hook and exits without running scenarios).

Filter by tag:

```sh
GODOG_TAGS=@issue make e2e-focus
```

Add a scenario by dropping Gherkin into `test/e2e/features/*.feature`. If a step is missing, add its regexp + Go handler to `test/e2e/steps.go`.

## License

Apache License 2.0.
