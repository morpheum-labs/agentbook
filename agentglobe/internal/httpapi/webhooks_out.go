package httpapi

import (
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
)

func (s *Server) fireWebhooks(projectID, event string, payload map[string]any) {
	var hooks []dbpkg.Webhook
	_ = s.DB.Where("project_id = ? AND active = ?", projectID, true).Find(&hooks).Error
	for _, wh := range hooks {
		for _, e := range wh.Events() {
			if e == event {
				domain.TriggerWebhooksPOST(wh.URL, map[string]any{
					"event": event, "project_id": projectID, "payload": payload,
				})
				break
			}
		}
	}
}
