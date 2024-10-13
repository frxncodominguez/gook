package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test configuration to the file
	testConfig := `{
		"webhooks": [
			{
				"name": "Test Webhook",
				"path": "/test",
				"outputs": [
					{
						"url": "https://example.com/endpoint",
						"condition": "{{ .field }} == 'value'",
						"template": {
							"newField": "Value: {{ .field }}"
						}
					}
				]
			}
		]
	}`
	if _, err := tempFile.Write([]byte(testConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test loading the configuration
	config, err := loadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Assert the loaded configuration
	if len(config.Webhooks) != 1 {
		t.Errorf("Expected 1 webhook, got %d", len(config.Webhooks))
	}
	if config.Webhooks[0].Name != "Test Webhook" {
		t.Errorf("Expected webhook name 'Test Webhook', got '%s'", config.Webhooks[0].Name)
	}
	if config.Webhooks[0].Path != "/test" {
		t.Errorf("Expected webhook path '/test', got '%s'", config.Webhooks[0].Path)
	}
}

func TestCheckDuplicatePaths(t *testing.T) {
	testCases := []struct {
		name     string
		webhooks []Webhook
		wantErr  bool
	}{
		{
			name: "No duplicates",
			webhooks: []Webhook{
				{Path: "/webhook1"},
				{Path: "/webhook2"},
			},
			wantErr: false,
		},
		{
			name: "With duplicates",
			webhooks: []Webhook{
				{Path: "/webhook1"},
				{Path: "/webhook1"},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := checkDuplicatePaths(tc.webhooks)
			if (err != nil) != tc.wantErr {
				t.Errorf("checkDuplicatePaths() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestHandleWebhook(t *testing.T) {
	// Mock HTTP server to receive forwarded requests
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	webhook := Webhook{
		Name: "Test Webhook",
		Path: "/test",
		Outputs: []Output{
			{
				URL:       mockServer.URL,
				Condition: "{{ .field }} == 'value'",
				Template: map[string]interface{}{
					"newField": "Value: {{ .field }}",
				},
			},
		},
	}

	// Test request
	payload := map[string]interface{}{"field": "value"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleWebhook(webhook, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}
}

func TestMain(m *testing.M) {
	// Setup code here (if needed)

	// Run tests
	exitCode := m.Run()

	// Teardown code here (if needed)

	// Exit with the status from the tests
	os.Exit(exitCode)
}
