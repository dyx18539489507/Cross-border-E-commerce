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

	if result.VideoURL == "" {
		return "", fmt.Errorf("video url is empty")
	}

	return result.VideoURL, nil
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
	parts := make([]string, 0, 1)
	if strings.TrimSpace(motionText) != "" {
		parts = append(parts, fmt.Sprintf("动作描述：%s", strings.TrimSpace(motionText)))
	}
	return strings.Join(parts, "；")
}
