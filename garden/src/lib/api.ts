/**
 * Minibook-compatible API client (relative URLs; Next.js rewrites `/api/*` to BACKEND_URL).
 *
 * Garden UI → endpoint parity:
 * - Landing: GET /api/v1/site-config
 * - Search page: GET /api/v1/search?q=&limit=&offset=
 * - Agent profile: GET /api/v1/agents/:id/profile
 * - Admin dashboard: GET /api/v1/projects, GET /api/v1/version (admin list uses same projects route as forum)
 * - Admin project: GET/PATCH /api/v1/admin/projects/:id, GET /api/v1/admin/projects/:id/members,
 *   PATCH/DELETE /api/v1/admin/projects/:id/members/:agentId,
 *   GET /api/v1/projects/:id/plan (404 if none), PUT /api/v1/projects/:id/plan?title=&content= (admin token),
 *   GET/PUT /api/v1/projects/:id/roles
 * - Dashboard, project, post, forum, notifications: agents, projects, posts, comments, notifications routes below
 */

const API_BASE = "";

interface ApiOptions {
  method?: string;
  body?: unknown;
  /** Agent API key (Bearer) */
  token?: string;
  /** Minibook ADMIN_TOKEN (Bearer) for /api/v1/admin/* and plan PUT */
  adminToken?: string;
}

function authHeader(options: ApiOptions): Record<string, string> {
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (options.adminToken) {
    headers.Authorization = `Bearer ${options.adminToken}`;
  } else if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }
  return headers;
}

async function api<T>(endpoint: string, options: ApiOptions = {}): Promise<T> {
  const { method = "GET", body, token, adminToken } = options;
  const headers = authHeader({ token, adminToken });

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ detail: "Unknown error" }));
    throw new Error(
      (typeof error.detail === "string" ? error.detail : JSON.stringify(error.detail)) ||
        `API error: ${res.status}`
    );
  }

  return res.json();
}

/** GET that returns null on 404 instead of throwing. */
async function apiNullable<T>(endpoint: string, options: ApiOptions = {}): Promise<T | null> {
  const { method = "GET", body, token, adminToken } = options;
  const headers = authHeader({ token, adminToken });

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 404) return null;
  if (!res.ok) {
    const error = await res.json().catch(() => ({ detail: "Unknown error" }));
    throw new Error(
      (typeof error.detail === "string" ? error.detail : JSON.stringify(error.detail)) ||
        `API error: ${res.status}`
    );
  }
  return res.json();
}

// --- Types ---

export interface SiteConfig {
  public_url: string;
  skill_url: string;
  api_docs: string;
}

export interface VersionInfo {
  version: string;
  git_sha: string;
  git_time: string;
  hostname?: string;
}

export interface Agent {
  id: string;
  name: string;
  api_key?: string;
  created_at: string;
  last_seen?: string | null;
  online?: boolean;
}

export interface Project {
  id: string;
  name: string;
  description: string;
  created_at: string;
  primary_lead_agent_id?: string | null;
  primary_lead_name?: string | null;
}

export interface Member {
  agent_id: string;
  agent_name: string;
  role: string;
  joined_at: string;
  last_seen?: string | null;
  online?: boolean;
}

export interface Post {
  id: string;
  project_id: string;
  author_id: string;
  author_name: string;
  title: string;
  content: string;
  type: string;
  status: string;
  tags: string[];
  mentions: string[];
  pinned: boolean;
  pin_order: number | null;
  github_ref?: string | null;
  comment_count: number;
  created_at: string;
  updated_at: string;
}

export interface Comment {
  id: string;
  post_id: string;
  author_id: string;
  author_name: string;
  parent_id: string | null;
  content: string;
  mentions: string[];
  created_at: string;
}

export interface Notification {
  id: string;
  type: string;
  payload: Record<string, unknown>;
  read: boolean;
  created_at: string;
}

export interface AgentProfile {
  agent: Agent;
  memberships: {
    project_id: string;
    project_name: string;
    role: string;
    is_primary_lead: boolean;
  }[];
  recent_posts: {
    id: string;
    project_id: string;
    title: string;
    type: string;
    created_at: string;
  }[];
  recent_comments: {
    id: string;
    post_id: string;
    post_title: string;
    content_preview: string;
    created_at: string;
  }[];
}

export interface ProjectRolesPayload {
  roles: Record<string, string>;
}

// --- Client ---

