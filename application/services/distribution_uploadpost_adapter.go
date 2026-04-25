package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
)

const defaultUploadPostTimeout = 90 * time.Second

type DistributionAdapterError struct {
	StatusCode  int
	Message     string
	Body        string
	Retriable   bool
	NeedsRebind bool
}

func (e *DistributionAdapterError) Error() string {
	if e == nil {
		return ""
	}
	if e.Message != "" {
		return e.Message
	}
	if e.Body != "" {
		return e.Body
	}
	return "distribution adapter error"
}

type UploadPostAdapter struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type UploadPostGenerateLinkRequest struct {
	Username           string
	RedirectURL        string
	LogoImage          string
	RedirectButton     string
	ConnectTitle       string
	ConnectDescription string
}

type UploadPostGenerateLinkResponse struct {
	AccessURL string `json:"access_url"`
	Duration  string `json:"duration"`
}

type UploadPostProfileResponse struct {
	Username       string                 `json:"username"`
	CreatedAt      string                 `json:"created_at"`
	SocialAccounts map[string]interface{} `json:"social_accounts"`
}

type UploadPostBoard struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Privacy       string `json:"privacy"`
	URL           string `json:"url"`
	CoverImage    string `json:"image_cover_url"`
	PinCount      int    `json:"pin_count"`
	FollowerCount int    `json:"follower_count"`
}

type UploadPostPublishRequest struct {
	Username           string
	Platform           models.DistributionPlatform
	ContentType        models.DistributionContentType
	Title              string
	Body               string
	MediaURL           string
	MediaPath          string
	MediaFileName      string
	ScheduledAt        *time.Time
	AsyncUpload        bool
	RedditSubreddit    string
	RedditFlairID      string
	RedditFirstComment string
	PinterestBoardID   string
}

type UploadPostPublishResponse struct {
	StatusCode   int
	RequestID    string
	JobID        string
	ScheduledAt  *time.Time
	Raw          map[string]interface{}
	ResultByName map[string]UploadPostSyncPlatformResult
}

type UploadPostSyncPlatformResult struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	PostID  string `json:"post_id"`
	Error   string `json:"error"`
}

type UploadPostStatusResponse struct {
	RequestID  string                 `json:"request_id"`
	JobID      string                 `json:"job_id"`
	Status     string                 `json:"status"`
	Completed  int                    `json:"completed"`
	Total      int                    `json:"total"`
	LastUpdate *time.Time             `json:"last_update"`
	Results    []UploadPostStatusItem `json:"results"`
}

type UploadPostStatusItem struct {
	Platform        string     `json:"platform"`
	Success         bool       `json:"success"`
	Message         string     `json:"message"`
	UploadTimestamp *time.Time `json:"upload_timestamp"`
}

type UploadPostHistoryResponse struct {
	History []UploadPostHistoryItem `json:"history"`
	Total   int                     `json:"total"`
	Page    int                     `json:"page"`
	Limit   int                     `json:"limit"`
}

type UploadPostHistoryItem struct {
	Platform        string     `json:"platform"`
	MediaType       string     `json:"media_type"`
	Success         bool       `json:"success"`
	PostURL         string     `json:"post_url"`
	PlatformPostID  string     `json:"platform_post_id"`
	ErrorMessage    string     `json:"error_message"`
	RequestID       string     `json:"request_id"`
	JobID           string     `json:"job_id"`
	UploadTimestamp *time.Time `json:"upload_timestamp"`
	PostTitle       string     `json:"post_title"`
	PostCaption     string     `json:"post_caption"`
}

func NewUploadPostAdapter(cfg *config.Config) *UploadPostAdapter {
	baseURL := strings.TrimRight(cfg.Distribution.UploadPostBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.upload-post.com/api"
	}

	return &UploadPostAdapter{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(os.Getenv("UPLOAD_POST_API_KEY")),
		client: &http.Client{
			Timeout: defaultUploadPostTimeout,
		},
	}
}

func (a *UploadPostAdapter) IsConfigured() bool {
	return a != nil && a.apiKey != ""
}

