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
	// ServiceRegistryToken authenticates external API services (newsapi, worldmon) to register.
	// When empty, service registry POSTs return 501. Env: SERVICE_REGISTRY_TOKEN.
	ServiceRegistryToken string `yaml:"service_registry_token"`
	AttachmentsDir     string `yaml:"attachments_dir"`
	MaxAttachmentBytes int64  `yaml:"max_attachment_bytes"`
	RateLimits         map[string]struct {
		Limit  int `yaml:"limit"`
		Window int `yaml:"window"`
	} `yaml:"rate_limits"`
	// CORSAllowedOrigins lists browser origins allowed to call the API cross-origin (e.g. https://www.example.com).
	// When empty, Access-Control-Allow-Origin is "*" (dev and backwards compatibility).
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
	// RssLib is a path to monitor-forge-style rss-library.json, or a URL. Relative paths are resolved
	// from the config file’s directory. Used by agentglobe GET /api/v1/public/world-context and worldmon with matching config.
	// Env override: RSS_LIB.
	RssLib string `yaml:"rss_lib"`
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
	if v := os.Getenv("SERVICE_REGISTRY_TOKEN"); v != "" {
		c.ServiceRegistryToken = v
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
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		c.CORSAllowedOrigins = parseCommaSeparatedList(v)
	}
	if v := os.Getenv("RSS_LIB"); v != "" {
		c.RssLib = v
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
	for i := range c.CORSAllowedOrigins {
		c.CORSAllowedOrigins[i] = normalizeOrigin(c.CORSAllowedOrigins[i])
	}
	return c, nil
}

// ResolvedRssLibPath returns an absolute file path for RssLib when it is a local path, or a trimmed URL string.
// Relative paths in config are joined with the directory of configFile; if configFile is empty, the path is resolved from cwd.
func (c *Config) ResolvedRssLibPath(configFile string) string {
	p := strings.TrimSpace(c.RssLib)
	if p == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(p), "http://") || strings.HasPrefix(strings.ToLower(p), "https://") {
		return p
	}
	if !filepath.IsAbs(p) {
		if configFile != "" {
			p = filepath.Join(filepath.Dir(configFile), p)
		} else {
			var err error
			p, err = filepath.Abs(p)
			if err != nil {
				return ""
			}
		}
	}
	return filepath.Clean(p)
}

func parseCommaSeparatedList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normalizeOrigin(o string) string {
	o = strings.TrimSpace(o)
	o = strings.TrimRight(o, "/")
	return o
}

// DefaultConfigPath returns the first existing file among common locations (cwd is usually repo root or agentglobe/).
// Includes dep/cf.yaml so shared deploy-style config can set database_url without exporting CONFIG_PATH.
func DefaultConfigPath() string {
	for _, p := range []string{
		"config.yaml",
		filepath.Join("minibook", "config.yaml"),
		filepath.Join("..", "minibook", "config.yaml"),
		filepath.Join("dep", "cf.yaml"),
		filepath.Join("..", "dep", "cf.yaml"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
