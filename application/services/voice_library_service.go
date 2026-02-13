package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/volcengine"
	"gorm.io/gorm"
)

const (
	defaultVoiceCloneUploadEndpoint = "https://openspeech.bytedance.com/api/v1/mega_tts/audio/upload"
	defaultVoiceCloneResourceID     = "volc.megatts.voiceclone"
	defaultVoiceCloneProjectName    = "default"

	voiceCloneStatusActionVersion = "2025-05-21"
)

type VoiceSpeaker struct {
	ID             string   `json:"id"`
	VoiceType      string   `json:"voice_type"`
	Name           string   `json:"name"`
	Avatar         string   `json:"avatar"`
	Gender         string   `json:"gender"`
	Age            string   `json:"age"`
	TrialURL       string   `json:"trial_url"`
	Categories     []string `json:"categories,omitempty"`
	IsCustom       bool     `json:"is_custom,omitempty"`
	Status         string   `json:"status,omitempty"`
	ResourceID     string   `json:"resource_id,omitempty"`
	SourceAudioURL string   `json:"source_audio_url,omitempty"`
	LastError      string   `json:"last_error,omitempty"`
}

type VoiceLibraryService struct {
	db                 *gorm.DB
	client             *volcengine.OpenAPIClient
	version            string
	cloneStatusVersion string
	cloneUploadURL     string
	cloneResourceID    string
	cloneProjectName   string
	speechAppID        string
	speechToken        string
	httpClient         *http.Client
	log                *logger.Logger
}

type CreateCustomVoiceRequest struct {
	Name           string
	SourceAudioURL string
	AudioBytes     []byte
	AudioFormat    string
}

func NewVoiceLibraryService(cfg *config.Config, db *gorm.DB, log *logger.Logger) (*VoiceLibraryService, error) {
	if cfg.Volcengine.AccessKeyID == "" || cfg.Volcengine.SecretAccessKey == "" {
		return nil, fmt.Errorf("missing volcengine access key config")
	}

	region := cfg.Volcengine.Region
	if region == "" {
		region = volcengine.DefaultSpeechSaasRegion
	}

	client := volcengine.NewOpenAPIClient(
		cfg.Volcengine.AccessKeyID,
		cfg.Volcengine.SecretAccessKey,
		region,
		volcengine.DefaultSpeechSaasService,
		volcengine.DefaultOpenAPIHost,
	)

	cloneUploadURL := strings.TrimSpace(cfg.Volcengine.Speech.CloneUploadEndpoint)
	if cloneUploadURL == "" {
		cloneUploadURL = defaultVoiceCloneUploadEndpoint
	}

	cloneResourceID := strings.TrimSpace(cfg.Volcengine.Speech.CloneResourceID)
	if cloneResourceID == "" {
		cloneResourceID = defaultVoiceCloneResourceID
	}

	cloneProjectName := strings.TrimSpace(cfg.Volcengine.Speech.CloneProjectName)
	if cloneProjectName == "" {
		cloneProjectName = defaultVoiceCloneProjectName
	}

	return &VoiceLibraryService{
		db:                 db,
		client:             client,
		version:            volcengine.DefaultSpeechSaasVersion,
		cloneStatusVersion: voiceCloneStatusActionVersion,
		cloneUploadURL:     cloneUploadURL,
		cloneResourceID:    cloneResourceID,
		cloneProjectName:   cloneProjectName,
		speechAppID:        strings.TrimSpace(cfg.Volcengine.Speech.AppID),
		speechToken:        strings.TrimSpace(cfg.Volcengine.Speech.Token),
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
		log: log,
	}, nil
}

type listBigModelTTSTimbresResult struct {
	Timbres []struct {
		SpeakerID   string `json:"SpeakerID"`
		TimbreInfos []struct {
			SpeakerName string `json:"SpeakerName"`
			Gender      string `json:"Gender"`
			Age         string `json:"Age"`
			Categories  []struct {
				Category     string `json:"Category"`
				NextCategory *struct {
					Category string `json:"Category"`
				} `json:"NextCategory"`
			} `json:"Categories"`
			Emotions []struct {
				DemoURL string `json:"DemoURL"`
			} `json:"Emotions"`
		} `json:"TimbreInfos"`
	} `json:"Timbres"`
}

