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

  Scenario: Adjust stock following a physical audit count (Stock ADJUST)
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "AUDIT-ITEM-01" and name "Headset"
    And I add 50 units of stock to item with SKU "AUDIT-ITEM-01" with reference "PO-INITIAL" and note "Initial stock"
    When I adjust stock for item with SKU "AUDIT-ITEM-01" to target quantity 42 with reference "AUDIT-Q1" and note "Physical count: 8 units damaged"
    Then the response status code should be 200
    And the current stock for item with SKU "AUDIT-ITEM-01" should be 42
    And a stock movement record of type "ADJUST" with quantity 8 should be logged

  Scenario: Successfully fulfill/deduct stock for an item (Stock OUT)
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "OUT-ITEM-01" and name "Gaming Headset"
    And I add 50 units of stock to item with SKU "OUT-ITEM-01" with reference "PO-INIT" and note "Initial stock"
    When I dispatch 20 units of stock from item with SKU "OUT-ITEM-01" with reference "SO-1001" and note "Customer order dispatch"
    Then the response status code should be 200
    And the current stock for item with SKU "OUT-ITEM-01" should be 30
    And a stock movement record of type "OUT" with quantity 20 should be logged

  Scenario: Reject stock dispatch when available stock is insufficient
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "GUARD-ITEM-01" and name "Limited Edition Mouse"
    And I add 10 units of stock to item with SKU "GUARD-ITEM-01" with reference "PO-SMALL" and note "Limited batch"
    When I dispatch 25 units of stock from item with SKU "GUARD-ITEM-01" with reference "SO-OVER" and note "Excessive order"
    Then the response status code should be 422
    And the response should contain error "insufficient stock for fulfillment"
    And the current stock for item with SKU "GUARD-ITEM-01" should be 10

  Scenario: Retrieve chronological stock movement history and audit trail
    Given I am authenticated as tenant "tenant-alpha"
    And an item exists with SKU "HIST-01" and name "History Tracked Item"
    And I add 100 units of stock to item with SKU "HIST-01" with reference "PO-100" and note "Initial batch"
    And I dispatch 30 units of stock from item with SKU "HIST-01" with reference "SO-200" and note "Customer shipment"
    And I adjust stock for item with SKU "HIST-01" to target quantity 65 with reference "AUDIT-300" and note "Discrepancy audit"
    When I query stock movements for item with SKU "HIST-01"
    Then the response status code should be 200
    And the stock movements history should contain 3 entries in reverse chronological order:
      | movement_type | quantity | balance_after | reference |
      | ADJUST        | 5        | 65            | AUDIT-300 |
      | OUT           | 30       | 70            | SO-200    |
      | IN            | 100      | 100           | PO-100    |




