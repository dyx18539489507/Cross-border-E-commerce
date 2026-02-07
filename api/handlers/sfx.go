package handlers

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

type SFXHandler struct {
	storagePath string
	baseURL     string
	log         *logger.Logger
}

func NewSFXHandler(cfg *config.Config, log *logger.Logger) *SFXHandler {
	return &SFXHandler{
		storagePath: cfg.Storage.LocalPath,
		baseURL:     cfg.Storage.BaseURL,
		log:         log,
	}
}

type SFXItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Category string `json:"category"`
	Duration int    `json:"duration,omitempty"`
}

type GenerateSFXRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	Count  int    `json:"count"`
}

func (h *SFXHandler) ListSFX(c *gin.Context) {
	category := strings.TrimSpace(c.DefaultQuery("category", "热门"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	items := h.generateSFXItems(category, limit, false)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SFXHandler) GenerateSFX(c *gin.Context) {
	var req GenerateSFXRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prompt is required"})
		return
	}
	count := req.Count
	if count <= 0 {
		count = 3
	}
	if count > 5 {
		count = 5
	}
	items := h.generateSFXItems(prompt, count, true)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SFXHandler) generateSFXItems(prompt string, count int, aiMode bool) []SFXItem {
	items := make([]SFXItem, 0, count)
	category := prompt
	for i := 0; i < count; i++ {
		suffix := fmt.Sprintf("%02d", i+1)
		if aiMode {
			switch i {
			case 0:
				suffix = "低沉"
			case 1:
				suffix = "清脆"
			default:
				suffix = "回响"
			}
		}
		name := prompt
		if aiMode {
			name = fmt.Sprintf("%s-%s", prompt, suffix)
		} else {
			name = fmt.Sprintf("%s-%s", prompt, suffix)
		}
		fileURL, duration := h.ensureSFXFile(prompt, i, aiMode)
		items = append(items, SFXItem{
			ID:       h.buildSFXID(prompt, i, aiMode),
			Name:     name,
			URL:      fileURL,
			Category: category,
			Duration: duration,
		})
	}
	return items
}

func (h *SFXHandler) ensureSFXFile(prompt string, index int, aiMode bool) (string, int) {
	dir := filepath.Join(h.storagePath, "sfx")
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.log.Warnw("Failed to create sfx directory", "error", err)
	}

	fileName := h.buildSFXFileName(prompt, index, aiMode)
	filePath := filepath.Join(dir, fileName)
	duration := h.estimateDuration(index)

	if _, err := os.Stat(filePath); err == nil {
		return h.buildSFXURL(fileName), duration
	}

	samples, sampleRate := buildSFXSamples(prompt, index, aiMode)
	wav := buildWav(samples, sampleRate)
	if err := os.WriteFile(filePath, wav, 0644); err != nil {
		h.log.Warnw("Failed to write sfx file", "error", err, "path", filePath)
	}

	return h.buildSFXURL(fileName), duration
}

func (h *SFXHandler) estimateDuration(index int) int {
	base := 1200 + index*250
	return int(math.Round(float64(base) / 1000))
}

func (h *SFXHandler) buildSFXID(prompt string, index int, aiMode bool) string {
	seed := fmt.Sprintf("%s|%d|%t", prompt, index, aiMode)
	sum := sha1.Sum([]byte(seed))
	return fmt.Sprintf("sfx-%s-%d", hex.EncodeToString(sum[:6]), index)
}

func (h *SFXHandler) buildSFXFileName(prompt string, index int, aiMode bool) string {
	seed := fmt.Sprintf("%s|%d|%t", prompt, index, aiMode)
	sum := sha1.Sum([]byte(seed))
	return fmt.Sprintf("sfx_%s_%d.wav", hex.EncodeToString(sum[:8]), index)
}

func (h *SFXHandler) buildSFXURL(fileName string) string {
	base := strings.TrimRight(h.baseURL, "/")
	if base == "" {
		base = "/static"
	}
	return fmt.Sprintf("%s/sfx/%s", base, fileName)
}

func buildSFXSamples(prompt string, index int, aiMode bool) ([]int16, int) {
	sampleRate := 44100
	baseFreq, noise := profileForPrompt(prompt)
	seed := hashSeed(prompt, index, aiMode)
	rng := rand.New(rand.NewSource(seed))

	duration := 1.2 + float64(index)*0.3
	freq := baseFreq * (1 + float64(index)*0.18)
	fadeOut := true
	modFreq := 0.0
	if aiMode && index == 2 {
		modFreq = 2.0
	}
	if aiMode && index == 1 {
		noise *= 0.6
	}
	if aiMode && index == 0 {
		noise *= 1.2
		freq *= 0.7
	}

	n := int(duration * float64(sampleRate))
	samples := make([]int16, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(sampleRate)
		value := math.Sin(2 * math.Pi * freq * t)
		if modFreq > 0 {
			value *= 0.6 + 0.4*math.Sin(2*math.Pi*modFreq*t)
		}
		if noise > 0 {
			value += noise * (rng.Float64()*2 - 1)
		}
		if fadeOut {
			value *= 1 - float64(i)/float64(n)
		}
		if value > 1 {
			value = 1
		} else if value < -1 {
			value = -1
		}
		samples[i] = int16(value * 30000)
	}
	return samples, sampleRate
}

func profileForPrompt(prompt string) (float64, float64) {
	text := strings.ToLower(prompt)
	switch {
	case strings.Contains(text, "爆") || strings.Contains(text, "炸") || strings.Contains(text, "boom"):
		return 90, 0.8
	case strings.Contains(text, "风") || strings.Contains(text, "雨") || strings.Contains(text, "wave") || strings.Contains(text, "噪"):
		return 220, 0.9
	case strings.Contains(text, "笑") || strings.Contains(text, "哈哈") || strings.Contains(text, "giggle"):
		return 420, 0.4
	case strings.Contains(text, "铃") || strings.Contains(text, "滴") || strings.Contains(text, "beep"):
		return 880, 0.1
	default:
		return 330, 0.5
	}
}

func buildWav(samples []int16, sampleRate int) []byte {
	dataSize := len(samples) * 2
	fileSize := 36 + dataSize
	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(fileSize))
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("fmt "))
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1)
	binary.LittleEndian.PutUint16(buf[22:24], 1)
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(sampleRate*2))
	binary.LittleEndian.PutUint16(buf[32:34], 2)
	binary.LittleEndian.PutUint16(buf[34:36], 16)
	copy(buf[36:40], []byte("data"))
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))
	offset := 44
	for i, sample := range samples {
		binary.LittleEndian.PutUint16(buf[offset+i*2:], uint16(sample))
	}
	return buf
}

func hashSeed(prompt string, index int, aiMode bool) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(prompt))
	_, _ = h.Write([]byte(fmt.Sprintf("-%d-%t", index, aiMode)))
	return int64(h.Sum64())
}
