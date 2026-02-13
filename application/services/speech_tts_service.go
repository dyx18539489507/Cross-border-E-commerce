package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
)

const (
	defaultVolcengineTTSEndpoint       = "https://openspeech.bytedance.com/api/v1/tts"
	defaultVolcengineTTSCluster        = "volcano_tts"
	defaultVolcengineTTSSubmitEndpoint = "https://openspeech.bytedance.com/api/v3/tts/submit"
	defaultVolcengineTTSQueryEndpoint  = "https://openspeech.bytedance.com/api/v3/tts/query"
	defaultVolcengineTTSResourceID     = "volc.service_type.10029"
	defaultVolcengineTTSNamespace      = "BidirectionalTTS"
)

type SpeechTTSService struct {
	client           *http.Client
	log              *logger.Logger
	endpoint         string
	appID            string
	token            string
	cluster          string
	submitEndpoint   string
	queryEndpoint    string
	resourceID       string
	namespace        string
	defaultVoiceType string
}

type TTSSynthesizeOptions struct {
	ResourceID string
}

type volcengineTTSV3SubmitRequest struct {
	User struct {
		UID string `json:"uid"`
	} `json:"user"`
	Namespace string `json:"namespace"`
	ReqParams struct {
		Text    string `json:"text"`
		Speaker string `json:"speaker"`
		Audio   struct {
			Format     string `json:"format"`
			SampleRate int    `json:"sample_rate"`
		} `json:"audio_params"`
	} `json:"req_params"`
}

type volcengineTTSV3SubmitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID          string `json:"task_id"`
		TaskStatus      int    `json:"task_status"`
		ReqTextLength   int    `json:"req_text_length"`
		SynthesisLength int    `json:"synthesize_text_length"`
	} `json:"data"`
}

type volcengineTTSV3QueryRequest struct {
	TaskID string `json:"task_id"`
}

type volcengineTTSV3QueryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TaskID            string `json:"task_id"`
		TaskStatus        int    `json:"task_status"`
		AudioURL          string `json:"audio_url"`
		ReqTextLength     int    `json:"req_text_length"`
		SynthesizeLength  int    `json:"synthesize_text_length"`
		FailureReason     string `json:"failure_reason"`
		FailureDetailCode string `json:"failure_code"`
	} `json:"data"`
}

func NewSpeechTTSService(cfg *config.Config, log *logger.Logger) *SpeechTTSService {
	speechCfg := cfg.Volcengine.Speech

	endpoint := strings.TrimSpace(speechCfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultVolcengineTTSEndpoint
	}

	cluster := strings.TrimSpace(speechCfg.Cluster)
	if cluster == "" {
		cluster = defaultVolcengineTTSCluster
	}

	submitEndpoint := strings.TrimSpace(speechCfg.SubmitEndpoint)
	if submitEndpoint == "" {
		submitEndpoint = defaultVolcengineTTSSubmitEndpoint
	}

	queryEndpoint := strings.TrimSpace(speechCfg.QueryEndpoint)
	if queryEndpoint == "" {
		queryEndpoint = defaultVolcengineTTSQueryEndpoint
	}

	resourceID := strings.TrimSpace(speechCfg.ResourceID)
	if resourceID == "" {
		resourceID = defaultVolcengineTTSResourceID
	}

	namespace := strings.TrimSpace(speechCfg.Namespace)
	if namespace == "" {
		namespace = defaultVolcengineTTSNamespace
	}

	return &SpeechTTSService{
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
		log:              log,
		endpoint:         endpoint,
		appID:            strings.TrimSpace(speechCfg.AppID),
		token:            strings.TrimSpace(speechCfg.Token),
		cluster:          cluster,
		submitEndpoint:   submitEndpoint,
		queryEndpoint:    queryEndpoint,
		resourceID:       resourceID,
		namespace:        namespace,
		defaultVoiceType: strings.TrimSpace(speechCfg.VoiceType),
	}
}

func (s *SpeechTTSService) IsConfigured() bool {
	return s != nil && s.appID != "" && s.token != ""
}

func (s *SpeechTTSService) Synthesize(ctx context.Context, text, voiceType string) ([]byte, string, error) {
	audioURL, err := s.SynthesizeToURL(ctx, text, voiceType)
	if err != nil {
		return nil, "", err
	}
	return s.downloadAudio(ctx, audioURL)
}

func (s *SpeechTTSService) SynthesizeToURL(ctx context.Context, text, voiceType string) (string, error) {
	return s.SynthesizeToURLWithOptions(ctx, text, voiceType, nil)
}

func (s *SpeechTTSService) SynthesizeToURLWithResource(ctx context.Context, text, voiceType, resourceID string) (string, error) {
	return s.SynthesizeToURLWithOptions(ctx, text, voiceType, &TTSSynthesizeOptions{ResourceID: resourceID})
}

