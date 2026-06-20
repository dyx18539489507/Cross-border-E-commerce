/**
 * 模块说明：数字丝路后端配置加载。
 * 业务场景：合规分析、丝路 Agent、视觉理解、数字人和一键分发需要从 YAML 与环境变量组合读取模型及第三方服务配置。
 * 核心职责：加载基础配置，并允许敏感密钥通过环境变量覆盖，避免业务代码直接依赖硬编码配置。
 */
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
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

// 火山引擎配置同时服务视觉模型与数字人链路，丝路 Agent 的图片理解会优先读取对应环境变量再回落到这里的默认接入信息。
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

// 合规分析使用 OpenAI 兼容协议调用 DeepSeek/火山方舟模型，字段命名保持通用以便切换模型供应商。
type ComplianceConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BaseURL  string `mapstructure:"base_url"`
	Endpoint string `mapstructure:"endpoint"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
}

// 分发配置描述 Upload-Post 账号绑定与 Discord 展示信息，真正的用户账号状态仍保存在数据库目标表中。
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

/**
 * 功能：加载 YAML、.env 与环境变量组合后的运行配置。
 * 参数：无。
 * 返回：完整 Config；配置文件缺失或解析失败时返回错误。
 */
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	_ = gotenv.Load(".env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyRuntimeOverrides(&config)

	// 合规校验配置支持通过环境变量注入，避免在配置文件硬编码敏感信息。
	// DeepSeek 与火山方舟均兼容 chat/completions，因此这里只保存协议层配置。
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
		// 分发账号绑定依赖外部 Upload-Post 服务，缺省值让本地演示环境无需额外 YAML 字段即可启动。
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
		config.Distribution.DiscordUsername = firstNonEmpty(os.Getenv("DISTRIBUTION_DISCORD_USERNAME"), "Digital Silk Road")
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

func applyRuntimeOverrides(config *Config) {
	config.App.Name = firstNonEmpty(os.Getenv("APP_NAME"), config.App.Name)
	config.App.Language = firstNonEmpty(os.Getenv("APP_LANGUAGE"), config.App.Language)
	if raw := strings.TrimSpace(os.Getenv("APP_DEBUG")); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			config.App.Debug = parsed
		}
	}

	config.Server.Host = firstNonEmpty(os.Getenv("SERVER_HOST"), config.Server.Host)
	config.Server.Port = readIntEnv("SERVER_PORT", config.Server.Port)
	config.Server.ReadTimeout = readIntEnv("SERVER_READ_TIMEOUT", config.Server.ReadTimeout)
	config.Server.WriteTimeout = readIntEnv("SERVER_WRITE_TIMEOUT", config.Server.WriteTimeout)
	if origins := splitEnvList(os.Getenv("SERVER_CORS_ORIGINS")); len(origins) > 0 {
		config.Server.CORSOrigins = origins
	}

	config.Database.Type = firstNonEmpty(os.Getenv("DATABASE_TYPE"), config.Database.Type)
	config.Database.Path = firstNonEmpty(os.Getenv("DATABASE_PATH"), config.Database.Path)
	config.Database.Host = firstNonEmpty(os.Getenv("DATABASE_HOST"), config.Database.Host)
	config.Database.Port = readIntEnv("DATABASE_PORT", config.Database.Port)
	config.Database.User = firstNonEmpty(os.Getenv("DATABASE_USER"), config.Database.User)
	config.Database.Password = firstNonEmpty(os.Getenv("DATABASE_PASSWORD"), config.Database.Password)
	config.Database.Database = firstNonEmpty(os.Getenv("DATABASE_NAME"), config.Database.Database)
	config.Database.Charset = firstNonEmpty(os.Getenv("DATABASE_CHARSET"), config.Database.Charset)

	config.Storage.Type = firstNonEmpty(os.Getenv("STORAGE_TYPE"), config.Storage.Type)
	config.Storage.LocalPath = firstNonEmpty(os.Getenv("STORAGE_LOCAL_PATH"), config.Storage.LocalPath)
	config.Storage.BaseURL = firstNonEmpty(os.Getenv("STORAGE_BASE_URL"), config.Storage.BaseURL)
	config.Storage.R2AccountID = firstNonEmpty(os.Getenv("R2_ACCOUNT_ID"), config.Storage.R2AccountID)
	config.Storage.R2AccessKeyID = firstNonEmpty(os.Getenv("R2_ACCESS_KEY_ID"), config.Storage.R2AccessKeyID)
	config.Storage.R2SecretKey = firstNonEmpty(os.Getenv("R2_SECRET_ACCESS_KEY"), config.Storage.R2SecretKey)
	config.Storage.R2Bucket = firstNonEmpty(os.Getenv("R2_BUCKET"), config.Storage.R2Bucket)
	config.Storage.R2Endpoint = firstNonEmpty(os.Getenv("R2_ENDPOINT"), config.Storage.R2Endpoint)
	config.Storage.R2Region = firstNonEmpty(os.Getenv("R2_REGION"), config.Storage.R2Region)

	config.Volcengine.AccessKeyID = firstNonEmpty(os.Getenv("VOLCENGINE_ACCESS_KEY_ID"), config.Volcengine.AccessKeyID)
	config.Volcengine.SecretAccessKey = firstNonEmpty(os.Getenv("VOLCENGINE_SECRET_ACCESS_KEY"), config.Volcengine.SecretAccessKey)
	config.Volcengine.Speech.AppID = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_APP_ID"), config.Volcengine.Speech.AppID)
	config.Volcengine.Speech.Token = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_TOKEN"), config.Volcengine.Speech.Token)
	config.Volcengine.Speech.VoiceType = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_VOICE_TYPE"), config.Volcengine.Speech.VoiceType)
}

func splitEnvList(value string) []string {
	items := strings.Split(value, ",")
	result := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
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

/**
 * 功能：按优先级选出第一个非空配置值。
 * 参数：values 通常按“环境变量、本地配置、默认值”的顺序传入。
 * 返回：第一个非空字符串；全部为空时返回空字符串。
 */
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

/**
 * 功能：读取整型环境变量。
 * 参数：key 为环境变量名，fallback 为缺省值。
 * 返回：合法整数；缺失或格式错误时返回 fallback，保证分发轮询等后台任务可继续启动。
 */
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
