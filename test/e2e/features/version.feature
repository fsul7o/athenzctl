@version
Feature: athenzctl version and root options

  Scenario: execute version with both root flag spellings
    Given a fresh athenz stack
    When I run athenzctl "version --domain ignored.example --output wide"
    Then the command should succeed
    And stdout should contain "athenzctl"
    When I run athenzctl "version -d ignored.example -o table"
    Then the command should succeed
