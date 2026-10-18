Feature: Item Catalog Management
  As a tenant
  I want to manage items in my inventory catalog
  So that I can track stock and product information

  Scenario: Successfully create a new item
    Given I am authenticated as tenant "tenant-alpha"
    When I create an item with the following details:
      | sku           | SKU-001             |
      | name          | Ergonomic Keyboard  |
      | description   | Mechanical keyboard |
      | unit_price    | 99.50               |
      | min_threshold | 5                   |
    Then the response status code should be 201
    And the response should contain item details:
      | sku       | SKU-001            |
      | name      | Ergonomic Keyboard |
      | tenant_id | tenant-alpha       |
      | is_active | true               |

  Scenario: Prevent duplicate SKU within the same tenant
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "KEY-DUPLICATE" and name "Original Keyboard"
    When I create an item with the following details:
      | sku           | KEY-DUPLICATE       |
      | name          | Another Keyboard    |
      | description   | Another description |
      | unit_price    | 49.99               |
      | min_threshold | 2                   |
    Then the response status code should be 409
    And the response should contain error "item with this SKU already exists for this tenant"

