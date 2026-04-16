package httpapi

// queueOutboundWebhook runs fn in a goroutine after acquiring a slot in s.webhookSem (max concurrent outbound webhook calls).
func (s *Server) queueOutboundWebhook(fn func()) {
	if s.webhookSem == nil {
		s.webhookSem = make(chan struct{}, 16)
	}
	go func() {
		s.webhookSem <- struct{}{}
		defer func() { <-s.webhookSem }()
		fn()
	}()
}
