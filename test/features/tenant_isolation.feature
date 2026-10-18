Feature: Multi-tenant request handling
  As an inventory service
  I want to enforce tenant context on every request
  So that multi-tenant isolation is strictly guaranteed

  Scenario: Request without X-Tenant-ID header is rejected
    When I send a GET request to "/api/v1/items" without a tenant ID
    Then the response status code should be 400
    And the response should contain error "X-Tenant-ID header is required"

  Scenario: Different tenants can create and read items with the exact same SKU independently
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "SHARED-SKU-99" and name "Alpha Product"
    When I am authenticated as tenant "tenant-beta"
    And I create an item with the following details:
      | sku           | SHARED-SKU-99 |
      | name          | Beta Product  |
      | description   | Beta version  |
      | unit_price    | 39.99         |
      | min_threshold | 2             |
    Then the response status code should be 201
    When I get item by SKU "SHARED-SKU-99" as tenant "tenant-alpha"
    Then the response status code should be 200
    And the response should contain item details:
      | sku       | SHARED-SKU-99  |
      | name      | Alpha Product  |
      | tenant_id | tenant-alpha   |
    When I get item by SKU "SHARED-SKU-99" as tenant "tenant-beta"
    Then the response status code should be 200
    And the response should contain item details:
      | sku       | SHARED-SKU-99 |
      | name      | Beta Product  |
      | tenant_id | tenant-beta   |

