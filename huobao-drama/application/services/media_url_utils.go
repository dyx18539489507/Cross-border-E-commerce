package services

import (
	"net/url"
	"strings"
)

// NormalizeImageURLForClient 统一将图片URL规范为前端可稳定加载的格式。
func NormalizeImageURLForClient(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "blob:") || strings.HasPrefix(value, "data:") {
		return value
	}

	if strings.HasPrefix(value, "/api/v1/media/proxy") {
		parsed, err := url.Parse(value)
		if err != nil {
			return value
		}
		target := parsed.Query().Get("url")
		if target == "" {
			return value
		}
		if decoded, err := url.QueryUnescape(target); err == nil {
			target = decoded
		}
		return NormalizeImageURLForClient(target)
	}

	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		parsed, err := url.Parse(value)
		if err != nil {
			return value
		}
		if strings.HasPrefix(parsed.Path, "/static/") {
			if parsed.RawQuery != "" {
				return parsed.Path + "?" + parsed.RawQuery
			}
			return parsed.Path
		}
		return value
	}

	if strings.HasPrefix(value, "/static/") {
		return value
	}
	if strings.HasPrefix(value, "/data/") {
		return "/static" + value
	}
	if strings.HasPrefix(value, "data/") {
		return "/static/" + value
	}
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/static/" + strings.TrimPrefix(value, "/")
}

func NormalizeImageURLPtr(raw *string) {
	if raw == nil {
		return
	}
	normalized := NormalizeImageURLForClient(*raw)
	if normalized != "" {
		*raw = normalized
	}
}
