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

  Scenario: set a context
    Given a unique context "e2e-config"
    When I run athenzctl "config set-context $CONTEXT --zms-url https://zms.example.test/zms/v1 --zts-url https://zts.example.test/zts/v1"
    Then the command should succeed
    And stdout should contain "$CONTEXT"
    And stdout should contain "created"

  Scenario: get contexts
    Given a unique context "e2e-config"
    When I run athenzctl "config set-context $CONTEXT --zms-url https://zms.example.test/zms/v1"
    Then the command should succeed
    When I run athenzctl "config get-contexts"
    Then the command should succeed
    And stdout should contain "$CONTEXT"

  Scenario: delete a context
    Given a unique context "e2e-config"
    When I run athenzctl "config set-context $CONTEXT --zms-url https://zms.example.test/zms/v1"
    Then the command should succeed
    When I run athenzctl "config delete-context $CONTEXT"
    Then the command should succeed
    And stdout should contain "$CONTEXT"
    And stdout should contain "Deleted context"
    When I run athenzctl "config get-contexts"
    Then the command should succeed
    And stdout should not contain "$CONTEXT"

  Scenario: use an existing context
    Given a fresh athenz stack
    When I run athenzctl "config use-context local"
    Then the command should succeed
    And stdout should contain "local"

  Scenario: save every context connection option
    Given a fresh athenz stack
    And a unique context "e2e-options-context"
    When I run athenzctl "config set-context $CONTEXT --zms-url https://zms.example.test/zms/v1 --zts-url https://zts.example.test/zts/v1 --cert $ADMIN_CERT --key $ADMIN_KEY --ca-cert $ADMIN_CA --zms-server-name zms.example.test --zts-server-name zts.example.test --auth-mode exec --exec-command /bin/true --exec-arg first --exec-arg second --exec-env E2E_OPTION=enabled --exec-cert-path $ADMIN_CERT --exec-key-path $ADMIN_KEY"
    Then the command should succeed
    When I run athenzctl "config view"
    Then the command should succeed
    And stdout should contain "auth-mode: exec"
    And stdout should contain "E2E_OPTION: enabled"
