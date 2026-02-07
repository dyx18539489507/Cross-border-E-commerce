package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/google/uuid"
)

type UploadService struct {
	storagePath string
	baseURL     string
	log         *logger.Logger
	r2Client    *s3.Client
	r2Bucket    string
	r2Enabled   bool
}

func NewUploadService(cfg *config.Config, log *logger.Logger) (*UploadService, error) {
	// 确保存储目录存在
	if err := os.MkdirAll(cfg.Storage.LocalPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	service := &UploadService{
		storagePath: cfg.Storage.LocalPath,
		baseURL:     cfg.Storage.BaseURL,
		log:         log,
	}

	if strings.EqualFold(cfg.Storage.Type, "r2") {
		r2Client, bucket, err := newR2Client(cfg)
		if err != nil {
			return nil, err
		}
		service.r2Client = r2Client
		service.r2Bucket = bucket
		service.r2Enabled = true
		if service.baseURL == "" {
			return nil, fmt.Errorf("storage.base_url is required for r2")
		}
	}

	return service, nil
}

// UploadFile 上传文件到本地存储
func (s *UploadService) UploadFile(file io.Reader, fileName, contentType string, category string) (string, error) {
	// 创建分类目录
	categoryPath := filepath.Join(s.storagePath, category)
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create category directory: %w", err)
	}

	// 生成唯一文件名
	ext := filepath.Ext(fileName)
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102_150405")
	newFileName := fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)
	filePath := filepath.Join(categoryPath, newFileName)

	// 创建文件
	dst, err := os.Create(filePath)
	if err != nil {
		s.log.Errorw("Failed to create file", "error", err, "path", filePath)
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 写入文件
	if _, err := io.Copy(dst, file); err != nil {
		s.log.Errorw("Failed to write file", "error", err, "path", filePath)
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	relPath := fmt.Sprintf("%s/%s", category, newFileName)

	if s.r2Enabled {
		if err := s.uploadToR2(filePath, relPath, contentType); err != nil {
			s.log.Errorw("Failed to upload to R2", "error", err, "path", filePath)
			return "", fmt.Errorf("上传到R2失败: %w", err)
		}
	}

	// 构建访问URL
	fileURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(s.baseURL, "/"), relPath)

	s.log.Infow("File uploaded successfully", "path", filePath, "url", fileURL)
	return fileURL, nil
}

// UploadCharacterImage 上传角色图片
func (s *UploadService) UploadCharacterImage(file io.Reader, fileName, contentType string) (string, error) {
	return s.UploadFile(file, fileName, contentType, "characters")
}

// DeleteFile 删除本地文件
func (s *UploadService) DeleteFile(fileURL string) error {
	// 从URL中提取相对路径
	// URL格式: http://localhost:8080/static/characters/20060102_150405_uuid.jpg
	relPath := s.extractRelativePathFromURL(fileURL)
	if relPath == "" {
		return fmt.Errorf("invalid file URL")
	}

	filePath := filepath.Join(s.storagePath, relPath)
	err := os.Remove(filePath)
	if err != nil {
		s.log.Errorw("Failed to delete file", "error", err, "path", filePath)
		return fmt.Errorf("删除文件失败: %w", err)
	}

	if s.r2Enabled {
		if err := s.deleteFromR2(relPath); err != nil {
			s.log.Errorw("Failed to delete file from R2", "error", err, "key", relPath)
		}
	}

	s.log.Infow("File deleted successfully", "path", filePath)
	return nil
}

// extractRelativePathFromURL 从URL中提取相对路径
func (s *UploadService) extractRelativePathFromURL(fileURL string) string {
	// 从baseURL后面提取路径
	// 例如: http://localhost:8080/static/characters/xxx.jpg -> characters/xxx.jpg
	if len(fileURL) <= len(s.baseURL) {
		return ""
	}
	return fileURL[len(s.baseURL)+1:] // +1 for the '/'
}

// GetPresignedURL 本地存储不需要预签名URL，直接返回原URL
func (s *UploadService) GetPresignedURL(objectName string, expiry time.Duration) (string, error) {
	// 本地存储通过静态文件服务直接访问，不需要预签名
	return fmt.Sprintf("%s/%s", s.baseURL, objectName), nil
}

func newR2Client(cfg *config.Config) (*s3.Client, string, error) {
	if cfg.Storage.R2AccessKeyID == "" || cfg.Storage.R2SecretKey == "" || cfg.Storage.R2Bucket == "" {
		return nil, "", fmt.Errorf("r2_access_key_id, r2_secret_access_key, r2_bucket are required for r2")
	}

	endpoint := cfg.Storage.R2Endpoint
	if endpoint == "" {
		if cfg.Storage.R2AccountID == "" {
			return nil, "", fmt.Errorf("r2_account_id or r2_endpoint is required for r2")
		}
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.Storage.R2AccountID)
	}

	region := cfg.Storage.R2Region
	if region == "" {
		region = "auto"
	}

	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(region),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.Storage.R2AccessKeyID, cfg.Storage.R2SecretKey, "")),
		awsConfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				if service == s3.ServiceID {
					return aws.Endpoint{
						URL:               endpoint,
						SigningRegion:     region,
						HostnameImmutable: true,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			},
		)),
	)
	if err != nil {
		return nil, "", fmt.Errorf("init r2 config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return client, cfg.Storage.R2Bucket, nil
}

func (s *UploadService) uploadToR2(filePath, objectKey, contentType string) error {
	if s.r2Client == nil {
		return fmt.Errorf("r2 client not initialized")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.r2Bucket),
		Key:         aws.String(objectKey),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	_, err = s.r2Client.PutObject(context.Background(), input)
	return err
}

func (s *UploadService) deleteFromR2(objectKey string) error {
	if s.r2Client == nil {
		return fmt.Errorf("r2 client not initialized")
	}
	_, err := s.r2Client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.r2Bucket),
		Key:    aws.String(objectKey),
	})
	return err
}
