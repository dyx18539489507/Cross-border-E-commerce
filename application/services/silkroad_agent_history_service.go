package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SilkroadAgentHistoryItem struct {
	ID             uint                         `json:"id"`
	RequestID      string                       `json:"requestId"`
	ProductName    string                       `json:"productName"`
	Category       string                       `json:"category"`
	TargetMarket   string                       `json:"targetMarket"`
	TargetPlatform string                       `json:"targetPlatform"`
	TargetAudience string                       `json:"targetAudience"`
	RawPrompt      string                       `json:"rawPrompt"`
	Status         string                       `json:"status"`
	Model          string                       `json:"model"`
	Input          SilkroadAgentInput           `json:"input"`
	Result         *SilkroadAgentResult         `json:"result,omitempty"`
	Workflow       *SilkroadAgentWorkflowResult `json:"workflow,omitempty"`
	CreatedAt      time.Time                    `json:"createdAt"`
	UpdatedAt      time.Time                    `json:"updatedAt"`
}

type SilkroadAgentHistoryService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewSilkroadAgentHistoryService(db *gorm.DB, log *logger.Logger) *SilkroadAgentHistoryService {
	return &SilkroadAgentHistoryService{db: db, log: log}
}

func (s *SilkroadAgentHistoryService) Save(deviceID string, input SilkroadAgentInput, result *SilkroadAgentResult) (*SilkroadAgentHistoryItem, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, fmt.Errorf("device id is required")
	}
	if result == nil {
		return nil, fmt.Errorf("agent result is required")
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal agent input failed: %w", err)
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal agent result failed: %w", err)
	}

	record := models.SilkroadAgentSession{
		DeviceID:       deviceID,
		RequestID:      strings.TrimSpace(input.RequestID),
		ProductName:    firstNonBlank(result.RecognizedInfo.ProductName, input.ProductName),
		Category:       firstNonBlank(result.RecognizedInfo.Category, input.Category),
		TargetMarket:   firstNonBlank(result.RecognizedInfo.TargetMarket, input.TargetMarket),
		TargetPlatform: firstNonBlank(result.RecognizedInfo.TargetPlatform, input.TargetPlatform),
		TargetAudience: firstNonBlank(result.RecognizedInfo.TargetAudience, input.TargetAudience),
		RawPrompt:      strings.TrimSpace(input.RawPrompt),
		InputSnapshot:  datatypes.JSON(inputJSON),
		ResultSnapshot: datatypes.JSON(resultJSON),
		Status:         "completed",
		Model:          strings.TrimSpace(result.Model),
	}

	if record.RequestID != "" {
		var existing models.SilkroadAgentSession
		err := s.db.Where("device_id = ? AND request_id = ?", deviceID, record.RequestID).First(&existing).Error
		if err == nil {
			existing.ProductName = record.ProductName
			existing.Category = record.Category
			existing.TargetMarket = record.TargetMarket
			existing.TargetPlatform = record.TargetPlatform
			existing.TargetAudience = record.TargetAudience
			existing.RawPrompt = record.RawPrompt
			existing.InputSnapshot = record.InputSnapshot
			existing.ResultSnapshot = record.ResultSnapshot
			existing.Status = record.Status
			existing.Model = record.Model
			if err := s.db.Save(&existing).Error; err != nil {
				return nil, fmt.Errorf("update agent history failed: %w", err)
			}
			return s.toHistoryItem(existing), nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("query agent history failed: %w", err)
		}
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create agent history failed: %w", err)
	}
	return s.toHistoryItem(record), nil
}

