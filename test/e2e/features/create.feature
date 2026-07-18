@create
Feature: athenzctl create

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-create"
    And a domain "$DOMAIN" exists

  Scenario: create role
    When I run athenzctl "create role role1 -d $DOMAIN --members user.athenz_admin"
    Then the command should succeed
    And I run athenzctl "get role role1 -d $DOMAIN -o yaml"
    And the command should succeed

  Scenario: create service
    When I run athenzctl "create service svc1 -d $DOMAIN"
    Then the command should succeed

  Scenario: create policy
    When I run athenzctl "create policy pol1 -d $DOMAIN"
    Then the command should succeed

  Scenario: create policyversion
    Given a policy "pol1" exists in domain "$DOMAIN"
    When I run athenzctl "create policyversion pol1:v1 -d $DOMAIN --from-version 0"
    Then the command should succeed

  Scenario: create group
    When I run athenzctl "create group grp1 -d $DOMAIN --members user.athenz_admin"
    Then the command should succeed

  Scenario: create membership
    Given a role "role1" exists in domain "$DOMAIN"
    When I run athenzctl "create membership -d $DOMAIN --role role1 --member user.newmember"
    Then the command should succeed

  Scenario: create domain-template
    # Uses the "instance_provider" template shipped by the athenz-distribution
    # default deployment; keywordsToReplace requires --param _provider_ / _dnssuffix_.
    When I run athenzctl "create domain-template instance_provider -d $DOMAIN --param provider=sys.auth.zts --param dnssuffix=athenz.cloud"
    Then the command should succeed

  Scenario: create domains with parent, description, audit, and user options
    Given a unique user "e2e-options-user"
    When I run athenzctl "create domain child -d $DOMAIN --parent $DOMAIN --admin-users user.athenz_admin --description child --audit-ref options-audit"
    Then the command should succeed
    When I run athenzctl "delete domain child -d $DOMAIN --parent $DOMAIN --audit-ref options-audit"
    Then the command should succeed
    When I run athenzctl "create domain --user $USER --description user-domain"
    Then the command should fail with "404"

  Scenario: create delegated role and service with options
    When I run athenzctl "create role delegated -d $DOMAIN --members user.athenz_admin --trust sys.auth --audit-ref options-audit"
    Then the command should succeed
    When I run athenzctl "create service endpoint-svc -d $DOMAIN --provider-endpoint https://provider.example.test --audit-ref options-audit"
    Then the command should succeed

  Scenario: create service keys using file and inline options
    Given a service "key-svc" exists in domain "$DOMAIN"
    And a public key file "$TEMP_DIR/public.pem" exists
    When I run athenzctl "create servicekey key-svc:0 -d $DOMAIN --pem $TEMP_DIR/public.pem --audit-ref options-audit"
    Then the command should succeed
    When I run athenzctl "create servicekey key-svc:1 -d $DOMAIN --key $PUBLIC_KEY"
    Then the command should succeed

  Scenario: exercise membership decision options
    Given a role "reviewers" exists in domain "$DOMAIN"
    When I run athenzctl "patch role reviewers -d $DOMAIN reviewEnabled=true"
    Then the command should succeed
    When I run athenzctl "create membership -d $DOMAIN --role reviewers --member user.approve"
    Then the command should succeed
    When I run athenzctl "create membership -d $DOMAIN --role reviewers --member user.approve --approve"
    Then the command should fail with "cannot approve / reject own request"
    When I run athenzctl "create membership -d $DOMAIN --role reviewers --member user.reject"
    Then the command should succeed
    When I run athenzctl "create membership -d $DOMAIN --role reviewers --member user.reject --reject"
    Then the command should fail with "cannot approve / reject own request"