type listSpeakersResult struct {
	Speakers []struct {
		ID         string `json:"ID"`
		VoiceType  string `json:"VoiceType"`
		Name       string `json:"Name"`
		Avatar     string `json:"Avatar"`
		Gender     string `json:"Gender"`
		Age        string `json:"Age"`
		TrialURL   string `json:"TrialURL"`
		Categories []struct {
			Categories []string `json:"Categories"`
		} `json:"Categories"`
	} `json:"Speakers"`
}

type listMegaTTSTrainStatusResult struct {
	Statuses []megaTrainSpeakerStatus `json:"Statuses"`
}

type batchListMegaTTSTrainStatusResult struct {
	Statuses   []megaTrainSpeakerStatus `json:"Statuses"`
	PageNumber int                      `json:"PageNumber"`
	PageSize   int                      `json:"PageSize"`
	TotalCount int                      `json:"TotalCount"`
	NextToken  string                   `json:"NextToken"`
}

type megaTrainSpeakerStatus struct {
	SpeakerID              string `json:"SpeakerID"`
	InstanceNO             string `json:"InstanceNO"`
	InstanceStatus         string `json:"InstanceStatus"`
	IsActivable            bool   `json:"IsActivable"`
	State                  string `json:"State"`
	DemoAudio              string `json:"DemoAudio"`
	Version                string `json:"Version"`
	CreateTime             int64  `json:"CreateTime"`
	ExpireTime             int64  `json:"ExpireTime"`
	OrderTime              int64  `json:"OrderTime"`
	Alias                  string `json:"Alias"`
	AvailableTrainingTimes int    `json:"AvailableTrainingTimes"`
	ResourceID             string `json:"ResourceID"`
}

func (s *VoiceLibraryService) ListSpeakers(ctx context.Context) ([]VoiceSpeaker, error) {
	customVoices, customErr := s.listCustomVoices(ctx)
	if customErr != nil {
		s.log.Warnw("Failed to load custom voices", "error", customErr)
	}

	cloudVoices, err := s.listCloudVoices(ctx)
	if err != nil {
		s.log.Warnw("Failed to load cloud voices, fallback to custom voices only", "error", err)
		if len(customVoices) > 0 {
			return customVoices, nil
		}
		return []VoiceSpeaker{}, nil
	}

	if len(customVoices) == 0 {
		return cloudVoices, nil
	}

	merged := make([]VoiceSpeaker, 0, len(customVoices)+len(cloudVoices))
	merged = append(merged, customVoices...)
	merged = append(merged, cloudVoices...)
	return merged, nil
}

func (s *VoiceLibraryService) listCloudVoices(ctx context.Context) ([]VoiceSpeaker, error) {
	timbres, err := s.listBigModelTimbres(ctx)
	if err == nil && len(timbres) > 0 {
		return timbres, nil
	}
	if err != nil {
		s.log.Warnw("ListBigModelTTSTimbres failed, fallback to ListSpeakers", "error", err)
	}

	return s.listLegacySpeakers(ctx)
}

func (s *VoiceLibraryService) listBigModelTimbres(ctx context.Context) ([]VoiceSpeaker, error) {
	resultRaw, err := s.client.Do(ctx, "ListBigModelTTSTimbres", s.version, map[string]any{})
	if err != nil {
		return nil, err
	}

	var parsed listBigModelTTSTimbresResult
	if err := json.Unmarshal(resultRaw, &parsed); err != nil {
		return nil, fmt.Errorf("parse big model timbre list: %w", err)
	}

	speakers := make([]VoiceSpeaker, 0, len(parsed.Timbres))
	for _, item := range parsed.Timbres {
		speakerID := strings.TrimSpace(item.SpeakerID)
		if speakerID == "" {
			continue
		}

		name := ""
		gender := ""
		age := ""
		trialURL := ""
		categorySet := make(map[string]struct{})

		for _, info := range item.TimbreInfos {
			if name == "" {
				name = strings.TrimSpace(info.SpeakerName)
			}
			if gender == "" {
				gender = strings.TrimSpace(info.Gender)
			}
			if age == "" {
				age = strings.TrimSpace(info.Age)
			}

			for _, emotion := range info.Emotions {
				if trialURL == "" {
					trialURL = strings.TrimSpace(emotion.DemoURL)
				}
			}

			for _, cat := range info.Categories {
				base := strings.TrimSpace(cat.Category)
				if base == "" {
					continue
				}
				categorySet[base] = struct{}{}
				if cat.NextCategory != nil {
					next := strings.TrimSpace(cat.NextCategory.Category)
					if next != "" {
						categorySet[base+"/"+next] = struct{}{}
					}
				}
			}
		}

		if name == "" {
			name = speakerID
		}

		categories := make([]string, 0, len(categorySet))
		for cat := range categorySet {
			categories = append(categories, cat)
		}

		speakers = append(speakers, VoiceSpeaker{
			ID:         speakerID,
			VoiceType:  speakerID,
			Name:       name,
			Avatar:     "",
			Gender:     gender,
			Age:        age,
			TrialURL:   trialURL,
			Categories: categories,
		})
	}

	return speakers, nil
}

