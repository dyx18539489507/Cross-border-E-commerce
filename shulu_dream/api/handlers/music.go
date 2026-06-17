package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/logger"
	"github.com/gin-gonic/gin"
)

type MusicHandler struct {
	log *logger.Logger
}

func NewMusicHandler(log *logger.Logger) *MusicHandler {
	return &MusicHandler{log: log}
}

type MusicSearchItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Duration  string `json:"duration"`
	Source    string `json:"source"`
	SongURL   string `json:"song_url"`
	Mid       string `json:"mid"`
	Hash      string `json:"hash"`
	ContentID string `json:"content_id"`
}

type MusicSearchResponse struct {
	Items []MusicSearchItem `json:"items"`
	Total int               `json:"total"`
}

func (h *MusicHandler) SearchNetease(c *gin.Context) {
	keywords := c.Query("keywords")
	if keywords == "" {
		keywords = c.Query("s")
	}
	if keywords == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keywords is required"})
		return
	}

	limit := c.DefaultQuery("limit", "30")
	searchURL := fmt.Sprintf("https://music.163.com/api/search/get?type=1&s=%s&limit=%s", url.QueryEscape(keywords), url.QueryEscape(limit))

	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	req.Header.Set("Referer", "https://music.163.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.log.Warnw("Netease search request failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "netease search failed"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "netease search read failed"})
		return
	}

	c.Data(resp.StatusCode, "application/json", body)
}

func (h *MusicHandler) SearchAll(c *gin.Context) {
	keywords := c.Query("keywords")
	if keywords == "" {
		keywords = c.Query("s")
	}
	if keywords == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keywords is required"})
		return
	}
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	type searchResult struct {
		source string
		items  []MusicSearchItem
		total  int
		err    error
	}

	resultsCh := make(chan searchResult, 2)
	go func() {
		items, total, err := h.searchNeteaseItems(keywords, page, pageSize)
		resultsCh <- searchResult{source: "netease", items: items, total: total, err: err}
	}()
	go func() {
		items, total, err := h.searchMusicDLItems(keywords, page, pageSize)
		resultsCh <- searchResult{source: "music_dl", items: items, total: total, err: err}
	}()

	var neteaseRes searchResult
	var musicDLRes searchResult
	for i := 0; i < 2; i++ {
		res := <-resultsCh
		if res.source == "netease" {
			neteaseRes = res
		} else {
			musicDLRes = res
		}
	}

	if neteaseRes.err != nil {
		h.log.Warnw("Netease search failed", "error", neteaseRes.err)
	}
	if musicDLRes.err != nil {
		h.log.Warnw("music-dl search failed", "error", musicDLRes.err)
	}

	items := make([]MusicSearchItem, 0, len(neteaseRes.items)+len(musicDLRes.items))
	seen := make(map[string]struct{})
	appendUnique := func(list []MusicSearchItem) {
		for _, item := range list {
			key := strings.ToLower(strings.Join([]string{item.Source, item.ID, item.Mid, item.Hash, item.Title, item.Artist}, "|"))
			if key == "" {
				key = strings.ToLower(strings.Join([]string{item.Source, item.Title, item.Artist}, "|"))
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, item)
		}
	}
	appendUnique(neteaseRes.items)
	appendUnique(musicDLRes.items)

	total := neteaseRes.total + musicDLRes.total
	if total < len(items) {
		total = len(items)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}

