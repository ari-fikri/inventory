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

  Scenario: Query items that are low in stock
    Given I am authenticated as tenant "tenant-alpha"
    When I create an item with the following details:
      | sku           | LOW-01             |
      | name          | Low Stock Monitor  |
      | description   | 4K Monitor         |
      | unit_price    | 299.99             |
      | min_threshold | 10                 |
    And I add 5 units of stock to item with SKU "LOW-01" with reference "PO-L1" and note "Low initial stock"
    When I create an item with the following details:
      | sku           | HIGH-01            |
      | name          | Abundant Mousepad  |
      | description   | Large gaming pad   |
      | unit_price    | 19.99              |
      | min_threshold | 5                  |
    And I add 20 units of stock to item with SKU "HIGH-01" with reference "PO-H1" and note "Plentiful stock"
    When I query low stock items
    Then the response status code should be 200
    And the low stock list should contain item with SKU "LOW-01"
    And the low stock list should not contain item with SKU "HIGH-01"

  Scenario: Successfully update existing item details
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "UPDATE-001" and name "Old Wireless Mouse"
    When I update item with SKU "UPDATE-001" with the following details:
      | name          | Pro Wireless Mouse                |
      | description   | Ergonomic rechargeable mouse 4Khz |
      | unit_price    | 79.99                             |
      | min_threshold | 8                                 |
    Then the response status code should be 200
    And the response should contain item details:
      | sku           | UPDATE-001                        |
      | name          | Pro Wireless Mouse                |
      | description   | Ergonomic rechargeable mouse 4Khz |
      | unit_price    | 79.99                             |
      | min_threshold | 8                                 |


