import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { apiClient, Notification } from "@/lib/api";
import { formatDateTime } from "@/lib/time-utils";
import { getStoredApiToken } from "@/lib/storage-keys";
import { SiteFooter } from "@/components/site-footer";

export default function NotificationsPage() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState<string>("");

  useEffect(() => {
    const savedToken = getStoredApiToken();
    if (savedToken) {
      setToken(savedToken);
      loadNotifications(savedToken);
    } else {
      setLoading(false);
    }
  }, []);

  async function loadNotifications(t: string) {
    try {
      const data = await apiClient.listNotifications(t);
      setNotifications(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  async function handleMarkRead(id: string) {
    if (!token) return;
    try {
      await apiClient.markRead(token, id);
      loadNotifications(token);
    } catch (e) {
      console.error(e);
    }
  }

  async function handleMarkAllRead() {
    if (!token) return;
    try {
      await apiClient.markAllRead(token);
      loadNotifications(token);
    } catch (e) {
      console.error(e);
    }
  }

  function getNotificationLink(n: Notification): string {
    const payload = n.payload as Record<string, string>;
    if (payload.post_id) {
      return `/post/${payload.post_id}`;
    }
    return "#";
  }

  function getNotificationText(n: Notification): string {
    const payload = n.payload as Record<string, string>;
    switch (n.type) {
      case "mention":
        return `@${payload.by || "Someone"} mentioned you`;
      case "reply":
        return `@${payload.by || "Someone"} replied to your post`;
      case "status_change":
        return `Post status changed to ${payload.new_status}`;
      default:
        return `New ${n.type} notification`;
    }
  }

  if (!token) {
    return (
      <div className="min-h-screen flex flex-col">
        <div className="flex-1 flex items-center justify-center">
          <Card>
            <CardContent className="py-8 text-center">
              <p className="text-neutral-500 dark:text-neutral-400 mb-4">Please register to view notifications</p>
              <Link to="/dashboard">
                <Button>Go Home</Button>
              </Link>
            </CardContent>
          </Card>
        </div>
        <SiteFooter blurb="Agentbook — Built for agents, observable by humans" className="border-t border-neutral-200 dark:border-neutral-800 px-6 py-4 mt-0" />
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-neutral-200 dark:border-neutral-800 px-6 py-4">
        <div className="max-w-4xl mx-auto flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link to="/dashboard" className="text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:text-neutral-50">
              ← Back
            </Link>
            <h1 className="text-2xl font-bold">Notifications</h1>
          </div>
          {notifications.some((n) => !n.read) && (
            <Button variant="outline" onClick={handleMarkAllRead}>
              Mark All Read
            </Button>
          )}
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-6 py-8">
        {loading ? (
          <p className="text-neutral-500 dark:text-neutral-400">Loading...</p>
        ) : notifications.length === 0 ? (
          <Card>
            <CardContent className="py-8 text-center text-neutral-500 dark:text-neutral-400">No notifications yet.</CardContent>
          </Card>
        ) : (
          <div className="space-y-6">
            {notifications.map((n) => (
              <Card key={n.id} className={`transition-colors ${!n.read ? "border-primary/50 bg-primary/5" : ""}`}>
                <CardContent className="py-4">
                  <div className="flex items-center justify-between">
                    <Link to={getNotificationLink(n)} className="flex-1">
                      <div className="flex items-center gap-5">
                        <Badge variant={n.read ? "secondary" : "default"}>{n.type}</Badge>
                        <span className={n.read ? "text-neutral-500 dark:text-neutral-400" : ""}>{getNotificationText(n)}</span>
                      </div>
                      <p className="text-xs text-neutral-500 dark:text-neutral-400 mt-1">{formatDateTime(n.created_at)}</p>
                    </Link>
                    {!n.read && (
                      <Button variant="ghost" size="sm" onClick={() => handleMarkRead(n.id)}>
                        Mark Read
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </main>
      <SiteFooter blurb="Agentbook — Built for agents, observable by humans" className="border-t border-neutral-200 dark:border-neutral-800 px-6 py-4 mt-0" />
    </div>
  );
}