func (s *VoiceLibraryService) listLegacySpeakers(ctx context.Context) ([]VoiceSpeaker, error) {
	resultRaw, err := s.client.Do(ctx, "ListSpeakers", s.version, map[string]any{
		"Offset": 0,
		"Limit":  1000,
	})
	if err != nil {
		return nil, err
	}

	var parsed listSpeakersResult
	if err := json.Unmarshal(resultRaw, &parsed); err != nil {
		return nil, fmt.Errorf("parse voice list: %w", err)
	}

	speakers := make([]VoiceSpeaker, 0, len(parsed.Speakers))
	for _, item := range parsed.Speakers {
		categories := make([]string, 0)
		for _, cat := range item.Categories {
			categories = append(categories, cat.Categories...)
		}

		speakers = append(speakers, VoiceSpeaker{
			ID:         item.ID,
			VoiceType:  item.VoiceType,
			Name:       item.Name,
			Avatar:     item.Avatar,
			Gender:     item.Gender,
			Age:        item.Age,
			TrialURL:   item.TrialURL,
			Categories: categories,
		})
	}

	return speakers, nil
}

func (s *VoiceLibraryService) CreateCustomVoice(ctx context.Context, req *CreateCustomVoiceRequest) (*VoiceSpeaker, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}
	if strings.TrimSpace(s.speechAppID) == "" || strings.TrimSpace(s.speechToken) == "" {
		return nil, fmt.Errorf("missing speech app_id/token config")
	}
	if len(req.AudioBytes) == 0 {
		return nil, fmt.Errorf("audio is empty")
	}

	audioFormat := normalizeAudioFormat(req.AudioFormat)
	speaker, err := s.pickTrainableSpeaker(ctx)
	if err != nil {
		return nil, err
	}

	if err := s.uploadVoiceCloneAudio(ctx, speaker.SpeakerID, req.AudioBytes, audioFormat); err != nil {
		return nil, err
	}

	status, statusErr := s.getSpeakerStatus(ctx, speaker.SpeakerID)
	if statusErr != nil {
		s.log.Warnw("Failed to query custom voice status after upload", "speaker_id", speaker.SpeakerID, "error", statusErr)
		status = speaker
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "自定义音色"
	}

	resourceID := strings.TrimSpace(status.ResourceID)
	if resourceID == "" {
		resourceID = strings.TrimSpace(speaker.ResourceID)
	}

	trialURL := strings.TrimSpace(status.DemoAudio)
	if trialURL == "" {
		trialURL = strings.TrimSpace(req.SourceAudioURL)
	}

	record := &models.CustomVoice{
		Name:           name,
		Provider:       "volcengine",
		SpeakerID:      speaker.SpeakerID,
		VoiceType:      speaker.SpeakerID,
		ResourceID:     resourceID,
		SourceAudioURL: strings.TrimSpace(req.SourceAudioURL),
		TrialURL:       trialURL,
		Status:         mapVoiceCloneState(status.State),
	}

	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("save custom voice: %w", err)
	}

	speakerResult := s.customVoiceToSpeaker(record)
	return &speakerResult, nil
}

