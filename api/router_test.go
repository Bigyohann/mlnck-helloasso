package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mlnck/pkg/helloasso"

	"github.com/gin-gonic/gin"
)

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Default Origins", func(t *testing.T) {
		os.Setenv("ALLOWED_ORIGINS", "")
		r := SetupRouter()
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/forms", nil)
		r.ServeHTTP(w, req)
		// We don't care about the response body here, just that the router is set up
	})

	t.Run("Custom Origins", func(t *testing.T) {
		os.Setenv("ALLOWED_ORIGINS", "http://test.com,http://other.com")
		r := SetupRouter()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/forms", nil)
		req.Header.Set("Origin", "http://test.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			t.Errorf("Expected CORS preflight success, got %d", w.Code)
		}
		
		if w.Header().Get("Access-Control-Allow-Origin") != "http://test.com" {
			t.Errorf("Expected Allow-Origin http://test.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("Rate Limit", func(t *testing.T) {
		// Mock helloasso to prevent 500 errors
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/oauth2/token" {
				json.NewEncoder(w).Encode(helloasso.LoginResponse{AccessToken: "token", ExpiresIn: 3600})
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		
		client := helloasso.NewClient()
		client.BaseURL = server.URL
		helloasso.SetDefaultClient(client)

		os.Setenv("ALLOWED_ORIGINS", "*")
		r := SetupRouter()

		// Send many requests quickly to trigger the rate limit (10 req/100ms)
		hitLimit := false
		for i := 0; i < 20; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/forms", nil)
			req.RemoteAddr = "1.2.3.4:1234"
			r.ServeHTTP(w, req)
			
			if w.Code == http.StatusTooManyRequests {
				hitLimit = true
				break
			}
		}
		
		if !hitLimit {
			t.Error("Expected to hit rate limit (429), but didn't")
		}
	})
}
