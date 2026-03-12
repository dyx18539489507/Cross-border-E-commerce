package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/volcengine"
)

const (
	visualActionSubmitTask = "CVSubmitTask"
	visualActionGetResult  = "CVGetResult"

	reqKeyVideoGeneration = "jimeng_realman_avatar_picture_omni_v15"
)

type DigitalHumanService struct {
	client *volcengine.VisualClient
	log    *logger.Logger
}

type DigitalHumanRequest struct {
	ImageURL    string
	ImageBase64 string
	AudioURL    string
	VoiceType   string
	SpeechText  string
	MotionText  string
}

type DigitalHumanResult struct {
	TaskID          string   `json:"task_id"`
	VideoURL        string   `json:"video_url"`
	MaskURLs        []string `json:"mask_urls,omitempty"`
	SubjectDetected bool     `json:"subject_detected"`
}

type submitTaskData struct {
	TaskID string `json:"task_id"`
}

type getResultData struct {
	Status         string `json:"status"`
	RespData       string `json:"resp_data"`
	VideoURL       string `json:"video_url"`
	AIGCMetaTagged *bool  `json:"aigc_meta_tagged"`
}

func NewDigitalHumanService(cfg *config.Config, log *logger.Logger) *DigitalHumanService {
	accessKeyID := strings.TrimSpace(cfg.Volcengine.AccessKeyID)
	secretAccessKey := strings.TrimSpace(cfg.Volcengine.SecretAccessKey)
	if accessKeyID == "" {
		accessKeyID = strings.TrimSpace(os.Getenv("VOLCENGINE_ACCESS_KEY_ID"))
	}
	if secretAccessKey == "" {
		secretAccessKey = strings.TrimSpace(os.Getenv("VOLCENGINE_SECRET_ACCESS_KEY"))
	}

	client := volcengine.NewVisualClient(
		accessKeyID,
		secretAccessKey,
		cfg.Volcengine.Region,
		cfg.Volcengine.Service,
		cfg.Volcengine.VisualHost,
	)

	return &DigitalHumanService{
		client: client,
		log:    log,
	}
}

func (s *DigitalHumanService) Generate(ctx context.Context, req *DigitalHumanRequest) (*DigitalHumanResult, error) {
	if req.ImageURL == "" && strings.TrimSpace(req.ImageBase64) == "" {
		return nil, fmt.Errorf("image is required")
	}

	audioURL := strings.TrimSpace(req.AudioURL)
	speechText := strings.TrimSpace(req.SpeechText)
	voiceType := strings.TrimSpace(req.VoiceType)
	if audioURL == "" && speechText == "" {
		return nil, fmt.Errorf("audio_url or speech_text is required")
	}
	if audioURL == "" && speechText != "" && voiceType == "" {
		return nil, fmt.Errorf("voice_type is required when speech_text is provided")
	}
	if s.client == nil || s.client.AccessKeyID == "" || s.client.SecretAccessKey == "" {
		return nil, fmt.Errorf("volcengine access key is not configured")
	}

	prompt := buildPrompt(req.MotionText)

	videoTaskID, err := s.submitVideoTask(ctx, req.ImageURL, req.ImageBase64, audioURL, voiceType, speechText, nil, prompt)
	if err != nil {
		return nil, err
	}

	videoURL, err := s.waitForVideo(ctx, videoTaskID)
	if err != nil {
		return nil, err
	}

	return &DigitalHumanResult{
		TaskID:          videoTaskID,
		VideoURL:        videoURL,
		SubjectDetected: true,
	}, nil
}

func (s *DigitalHumanService) submitVideoTask(ctx context.Context, imageURL, imageBase64, audioURL, voiceType, speechText string, maskURLs []string, prompt string) (string, error) {
	payload := map[string]any{
		"req_key":           reqKeyVideoGeneration,
		"output_resolution": 1080,
		"pe_fast_mode":      false,
	}
	if strings.TrimSpace(audioURL) != "" {
		payload["audio_url"] = strings.TrimSpace(audioURL)
	}
	if strings.TrimSpace(voiceType) != "" {
		payload["voice_type"] = strings.TrimSpace(voiceType)
	}
	if strings.TrimSpace(speechText) != "" {
		payload["speech_text"] = strings.TrimSpace(speechText)
	}
	if strings.TrimSpace(imageBase64) != "" {
		payload["image_base64"] = strings.TrimSpace(imageBase64)
	} else {
		payload["image_url"] = imageURL
	}
	if len(maskURLs) > 0 {
		payload["mask_url"] = maskURLs
	}
	if strings.TrimSpace(prompt) != "" {
		payload["prompt"] = prompt
	}

	return s.submitTask(ctx, reqKeyVideoGeneration, payload)
}

func (s *DigitalHumanService) waitForVideo(ctx context.Context, taskID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Minute)
	defer cancel()

	result, err := s.pollTask(ctx, reqKeyVideoGeneration, taskID, 5*time.Second)
	if err != nil {
		return "", err
	}

	videoURL := resolveDigitalHumanVideoURL(result)
	if videoURL == "" {
		if failure := describeDigitalHumanTaskFailure(result); failure != "" {
			return "", fmt.Errorf("video url is empty: %s", failure)
		}
		return "", fmt.Errorf("video url is empty")
	}

	return videoURL, nil
}

