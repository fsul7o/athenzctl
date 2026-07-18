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
