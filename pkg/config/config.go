package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Storage      StorageConfig      `mapstructure:"storage"`
	AI           AIConfig           `mapstructure:"ai"`
	Volcengine   VolcengineConfig   `mapstructure:"volcengine"`
	Compliance   ComplianceConfig   `mapstructure:"compliance"`
	Distribution DistributionConfig `mapstructure:"distribution"`
}

type AppConfig struct {
	Name     string `mapstructure:"name"`
	Version  string `mapstructure:"version"`
	Debug    bool   `mapstructure:"debug"`
	Language string `mapstructure:"language"` // zh 或 en
}

type ServerConfig struct {
	Port         int      `mapstructure:"port"`
	Host         string   `mapstructure:"host"`
	CORSOrigins  []string `mapstructure:"cors_origins"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // sqlite, mysql
	Path     string `mapstructure:"path"` // SQLite数据库文件路径
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
	MaxIdle  int    `mapstructure:"max_idle"`
	MaxOpen  int    `mapstructure:"max_open"`
}

type StorageConfig struct {
	Type          string `mapstructure:"type"`       // local, minio
	LocalPath     string `mapstructure:"local_path"` // 本地存储路径
	BaseURL       string `mapstructure:"base_url"`   // 访问URL前缀
	R2AccountID   string `mapstructure:"r2_account_id"`
	R2AccessKeyID string `mapstructure:"r2_access_key_id"`
	R2SecretKey   string `mapstructure:"r2_secret_access_key"`
	R2Bucket      string `mapstructure:"r2_bucket"`
	R2Endpoint    string `mapstructure:"r2_endpoint"`
	R2Region      string `mapstructure:"r2_region"`
}

type AIConfig struct {
	DefaultTextProvider  string `mapstructure:"default_text_provider"`
	DefaultImageProvider string `mapstructure:"default_image_provider"`
	DefaultVideoProvider string `mapstructure:"default_video_provider"`
}

type VolcengineConfig struct {
	AccessKeyID     string                 `mapstructure:"access_key_id"`
	SecretAccessKey string                 `mapstructure:"secret_access_key"`
	Region          string                 `mapstructure:"region"`
	Service         string                 `mapstructure:"service"`
	VisualHost      string                 `mapstructure:"visual_host"`
	Speech          VolcengineSpeechConfig `mapstructure:"speech"`
}

type VolcengineSpeechConfig struct {
	AppID               string `mapstructure:"app_id"`
	Token               string `mapstructure:"token"`
	Cluster             string `mapstructure:"cluster"`
	Endpoint            string `mapstructure:"endpoint"`
	SubmitEndpoint      string `mapstructure:"submit_endpoint"`
	QueryEndpoint       string `mapstructure:"query_endpoint"`
	ResourceID          string `mapstructure:"resource_id"`
	Namespace           string `mapstructure:"namespace"`
	VoiceType           string `mapstructure:"voice_type"`
	CloneUploadEndpoint string `mapstructure:"clone_upload_endpoint"`
	CloneResourceID     string `mapstructure:"clone_resource_id"`
	CloneProjectName    string `mapstructure:"clone_project_name"`
}

type ComplianceConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BaseURL  string `mapstructure:"base_url"`
	Endpoint string `mapstructure:"endpoint"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
}

type DistributionConfig struct {
	UploadPostBaseURL        string `mapstructure:"upload_post_base_url"`
	UploadPostConnectTitle   string `mapstructure:"upload_post_connect_title"`
	UploadPostConnectDesc    string `mapstructure:"upload_post_connect_description"`
	UploadPostRedirectURL    string `mapstructure:"upload_post_redirect_url"`
	UploadPostLogoImage      string `mapstructure:"upload_post_logo_image"`
	DiscordUsername          string `mapstructure:"discord_username"`
	DiscordAvatarURL         string `mapstructure:"discord_avatar_url"`
	StatusPollIntervalSecond int    `mapstructure:"status_poll_interval_seconds"`
	HistoryLookbackPages     int    `mapstructure:"history_lookback_pages"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 合规校验配置支持通过环境变量注入，避免在配置文件硬编码敏感信息
	if config.Compliance.BaseURL == "" {
		config.Compliance.BaseURL = firstNonEmpty(os.Getenv("COMPLIANCE_BASE_URL"), "https://ark.cn-beijing.volces.com/api/v3")
	}
	if config.Compliance.Endpoint == "" {
		config.Compliance.Endpoint = firstNonEmpty(os.Getenv("COMPLIANCE_ENDPOINT"), "/chat/completions")
	}
	if config.Compliance.Model == "" {
		config.Compliance.Model = firstNonEmpty(os.Getenv("COMPLIANCE_MODEL"), "deepseek-v3-2-251201")
	}
	config.Compliance.APIKey = firstNonEmpty(os.Getenv("COMPLIANCE_API_KEY"), os.Getenv("DEEPSEEK_API_KEY"), config.Compliance.APIKey)

	if envEnabled := os.Getenv("COMPLIANCE_ENABLED"); envEnabled != "" {
		if parsed, err := strconv.ParseBool(envEnabled); err == nil {
			config.Compliance.Enabled = parsed
		}
	} else if !config.Compliance.Enabled {
		// 未显式配置时，默认启用并在运行期根据 API Key 自动回退
		config.Compliance.Enabled = true
	}

	if config.Distribution.UploadPostBaseURL == "" {
		config.Distribution.UploadPostBaseURL = firstNonEmpty(os.Getenv("UPLOAD_POST_BASE_URL"), "https://api.upload-post.com/api")
	}
	if config.Distribution.UploadPostConnectTitle == "" {
		config.Distribution.UploadPostConnectTitle = firstNonEmpty(os.Getenv("UPLOAD_POST_CONNECT_TITLE"), "Connect Pinterest / Reddit")
	}
	if config.Distribution.UploadPostConnectDesc == "" {
		config.Distribution.UploadPostConnectDesc = firstNonEmpty(os.Getenv("UPLOAD_POST_CONNECT_DESCRIPTION"), "Connect your own Pinterest and Reddit accounts before distributing content.")
	}
	if config.Distribution.UploadPostRedirectURL == "" {
		config.Distribution.UploadPostRedirectURL = os.Getenv("UPLOAD_POST_REDIRECT_URL")
	}
	if config.Distribution.UploadPostLogoImage == "" {
		config.Distribution.UploadPostLogoImage = os.Getenv("UPLOAD_POST_LOGO_IMAGE")
	}
	if config.Distribution.DiscordUsername == "" {
		config.Distribution.DiscordUsername = firstNonEmpty(os.Getenv("DISTRIBUTION_DISCORD_USERNAME"), "Drama Generator")
	}
	if config.Distribution.DiscordAvatarURL == "" {
		config.Distribution.DiscordAvatarURL = os.Getenv("DISTRIBUTION_DISCORD_AVATAR_URL")
	}
	if config.Distribution.StatusPollIntervalSecond <= 0 {
		config.Distribution.StatusPollIntervalSecond = readIntEnv("DISTRIBUTION_STATUS_POLL_INTERVAL_SECONDS", 20)
	}
	if config.Distribution.HistoryLookbackPages <= 0 {
		config.Distribution.HistoryLookbackPages = readIntEnv("DISTRIBUTION_HISTORY_LOOKBACK_PAGES", 3)
	}

	return &config, nil
}

func (c *DatabaseConfig) DSN() string {
	if c.Type == "sqlite" {
		return c.Path
	}
	// MySQL DSN
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
	)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func readIntEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}

	return parsed
}
