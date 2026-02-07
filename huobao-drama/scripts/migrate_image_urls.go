package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/drama-generator/backend/infrastructure/storage"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

// This script migrates external image URLs to local storage and updates DB.
// Usage:
//
//	go run scripts/migrate_image_urls.go --config ./configs/config.yaml
//	or provide overrides:
//	go run scripts/migrate_image_urls.go --dsn "user:pass@tcp(127.0.0.1:3306)/db?parseTime=true" \
//	  --local-path "/path/to/data" --base-url "http://localhost:5678/static" --db-type mysql
func main() {
	var dsn string
	var localPath string
	var baseURL string
	var dbType string
	var configPath string
	flag.StringVar(&dsn, "dsn", "", "MySQL DSN")
	flag.StringVar(&localPath, "local-path", "", "local storage path")
	flag.StringVar(&baseURL, "base-url", "", "base URL for static files")
	flag.StringVar(&dbType, "db-type", "", "db type (mysql|sqlite)")
	flag.StringVar(&configPath, "config", "./configs/config.yaml", "config file path")
	flag.Parse()

	if dsn == "" || localPath == "" || baseURL == "" || dbType == "" {
		cfg, err := readConfig(configPath)
		if err != nil {
			log.Fatalf("failed to read config: %v", err)
		}
		if dsn == "" {
			if cfg.Database.Type == "sqlite" {
				dsn = filepath.Clean(cfg.Database.Path)
				dbType = "sqlite"
			} else {
				dsn = cfg.Database.DSN
				dbType = "mysql"
			}
		}
		if localPath == "" {
			localPath = cfg.Storage.LocalPath
		}
		if baseURL == "" {
			baseURL = cfg.Storage.BaseURL
		}
	}

	store, err := storage.NewLocalStorage(localPath, baseURL)
	if err != nil {
		log.Fatalf("failed to init local storage: %v", err)
	}

	driver := "mysql"
	if strings.ToLower(dbType) == "sqlite" {
		driver = "sqlite"
	}
	db, err := sql.Open(driver, dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	migrateTable(ctx, db, store, "characters", "image_url")
	migrateTable(ctx, db, store, "scenes", "image_url")
	migrateTable(ctx, db, store, "image_generations", "image_url")
}

type appConfig struct {
	Database struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
		DSN  string `yaml:"dsn"`
	} `yaml:"database"`
	Storage struct {
		LocalPath string `yaml:"local_path"`
		BaseURL   string `yaml:"base_url"`
	} `yaml:"storage"`
}

func readConfig(path string) (*appConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg appConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func migrateTable(ctx context.Context, db *sql.DB, store *storage.LocalStorage, table, column string) {
	query := fmt.Sprintf("SELECT id, %s FROM %s WHERE %s IS NOT NULL AND %s != ''", column, table, column, column)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("%s query failed: %v", table, err)
		return
	}
	defer rows.Close()

	updated := 0
	for rows.Next() {
		var id int64
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			log.Printf("%s scan failed: %v", table, err)
			continue
		}

		if !isRemoteURL(url) {
			continue
		}

		localURL, err := store.DownloadFromURL(url, "images")
		if err != nil {
			log.Printf("%s id=%d download failed: %v", table, id, err)
			continue
		}
		if localURL == "" {
			continue
		}

		update := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id = ?", table, column)
		if _, err := db.ExecContext(ctx, update, localURL, id); err != nil {
			log.Printf("%s id=%d update failed: %v", table, id, err)
			continue
		}
		updated++
	}

	log.Printf("%s migrated: %d", table, updated)
}

func isRemoteURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}
