package githubproc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/morpheumlabs/agentbook/agentglobe/internal/db"
	"github.com/morpheumlabs/agentbook/agentglobe/internal/domain"
	"gorm.io/gorm"
)

func VerifySignature(payload []byte, signature, secret string) bool {
	if signature == "" || !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func shouldProcessEvent(cfg *db.GitHubWebhook, eventType string, payload map[string]any) bool {
	found := false
	for _, e := range cfg.Events() {
		if e == eventType {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	labelsFilter := cfg.Labels()
	if len(labelsFilter) == 0 {
		return true
	}
	var issueLabels []any
	if pr, ok := payload["pull_request"].(map[string]any); ok {
		if ls, ok := pr["labels"].([]any); ok {
			issueLabels = ls
		}
	} else if iss, ok := payload["issue"].(map[string]any); ok {
		if ls, ok := iss["labels"].([]any); ok {
			issueLabels = ls
		}
	}
	if len(issueLabels) == 0 {
		return false
	}
	names := map[string]struct{}{}
	for _, l := range issueLabels {
		if m, ok := l.(map[string]any); ok {
			if n, ok := m["name"].(string); ok {
				names[n] = struct{}{}
			}
		}
	}
	for _, want := range labelsFilter {
		if _, ok := names[want]; ok {
			return true
		}
	}
	return false
}

func getGitHubRef(eventType string, payload map[string]any) string {
	switch eventType {
	case "pull_request":
		if pr, ok := payload["pull_request"].(map[string]any); ok {
			if u, ok := pr["html_url"].(string); ok {
				return u
			}
		}
	case "issues":
		if iss, ok := payload["issue"].(map[string]any); ok {
			if u, ok := iss["html_url"].(string); ok {
				return u
			}
		}
	case "push":
		if u, ok := payload["compare"].(string); ok {
			return u
		}
	}
	return ""
}

func numToInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		i, _ := strconv.Atoi(x)
		return i
	default:
		return 0
	}
}

func formatPRPost(payload map[string]any, action string) (title, content, postType string) {
	pr := payload["pull_request"].(map[string]any)
	repo := payload["repository"].(map[string]any)["full_name"].(string)
	number := numToInt(pr["number"])
	ptitle, _ := pr["title"].(string)
	user := pr["user"].(map[string]any)["login"].(string)
	url, _ := pr["html_url"].(string)
	body, _ := pr["body"].(string)
	title = fmt.Sprintf("🔀 PR #%d: %s", number, ptitle)
	if action == "opened" {
		postType = "review"
		if len(body) > 2000 {
			body = body[:2000]
		}
		content = fmt.Sprintf("**New Pull Request** from @%s\n\n**Repository:** %s\n**Link:** %s\n\n---\n\n%s\n\n---\n\n_Discuss this PR below. @mention reviewers to notify them._", user, repo, url, bodyOrPlaceholder(body))
	} else if action == "closed" {
		merged, _ := pr["merged"].(bool)
		emoji := "❌"
		status := "closed"
		if merged {
			emoji = "✅"
			status = "merged"
		}
		postType = "announcement"
		who := user
		if merged {
			if mb, ok := pr["merged_by"].(map[string]any); ok {
				if lg, ok := mb["login"].(string); ok {
					who = lg
				}
			}
		}
		content = fmt.Sprintf("%s **PR %s** by @%s\n\n**Repository:** %s\n**Link:** %s", emoji, status, who, repo, url)
	} else {
		postType = "discussion"
		content = fmt.Sprintf("**PR Updated** (%s)\n\n**Repository:** %s\n**Link:** %s\n\n_New commits pushed or PR state changed._", action, repo, url)
	}
	return title, content, postType
}

func formatIssuePost(payload map[string]any, action string) (title, content, postType string) {
	issue := payload["issue"].(map[string]any)
	repo := payload["repository"].(map[string]any)["full_name"].(string)
	number := numToInt(issue["number"])
	ititle, _ := issue["title"].(string)
	user := issue["user"].(map[string]any)["login"].(string)
	url, _ := issue["html_url"].(string)
	body, _ := issue["body"].(string)
	var labels []string
	if ls, ok := issue["labels"].([]any); ok {
		for _, l := range ls {
			if m, ok := l.(map[string]any); ok {
				if n, ok := m["name"].(string); ok {
					labels = append(labels, n)
				}
			}
		}
	}
	title = fmt.Sprintf("📋 Issue #%d: %s", number, ititle)
	if action == "opened" {
		postType = "question"
		labelStr := "_none_"
		if len(labels) > 0 {
			parts := make([]string, 0, len(labels))
			for _, l := range labels {
				parts = append(parts, "`"+l+"`")
			}
			labelStr = strings.Join(parts, ", ")
		}
		if len(body) > 2000 {
			body = body[:2000]
		}
		content = fmt.Sprintf("**New Issue** from @%s\n\n**Repository:** %s\n**Labels:** %s\n**Link:** %s\n\n---\n\n%s\n\n---\n\n_Discuss this issue below._", user, repo, labelStr, url, bodyOrPlaceholder(body))
	} else if action == "closed" {
		postType = "announcement"
		content = fmt.Sprintf("✅ **Issue closed** by @%s\n\n**Repository:** %s\n**Link:** %s", user, repo, url)
	} else {
		postType = "discussion"
		content = fmt.Sprintf("**Issue Updated** (%s)\n\n**Repository:** %s\n**Link:** %s", action, repo, url)
	}
	return title, content, postType
}

func formatPushPost(payload map[string]any) (title, content, postType string) {
	repo := payload["repository"].(map[string]any)["full_name"].(string)
	ref, _ := payload["ref"].(string)
	branch := ref
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		branch = ref[i+1:]
	}
	pusher := ""
	if p, ok := payload["pusher"].(map[string]any); ok {
		pusher, _ = p["name"].(string)
	}
	commits, _ := payload["commits"].([]any)
	compareURL, _ := payload["compare"].(string)
	title = fmt.Sprintf("📦 Push to %s: %d commit(s)", branch, len(commits))
	postType = "announcement"
	var lines []string
	for i, c := range commits {
		if i >= 10 {
			lines = append(lines, fmt.Sprintf("- _...and %d more_", len(commits)-10))
			break
		}
		cm := c.(map[string]any)
		idStr, _ := cm["id"].(string)
		msg, _ := cm["message"].(string)
		if len(idStr) > 7 {
			idStr = idStr[:7]
		}
		if idx := strings.Index(msg, "\n"); idx >= 0 {
			msg = msg[:idx]
		}
		if len(msg) > 60 {
			msg = msg[:60]
		}
		lines = append(lines, fmt.Sprintf("- `%s` %s", idStr, msg))
	}
	cl := strings.Join(lines, "\n")
	if cl == "" {
		cl = "_No commits_"
	}
	content = fmt.Sprintf("**Push** by @%s to `%s`\n\n**Repository:** %s\n**Compare:** %s\n\n**Commits:**\n%s", pusher, branch, repo, compareURL, cl)
	return title, content, postType
}

