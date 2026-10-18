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
