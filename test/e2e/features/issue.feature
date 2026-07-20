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

  Scenario: issue an access token with proxy for principal
    # user.athenz_admin is listed in ZTS's athenz.zts.authorized_proxy_users
    # config (see scripts/patches/athenz-distribution-zts-authorized-proxy-users.patch),
    # so it is allowed to request a token proxied on its own behalf.
    When I run athenzctl "issue accesstoken -d sys.auth --role admin --proxy-for-principal user.athenz_admin -o json"
    Then the command should succeed
    And stdout should be valid json
    And stdout should contain "access_token"

  Scenario: issue an access token with an unauthorized authorization-details request
    # sys.auth:role.admin has no configured AuthzDetailsEntity and this e2e
    # stack has no system-wide authorization details configured, so any
    # non-empty authorization_details value is rejected regardless of the
    # proxy-for-principal check above.
    When I run athenzctl "issue accesstoken -d sys.auth --role admin --authorization-details {} -o json"
    Then the command should fail with "Authorization Details not valid for this request"

  Scenario: issue rolecert
    # user.athenz_admin is a member of sys.auth:role.admin via bootstrap, so
    # rolecert works without any per-scenario setup.
    When I run athenzctl "issue rolecert --role-domain sys.auth --role-name admin --dns-domain .athenz.cloud"
    Then the command should succeed
    And stdout should be a valid PEM certificate

  Scenario: generate a role certificate CSR with every CSR option
    When I run athenzctl "issue rolecert --domain sys.auth --role-domain sys.auth --role-name admin --service e2e-svc --role-key-file $ADMIN_KEY --dns-domain athenz.cloud --subj-c JP --subj-p Tokyo --subj-o Example --subj-ou E2E --spiffe=false --spiffe-trust-domain example.test --ip 127.0.0.1 --old-role-cert $ADMIN_CERT --csr --expiry-time 5 --out $TEMP_DIR/role.cert.pem --proxy-for-principal user.athenz_admin --concat-intermediate-cert --cacert-bundle-name athenz --signer-key-id 0"
    Then the command should succeed
    And stdout should contain "CERTIFICATE REQUEST"

  Scenario: issue instance-register-token
    # sys.auth.zts is the built-in InstanceZTSProvider. The tenant service is
    # created in the fresh test domain and authorized to launch by ZTS.
    Given a service "e2e-svc" exists in domain "$DOMAIN"
    When I run athenzctl "create domain-template zts_instance_launch_provider -d $DOMAIN --param service=e2e-svc"
    Then the command should succeed
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "issue instance-register-token -d $DOMAIN --service e2e-svc --provider sys.auth.zts --instance i-1"
    Then the command should succeed

  Scenario: issue an instance register token to a file
    Given a service "e2e-svc" exists in domain "$DOMAIN"
    When I run athenzctl "create domain-template zts_instance_launch_provider -d $DOMAIN --param service=e2e-svc"
    Then the command should succeed
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "issue instance-register-token -d $DOMAIN --service e2e-svc --provider sys.auth.zts --instance i-1 --out $TEMP_DIR/register.token"
    Then the command should succeed
    And the file "$TEMP_DIR/register.token" should exist