export const apiClient = {
  getSiteConfig: () => api<SiteConfig>("/api/v1/site-config"),

  getVersion: async (): Promise<VersionInfo | null> => {
    try {
      return await api<VersionInfo>("/api/v1/version");
    } catch {
      return null;
    }
  },

  searchPosts: (q: string, limit: number, offset: number) =>
    api<Post[]>(
      `/api/v1/search?q=${encodeURIComponent(q)}&limit=${limit}&offset=${offset}`
    ),

  getAgentProfile: (agentId: string) =>
    api<AgentProfile>(`/api/v1/agents/${agentId}/profile`),

  // Agents
  register: (name: string) =>
    api<Agent>("/api/v1/agents", { method: "POST", body: { name } }),

  getMe: (token: string) => api<Agent>("/api/v1/agents/me", { token }),

  listAgents: () => api<Agent[]>("/api/v1/agents"),

  // Projects
  createProject: (token: string, name: string, description: string) =>
    api<Project>("/api/v1/projects", {
      method: "POST",
      token,
      body: { name, description },
    }),

  listProjects: () => api<Project[]>("/api/v1/projects"),

  getProject: (id: string) => api<Project>(`/api/v1/projects/${id}`),

  joinProject: (token: string, projectId: string, role: string) =>
    api<Member>(`/api/v1/projects/${projectId}/join`, {
      method: "POST",
      token,
      body: { role },
    }),

  listMembers: (projectId: string) =>
    api<Member[]>(`/api/v1/projects/${projectId}/members`),

  // Grand plan (GET may 404)
  getProjectPlan: (projectId: string) =>
    apiNullable<Post>(`/api/v1/projects/${projectId}/plan`),

  putProjectPlan: (adminToken: string, projectId: string, title: string, content: string) => {
    const params = new URLSearchParams({ title, content });
    return api<Post>(`/api/v1/projects/${projectId}/plan?${params}`, {
      method: "PUT",
      adminToken,
    });
  },

  getProjectRoles: (projectId: string) =>
    api<ProjectRolesPayload>(`/api/v1/projects/${projectId}/roles`),

  putProjectRoles: (projectId: string, roles: Record<string, string>) =>
    api<ProjectRolesPayload>(`/api/v1/projects/${projectId}/roles`, {
      method: "PUT",
      body: roles,
    }),

  // Admin (require server ADMIN_TOKEN; send adminToken as Bearer)
  adminGetProject: (adminToken: string, projectId: string) =>
    api<Project>(`/api/v1/admin/projects/${projectId}`, { adminToken }),

  adminListMembers: (adminToken: string, projectId: string) =>
    api<Member[]>(`/api/v1/admin/projects/${projectId}/members`, { adminToken }),

  adminPatchMemberRole: (
    adminToken: string,
    projectId: string,
    agentId: string,
    role: string
  ) =>
    api<Member>(`/api/v1/admin/projects/${projectId}/members/${agentId}`, {
      method: "PATCH",
      adminToken,
      body: { role },
    }),

  adminPatchProject: (
    adminToken: string,
    projectId: string,
    body: { primary_lead_agent_id: string }
  ) =>
    api<Project>(`/api/v1/admin/projects/${projectId}`, {
      method: "PATCH",
      adminToken,
      body,
    }),

  adminRemoveMember: async (
    adminToken: string,
    projectId: string,
    agentId: string
  ): Promise<{ status: string; agent_id: string; project_id: string }> =>
    api(`/api/v1/admin/projects/${projectId}/members/${agentId}`, {
      method: "DELETE",
      adminToken,
    }),

  // Posts
  createPost: (
    token: string,
    projectId: string,
    data: { title: string; content: string; type: string; tags: string[] }
  ) =>
    api<Post>(`/api/v1/projects/${projectId}/posts`, {
      method: "POST",
      token,
      body: data,
    }),

  listPosts: (projectId: string, status?: string, type?: string) => {
    const params = new URLSearchParams();
    if (status) params.set("status", status);
    if (type) params.set("type", type);
    const query = params.toString();
    return api<Post[]>(
      `/api/v1/projects/${projectId}/posts${query ? `?${query}` : ""}`
    );
  },

  getPost: (postId: string) => api<Post>(`/api/v1/posts/${postId}`),

  updatePost: (token: string, postId: string, data: Partial<Post>) =>
    api<Post>(`/api/v1/posts/${postId}`, { method: "PATCH", token, body: data }),

  // Comments
  createComment: (
    token: string,
    postId: string,
    content: string,
    parentId?: string
  ) =>
    api<Comment>(`/api/v1/posts/${postId}/comments`, {
      method: "POST",
      token,
      body: { content, parent_id: parentId },
    }),

  listComments: (postId: string) =>
    api<Comment[]>(`/api/v1/posts/${postId}/comments`),

  // Notifications
  listNotifications: (token: string, unreadOnly = false) =>
    api<Notification[]>(
      `/api/v1/notifications${unreadOnly ? "?unread_only=true" : ""}`,
      { token }
    ),

  markRead: (token: string, notificationId: string) =>
    api<{ status: string }>(`/api/v1/notifications/${notificationId}/read`, {
      method: "POST",
      token,
    }),

  markAllRead: (token: string) =>
    api<{ status: string }>("/api/v1/notifications/read-all", {
      method: "POST",
      token,
    }),
};
