package handlers

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type VoiceLibraryHandler struct {
	service          *services.VoiceLibraryService
	speechTTSService *services.SpeechTTSService
	resourceCacheMu  sync.RWMutex
	resourceCache    map[string]resourceAccessCacheEntry
	log              *logger.Logger
}

type resourceAccessCacheEntry struct {
	Allowed   bool
	CheckedAt time.Time
}

var blockedVoiceKeywords = []string{
	"娇喘",
}

func NewVoiceLibraryHandler(cfg *config.Config, log *logger.Logger) (*VoiceLibraryHandler, error) {
	service, err := services.NewVoiceLibraryService(cfg, log)
	if err != nil {
		return nil, err
	}

	return &VoiceLibraryHandler{
		service:          service,
		speechTTSService: services.NewSpeechTTSService(cfg, log),
		resourceCache:    make(map[string]resourceAccessCacheEntry),
		log:              log,
	}, nil
}

func (h *VoiceLibraryHandler) List(c *gin.Context) {
	voices, err := h.service.ListSpeakers(c.Request.Context())
	if err != nil {
		h.log.Errorw("Failed to list voice library", "error", err)
		response.InternalError(c, err.Error())
		return
	}

	voices = h.filterUsableVoices(c.Request.Context(), voices)
	response.Success(c, voices)
}

func (h *VoiceLibraryHandler) filterUsableVoices(ctx context.Context, voices []services.VoiceSpeaker) []services.VoiceSpeaker {
	if len(voices) == 0 {
		return voices
	}
	if h.speechTTSService == nil || !h.speechTTSService.IsConfigured() {
		return []services.VoiceSpeaker{}
	}

	allowedResources := make(map[string]bool)
	for _, voice := range voices {
		if isBlockedVoice(voice.Name) {
			continue
		}
		resourceID := strings.TrimSpace(voice.ResourceID)
		voiceType := strings.TrimSpace(voice.VoiceType)
		if resourceID == "" || voiceType == "" {
			continue
		}
		if _, ok := allowedResources[resourceID]; ok {
			continue
		}
		allowedResources[resourceID] = h.checkResourceAccess(ctx, resourceID, voiceType)
	}

	filtered := make([]services.VoiceSpeaker, 0, len(voices))
	for _, voice := range voices {
		if isBlockedVoice(voice.Name) {
			continue
		}
		resourceID := strings.TrimSpace(voice.ResourceID)
		if resourceID == "" {
			continue
		}
		if allowedResources[resourceID] {
			filtered = append(filtered, voice)
		}
	}
	return filtered
}

func (h *VoiceLibraryHandler) checkResourceAccess(ctx context.Context, resourceID, voiceType string) bool {
	cacheKey := strings.TrimSpace(resourceID)
	if cacheKey == "" {
		return false
	}

	h.resourceCacheMu.RLock()
	cached, ok := h.resourceCache[cacheKey]
	h.resourceCacheMu.RUnlock()
	if ok && time.Since(cached.CheckedAt) < 10*time.Minute {
		return cached.Allowed
	}

	probeCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	_, err := h.speechTTSService.SynthesizeToURLWithResource(probeCtx, "你好", strings.TrimSpace(voiceType), cacheKey)
	allowed := true
	if err != nil {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "requested resource not granted"),
			strings.Contains(msg, "resource ID is mismatched with speaker related resource"),
			strings.Contains(msg, "invalid auth token"),
			strings.Contains(msg, "missing Authorization header"):
			allowed = false
		default:
			// Keep the voice visible on transient upstream/network errors.
			allowed = true
		}
	}

	h.resourceCacheMu.Lock()
	h.resourceCache[cacheKey] = resourceAccessCacheEntry{
		Allowed:   allowed,
		CheckedAt: time.Now(),
	}
	h.resourceCacheMu.Unlock()
	return allowed
}

func isBlockedVoice(name string) bool {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return false
	}

	for _, keyword := range blockedVoiceKeywords {
		if strings.Contains(trimmed, keyword) {
			return true
		}
	}
	return false
}
