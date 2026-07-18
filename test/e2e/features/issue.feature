@issue
Feature: athenzctl issue (ZTS credentials)

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-issue"
    And a domain "$DOMAIN" exists

  Scenario: issue accesstoken for own domain
    When I run athenzctl "issue accesstoken -d sys.auth -r admin -o json"
    Then the command should succeed
    And stdout should be valid json
    And stdout should contain "access_token"

  Scenario: issue an access token with request options
    When I run athenzctl "issue accesstoken -d sys.auth --role admin --expires-in 60 -o json"
    Then the command should succeed
    And stdout should be valid json
    And stdout should contain "access_token"

  Scenario: issue an access token with proxy and authorization details
    When I run athenzctl "issue accesstoken -d sys.auth --role admin --proxy-for-principal user.athenz_admin --authorization-details {} -o json"
    Then the command should fail with "not authorized for proxy access token request"

  Scenario: issue rolecert
    # user.athenz_admin is a member of sys.auth:role.admin via bootstrap, so
    # rolecert works without any per-scenario setup. Using a fresh test-domain
    # role instead would race the ZTS domain-update interval.
    When I run athenzctl "issue rolecert --role-domain sys.auth --role-name admin --dns-domain athenz.cloud"
    Then the command should fail with "Unable to validate cert request"

  Scenario: generate a role certificate CSR with every CSR option
    When I run athenzctl "issue rolecert --domain sys.auth --role-domain sys.auth --role-name admin --role admin --service e2e-svc --role-key-file $ADMIN_KEY --dns-domain athenz.cloud --subj-c JP --subj-o Example --subj-ou E2E --spiffe=false --spiffe-trust-domain example.test --ip 127.0.0.1 --old-role-cert $ADMIN_CERT --csr --expiry-time 5 --expiry-mins 5 --out $TEMP_DIR/role.cert.pem --proxy-for-principal user.athenz_admin"
    Then the command should succeed
    And stdout should contain "CERTIFICATE REQUEST"

  Scenario: issue servicecert (CSR only)
    When I run athenzctl "issue servicecert -d sys.auth --service e2e-svc --dns-domain athenz.cloud --private-key $ADMIN_KEY --csr"
    Then the command should succeed
    And stdout should contain "CERTIFICATE REQUEST"

  Scenario: generate a service certificate CSR with every CSR option
    When I run athenzctl "issue servicecert -d sys.auth --service e2e-svc --provider sys.auth.zts --instance i-1 --instance-id i-legacy --private-key $ADMIN_KEY --dns-domain athenz.cloud --subj-c JP --subj-o Example --subj-ou E2E --spiffe=false --spiffe-trust-domain example.test --ip 127.0.0.1 --attestation-data $TEMP_DIR/attestation.data --signer-key-id 0 --expiry-time 5 --expiry-mins 5 --out $TEMP_DIR/service.cert.pem --signer-cert-out $TEMP_DIR/signer.cert.pem --csr --use-instance-register-token"
    Then the command should succeed
    And stdout should contain "CERTIFICATE REQUEST"

  Scenario: issue instance-register-token
    # sys.auth.zts is the built-in InstanceZTSProvider. The tenant service is
    # created in the fresh test domain; wait for ZTS to sync before requesting.
    Given a service "e2e-svc" exists in domain "$DOMAIN"
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "issue instance-register-token -d $DOMAIN --service e2e-svc --provider sys.auth.zts --instance i-1"
    Then the command should fail with "not authorized to launch"

  Scenario: issue an instance register token to a file
    Given a service "e2e-svc" exists in domain "$DOMAIN"
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "issue instance-register-token -d $DOMAIN --service e2e-svc --provider sys.auth.zts --instance i-1 --out $TEMP_DIR/register.token"
    Then the command should fail with "not authorized to launch"
