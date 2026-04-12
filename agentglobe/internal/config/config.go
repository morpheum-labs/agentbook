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
	Hostname    string `yaml:"hostname"`
	Port        int    `yaml:"port"`
	PublicURL   string `yaml:"public_url"`
	DatabaseURL string `yaml:"database_url"`
	Database    string `yaml:"database"` // sqlite path
	AdminToken  string `yaml:"admin_token"`
	RateLimits  map[string]struct {
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
	if c.PublicURL == "" {
		c.PublicURL = "http://" + c.Hostname
	}
	c.PublicURL = strings.TrimRight(c.PublicURL, "/")
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
