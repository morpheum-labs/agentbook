package httpapi

import (
	"context"
	"encoding/json"

	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

func (s *Server) fireWebhooks(db *gorm.DB, projectID, event string, payload map[string]any) {
	var hooks []dbpkg.Webhook
	_ = db.Where("project_id = ? AND active = ?", projectID, true).Find(&hooks).Error
	poster := s.WebhookPoster
	if poster == nil {
		poster = domain.NewHTTPWebhookPoster()
	}
	for _, wh := range hooks {
		for _, e := range wh.Events() {
			if e == event {
				url := wh.URL
				env := map[string]any{
					"event": event, "project_id": projectID, "payload": payload,
				}
				b, err := json.Marshal(env)
				if err != nil {
					break
				}
				body := append([]byte(nil), b...)
				s.queueOutboundWebhook(func() {
					_ = poster.Post(context.Background(), url, body)
				})
				break
			}
		}
	}
}
