package services

import "strings"

// normalizeVolcesEndpoint ensures the /api/v3 prefix is applied exactly once.
// It supports base URLs with or without a trailing /api/v3 segment.
func normalizeVolcesEndpoint(baseURL, endpoint, defaultEndpoint string) string {
	ep := strings.TrimSpace(endpoint)
	if ep == "" {
		ep = strings.TrimSpace(defaultEndpoint)
	}
	if ep == "" {
		return ""
	}

	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	hasAPIV3 := strings.Contains(base, "/api/v3")

	if hasAPIV3 {
		if strings.HasPrefix(ep, "/api/v3/") {
			ep = strings.TrimPrefix(ep, "/api/v3")
		} else if ep == "/api/v3" {
			ep = "/"
		}
	} else {
		if !strings.HasPrefix(ep, "/api/v3/") && ep != "/api/v3" {
			if strings.HasPrefix(ep, "/") {
				ep = "/api/v3" + ep
			} else {
				ep = "/api/v3/" + ep
			}
		}
	}

	if !strings.HasPrefix(ep, "/") {
		ep = "/" + ep
	}

	return ep
}
