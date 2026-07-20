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

## Development and AI context

See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow, unit and end-to-end tests, and Microsoft APM setup. The repository's AI instructions and review agent are maintained under `.apm/`; choose a target such as `apm install --target codex` followed by `apm compile --target codex --clean` to generate only the client-specific context you need.

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

- **Verbs:** `get`, `describe`, `create`, `delete`, `edit`, `lookup`, `issue`, `fetch`, `config`.
- **Auth:** mTLS (service identity certificate + private key, the default) or `auth-mode: exec` to obtain the client cert from an external command (kubectl exec-credential style — see below). `auth-mode: ntoken`/`auth-mode: copperargos` mint a fresh short-lived service certificate via ZTS on every invocation (with on-disk caching) instead of relying on a pre-issued static certificate — see below. NToken itself is not otherwise exposed as a general-purpose auth mode, and legacy role-token issuance is out of scope.
- **Config file:** `~/.athenzctl/config.yaml` (kubeconfig-style; override with `$ATHENZCTL_CONFIG` or `--config`).

## Issue certificate defaults

The distribution build can embed the defaults used by `issue rolecert` and by the `ntoken`/`copperargos` auth-modes' CSR building, so users do not need to configure organization-specific CSR values after installing the binary. `issue rolecert`'s public builds keep the generic defaults (`US`, `Oath Inc.`, `Athenz`, and SPIFFE enabled). The `ntoken`/`copperargos` auth-modes have no such hardcoded fallback for `subj-c`/`subj-p`/`subj-o`/`subj-ou`: each defaults to an empty string — and is simply omitted from the CSR subject — unless a build-time default or context override configures it (SPIFFE still defaults to enabled). A private distribution can override them at build time, for example:

```sh
make build \
    ISSUE_DEFAULT_SERVICECERT_SUBJ_C=JP \
    ISSUE_DEFAULT_SERVICECERT_SUBJ_P=Tokyo \
    ISSUE_DEFAULT_SERVICECERT_SUBJ_O='Example Inc.' \
    ISSUE_DEFAULT_SERVICECERT_SUBJ_OU=Services \
    ISSUE_DEFAULT_SERVICECERT_DNS_DOMAIN=athenz.example \
    ISSUE_DEFAULT_SERVICECERT_CONCAT_INTERMEDIATE_CERT=true \
    NTOKEN_AUTH_DEFAULT_HDR=Athenz-Principal-Auth \
    ISSUE_DEFAULT_ROLECERT_SUBJ_C=JP \
    ISSUE_DEFAULT_ROLECERT_SUBJ_P=Tokyo \
    ISSUE_DEFAULT_ROLECERT_SUBJ_O='Example Inc.' \
    ISSUE_DEFAULT_ROLECERT_SUBJ_OU=Roles \
    ISSUE_DEFAULT_ROLECERT_DNS_DOMAIN=athenz.example \
    ISSUE_DEFAULT_ROLECERT_CONCAT_INTERMEDIATE_CERT=true \
    ISSUE_DEFAULT_ROLECERT_CACERT_BUNDLE_NAME=athenz \
    ISSUE_DEFAULT_ROLECERT_EXPIRY_TIME=43200 \
    ISSUE_DEFAULT_ROLECERT_IP=10.0.0.1 \
    ISSUE_DEFAULT_ROLECERT_SIGNER_KEY_ID=0
```

The same values can be supplied to GoReleaser through the corresponding `ATHENZCTL_ISSUE_DEFAULT_SERVICECERT_*`/`ATHENZCTL_ISSUE_DEFAULT_ROLECERT_*`/`ATHENZCTL_NTOKEN_AUTH_DEFAULT_HDR` environment variables. Service and role certificate defaults are independent. A context may override the embedded values under `issue-defaults.servicecert`/`issue-defaults.rolecert`, and (for `issue rolecert` specifically) an explicit command-line flag always has the highest priority — `issue-defaults.servicecert` has no such CLI-flag layer above it, since it's consumed by the `ntoken`/`copperargos` auth-modes (see Auth modes below) rather than by a dedicated `issue servicecert` command (athenzctl does not implement one; use the official `zts-svccert` tool for one-off/manual service cert issuance). Of `issue-defaults.servicecert`'s fields, `dns-domain`/`subj-*`/`spiffe`/`spiffe-trust-domain`/`concat-intermediate-cert` are all used by `ntoken`/`copperargos`; its `cacert-bundle-name`/`expiry-time`/`ip`/`signer-key-id` fields are `issue rolecert`-only and are ignored by `ntoken`/`copperargos`.

For example, the optional context-level override is:

```yaml
contexts:
  - name: prod
    issue-defaults:
      servicecert:
        subj-c: JP
        subj-p: Tokyo
        subj-o: Example Inc.
        subj-ou: Services
        spiffe: false
        dns-domain: athenz.example
        concat-intermediate-cert: true
      rolecert:
        subj-c: JP
        subj-p: Tokyo
        subj-o: Example Inc.
        subj-ou: Roles
        spiffe-trust-domain: spiffe.example
        dns-domain: athenz.example
        concat-intermediate-cert: true
        cacert-bundle-name: athenz
        expiry-time: 43200
        ip: 10.0.0.1
        signer-key-id: "0"
```

