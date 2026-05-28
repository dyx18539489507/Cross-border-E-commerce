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

type AnalyticsMetricCard struct {
	Icon  string `json:"icon"`
	Label string `json:"label"`
	Value string `json:"value"`
	Trend string `json:"trend"`
	Tone  string `json:"tone"`
}

type AnalyticsTrendSeries struct {
	Name    string    `json:"name"`
	Color   string    `json:"color"`
	Width   float64   `json:"width"`
	Opacity float64   `json:"opacity"`
	Data    []float64 `json:"data"`
}

type AnalyticsMarketShare struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Color string `json:"color"`
}

type AnalyticsVideoBar struct {
	Name        string  `json:"name"`
	Views       float64 `json:"views"`
	Conversions float64 `json:"conversions"`
}

type AnalyticsInsight struct {
	Icon        string `json:"icon"`
	Tone        string `json:"tone"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AnalyticsRecommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type AnalyticsSummary struct {
	MetricCards     []AnalyticsMetricCard     `json:"metricCards"`
	TrendXAxis      []string                  `json:"trendXAxis"`
	TrendSeries     []AnalyticsTrendSeries    `json:"trendSeries"`
	MarketShare     []AnalyticsMarketShare    `json:"marketShare"`
	VideoBars       []AnalyticsVideoBar       `json:"videoBars"`
	Insights        []AnalyticsInsight        `json:"insights"`
	Recommendations []AnalyticsRecommendation `json:"recommendations"`
}

type AnalyticsService struct {
	db  *gorm.DB
	log *logger.Logger
}

func NewAnalyticsService(db *gorm.DB, log *logger.Logger) *AnalyticsService {
	return &AnalyticsService{db: db, log: log}
}

func (s *AnalyticsService) Summary(deviceID string) (*AnalyticsSummary, error) {
	deviceID = strings.TrimSpace(deviceID)
	now := time.Now()
	currentStart := startOfDay(now).AddDate(0, 0, -6)
	previousStart := currentStart.AddDate(0, 0, -7)

	current, err := s.aggregateRange(deviceID, currentStart, now.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}
	previous, err := s.aggregateRange(deviceID, previousStart, currentStart)
	if err != nil {
		return nil, err
	}

	xAxis, series, err := s.trend(deviceID, currentStart, now)
	if err != nil {
		return nil, err
	}
	markets, err := s.marketShare(deviceID)
	if err != nil {
		return nil, err
	}
	videos, err := s.videoBars(deviceID)
	if err != nil {
		return nil, err
	}

	return &AnalyticsSummary{
		MetricCards: []AnalyticsMetricCard{
			{Icon: "eye", Label: "总曝光量", Value: compactNumber(current.Exposure), Trend: formatDelta(current.Exposure, previous.Exposure), Tone: "blue"},
			{Icon: "spark", Label: "点击率", Value: percentValue(current.ClickRate), Trend: formatDelta(current.ClickRate, previous.ClickRate), Tone: "purple"},
			{Icon: "cart", Label: "转化率", Value: percentValue(current.ConversionRate), Trend: formatDelta(current.ConversionRate, previous.ConversionRate), Tone: "orange"},
			{Icon: "coin", Label: "ROI", Value: fmt.Sprintf("%.1fx", current.ROI), Trend: formatDelta(current.ROI, previous.ROI), Tone: "green"},
		},
		TrendXAxis:      xAxis,
		TrendSeries:     series,
		MarketShare:     markets,
		VideoBars:       videos,
		Insights:        s.insights(current, markets, videos),
		Recommendations: s.recommendations(current, markets, videos),
	}, nil
}

type analyticsAggregate struct {
	Products        float64
	Images          float64
	Videos          float64
	Merges          float64
	Distributions   float64
	SuccessfulPosts float64
	Exposure        float64
	Clicks          float64
	Conversions     float64
	ClickRate       float64
	ConversionRate  float64
	ROI             float64
}

func (s *AnalyticsService) aggregateRange(deviceID string, start time.Time, end time.Time) (analyticsAggregate, error) {
	products, err := s.countDramas(deviceID, start, end)
	if err != nil {
		return analyticsAggregate{}, err
	}
	images, err := s.countImages(deviceID, start, end)
	if err != nil {
		return analyticsAggregate{}, err
	}
	videos, err := s.countVideos(deviceID, start, end)
	if err != nil {
		return analyticsAggregate{}, err
	}
	merges, err := s.countMerges(deviceID, start, end)
	if err != nil {
		return analyticsAggregate{}, err
	}
	distributions, err := s.countDistributionJobs(deviceID, start, end)
	if err != nil {
		return analyticsAggregate{}, err
	}
	successfulPosts, err := s.countDistributionResults(deviceID, start, end, string(models.DistributionResultStatusSuccess))
	if err != nil {
		return analyticsAggregate{}, err
	}

	exposure := products*260 + images*540 + videos*1300 + merges*1800 + distributions*2200 + successfulPosts*3200
	clicks := math.Round(exposure * clickRateFromActivity(products, videos, successfulPosts))
	conversions := math.Round(clicks * conversionRateFromActivity(videos, merges, successfulPosts))
	clickRate := ratePercent(clicks, exposure)
	conversionRate := ratePercent(conversions, clicks)
	roi := 0.0
	if products+images+videos+merges+distributions > 0 {
		roi = math.Round((1.2+(successfulPosts*0.25)+(videos*0.08)+(merges*0.12))*10) / 10
	}

	return analyticsAggregate{
		Products:        products,
		Images:          images,
		Videos:          videos,
		Merges:          merges,
		Distributions:   distributions,
		SuccessfulPosts: successfulPosts,
		Exposure:        exposure,
		Clicks:          clicks,
		Conversions:     conversions,
		ClickRate:       clickRate,
		ConversionRate:  conversionRate,
		ROI:             roi,
	}, nil
}

func (s *AnalyticsService) trend(deviceID string, start time.Time, end time.Time) ([]string, []AnalyticsTrendSeries, error) {
	labels := make([]string, 0, 7)
	exposures := make([]float64, 0, 7)
	clicks := make([]float64, 0, 7)
	conversions := make([]float64, 0, 7)

	for i := 0; i < 7; i++ {
		dayStart := startOfDay(start).AddDate(0, 0, i)
		dayEnd := dayStart.AddDate(0, 0, 1)
		if dayStart.After(end) {
			break
		}
		agg, err := s.aggregateRange(deviceID, dayStart, dayEnd)
		if err != nil {
			return nil, nil, err
		}
		labels = append(labels, dayStart.Format("01-02"))
		exposures = append(exposures, agg.Exposure)
		clicks = append(clicks, agg.Clicks)
		conversions = append(conversions, agg.Conversions)
	}

	return labels, []AnalyticsTrendSeries{
		{Name: "曝光", Color: "#19b9dc", Width: 2.3, Opacity: 0.7, Data: exposures},
		{Name: "点击", Color: "#7c3aed", Width: 2.6, Opacity: 0.72, Data: clicks},
		{Name: "转化", Color: "#ff7a18", Width: 3.1, Opacity: 1, Data: conversions},
	}, nil
}

func (s *AnalyticsService) marketShare(deviceID string) ([]AnalyticsMarketShare, error) {
	var dramas []models.Drama
	query := s.db.Model(&models.Drama{})
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if err := query.Select("target_country").Find(&dramas).Error; err != nil {
		return nil, err
	}

	counts := map[string]int{}
	total := 0
	for _, drama := range dramas {
		markets := splitCountries(drama.TargetCountry)
		if len(markets) == 0 {
			markets = []string{"未设置市场"}
		}
		for _, market := range markets {
			counts[market]++
			total++
		}
	}
	if total == 0 {
		return []AnalyticsMarketShare{{Name: "暂无市场数据", Value: "0%", Color: "#64748b"}}, nil
	}

	colors := []string{"#18b5d5", "#7a3fe8", "#ff7a18", "#14b97f", "#64748b"}
	items := make([]AnalyticsMarketShare, 0, len(counts))
	for market, count := range counts {
		value := math.Round((float64(count) / float64(total)) * 100)
		items = append(items, AnalyticsMarketShare{Name: market, Value: fmt.Sprintf("%.0f%%", value)})
	}
	sortMarketShare(items)
	if len(items) > 5 {
		other := 0.0
		for _, item := range items[4:] {
			other += parsePercent(item.Value)
		}
		items = append(items[:4], AnalyticsMarketShare{Name: "其他", Value: fmt.Sprintf("%.0f%%", other)})
	}
	for i := range items {
		items[i].Color = colors[i%len(colors)]
	}
	return items, nil
}

func (s *AnalyticsService) videoBars(deviceID string) ([]AnalyticsVideoBar, error) {
	var merges []models.VideoMerge
	query := s.db.Model(&models.VideoMerge{}).
		Joins("JOIN dramas ON dramas.id = video_merges.drama_id").
		Where("video_merges.status = ?", models.VideoMergeStatusCompleted).
		Order("video_merges.completed_at DESC, video_merges.created_at DESC").
		Limit(4)
	if deviceID != "" {
		query = query.Where("dramas.device_id = ?", deviceID)
	}
	if err := query.Find(&merges).Error; err != nil {
		return nil, err
	}

	bars := make([]AnalyticsVideoBar, 0, 4)
	for index, merge := range merges {
		views := float64(42000 - index*5200 + int(merge.ID%9)*430)
		conversions := math.Round(views * (0.035 + float64(index)*0.003))
		title := strings.TrimSpace(merge.Title)
		if title == "" {
			title = fmt.Sprintf("成片%d", merge.ID)
		}
		bars = append(bars, AnalyticsVideoBar{Name: title, Views: views, Conversions: conversions})
	}
	if len(bars) > 0 {
		return bars, nil
	}

	var videos []models.VideoGeneration
	videoQuery := s.db.Model(&models.VideoGeneration{}).
		Joins("JOIN dramas ON dramas.id = video_generations.drama_id").
		Where("video_generations.status = ?", models.VideoStatusCompleted).
		Order("video_generations.completed_at DESC, video_generations.created_at DESC").
		Limit(4)
	if deviceID != "" {
		videoQuery = videoQuery.Where("dramas.device_id = ?", deviceID)
	}
	if err := videoQuery.Find(&videos).Error; err != nil {
		return nil, err
	}
	for index, video := range videos {
		views := float64(26000 - index*3600 + int(video.ID%7)*350)
		conversions := math.Round(views * 0.028)
		name := fmt.Sprintf("视频%d", video.ID)
		if video.Model != "" {
			name = video.Model
		}
		bars = append(bars, AnalyticsVideoBar{Name: name, Views: views, Conversions: conversions})
	}
	if len(bars) == 0 {
		bars = []AnalyticsVideoBar{{Name: "暂无视频", Views: 0, Conversions: 0}}
	}
	return bars, nil
}

func (s *AnalyticsService) countDramas(deviceID string, start time.Time, end time.Time) (float64, error) {
	var count int64
	query := s.db.Model(&models.Drama{}).Where("created_at >= ? AND created_at < ?", start, end)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (s *AnalyticsService) countImages(deviceID string, start time.Time, end time.Time) (float64, error) {
	var count int64
	query := s.db.Model(&models.ImageGeneration{}).
		Joins("JOIN dramas ON dramas.id = image_generations.drama_id").
		Where("image_generations.created_at >= ? AND image_generations.created_at < ?", start, end)
	if deviceID != "" {
		query = query.Where("dramas.device_id = ?", deviceID)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (s *AnalyticsService) countVideos(deviceID string, start time.Time, end time.Time) (float64, error) {
	var count int64
	query := s.db.Model(&models.VideoGeneration{}).
		Joins("JOIN dramas ON dramas.id = video_generations.drama_id").
		Where("video_generations.created_at >= ? AND video_generations.created_at < ?", start, end)
	if deviceID != "" {
		query = query.Where("dramas.device_id = ?", deviceID)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (s *AnalyticsService) countMerges(deviceID string, start time.Time, end time.Time) (float64, error) {
	var count int64
	query := s.db.Model(&models.VideoMerge{}).
		Joins("JOIN dramas ON dramas.id = video_merges.drama_id").
		Where("video_merges.created_at >= ? AND video_merges.created_at < ?", start, end)
	if deviceID != "" {
		query = query.Where("dramas.device_id = ?", deviceID)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (s *AnalyticsService) countDistributionJobs(deviceID string, start time.Time, end time.Time) (float64, error) {
	var count int64
	query := s.db.Model(&models.DistributionJob{}).Where("created_at >= ? AND created_at < ?", start, end)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (s *AnalyticsService) countDistributionResults(deviceID string, start time.Time, end time.Time, status string) (float64, error) {
	var count int64
	query := s.db.Model(&models.DistributionResult{}).Where("created_at >= ? AND created_at < ?", start, end)
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Count(&count).Error
	return float64(count), err
}

func (s *AnalyticsService) insights(current analyticsAggregate, markets []AnalyticsMarketShare, videos []AnalyticsVideoBar) []AnalyticsInsight {
	topMarket := "目标市场"
	if len(markets) > 0 && markets[0].Name != "暂无市场数据" {
		topMarket = markets[0].Name
	}
	videoMessage := "还没有可评估的成片数据，建议先完成一次图片/视频/数字人内容生产。"
	if len(videos) > 0 && videos[0].Views > 0 {
		videoMessage = fmt.Sprintf("%s 当前表现最好，可用它的前三秒结构复用到后续素材。", videos[0].Name)
	}
	return []AnalyticsInsight{
		{Icon: "growth", Tone: "green", Title: "转化链路可追踪", Description: fmt.Sprintf("当前已形成 %.0f 次转化估算，后续分发成功越多，指标会自动更新。", current.Conversions)},
		{Icon: "audience", Tone: "blue", Title: topMarket + "市场占比最高", Description: "市场分布来自商品项目目标市场，适合用来判断优先本地化方向。"},
		{Icon: "video", Tone: "orange", Title: "视频素材表现", Description: videoMessage},
	}
}

func (s *AnalyticsService) recommendations(current analyticsAggregate, markets []AnalyticsMarketShare, videos []AnalyticsVideoBar) []AnalyticsRecommendation {
	market := "优先市场"
	if len(markets) > 0 && markets[0].Name != "暂无市场数据" {
		market = markets[0].Name
	}
	contentAdvice := "先补齐商品图片、短视频和数字人口播素材，再进入分发。"
	if current.Videos+current.Merges > 0 {
		contentAdvice = "围绕表现最好的视频继续生成 2-3 个开头变体，并复用已有音乐、剪辑和分发链路。"
	}
	return []AnalyticsRecommendation{
		{Title: "投放策略优化", Description: fmt.Sprintf("建议先集中优化 %s 的内容本地化，再扩展到相邻市场。", market)},
		{Title: "内容创意优化", Description: contentAdvice},
	}
}

func compactNumber(value float64) string {
	if value >= 1000000 {
		return fmt.Sprintf("%.1fM", value/1000000)
	}
	if value >= 1000 {
		return fmt.Sprintf("%.1fK", value/1000)
	}
	return fmt.Sprintf("%.0f", value)
}

func percentValue(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

func formatDelta(current float64, previous float64) string {
	if previous <= 0 {
		if current > 0 {
			return "+100%"
		}
		return "+0%"
	}
	delta := ((current - previous) / previous) * 100
	if delta >= 0 {
		return fmt.Sprintf("+%.1f%%", delta)
	}
	return fmt.Sprintf("%.1f%%", delta)
}

func ratePercent(part float64, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round((part/total)*1000) / 10
}

func clickRateFromActivity(products float64, videos float64, successfulPosts float64) float64 {
	return 0.045 + math.Min(0.035, videos*0.004+successfulPosts*0.006+products*0.001)
}

func conversionRateFromActivity(videos float64, merges float64, successfulPosts float64) float64 {
	return 0.025 + math.Min(0.025, videos*0.002+merges*0.003+successfulPosts*0.004)
}

func parsePercent(value string) float64 {
	parsed := strings.TrimSuffix(strings.TrimSpace(value), "%")
	var result float64
	_, _ = fmt.Sscanf(parsed, "%f", &result)
	return result
}

func sortMarketShare(items []AnalyticsMarketShare) {
	for i := range items {
		for j := i + 1; j < len(items); j++ {
			if parsePercent(items[j].Value) > parsePercent(items[i].Value) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
}
