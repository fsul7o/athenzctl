# Athenz official tools ↔ athenzctl option mapping

This document maps the options and subcommands of the official Athenz command-line tools that athenzctl replaces (`zms-cli`, `zts-accesstoken`, `zts-svccert`, `zts-rolecert`, `zpu`/`zpe-updater`) to their corresponding athenzctl commands and flags.

Reference version: `github.com/AthenZ/athenz@v1.12.43` (source read from the local `go pkg mod` cache). The athenzctl side was read directly from this repository's `internal/cli/**`.

## Table of contents

1. [Architectural notes](#architectural-notes)
2. [Auth / connection configuration mapping](#auth--connection-configuration-mapping)
3. [zms-cli global flag mapping](#zms-cli-global-flag-mapping)
4. [zms-cli ↔ athenzctl (by resource)](#zms-cli--athenzctl-by-resource)
5. [zts-accesstoken ↔ `athenzctl issue accesstoken`](#zts-accesstoken--athenzctl-issue-accesstoken)
6. [zts-svccert ↔ `athenzctl issue servicecert` / `issue instance-register-token`](#zts-svccert--athenzctl-issue-servicecert--issue-instance-register-token)
7. [zts-rolecert ↔ `athenzctl issue rolecert`](#zts-rolecert--athenzctl-issue-rolecert)
8. [zpu (zpe-updater) ↔ `athenzctl fetch signedpolicy`](#zpu-zpe-updater--athenzctl-fetch-signedpolicy)
9. [Summary of unsupported / out-of-scope features](#summary-of-unsupported--out-of-scope-features)

---

## Architectural notes

- athenzctl's `get` / `describe` / `create` / `edit` / `patch` / `delete` are each **a single Cobra command** (`internal/cli/<verb>/<verb>.go`) that takes a `KIND` argument (`role`, `policy`, `service`, etc.) as `athenzctl <verb> KIND [NAME]` and dispatches internally. Unlike zms-cli, where each resource gets its own command name (`add-role`, `show-policy`, etc.), athenzctl does not split commands per resource.
- `issue` / `fetch` / `check` / `config`, on the other hand, are command groups with real Cobra subcommands per resource/operation.
- The large number of zms-cli `set-<resource>-<field>` commands (e.g. `set-role-max-members`, `set-domain-slack-channel`) collapse in athenzctl into the generic `edit <kind> <name>` (interactive YAML edit via `$EDITOR`) or `patch <kind> <name> FIELD=VALUE` (non-interactive patch). Field names are defined per-resource in a Whitelist (see below).
- `-o/--output` is shared across all of athenzctl (`table|wide|json|yaml`). Note that zms-cli's `-o` (`json|yaml|manualYaml`) has different accepted values and a different default.

## Auth / connection configuration mapping

athenzctl centralizes connection targets and credentials in a **kubeconfig-style context file** (`~/.athenzctl/config.yaml`). What the legacy tools required as command-line flags or separate config files on every invocation is consolidated under `athenzctl config set-context`.

| Legacy tool option | athenzctl equivalent |
|---|---|
| zms-cli: `-z zms_url` | `config set-context --zms-url` |
| zts-accesstoken/zts-svccert/zts-rolecert: `-zts` | `config set-context --zts-url` |
| zms-cli: `-cert`/`-key`; each zts tool: `-svc-cert-file`/`-svc-key-file` | `config set-context --cert`/`--key` |
| zms-cli: `-c cacert_file`; each zts tool: `-cacert`/`-svc-cacert-file` | `config set-context --ca-cert` |
| zms-cli: `-k` / `-s host:port` | `-k`/`--insecure-skip-tls-verify` and `-s`/`--proxy`; persist with `config set-context --insecure-skip-tls-verify` / `--proxy`. Applies to ZMS and ZTS; bare `host:port` is SOCKS5 and `socks5://`, `http://`, and `https://` URLs are supported |
| zms-cli: `-f ntoken_file` / `-i identity` (NToken auth); each zts tool: `-ntoken-file` | **Out of scope** (athenzctl only supports mTLS or `auth-mode: exec`; NToken and legacy role-token issuance are not supported — see README) |
| (none) | `config set-context --auth-mode exec --exec-command ... --exec-arg ... --exec-env ... --exec-cert-path ... --exec-key-path ...` (obtains a fresh certificate via an external command each invocation; intended to pair with tools like `ctyano/athenz-user-cert`) |
| (none) | `config set-context --zms-server-name`/`--zts-server-name` (TLS SNI/verification-name override; used e.g. for local e2e) |
| Switching between multiple contexts (no such concept in the legacy tools) | `config use-context` / `config get-contexts` / `config current-context` / `config delete-context` / `config view` |

---

## zms-cli global flag mapping

| zms-cli flag | Meaning | athenzctl equivalent |
|---|---|---|
| `-a auditRef` | Audit reference token (required for audit-enabled domains) | `--audit-ref` on each command |
| `-b` (bulk mode) | Suppress re-fetching/displaying updated objects | No equivalent (athenzctl is designed to always return minimal results) |
| `-c cacert_file` | CA certificate file | `config set-context --ca-cert` |
| `-cert` / `-key` | Client certificate/private key (mTLS) | `config set-context --cert`/`--key` |
| `-d domain` | Domain scope | `-d`/`--domain` (global persistent flag) |
| `-e skip_errors` | Skip errors during `import-domain` | No equivalent (`import-domain` itself is unsupported) |
| `-f ntoken_file` / `-i identity` | NToken auth | **Out of scope** (NToken not supported) |
| `-k` | Disable TLS peer verification | `-k`/`--insecure-skip-tls-verify`; can also be stored in context with `config set-context --insecure-skip-tls-verify` |
| `-o output_format` (json/yaml/manualYaml) | Output format | `-o/--output` (table/wide/json/yaml — different meaning and default) |
| `-overwrite` | Overwrite without existence checks | No equivalent (`create` errors if the resource already exists; `edit`/`patch` are explicit update operations) |
| `-r resource_owner` | Stamp resource-owner metadata | No equivalent (resource-ownership metadata is unimplemented; see below) |
| `-s host:port` | SOCKS5 proxy | `-s`/`--proxy host:port`; also accepts `socks5://`, `http://`, and `https://` proxy URLs |
| `-v` | Verbose (full resource names) | `-o wide` is close but not an exact match |
| `-x` | Omit header name in `get-user-token` output | No equivalent (`get-user-token` itself is out of scope) |
| `-z zms_url` | ZMS URL | `config set-context --zms-url` |
| `-debug` | Debug mode — issues fake NTokens | No equivalent |
| `-p` / `-addself` | Require product-id / auto-add caller as admin when creating a top-level domain | No equivalent (`create domain --admin-users` specifies admins explicitly) |
| `-u user_domain` / `-h home_domain` | Customize the user/home domain prefix | No equivalent (`user.`/`home.` are fixed) |
| `-version` | Print version | `athenzctl version` |

---

## zms-cli ↔ athenzctl (by resource)

### Domain

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `add-domain domain [product-id] [admin ...]` | `create domain NAME --admin-users a,b [--description D]` | Subdomains use `--parent`; user domains use `--user` |
| `list-domain [prefix]` / `list-domain limit skip prefix depth` | `get domains` | Pagination (limit/skip/depth) is unsupported — always fetches everything |
| `show-domain [domain]` / `show-domain-attrs` | `describe domain NAME` / `get domain-meta` | |
| `delete-domain domain` | `delete domain NAME [--parent P]` | User domains use `delete domain --user U` |
| `disable-domain` / `enable-domain` | `patch domain-meta NAME enabled=false/true` | |
| `check-domain` | No equivalent | Domain validation feature |
| `use-domain` | No equivalent (`-d`/`--context` serve a similar purpose) | No interactive-mode concept exists |
| `lookup-domain-by-*` (role/tag/aws-account/azure-subscription/gcp-project/product-id/business-service) | No equivalent | Search/lookup commands are entirely unimplemented |
| `import-domain` / `update-domain` / `export-domain` (bulk YAML) | No equivalent | `edit` only supports interactive editing of a single resource — no bulk import/export |
| `system-backup dir` | No equivalent | |
| `get-signed-domains [matching_tag]` | `fetch signedpolicy DOMAIN...` (related but a different API) | zms-cli's version calls ZMS `SignedDomains` (full data for ZPU); athenzctl's calls ZTS `PostSignedPolicyRequest` (JWS for a single domain) |
| `set-default-admins domain admin [admin ...]` | No equivalent | Force-overwrite of the admin role |
| `set-domain-meta description` | `patch domain-meta NAME description="..."` | |
| `set-audit-enabled` | `patch domain-meta NAME auditEnabled=true` | |
| `set-application-id` | `patch domain-meta NAME applicationId=...` | |
| `set-business-service` | `patch domain-meta NAME businessService=...` | |
| `set-org-name` | `patch domain-meta NAME org=...` | |
| `set-product-id` | `patch domain-meta NAME productId=...` | |
| `set-domain-member-expiry-days` | `patch domain-meta NAME memberExpiryDays=N` | |
| `set-domain-member-purge-expiry-days` | `patch domain-meta NAME memberPurgeExpiryDays=N` | |
| `set-domain-service-expiry-days` | `patch domain-meta NAME serviceExpiryDays=N` | |
| `set-domain-group-expiry-days` | `patch domain-meta NAME groupExpiryDays=N` | |
| `set-domain-token-expiry-mins` | `patch domain-meta NAME tokenExpiryMins=N` | |
| `set-domain-service-cert-expiry-mins` | `patch domain-meta NAME serviceCertExpiryMins=N` | |
| `set-domain-role-cert-expiry-mins` | `patch domain-meta NAME roleCertExpiryMins=N` | |
| `set-domain-token-sign-algorithm` | `patch domain-meta NAME signAlgorithm=rsa\|ec` | |
| `set-domain-user-authority-filter` | `patch domain-meta NAME userAuthorityFilter=...` | |
| `set-domain-x509-cert-signer-keyid` | `patch domain-meta NAME x509CertSignerKeyId=...` | |
| `set-domain-ssh-cert-signer-keyid` | `patch domain-meta NAME sshCertSignerKeyId=...` | |
| `set-domain-slack-channel` | `patch domain-meta NAME slackChannel=...` | |
| `set-domain-on-call` | `patch domain-meta NAME onCall=...` | |
| `set-domain-environment` | `patch domain-meta NAME environment=...` | |
| `set-domain-feature-flags` | `patch domain-meta NAME featureFlags=N` | |
| `set-aws-account` / `set-azure-subscription` / `set-gcp-project` | `patch domain-meta NAME account=... / azureSubscription=... / gcpProject=...` | zms-cli takes multiple values (account id+name, subscription+tenant+client, etc.); athenzctl's Whitelist only exposes a single field, so the finer-grained sub-values are unsupported |
| `set-cost-center` / `set-external-member-validator` / `set-cert-dns-domain` | No equivalent | No corresponding field in `DomainMetaWhitelist` |
| `set-domain-contact type user` | `patch domain-meta NAME contacts=...` (approximate) | `contacts` is a map field; there's no dedicated per-type add command, so editing the whole map via `edit` is the practical route |
| `add-domain-tag` / `delete-domain-tag` | `edit domain-meta NAME` (edit the `tags` field in YAML) | No dedicated add/delete-one command — the whole map is edited |
| `set-domain-resource-ownership` / `reset-domain-resource-ownership` | No equivalent | resource-ownership metadata is unimplemented across the board (same for Policy/Role/Group/Service) |
| `overdue-review` / `get-stats` / `get-auth-history` / `get-dependent-domain-list` / `put-domain-dependency` / `delete-domain-dependency` / `get-dependent-service-list` | No equivalent | Review-due tracking, stats, and dependency management are unimplemented |

### Role

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `add-regular-role role [-audit-enabled] [member ...]` | `create role NAME -d DOMAIN --members m1,m2` | |
| `add-delegated-role role trusted_domain` | `create role NAME -d DOMAIN --trust TRUSTDOM` | |
| `list-role` | `get roles -d DOMAIN` | |
| `show-role role [log\|expand\|pending]` | `describe role NAME -d DOMAIN` | The `log`/`expand`/`pending` display modes are unsupported |
| `show-roles [tag_key] [tag_value]` | `get roles -d DOMAIN` (no tag filtering) | |
| `show-roles-principal` / `list-roles-for-review` | No equivalent | |
| `delete-role role` | `delete role NAME -d DOMAIN` | |
| `add-member` / `add-temporary-member` / `add-reviewed-member` | `create membership -d DOMAIN --member M --role R` | Adding with an expiration or review date is unsupported |
| `delete-member` | `delete membership -d DOMAIN --member M --role R` | |
| `check-member` / `check-active-member` | `get membership MEMBER -d DOMAIN --role R` | |
| `put-membership-decision` (approve/reject) | `create membership -d DOMAIN --member M --role R --approve`/`--reject` | |
| `add-provider-role-member` / `show-provider-role-member` / `delete-provider-role-member` | No equivalent | Tenancy (provider-role) features are entirely unimplemented |
| `list-domain-role-members` / `delete-domain-role-member` | No equivalent | |
| `set-role-audit-enabled` | `patch role NAME -d DOMAIN auditEnabled=true` | |
| `set-role-review-enabled` | `patch role NAME -d DOMAIN reviewEnabled=true` | |
| `set-role-delete-protection` | `patch role NAME -d DOMAIN deleteProtection=true` | |
| `set-role-self-renew` / `set-role-self-renew-mins` | `patch role NAME -d DOMAIN selfRenew=true / selfRenewMins=N` | |
| `set-role-self-serve` | `patch role NAME -d DOMAIN selfServe=true` | |
| `set-role-max-members` | `patch role NAME -d DOMAIN maxMembers=N` | |
| `set-role-member-expiry-days` / `-service-expiry-days` / `-group-expiry-days` | `patch role NAME -d DOMAIN memberExpiryDays=N / serviceExpiryDays=N / groupExpiryDays=N` | |
| `set-role-member-review-days` / `-service-review-days` / `-group-review-days` | `patch role NAME -d DOMAIN memberReviewDays=N / serviceReviewDays=N / groupReviewDays=N` | |
| `set-role-token-expiry-mins` / `-cert-expiry-mins` | `patch role NAME -d DOMAIN tokenExpiryMins=N / certExpiryMins=N` | |
| `set-role-token-sign-algorithm` | `patch role NAME -d DOMAIN signAlgorithm=rsa\|ec` | |
| `set-role-notify-roles` / `-notify-details` | `patch role NAME -d DOMAIN notifyRoles=... / notifyDetails=...` | |
| `set-role-user-authority-filter` / `-user-authority-expiration` | `patch role NAME -d DOMAIN userAuthorityFilter=... / userAuthorityExpiration=...` | |
| `set-role-description` | `patch role NAME -d DOMAIN description=...` | |
| `set-role-principal-domain-filter` | `patch role NAME -d DOMAIN principalDomainFilter=...` | |
| `add-role-tag` / `delete-role-tag` | `edit role NAME -d DOMAIN` (edit `tags` in YAML) | |
| `set-role-resource-ownership` | No equivalent | |

### Group

Structurally almost identical to Role (minus the certificate/token-related
fields, which groups don't have).

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `add-group group [-audit-enabled] [member ...]` | `create group NAME -d DOMAIN --members m1,m2` | |
| `list-group` / `show-group` / `show-groups` | `get group(s)` / `describe group NAME` | |
| `delete-group` | `delete group NAME -d DOMAIN` | |
| `add-group-member` / `add-temporary-group-member` / `delete-group-member` | `create membership --member M --group G` / `delete membership --member M --group G` | Adding with an expiration is unsupported |
| `check-group-member` / `check-active-group-member` | `get membership MEMBER -d DOMAIN --group G` | |
| `put-group-membership-decision` | `create membership --member M --group G --approve`/`--reject` | |
| `set-group-audit-enabled` / `-review-enabled` / `-delete-protection` / `-self-renew` / `-self-renew-mins` / `-self-serve` / `-max-members` | `patch group NAME -d DOMAIN <field>=<value>` | Field names follow the same pattern as role (`auditEnabled`, etc.) |
| `set-group-member-expiry-days` / `-service-expiry-days` | `patch group NAME -d DOMAIN memberExpiryDays=N / serviceExpiryDays=N` | |
| `set-group-notify-roles` / `-notify-details` | `patch group NAME -d DOMAIN notifyRoles=... / notifyDetails=...` | |
| `set-group-user-authority-filter` / `-user-authority-expiration` | `patch group NAME -d DOMAIN userAuthorityFilter=... / userAuthorityExpiration=...` | |
| `set-group-principal-domain-filter` | `patch group NAME -d DOMAIN principalDomainFilter=...` | |
| `add-group-tag` / `delete-group-tag` | `edit group NAME -d DOMAIN` (edit `tags` in YAML) | |
| `show-groups-principal` / `list-groups-for-review` / `list-domain-group-members` | No equivalent | |
| `set-group-resource-ownership` | No equivalent | |

### Policy / Policy version

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `add-policy policy [assertion]` | `create policy NAME -d DOMAIN` | Specifying an assertion at creation time is unsupported (add it afterward via `edit`/`patch`) |
| `add-policy-version policy version source_version` | `create policyversion POLICY:VERSION -d DOMAIN --from-version V` | |
| `list-policy` / `list-policy-versions` | `get policies` / `get policyversions POLICY -d DOMAIN` | |
| `show-policy` / `show-policy-version` / `show-policies` | `describe policy NAME` / `describe policyversion POLICY:VERSION` | |
| `delete-policy` / `delete-policy-version` | `delete policy NAME` / `delete policyversion POLICY:VERSION` | |
| `set-active-policy-version` | `patch policyversion POLICY:VERSION -d DOMAIN active=true` | Setting `active=true` internally calls `SetActivePolicyVersion` |
| `add-assertion` / `delete-assertion` / `add-assertion-policy-version` / `delete-assertion-policy-version` | `edit policy NAME` / `edit policyversion POLICY:VERSION` (edit `assertions[]` in YAML) | No dedicated add/delete-one command — the whole array is edited |
| `add-policy-tag` / `delete-policy-tag` | `edit policy NAME` (edit `tags` in YAML) | |
| `show-access action resource [...]` | `check access ACTION RESOURCE [--principal P]` | |
| `show-access-ext` | `check access ACTION RESOURCE --ext` | |
| `show-resource principal action` | `check resource --principal P [--action A]` | |
| `set-policy-resource-ownership` | No equivalent | |

### Service

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `add-service service key_id pubkey` | `create service NAME -d DOMAIN` (+ separate `create servicekey`) | zms-cli can set an initial public key at creation time; athenzctl requires a two-step `create service` → `create servicekey` |
| `add-provider-service` | No equivalent | Special creation of a provider (tenancy-capable) service is unsupported |
| `list-service` / `show-service` / `show-services` | `get services` / `describe service NAME` | |
| `search-services` | No equivalent | |
| `delete-service` | `delete service NAME -d DOMAIN` | |
| `set-service-endpoint` | `patch service NAME -d DOMAIN providerEndpoint=...` | |
| `set-service-exe service executable user group` | `patch service NAME -d DOMAIN executable=... user=... group=...` | |
| `add-service-host` / `delete-service-host` | `edit service NAME` (edit `hosts` in YAML) | |
| `add-public-key` / `show-public-key` / `delete-public-key` | `create servicekey SERVICE:KEYID --pem/--key` / `describe servicekey` / `delete servicekey` | |
| `add-service-tag` / `delete-service-tag` | `edit service NAME` (edit `tags` in YAML) | |
| `set-service-client-id` / `-x509-cert-signer-keyid` / `-ssh-cert-signer-keyid` / `-feature-flags` / `-creds` / `-resource-ownership` | No equivalent | No corresponding field in `ServiceWhitelist` |

### Quota

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `get-quota` | `get quota -d DOMAIN` / `describe quota -d DOMAIN` | |
| `set-quota key=value [...]` | `patch quota DOMAIN key1=v1 key2=v2` | Field names correspond directly (`role`, `role-member`→`roleMember`, `group`, `group-member`→`groupMember`, `subdomain`, `policy`, `assertion`, `service`, `service-host`→`serviceHost`, `public-key`→`publicKey`, `entity`) |
| `delete-quota` | `delete quota -d DOMAIN` | |

### Template

| zms-cli command | athenzctl equivalent | Notes |
|---|---|---|
| `list-server-template(s)` | `get templates` | |
| `show-server-template` | `describe template NAME` | |
| `list-domain-template(s)` | `get domain-templates -d DOMAIN` | |
| `set-domain-template name [param=value ...]` | `create domain-template NAME -d DOMAIN --param K=V` | |
| `delete-domain-template` | `delete domain-template NAME -d DOMAIN` | |

### Tenancy / Entity / System administration (unsupported areas)

The following exist in zms-cli but have no corresponding resource kind or
command in athenzctl.

| zms-cli command group | Summary |
|---|---|
| `add-tenant` / `delete-tenant` / `add-tenancy` / `delete-tenancy` | Tenant registration / establishing a tenancy relationship |
| `show-tenant-resource-group-roles` / `add-tenant-resource-group-roles` / `delete-tenant-resource-group-roles` | Tenant-side resource-group roles |
| `show-provider-resource-group-roles` / `add-provider-resource-group-roles` / `delete-provider-resource-group-roles` | Provider-side resource-group roles |
| `list-entity` / `show-entity` / `add-entity` / `delete-entity` | Generic entity (key=value store) |
| `set-default-admins` / `list-user` / `delete-user` / `disable-principal` / `enable-principal` | System administrator / principal management |
| `get-user-token` | NToken issuance (**out of scope**, per README) |

---

## zts-accesstoken ↔ `athenzctl issue accesstoken`

| zts-accesstoken flag | athenzctl equivalent | Notes |
|---|---|---|
| `-domain` | `-d`/`--domain` (global) | |
| `-service` | No equivalent | athenzctl identifies the calling service from the context's certificate; there is no explicit flag |
| `-roles` | `-r`/`--role` (repeatable) | zts-accesstoken takes one comma-separated string; athenzctl repeats `-r` |
| `-ntoken-file` / `-hdr` | **Out of scope** | NToken auth is unsupported — mTLS only |
| `-svc-key-file` / `-svc-cert-file` / `-svc-cacert-file` | `config set-context --key/--cert/--ca-cert` | Credentials are centralized in the context |
| `-zts` | `config set-context --zts-url` | |
| `-expire-time` (minutes) | `--expires-in` (**seconds**) | ⚠️ Note the unit difference |
| `-proxy` | No equivalent | No proxy-enable flag |
| `-validate` / `-claims` / `-access-token` / `-conf` | No equivalent | Access-token validation mode is not implemented in athenzctl (issuance only) |
| `-authorization-details` | `--authorization-details` | |
| `-proxy-principal-spiffe-uris` | No equivalent | |
| `-token-only` | Default behavior (prints only the token string when `-o` is not given) | Use `-o json/yaml` to get the full response |
| `-actor` | `--proxy-for-principal` (conceptually close) | Not an exact 1:1 mapping |
| `-version` | `athenzctl version` | |

## zts-svccert ↔ `athenzctl issue servicecert` / `issue instance-register-token`

| zts-svccert flag | athenzctl equivalent | Notes |
|---|---|---|
| `-csr` | `--csr` | |
| `-get-instance-register-token` | `issue instance-register-token` (a separate subcommand) | |
| `-use-instance-register-token` | `--use-instance-register-token` | |
| `-spiffe` / `-spiffe-trust-domain` | `--spiffe` / `--spiffe-trust-domain` | |
| `-expiry-time` | `--expiry-time` | |
| `-cert-file` | `--out` | |
| `-signer-cert-file` | `--signer-cert-out` | |
| `-cacert` | `config set-context --ca-cert` | |
| `-private-key` | `--private-key` | |
| `-domain` | `-d`/`--domain` (global) | |
| `-service` | `--service` | |
| `-dns-domain` | `--dns-domain` | |
| `-subj-c` / `-subj-o` / `-subj-ou` | `--subj-c` / `--subj-p` / `--subj-o` / `--subj-ou` | `--subj-p` is athenzctl extension |
| (none) | Build-time `ISSUE_DEFAULT_SERVICECERT_*` or context `issue-defaults.servicecert` | CLI flags override context and build-time defaults |
| `-ip` | `--ip` | |
| `-provider` / `-instance` | `--provider` / `--instance` | Used by both `issue servicecert` (`--instance-id` is a deprecated alias) and `issue instance-register-token` |
| `-attestation-data` | `--attestation-data` | |
| `-svc-key-file` / `-svc-cert-file` / `-ntoken-file` / `-key-version` / `-hdr` | `config set-context --key/--cert` (NToken-related flags are **out of scope**) | Auth for fetching the instance register token is also centralized in the context |
| `-signer-key-id` | `--signer-key-id` | |
| (none) | `--concat-intermediate-cert` | athenzctl appends the CA bundle returned with the service certificate to the output certificate |
| `-service-cert` | No equivalent | |
| `-version` | `athenzctl version` | |

## zts-rolecert ↔ `athenzctl issue rolecert`

| zts-rolecert flag | athenzctl equivalent | Notes |
|---|---|---|
| `-role-key-file` | `--role-key-file` | |
| `-role-cert-file` | `--out` | |
| `-cacert` | `config set-context --ca-cert` | |
| `-svc-key-file` / `-svc-cert-file` | `config set-context --key/--cert` | |
| `-domain` / `-service` | `-d`/`--domain` (global) / `--service` | ⚠️ Upstream `zts-rolecert` has a known quirk: `-domain`/`-service` are declared but unused in practice (the values are extracted from the cert CN instead). athenzctl's `--service` actually works |
| `-zts` | `config set-context --zts-url` | |
| `-role-domain` / `-role-name` | `--role-domain` / `--role-name` (`--role` is a deprecated alias) | |
| `-dns-domain` | `--dns-domain` | |
| `-subj-c` / `-subj-o` / `-subj-ou` | `--subj-c` / `--subj-p` / `--subj-o` / `--subj-ou` | `--subj-p` is athenzctl extension |
| (none) | Build-time `ISSUE_DEFAULT_ROLECERT_*` or context `issue-defaults.rolecert` | CLI flags override context and build-time defaults |
| `-ip` | `--ip` | |
| `-old-role-cert` | `--old-role-cert` | |
| `-spiffe` / `-spiffe-trust-domain` | `--spiffe` / `--spiffe-trust-domain` | |
| `-csr` | `--csr` | |
| `-expiry-time` | `--expiry-time` (`--expiry-mins` is a deprecated alias) | |
| (none) | `--concat-intermediate-cert` | When the response does not already contain a chain, fetch and append the CA bundle named by `--cacert-bundle-name` |
| (none) | `--cacert-bundle-name` | CA bundle name used with `--concat-intermediate-cert` |
| `-proxy` | `--proxy-for-principal` (different concept — see notes) | zts-rolecert's `-proxy` is an HTTP-proxy-enable flag; athenzctl's `--proxy-for-principal` specifies a delegate principal. Neither is a direct equivalent of the other |
| `-version` | `athenzctl version` | |

## zpu (zpe-updater) ↔ `athenzctl fetch signedpolicy`

**Note the architectural difference**: `zpu` operates over a **domain list**
configured in `zpu.conf`, and is designed to run periodically (cron/daemon)
to keep locally cached, signature-verified policy files up to date. By
contrast, `athenzctl fetch signedpolicy` is a simple on-demand command that
fetches and prints/writes the domains given as command-line arguments **on
the spot** — it has no caching, state-tracking, or signature-verification
functionality.

| zpu flag / zpu.conf field | athenzctl equivalent | Notes |
|---|---|---|
| `zpu.conf`: `domains` (comma-separated list) | `fetch signedpolicy DOMAIN [DOMAIN...]` (positional args) | Specified per-invocation as command-line arguments rather than in a config file |
| `-zts` | `config set-context --zts-url` | |
| `-cacert` / `-private-key` / `-cert-file` (equivalently `zpu.conf`'s `caCertFile`/`privateKeyFile`/`certFile`) | `config set-context --ca-cert/--key/--cert` | |
| `zpu.conf`: `policyDir` | `--output-dir` (writes to `<output-dir>/<domain>.pol`) | zpu writes atomically via a temp directory; athenzctl writes directly |
| `zpu.conf`: `tempPolicyDir` | No equivalent | No atomic-write mechanism |
| `zpu.conf`: `policyVersions` (per-domain version map) | `--policy-version` (a single value applied to all domains in the fetch call) | Per-domain version pinning is not supported |
| `-force-refresh` | No equivalent | athenzctl always fetches fresh (there is no cache to refresh) |
| `-check-status` / `-check-details` | No equivalent | No status-check/metrics-output functionality |
| `-view-domain` | No equivalent | No local-cache viewing mode — athenzctl always calls ZTS |
| `-sia-dir` | No equivalent | |
| `zpu.conf`: `expiryCheck` / `checkZMSSignature` / `athenz.conf`'s `ztsPublicKeys`/`zmsPublicKeys` | No equivalent | athenzctl does not perform signature verification itself — it just outputs the raw JWS policy data |
| `zpu.conf`: `jwsPolicySupport` | Always on (JWS is the only supported format) | athenzctl does not handle the legacy non-JWS format |
| (none) | `--p1363` | athenzctl-only addition — lets you choose the signature format (ASN.1 DER vs. P1363) |
| `-debug` / `-logFile` / zpu.conf's logging-related fields in general | No equivalent | athenzctl is a one-shot CLI rather than a resident daemon, so there's no concept of log rotation, etc. |

---

## Summary of unsupported / out-of-scope features

athenzctl is in "early development (pre-v0.1)"; the following are currently clearly unimplemented, or deliberately out of scope (see `README.md`).

- **NToken auth / legacy role-token issuance**: entirely out of scope. Only mTLS or `auth-mode: exec` are supported.
- **Tenancy features in general** (`add-tenant*`, `*-resource-group-roles`, `add-provider-service`): unimplemented.
- **Generic entities** (`*-entity`): no corresponding resource kind exists.
- **resource-ownership metadata** (`set-*-resource-ownership`): unimplemented for every resource.
- **Bulk domain import/export** (`import-domain`/`export-domain`/`update-domain`/`system-backup`): unimplemented. `edit`/`patch` only operate on a single resource at a time.
- **Search/lookup commands** (`lookup-domain-by-*`, `search-services`, `show-roles-principal`, etc.): unimplemented.
- **Review-due tracking, stats, dependency management** (`overdue-review`, `get-stats`, `get-auth-history`, `get-dependent-*`, `put-*-dependency`): unimplemented.
- **System administrator / principal management** (`set-default-admins`, `list-user`, `delete-user`, `disable-principal`, `enable-principal`): unimplemented.
- **zpu's state-management features** (`-check-status`, `-check-details`, `-view-domain`, caching, signature verification): `fetch signedpolicy` is on-demand fetch only and has none of zpu's daemon-like functionality.
- **zts-accesstoken's token validation mode** (`-validate`): not implemented (issuance only).

Conversely, features athenzctl adds that have no equivalent in the legacy tool set:

- **Unified context management** (the `config` command group): switch between multiple environments kubeconfig-style.
- **`auth-mode: exec`**: obtain a fresh certificate on the spot via an external command (intended to pair with OIDC-login-style wrapper tools).
- **`issue instance-register-token`**: instance-register-token retrieval as its own command.
- **`--p1363`** (`fetch signedpolicy`): explicit choice of signature format.
- **`patch` command**: alongside `edit` (interactive YAML editing), provides a scriptable, non-interactive `FIELD=VALUE` patch — a generalization of zms-cli's large set of `set-*` commands.
