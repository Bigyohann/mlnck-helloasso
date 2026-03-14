package helloasso

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetForms(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			res := LoginResponse{
				AccessToken: "test-token",
				ExpiresIn:   3600,
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		if r.URL.Path == "/v5/organizations/test-org/forms" {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			res := ListResponse[Form]{
				Data: []Form{
					{Title: "Test Form 1"},
					{Title: "Test Form 2"},
				},
			}
			json.NewEncoder(w).Encode(res)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{
		BaseURL:      server.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Organization: "test-org",
		HTTPClient:   server.Client(),
	}

	forms, err := client.GetForms()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(forms) != 2 {
		t.Errorf("Expected 2 forms, got %d", len(forms))
	}

	if forms[0].Title != "Test Form 1" {
		t.Errorf("Expected Title 'Test Form 1', got '%s'", forms[0].Title)
	}

	// Test cache
	forms2, err := client.GetForms()
	if err != nil {
		t.Fatalf("Expected no error on second call, got %v", err)
	}
	if len(forms2) != 2 {
		t.Errorf("Expected 2 forms from cache, got %d", len(forms2))
	}
}

func TestClient_Login_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid client"))
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	err := client.login()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestClient_GetForms_Errors(t *testing.T) {
	t.Run("Status Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth2/token" {
				json.NewEncoder(w).Encode(LoginResponse{AccessToken: "token"})
				return
			}
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("forbidden"))
		}))
		defer server.Close()

		client := &Client{BaseURL: server.URL, HTTPClient: server.Client()}
		_, err := client.GetForms()
		if err == nil {
			t.Fatal("Expected error on 403, got nil")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth2/token" {
				json.NewEncoder(w).Encode(LoginResponse{AccessToken: "token"})
				return
			}
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := &Client{BaseURL: server.URL, HTTPClient: server.Client()}
		_, err := client.GetForms()
		if err == nil {
			t.Fatal("Expected error on invalid JSON, got nil")
		}
	})

	t.Run("Invalid URL", func(t *testing.T) {
		client := &Client{BaseURL: "http:// invalid-url", HTTPClient: http.DefaultClient}
		client.token = "token"
		client.tokenExpiry = time.Now().Add(time.Hour)
		_, err := client.GetForms()
		if err == nil {
			t.Fatal("Expected error on invalid URL, got nil")
		}
	})
}

func TestClient_Login_JSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, HTTPClient: server.Client()}
	err := client.login()
	if err == nil {
		t.Fatal("Expected error on invalid login JSON, got nil")
	}
}

func TestClient_TokenExpiry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(LoginResponse{AccessToken: "new-token", ExpiresIn: 3600})
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		token:      "old-token",
		tokenExpiry: time.Now().Add(-time.Hour), // expired
	}

	err := client.login()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client.token != "new-token" {
		t.Errorf("Expected token to be renewed, got %s", client.token)
	}
}

func TestClient_GetForms_RequestError(t *testing.T) {
	client := &Client{
		BaseURL:    " http://invalid", // space makes URL invalid for NewRequest
		HTTPClient: http.DefaultClient,
		token:      "token",
		tokenExpiry: time.Now().Add(time.Hour),
	}
	_, err := client.GetForms()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestClient_GetForms_LoginError(t *testing.T) {
	client := &Client{
		BaseURL:    " http://invalid",
		HTTPClient: http.DefaultClient,
	}
	_, err := client.GetForms()
	if err == nil {
		t.Fatal("Expected error when login fails, got nil")
	}
}

func TestClient_Login_NetworkError(t *testing.T) {
	client := &Client{
		BaseURL:    "http://localhost:1", // Should fail to connect
		HTTPClient: http.DefaultClient,
	}
	err := client.login()
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}

func TestClient_GetForms_NetworkError(t *testing.T) {
	client := &Client{
		BaseURL:    "http://localhost:1",
		HTTPClient: http.DefaultClient,
		token:      "token",
		tokenExpiry: time.Now().Add(time.Hour),
	}
	_, err := client.GetForms()
	if err == nil {
		t.Fatal("Expected network error, got nil")
	}
}


func TestSetDefaultClient(t *testing.T) {
	oldClient := defaultClient
	defer func() { defaultClient = oldClient }()

	newClient := &Client{BaseURL: "http://new-client"}
	SetDefaultClient(newClient)
	if defaultClient != newClient {
		t.Errorf("Expected defaultClient to be updated to newClient")
	}
}

func TestGetForms_Integration(t *testing.T) {
	if getDefaultClient() == nil {
		t.Fatal("Default client should be initialized")
	}
	// Test the public wrapper
	// Pointing defaultClient to mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ListResponse[Form]{})
	}))
	defer server.Close()
	
	oldClient := defaultClient
	defer func() { defaultClient = oldClient }()
	
	defaultClient = &Client{
		BaseURL: server.URL,
		HTTPClient: server.Client(),
		token: "token",
		tokenExpiry: time.Now().Add(time.Hour),
	}
	
	_, _ = GetForms()
}

