# athenzctl

> **Disclaimer:** This is **not** an official Athenz project.

`athenzctl` is a kubectl-style command-line client for [Athenz](https://www.athenz.io/). It provides a unified verb-resource interface for common ZMS and ZTS operations, with kubeconfig-style contexts for connection and credentials.

> Status: **early development (pre-v0.1)**. Commands are added incrementally.

## Install

```sh
go install github.com/fsul7o/athenzctl/cmd/athenzctl@latest
```

## Quick start

Create and select a context with your ZMS/ZTS endpoints and mTLS credentials:

```sh
athenzctl config set-context prod \
  --zms-url https://athenz-zms-server:4443/zms/v1 \
  --zts-url https://athenz-zts-server:4443/zts/v1 \
  --cert /path/to/.athenz/service.cert \
  --key /path/to/.athenz/private.key

athenzctl config use-context prod
```

Then run operations using a verb and resource:

```sh
athenzctl get domains
athenzctl get roles -d my.domain
athenzctl issue accesstoken -d my.domain -r admin
athenzctl fetch signedpolicy my.domain --output-dir /var/lib/zpe/
```

The default configuration path is `~/.athenzctl/config.yaml`. Override it with `$ATHENZCTL_CONFIG` or `--config`.

## Commands

| Group | Purpose |
| --- | --- |
| `get`, `describe` | List or inspect Athenz resources |
| `create`, `delete`, `edit`, `patch` | Manage resources |
| `check`, `lookup` | Check access or find domains |
| `issue`, `fetch` | Obtain tokens, certificates, and signed policies |
| `config` | Manage contexts and connection settings |

Run `athenzctl --help` or `athenzctl <command> --help` for available resources and flags.

## Documentation

- [Authentication and connection](docs/authentication.md): mTLS, exec, ntoken, copperargos, TLS, and proxies.
- [Certificate defaults](docs/certificate-defaults.md): build-time defaults and context overrides for certificate CSRs.
- [Local E2E](docs/local-e2e.md): run and filter the local Athenz test stack.
- [Command mapping](docs/command-mapping.md): migration reference for official Athenz CLI tools.
- [Contributing guide](CONTRIBUTING.md): development setup, test commands, and AI context setup.

## License

Apache License 2.0.
