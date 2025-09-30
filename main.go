package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.0",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; SM-G998B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Mobile Safari/537.36",
	}
	endpoints = []string{
		"https://autoparse-41617314059.us-east1.run.app/screenshot",
		"https://autoparse-41617314059.us-central1.run.app/screenshot",
		"https://autoparse-41617314059.us-east4.run.app/screenshot",
		"https://autoparse-41617314059.us-east5.run.app/screenshot",
		"https://autoparse-41617314059.europe-west1.run.app/screenshot",
	}
	httpClient = &http.Client{Timeout: 30 * time.Second}
)

// ScreenshotRequest defines the expected JSON payload
type ScreenshotRequest struct {
	URL string `json:"url" binding:"required,uri"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	defaultAddr := ":8080"
	if envPort := strings.TrimSpace(os.Getenv("PORT")); envPort != "" {
		defaultAddr = normalizeAddr(envPort)
	}

	listenAddr := flag.String("port", defaultAddr, "HTTP listen address (e.g. ':8080' or '0.0.0.0:8080')")
	flag.Parse()

	router := gin.Default()
	router.POST("/screenshot", handleScreenshot)

	if err := router.Run(normalizeAddr(*listenAddr)); err != nil {
		fmt.Printf("failed to start server: %v\n", err)
	}
}

func normalizeAddr(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ":8080"
	}
	if strings.HasPrefix(trimmed, ":") || strings.Contains(trimmed, ":") {
		return trimmed
	}
	return ":" + trimmed
}

func handleScreenshot(c *gin.Context) {
	var req ScreenshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload: " + err.Error()})
		return
	}

	encodedURL := url.QueryEscape(req.URL)
	queryParams := fmt.Sprintf("url=%s&disable_js=true&quality=low", encodedURL)

	client := httpClient

	var lastError error

	for _, endpoint := range endpoints {
		userAgent := userAgents[rand.Intn(len(userAgents))]
		reqURL := fmt.Sprintf("%s?%s", endpoint, queryParams)

		httpReq, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			lastError = err
			continue
		}
		httpReq.Header.Set("Accept", "application/json")
		httpReq.Header.Set("User-Agent", userAgent)

		resp, err := client.Do(httpReq)
		if err != nil {
			lastError = err
			continue
		}

		func() {
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				lastError = fmt.Errorf("HTTP status %d from %s", resp.StatusCode, endpoint)
				return
			}

			var result map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				lastError = err
				return
			}

			request, ok := result["request"].(map[string]any)
			if !ok {
				lastError = fmt.Errorf("missing request status in response from %s", endpoint)
				return
			}

			status, _ := request["status"].(string)
			if status != "success" && status != "partial_load" {
				lastError = fmt.Errorf("API status %q from %s", status, endpoint)
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":          status,
				"screenshot_url":  result["screenshot_url"],
				"network_log_url": result["network_log_url"],
				"txs_log_url":     result["txs_log_url"],
				"endpoint":        endpoint,
			})
			lastError = nil
		}()

		if lastError == nil {
			return
		}
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": fmt.Sprintf("all endpoints failed: %v", lastError),
	})
}