`dns-domain` can be omitted from `issue rolecert` when it is embedded in the binary or configured in the selected context; otherwise it remains required (`ntoken`/`copperargos` similarly require `issue-defaults.servicecert.dns-domain` to be configured via the build-time default or the context, since they have no command-line flag of their own to supply it). For an exceptional context-specific override, use the certificate-prefixed `config set-context` flags such as `--servicecert-subj-o` or `--rolecert-dns-domain`. The corresponding Province defaults are available as `--servicecert-subj-p` and `--rolecert-subj-p`. `--servicecert-concat-intermediate-cert` is available for `ntoken`/`copperargos`. The remaining `rolecert`-only detail flags are configurable the same way: `--rolecert-concat-intermediate-cert`, `--rolecert-cacert-bundle-name`, `--rolecert-expiry-time`, `--rolecert-ip`, and `--rolecert-signer-key-id`.

Note that `--servicecert-subj-c`/`-subj-p`/`-subj-o`/`-subj-ou` (and their `ISSUE_DEFAULT_SERVICECERT_SUBJ_*` build-time equivalents) have no non-empty fallback for `ntoken`/`copperargos`: leaving one unset at every layer results in that RDN being left out of the minted certificate's subject entirely, rather than substituting a generic placeholder value.

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

**ntoken** — mints a fresh service X.509 certificate on every invocation using a private key already registered in ZMS (a public/private key pair added via `create servicekey`/`describe servicekey`), mirroring upstream `zts-svccert`'s "Registered Public/Private Key Pair" mode. A ZMS service token (NToken) is signed with that key and used to authenticate a ZTS refresh call — no pre-issued certificate is needed at all, only the long-lived key pair and its ZMS key version:

```sh
athenzctl config set-context prod-ntoken \
    --zms-url https://athenz-zms-server:4443/zms/v1 \
    --zts-url https://athenz-zts-server:4443/zts/v1 \
    --auth-mode ntoken \
    --ntoken-auth-domain my.domain \
    --ntoken-auth-service api \
    --ntoken-auth-private-key "$HOME/.athenz/api.key.pem" \
    --ntoken-auth-key-id 0

athenzctl config use-context prod-ntoken
athenzctl get domains   # signs an NToken, refreshes a cert via ZTS, and caches it
```

The resulting certificate/key pair is cached (by default under `~/.athenzctl/cache/<context>/ntoken/`, overridable with `--auth-cache-dir`) and reused across invocations until it nears expiry, so ZTS is not contacted on every single command. CSR details (`dns-domain`, `subj-*`, `spiffe`, `concat-intermediate-cert`) come from that context's `issue-defaults.servicecert` (see the CSR defaults section above) — configure those alongside `--ntoken-auth-*` if the built-in defaults don't fit. The HTTP header the signed NToken is sent under defaults to `Athenz-Principal-Auth` (mirroring `zts-svccert`'s `-hdr` flag); override it per-context with `--ntoken-auth-hdr`, or embed a different default at build time with `NTOKEN_AUTH_DEFAULT_HDR` (see Issue certificate defaults above).

**copperargos** — obtains a service X.509 certificate via ZTS's Copper Argos provider registration flow, authenticated by a previously prepared attestation-data file rather than a registered key pair. athenzctl does **not** generate or refresh that attestation-data itself — prepare it out of band first, e.g. with `athenzctl issue instance-register-token --out` using any other context/auth-mode, or with a cloud-native attestation mechanism:

```sh
athenzctl config set-context prod-copperargos \
    --zms-url https://athenz-zms-server:4443/zms/v1 \
    --zts-url https://athenz-zts-server:4443/zts/v1 \
    --auth-mode copperargos \
    --copperargos-auth-domain my.domain \
    --copperargos-auth-service api \
    --copperargos-auth-provider sys.auth.zts \
    --copperargos-auth-instance i-0123456789abcdef0 \
    --copperargos-auth-attestation-data "$HOME/.athenz/api.attestation-data"

athenzctl config use-context prod-copperargos
athenzctl get domains   # registers (once) or self-refreshes (thereafter) via ZTS, and caches the result
```

Once a certificate has been issued, later invocations refresh it using the cached certificate itself as the mTLS credential (`PostInstanceRefreshRequest`), rather than re-registering — most providers only allow registering a given provider/instance pair once. Registration from the attestation-data file is only retried if no usable cache exists yet, or a refresh attempt fails (in which case you may need to point `--copperargos-auth-instance` at a new, not-yet-registered instance ID together with fresh attestation-data). The CSR private key is always a fresh ephemeral key generated in memory for each registration attempt (not user-configurable, since there's never an existing key worth preserving at that point). As with `ntoken`, results are cached under `~/.athenzctl/cache/<context>/copperargos/` by default (`--auth-cache-dir` to override — shared with `ntoken` since a context only ever uses one auth-mode at a time), and CSR details come from `issue-defaults.servicecert`.

## TLS verification and proxies

For local or otherwise privately-issued endpoints, `-k` / `--insecure-skip-tls-verify` disables TLS certificate and hostname verification. Requests can be routed through a proxy with `-s` / `--proxy`:

```sh
athenzctl -k -s 127.0.0.1:1080 get domains
athenzctl --proxy https://proxy.example:8443 get domains
```

The bare `host:port` form is treated as SOCKS5. `socks5://`, `http://`, and `https://` URLs are also accepted, including URL userinfo for proxy authentication. These options apply to both ZMS and ZTS. To save them in a context, use `config set-context --insecure-skip-tls-verify --proxy ...`; command-line values take precedence over context values.

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
