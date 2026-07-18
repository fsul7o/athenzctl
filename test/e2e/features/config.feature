@config
Feature: athenzctl config

  Scenario: view current context
    When I run athenzctl "config view"
    Then the command should succeed
    And stdout should contain "current-context: local"

  Scenario: list contexts
    When I run athenzctl "config get-contexts"
    Then the command should succeed
    And stdout should contain "local"

  Scenario: show current-context
    When I run athenzctl "config current-context"
    Then the command should succeed
    And stdout should contain "local"
