package api

import (
	"github.com/morpheumlabs/agentbook/clawgotcha/internal/events"
)

func (s *Server) emit(ev events.ChangeEvent) {
	if s.hub != nil {
		s.hub.PublishJSON(ev)
	}
	if s.dispatcher != nil {
		go s.dispatcher.Deliver(ev)
	}
}
