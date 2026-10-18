Feature: Purchase Order and Supplier Receiving
  As a warehouse procurement manager
  I want to manage suppliers, create purchase orders, and receive shipments directly into warehouse stock
  So that inventory restocks are fully audited with real purchase costs

  Scenario: Create a supplier and draft a purchase order with line items
    Given I am authenticated as tenant "tenant-alpha"
    When I create a supplier with name "Apex Electronics", email "vendor@apex.com", and phone "555-0100"
    Then the response status code should be 201
    And an item exists with SKU "PO-ITEM-01" and name "Headphones"
    When I create a purchase order with PO number "PO-1001", supplier "Apex Electronics", destination warehouse "WH-MAIN", and line items:
      | sku        | quantity | unit_cost |
      | PO-ITEM-01 | 100      | 45.00     |
    Then the response status code should be 201
    And the purchase order status should be "DRAFT"

  Scenario: Submit purchase order and receive partial delivery into warehouse stock
    Given I am authenticated as tenant "tenant-alpha"
    And a supplier exists with name "Beta Logistics", email "ops@beta.com", and phone "555-0200"
    And an item exists with SKU "PO-ITEM-02" and name "Microphones"
    And a draft purchase order "PO-1002" exists for supplier "Beta Logistics" in warehouse "WH-EAST" with items:
      | sku        | quantity | unit_cost |
      | PO-ITEM-02 | 100      | 50.00     |
    When I submit purchase order "PO-1002"
    Then the response status code should be 200
    And the purchase order status should be "ORDERED"
    When I receive delivery for purchase order "PO-1002" with items:
      | sku        | received_quantity |
      | PO-ITEM-02 | 60                |
    Then the response status code should be 200
    And the purchase order status should be "PARTIALLY_RECEIVED"
    And the stock for item with SKU "PO-ITEM-02" in warehouse "WH-EAST" should be 60
    When I query inventory valuation for item with SKU "PO-ITEM-02"
    Then the response status code should be 200
    And the valuation response should show total quantity 60 and total value 3000.00

