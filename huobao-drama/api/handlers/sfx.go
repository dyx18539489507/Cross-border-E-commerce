package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

// 旧版本地静态音效（configs/sfx_gaudio.json）逻辑已停用。
// 当前实现改为聚合 Freesound + Pixabay 两个外部音效源。
type SfxHandler struct {
	cfg                *config.Config
	log                *logger.Logger
	httpClient         *http.Client
	sfxNameTranslation sync.Map
}

type SFXItem struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	URL         string  `json:"url"`
	Duration    int     `json:"duration,omitempty"`
	ViewCount   int     `json:"view_count,omitempty"`
	Artist      string  `json:"artist,omitempty"`
	Cover       string  `json:"cover,omitempty"`
	Description string  `json:"description,omitempty"`
	Source      string  `json:"source,omitempty"`
	Rank        int     `json:"rank,omitempty"`
	Heat        float64 `json:"-"`
}

type sfxSearchResult struct {
	source string
	items  []SFXItem
	err    error
}

func NewSfxHandler(cfg *config.Config, log *logger.Logger) *SfxHandler {
	timeoutSeconds := 30
	if cfg != nil && cfg.SFX.RequestTimeout > 0 {
		timeoutSeconds = cfg.SFX.RequestTimeout
	}

	return &SfxHandler{
		cfg: cfg,
		log: log,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

func (h *SfxHandler) List(c *gin.Context) {
	keywords := strings.TrimSpace(c.Query("keywords"))
	category := strings.TrimSpace(c.DefaultQuery("category", "热门"))
	if keywords == "" {
		keywords = mapSfxCategoryToQuery(category)
	}

	limit := parsePositiveInt(c.DefaultQuery("limit", ""), h.defaultLimit(), 60)
	page := parsePositiveInt(c.DefaultQuery("page", ""), 1, 200)

	items, warnings, err := h.searchSFX(c.Request.Context(), keywords, limit, page)
	if err != nil && len(items) == 0 {
		resp := gin.H{"error": "failed to load sfx"}
		if len(warnings) > 0 {
			resp["warnings"] = warnings
		}
		c.JSON(http.StatusBadGateway, resp)
		return
	}

	resp := gin.H{
		"items":    items,
		"total":    len(items),
		"page":     page,
		"limit":    limit,
		"has_more": len(items) >= limit,
	}
	if len(warnings) > 0 {
		resp["warnings"] = warnings
	}
	c.JSON(http.StatusOK, resp)
}

func mapSfxCategoryToQuery(category string) string {
	trimmed := strings.TrimSpace(category)
	normalized := strings.ToLower(trimmed)
	switch normalized {
	case "", "all", "热门", "hot":
		return "sound effect"
	case "转场", "transition":
		return "transition whoosh sweep swipe"
	case "笑声", "laugh", "laughter":
		return "laugh laughter crowd laugh"
	case "尴尬", "awkward":
		return "awkward record scratch fail buzzer"
	case "震惊", "shock", "surprise":
		return "dramatic hit shock stinger surprise"
	default:
		if trimmed == "" {
			return "sound effect"
		}
		return fmt.Sprintf("%s sound effect", trimmed)
	}
}

func (h *SfxHandler) defaultLimit() int {
	if h.cfg != nil && h.cfg.SFX.DefaultLimit > 0 {
		return h.cfg.SFX.DefaultLimit
	}
	return 20
}

func (h *SfxHandler) searchSFX(ctx context.Context, query string, limit int, page int) ([]SFXItem, []string, error) {
	if limit <= 0 {
		limit = h.defaultLimit()
	}
	if limit > 60 {
		limit = 60
	}
	if page <= 0 {
		page = 1
	}

	perSource := limit
	if perSource < 10 {
		perSource = 10
	}
	if perSource > 40 {
		perSource = 40
	}

	results := make(chan sfxSearchResult, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		items, err := h.fetchFreesound(ctx, query, perSource, page)
		results <- sfxSearchResult{source: "freesound", items: items, err: err}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		items, err := h.fetchPixabay(ctx, query, perSource, page)
		results <- sfxSearchResult{source: "pixabay", items: items, err: err}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	warnings := make([]string, 0, 2)
	merged := make([]SFXItem, 0, limit*2)
	for result := range results {
		if result.err != nil {
			warnings = append(warnings, fmt.Sprintf("%s: %v", result.source, result.err))
			if h.log != nil {
				h.log.Warnw("SFX source fetch failed", "source", result.source, "error", result.err)
			}
			continue
		}
		merged = append(merged, result.items...)
	}

	if len(merged) == 0 {
		if len(warnings) == 0 {
			return nil, warnings, errors.New("no sfx items available")
		}
		return nil, warnings, errors.New(strings.Join(warnings, "; "))
	}

	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].Heat == merged[j].Heat {
			return merged[i].Name < merged[j].Name
		}
		return merged[i].Heat > merged[j].Heat
	})

	unique := dedupeSFX(merged)
	if len(unique) > limit {
		unique = unique[:limit]
	}
	for i := range unique {
		unique[i].Rank = i + 1
	}
	if err := h.translateSFXNames(ctx, unique); err != nil {
		warnings = append(warnings, fmt.Sprintf("translation: %v", err))
		if h.log != nil {
			h.log.Warnw("SFX name translation failed", "error", err)
		}
	}

	return unique, warnings, nil
}

