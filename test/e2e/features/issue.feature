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

  @skip
  Scenario: issue rolecert
    # user.athenz_admin is a member of sys.auth:role.admin via bootstrap, so
    # rolecert works without any per-scenario setup. Using a fresh test-domain
    # role instead would race the ZTS domain-update interval.
    When I run athenzctl "issue rolecert --role-domain sys.auth --role-name admin --dns-domain athenz.cloud"
    Then the command should succeed
    And stdout should be a valid PEM certificate

  Scenario: issue servicecert (CSR only)
    When I run athenzctl "issue servicecert -d sys.auth --service e2e-svc --dns-domain athenz.cloud --private-key $ADMIN_KEY --csr"
    Then the command should succeed
    And stdout should contain "CERTIFICATE REQUEST"

  @skip
  Scenario: issue instance-register-token
    # sys.auth.zts is the built-in InstanceZTSProvider. The tenant service is
    # created in the fresh test domain; wait for ZTS to sync before requesting.
    Given a service "e2e-svc" exists in domain "$DOMAIN"
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "issue instance-register-token -d $DOMAIN --service e2e-svc --provider sys.auth.zts --instance i-1"
    Then the command should succeed
