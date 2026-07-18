@check
Feature: athenzctl check

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-check"
    And a domain "$DOMAIN" exists
    And a role "readers" exists in domain "$DOMAIN"

  Scenario: check access for admin
    When I run athenzctl "check access read $DOMAIN:resource1 -d $DOMAIN --principal user.athenz_admin"
    Then the command should succeed

  Scenario: check resource enumerates admin's access
    When I run athenzctl "check resource --principal user.athenz_admin"
    Then the command should succeed
