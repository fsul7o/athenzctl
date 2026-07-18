@describe
Feature: athenzctl describe

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-describe"
    And a domain "$DOMAIN" exists

  Scenario Outline: describe <kind>
    Given "<kind>" prerequisites exist
    When I run athenzctl "describe <kind> <name> -d $DOMAIN <flags>"
    Then the command should succeed

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
      | template      | instance_provider |              |
      | domain-template|                  |              |
      | quota         |                   |              |

  Scenario: describe a group membership
    Given a group "describe-group" exists in domain "$DOMAIN"
    When I run athenzctl "create membership -d $DOMAIN --group describe-group --member user.describe-group"
    Then the command should succeed
    When I run athenzctl "describe membership user.describe-group -d $DOMAIN --group describe-group"
    Then the command should succeed
