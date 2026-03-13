package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DeviceIDHeader       = "X-Device-ID"
	DeviceIDCookie       = "drama_device_id"
	DeviceIDContextKey   = "device_id"
	deviceIDCookieMaxAge = 365 * 24 * 60 * 60
)

var deviceIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{16,128}$`)

func DeviceIdentityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := normalizeDeviceID(c.GetHeader(DeviceIDHeader))
		if deviceID == "" {
			if cookieValue, err := c.Cookie(DeviceIDCookie); err == nil {
				deviceID = normalizeDeviceID(cookieValue)
			}
		}
		if deviceID == "" {
			deviceID = generateDeviceID()
		}

		c.Set(DeviceIDContextKey, deviceID)
		c.Request.Header.Set(DeviceIDHeader, deviceID)
		c.Header(DeviceIDHeader, deviceID)
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     DeviceIDCookie,
			Value:    deviceID,
			Path:     "/",
			MaxAge:   deviceIDCookieMaxAge,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   c.Request.TLS != nil,
		})

		c.Next()
	}
}

func GetDeviceID(c *gin.Context) string {
	if value, exists := c.Get(DeviceIDContextKey); exists {
		if deviceID, ok := value.(string); ok && deviceID != "" {
			return deviceID
		}
	}

	return ""
}

func normalizeDeviceID(raw string) string {
	deviceID := strings.TrimSpace(raw)
	if deviceID == "" {
		return ""
	}
	if !deviceIDPattern.MatchString(deviceID) {
		return ""
	}
	return deviceID
}

func generateDeviceID() string {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "dev_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return "dev_" + hex.EncodeToString(randomBytes)
}
