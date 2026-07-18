@edit
Feature: athenzctl edit
  # steps.go writes a fake editor to the scenario tempDir and points
  # $ATHENZCTL_EDITOR at it. The stub appends a trailing YAML comment to
  # force a non-empty diff so `edit` will PUT the resource back.

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-edit"
    And a domain "$DOMAIN" exists

  Scenario: edit role via fake editor
    Given a role "r1" exists in domain "$DOMAIN"
    When I run athenzctl "edit role r1 -d $DOMAIN --audit-ref edit-audit"
    Then the command should succeed

  Scenario Outline: edit every supported kind via fake editor
    Given "<kind>" prerequisites exist
    When I run athenzctl "edit <kind> <name> -d $DOMAIN --audit-ref edit-audit"
    Then the command should succeed

    Examples:
      | kind          | name              |
      | domain-meta   |                   |
      | quota         |                   |
      | role          | e2e-role          |
      | policy        | e2e-policy        |
      | policyversion | e2e-policy:v1     |
      | service       | e2e-svc           |
      | group         | e2e-group         |
      | role-meta     | e2e-role          |
      | group-meta    | e2e-group         |
