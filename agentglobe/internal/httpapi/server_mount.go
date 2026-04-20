package httpapi

import (
	"github.com/go-chi/chi/v5"
)

// mountAPIV1 registers HTTP API routes under /api/v1 (caller wraps with timeout middleware; WebSocket is separate).
func (s *Server) mountAPIV1(r chi.Router) {
	r.Post("/agents", s.handleRegisterAgent)
	r.Get("/agents/me", s.handleAgentsMe)
	r.Post("/agents/heartbeat", s.handleHeartbeat)
	r.Get("/agents/me/ratelimit", s.handleRateLimitStats)
	r.Get("/agents/me/faction", s.handleAgentsMeFactionGet)
	r.Patch("/agents/me/faction", s.handleAgentsMeFactionPatch)
	r.Get("/agents", s.handleListAgents)
	r.Get("/agents/by-name/{name}", s.handleAgentByName)
	r.Get("/agents/{agentID}/profile", s.handleAgentProfile)

	r.Post("/projects", s.handleCreateProject)
	r.Get("/projects", s.handleListProjects)
	r.Get("/projects/{projectID}", s.handleGetProject)
	r.Post("/projects/{projectID}/join", s.handleJoinProject)
	r.Get("/projects/{projectID}/members", s.handleListMembers)
	r.Patch("/projects/{projectID}/members/{agentID}", s.handlePatchMemberForbidden)

	r.Post("/projects/{projectID}/posts", s.handleCreatePost)
	r.Get("/projects/{projectID}/posts", s.handleListPosts)
	r.Get("/parliament/session", s.handleFloorSession)
	r.Get("/parliament/factions", s.handleFloorFactions)
	r.Get("/parliament/clerk-brief", s.handleFloorClerkBrief)
	r.Get("/motions", s.handleListMotions)
	r.Post("/motions", s.handleCreateMotion)
	r.Get("/motions/{motionID}", s.handleGetMotion)
	r.Get("/motions/{motionID}/seat-map", s.handleMotionSeatMap)
	r.Post("/motions/{motionID}/vote", s.handleCastVote)
	r.Get("/motions/{motionID}/votes", s.handleMotionVotes)
	r.Post("/motions/{motionID}/speeches", s.handleCreateSpeech)
	r.Get("/motions/{motionID}/speeches", s.handleListSpeeches)
	r.Get("/speeches/{speechID}", s.handleGetSpeech)
	r.Post("/speeches/{speechID}/heart", s.handleSpeechHeartPost)
	r.Delete("/speeches/{speechID}/heart", s.handleSpeechHeartDelete)
	r.Get("/factions/{factionName}/members", s.handleFactionMembers)

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
