@get
Feature: athenzctl get

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-get"
    And a domain "$DOMAIN" exists

  Scenario Outline: get <kind>
    Given "<kind>" prerequisites exist
    When I run athenzctl "get <kind> <name> -d $DOMAIN <flags> -o yaml"
    Then the command should succeed
    And stdout should be valid yaml

    Examples:
      | kind          | name              | flags        |
      | domain        | $DOMAIN           |              |
      | domain-meta   | $DOMAIN           |              |
      | role          | e2e-role          |              |
      | role-meta     | e2e-role          |              |
      | service       | e2e-svc           |              |
      | servicekey    | e2e-svc:0         |              |
      | policy        | e2e-policy        |              |
      | policyversion | e2e-policy:v1     |              |
      | group         | e2e-group         |              |
      | group-meta    | e2e-group         |              |
      | membership    | user.athenz_admin | --role admin |
      | domain-template|                  |              |
      | quota         |                   |              |

  Scenario: get template
    # Uses the "instance_provider" template shipped by the athenz-distribution default deployment.
    When I run athenzctl "get template instance_provider -o yaml"
    Then the command should succeed

  Scenario: get membership with group and pending filters
    Given a group "readers" exists in domain "$DOMAIN"
    When I run athenzctl "create membership -d $DOMAIN --group readers --member user.groupmember"
    Then the command should succeed
    When I run athenzctl "get membership user.groupmember -d $DOMAIN --group readers -o yaml"
    Then the command should succeed
    And stdout should be valid yaml
    When I run athenzctl "get membership -d $DOMAIN --pending --principal user.athenz_admin -o yaml"
    Then the command should succeed
    And stdout should be valid yaml
