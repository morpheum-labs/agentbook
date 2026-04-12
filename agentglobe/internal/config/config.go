package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config mirrors minibook config.yaml + env overrides.
type Config struct {
	Hostname           string `yaml:"hostname"`
	Port               int    `yaml:"port"`
	PublicURL          string `yaml:"public_url"`
	DatabaseURL        string `yaml:"database_url"`
	Database           string `yaml:"database"` // sqlite path
	AdminToken         string `yaml:"admin_token"`
	AttachmentsDir     string `yaml:"attachments_dir"`
	MaxAttachmentBytes int64  `yaml:"max_attachment_bytes"`
	RateLimits         map[string]struct {
		Limit  int `yaml:"limit"`
		Window int `yaml:"window"`
	} `yaml:"rate_limits"`
}

func Load(configPath string) (*Config, error) {
	c := &Config{
		Hostname:  "localhost:3456",
		Port:      3456,
		PublicURL: "",
		Database:  "data/minibook.db",
	}
	if configPath != "" {
		if b, err := os.ReadFile(configPath); err == nil {
			_ = yaml.Unmarshal(b, c)
		}
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		c.DatabaseURL = v
	}
	if c.DatabaseURL == "" {
		if v := os.Getenv("SQLITE_PATH"); v != "" {
			c.Database = v
		}
	}
	if v := os.Getenv("ADMIN_TOKEN"); v != "" {
		c.AdminToken = v
	}
	if v := os.Getenv("PUBLIC_URL"); v != "" {
		c.PublicURL = v
	}
	if v := os.Getenv("HOSTNAME"); v != "" {
		c.Hostname = v
	}
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Port = p
		}
	}
	if v := os.Getenv("ATTACHMENTS_DIR"); v != "" {
		c.AttachmentsDir = v
	}
	if v := os.Getenv("MAX_ATTACHMENT_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			c.MaxAttachmentBytes = n
		}
	}
	if c.PublicURL == "" {
		c.PublicURL = "http://" + c.Hostname
	}
	c.PublicURL = strings.TrimRight(c.PublicURL, "/")
	if strings.TrimSpace(c.AttachmentsDir) == "" {
		c.AttachmentsDir = "data/attachments"
	}
	if c.MaxAttachmentBytes <= 0 {
		c.MaxAttachmentBytes = 10 * 1024 * 1024
	}
	return c, nil
}

// DefaultConfigPath returns ../../minibook/config.yaml from cwd agentglobe, or ./config.yaml.
func DefaultConfigPath() string {
	for _, p := range []string{
		"config.yaml",
		filepath.Join("minibook", "config.yaml"),
		filepath.Join("..", "minibook", "config.yaml"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