func (s *SpeechTTSService) SynthesizeToURLWithOptions(ctx context.Context, text, voiceType string, options *TTSSynthesizeOptions) (string, error) {
	if !s.IsConfigured() {
		return "", fmt.Errorf("volcengine speech tts is not configured")
	}

	trimmedText := strings.TrimSpace(text)
	if trimmedText == "" {
		return "", fmt.Errorf("text is empty")
	}

	resolvedVoiceType := strings.TrimSpace(voiceType)
	if resolvedVoiceType == "" {
		resolvedVoiceType = s.defaultVoiceType
	}
	if resolvedVoiceType == "" {
		return "", fmt.Errorf("voice_type is required")
	}

	resolvedResourceID := strings.TrimSpace(s.resourceID)
	if options != nil {
		if override := strings.TrimSpace(options.ResourceID); override != "" {
			resolvedResourceID = override
		}
	}
	if resolvedResourceID == "" {
		return "", fmt.Errorf("resource_id is required")
	}

	taskID, err := s.submitV3Task(ctx, trimmedText, resolvedVoiceType, resolvedResourceID)
	if err != nil {
		return "", err
	}

	return s.pollV3TaskAudioURL(ctx, taskID, resolvedResourceID)
}

func (s *SpeechTTSService) submitV3Task(ctx context.Context, text, voiceType, resourceID string) (string, error) {
	payload := volcengineTTSV3SubmitRequest{}
	payload.User.UID = fmt.Sprintf("uid-%d", time.Now().UnixNano())
	payload.Namespace = s.namespace
	payload.ReqParams.Text = text
	payload.ReqParams.Speaker = voiceType
	payload.ReqParams.Audio.Format = "mp3"
	payload.ReqParams.Audio.SampleRate = 24000

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal tts submit payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.submitEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create tts submit request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-App-Id", s.appID)
	httpReq.Header.Set("X-Api-Access-Key", s.token)
	httpReq.Header.Set("X-Api-Resource-Id", resourceID)
	httpReq.Header.Set("X-Api-Request-Id", newTTSRequestID())

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("send tts submit request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read tts submit response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("tts submit http status %d: %s", resp.StatusCode, truncateTTSString(string(respBody), 240))
	}

	var parsed volcengineTTSV3SubmitResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("parse tts submit response: %w", err)
	}

	if parsed.Code != 20000000 {
		if parsed.Message == "" {
			parsed.Message = "unknown tts submit error"
		}
		return "", fmt.Errorf("tts submit failed (code=%d): %s", parsed.Code, parsed.Message)
	}

	taskID := strings.TrimSpace(parsed.Data.TaskID)
	if taskID == "" {
		return "", fmt.Errorf("tts submit task_id is empty")
	}

	return taskID, nil
}

func (s *SpeechTTSService) pollV3TaskAudioURL(ctx context.Context, taskID, resourceID string) (string, error) {
	pollCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	ticker := time.NewTicker(1200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-pollCtx.Done():
			return "", fmt.Errorf("tts query timeout: %w", pollCtx.Err())
		case <-ticker.C:
			result, err := s.queryV3Task(pollCtx, taskID, resourceID)
			if err != nil {
				return "", err
			}

			audioURL := strings.TrimSpace(result.Data.AudioURL)
			if audioURL != "" {
				return audioURL, nil
			}

			status := result.Data.TaskStatus
			if status < 0 || status >= 4 {
				reason := strings.TrimSpace(result.Data.FailureReason)
				if reason == "" {
					reason = strings.TrimSpace(result.Message)
				}
				if reason == "" {
					reason = fmt.Sprintf("task_status=%d", status)
				}
				return "", fmt.Errorf("tts query failed: %s", reason)
			}
		}
	}
}

func (s *SpeechTTSService) queryV3Task(ctx context.Context, taskID, resourceID string) (*volcengineTTSV3QueryResponse, error) {
	payload := volcengineTTSV3QueryRequest{TaskID: taskID}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal tts query payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.queryEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create tts query request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-App-Id", s.appID)
	httpReq.Header.Set("X-Api-Access-Key", s.token)
	httpReq.Header.Set("X-Api-Resource-Id", resourceID)
	httpReq.Header.Set("X-Api-Request-Id", newTTSRequestID())

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send tts query request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tts query response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("tts query http status %d: %s", resp.StatusCode, truncateTTSString(string(respBody), 240))
	}

	var parsed volcengineTTSV3QueryResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("parse tts query response: %w", err)
	}

	if parsed.Code != 20000000 {
		if parsed.Message == "" {
			parsed.Message = "unknown tts query error"
		}
		return nil, fmt.Errorf("tts query failed (code=%d): %s", parsed.Code, parsed.Message)
	}

	return &parsed, nil
}

func (s *SpeechTTSService) downloadAudio(ctx context.Context, audioURL string) ([]byte, string, error) {
	trimmedURL := strings.TrimSpace(audioURL)
	if trimmedURL == "" {
		return nil, "", fmt.Errorf("tts audio url is empty")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create tts audio download request: %w", err)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, "", fmt.Errorf("download tts audio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 240))
		return nil, "", fmt.Errorf("tts audio download http status %d: %s", resp.StatusCode, truncateTTSString(string(body), 240))
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read tts audio bytes: %w", err)
	}

	if len(audioBytes) == 0 {
		return nil, "", fmt.Errorf("downloaded tts audio is empty")
	}

	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	return audioBytes, contentType, nil
}

func newTTSRequestID() string {
	return fmt.Sprintf("tts-%d", time.Now().UnixNano())
}

func truncateTTSString(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max] + "..."
}
