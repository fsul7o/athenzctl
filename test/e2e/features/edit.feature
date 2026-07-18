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
    When I run athenzctl "edit role r1 -d $DOMAIN"
    Then the command should succeed