func (a *UploadPostAdapter) EnsureUserProfile(ctx context.Context, username string) (*UploadPostProfileResponse, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	profile, err := a.GetUserProfile(ctx, username)
	if err == nil {
		return profile, nil
	}

	var adapterErr *DistributionAdapterError
	if !errors.As(err, &adapterErr) || adapterErr.StatusCode != http.StatusNotFound {
		return nil, err
	}

	payload := map[string]string{
		"username": username,
	}

	_, err = a.doJSON(ctx, http.MethodPost, "/uploadposts/users", payload, nil)
	if err != nil {
		if errors.As(err, &adapterErr) && adapterErr.StatusCode == http.StatusConflict {
			return a.GetUserProfile(ctx, username)
		}
		return nil, err
	}

	return a.GetUserProfile(ctx, username)
}

func (a *UploadPostAdapter) GetUserProfile(ctx context.Context, username string) (*UploadPostProfileResponse, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	body, err := a.doJSON(ctx, http.MethodGet, "/uploadposts/users/"+url.PathEscape(username), nil, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Profile UploadPostProfileResponse `json:"profile"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload-post profile response failed: %w", err)
	}

	return &resp.Profile, nil
}

func (a *UploadPostAdapter) GenerateConnectURL(ctx context.Context, req UploadPostGenerateLinkRequest) (*UploadPostGenerateLinkResponse, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"username":            req.Username,
		"connect_title":       req.ConnectTitle,
		"connect_description": req.ConnectDescription,
	}
	if req.RedirectURL != "" {
		payload["redirect_url"] = req.RedirectURL
	}
	if req.LogoImage != "" {
		payload["logo_image"] = req.LogoImage
	}
	if req.RedirectButton != "" {
		payload["redirect_button_text"] = req.RedirectButton
	}

	body, err := a.doJSON(ctx, http.MethodPost, "/uploadposts/users/generate-jwt", payload, nil)
	if err != nil {
		return nil, err
	}

	var resp UploadPostGenerateLinkResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload-post connect response failed: %w", err)
	}

	return &resp, nil
}

func (a *UploadPostAdapter) GetPinterestBoards(ctx context.Context, username string) ([]UploadPostBoard, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("profile", username)

	body, err := a.doJSON(ctx, http.MethodGet, "/uploadposts/pinterest/boards", nil, query)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Boards []UploadPostBoard `json:"boards"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload-post boards response failed: %w", err)
	}

	return resp.Boards, nil
}

func (a *UploadPostAdapter) Upload(ctx context.Context, req UploadPostPublishRequest) (*UploadPostPublishResponse, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	endpoint, err := a.publishEndpoint(req.ContentType)
	if err != nil {
		return nil, err
	}

	httpReq, cleanup, err := a.newUploadRequest(ctx, endpoint, req)
	if err != nil {
		return nil, err
	}
	if cleanup != nil {
		defer cleanup()
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, &DistributionAdapterError{
			Message:   fmt.Sprintf("upload-post request failed: %v", err),
			Retriable: true,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read upload-post response failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, &DistributionAdapterError{
			StatusCode:  resp.StatusCode,
			Message:     fmt.Sprintf("upload-post request failed: %s", resp.Status),
			Body:        string(body),
			Retriable:   resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError,
			NeedsRebind: resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
		}
	}

	result := &UploadPostPublishResponse{
		StatusCode: resp.StatusCode,
	}

	if err := json.Unmarshal(body, &result.Raw); err != nil {
		return nil, fmt.Errorf("unmarshal upload-post response failed: %w", err)
	}

	if requestID, _ := result.Raw["request_id"].(string); requestID != "" {
		result.RequestID = requestID
	}
	if jobID, _ := result.Raw["job_id"].(string); jobID != "" {
		result.JobID = jobID
	}
	if scheduledRaw, _ := result.Raw["scheduled_date"].(string); scheduledRaw != "" {
		if scheduledAt, parseErr := time.Parse(time.RFC3339, scheduledRaw); parseErr == nil {
			result.ScheduledAt = &scheduledAt
		}
	}

	if rawResults, ok := result.Raw["results"].(map[string]interface{}); ok {
		result.ResultByName = make(map[string]UploadPostSyncPlatformResult, len(rawResults))
		for platform, value := range rawResults {
			entryBytes, _ := json.Marshal(value)
			var item UploadPostSyncPlatformResult
			if err := json.Unmarshal(entryBytes, &item); err != nil {
				continue
			}
			result.ResultByName[platform] = item
		}
	}

	return result, nil
}

func (a *UploadPostAdapter) GetUploadStatus(ctx context.Context, requestID, jobID string) (*UploadPostStatusResponse, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	query := url.Values{}
	if requestID != "" {
		query.Set("request_id", requestID)
	}
	if jobID != "" {
		query.Set("job_id", jobID)
	}

	body, err := a.doJSON(ctx, http.MethodGet, "/uploadposts/status", nil, query)
	if err != nil {
		return nil, err
	}

	var resp UploadPostStatusResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload-post status response failed: %w", err)
	}

	return &resp, nil
}