func dedupeSFX(items []SFXItem) []SFXItem {
	seen := make(map[string]struct{}, len(items))
	out := make([]SFXItem, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Source) + "|" + strings.TrimSpace(item.ID)
		if strings.TrimSpace(item.URL) != "" {
			key = strings.TrimSpace(item.URL)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, item)
	}
	return out
}

func (h *SfxHandler) fetchFreesound(ctx context.Context, query string, limit int, page int) ([]SFXItem, error) {
	apiKey := ""
	baseURL := "https://freesound.org/apiv2"
	if h.cfg != nil {
		apiKey = strings.TrimSpace(h.cfg.SFX.Freesound.APIKey)
		if configured := strings.TrimSpace(h.cfg.SFX.Freesound.BaseURL); configured != "" {
			baseURL = configured
		}
	}
	if apiKey == "" {
		return nil, errors.New("missing freesound api_key")
	}

	endpoint := strings.TrimRight(baseURL, "/") + "/search/text/"
	values := url.Values{}
	values.Set("query", strings.TrimSpace(query))
	values.Set("token", apiKey)
	values.Set("sort", "downloads_desc")
	values.Set("page_size", strconv.Itoa(limit))
	values.Set("page", strconv.Itoa(page))
	values.Set("fields", "id,name,description,tags,previews,images,duration,username,num_downloads,avg_rating,num_ratings")

	requestURL := endpoint + "?" + values.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "DramaGenerator/1.0")

	response, err := h.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("status %d: %s", response.StatusCode, truncateString(string(body), 160))
	}

	type freesoundResult struct {
		ID           int64    `json:"id"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Tags         []string `json:"tags"`
		Duration     float64  `json:"duration"`
		Username     string   `json:"username"`
		NumDownloads int      `json:"num_downloads"`
		AvgRating    float64  `json:"avg_rating"`
		NumRatings   int      `json:"num_ratings"`
		Previews     struct {
			PreviewHqMp3 string `json:"preview-hq-mp3"`
			PreviewLqMp3 string `json:"preview-lq-mp3"`
			PreviewHqOgg string `json:"preview-hq-ogg"`
			PreviewLqOgg string `json:"preview-lq-ogg"`
		} `json:"previews"`
		Images struct {
			WaveformM string `json:"waveform_m"`
			WaveformL string `json:"waveform_l"`
			SpectralM string `json:"spectral_m"`
			SpectralL string `json:"spectral_l"`
		} `json:"images"`
	}
	type freesoundResponse struct {
		Results []freesoundResult `json:"results"`
	}

	var parsed freesoundResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	items := make([]SFXItem, 0, len(parsed.Results))
	for _, result := range parsed.Results {
		previewURL := pickFirstNonEmpty(
			result.Previews.PreviewHqMp3,
			result.Previews.PreviewLqMp3,
			result.Previews.PreviewHqOgg,
			result.Previews.PreviewLqOgg,
		)
		if previewURL == "" {
			continue
		}
		name := strings.TrimSpace(result.Name)
		if name == "" {
			name = fmt.Sprintf("Freesound-%d", result.ID)
		}
		description := strings.TrimSpace(result.Description)
		if description == "" && len(result.Tags) > 0 {
			description = strings.Join(result.Tags, ", ")
		}
		duration := int(math.Round(result.Duration))
		if duration < 0 {
			duration = 0
		}
		heat := float64(result.NumDownloads) + result.AvgRating*300 + float64(result.NumRatings)*20

		items = append(items, SFXItem{
			ID:          fmt.Sprintf("freesound-%d", result.ID),
			Name:        name,
			Category:    "热门音效",
			URL:         previewURL,
			Duration:    duration,
			ViewCount:   result.NumDownloads,
			Artist:      strings.TrimSpace(result.Username),
			Cover:       pickFirstNonEmpty(result.Images.WaveformL, result.Images.WaveformM, result.Images.SpectralL, result.Images.SpectralM),
			Description: description,
			Source:      "freesound",
			Heat:        heat,
		})
	}

	return items, nil
}

func (h *SfxHandler) fetchPixabay(ctx context.Context, query string, limit int, page int) ([]SFXItem, error) {
	apiKey := ""
	baseURL := "https://pixabay.com"
	if h.cfg != nil {
		apiKey = strings.TrimSpace(h.cfg.SFX.Pixabay.APIKey)
		if configured := strings.TrimSpace(h.cfg.SFX.Pixabay.BaseURL); configured != "" {
			baseURL = configured
		}
	}
	if apiKey == "" {
		return nil, errors.New("missing pixabay api_key")
	}

	baseURL = strings.TrimRight(baseURL, "/")
	candidates := []string{
		baseURL + "/api/audio/",
		baseURL + "/api/sounds/",
	}

	var errs []string
	for _, endpoint := range candidates {
		items, err := h.fetchPixabayByEndpoint(ctx, endpoint, apiKey, query, limit, page)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s -> %v", endpoint, err))
			continue
		}
		return items, nil
	}

	if len(errs) == 0 {
		return nil, errors.New("pixabay returned no usable endpoint")
	}
	return nil, errors.New(strings.Join(errs, " | "))
}

func (h *SfxHandler) fetchPixabayByEndpoint(ctx context.Context, endpoint, apiKey, query string, limit int, page int) ([]SFXItem, error) {
	values := url.Values{}
	values.Set("key", apiKey)
	values.Set("per_page", strconv.Itoa(limit))
	values.Set("page", strconv.Itoa(page))
	values.Set("order", "popular")
	values.Set("safesearch", "true")
	if strings.TrimSpace(query) != "" {
		values.Set("q", strings.TrimSpace(query))
	} else {
		values.Set("q", "sound effect")
	}

	requestURL := endpoint + "?" + values.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "DramaGenerator/1.0")

	response, err := h.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("status %d: %s", response.StatusCode, truncateString(string(body), 160))
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	hitsRaw, ok := parsed["hits"]
	if !ok {
		return nil, errors.New("missing hits field")
	}
	hits, ok := hitsRaw.([]any)
	if !ok {
		return nil, errors.New("invalid hits field")
	}

	items := make([]SFXItem, 0, len(hits))
	for idx, rawHit := range hits {
		hit, ok := rawHit.(map[string]any)
		if !ok {
			continue
		}

		id := asString(hit["id"])
		if id == "" {
			id = fmt.Sprintf("pixabay-%d", idx)
		}

		name := strings.TrimSpace(asString(hit["name"]))
		if name == "" {
			name = firstTagOrFallback(asString(hit["tags"]), fmt.Sprintf("Pixabay-%s", id))
		}

		previewURL := ""
		if audioMap, ok := hit["audio"].(map[string]any); ok {
			previewURL = pickFirstNonEmpty(
				asString(audioMap["mp3"]),
				asString(audioMap["preview-hq-mp3"]),
				asString(audioMap["preview-lq-mp3"]),
				asString(audioMap["ogg"]),
				asString(audioMap["wav"]),
				asString(audioMap["url"]),
			)
		}
		if previewURL == "" {
			if previewsMap, ok := hit["previews"].(map[string]any); ok {
				previewURL = pickFirstNonEmpty(
					asString(previewsMap["preview-hq-mp3"]),
					asString(previewsMap["preview-lq-mp3"]),
					asString(previewsMap["mp3"]),
					asString(previewsMap["ogg"]),
				)
			}
		}
		if previewURL == "" {
			previewURL = pickFirstNonEmpty(
				asString(hit["previewURL"]),
				asString(hit["audio_url"]),
				asString(hit["url"]),
			)
		}
		if previewURL == "" {
			continue
		}

		duration := int(math.Round(asFloat(hit["duration"])))
		if duration < 0 {
			duration = 0
		}
		downloads := asInt(hit["downloads"])
		likes := asInt(hit["likes"])
		heat := float64(downloads) + float64(likes)*20

		description := firstTagOrFallback(asString(hit["tags"]), "Pixabay 音效")
		items = append(items, SFXItem{
			ID:          "pixabay-" + id,
			Name:        name,
			Category:    "热门音效",
			URL:         previewURL,
			Duration:    duration,
			ViewCount:   downloads,
			Artist:      asString(hit["user"]),
			Cover:       asString(hit["userImageURL"]),
			Description: description,
			Source:      "pixabay",
			Heat:        heat,
		})
	}

	return items, nil
}

func parsePositiveInt(raw string, fallback int, max int) int {
	value := fallback
	if strings.TrimSpace(raw) != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			value = parsed
		}
	}
	if value <= 0 {
		value = fallback
	}
	if max > 0 && value > max {
		return max
	}
	return value
}

func pickFirstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func truncateString(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= limit {
		return trimmed
	}
	return trimmed[:limit] + "..."
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func asFloat(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed
	default:
		return 0
	}
}

func asInt(value any) int {
	return int(math.Round(asFloat(value)))
}

func firstTagOrFallback(tags string, fallback string) string {
	trimmed := strings.TrimSpace(tags)
	if trimmed == "" {
		return fallback
	}
	parts := strings.Split(trimmed, ",")
	if len(parts) == 0 {
		return trimmed
	}
	candidate := strings.TrimSpace(parts[0])
	if candidate == "" {
		return trimmed
	}
	return candidate
}

func (h *SfxHandler) translateSFXNames(ctx context.Context, items []SFXItem) error {
	cfg := h.getTranslationConfig()
	if !cfg.enabled || cfg.appID == "" || cfg.apiKey == "" {
		return nil
	}
	if len(items) == 0 {
		return nil
	}

	queryToIndexes := make(map[string][]int)
	queryOrder := make([]string, 0, len(items))
	for i := range items {
		if strings.TrimSpace(items[i].Name) == "" {
			continue
		}
		if containsChineseText(items[i].Name) {
			continue
		}
		query := buildSFXTranslationQuery(items[i].Name)
		if query == "" {
			continue
		}
		if cached, ok := h.sfxNameTranslation.Load(query); ok {
			if translated, ok := cached.(string); ok && strings.TrimSpace(translated) != "" {
				items[i].Name = translated
			}
			continue
		}
		if _, exists := queryToIndexes[query]; !exists {
			queryOrder = append(queryOrder, query)
		}
		queryToIndexes[query] = append(queryToIndexes[query], i)
	}

	if len(queryOrder) == 0 {
		return nil
	}

	const batchSize = 25
	for start := 0; start < len(queryOrder); start += batchSize {
		end := start + batchSize
		if end > len(queryOrder) {
			end = len(queryOrder)
		}
		batchQueries := queryOrder[start:end]
		translations, err := h.translateBatchWithYoudao(ctx, cfg, batchQueries)
		if err != nil {
			return err
		}
		for i, source := range batchQueries {
			translated := ""
			if i < len(translations) {
				translated = strings.TrimSpace(translations[i])
			}
			if translated == "" {
				continue
			}
			h.sfxNameTranslation.Store(source, translated)
			for _, idx := range queryToIndexes[source] {
				items[idx].Name = translated
			}
		}
	}

	return nil
}

type sfxTranslationRuntimeConfig struct {
	enabled  bool
	appID    string
	apiKey   string
	endpoint string
	from     string
	to       string
}

func (h *SfxHandler) getTranslationConfig() sfxTranslationRuntimeConfig {
	cfg := sfxTranslationRuntimeConfig{
		endpoint: "https://openapi.youdao.com/api",
		from:     "auto",
		to:       "zh-CHS",
	}
	if h.cfg == nil {
		return cfg
	}
	source := h.cfg.SFX.Translation
	cfg.enabled = source.Enabled
	cfg.appID = strings.TrimSpace(source.AppID)
	cfg.apiKey = strings.TrimSpace(source.APIKey)
	if endpoint := strings.TrimSpace(source.Endpoint); endpoint != "" {
		cfg.endpoint = endpoint
	}
	if from := strings.TrimSpace(source.From); from != "" {
		cfg.from = from
	}
	if to := strings.TrimSpace(source.To); to != "" {
		cfg.to = to
	}
	return cfg
}

type youdaoTranslateResponse struct {
	ErrorCode   string   `json:"errorCode"`
	Translation []string `json:"translation"`
}

func (h *SfxHandler) translateBatchWithYoudao(ctx context.Context, cfg sfxTranslationRuntimeConfig, queries []string) ([]string, error) {
	if len(queries) == 0 {
		return nil, nil
	}

	rawQuery := strings.Join(queries, "\n")
	salt := strconv.FormatInt(time.Now().UnixNano(), 10)
	curtime := strconv.FormatInt(time.Now().Unix(), 10)
	signSource := cfg.appID + truncateForYoudaoSign(rawQuery) + salt + curtime + cfg.apiKey
	signBytes := sha256.Sum256([]byte(signSource))
	sign := hex.EncodeToString(signBytes[:])

	form := url.Values{}
	form.Set("q", rawQuery)
	form.Set("from", cfg.from)
	form.Set("to", cfg.to)
	form.Set("appKey", cfg.appID)
	form.Set("salt", salt)
	form.Set("sign", sign)
	form.Set("signType", "v3")
	form.Set("curtime", curtime)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "DramaGenerator/1.0")

	response, err := h.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("youdao status %d: %s", response.StatusCode, truncateString(string(body), 160))
	}

	var parsed youdaoTranslateResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	if parsed.ErrorCode != "" && parsed.ErrorCode != "0" {
		return nil, fmt.Errorf("youdao errorCode=%s", parsed.ErrorCode)
	}

	results := parsed.Translation
	if len(queries) > 1 && len(results) == 1 && strings.Contains(results[0], "\n") {
		results = strings.Split(results[0], "\n")
	}
	if len(results) < len(queries) {
		padded := make([]string, len(queries))
		copy(padded, results)
		results = padded
	}
	if len(results) > len(queries) {
		results = results[:len(queries)]
	}
	return results, nil
}

func truncateForYoudaoSign(text string) string {
	runes := []rune(text)
	length := len(runes)
	if length <= 20 {
		return text
	}
	return string(runes[:10]) + strconv.Itoa(length) + string(runes[length-10:])
}

func buildSFXTranslationQuery(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	for _, ext := range []string{".wav", ".mp3", ".ogg", ".flac", ".m4a", ".aac"} {
		if strings.HasSuffix(strings.ToLower(trimmed), ext) {
			trimmed = strings.TrimSpace(trimmed[:len(trimmed)-len(ext)])
			break
		}
	}
	replaced := strings.NewReplacer("_", " ", "-", " ").Replace(trimmed)
	return strings.Join(strings.Fields(replaced), " ")
}

func containsChineseText(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