func (s *VoiceLibraryService) RefreshCustomVoiceStatus(ctx context.Context, voiceID uint) (*VoiceSpeaker, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database is not configured")
	}

	var record models.CustomVoice
	if err := s.db.WithContext(ctx).First(&record, voiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("custom voice not found")
		}
		return nil, err
	}

	status, err := s.getSpeakerStatus(ctx, record.SpeakerID)
	if err != nil {
		return nil, err
	}

	record.Status = mapVoiceCloneState(status.State)
	if status.ResourceID != "" {
		record.ResourceID = status.ResourceID
	}
	if status.DemoAudio != "" {
		record.TrialURL = status.DemoAudio
	}
	if status.SpeakerID != "" {
		record.VoiceType = status.SpeakerID
	}

	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, err
	}

	result := s.customVoiceToSpeaker(&record)
	return &result, nil
}

func (s *VoiceLibraryService) ResolveCustomVoiceByPublicID(ctx context.Context, publicID string) (*models.CustomVoice, bool, error) {
	id, isCustom := parseCustomVoicePublicID(publicID)
	if !isCustom {
		return nil, false, nil
	}
	if s.db == nil {
		return nil, true, fmt.Errorf("database is not configured")
	}

	var voice models.CustomVoice
	if err := s.db.WithContext(ctx).First(&voice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, true, fmt.Errorf("custom voice not found")
		}
		return nil, true, err
	}
	return &voice, true, nil
}

