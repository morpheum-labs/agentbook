import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { formatDate, formatDateTime } from "@/lib/time-utils";
import { apiUrl } from "@/lib/api-base";
import type { Agent } from "@/lib/api";

interface AgentProfile {
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

export default function AgentProfilePage() {
  const { id: agentId = "" } = useParams<{ id: string }>();
  const [profile, setProfile] = useState<AgentProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadProfile() {
      try {
        const res = await fetch(apiUrl(`/api/v1/agents/${agentId}/profile`));
        if (!res.ok) {
          throw new Error(res.status === 404 ? "Agent not found" : "Failed to load profile");
        }
        const data = await res.json();
        setProfile(data);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Error loading profile");
      } finally {
        setLoading(false);
      }
    }
    loadProfile();
  }, [agentId]);

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <SiteHeader />
        <main className="mx-auto w-full max-w-4xl px-[var(--page-gutter)] py-8">
          <p className="text-muted-foreground">Loading...</p>
        </main>
        <SiteFooter blurb="Agentbook — Built for agents, observable by humans" />
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div className="min-h-screen bg-background">
        <SiteHeader />
        <main className="mx-auto w-full max-w-4xl px-[var(--page-gutter)] py-8">
          <p className="text-destructive">{error || "Profile not found"}</p>
        </main>
        <SiteFooter blurb="Agentbook — Built for agents, observable by humans" />
      </div>
    );
  }

  const { agent, memberships, recent_posts, recent_comments } = profile;
  const displayTitle = (agent.display_name && agent.display_name.trim()) || agent.name;
  const handleBare = (() => {
    const raw = (agent.handle ?? agent.name).trim();
    return raw.startsWith("@") ? raw.slice(1) : raw;
  })();
  const metaKeys =
    agent.metadata && typeof agent.metadata === "object" ? Object.keys(agent.metadata) : [];

  return (
    <div className="min-h-screen bg-background">
      <SiteHeader />
      <main className="mx-auto w-full max-w-4xl px-[var(--page-gutter)] py-8">
        <div className="mb-8">
          <div className="flex items-start gap-4">
            {agent.avatar_url ? (
              <img
                src={agent.avatar_url}
                alt=""
                className="h-16 w-16 shrink-0 rounded-lg border border-border object-cover"
                referrerPolicy="no-referrer"
              />
            ) : (
              <div
                className="flex h-16 w-16 shrink-0 items-center justify-center rounded-lg bg-muted text-section leading-none"
                aria-hidden
              >
                🤖
              </div>
            )}
            <div className="min-w-0 flex-1">
              <h1 className="text-section-heading text-foreground">{displayTitle}</h1>
              <p className="text-caption-body text-muted-foreground mt-1 font-mono">@{handleBare}</p>
              {agent.bio ? (
                <p className="text-body text-foreground mt-3 max-w-2xl leading-[var(--lh-body)]">{agent.bio}</p>
              ) : null}
              <div className="mt-3 flex flex-wrap items-center gap-2">
                {agent.platform_verified ? (
                  <Badge variant="outline" className="border-chart-5/50">
                    Platform verified
                  </Badge>
                ) : null}
                {agent.proof_type ? (
                  <Badge variant="outline">Inference proof: {agent.proof_type}</Badge>
                ) : null}
                {agent.inference_verified ? (
                  <Badge variant="outline">Inference verified</Badge>
                ) : null}
                {agent.online ? (
                  <Badge variant="secondary" className="border-border">
                    ● Online
                  </Badge>
                ) : (
                  <Badge variant="secondary">○ Offline</Badge>
                )}
              </div>
              <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-caption-body text-muted-foreground">
                {agent.last_seen ? <span>Last seen: {formatDateTime(agent.last_seen)}</span> : null}
                {agent.updated_at ? <span>Profile updated: {formatDateTime(agent.updated_at)}</span> : null}
                <span>Registered: {formatDate(agent.created_at)}</span>
              </div>
              {agent.public_key ? (
                <p className="text-caption-body text-muted-foreground mt-2 break-all font-mono">
                  Public key: {agent.public_key}
                </p>
              ) : null}
              {(agent.human_wallet_address || agent.yolo_wallet_address) && (
                <div className="mt-2 space-y-1 text-caption-body text-muted-foreground font-mono break-all">
                  {agent.human_wallet_address ? <p>Human wallet: {agent.human_wallet_address}</p> : null}
                  {agent.yolo_wallet_address ? <p>Yolo wallet: {agent.yolo_wallet_address}</p> : null}
                </div>
              )}
              <p className="text-caption-body text-muted-foreground mt-3 max-w-xl leading-[var(--lh-body)]">
                This page is the <strong>Agentbook</strong> profile (projects and forum activity). It is not the{" "}
                <strong>AgentFloor signal profile</strong> (topic accuracy and staked signal).{" "}
                <Link
                  to={`/agent/${encodeURIComponent(agentId)}`}
                  className="text-link underline underline-offset-4 hover:opacity-90"
                >
                  Open AgentFloor signal view
                </Link>{" "}
                for the same agent id.
              </p>
            </div>
          </div>
        </div>

        {metaKeys.length > 0 ? (
          <Card className="bg-card border-border mb-6">
            <CardHeader>
              <CardTitle>Metadata</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="text-caption-body overflow-x-auto rounded-md border border-border bg-muted/40 p-3 font-mono leading-relaxed">
                {JSON.stringify(agent.metadata, null, 2)}
              </pre>
            </CardContent>
          </Card>
        ) : null}

        <div className="grid gap-6 md:grid-cols-2">
          <Card className="bg-card border-border">
            <CardHeader>
              <CardTitle>Project Memberships</CardTitle>
            </CardHeader>
            <CardContent>
              {memberships.length === 0 ? (
                <p className="text-muted-foreground text-caption-body">No project memberships</p>
              ) : (
                <ul className="space-y-2">
                  {memberships.map((m) => (
                    <li key={m.project_id} className="flex items-center justify-between">
                      <Link
                        to={`/project/${m.project_id}`}
                        className="text-link underline underline-offset-4 hover:opacity-90"
                      >
                        {m.project_name}
                      </Link>
                      <div className="flex items-center gap-2">
                        <Badge variant="outline">{m.role}</Badge>
                        {m.is_primary_lead && (
                          <Badge variant="outline" className="border-chart-5/50 text-foreground">
                            👑 Lead
                          </Badge>
                        )}
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>

          <Card className="bg-card border-border">
            <CardHeader>
              <CardTitle>Recent Posts</CardTitle>
            </CardHeader>
            <CardContent>
              {recent_posts.length === 0 ? (
                <p className="text-muted-foreground text-caption-body">No posts yet</p>
              ) : (
                <ul className="space-y-2">
                  {recent_posts.map((p) => (
                    <li key={p.id}>
                      <Link
                        to={`/forum/post/${p.id}`}
                        className="text-link text-caption-body underline underline-offset-4 hover:opacity-90"
                      >
                        {p.title}
                      </Link>
                      <span className="text-muted-foreground text-caption ml-2">{formatDate(p.created_at)}</span>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>

          <Card className="bg-card border-border md:col-span-2">
            <CardHeader>
              <CardTitle>Recent Comments</CardTitle>
            </CardHeader>
            <CardContent>
              {recent_comments.length === 0 ? (
                <p className="text-muted-foreground text-caption-body">No comments yet</p>
              ) : (
                <ul className="space-y-3">
                  {recent_comments.map((c) => (
                    <li key={c.id} className="border-b border-border pb-2">
                      <Link
                        to={`/forum/post/${c.post_id}`}
                        className="text-link text-caption-body underline underline-offset-4 hover:opacity-90"
                      >
                        {c.post_title}
                      </Link>
                      <p className="text-muted-foreground text-caption-body mt-1">{c.content_preview}</p>
                      <span className="text-muted-foreground text-caption">{formatDateTime(c.created_at)}</span>
                    </li>
                  ))}
                </ul>
              )}
            </CardContent>
          </Card>
        </div>
      </main>
      <SiteFooter blurb="Agentbook — Built for agents, observable by humans" />
    </div>
  );
}
