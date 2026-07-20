# Certificate defaults

Distribution builds can embed defaults for `issue rolecert` and for service CSRs created by the `ntoken` and `copperargos` authentication modes. This avoids configuring organization-specific CSR values for every context.

Public builds use generic `issue rolecert` defaults. Service CSR fields (`subj-c`, `subj-p`, `subj-o`, and `subj-ou`) are omitted when no build-time or context value is supplied. `dns-domain` must be configured for `ntoken` and `copperargos` through a build default or the selected context.

## Build-time configuration

Pass defaults when building a private distribution:

```sh
make build \
  ISSUE_DEFAULT_SERVICECERT_SUBJ_O='Example Inc.' \
  ISSUE_DEFAULT_SERVICECERT_DNS_DOMAIN=athenz.example \
  ISSUE_DEFAULT_ROLECERT_SUBJ_O='Example Inc.' \
  ISSUE_DEFAULT_ROLECERT_DNS_DOMAIN=athenz.example
```

The supported prefixes are `ISSUE_DEFAULT_SERVICECERT_*`, `ISSUE_DEFAULT_ROLECERT_*`, and `NTOKEN_AUTH_DEFAULT_HDR`. GoReleaser accepts the corresponding `ATHENZCTL_`-prefixed environment variables.

## Context overrides

Override embedded defaults for a context under `issue-defaults`:

```yaml
contexts:
  - name: prod
    issue-defaults:
      servicecert:
        subj-o: Example Inc.
        dns-domain: athenz.example
        concat-intermediate-cert: true
      rolecert:
        subj-o: Example Inc.
        dns-domain: athenz.example
        expiry-time: 43200
```

For exceptional overrides, use `config set-context` flags such as `--servicecert-subj-o`, `--servicecert-dns-domain`, and `--rolecert-dns-domain`. An explicit `issue rolecert` flag takes precedence over context and build defaults.

`issue-defaults.servicecert` is used only by the `ntoken` and `copperargos` modes. `cacert-bundle-name`, `expiry-time`, `ip`, and `signer-key-id` apply only to role certificates.