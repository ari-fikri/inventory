package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

type TestContext struct {
	Handler      http.Handler
	LastResponse *http.Response
	LastBody     []byte
}

func NewTestContext(handler http.Handler) *TestContext {
	return &TestContext{
		Handler: handler,
	}
}

func (c *TestContext) Close() {
	// Clean up resources if needed
}

func (c *TestContext) SendRequest(method, path string, tenantID string, payload any) error {
	var bodyReader io.Reader
	if payload != nil {
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	}

	recorder := httptest.NewRecorder()
	c.Handler.ServeHTTP(recorder, req)

	resp := recorder.Result()
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to read recorded response body: %w", err)
	}

	c.LastResponse = resp
	c.LastBody = body
	return nil
}
