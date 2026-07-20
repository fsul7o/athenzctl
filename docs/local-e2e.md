# Local E2E

End-to-end tests run athenzctl subcommands in-process against a local Athenz stack from [ctyano/athenz-distribution](https://github.com/ctyano/athenz-distribution). Scenarios live in `test/e2e/features/*.feature` and run with [godog](https://github.com/cucumber/godog).

Docker and Docker Compose are required. The generated context uses TLS server-name overrides, so no `/etc/hosts` edit is needed.

```sh
make e2e-up    # clone and start the local stack; write .local/e2e/config.yaml
make e2e       # run all scenarios except @skip
make e2e-down  # stop the local stack
```

Reuse a running stack across `make e2e` runs. Each scenario uses a unique `e2e-*` domain and cleanup removes its resources. Run `make e2e-sweep` to remove domains left behind by an interrupted run.

Run tagged scenarios with:

```sh
GODOG_TAGS=@issue make e2e-focus
```

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development prerequisites and the complete testing workflow.