func (a *UploadPostAdapter) GetUploadHistory(ctx context.Context, page, limit int) (*UploadPostHistoryResponse, error) {
	if err := a.requireConfigured(); err != nil {
		return nil, err
	}

	query := url.Values{}
	if page > 0 {
		query.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}

	body, err := a.doJSON(ctx, http.MethodGet, "/uploadposts/history", nil, query)
	if err != nil {
		return nil, err
	}

	var resp UploadPostHistoryResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal upload-post history response failed: %w", err)
	}

	return &resp, nil
}

func (a *UploadPostAdapter) publishEndpoint(contentType models.DistributionContentType) (string, error) {
	switch contentType {
	case models.DistributionContentTypeText:
		return "/upload_text", nil
	case models.DistributionContentTypeImage:
		return "/upload_photos", nil
	case models.DistributionContentTypeVideo:
		return "/upload", nil
	default:
		return "", fmt.Errorf("unsupported content type: %s", contentType)
	}
}

func (a *UploadPostAdapter) newUploadRequest(ctx context.Context, endpoint string, req UploadPostPublishRequest) (*http.Request, func(), error) {
	bodyReader, contentType, cleanup, err := a.buildUploadBody(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+endpoint, bodyReader)
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, nil, fmt.Errorf("create upload-post request failed: %w", err)
	}

	httpReq.Header.Set("Authorization", "Apikey "+a.apiKey)
	httpReq.Header.Set("Content-Type", contentType)

	return httpReq, cleanup, nil
}

