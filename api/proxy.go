// Package api provides the API handlers for the mlnck application.
package api

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ProxyImageHandler(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image"})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", resp.Header.Get("Content-Type"))
	c.Header("Cross-Origin-Resource-Policy", "cross-origin")
	c.Header("Cache-Control", "public, max-age=3600")
	_, _ = io.Copy(c.Writer, resp.Body)
}
