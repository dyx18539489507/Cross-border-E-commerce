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

	reqKeySubjectRecognition = "jimeng_realman_avatar_picture_create_role_omni_v15"
	reqKeyVideoGeneration    = "jimeng_realman_avatar_picture_omni_v15"
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
	TaskID          string `json:"task_id"`
	VideoURL        string `json:"video_url"`
	SubjectDetected bool   `json:"subject_detected"`
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
	subjectDetected, err := s.recognizeSubject(ctx, req.ImageURL)
	if err != nil {
		s.log.Warnw("Subject recognition fallback to direct generation", "error", err, "image_url", req.ImageURL)
		subjectDetected = false
	} else if !subjectDetected {
		s.log.Warnw("Subject recognition returned no subject, fallback to direct generation", "image_url", req.ImageURL)
	}

	videoTaskID, err := s.submitVideoTask(ctx, req.ImageURL, req.ImageBase64, audioURL, voiceType, speechText, prompt)
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
		SubjectDetected: subjectDetected,
	}, nil
}

func (s *DigitalHumanService) submitVideoTask(ctx context.Context, imageURL, imageBase64, audioURL, voiceType, speechText string, prompt string) (string, error) {
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
	if strings.TrimSpace(imageURL) != "" {
		payload["image_url"] = imageURL
	} else if strings.TrimSpace(imageBase64) != "" {
		payload["image_base64"] = strings.TrimSpace(imageBase64)
	}
	if strings.TrimSpace(prompt) != "" {
		payload["prompt"] = prompt
	}

	taskID, err := s.submitTask(ctx, reqKeyVideoGeneration, payload)
	if err == nil {
		return taskID, nil
	}
	if strings.TrimSpace(prompt) == "" || !shouldRetryDigitalHumanWithoutPrompt(err) {
		return "", err
	}

	s.log.Warnw("Retrying digital human generation without prompt", "error", err)
	delete(payload, "prompt")
	return s.submitTask(ctx, reqKeyVideoGeneration, payload)
}

func (s *DigitalHumanService) waitForVideo(ctx context.Context, taskID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Minute)
	defer cancel()

	result, err := s.pollTask(ctx, reqKeyVideoGeneration, taskID, 5*time.Second)
	if err != nil {
		return "", err
	}

	if result.VideoURL == "" {
		return "", fmt.Errorf("video url is empty")
	}

	return result.VideoURL, nil
}

func (s *DigitalHumanService) submitTask(ctx context.Context, reqKey string, payload map[string]any) (string, error) {
	resp, err := s.doVisualRequest(ctx, visualActionSubmitTask, payload)
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

	return data.TaskID, nil
}

func (s *DigitalHumanService) pollTask(ctx context.Context, reqKey, taskID string, interval time.Duration) (*getResultData, error) {
	for {
		result, err := s.getTaskResult(ctx, reqKey, taskID)
		if err != nil {
			return nil, err
		}

		switch result.Status {
		case "done":
			return result, nil
		case "expired", "not_found":
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
	resp, err := s.doVisualRequest(ctx, visualActionGetResult, map[string]any{
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

func (s *DigitalHumanService) doVisualRequest(ctx context.Context, action string, payload map[string]any) (*volcengine.VisualResponse, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := s.client.Do(ctx, action, volcengine.DefaultVisualVersion, payload)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryableDigitalHumanError(err) || attempt == 2 {
			return nil, err
		}

		wait := time.Duration(attempt+1) * 2 * time.Second
		s.log.Warnw("Retrying digital human visual request", "action", action, "attempt", attempt+2, "error", err, "wait", wait.String())

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}

	return nil, lastErr
}

func isRetryableDigitalHumanError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "504 Gateway Time-out") ||
		strings.Contains(msg, "volcengine http status 504") ||
		strings.Contains(msg, "volcengine http status 502")
}

func shouldRetryDigitalHumanWithoutPrompt(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	return strings.Contains(msg, "get prompt json fail") ||
		strings.Contains(msg, "algoProxy exceed max retry times") ||
		strings.Contains(msg, "\"code\":50501") ||
		strings.Contains(msg, "status\":50501")
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

func (s *DigitalHumanService) recognizeSubject(ctx context.Context, imageURL string) (bool, error) {
	if strings.TrimSpace(imageURL) == "" {
		return false, fmt.Errorf("image url is required for subject recognition")
	}

	taskID, err := s.submitTask(ctx, reqKeySubjectRecognition, map[string]any{
		"req_key":   reqKeySubjectRecognition,
		"image_url": strings.TrimSpace(imageURL),
	})
	if err != nil {
		return false, fmt.Errorf("subject recognition submit failed: %w", err)
	}

	result, err := s.pollTask(ctx, reqKeySubjectRecognition, taskID, 3*time.Second)
	if err != nil {
		return false, fmt.Errorf("subject recognition failed: %w", err)
	}

	return resolveSubjectRecognition(result), nil
}

func resolveSubjectRecognition(result *getResultData) bool {
	if result == nil {
		return false
	}

	respData := parseDigitalHumanRespData(strings.TrimSpace(result.RespData))
	if respData == nil {
		return false
	}

	status := digitalHumanAnyToInt(respData["status"])
	return status == 1
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
		text, ok := parsed.(string)
		if !ok {
			break
		}
		text = strings.TrimSpace(text)
		if text == "" || text == "null" {
			return nil
		}
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			return nil
		}
	}

	data, _ := parsed.(map[string]any)
	return data
}

func digitalHumanAnyToString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func digitalHumanAnyToInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		if typed == "" {
			return 0
		}
		var parsed int
		_, _ = fmt.Sscanf(typed, "%d", &parsed)
		return parsed
	default:
		return 0
	}
}