func (s *VoiceLibraryService) listCustomVoices(ctx context.Context) ([]VoiceSpeaker, error) {
	if s.db == nil {
		return nil, nil
	}

	var records []models.CustomVoice
	if err := s.db.WithContext(ctx).Order("updated_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	result := make([]VoiceSpeaker, 0, len(records))
	for index := range records {
		item := s.customVoiceToSpeaker(&records[index])
		result = append(result, item)
	}
	return result, nil
}

func (s *VoiceLibraryService) customVoiceToSpeaker(item *models.CustomVoice) VoiceSpeaker {
	lastErr := ""
	if item.LastError != nil {
		lastErr = strings.TrimSpace(*item.LastError)
	}

	trialURL := strings.TrimSpace(item.TrialURL)
	if trialURL == "" {
		trialURL = strings.TrimSpace(item.SourceAudioURL)
	}

	return VoiceSpeaker{
		ID:             item.PublicID(),
		VoiceType:      strings.TrimSpace(item.VoiceType),
		Name:           strings.TrimSpace(item.Name),
		TrialURL:       trialURL,
		Categories:     []string{"自定义音色"},
		IsCustom:       true,
		Status:         strings.TrimSpace(item.Status),
		ResourceID:     strings.TrimSpace(item.ResourceID),
		SourceAudioURL: strings.TrimSpace(item.SourceAudioURL),
		LastError:      lastErr,
	}
}

func (s *VoiceLibraryService) pickTrainableSpeaker(ctx context.Context) (*megaTrainSpeakerStatus, error) {
	resultRaw, err := s.client.Do(ctx, "BatchListMegaTTSTrainStatus", s.cloneStatusVersion, map[string]any{
		"ProjectName": s.cloneProjectName,
		"PageNumber":  1,
		"PageSize":    100,
	})
	if err != nil {
		return nil, fmt.Errorf("query trainable speaker list: %w", err)
	}

	var parsed batchListMegaTTSTrainStatusResult
	if err := json.Unmarshal(resultRaw, &parsed); err != nil {
		return nil, fmt.Errorf("parse trainable speaker list: %w", err)
	}

	if len(parsed.Statuses) == 0 {
		return nil, fmt.Errorf("no available speaker slot")
	}

	type scoredSpeaker struct {
		score   int
		speaker megaTrainSpeakerStatus
	}

	scored := make([]scoredSpeaker, 0, len(parsed.Statuses))
	for _, item := range parsed.Statuses {
		speakerID := strings.TrimSpace(item.SpeakerID)
		if speakerID == "" {
			continue
		}
		if item.AvailableTrainingTimes <= 0 {
			continue
		}
		score := item.AvailableTrainingTimes * 10
		if strings.EqualFold(strings.TrimSpace(item.InstanceStatus), "active") {
			score += 100
		}
		stateLower := strings.ToLower(strings.TrimSpace(item.State))
		if strings.Contains(stateLower, "training") || strings.Contains(stateLower, "queue") || strings.Contains(stateLower, "process") {
			score -= 50
		}
		scored = append(scored, scoredSpeaker{score: score, speaker: item})
	}

	if len(scored) == 0 {
		return nil, fmt.Errorf("no speaker slot with available training times")
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	picked := scored[0].speaker
	return &picked, nil
}

func (s *VoiceLibraryService) getSpeakerStatus(ctx context.Context, speakerID string) (*megaTrainSpeakerStatus, error) {
	resultRaw, err := s.client.Do(ctx, "ListMegaTTSTrainStatus", s.cloneStatusVersion, map[string]any{
		"ProjectName": s.cloneProjectName,
		"SpeakerIDs":  []string{speakerID},
	})
	if err != nil {
		return nil, fmt.Errorf("query speaker status: %w", err)
	}

	var parsed listMegaTTSTrainStatusResult
	if err := json.Unmarshal(resultRaw, &parsed); err != nil {
		return nil, fmt.Errorf("parse speaker status: %w", err)
	}

	for index := range parsed.Statuses {
		if strings.EqualFold(strings.TrimSpace(parsed.Statuses[index].SpeakerID), strings.TrimSpace(speakerID)) {
			return &parsed.Statuses[index], nil
		}
	}
	return nil, fmt.Errorf("speaker status not found")
}

func (s *VoiceLibraryService) uploadVoiceCloneAudio(ctx context.Context, speakerID string, audioBytes []byte, audioFormat string) error {
	payload := map[string]any{
		"appid":      s.speechAppID,
		"speaker_id": strings.TrimSpace(speakerID),
		"audios": []map[string]any{
			{
				"audio_bytes":  base64.StdEncoding.EncodeToString(audioBytes),
				"audio_format": audioFormat,
			},
		},
		"source":     2,
		"language":   0,
		"model_type": 1,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal voice clone payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cloneUploadURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create voice clone upload request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer;"+s.speechToken)
	httpReq.Header.Set("Resource-Id", s.cloneResourceID)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send voice clone upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read voice clone upload response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("voice clone upload http status %d: %s", resp.StatusCode, truncateVoiceCloneMessage(string(respBody), 240))
	}

	var parsed map[string]any
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil
	}

	if code, ok := parseCodeField(parsed["code"]); ok {
		if code != 0 && code != 20000000 {
			msg := strings.TrimSpace(anyToString(parsed["message"]))
			if msg == "" {
				msg = strings.TrimSpace(anyToString(parsed["msg"]))
			}
			if msg == "" {
				msg = "unknown voice clone upload error"
			}
			return fmt.Errorf("voice clone upload failed (code=%d): %s", code, msg)
		}
	}

	return nil
}

func parseCustomVoicePublicID(raw string) (uint, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "custom-") {
		return 0, false
	}
	idValue := strings.TrimPrefix(trimmed, "custom-")
	parsed, err := strconv.ParseUint(idValue, 10, 64)
	if err != nil {
		return 0, true
	}
	return uint(parsed), true
}

func mapVoiceCloneState(raw string) string {
	state := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case state == "":
		return models.CustomVoiceStatusProcessing
	case strings.Contains(state, "success") || strings.Contains(state, "active") || strings.Contains(state, "finish") || strings.Contains(state, "complete"):
		return models.CustomVoiceStatusCompleted
	case strings.Contains(state, "fail") || strings.Contains(state, "error") || strings.Contains(state, "reject"):
		return models.CustomVoiceStatusFailed
	default:
		return models.CustomVoiceStatusProcessing
	}
}

func normalizeAudioFormat(format string) string {
	trimmed := strings.ToLower(strings.Trim(strings.TrimSpace(format), "."))
	switch trimmed {
	case "wav", "mp3", "m4a", "aac", "ogg", "flac":
		return trimmed
	default:
		return "wav"
	}
}

func parseCodeField(code any) (int, bool) {
	switch value := code.(type) {
	case float64:
		return int(value), true
	case int:
		return value, true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func anyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func truncateVoiceCloneMessage(raw string, maxLen int) string {
	trimmed := strings.TrimSpace(raw)
	if maxLen <= 0 || len(trimmed) <= maxLen {
		return trimmed
	}
	return trimmed[:maxLen]
}