func (s *SilkroadAgentHistoryService) SaveWorkflow(deviceID string, input SilkroadAgentInput, workflow *SilkroadAgentWorkflowResult) (*SilkroadAgentHistoryItem, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, fmt.Errorf("device id is required")
	}
	if workflow == nil {
		return nil, fmt.Errorf("agent workflow result is required")
	}

	result := workflow.Result
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal agent input failed: %w", err)
	}
	resultJSON, err := json.Marshal(&result)
	if err != nil {
		return nil, fmt.Errorf("marshal agent result failed: %w", err)
	}
	workflowJSON, err := json.Marshal(workflow)
	if err != nil {
		return nil, fmt.Errorf("marshal agent workflow failed: %w", err)
	}

	record := models.SilkroadAgentSession{
		DeviceID:         deviceID,
		RequestID:        strings.TrimSpace(input.RequestID),
		ProductName:      firstNonBlank(result.RecognizedInfo.ProductName, input.ProductName),
		Category:         firstNonBlank(result.RecognizedInfo.Category, input.Category),
		TargetMarket:     firstNonBlank(result.RecognizedInfo.TargetMarket, input.TargetMarket),
		TargetPlatform:   firstNonBlank(result.RecognizedInfo.TargetPlatform, input.TargetPlatform),
		TargetAudience:   firstNonBlank(result.RecognizedInfo.TargetAudience, input.TargetAudience),
		RawPrompt:        strings.TrimSpace(input.RawPrompt),
		InputSnapshot:    datatypes.JSON(inputJSON),
		ResultSnapshot:   datatypes.JSON(resultJSON),
		WorkflowSnapshot: datatypes.JSON(workflowJSON),
		Status:           firstNonBlank(workflow.WorkflowStatus, "completed"),
		Model:            strings.TrimSpace(result.Model),
	}

	if record.RequestID != "" {
		var existing models.SilkroadAgentSession
		err := s.db.Where("device_id = ? AND request_id = ?", deviceID, record.RequestID).First(&existing).Error
		if err == nil {
			existing.ProductName = record.ProductName
			existing.Category = record.Category
			existing.TargetMarket = record.TargetMarket
			existing.TargetPlatform = record.TargetPlatform
			existing.TargetAudience = record.TargetAudience
			existing.RawPrompt = record.RawPrompt
			existing.InputSnapshot = record.InputSnapshot
			existing.ResultSnapshot = record.ResultSnapshot
			existing.WorkflowSnapshot = record.WorkflowSnapshot
			existing.Status = record.Status
			existing.Model = record.Model
			if err := s.db.Save(&existing).Error; err != nil {
				return nil, fmt.Errorf("update agent workflow history failed: %w", err)
			}
			return s.toHistoryItem(existing), nil
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("query agent workflow history failed: %w", err)
		}
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create agent workflow history failed: %w", err)
	}
	return s.toHistoryItem(record), nil
}

func (s *SilkroadAgentHistoryService) List(deviceID string, limit int) ([]SilkroadAgentHistoryItem, error) {
	deviceID = strings.TrimSpace(deviceID)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var records []models.SilkroadAgentSession
	query := s.db.Where("device_id = ?", deviceID).Order("updated_at DESC").Limit(limit)
	if err := query.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("list agent history failed: %w", err)
	}

	items := make([]SilkroadAgentHistoryItem, 0, len(records))
	for _, record := range records {
		items = append(items, *s.toHistoryItem(record))
	}
	return items, nil
}

func (s *SilkroadAgentHistoryService) Get(deviceID string, id uint) (*SilkroadAgentHistoryItem, error) {
	var record models.SilkroadAgentSession
	if err := s.db.Where("device_id = ? AND id = ?", strings.TrimSpace(deviceID), id).First(&record).Error; err != nil {
		return nil, err
	}
	return s.toHistoryItem(record), nil
}

func (s *SilkroadAgentHistoryService) toHistoryItem(record models.SilkroadAgentSession) *SilkroadAgentHistoryItem {
	item := &SilkroadAgentHistoryItem{
		ID:             record.ID,
		RequestID:      record.RequestID,
		ProductName:    record.ProductName,
		Category:       record.Category,
		TargetMarket:   record.TargetMarket,
		TargetPlatform: record.TargetPlatform,
		TargetAudience: record.TargetAudience,
		RawPrompt:      record.RawPrompt,
		Status:         record.Status,
		Model:          record.Model,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
	if len(record.InputSnapshot) > 0 {
		_ = json.Unmarshal(record.InputSnapshot, &item.Input)
	}
	if len(record.ResultSnapshot) > 0 {
		var result SilkroadAgentResult
		if err := json.Unmarshal(record.ResultSnapshot, &result); err == nil {
			item.Result = &result
		}
	}
	if len(record.WorkflowSnapshot) > 0 {
		var workflow SilkroadAgentWorkflowResult
		if err := json.Unmarshal(record.WorkflowSnapshot, &workflow); err == nil {
			item.Workflow = &workflow
		}
	}
	return item
}
