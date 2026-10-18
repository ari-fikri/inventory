Feature: Stock Movement and Ledger Management
  As a tenant
  I want to record stock movements for items
  So that I maintain an accurate inventory count and audit trail

  Scenario: Successfully add stock to an existing item (Stock IN)
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "STOCK-ITEM-01" and name "Wireless Mouse"
    When I add 50 units of stock to item with SKU "STOCK-ITEM-01" with reference "PO-1001" and note "Initial shipment"
    Then the response status code should be 200
    And the current stock for item with SKU "STOCK-ITEM-01" should be 50
    And a stock movement record of type "IN" with quantity 50 should be logged
