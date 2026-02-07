package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/volcengine"
)

type VoiceSpeaker struct {
	ID         string   `json:"id"`
	VoiceType  string   `json:"voice_type"`
	Name       string   `json:"name"`
	Avatar     string   `json:"avatar"`
	Gender     string   `json:"gender"`
	Age        string   `json:"age"`
	TrialURL   string   `json:"trial_url"`
	Categories []string `json:"categories,omitempty"`
}

type VoiceLibraryService struct {
	client  *volcengine.OpenAPIClient
	version string
	log     *logger.Logger
}

func NewVoiceLibraryService(cfg *config.Config, log *logger.Logger) (*VoiceLibraryService, error) {
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

	return &VoiceLibraryService{
		client:  client,
		version: volcengine.DefaultSpeechSaasVersion,
		log:     log,
	}, nil
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

func (s *VoiceLibraryService) ListSpeakers(ctx context.Context) ([]VoiceSpeaker, error) {
	// Volcengine's ListSpeakers API defaults to a small page size (e.g. 10).
	// Use a large Limit to fetch the full public voice library in one call.
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
