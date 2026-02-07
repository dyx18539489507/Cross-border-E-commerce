package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/gin-gonic/gin"
)

type SfxHandler struct {
	cfg *config.Config
}

type SfxItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	File     string `json:"file"`
	Category string `json:"category"`
	URL      string `json:"url"`
}

type sfxConfig struct {
	Items []SfxItem `json:"items"`
}

func NewSfxHandler(cfg *config.Config) *SfxHandler {
	return &SfxHandler{cfg: cfg}
}

func (h *SfxHandler) List(c *gin.Context) {
	category := strings.TrimSpace(c.DefaultQuery("category", "热门"))
	limit := parseInt(c.DefaultQuery("limit", "20"), 20)

	items, err := h.loadItems()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to load sfx"})
		return
	}

	filtered := make([]SfxItem, 0, len(items))
	for _, item := range items {
		if category == "热门" || category == "" || item.Category == category {
			filtered = append(filtered, item)
		}
	}

	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}
	if category == "热门" && limit > 0 && len(filtered) > 0 && len(filtered) < limit {
		// repeat to fill default list size
		orig := filtered
		for len(filtered) < limit {
			for _, item := range orig {
				if len(filtered) >= limit {
					break
				}
				copyItem := item
				copyItem.ID = fmt.Sprintf("%s-%d", item.ID, len(filtered))
				filtered = append(filtered, copyItem)
			}
		}
	}

	c.JSON(200, gin.H{"items": filtered})
}

func (h *SfxHandler) loadItems() ([]SfxItem, error) {
	cfgPath := filepath.Join("configs", "sfx_gaudio.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	var cfg sfxConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	baseURL := strings.TrimRight(h.cfg.Storage.BaseURL, "/")
	for i := range cfg.Items {
		cfg.Items[i].URL = baseURL + "/sfx/" + cfg.Items[i].File
	}
	return cfg.Items, nil
}

func parseInt(v string, def int) int {
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil || n <= 0 {
		return def
	}
	return n
}
