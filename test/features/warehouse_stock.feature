Feature: Multi-Location Warehouse Stock Management
  As a tenant with multiple fulfillment facilities
  I want to allocate and track stock per warehouse
  So that I have accurate visibility into inventory across locations

  Scenario: Stock items into specific warehouses independently
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "MULTI-WH-01" and name "Logistics Item"
    When I add 100 units of stock to item with SKU "MULTI-WH-01" in warehouse "WH-NORTH" with reference "PO-WH-1" and note "North delivery"
    And I add 40 units of stock to item with SKU "MULTI-WH-01" in warehouse "WH-SOUTH" with reference "PO-WH-2" and note "South delivery"
    Then the response status code should be 200
    And the stock for item with SKU "MULTI-WH-01" in warehouse "WH-NORTH" should be 100
    And the stock for item with SKU "MULTI-WH-01" in warehouse "WH-SOUTH" should be 40
