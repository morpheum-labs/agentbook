package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/morpheumlabs/agentbook/clawgotcha/internal/db"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/httperr"
)

type publishEventBody struct {
	EventType          string   `json:"event_type"`
	AffectedEntityType string   `json:"affected_entity_type"`
	AffectedIDs        []string `json:"affected_ids"`
	NewRevision        int64    `json:"new_revision"`
}

func (s *Server) publishEvent(w http.ResponseWriter, r *http.Request) {
	var b publishEventBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httperr.Write(w, r, httperr.BadRequest("invalid json body", err))
		return
	}
	if b.EventType == "" || b.AffectedEntityType == "" {
		httperr.Write(w, r, httperr.BadRequest("event_type and affected_entity_type required", nil))
		return
	}
	ev := events.ChangeEvent{
		EventType:          b.EventType,
		AffectedEntityType: b.AffectedEntityType,
		AffectedIDs:        b.AffectedIDs,
		NewRevision:        b.NewRevision,
		TS:                 events.NowRFC3339Nano(),
	}
	s.emit(ev)
	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request) {
	sum, err := db.LoadRevisionSummary(s.db)
	if err != nil {
		httperr.Write(w, r, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	writeSSE(w, "revision_summary", sum)
	flusher.Flush()

	if s.hub == nil {
		return
	}
	ch := s.hub.Subscribe(r.Context(), 16)
	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "event: change\ndata: %s\n\n", string(msg))
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, event string, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(b))
}
