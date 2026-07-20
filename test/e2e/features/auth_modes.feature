@auth-modes
Feature: athenzctl auth-mode ntoken/copperargos end-to-end usage

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-auth-modes"
    And a domain "$DOMAIN" exists
    And a unique context "e2e-auth-modes-context"

  Scenario: ntoken auth-mode mints a service cert via a registered key pair on every invocation
    Given a domain system attribute "certdnsdomain" with value "athenz.cloud" exists in domain "$DOMAIN"
    And ZTS has synced domain "$DOMAIN"
    And a registered service key pair "1" for service "e2e-ntoken-svc" exists in domain "$DOMAIN"
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "config set-context $CONTEXT --zms-url $ZMS_URL --zts-url $ZTS_URL --zms-server-name $ZMS_SNI --zts-server-name $ZTS_SNI --ca-cert $ADMIN_CA --auth-mode ntoken --ntoken-auth-domain $DOMAIN --ntoken-auth-service e2e-ntoken-svc --ntoken-auth-private-key $SVC_KEY_PATH --ntoken-auth-key-id 1 --auth-cache-dir $TEMP_DIR/ntoken-cache --servicecert-dns-domain athenz.cloud"
    Then the command should succeed
    And I use the "$CONTEXT" context
    When I run athenzctl "get domain $DOMAIN -o yaml"
    Then the command should succeed
    And stdout should be valid yaml
    And the file "$TEMP_DIR/ntoken-cache/cert.pem" should exist

  Scenario: copperargos auth-mode registers a service cert using prepared attestation-data
    # The built-in sys.auth.zts InstanceZTSProvider validates the CSR's SAN
    # DNS hostname against its own fixed "zts.athenz.cloud" dns suffix
    # (athenz.zts.provider_dns_suffix server config), regardless of the
    # domain's certDnsDomain, so --servicecert-dns-domain must match it here.
    Given a domain system attribute "certdnsdomain" with value "zts.athenz.cloud" exists in domain "$DOMAIN"
    And ZTS has synced domain "$DOMAIN"
    And a service "e2e-ca-svc" exists in domain "$DOMAIN"
    When I run athenzctl "create domain-template zts_instance_launch_provider -d $DOMAIN --param service=e2e-ca-svc"
    Then the command should succeed
    And ZTS has synced domain "$DOMAIN"
    When I run athenzctl "issue instance-register-token -d $DOMAIN --service e2e-ca-svc --provider sys.auth.zts --instance i-e2e-ca-1 --out $TEMP_DIR/attestation.data"
    Then the command should succeed
    When I run athenzctl "config set-context $CONTEXT --zms-url $ZMS_URL --zts-url $ZTS_URL --zms-server-name $ZMS_SNI --zts-server-name $ZTS_SNI --ca-cert $ADMIN_CA --auth-mode copperargos --copperargos-auth-domain $DOMAIN --copperargos-auth-service e2e-ca-svc --copperargos-auth-provider sys.auth.zts --copperargos-auth-instance i-e2e-ca-1 --copperargos-auth-attestation-data $TEMP_DIR/attestation.data --auth-cache-dir $TEMP_DIR/copperargos-cache --servicecert-dns-domain zts.athenz.cloud"
    Then the command should succeed
    And I use the "$CONTEXT" context
    When I run athenzctl "get domain $DOMAIN -o yaml"
    Then the command should succeed
    And stdout should be valid yaml
    And the file "$TEMP_DIR/copperargos-cache/cert.pem" should exist