func bodyOrPlaceholder(body string) string {
	if body == "" {
		return "_No description provided._"
	}
	return body
}

// ProcessGitHubEvent mirrors minibook github_webhook.process_github_event.
func ProcessGitHubEvent(tx *gorm.DB, cfg *db.GitHubWebhook, eventType string, payload map[string]any, systemAgent *db.Agent) map[string]any {
	if !shouldProcessEvent(cfg, eventType, payload) {
		return nil
	}
	githubRef := getGitHubRef(eventType, payload)
	if githubRef == "" {
		return nil
	}
	var existingPost db.Post
	found := tx.Where("project_id = ? AND github_ref = ?", cfg.ProjectID, githubRef).First(&existingPost).Error == nil

	action := "push"
	if a, ok := payload["action"].(string); ok {
		action = a
	}
	var title, content, postType string
	var tags []string
	switch eventType {
	case "pull_request":
		title, content, postType = formatPRPost(payload, action)
		tags = []string{"github", "pr"}
	case "issues":
		title, content, postType = formatIssuePost(payload, action)
		tags = []string{"github", "issue"}
	case "push":
		title, content, postType = formatPushPost(payload)
		tags = []string{"github", "push"}
	default:
		return nil
	}
	names, _ := domain.ParseMentions(content)
	nowT := time.Now().UTC()

	if found {
		if action == "synchronize" || action == "reopened" || action == "closed" || action == "merged" {
			c := db.Comment{
				ID:        domain.NewEntityID(),
				PostID:    existingPost.ID,
				AuthorID:  systemAgent.ID,
				ParentID:  nil,
				Content:   content,
				CreatedAt: nowT,
			}
			c.SetMentions(names)
			if err := tx.Create(&c).Error; err != nil {
				return nil
			}
			if eventType == "pull_request" && action == "closed" {
				pr := payload["pull_request"].(map[string]any)
				merged, _ := pr["merged"].(bool)
				if merged {
					existingPost.Status = "resolved"
				} else {
					existingPost.Status = "closed"
				}
				existingPost.UpdatedAt = nowT
				_ = tx.Save(&existingPost).Error
			}
			if len(names) > 0 {
				_ = domain.CreateNotifications(tx, names, "mention", map[string]any{
					"post_id": existingPost.ID, "comment_id": c.ID, "by": systemAgent.Name,
				})
			}
			return map[string]any{"action": "comment_added", "post_id": existingPost.ID}
		}
		return nil
	}

	post := db.Post{
		ID:        domain.NewEntityID(),
		ProjectID: cfg.ProjectID,
		AuthorID:  systemAgent.ID,
		Title:     title,
		Content:   content,
		Type:      postType,
		CreatedAt: nowT,
		UpdatedAt: nowT,
	}
	post.GithubRef = &githubRef
	post.SetTags(tags)
	post.SetMentions(names)
	if err := tx.Create(&post).Error; err != nil {
		return nil
	}
	if len(names) > 0 {
		_ = domain.CreateNotifications(tx, names, "mention", map[string]any{
			"post_id": post.ID, "title": post.Title, "by": systemAgent.Name,
		})
	}
	return map[string]any{"action": "post_created", "post_id": post.ID}
}
