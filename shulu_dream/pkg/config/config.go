package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Storage    StorageConfig    `mapstructure:"storage"`
	AI         AIConfig         `mapstructure:"ai"`
	Volcengine VolcengineConfig `mapstructure:"volcengine"`
	Compliance ComplianceConfig `mapstructure:"compliance"`
	SFX        SFXConfig        `mapstructure:"sfx"`
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
	Type      string `mapstructure:"type"`       // local, minio
	LocalPath string `mapstructure:"local_path"` // 本地存储路径
	BaseURL   string `mapstructure:"base_url"`   // 访问URL前缀
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
	AppID          string `mapstructure:"app_id"`
	Token          string `mapstructure:"token"`
	Cluster        string `mapstructure:"cluster"`
	Endpoint       string `mapstructure:"endpoint"`
	SubmitEndpoint string `mapstructure:"submit_endpoint"`
	QueryEndpoint  string `mapstructure:"query_endpoint"`
	ResourceID     string `mapstructure:"resource_id"`
	Namespace      string `mapstructure:"namespace"`
	VoiceType      string `mapstructure:"voice_type"`
}

type ComplianceConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	BaseURL  string `mapstructure:"base_url"`
	Endpoint string `mapstructure:"endpoint"`
	APIKey   string `mapstructure:"api_key"`
	Model    string `mapstructure:"model"`
}

type SFXConfig struct {
	DefaultLimit   int                  `mapstructure:"default_limit"`
	RequestTimeout int                  `mapstructure:"request_timeout"`
	Freesound      FreesoundConfig      `mapstructure:"freesound"`
	Pixabay        PixabaySFXConfig     `mapstructure:"pixabay"`
	Translation    SFXTranslationConfig `mapstructure:"translation"`
}

type FreesoundConfig struct {
	ClientID string `mapstructure:"client_id"`
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
}

type PixabaySFXConfig struct {
	APIKey  string `mapstructure:"api_key"`
	BaseURL string `mapstructure:"base_url"`
}

type SFXTranslationConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	AppID    string `mapstructure:"app_id"`
	APIKey   string `mapstructure:"api_key"`
	Endpoint string `mapstructure:"endpoint"`
	From     string `mapstructure:"from"`
	To       string `mapstructure:"to"`
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

	mergeVolcengineSpeechConfig(&config.Volcengine.Speech, loadSharedVolcengineSpeechConfig())
	config.Storage.LocalPath = firstNonEmpty(os.Getenv("STORAGE_LOCAL_PATH"), config.Storage.LocalPath)
	config.Storage.BaseURL = firstNonEmpty(os.Getenv("STORAGE_BASE_URL"), config.Storage.BaseURL)
	config.Volcengine.Speech.AppID = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_APP_ID"), config.Volcengine.Speech.AppID)
	config.Volcengine.Speech.Token = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_TOKEN"), config.Volcengine.Speech.Token)
	config.Volcengine.Speech.Cluster = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_CLUSTER"), config.Volcengine.Speech.Cluster)
	config.Volcengine.Speech.Endpoint = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_ENDPOINT"), config.Volcengine.Speech.Endpoint)
	config.Volcengine.Speech.SubmitEndpoint = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_SUBMIT_ENDPOINT"), config.Volcengine.Speech.SubmitEndpoint)
	config.Volcengine.Speech.QueryEndpoint = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_QUERY_ENDPOINT"), config.Volcengine.Speech.QueryEndpoint)
	config.Volcengine.Speech.ResourceID = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_RESOURCE_ID"), config.Volcengine.Speech.ResourceID)
	config.Volcengine.Speech.Namespace = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_NAMESPACE"), config.Volcengine.Speech.Namespace)
	config.Volcengine.Speech.VoiceType = firstNonEmpty(os.Getenv("VOLCENGINE_SPEECH_VOICE_TYPE"), config.Volcengine.Speech.VoiceType)

	if envPort := os.Getenv("SERVER_PORT"); envPort != "" {
		port, err := strconv.Atoi(envPort)
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("invalid SERVER_PORT %q: must be an integer between 1 and 65535", envPort)
		}
		config.Server.Port = port
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

	// 音效名称翻译配置（Youdao）
	if config.SFX.Translation.Endpoint == "" {
		config.SFX.Translation.Endpoint = firstNonEmpty(os.Getenv("SFX_TRANSLATION_ENDPOINT"), "https://openapi.youdao.com/api")
	} else {
		config.SFX.Translation.Endpoint = firstNonEmpty(os.Getenv("SFX_TRANSLATION_ENDPOINT"), config.SFX.Translation.Endpoint)
	}
	config.SFX.Translation.AppID = firstNonEmpty(os.Getenv("SFX_TRANSLATION_APP_ID"), os.Getenv("YOUDAO_APP_ID"), config.SFX.Translation.AppID)
	config.SFX.Translation.APIKey = firstNonEmpty(os.Getenv("SFX_TRANSLATION_API_KEY"), os.Getenv("YOUDAO_API_KEY"), config.SFX.Translation.APIKey)
	config.SFX.Translation.From = firstNonEmpty(os.Getenv("SFX_TRANSLATION_FROM"), config.SFX.Translation.From, "auto")
	config.SFX.Translation.To = firstNonEmpty(os.Getenv("SFX_TRANSLATION_TO"), config.SFX.Translation.To, "zh-CHS")
	if envEnabled := os.Getenv("SFX_TRANSLATION_ENABLED"); envEnabled != "" {
		if parsed, err := strconv.ParseBool(envEnabled); err == nil {
			config.SFX.Translation.Enabled = parsed
		}
	} else if !config.SFX.Translation.Enabled {
		// 未显式设置 enabled 时，只要配置了 app_id/api_key 即启用
		config.SFX.Translation.Enabled = config.SFX.Translation.AppID != "" && config.SFX.Translation.APIKey != ""
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

func loadSharedVolcengineSpeechConfig() VolcengineSpeechConfig {
	sharedViper := viper.New()
	sharedViper.SetConfigName("config")
	sharedViper.SetConfigType("yaml")
	sharedViper.AddConfigPath("../configs")

	if err := sharedViper.ReadInConfig(); err != nil {
		return VolcengineSpeechConfig{}
	}

	var shared struct {
		Volcengine struct {
			Speech VolcengineSpeechConfig `mapstructure:"speech"`
		} `mapstructure:"volcengine"`
	}
	if err := sharedViper.Unmarshal(&shared); err != nil {
		return VolcengineSpeechConfig{}
	}

	return shared.Volcengine.Speech
}

func mergeVolcengineSpeechConfig(target *VolcengineSpeechConfig, fallback VolcengineSpeechConfig) {
	if target == nil {
		return
	}

	target.AppID = firstNonEmpty(target.AppID, fallback.AppID)
	target.Token = firstNonEmpty(target.Token, fallback.Token)
	target.Cluster = firstNonEmpty(target.Cluster, fallback.Cluster)
	target.Endpoint = firstNonEmpty(target.Endpoint, fallback.Endpoint)
	target.SubmitEndpoint = firstNonEmpty(target.SubmitEndpoint, fallback.SubmitEndpoint)
	target.QueryEndpoint = firstNonEmpty(target.QueryEndpoint, fallback.QueryEndpoint)
	target.ResourceID = firstNonEmpty(target.ResourceID, fallback.ResourceID)
	target.Namespace = firstNonEmpty(target.Namespace, fallback.Namespace)
	target.VoiceType = firstNonEmpty(target.VoiceType, fallback.VoiceType)
}
