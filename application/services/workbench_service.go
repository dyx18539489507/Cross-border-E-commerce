package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/logger"
	"gorm.io/gorm"
)

type WorkbenchOverview struct {
	PendingProducts     int64 `json:"pendingProducts"`
	ComplianceCompleted int64 `json:"complianceCompleted"`
	VideosGenerated     int64 `json:"videosGenerated"`
	CoveredMarkets      int64 `json:"coveredMarkets"`
}

type WorkbenchOverviewTrend struct {
	PendingProducts     int64 `json:"pendingProducts"`
	ComplianceCompleted int64 `json:"complianceCompleted"`
	VideosGenerated     int64 `json:"videosGenerated"`
	CoveredMarkets      int64 `json:"coveredMarkets"`
}

type WorkbenchMetricPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type WorkbenchRecentTask struct {
	Title  string `json:"title"`
	Market string `json:"market"`
	Status string `json:"status"`
	Meta   string `json:"meta"`
	Tone   string `json:"tone"`
	Path   string `json:"path,omitempty"`
}

type WorkbenchSummary struct {
	Overview        WorkbenchOverview      `json:"overview"`
	Trends          WorkbenchOverviewTrend `json:"trends"`
	WeeklyActivity  []WorkbenchMetricPoint `json:"weeklyActivity"`
	ConversionTrend []WorkbenchMetricPoint `json:"conversionTrend"`
	RecentTasks     []WorkbenchRecentTask  `json:"recentTasks"`
}

type WorkbenchService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewWorkbenchService(db *gorm.DB, log *logger.Logger) *WorkbenchService {
	return &WorkbenchService{db: db, log: log}
}

func (s *WorkbenchService) Summary(deviceID string) (*WorkbenchSummary, error) {
	deviceID = strings.TrimSpace(deviceID)

	pendingProducts, err := s.countDramas(deviceID, "status IN ?", []string{"draft", "planning", "generating", "error"})
	if err != nil {
		return nil, err
	}
	complianceCompleted, err := s.countDramas(deviceID, "compliance_report IS NOT NULL")
	if err != nil {
		return nil, err
	}
	videosGenerated, err := s.countVideos(deviceID, string(models.VideoStatusCompleted), time.Time{}, time.Time{})
	if err != nil {
		return nil, err
	}
	coveredMarkets, err := s.countCoveredMarkets(deviceID)
	if err != nil {
		return nil, err
	}

	sinceYesterday := time.Now().Add(-24 * time.Hour)
	newPending, _ := s.countDramasSince(deviceID, sinceYesterday)
	newCompliance, _ := s.countDramasSince(deviceID, sinceYesterday, "compliance_report IS NOT NULL")
	newVideos, _ := s.countVideos(deviceID, string(models.VideoStatusCompleted), sinceYesterday, time.Time{})

	weeklyActivity, err := s.weeklyActivity(deviceID)
	if err != nil {
		return nil, err
	}
	conversionTrend, err := s.conversionTrend(deviceID)
	if err != nil {
		return nil, err
	}
	recentTasks, err := s.recentTasks(deviceID)
	if err != nil {
		return nil, err
	}

	return &WorkbenchSummary{
		Overview: WorkbenchOverview{
			PendingProducts:     pendingProducts,
			ComplianceCompleted: complianceCompleted,
			VideosGenerated:     videosGenerated,
			CoveredMarkets:      coveredMarkets,
		},
		Trends: WorkbenchOverviewTrend{
			PendingProducts:     newPending,
			ComplianceCompleted: newCompliance,
			VideosGenerated:     newVideos,
			CoveredMarkets:      0,
		},
		WeeklyActivity:  weeklyActivity,
		ConversionTrend: conversionTrend,
		RecentTasks:     recentTasks,
	}, nil
}

func (s *WorkbenchService) countDramas(deviceID string, condition string, args ...interface{}) (int64, error) {
	var count int64
	query := s.scopeDrama(deviceID)
	if condition != "" {
		query = query.Where(condition, args...)
	}
	err := query.Count(&count).Error
	return count, err
}

