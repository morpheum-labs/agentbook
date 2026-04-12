import { useEffect, useRef } from "react";
import { apiOrigin } from "@/lib/api-base";

/** WebSocket URL for agentglobe realtime (same host as `VITE_API_URL`). */
export function buildRealtimeWsUrl(apiKey: string): string {
  const o = apiOrigin();
  let wsBase = o;
  if (o.startsWith("https://")) {
    wsBase = "wss://" + o.slice("https://".length);
  } else if (o.startsWith("http://")) {
    wsBase = "ws://" + o.slice("http://".length);
  } else {
    wsBase = "ws://" + o;
  }
  return `${wsBase}/api/v1/ws?${new URLSearchParams({ token: apiKey }).toString()}`;
}

/**
 * One WebSocket per (projectId, token); invokes onEvent for frames whose `project_id` matches or type is `connected`.
 */
export function useProjectRealtime(
  projectId: string | undefined,
  token: string | undefined,
  onEvent: (msg: Record<string, unknown>) => void,
): void {
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;
  useEffect(() => {
    if (!projectId || !token) return;
    const ws = new WebSocket(buildRealtimeWsUrl(token));
    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(String(ev.data)) as Record<string, unknown>;
        const t = msg.type;
        const pid = msg.project_id;
        if (t === "connected" || pid === projectId) {
          onEventRef.current(msg);
        }
      } catch {
        /* ignore malformed */
      }
    };
    return () => {
      ws.close();
    };
  }, [projectId, token]);
}
