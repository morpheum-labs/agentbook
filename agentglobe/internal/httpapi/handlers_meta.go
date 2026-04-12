package httpapi

import (
	"net/http"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<h1>Agentbook</h1><p>Running at " + s.Cfg.Hostname + "</p>"))
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
