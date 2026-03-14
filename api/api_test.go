package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mlnck/pkg/helloasso"

	"github.com/gin-gonic/gin"
)

func TestGetFormsHandler(t *testing.T) {
	// Mock HelloAsso API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			json.NewEncoder(w).Encode(helloasso.LoginResponse{AccessToken: "token"})
			return
		}
		res := helloasso.ListResponse[helloasso.Form]{
			Data: []helloasso.Form{
				{Title: "B", StartDate: "2024-02-01"},
				{Title: "A", StartDate: "2024-01-01"},
			},
		}
		json.NewEncoder(w).Encode(res)
	}))
	defer server.Close()

	// Configure helloasso client to use mock server
	client := helloasso.NewClient()
	client.BaseURL = server.URL
	helloasso.SetDefaultClient(client)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/forms", GetFormsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forms", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var results []helloasso.Form
	json.NewDecoder(w.Body).Decode(&results)

	if len(results) != 2 {
		t.Errorf("Expected 2 forms, got %d", len(results))
	}

	// Verify sorting (A should be first because 2024-01-01 < 2024-02-01)
	if results[0].Title != "A" {
		t.Errorf("Expected first form to be 'A', got '%s'", results[0].Title)
	}
}

func TestGetFormsHandler_Error(t *testing.T) {
	// Mock HelloAsso API with error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/oauth2/token" {
			json.NewEncoder(w).Encode(helloasso.LoginResponse{AccessToken: "token"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := helloasso.NewClient()
	client.BaseURL = server.URL
	helloasso.SetDefaultClient(client)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/forms", GetFormsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/forms", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestProxyImageHandler(t *testing.T) {
	// Mock image server
	imageServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("fake-image-content"))
	}))
	defer imageServer.Close()

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/proxy-image", ProxyImageHandler)

	t.Run("Valid URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/proxy-image?url="+imageServer.URL, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		if w.Header().Get("Content-Type") != "image/png" {
			t.Errorf("Expected Content-Type image/png, got %s", w.Header().Get("Content-Type"))
		}

		if w.Header().Get("Cross-Origin-Resource-Policy") != "cross-origin" {
			t.Errorf("Expected CORP header, got %s", w.Header().Get("Cross-Origin-Resource-Policy"))
		}

		if w.Body.String() != "fake-image-content" {
			t.Errorf("Expected body 'fake-image-content', got '%s'", w.Body.String())
		}
	})

	t.Run("Missing URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/proxy-image", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("Invalid URL", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/proxy-image?url=http://invalid-url-that-does-not-exist-123.com", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
	})
}
