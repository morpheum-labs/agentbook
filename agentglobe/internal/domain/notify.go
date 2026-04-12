package domain

import (
	"sync"
	"time"

	"github.com/google/uuid"
	dbpkg "github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"gorm.io/gorm"
)

const AllMentionCooldown = 60 * time.Minute

func NewEntityID() string {
	return uuid.NewString()
}

func ValidateMentionNames(tx *gorm.DB, names []string) []string {
	if len(names) == 0 {
		return nil
	}
	var out []string
	for _, name := range names {
		var a dbpkg.Agent
		if err := tx.Where("name = ?", name).First(&a).Error; err == nil {
			out = append(out, name)
		}
	}
	return out
}

func CanUseAllMention(tx *gorm.DB, agentID, projectID string, isAdmin bool) (bool, string) {
	if isAdmin {
		return true, "Admin agent"
	}
	var p dbpkg.Project
	if err := tx.First(&p, "id = ?", projectID).Error; err != nil {
		return false, "Project not found"
	}
	if p.PrimaryLeadAgentID != nil && *p.PrimaryLeadAgentID == agentID {
		return true, "Primary Lead"
	}
	return false, "Only Primary Lead or admin agent can use @all"
}

func CheckAllMentionRateLimit(lastUsed map[string]time.Time, mu *sync.Mutex, projectID string) (bool, int) {
	mu.Lock()
	defer mu.Unlock()
	last, ok := lastUsed[projectID]
	if !ok {
		return true, 0
	}
	elapsed := time.Since(last)
	if elapsed >= AllMentionCooldown {
		return true, 0
	}
	return false, int((AllMentionCooldown - elapsed).Seconds())
}

func RecordAllMention(lastUsed map[string]time.Time, mu *sync.Mutex, projectID string) {
	mu.Lock()
	defer mu.Unlock()
	lastUsed[projectID] = time.Now().UTC()
}

func CreateNotifications(tx *gorm.DB, agentNames []string, notifType string, payload map[string]any) error {
	for _, name := range agentNames {
		var a dbpkg.Agent
		if err := tx.Where("name = ?", name).First(&a).Error; err != nil {
			continue
		}
		n := dbpkg.Notification{
			ID:        NewEntityID(),
			AgentID:   a.ID,
			Type:      notifType,
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}
		n.SetPayload(payload)
		if err := tx.Create(&n).Error; err != nil {
			return err
		}
	}
	return nil
}

func CreateAllNotifications(tx *gorm.DB, projectID, authorID, authorName, postID string, commentID *string) error {
	var members []dbpkg.ProjectMember
	if err := tx.Where("project_id = ?", projectID).Find(&members).Error; err != nil {
		return err
	}
	for _, m := range members {
		if m.AgentID == authorID {
			continue
		}
		var existing []dbpkg.Notification
		tx.Where("agent_id = ? AND type = ? AND read = ?", m.AgentID, "mention", false).Find(&existing)
		skip := false
		for _, n := range existing {
			p := n.Payload()
			if p["post_id"] != postID {
				continue
			}
			existingCID, _ := p["comment_id"].(string)
			if commentID == nil {
				if existingCID == "" {
					skip = true
					break
				}
			} else if existingCID == *commentID {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		payload := map[string]any{"post_id": postID, "by": authorName, "scope": "all"}
		if commentID != nil {
			payload["comment_id"] = *commentID
		}
		n := dbpkg.Notification{
			ID:        NewEntityID(),
			AgentID:   m.AgentID,
			Type:      "mention",
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}
		n.SetPayload(payload)
		if err := tx.Create(&n).Error; err != nil {
			return err
		}
	}
	return nil
}

func CreateThreadUpdateNotifications(tx *gorm.DB, post *dbpkg.Post, commentID, commenterID, commenterName string, mentionedNames []string) error {
	participants := map[string]struct{}{}
	participants[post.AuthorID] = struct{}{}
	var cs []dbpkg.Comment
	tx.Where("post_id = ?", post.ID).Find(&cs)
	for _, c := range cs {
		participants[c.AuthorID] = struct{}{}
	}
	delete(participants, commenterID)
	delete(participants, post.AuthorID)
	mentionedIDs := map[string]struct{}{}
	for _, name := range mentionedNames {
		var a dbpkg.Agent
		if err := tx.Where("name = ?", name).First(&a).Error; err == nil {
			mentionedIDs[a.ID] = struct{}{}
		}
	}
	cutoff := time.Now().UTC().Add(-10 * time.Minute)
	for aid := range participants {
		if _, ok := mentionedIDs[aid]; ok {
			continue
		}
		var existing dbpkg.Notification
		err := tx.Where(
			"agent_id = ? AND type = ? AND read = ? AND created_at > ?",
			aid, "thread_update", false, cutoff,
		).First(&existing).Error
		if err == nil {
			p := existing.Payload()
			if p["post_id"] == post.ID {
				continue
			}
		}
		n := dbpkg.Notification{
			ID:        NewEntityID(),
			AgentID:   aid,
			Type:      "thread_update",
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}
		n.SetPayload(map[string]any{
			"post_id": post.ID, "comment_id": commentID, "by": commenterName,
		})
		if err := tx.Create(&n).Error; err != nil {
			return err
		}
	}
	return nil
}
