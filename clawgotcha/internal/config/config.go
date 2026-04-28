package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds clawgotcha process settings from defaults, optional YAML, and the environment.
// Override priority: environment > YAML file > default (see [Load]).
// YAML keys match shared deploy config (e.g. dep/cl.yaml): database_url, port, hostname, public_url.
type Config struct {
	Hostname    string `yaml:"hostname"`
	Port        int    `yaml:"port"`
	PublicURL   string `yaml:"public_url"`
	DatabaseURL string `yaml:"database_url"`
	// InternalToken gates POST /api/v1/events/publish (Bearer or X-Internal-Token). Set via CLAWGOTCHA_INTERNAL_TOKEN.
	InternalToken string `yaml:"-"`
	// HTTPAddr is the full listen address (e.g. :3477). Set by [Load] from env, or derived from Port; not read from YAML.
	HTTPAddr string `yaml:"-"`
}

// Load merges configuration in this order: defaults, then optional YAML, then environment (highest priority).
// If configPath is empty, YAML is skipped. If configPath is non-empty, the file must exist and be readable.
func Load(configPath string) (*Config, error) {
	c := newDefaults()
	if configPath != "" {
		if err := mergeYAMLFile(c, configPath); err != nil {
			return nil, err
		}
	}
	applyEnv(c)
	deriveHTTPAddr(c)
	return c, nil
}

func newDefaults() *Config {
	return &Config{Port: 3477}
}

func mergeYAMLFile(c *Config, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %q: %w", path, err)
	}
	if err := yaml.Unmarshal(b, c); err != nil {
		return fmt.Errorf("parse config %q: %w", path, err)
	}
	return nil
}

// applyEnv overwrites c with any set environment variable (highest-priority layer).
func applyEnv(c *Config) {
	if v := stringsTrimEnv("DATABASE_URL"); v != nil {
		c.DatabaseURL = *v
	}
	if v := stringsTrimEnv("HTTP_ADDR"); v != nil {
		c.HTTPAddr = *v
	}
	if v := stringsTrimEnv("HOSTNAME"); v != nil {
		c.Hostname = *v
	}
	if v, ok := intFromEnv("PORT"); ok {
		c.Port = v
	}
	if v := stringsTrimEnv("PUBLIC_URL"); v != nil {
		c.PublicURL = *v
	}
	if v := stringsTrimEnv("CLAWGOTCHA_INTERNAL_TOKEN"); v != nil {
		c.InternalToken = *v
	}
}

// deriveHTTPAddr fills HTTPAddr when not set by env, using Port (from defaults + yaml + env PORT).
func deriveHTTPAddr(c *Config) {
	if strings.TrimSpace(c.HTTPAddr) != "" {
		return
	}
	if c.Port > 0 {
		c.HTTPAddr = ":" + strconv.Itoa(c.Port)
	} else {
		c.HTTPAddr = ":8080"
	}
}

func stringsTrimEnv(key string) *string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	s := strings.TrimSpace(v)
	return &s
}

func intFromEnv(key string) (int, bool) {
	v, ok := os.LookupEnv(key)
	if !ok {
		return 0, false
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

// DefaultConfigPath returns the first existing file among common locations (repo root, clawgotcha/, parent dep/).
func DefaultConfigPath() string {
	for _, p := range []string{
		filepath.Join("dep", "cl.yaml"),
		filepath.Join("..", "dep", "cl.yaml"),
		"config.yaml",
		filepath.Join("clawgotcha", "config.yaml"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