func (h *MusicHandler) searchNeteaseItems(keyword, page, pageSize string) ([]MusicSearchItem, int, error) {
	p, _ := strconv.Atoi(page)
	if p <= 0 {
		p = 1
	}
	ps, _ := strconv.Atoi(pageSize)
	if ps <= 0 {
		ps = 20
	}
	offset := (p - 1) * ps

	searchURL := fmt.Sprintf("https://music.163.com/api/search/get?type=1&s=%s&limit=%d&offset=%d", url.QueryEscape(keyword), ps, offset)
	req, err := http.NewRequest(http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Referer", "https://music.163.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	type neteaseResp struct {
		Result struct {
			SongCount int `json:"songCount"`
			Songs     []struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
				DT   int64  `json:"dt"`
				AR   []struct {
					Name string `json:"name"`
				} `json:"ar"`
				AL struct {
					Name string `json:"name"`
				} `json:"al"`
			} `json:"songs"`
		} `json:"result"`
	}

	var parsed neteaseResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, 0, err
	}
	items := make([]MusicSearchItem, 0, len(parsed.Result.Songs))
	for _, song := range parsed.Result.Songs {
		artists := make([]string, 0, len(song.AR))
		for _, a := range song.AR {
			if a.Name != "" {
				artists = append(artists, a.Name)
			}
		}
		items = append(items, MusicSearchItem{
			ID:       fmt.Sprintf("%d", song.ID),
			Title:    song.Name,
			Artist:   strings.Join(artists, "/"),
			Album:    song.AL.Name,
			Duration: fmt.Sprintf("%d", song.DT/1000),
			Source:   "netease",
		})
	}
	return items, parsed.Result.SongCount, nil
}

func (h *MusicHandler) searchMusicDLItems(keyword, page, pageSize string) ([]MusicSearchItem, int, error) {
	return h.searchMusicDLItemsWithSources(keyword, page, pageSize, "qq,kugou,migu,baidu", 0)
}

func (h *MusicHandler) searchMusicDLItemsWithSources(keyword, page, pageSize, sources string, resolveLimit int) ([]MusicSearchItem, int, error) {
	cwd, _ := os.Getwd()
	script := filepath.Join(cwd, "scripts", "music_dl_search.py")
	pythonPath := getPythonPath(cwd)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonPath, script, keyword, page, pageSize, sources, strconv.Itoa(resolveLimit))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, 0, fmt.Errorf("music-dl search timeout: %w", err)
		}
		return nil, 0, fmt.Errorf("music-dl search error: %w: %s", err, out.String())
	}
	var resp MusicSearchResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return nil, 0, err
	}
	return resp.Items, resp.Total, nil
}

func (h *MusicHandler) GetNeteaseSongURL(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	outerURL := fmt.Sprintf("https://music.163.com/song/media/outer/url?id=%s.mp3", url.QueryEscape(id))

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, outerURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	req.Header.Set("Referer", "https://music.163.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		h.log.Warnw("Netease song url request failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "netease song url failed"})
		return
	}
	defer resp.Body.Close()

	resolved := resp.Header.Get("Location")
	if resolved == "" {
		resolved = outerURL
	}

	c.JSON(http.StatusOK, gin.H{
		"data": []gin.H{{"url": resolved}},
	})
}

func (h *MusicHandler) StreamNeteaseSong(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	// Try player url first (better availability)
	playerURL := fmt.Sprintf("https://music.163.com/api/song/enhance/player/url?ids=[%s]&br=128000", url.QueryEscape(id))
	req, err := http.NewRequest(http.MethodGet, playerURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}
	req.Header.Set("Referer", "https://music.163.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			type playerData struct {
				Data []struct {
					URL string `json:"url"`
				} `json:"data"`
			}
			var parsed playerData
			_ = json.Unmarshal(body, &parsed)
			if len(parsed.Data) > 0 && parsed.Data[0].URL != "" {
				h.proxyStream(c, parsed.Data[0].URL)
				return
			}
		}
	}

	// Fallback to outer url redirect
	outerURL := fmt.Sprintf("https://music.163.com/song/media/outer/url?id=%s.mp3", url.QueryEscape(id))
	if !h.proxyStream(c, outerURL) {
		writeSilentWav(c)
	}
}