func (a *UploadPostAdapter) buildUploadBody(ctx context.Context, req UploadPostPublishRequest) (io.Reader, string, func(), error) {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	var cleanupFns []func()
	cleanup := func() {
		for _, fn := range cleanupFns {
			fn()
		}
	}

	go func() {
		defer cleanup()
		defer pipeWriter.Close()
		defer writer.Close()

		writeErr := func(err error) {
			_ = pipeWriter.CloseWithError(err)
		}

		if err := writer.WriteField("user", req.Username); err != nil {
			writeErr(err)
			return
		}
		if err := writer.WriteField("platform[]", string(req.Platform)); err != nil {
			writeErr(err)
			return
		}
		if req.Title != "" {
			if err := writer.WriteField("title", req.Title); err != nil {
				writeErr(err)
				return
			}
		}
		if req.Body != "" {
			if err := writer.WriteField("description", req.Body); err != nil {
				writeErr(err)
				return
			}
		}
		if req.ScheduledAt != nil {
			if err := writer.WriteField("scheduled_date", req.ScheduledAt.UTC().Format(time.RFC3339)); err != nil {
				writeErr(err)
				return
			}
		} else if req.AsyncUpload {
			if err := writer.WriteField("async_upload", "true"); err != nil {
				writeErr(err)
				return
			}
		}
		if req.RedditSubreddit != "" {
			if err := writer.WriteField("subreddit", req.RedditSubreddit); err != nil {
				writeErr(err)
				return
			}
		}
		if req.RedditFlairID != "" {
			if err := writer.WriteField("flair_id", req.RedditFlairID); err != nil {
				writeErr(err)
				return
			}
		}
		if req.RedditFirstComment != "" {
			if err := writer.WriteField("reddit_first_comment", req.RedditFirstComment); err != nil {
				writeErr(err)
				return
			}
		}
		if req.PinterestBoardID != "" {
			if err := writer.WriteField("pinterest_board_id", req.PinterestBoardID); err != nil {
				writeErr(err)
				return
			}
		}

		switch req.ContentType {
		case models.DistributionContentTypeText:
			return
		case models.DistributionContentTypeImage:
			reader, fileName, closeFn, err := a.openMediaReader(ctx, req.MediaPath, req.MediaURL, req.MediaFileName)
			if err != nil {
				writeErr(err)
				return
			}
			if closeFn != nil {
				cleanupFns = append(cleanupFns, closeFn)
			}
			part, err := writer.CreateFormFile("photos[]", fileName)
			if err != nil {
				writeErr(err)
				return
			}
			if _, err := io.Copy(part, reader); err != nil {
				writeErr(err)
				return
			}
		case models.DistributionContentTypeVideo:
			if req.MediaPath == "" && req.MediaURL != "" && (strings.HasPrefix(req.MediaURL, "http://") || strings.HasPrefix(req.MediaURL, "https://")) {
				if err := writer.WriteField("video", req.MediaURL); err != nil {
					writeErr(err)
					return
				}
				return
			}

			reader, fileName, closeFn, err := a.openMediaReader(ctx, req.MediaPath, req.MediaURL, req.MediaFileName)
			if err != nil {
				writeErr(err)
				return
			}
			if closeFn != nil {
				cleanupFns = append(cleanupFns, closeFn)
			}
			part, err := writer.CreateFormFile("video", fileName)
			if err != nil {
				writeErr(err)
				return
			}
			if _, err := io.Copy(part, reader); err != nil {
				writeErr(err)
				return
			}
		default:
			writeErr(fmt.Errorf("unsupported content type: %s", req.ContentType))
		}
	}()

	return pipeReader, writer.FormDataContentType(), cleanup, nil
}

func (a *UploadPostAdapter) openMediaReader(ctx context.Context, mediaPath, mediaURL, fallbackName string) (io.Reader, string, func(), error) {
	if mediaPath != "" {
		file, err := os.Open(mediaPath)
		if err != nil {
			return nil, "", nil, fmt.Errorf("open media file failed: %w", err)
		}

		name := filepath.Base(mediaPath)
		if fallbackName != "" {
			name = fallbackName
		}

		return file, name, func() {
			_ = file.Close()
		}, nil
	}

	if mediaURL == "" {
		return nil, "", nil, errors.New("media is required")
	}

	resp, err := a.client.Get(mediaURL)
	if err != nil {
		return nil, "", nil, &DistributionAdapterError{
			Message:   fmt.Sprintf("download media failed: %v", err),
			Retriable: true,
		}
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer resp.Body.Close()
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, "", nil, &DistributionAdapterError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("download media failed: %s", resp.Status),
			Body:       string(body),
			Retriable:  resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError,
		}
	}

	name := fallbackName
	if name == "" {
		parsedURL, _ := url.Parse(mediaURL)
		name = filepath.Base(parsedURL.Path)
		if name == "" || name == "." || name == "/" {
			name = "upload.bin"
		}
	}

	return resp.Body, name, func() {
		_ = resp.Body.Close()
	}, nil
}

func (a *UploadPostAdapter) doJSON(ctx context.Context, method, path string, payload interface{}, query url.Values) ([]byte, error) {
	reqURL := a.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal request failed: %w", err)
		}
		body = strings.NewReader(string(raw))
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", "Apikey "+a.apiKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, &DistributionAdapterError{
			Message:   fmt.Sprintf("upload-post request failed: %v", err),
			Retriable: true,
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, &DistributionAdapterError{
			StatusCode:  resp.StatusCode,
			Message:     fmt.Sprintf("upload-post request failed: %s", resp.Status),
			Body:        string(bodyBytes),
			Retriable:   resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError,
			NeedsRebind: resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
		}
	}

	return bodyBytes, nil
}

func (a *UploadPostAdapter) requireConfigured() error {
	if a.IsConfigured() {
		return nil
	}
	return errors.New("UPLOAD_POST_API_KEY 未配置")
}
