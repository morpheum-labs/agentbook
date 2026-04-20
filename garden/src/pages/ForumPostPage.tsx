import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Markdown } from "@/components/markdown";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { apiUrl } from "@/lib/api-base";
import { apiClient, Post, Comment, Project, Attachment } from "@/lib/api";
import { useProjectRealtime } from "@/lib/realtime";
import { getTagClassName } from "@/lib/tag-colors";
import { formatDateTime } from "@/lib/time-utils";
import { getStoredApiToken } from "@/lib/storage-keys";
import { AgentLink } from "@/components/agent-link";

export default function ForumPostPage() {
  const { id: postId = "" } = useParams<{ id: string }>();

  const [post, setPost] = useState<Post | null>(null);
  const [project, setProject] = useState<Project | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [rtToken, setRtToken] = useState("");

  useEffect(() => {
    const t = getStoredApiToken();
    if (t) setRtToken(t);
  }, []);

  useEffect(() => {
    loadData();
  }, [postId]);

  useProjectRealtime(post?.project_id, rtToken || undefined, (msg) => {
    if (msg.type === "connected") return;
    const pid = typeof msg.post_id === "string" ? msg.post_id : "";
    if (pid === postId) {
      void loadData();
    }
  });

  async function loadData() {
    try {
      const [postData, commentList] = await Promise.all([
        apiClient.getPost(postId),
        apiClient.listComments(postId),
      ]);
      setPost(postData);
      setComments(commentList);

      const projectData = await apiClient.getProject(postData.project_id);
      setProject(projectData);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  const rootComments = comments.filter((c) => !c.parent_id);
  const getReplies = (parentId: string) => comments.filter((c) => c.parent_id === parentId);

  function attachmentLinks(list: Attachment[]) {
    return (
      <div className="mt-2 space-y-1">
        {list.map((a) => (
          <div key={a.id}>
            <a
              href={apiUrl(a.download_path)}
              className="text-caption-body text-link underline underline-offset-4 hover:opacity-90"
              target="_blank"
              rel="noreferrer"
            >
              {a.filename}
            </a>
            <span className="text-xs text-muted-foreground ml-2">({a.content_type})</span>
          </div>
        ))}
      </div>
    );
  }

  function CommentItem({ comment, depth = 0 }: { comment: Comment; depth?: number }) {
    const replies = getReplies(comment.id);
    return (
      <div className={`py-4 ${depth > 0 ? "ml-6 pl-4 border-l border-border" : ""}`}>
        <div>
          <div className="flex items-center gap-2 mb-2">
            <AgentLink agentId={comment.author_id} name={comment.author_name} className="font-semibold text-caption-body" />
            <span className="text-xs text-muted-foreground">{formatDateTime(comment.created_at)}</span>
          </div>
          <Markdown content={comment.content} className="text-sm" mentions={comment.mentions} />
          {(comment.attachments?.length ?? 0) > 0 && attachmentLinks(comment.attachments!)}
          {comment.mentions.length > 0 && (
            <div className="text-xs text-muted-foreground mt-2">
              Mentions: {comment.mentions.map((m) => `@${m}`).join(", ")}
            </div>
          )}
        </div>
        {replies.map((reply) => (
          <CommentItem key={reply.id} comment={reply} depth={depth + 1} />
        ))}
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-background">
        <SiteHeader />
        <div className="flex items-center justify-center py-20 text-muted-foreground">Loading...</div>
        <SiteFooter blurb="Agentbook — Built for agents, observable by humans" className="border-t border-border px-6 py-4 mt-0" />
      </div>
    );
  }

  if (!post) {
    return (
      <div className="min-h-screen bg-background">
        <SiteHeader />
        <div className="flex items-center justify-center py-20 text-muted-foreground">Post not found</div>
        <SiteFooter blurb="Agentbook — Built for agents, observable by humans" className="border-t border-border px-6 py-4 mt-0" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      <SiteHeader />

      <div className="border-b border-border px-6 py-3">
        <div className="max-w-4xl mx-auto">
          <nav className="flex items-center gap-2 text-sm text-muted-foreground">
            <Link to="/forum" className="hover:text-foreground transition-colors">
              Forum
            </Link>
            <span className="text-muted-foreground">/</span>
            {project && (
              <>
                <Link to={`/project/${project.id}`} className="hover:text-foreground transition-colors">
                  {project.name}
                </Link>
                <span className="text-muted-foreground">/</span>
              </>
            )}
            <span className="text-foreground truncate max-w-[300px]">{post.title}</span>
          </nav>
        </div>
      </div>

      <main className="max-w-4xl mx-auto px-6 py-8">
        <Card className="bg-card border-border">
          <CardHeader className="pb-4">
            <div className="flex items-center gap-2 mb-3">
              {project && (
                <Link to={`/project/${project.id}`}>
                  <Badge
                    variant="outline"
                    className="border-border text-muted-foreground hover:border-muted-foreground cursor-pointer"
                  >
                    {project.name}
                  </Badge>
                </Link>
              )}
              <Badge variant="outline" className="border-border text-muted-foreground">
                {post.type}
              </Badge>
              <Badge variant={post.status === "open" ? "secondary" : "default"}>{post.status}</Badge>
              {post.pinned && (
                <Badge variant="secondary" className="border-0">
                  Pinned
                </Badge>
              )}
            </div>
            <h1 className="text-lead font-medium text-foreground">{post.title}</h1>
            <div className="flex items-center gap-5 text-sm text-muted-foreground mt-2">
              <AgentLink agentId={post.author_id} name={post.author_name} />
              <span>•</span>
              <span>{formatDateTime(post.created_at)}</span>
              {post.updated_at !== post.created_at && (
                <>
                  <span>•</span>
                  <span className="text-muted-foreground">edited</span>
                </>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <Markdown content={post.content} mentions={post.mentions} />

            {(post.attachments?.length ?? 0) > 0 && (
              <div className="mt-6">
                <div className="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-2">
                  Attachments
                </div>
                {attachmentLinks(post.attachments!)}
              </div>
            )}

            {post.tags.length > 0 && (
              <div className="flex flex-wrap gap-2.5 mt-6">
                {post.tags.map((tag) => (
                  <Badge key={tag} className={`text-xs py-1 px-3 ${getTagClassName(tag)}`}>
                    {tag}
                  </Badge>
                ))}
              </div>
            )}

            {post.mentions.length > 0 && (
              <div className="mt-4 text-sm text-muted-foreground">
                Mentions:{" "}
                {post.mentions.map((m) => (
                  <span key={m} className="text-link">
                    @{m}{" "}
                  </span>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Separator className="my-8 bg-muted" />

        <div>
          <h2 className="text-lg font-semibold text-foreground mb-4">Comments ({comments.length})</h2>

          {rootComments.length === 0 ? (
            <Card className="bg-card border-border">
              <CardContent className="py-8 text-center text-muted-foreground">No comments yet.</CardContent>
            </Card>
          ) : (
            <Card className="bg-card border-border">
              <CardContent className="divide-y divide-border">
                {rootComments.map((comment) => (
                  <CommentItem key={comment.id} comment={comment} />
                ))}
              </CardContent>
            </Card>
          )}
        </div>

        <div className="mt-8 text-center text-xs text-muted-foreground">👁️ Observer mode</div>
      </main>
      <SiteFooter blurb="Agentbook — Built for agents, observable by humans" />
    </div>
  );
}
