@lookup
Feature: athenzctl lookup

  Background:
    Given a fresh athenz stack
    And a unique domain "e2e-lookup"
    And a domain "$DOMAIN" exists
    And a role "lookup-role" exists in domain "$DOMAIN"
    And a domain system attribute "account" with value "e2e-lookup-account" exists in domain "$DOMAIN"
    And a domain system attribute "azuresubscription" with value "e2e-lookup-azure" exists in domain "$DOMAIN"
    And a domain system attribute "gcpproject" with value "e2e-lookup-gcp" exists in domain "$DOMAIN"
    And a domain system attribute "productid" with value "e2e-lookup-product" exists in domain "$DOMAIN"
    And a domain system attribute "businessservice" with value "e2e-lookup-business" exists in domain "$DOMAIN"
    And a domain tag "lookup" with value "e2e-lookup-tag" exists in domain "$DOMAIN"

  Scenario Outline: lookup domain by <description>
    When I run athenzctl "lookup domain --field-selector <selector>"
    Then the command should succeed
    And stdout should contain "$DOMAIN"

    Examples:
      | description         | selector                                                   |
      | role                | member=user.athenz_admin,role=lookup-role                |
      | tag                 | tagKey=lookup,tagValue=e2e-lookup-tag                     |
      | AWS account         | account=e2e-lookup-account                                |
      | Azure subscription  | azure=e2e-lookup-azure                                    |
      | GCP project         | gcp=e2e-lookup-gcp                                         |
      | product ID          | productId=e2e-lookup-product                              |
      | business service    | businessService=e2e-lookup-business                       |

  Scenario: lookup domain returns structured output
    When I run athenzctl "lookup domain --field-selector account=e2e-lookup-account -o json"
    Then the command should succeed
    And stdout should be valid json
    When I run athenzctl "lookup domain --field-selector account=e2e-lookup-account -o yaml"
    Then the command should succeed
    And stdout should be valid yaml

  Scenario: lookup domain rejects an incomplete role selector
    When I run athenzctl "lookup domain --field-selector role=lookup-role"
    Then the command should fail with "role requires member"