func (h *MusicHandler) StreamMusic(c *gin.Context) {
	rawURL := c.Query("url")
	source := strings.ToLower(c.Query("source"))
	if rawURL != "" {
		if !isAllowedAudioURL(rawURL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
			return
		}
		if !h.proxyStream(c, rawURL) {
			writeSilentWav(c)
		}
		return
	}

	payload := map[string]string{
		"id":         c.Query("id"),
		"mid":        c.Query("mid"),
		"hash":       c.Query("hash"),
		"content_id": c.Query("content_id"),
		"title":      c.Query("title"),
		"artist":     c.Query("artist"),
	}
	data, _ := json.Marshal(payload)
	cwd, _ := os.Getwd()
	script := filepath.Join(cwd, "scripts", "music_dl_resolve.py")
	pythonPath := getPythonPath(cwd)
	cmd := exec.Command(pythonPath, script, source, string(data))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		h.log.Warnw("music-dl resolve failed", "error", err, "output", out.String())
		writeSilentWav(c)
		return
	}
	var res struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(out.Bytes(), &res); err != nil || res.URL == "" {
		keyword := strings.TrimSpace(payload["title"])
		if keyword == "" {
			keyword = strings.TrimSpace(strings.TrimSpace(payload["title"]) + " " + strings.TrimSpace(payload["artist"]))
		}
		if keyword != "" {
			if items, _, err := h.searchMusicDLItemsWithSources(keyword, "1", "20", "qq,kugou,migu,baidu,netease", 8); err == nil {
				for _, item := range items {
					if item.SongURL == "" {
						continue
					}
					if h.proxyStream(c, item.SongURL) {
						return
					}
				}
			}
		}
		writeSilentWav(c)
		return
	}
	if !h.proxyStream(c, res.URL) {
		writeSilentWav(c)
	}
}

func isAllowedAudioURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Host)
	allowed := []string{"qq.com", "kugou.com", "migu.cn", "music.163.com", "music.126.net", "stream.qqmusic.qq.com", "dl.stream.qqmusic.qq.com"}
	for _, a := range allowed {
		if strings.Contains(host, a) {
			return true
		}
	}
	return false
}

func getPythonPath(base string) string {
	venvPython := filepath.Join(base, ".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err == nil {
		return venvPython
	}
	return "python3"
}

func (h *MusicHandler) proxyStream(c *gin.Context, target string) bool {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return false
	}
	referer := ""
	switch {
	case strings.Contains(target, "music.163.com") || strings.Contains(target, "music.126.net"):
		referer = "https://music.163.com"
	case strings.Contains(target, "qqmusic") || strings.Contains(target, "y.qq.com") || strings.Contains(target, "qq.com"):
		referer = "https://y.qq.com/"
	case strings.Contains(target, "kugou.com"):
		referer = "https://www.kugou.com/"
	case strings.Contains(target, "migu.cn"):
		referer = "https://music.migu.cn/"
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if rangeHeader := c.GetHeader("Range"); rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp == nil {
		h.log.Warnw("Netease stream request failed", "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		h.log.Warnw("audio stream upstream status not ok", "status", resp.StatusCode, "target", target)
		return false
	}
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/plain") || strings.Contains(contentType, "text/html") {
		h.log.Warnw("audio stream upstream returned non-audio content", "content_type", contentType, "target", target)
		return false
	}

	// Pass through status and headers
	for key, values := range resp.Header {
		for _, v := range values {
			c.Writer.Header().Add(key, v)
		}
	}
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
	return true
}

func writeSilentWav(c *gin.Context) {
	// 1 second of silence, 16-bit mono, 44100Hz
	sampleRate := 44100
	seconds := 1
	numSamples := sampleRate * seconds
	dataSize := numSamples * 2
	fileSize := 36 + dataSize

	buf := make([]byte, 44+dataSize)
	copy(buf[0:4], []byte("RIFF"))
	putLE32(buf[4:8], uint32(fileSize))
	copy(buf[8:12], []byte("WAVE"))
	copy(buf[12:16], []byte("fmt "))
	putLE32(buf[16:20], 16)
	putLE16(buf[20:22], 1)
	putLE16(buf[22:24], 1)
	putLE32(buf[24:28], uint32(sampleRate))
	putLE32(buf[28:32], uint32(sampleRate*2))
	putLE16(buf[32:34], 2)
	putLE16(buf[34:36], 16)
	copy(buf[36:40], []byte("data"))
	putLE32(buf[40:44], uint32(dataSize))

	c.Header("Content-Type", "audio/wav")
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write(buf)
}

func putLE16(buf []byte, v uint16) {
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
}

func putLE32(buf []byte, v uint32) {
	buf[0] = byte(v)
	buf[1] = byte(v >> 8)
	buf[2] = byte(v >> 16)
	buf[3] = byte(v >> 24)
}
