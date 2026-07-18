@patch
Feature: athenzctl patch

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-patch"
    And a domain "$DOMAIN" exists

  Scenario: patch role
    Given a role "role1" exists in domain "$DOMAIN"
    When I run athenzctl "patch role role1 -d $DOMAIN description=patched"
    Then the command should succeed

  Scenario: patch role-meta
    Given a role "role1" exists in domain "$DOMAIN"
    When I run athenzctl "patch role-meta role1 -d $DOMAIN selfServe=true"
    Then the command should succeed

  Scenario: patch policy
    Given a policy "pol1" exists in domain "$DOMAIN"
    When I run athenzctl "patch policy pol1 -d $DOMAIN active=true"
    Then the command should succeed

  Scenario: patch service
    Given a service "svc1" exists in domain "$DOMAIN"
    When I run athenzctl "patch service svc1 -d $DOMAIN description=svc"
    Then the command should succeed

  Scenario: patch group-meta
    Given a group "grp1" exists in domain "$DOMAIN"
    When I run athenzctl "patch group-meta grp1 -d $DOMAIN selfServe=true"
    Then the command should succeed

  Scenario: patch domain-meta
    When I run athenzctl "patch domain-meta $DOMAIN description=updated"
    Then the command should succeed

  Scenario: patch quota
    When I run athenzctl "patch quota -d $DOMAIN role=100 policy=200"
    Then the command should succeed