func (s *DigitalHumanService) submitTask(ctx context.Context, reqKey string, payload map[string]any) (string, error) {
	resp, err := s.client.Do(ctx, visualActionSubmitTask, volcengine.DefaultVisualVersion, payload)
	if err != nil {
		return "", err
	}

	var data submitTaskData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", fmt.Errorf("parse submit task data: %w", err)
	}

	if data.TaskID == "" {
		return "", fmt.Errorf("task_id is empty")
	}

	s.log.Infow("Submitted digital human task", "task_id", data.TaskID, "req_key", reqKey)
	return data.TaskID, nil
}

func (s *DigitalHumanService) pollTask(ctx context.Context, reqKey, taskID string, interval time.Duration) (*getResultData, error) {
	for {
		result, err := s.getTaskResult(ctx, reqKey, taskID)
		if err != nil {
			return nil, err
		}

		status := strings.ToLower(strings.TrimSpace(result.Status))
		switch status {
		case "done", "success", "succeeded", "completed":
			return result, nil
		case "expired", "not_found":
			return nil, fmt.Errorf("task status %s", result.Status)
		case "fail", "failed", "error", "rejected", "canceled", "cancelled", "aborted":
			if failure := describeDigitalHumanTaskFailure(result); failure != "" {
				return nil, fmt.Errorf("task status %s: %s", result.Status, failure)
			}
			return nil, fmt.Errorf("task status %s", result.Status)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func (s *DigitalHumanService) getTaskResult(ctx context.Context, reqKey, taskID string) (*getResultData, error) {
	resp, err := s.client.Do(ctx, visualActionGetResult, volcengine.DefaultVisualVersion, map[string]any{
		"req_key": reqKey,
		"task_id": taskID,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		return nil, fmt.Errorf("empty task result")
	}

	var data getResultData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse task result: %w", err)
	}

	return &data, nil
}

func buildPrompt(motionText string) string {
	parts := []string{
		"严格以上传角色为唯一主角，保持同一张脸、五官、发型、肤色和服饰，不替换人物，不改变性别年龄，不新增第二人物",
		"优先保持参考图原始景别与构图，自然说话并保持口型同步，避免生成与参考图不一致的新人物",
	}

	if strings.TrimSpace(motionText) != "" {
		parts = append(parts, fmt.Sprintf("在人物身份不变的前提下执行以下动作：%s", strings.TrimSpace(motionText)))
	} else {
		parts = append(parts, "仅做自然说话、轻微点头或手势动作，避免大幅改造人物造型")
	}

	return strings.Join(parts, "；")
}

func resolveDigitalHumanVideoURL(result *getResultData) string {
	if result == nil {
		return ""
	}

	if videoURL := strings.TrimSpace(result.VideoURL); videoURL != "" {
		return videoURL
	}

	respData := parseDigitalHumanRespData(strings.TrimSpace(result.RespData))
	if respData == nil {
		return ""
	}

	return firstDigitalHumanURL(respData, "video_url", "videoUrl", "url", "video_urls", "videos", "data", "result")
}

func describeDigitalHumanTaskFailure(result *getResultData) string {
	if result == nil {
		return ""
	}

	raw := strings.TrimSpace(result.RespData)
	if raw == "" || raw == "null" {
		return ""
	}

	respData := parseDigitalHumanRespData(raw)
	if respData == nil {
		return raw
	}

	message := firstDigitalHumanString(respData, "message", "msg", "detail", "reason")
	nestedError := firstDigitalHumanMap(respData, "error")
	if message == "" {
		if nestedError != nil {
			message = firstDigitalHumanString(nestedError, "message", "msg", "detail", "reason")
		}
	}

	code := strings.TrimSpace(digitalHumanAnyToString(respData["code"]))
	if code == "" && nestedError != nil {
		code = strings.TrimSpace(digitalHumanAnyToString(nestedError["code"]))
	}
	switch {
	case code != "" && message != "":
		return fmt.Sprintf("code=%s: %s", code, message)
	case code != "":
		return fmt.Sprintf("code=%s", code)
	case message != "":
		return message
	default:
		return raw
	}
}

func parseDigitalHumanRespData(raw string) map[string]any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "null" {
		return nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil
	}

	for i := 0; i < 2; i++ {
		if text, ok := parsed.(string); ok {
			text = strings.TrimSpace(text)
			if text == "" || text == "null" {
				return nil
			}
			if err := json.Unmarshal([]byte(text), &parsed); err != nil {
				return nil
			}
			continue
		}
		break
	}

	data, _ := parsed.(map[string]any)
	return data
}

func firstDigitalHumanURL(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if nested := firstDigitalHumanMap(data, key); nested != nil {
			if value := firstDigitalHumanURL(nested, keys...); value != "" {
				return value
			}
			continue
		}
		if list := firstDigitalHumanList(data, key); len(list) > 0 {
			for _, item := range list {
				switch typed := item.(type) {
				case string:
					if value := strings.TrimSpace(typed); value != "" {
						return value
					}
				case map[string]any:
					if value := firstDigitalHumanURL(typed, keys...); value != "" {
						return value
					}
				}
			}
			continue
		}
		switch typed := data[key].(type) {
		case string:
			if value := strings.TrimSpace(typed); value != "" {
				return value
			}
		}
	}
	return ""
}

func firstDigitalHumanString(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(digitalHumanAnyToString(data[key])); value != "" {
			return value
		}
	}
	return ""
}

func firstDigitalHumanMap(data map[string]any, key string) map[string]any {
	value, ok := data[key]
	if !ok {
		return nil
	}
	nested, _ := value.(map[string]any)
	return nested
}

func firstDigitalHumanList(data map[string]any, key string) []any {
	value, ok := data[key]
	if !ok {
		return nil
	}
	list, _ := value.([]any)
	return list
}

func digitalHumanAnyToString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprintf("%v", typed)
	}
}
