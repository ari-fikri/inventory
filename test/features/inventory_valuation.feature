Feature: Inventory Valuation and Cost Layer Tracking
  As a tenant warehouse or finance manager
  I want inventory receipts and dispatches to track unit costs and calculate asset valuation using FIFO
  So that I have an accurate Cost of Goods Sold (COGS) and financial balance sheet

  Scenario: Receive inventory with unit cost creating FIFO cost layers and calculate inventory asset value
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "FIFO-VAL-01" and name "Valued Widget"
    When I add 10 units of stock to item with SKU "FIFO-VAL-01" with unit cost 15.50 and reference "PO-100"
    And I add 20 units of stock to item with SKU "FIFO-VAL-01" with unit cost 20.00 and reference "PO-101"
    Then the response status code should be 200
    When I query inventory valuation for item with SKU "FIFO-VAL-01"
    Then the response status code should be 200
    And the valuation response should show total quantity 30 and total value 555.00

  Scenario: Dispatch stock using FIFO calculating Cost of Goods Sold
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "FIFO-OUT-01" and name "COGS Widget"
    And I add 10 units of stock to item with SKU "FIFO-OUT-01" with unit cost 10.00 and reference "BATCH-1"
    And I add 10 units of stock to item with SKU "FIFO-OUT-01" with unit cost 20.00 and reference "BATCH-2"
    When I dispatch 15 units of stock from item with SKU "FIFO-OUT-01" with reference "ORDER-99" and note "Customer order"
    Then the response status code should be 200
    And the response should contain total COGS 200.00
    When I query inventory valuation for item with SKU "FIFO-OUT-01"
    Then the response status code should be 200
    And the valuation response should show total quantity 5 and total value 100.00

