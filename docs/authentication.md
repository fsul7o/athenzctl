# Authentication and connection

`athenzctl` stores endpoints, credentials, and connection options in a selected context. The default location is `~/.athenzctl/config.yaml`; override it with `$ATHENZCTL_CONFIG` or `--config`.

Command-line settings take precedence over the selected context.

## mTLS

mTLS is the default authentication mode. Configure a service identity certificate and private key:

```sh
athenzctl config set-context prod \
  --zms-url https://athenz-zms-server:4443/zms/v1 \
  --zts-url https://athenz-zts-server:4443/zts/v1 \
  --cert /path/to/.athenz/service.cert \
  --key /path/to/.athenz/private.key
```

This creates the following `config.yaml` fragment:

```yaml
contexts:
    - name: prod
      zms-url: https://athenz-zms-server:4443/zms/v1
      zts-url: https://athenz-zts-server:4443/zts/v1
      cert: /path/to/.athenz/service.cert
      key: /path/to/.athenz/private.key
```

Select the context with `athenzctl config use-context prod`; this adds `current-context: prod` at the top of the file.

## exec

The `exec` mode runs an external command before each ZMS or ZTS request. The command must write a fresh certificate and key to configured paths; athenzctl then reads those files. This supports external credential tools and interactive login flows without athenzctl implementing them.

```sh
athenzctl config set-context prod-usercert \
  --zms-url https://athenz-zms-server:4443/zms/v1 \
  --zts-url https://athenz-zts-server:4443/zts/v1 \
  --auth-mode exec \
  --exec-command athenzusercert \
  --exec-arg -oidc-issuer \
  --exec-arg https://login.example.com/dex \
  --exec-cert-path /path/to/.athenz/user.cert.pem \
  --exec-key-path /path/to/.athenz/user.key.pem
```

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
        cert-path: /path/to/.athenz/user.cert.pem
        key-path: /path/to/.athenz/user.key.pem
```

`--exec-arg` and `--exec-env KEY=VALUE` are repeatable and order-preserving. The external command's standard output and error are passed through, and the command is run on every invocation.

## ntoken

The `ntoken` mode signs an NToken with a service key registered in ZMS, then obtains and caches a short-lived service certificate from ZTS. It requires the service domain, service name, registered private key, and key version.

```sh
athenzctl config set-context prod-ntoken \
  --zms-url https://athenz-zms-server:4443/zms/v1 \
  --zts-url https://athenz-zts-server:4443/zts/v1 \
  --auth-mode ntoken \
  --ntoken-auth-domain my.domain \
  --ntoken-auth-service api \
  --ntoken-auth-private-key /path/to/.athenz/api.key.pem \
  --ntoken-auth-key-id 0
```

```yaml
contexts:
    - name: prod-ntoken
      zms-url: https://athenz-zms-server:4443/zms/v1
      zts-url: https://athenz-zts-server:4443/zts/v1
      auth-mode: ntoken
      ntoken-auth:
        domain: my.domain
        service: api
        private-key: /path/to/.athenz/api.key.pem
        key-id: "0"
```

Cached credentials are stored under `~/.athenzctl/cache/<context>/ntoken/` by default. Use `--auth-cache-dir` to override the cache location. CSR attributes come from `issue-defaults.servicecert`; see [certificate defaults](certificate-defaults.md).

## copperargos

The `copperargos` mode registers an instance with ZTS using an attestation-data file, then caches the returned certificate. Prepare attestation data separately, for example with `athenzctl issue instance-register-token --out` using another context.

```sh
athenzctl config set-context prod-copperargos \
  --zms-url https://athenz-zms-server:4443/zms/v1 \
  --zts-url https://athenz-zts-server:4443/zts/v1 \
  --auth-mode copperargos \
  --copperargos-auth-domain my.domain \
  --copperargos-auth-service api \
  --copperargos-auth-provider sys.auth.zts \
  --copperargos-auth-instance i-0123456789abcdef0 \
  --copperargos-auth-attestation-data /path/to/.athenz/api.attestation-data
```

```yaml
contexts:
    - name: prod-copperargos
      zms-url: https://athenz-zms-server:4443/zms/v1
      zts-url: https://athenz-zts-server:4443/zts/v1
      auth-mode: copperargos
      copperargos-auth:
        domain: my.domain
        service: api
        provider: sys.auth.zts
        instance: i-0123456789abcdef0
        attestation-data-path: /path/to/.athenz/api.attestation-data
```

Later invocations refresh the cached certificate with mTLS. If no usable cache exists, or refresh fails, registration is retried with the configured attestation data. Credentials are cached under `~/.athenzctl/cache/<context>/copperargos/` by default.

## TLS and proxies

Use `--ca-cert` for a private CA. For local or privately issued endpoints, `-k` / `--insecure-skip-tls-verify` disables certificate and hostname verification. Route ZMS and ZTS requests through a proxy with `-s` / `--proxy`:

```sh
athenzctl -k -s 127.0.0.1:1080 get domains
athenzctl --proxy https://proxy.example:8443 get domains
```

A bare `host:port` value uses SOCKS5. `socks5://`, `http://`, and `https://` URLs, including proxy credentials in URL userinfo, are supported. Persist these values with `config set-context --insecure-skip-tls-verify --proxy ...`:

```sh
athenzctl config set-context prod \
  --insecure-skip-tls-verify \
  --proxy http://proxy.example:8080
```

```yaml
contexts:
    - name: prod
      insecure-skip-tls-verify: true
      proxy: http://proxy.example:8080
```