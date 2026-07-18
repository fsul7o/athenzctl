@fetch
Feature: athenzctl fetch

  Background:
    Given a fresh athenz stack

  Scenario: fetch signedpolicy for sys.auth
    When I run athenzctl "fetch signedpolicy sys.auth"
    Then the command should succeed
