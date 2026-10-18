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