func (s *WorkbenchService) countDramasSince(deviceID string, since time.Time, extra ...string) (int64, error) {
	query := s.scopeDrama(deviceID).Where("created_at >= ?", since)
	for _, condition := range extra {
		if strings.TrimSpace(condition) != "" {
			query = query.Where(condition)
		}
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (s *WorkbenchService) countVideos(deviceID string, status string, since time.Time, before time.Time) (int64, error) {
	var count int64
	query := s.db.Model(&models.VideoGeneration{}).
		Joins("JOIN dramas ON dramas.id = video_generations.drama_id")
	if deviceID != "" {
		query = query.Where("dramas.device_id = ?", deviceID)
	}
	if status != "" {
		query = query.Where("video_generations.status = ?", status)
	}
	if !since.IsZero() {
		query = query.Where("video_generations.created_at >= ?", since)
	}
	if !before.IsZero() {
		query = query.Where("video_generations.created_at < ?", before)
	}
	err := query.Count(&count).Error
	return count, err
}

func (s *WorkbenchService) countCoveredMarkets(deviceID string) (int64, error) {
	var dramas []models.Drama
	if err := s.scopeDrama(deviceID).Select("target_country").Find(&dramas).Error; err != nil {
		return 0, err
	}

	markets := make(map[string]struct{})
	for _, drama := range dramas {
		for _, market := range splitCountries(drama.TargetCountry) {
			markets[market] = struct{}{}
		}
	}
	return int64(len(markets)), nil
}

func (s *WorkbenchService) weeklyActivity(deviceID string) ([]WorkbenchMetricPoint, error) {
	labels := []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"}
	now := time.Now()
	monday := startOfDay(now).AddDate(0, 0, -weekdayOffset(now))
	points := make([]WorkbenchMetricPoint, 0, 7)
	for i, label := range labels {
		start := monday.AddDate(0, 0, i)
		end := start.AddDate(0, 0, 1)
		count, err := s.countDramasInRange(deviceID, start, end)
		if err != nil {
			return nil, err
		}
		points = append(points, WorkbenchMetricPoint{Label: label, Value: float64(count)})
	}
	return points, nil
}

func (s *WorkbenchService) conversionTrend(deviceID string) ([]WorkbenchMetricPoint, error) {
	now := time.Now()
	points := make([]WorkbenchMetricPoint, 0, 4)
	for i := 3; i >= 0; i-- {
		start := time.Date(now.Year(), now.Month()-time.Month(i), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0)
		total, err := s.countDramasInRange(deviceID, start, end)
		if err != nil {
			return nil, err
		}
		completed, err := s.countDramasInRange(deviceID, start, end, "status = ?", "completed")
		if err != nil {
			return nil, err
		}
		value := 0.0
		if total > 0 {
			value = math.Round((float64(completed)/float64(total))*800) / 100
		}
		points = append(points, WorkbenchMetricPoint{
			Label: fmt.Sprintf("%d月", int(start.Month())),
			Value: value,
		})
	}
	return points, nil
}

func (s *WorkbenchService) recentTasks(deviceID string) ([]WorkbenchRecentTask, error) {
	var dramas []models.Drama
	if err := s.scopeDrama(deviceID).Order("updated_at DESC").Limit(4).Find(&dramas).Error; err != nil {
		return nil, err
	}

	tasks := make([]WorkbenchRecentTask, 0, len(dramas))
	for _, drama := range dramas {
		tasks = append(tasks, WorkbenchRecentTask{
			Title:  drama.Title,
			Market: firstCountry(drama.TargetCountry),
			Status: statusLabel(drama.Status),
			Meta:   relativeTime(drama.UpdatedAt),
			Tone:   statusTone(drama.Status),
			Path:   fmt.Sprintf("/dramas/%d", drama.ID),
		})
	}
	return tasks, nil
}

func (s *WorkbenchService) countDramasInRange(deviceID string, start time.Time, end time.Time, extra ...interface{}) (int64, error) {
	query := s.scopeDrama(deviceID).Where("created_at >= ? AND created_at < ?", start, end)
	if len(extra) >= 1 {
		condition, _ := extra[0].(string)
		if condition != "" {
			query = query.Where(condition, extra[1:]...)
		}
	}
	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (s *WorkbenchService) scopeDrama(deviceID string) *gorm.DB {
	query := s.db.Model(&models.Drama{})
	if strings.TrimSpace(deviceID) != "" {
		query = query.Where("device_id = ?", strings.TrimSpace(deviceID))
	}
	return query
}

func splitCountries(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '，' || r == ';' || r == '；' || r == '|' || r == '/'
	})
	countries := make([]string, 0, len(parts))
	for _, part := range parts {
		country := strings.TrimSpace(part)
		if country != "" {
			countries = append(countries, country)
		}
	}
	return countries
}

func firstCountry(raw string) string {
	countries := splitCountries(raw)
	if len(countries) == 0 {
		return "未设置"
	}
	return countries[0]
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func weekdayOffset(t time.Time) int {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return 6
	}
	return weekday - 1
}

func statusLabel(status string) string {
	switch status {
	case "completed":
		return "已完成"
	case "production", "generating":
		return "进行中"
	case "error":
		return "异常"
	default:
		return "待处理"
	}
}

func statusTone(status string) string {
	switch status {
	case "completed":
		return "done"
	case "production", "generating":
		return "progress"
	default:
		return "pending"
	}
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "刚刚"
	}
	duration := time.Since(t)
	if duration < time.Minute {
		return "刚刚"
	}
	if duration < time.Hour {
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	}
	return fmt.Sprintf("%d天前", int(duration.Hours()/24))
}
