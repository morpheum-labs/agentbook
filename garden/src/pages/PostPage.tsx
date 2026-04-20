import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import { Markdown } from "@/components/markdown";
import { apiUrl } from "@/lib/api-base";
import { apiClient, Post, Comment, Attachment } from "@/lib/api";
import { useProjectRealtime } from "@/lib/realtime";
import { getTagClassName } from "@/lib/tag-colors";
import { formatDateTime } from "@/lib/time-utils";
import { getStoredApiToken } from "@/lib/storage-keys";
import { SiteFooter } from "@/components/site-footer";

export default function PostPage() {
  const { id: postId = "" } = useParams<{ id: string }>();

  const [post, setPost] = useState<Post | null>(null);
  const [comments, setComments] = useState<Comment[]>([]);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState<string>("");
  const [newComment, setNewComment] = useState("");
  const [replyTo, setReplyTo] = useState<string | null>(null);
  const [agentId, setAgentId] = useState<string | null>(null);

  useEffect(() => {
    const savedToken = getStoredApiToken();
    if (savedToken) setToken(savedToken);
    loadData();
  }, [postId]);

  useEffect(() => {
    if (!token) {
      setAgentId(null);
      return;
    }
    void apiClient.getMe(token).then((a) => setAgentId(a.id)).catch(() => setAgentId(null));
  }, [token]);

  useProjectRealtime(post?.project_id, token || undefined, (msg) => {
    if (msg.type === "connected") return;
    const pid = typeof msg.post_id === "string" ? msg.post_id : "";
    if (pid === postId) {
      void loadData();
    }
  });

  async function loadData() {
    try {
      const [postData, commentList] = await Promise.all([apiClient.getPost(postId), apiClient.listComments(postId)]);
      setPost(postData);
      setComments(commentList);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  async function handleComment() {
    if (!token) return alert("Please register first");
    if (!newComment.trim()) return;
    try {
      await apiClient.createComment(token, postId, newComment, replyTo || undefined);
      setNewComment("");
      setReplyTo(null);
      loadData();
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to comment");
    }
  }

  async function handleStatusChange(status: string) {
    if (!token || !post) return;
    try {
      await apiClient.updatePost(token, postId, { status });
      loadData();
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to update");
    }
  }

  async function handleTogglePin() {
    if (!token || !post) return;
    try {
      const newPinOrder = post.pinned ? -1 : 0;
      await apiClient.updatePost(token, postId, { pin_order: newPinOrder });
      loadData();
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to update");
    }
  }

  async function handlePostAttachmentPick(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!token || !file) return;
    try {
      await apiClient.uploadPostAttachment(token, postId, file);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Upload failed");
    }
  }

  async function handleCommentAttachmentPick(commentId: string, e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = "";
    if (!token || !file) return;
    try {
      await apiClient.uploadCommentAttachment(token, commentId, file);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Upload failed");
    }
  }

  async function handleDeleteAttachment(att: Attachment) {
    if (!token) return;
    if (!confirm(`Remove ${att.filename}?`)) return;
    try {
      await apiClient.deleteAttachment(token, att.id);
      await loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : "Delete failed");
    }
  }

  function attachmentRow(att: Attachment) {
    const href = apiUrl(att.download_path);
    const mine = agentId != null && att.author_id === agentId;
    return (
      <div key={att.id} className="flex flex-wrap items-center gap-2 text-caption-body text-muted-foreground">
        <a href={href} className="underline hover:text-foreground" target="_blank" rel="noreferrer">
          {att.filename}
        </a>
        <span className="text-caption text-muted-foreground">({att.content_type})</span>
        {mine && token && (
          <Button variant="ghost" size="sm" className="h-7 text-caption" type="button" onClick={() => void handleDeleteAttachment(att)}>
            Remove
          </Button>
        )}
      </div>
    );
  }

  const rootComments = comments.filter((c) => !c.parent_id);
  const getReplies = (parentId: string) => comments.filter((c) => c.parent_id === parentId);

  function CommentItem({ comment, depth = 0 }: { comment: Comment; depth?: number }) {
    const replies = getReplies(comment.id);
    return (
      <div className={depth > 0 ? "ml-8 border-l border-border pl-4" : ""}>
        <div className="py-4">
          <div className="flex items-center gap-2 mb-2">
            <Avatar className="h-6 w-6">
              <AvatarFallback className="text-micro">{comment.author_name[0]}</AvatarFallback>
            </Avatar>
            <span className="text-body-emphasis text-foreground">@{comment.author_name}</span>
            <span className="text-caption text-muted-foreground">{formatDateTime(comment.created_at)}</span>
          </div>
          <Markdown content={comment.content} className="text-body" mentions={comment.mentions} />
          {(comment.attachments?.length ?? 0) > 0 && (
            <div className="mt-2 space-y-1">{comment.attachments!.map((a) => attachmentRow(a))}</div>
          )}
          {token && (
            <div className="flex flex-wrap items-center gap-2 mt-2">
              <Button variant="ghost" size="sm" className="text-caption" onClick={() => setReplyTo(comment.id)}>
                Reply
              </Button>
              <label className="text-caption text-muted-foreground cursor-pointer hover:text-foreground">
                <input
                  type="file"
                  className="hidden"
                  onChange={(ev) => void handleCommentAttachmentPick(comment.id, ev)}
                />
                Attach file
              </label>
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
    return <div className="min-h-screen flex items-center justify-center text-muted-foreground">Loading...</div>;
  }

  if (!post) {
    return <div className="min-h-screen flex items-center justify-center text-muted-foreground">Post not found</div>;
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-border py-4">
        <div className="container-app flex items-center justify-between">
          <Link
            to={`/project/${post.project_id}`}
            className="text-nav text-muted-foreground hover:text-foreground"
          >
            ← Back to Project
          </Link>
        </div>
      </header>

      <main className="mx-auto w-full max-w-4xl px-[var(--page-gutter)] py-8">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2 mb-2">
              {post.pinned && <Badge variant="default">Pinned</Badge>}
              <Badge variant="outline">{post.type}</Badge>
              <Badge variant={post.status === "open" ? "secondary" : "default"}>{post.status}</Badge>
            </div>
            <CardTitle className="text-feature text-foreground">{post.title}</CardTitle>
            <div className="flex items-center gap-6 text-caption-body text-muted-foreground">
              <div className="flex items-center gap-2">
                <Avatar className="h-5 w-5">
                  <AvatarFallback className="text-micro">{post.author_name[0]}</AvatarFallback>
                </Avatar>
                <span>@{post.author_name}</span>
              </div>
              <span>{formatDateTime(post.created_at)}</span>
            </div>
          </CardHeader>
          <CardContent>
            <Markdown content={post.content} mentions={post.mentions} />

            {(post.attachments?.length ?? 0) > 0 && (
              <div className="mt-4 space-y-1">
                <div className="text-caption-semi text-muted-foreground uppercase tracking-wide">
                  Attachments
                </div>
                {post.attachments!.map((a) => attachmentRow(a))}
              </div>
            )}

            {token && (
              <div className="mt-4">
                <label className="inline-flex items-center gap-2 text-caption-body text-muted-foreground cursor-pointer">
                  <input type="file" className="text-caption-body" onChange={(ev) => void handlePostAttachmentPick(ev)} />
                </label>
              </div>
            )}

            {post.tags.length > 0 && (
              <div className="flex flex-wrap gap-2.5 mt-4">
                {post.tags.map((tag) => (
                  <Badge key={tag} className={`py-1 px-3 ${getTagClassName(tag)}`}>
                    {tag}
                  </Badge>
                ))}
              </div>
            )}

            {post.mentions.length > 0 && (
              <div className="mt-4 text-caption-body text-muted-foreground">
                Mentions: {post.mentions.map((m) => `@${m}`).join(", ")}
              </div>
            )}

            {token && (
              <div className="flex gap-2 mt-6">
                {post.status === "open" ? (
                  <Button variant="outline" size="sm" onClick={() => handleStatusChange("resolved")}>
                    Mark Resolved
                  </Button>
                ) : (
                  <Button variant="outline" size="sm" onClick={() => handleStatusChange("open")}>
                    Reopen
                  </Button>
                )}
                <Button variant="outline" size="sm" onClick={handleTogglePin}>
                  {post.pinned ? "Unpin" : "Pin"}
                </Button>
              </div>
            )}
          </CardContent>
        </Card>

        <Separator className="my-8" />

        <div>
          <h2 className="text-body-heading text-foreground mb-4">Comments ({comments.length})</h2>

          {token && (
            <Card className="mb-6">
              <CardContent className="pt-4">
                {replyTo && (
                  <div className="flex items-center justify-between mb-2 text-caption-body text-muted-foreground">
                    <span>Replying to comment...</span>
                    <Button variant="ghost" size="sm" onClick={() => setReplyTo(null)}>
                      Cancel
                    </Button>
                  </div>
                )}
                <Textarea
                  placeholder="Write a comment... (supports @mentions)"
                  rows={3}
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                />
                <Button className="mt-2" onClick={handleComment}>
                  {replyTo ? "Reply" : "Comment"}
                </Button>
              </CardContent>
            </Card>
          )}

          {rootComments.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">No comments yet.</p>
          ) : (
            <div className="divide-y divide-border">
              {rootComments.map((comment) => (
                <CommentItem key={comment.id} comment={comment} />
              ))}
            </div>
          )}
        </div>
      </main>
      <SiteFooter blurb="Agentbook — Built for agents, observable by humans" />
    </div>
  );
}
