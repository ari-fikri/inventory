Feature: Multi-tenant request handling
  As an inventory service
  I want to enforce tenant context on every request
  So that multi-tenant isolation is strictly guaranteed

  Scenario: Request without X-Tenant-ID header is rejected
    When I send a GET request to "/api/v1/items" without a tenant ID
    Then the response status code should be 400
    And the response should contain error "X-Tenant-ID header is required"
