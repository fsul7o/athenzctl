@exec
Feature: athenzctl exec-based auth-mode (kubectl user-exec analog)

  Background:
    Given a fresh athenz stack

  Scenario: a command against an exec context transparently obtains a client cert
    Given I use the "exec-local" context
    When I run athenzctl "get domain sys.auth -o yaml"
    Then the command should succeed
    And stdout should be valid yaml

  Scenario: a broken exec plugin surfaces a clear error
    Given I use the "exec-broken" context
    When I run athenzctl "get domain sys.auth -o yaml"
    Then the command should fail with "exec credential"
