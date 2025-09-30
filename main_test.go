package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleScreenshotSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodGet {
			t.Fatalf("expected GET request, got %s", got)
		}
		if got := r.URL.Query().Get("url"); got != "https://example.com" {
			t.Fatalf("expected encoded URL parameter, got %s", got)
		}

		resp := map[string]any{
			"request":         map[string]any{"status": "success"},
			"screenshot_url":  "https://cdn.example/screenshot.png",
			"network_log_url": "https://cdn.example/network.log",
			"txs_log_url":     "https://cdn.example/trace.log",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	prevEndpoints := endpoints
	prevUserAgents := userAgents
	prevClient := httpClient

	endpoints = []string{srv.URL}
	userAgents = []string{"test-agent"}
	httpClient = srv.Client()

	t.Cleanup(func() {
		endpoints = prevEndpoints
		userAgents = prevUserAgents
		httpClient = prevClient
	})

	router := gin.Default()
	router.POST("/screenshot", handleScreenshot)

	body := bytes.NewBufferString(`{"url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/screenshot", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["status"] != "success" {
		t.Fatalf("expected success status, got %#v", resp["status"])
	}
	if resp["endpoint"] != srv.URL {
		t.Fatalf("expected endpoint %q, got %q", srv.URL, resp["endpoint"])
	}
}

func TestHandleScreenshotFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "backend failure", http.StatusBadGateway)
	}))
	defer srv.Close()

	prevEndpoints := endpoints
	prevClient := httpClient

	endpoints = []string{srv.URL}
	httpClient = srv.Client()

	t.Cleanup(func() {
		endpoints = prevEndpoints
		httpClient = prevClient
	})

	router := gin.Default()
	router.POST("/screenshot", handleScreenshot)

	body := bytes.NewBufferString(`{"url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/screenshot", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
