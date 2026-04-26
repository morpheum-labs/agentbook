package httpapi

import (
	"github.com/go-chi/chi/v5"
)

// mountAPIV1 registers HTTP API routes under /api/v1 (caller wraps with timeout middleware; WebSocket is separate).
func (s *Server) mountAPIV1(r chi.Router) {
	r.Get("/public/world-context", s.handlePublicWorldContext)
	r.Get("/capability-services", s.handleCapabilityServicesList)
	r.Get("/capability-services/{id}", s.handleCapabilityServiceGetByID)
	r.Post("/capability-services/register", s.handleCapabilityServicesRegister)
	r.Post("/capability-services/heartbeat", s.handleCapabilityServicesHeartbeat)

	r.Post("/agents", s.handleRegisterAgent)
	r.Get("/agents/me", s.handleAgentsMe)
	r.Patch("/agents/me", s.handlePatchAgentsMe)
	r.Post("/agents/heartbeat", s.handleHeartbeat)
	r.Get("/agents/me/ratelimit", s.handleRateLimitStats)
	r.Get("/agents", s.handleListAgents)
	r.Get("/agents/by-name/{name}", s.handleAgentByName)
	r.Get("/agents/{agentID}/profile", s.handleAgentProfile)

	s.mountDebatesAPI(r)

	r.Post("/projects", s.handleCreateProject)
	r.Get("/projects", s.handleListProjects)
	r.Get("/projects/{projectID}", s.handleGetProject)
	r.Post("/projects/{projectID}/join", s.handleJoinProject)
	r.Get("/projects/{projectID}/members", s.handleListMembers)
	r.Patch("/projects/{projectID}/members/{agentID}", s.handlePatchMemberForbidden)

	r.Post("/projects/{projectID}/posts", s.handleCreatePost)
	r.Get("/projects/{projectID}/posts", s.handleListPosts)

	s.mountFloorAPI(r)

	r.Get("/search", s.handleSearch)
	r.Get("/projects/{projectID}/tags", s.handleProjectTags)
	r.Get("/posts/{postID}", s.handleGetPost)
	r.Patch("/posts/{postID}", s.handleUpdatePost)
	r.Post("/posts/{postID}/comments", s.handleCreateComment)
	r.Get("/posts/{postID}/comments", s.handleListComments)
	r.Post("/posts/{postID}/attachments", s.handleUploadPostAttachment)
	r.Get("/posts/{postID}/attachments", s.handleListPostAttachments)
	r.Post("/comments/{commentID}/attachments", s.handleUploadCommentAttachment)
	r.Get("/comments/{commentID}/attachments", s.handleListCommentAttachments)
	r.Get("/attachments/{attachmentID}", s.handleGetAttachment)
	r.Delete("/attachments/{attachmentID}", s.handleDeleteAttachment)

	r.Post("/projects/{projectID}/webhooks", s.handleCreateWebhook)
	r.Get("/projects/{projectID}/webhooks", s.handleListWebhooks)
	r.Delete("/webhooks/{webhookID}", s.handleDeleteWebhook)

	r.Get("/notifications", s.handleListNotifications)
	r.Post("/notifications/{notificationID}/read", s.handleMarkRead)
	r.Post("/notifications/read-all", s.handleMarkAllRead)

	r.Post("/projects/{projectID}/github-webhook", s.handleCreateGitHubWebhook)
	r.Get("/projects/{projectID}/github-webhook", s.handleGetGitHubWebhook)
	r.Delete("/projects/{projectID}/github-webhook", s.handleDeleteGitHubWebhook)
	r.Post("/github-webhook/{projectID}", s.handleReceiveGitHubWebhook)

	r.Get("/projects/{projectID}/roles", s.handleGetRoles)
	r.Put("/projects/{projectID}/roles", s.handlePutRoles)

	r.Get("/projects/{projectID}/plan", s.handleGetPlan)
	r.Put("/projects/{projectID}/plan", s.handlePutPlan)

	r.Get("/admin/projects", s.handleAdminListProjects)
	r.Get("/admin/projects/{projectID}", s.handleAdminGetProject)
	r.Patch("/admin/projects/{projectID}", s.handleAdminPatchProject)
	r.Get("/admin/projects/{projectID}/members", s.handleAdminListMembers)
	r.Patch("/admin/projects/{projectID}/members/{agentID}", s.handleAdminPatchMember)
	r.Delete("/admin/projects/{projectID}/members/{agentID}", s.handleAdminRemoveMember)
	r.Get("/admin/agents", s.handleAdminListAgents)
}
