@fetch
Feature: athenzctl fetch

  Background:
    Given a fresh athenz stack

  Scenario: fetch signedpolicy for sys.auth
    When I run athenzctl "fetch signedpolicy sys.auth"
    Then the command should succeed

  Scenario: fetch signed policy with output options
    When I run athenzctl "fetch signedpolicy sys.auth --output-dir $TEMP_DIR/policies --policy-version 0 --p1363"
    Then the command should succeed
    And the file "$TEMP_DIR/policies/sys.auth.pol" should exist
