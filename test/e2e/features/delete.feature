@delete
Feature: athenzctl delete

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-delete"
    And a domain "$DOMAIN" exists

  Scenario: delete role
    Given a role "role1" exists in domain "$DOMAIN"
    When I run athenzctl "delete role role1 -d $DOMAIN --audit-ref delete-audit"
    Then the command should succeed

  Scenario: delete service
    Given a service "svc1" exists in domain "$DOMAIN"
    When I run athenzctl "delete service svc1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete policy
    Given a policy "pol1" exists in domain "$DOMAIN"
    When I run athenzctl "delete policy pol1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete group
    Given a group "grp1" exists in domain "$DOMAIN"
    When I run athenzctl "delete group grp1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete a role membership
    Given a role "role-membership" exists in domain "$DOMAIN"
    When I run athenzctl "create membership -d $DOMAIN --role role-membership --member user.delete-role"
    Then the command should succeed
    When I run athenzctl "delete membership -d $DOMAIN --role role-membership --member user.delete-role --audit-ref delete-audit"
    Then the command should succeed

  Scenario: delete a group membership
    Given a group "group-membership" exists in domain "$DOMAIN"
    When I run athenzctl "create membership -d $DOMAIN --group group-membership --member user.delete-group"
    Then the command should succeed
    When I run athenzctl "delete membership -d $DOMAIN --group group-membership --member user.delete-group"
    Then the command should succeed

  Scenario: delete a service key
    Given a service "key-service" exists in domain "$DOMAIN"
    When I run athenzctl "create servicekey key-service:0 -d $DOMAIN --key $PUBLIC_KEY"
    Then the command should succeed
    When I run athenzctl "delete servicekey key-service:0 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete a policy version
    Given "policyversion" prerequisites exist
    When I run athenzctl "delete policyversion e2e-policy:v1 -d $DOMAIN"
    Then the command should succeed

  Scenario: delete a domain template
    When I run athenzctl "create domain-template instance_provider -d $DOMAIN --param provider=sys.auth.zts --param dnssuffix=athenz.cloud"
    Then the command should succeed
    When I run athenzctl "delete domain-template instance_provider -d $DOMAIN"
    Then the command should succeed

  Scenario: reset a quota
    When I run athenzctl "patch quota -d $DOMAIN role=100"
    Then the command should succeed
    When I run athenzctl "delete quota -d $DOMAIN --audit-ref delete-audit"
    Then the command should succeed

  Scenario: delete a user domain through the CLI
    Given a unique user "e2e-delete-user"
    When I run athenzctl "delete domain --user $USER"
    Then the command should fail with "404"
