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

  Scenario: Successfully transfer stock between warehouses
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "TRANSFER-01" and name "Transferable Widget"
    And I add 100 units of stock to item with SKU "TRANSFER-01" in warehouse "WH-ORIGIN" with reference "PO-INIT" and note "Initial batch"
    When I transfer 35 units of stock for item with SKU "TRANSFER-01" from warehouse "WH-ORIGIN" to warehouse "WH-DEST" with reference "TRF-901" and note "Rebalancing"
    Then the response status code should be 200
    And the stock for item with SKU "TRANSFER-01" in warehouse "WH-ORIGIN" should be 65
    And the stock for item with SKU "TRANSFER-01" in warehouse "WH-DEST" should be 35

  Scenario: Reject warehouse transfer when source warehouse has insufficient stock
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "TRANSFER-LOW" and name "Low Stock Item"
    And I add 10 units of stock to item with SKU "TRANSFER-LOW" in warehouse "WH-ORIGIN" with reference "PO-INIT" and note "Small batch"
    When I transfer 50 units of stock for item with SKU "TRANSFER-LOW" from warehouse "WH-ORIGIN" to warehouse "WH-DEST" with reference "TRF-902" and note "Excessive transfer"
    Then the response status code should be 422
    And the response should contain error "insufficient stock for transfer in source warehouse"
    And the stock for item with SKU "TRANSFER-LOW" in warehouse "WH-ORIGIN" should be 10
    And the stock for item with SKU "TRANSFER-LOW" in warehouse "WH-DEST" should be 0


