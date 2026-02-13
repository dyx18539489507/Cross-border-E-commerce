package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	log *logger.Logger
}

func NewMediaHandler(log *logger.Logger) *MediaHandler {
	return &MediaHandler{log: log}
}

func (h *MediaHandler) Proxy(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("url"))
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}
	if !isSafeMediaURL(raw) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		h.log.Warnw("Media proxy failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "proxy failed"})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func isSafeMediaURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(u.Scheme), "http")
}
