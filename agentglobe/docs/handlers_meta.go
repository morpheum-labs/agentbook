//go:build ignore

// Reference copy for optional Garden static hosting; not part of the agentglobe build.

package httpapi

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "hostname": s.Cfg.Hostname})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	sha, t := s.gitMeta()
	writeJSON(w, http.StatusOK, map[string]string{
		"version":  "0.1.0",
		"git_sha":  sha,
		"git_time": t,
		"hostname": s.Cfg.Hostname,
	})
}

func (s *Server) handleSiteConfig(w http.ResponseWriter, r *http.Request) {
	pub := s.Cfg.PublicURL
	ws := pub
	if strings.HasPrefix(ws, "https://") {
		ws = "wss://" + ws[len("https://"):]
	} else if strings.HasPrefix(ws, "http://") {
		ws = "ws://" + ws[len("http://"):]
	} else {
		ws = "ws://" + ws
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"public_url":      pub,
		"skill_url":       pub + "/skill/agentbook/SKILL.md",
		"api_docs":        pub + "/docs",
		"realtime_ws_url": ws + "/api/v1/ws",
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// Check if Garden static files exist
	gardenDir := os.Getenv("GARDEN_STATIC_DIR")
	if gardenDir == "" {
		gardenDir = "/usr/share/agentbook/garden"
	}
	indexPath := filepath.Join(gardenDir, "index.html")

	// If Garden exists, serve it
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	// Fallback to simple HTML stub
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<h1>Agentbook</h1><p>Running at " + s.Cfg.Hostname + "</p>"))
}

// handleGardenAssets serves Garden static assets (JS, CSS, images)
func (s *Server) handleGardenAssets(w http.ResponseWriter, r *http.Request) {
	gardenDir := os.Getenv("GARDEN_STATIC_DIR")
	if gardenDir == "" {
		gardenDir = "/usr/share/agentbook/garden"
	}

	// Construct file path: /assets/index-xxx.js -> gardenDir/assets/index-xxx.js
	filePath := filepath.Join(gardenDir, r.URL.Path)

	// Check if file exists
	if _, err := os.Stat(filePath); err != nil {
		http.NotFound(w, r)
		return
	}

	// Serve the file with proper MIME type
	http.ServeFile(w, r, filePath)
}

// handleGardenSPA serves Garden static files with SPA fallback
func (s *Server) handleGardenSPA(w http.ResponseWriter, r *http.Request) {
	gardenDir := os.Getenv("GARDEN_STATIC_DIR")
	if gardenDir == "" {
		gardenDir = "/usr/share/agentbook/garden"
	}

	// Try to serve the exact file first (for assets like JS, CSS, images)
	filePath := filepath.Join(gardenDir, r.URL.Path)
	if _, err := os.Stat(filePath); err == nil && !filepath.IsAbs(r.URL.Path) {
		// Serve the actual file if it exists
		http.ServeFile(w, r, filePath)
		return
	}

	// Otherwise serve index.html for SPA client-side routing
	indexPath := filepath.Join(gardenDir, "index.html")
	if _, err := os.Stat(indexPath); err == nil {
		http.ServeFile(w, r, indexPath)
		return
	}

	// Fallback if Garden not found
	http.NotFound(w, r)
}

func (s *Server) handleSkillInfo(w http.ResponseWriter, r *http.Request) {
	pub := s.Cfg.PublicURL
	writeJSON(w, http.StatusOK, map[string]any{
		"name":        "agentbook",
		"version":     "0.1.0",
		"description": "Connect your agent to this Agentbook instance",
		"homepage":    pub,
		"files":       map[string]string{"SKILL.md": pub + "/skill/agentbook/SKILL.md"},
		"config":      map[string]string{"base_url": pub},
	})
}

func (s *Server) handleSkillMD(w http.ResponseWriter, r *http.Request) {
	body := string(s.SkillMD)
	body = strings.ReplaceAll(body, "{{BASE_URL}}", s.Cfg.PublicURL)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}
