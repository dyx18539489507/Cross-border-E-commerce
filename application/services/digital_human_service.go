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
	visualActionProcess    = "CVProcess"

	reqKeySubjectRecognition = "jimeng_realman_avatar_picture_create_role_omni_v15"
	reqKeySubjectDetection   = "jimeng_realman_avatar_object_detection"
	reqKeyVideoGeneration    = "jimeng_realman_avatar_picture_omni_v15"
)

type DigitalHumanService struct {
	client *volcengine.VisualClient
	log    *logger.Logger
}

type DigitalHumanRequest struct {
	ImageURL   string
	AudioURL   string
	SpeechText string
	MotionText string
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
	Status        string `json:"status"`
	RespData      string `json:"resp_data"`
	VideoURL      string `json:"video_url"`
	AIGCMetaTagged *bool  `json:"aigc_meta_tagged"`
}

type subjectRecognitionResp struct {
	Status int `json:"status"`
}

type objectDetectionResp struct {
	Code   int `json:"code"`
	Status int `json:"status"`
	ObjectDetectionResult struct {
		Mask struct {
			URL []string `json:"url"`
		} `json:"mask"`
	} `json:"object_detection_result"`
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
	if req.ImageURL == "" || req.AudioURL == "" {
		return nil, fmt.Errorf("image_url and audio_url are required")
	}
	if s.client == nil || s.client.AccessKeyID == "" || s.client.SecretAccessKey == "" {
		return nil, fmt.Errorf("volcengine access key is not configured")
	}

	subjectDetected, err := s.checkSubject(ctx, req.ImageURL)
	if err != nil {
		return nil, err
	}
	if !subjectDetected {
		return nil, fmt.Errorf("image does not contain a valid subject")
	}

	maskURLs, err := s.detectSubjectMasks(ctx, req.ImageURL)
	if err != nil {
		s.log.Warnw("Subject mask detection failed, continue without mask", "error", err)
	}

	prompt := buildPrompt(req.SpeechText, req.MotionText)

	videoTaskID, err := s.submitVideoTask(ctx, req.ImageURL, req.AudioURL, maskURLs, prompt)
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
		MaskURLs:        maskURLs,
		SubjectDetected: subjectDetected,
	}, nil
}

func (s *DigitalHumanService) checkSubject(ctx context.Context, imageURL string) (bool, error) {
	taskID, err := s.submitTask(ctx, reqKeySubjectRecognition, map[string]any{
		"req_key":  reqKeySubjectRecognition,
		"image_url": imageURL,
	})
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	result, err := s.pollTask(ctx, reqKeySubjectRecognition, taskID, 3*time.Second)
	if err != nil {
		return false, err
	}

	if result.RespData == "" {
		return false, fmt.Errorf("subject recognition returned empty result")
	}

	var resp subjectRecognitionResp
	if err := json.Unmarshal([]byte(result.RespData), &resp); err != nil {
		return false, fmt.Errorf("parse subject recognition: %w", err)
	}

	return resp.Status == 1, nil
}

func (s *DigitalHumanService) detectSubjectMasks(ctx context.Context, imageURL string) ([]string, error) {
	resp, err := s.client.Do(ctx, visualActionProcess, volcengine.DefaultVisualVersion, map[string]any{
		"req_key":  reqKeySubjectDetection,
		"image_url": imageURL,
	})
	if err != nil {
		return nil, err
	}

	var data struct {
		RespData string `json:"resp_data"`
	}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("parse subject detection data: %w", err)
	}
	if data.RespData == "" {
		return nil, nil
	}

	var respData objectDetectionResp
	if err := json.Unmarshal([]byte(data.RespData), &respData); err != nil {
		return nil, fmt.Errorf("parse subject detection result: %w", err)
	}

	if respData.Status != 1 {
		return nil, nil
	}

	return respData.ObjectDetectionResult.Mask.URL, nil
}

func (s *DigitalHumanService) submitVideoTask(ctx context.Context, imageURL, audioURL string, maskURLs []string, prompt string) (string, error) {
	payload := map[string]any{
		"req_key":           reqKeyVideoGeneration,
		"image_url":         imageURL,
		"audio_url":         audioURL,
		"output_resolution": 1080,
		"pe_fast_mode":       false,
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

func buildPrompt(speechText, motionText string) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(speechText) != "" {
		parts = append(parts, fmt.Sprintf("说话内容：%s", strings.TrimSpace(speechText)))
	}
	if strings.TrimSpace(motionText) != "" {
		parts = append(parts, fmt.Sprintf("动作描述：%s", strings.TrimSpace(motionText)))
	}
	return strings.Join(parts, "；")
}
