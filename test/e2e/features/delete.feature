@delete
Feature: athenzctl delete

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-delete"
    And a domain "$DOMAIN" exists

  Scenario: delete role
    Given a role "role1" exists in domain "$DOMAIN"
    When I run athenzctl "delete role role1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete service
    Given a service "svc1" exists in domain "$DOMAIN"
    When I run athenzctl "delete service svc1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete policy
    Given a policy "pol1" exists in domain "$DOMAIN"
    When I run athenzctl "delete policy pol1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete group
    Given a group "grp1" exists in domain "$DOMAIN"
    When I run athenzctl "delete group grp1 -d $DOMAIN"
    Then the command should succeed
