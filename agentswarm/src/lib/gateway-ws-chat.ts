/** Browser client for ZeroClaw gateway `GET /ws/chat` (see miroclaw `gateway/ws.rs`). */

const WS_PROTOCOL = "zeroclaw.v1";

export type GatewayWsTurnResult = {
  fullResponse: string;
};

/**
 * Maps a runtime HTTP(S) `public_url` (clawgotcha `SwarmRuntimeInstance.PublicURL`) to a WebSocket origin
 * used before `/ws/chat`: same host/port, `ws`/`wss` scheme, optional **path prefix** (for reverse proxies),
 * no query or fragment.
 */
export function publicHttpUrlToGatewayWsBase(publicUrl: string): string {
  const raw = publicUrl.trim();
  if (!raw) {
    throw new Error("Public URL is empty");
  }
  const u = new URL(/^https?:\/\//i.test(raw) ? raw : `https://${raw}`);
  u.protocol = u.protocol === "https:" ? "wss:" : "ws:";
  u.search = "";
  u.hash = "";
  u.pathname = u.pathname.replace(/\/+$/, "") || "/";
  let out = u.toString();
  if (u.pathname === "/" && out.endsWith("/")) {
    out = out.slice(0, -1);
  }
  return out;
}

/** Apply MiroClaw gateway chat path: root → `/ws/chat`, prefix → `{prefix}/ws/chat`, already `…/ws/chat` unchanged. */
export function joinGatewayChatPath(pathname: string): string {
  const p = pathname.replace(/\/+$/, "") || "";
  if (!p || p === "/") {
    return "/ws/chat";
  }
  if (p === "/ws/chat" || p.endsWith("/ws/chat")) {
    return p;
  }
  return `${p}/ws/chat`;
}

/**
 * Example full chat URL for UI previews (placeholder `session_id`).
 */
export function previewGatewayChatWsUrl(wsBase: string): string {
  const base = wsBase.trim();
  if (!base) return "";
  try {
    return buildGatewayChatWsUrl(base, {
      sessionId: "_preview_session",
      name: "Hand name",
    });
  } catch {
    return "";
  }
}

/**
 * Build a WebSocket URL for `/ws/chat`. Accepts `ws://host`, `http://host`, or `host:port`.
 */
export function buildGatewayChatWsUrl(
  base: string,
  opts: {
    sessionId: string;
    name?: string;
    token?: string;
    fresh?: boolean;
  }
): string {
  let s = base.trim();
  if (!s) {
    throw new Error("Gateway URL is empty");
  }
  if (!/^wss?:\/\//i.test(s)) {
    if (/^https?:\/\//i.test(s)) {
      s = s.replace(/^http/i, "ws");
    } else {
      s = `ws://${s}`;
    }
  }
  const u = new URL(s);
  u.pathname = joinGatewayChatPath(u.pathname);
  u.search = "";
  u.searchParams.set("session_id", opts.sessionId);
  if (opts.name?.trim()) {
    u.searchParams.set("name", opts.name.trim());
  }
  if (opts.token?.trim()) {
    u.searchParams.set("token", opts.token.trim());
  }
  if (opts.fresh) {
    u.searchParams.set("fresh", "1");
  }
  return u.toString();
}

function parseWsMessage(raw: string): { type?: string; content?: string; full_response?: string; message?: string } {
  try {
    return JSON.parse(raw) as { type?: string; content?: string; full_response?: string; message?: string };
  } catch {
    return {};
  }
}

/** Redact token for safe display in errors (browser WebSocket URLs often carry ?token=). */
export function redactGatewayWsUrlForDisplay(wsUrl: string): string {
  try {
    const u = new URL(wsUrl);
    if (u.searchParams.has("token")) {
      u.searchParams.set("token", "(redacted)");
    }
    return u.toString();
  } catch {
    return wsUrl;
  }
}

/**
 * One user turn over an ephemeral WebSocket (gateway persists history by `session_id`).
 */
export function gatewayChatSingleTurn(
  wsUrl: string,
  userContent: string,
  options?: { signal?: AbortSignal }
): Promise<GatewayWsTurnResult> {
  return new Promise((resolve, reject) => {
    let settled = false;
    let accum = "";
    let transportErrored = false;

    const fail = (e: Error) => {
      if (settled) return;
      settled = true;
      reject(e);
    };

    const succeed = (full: string) => {
      if (settled) return;
      settled = true;
      resolve({ fullResponse: full });
    };

    let ws: WebSocket;
    try {
      ws = new WebSocket(wsUrl, [WS_PROTOCOL]);
    } catch (e) {
      fail(e instanceof Error ? e : new Error("WebSocket constructor failed"));
      return;
    }

    const onAbort = () => {
      try {
        ws.close();
      } catch {
        // ignore
      }
      fail(new Error("Aborted"));
    };
    if (options?.signal) {
      if (options.signal.aborted) {
        onAbort();
        return;
      }
      options.signal.addEventListener("abort", onAbort, { once: true });
    }

    const target = redactGatewayWsUrlForDisplay(wsUrl);

    ws.onerror = () => {
      transportErrored = true;
    };

    ws.onclose = (ev: CloseEvent) => {
      if (settled) {
        return;
      }
      const bits = [`code ${ev.code}`];
      if (ev.reason) {
        bits.push(ev.reason);
      }
      const wasClean = ev.wasClean;
      const hint =
        "HTTPS pages require wss:// (not ws://). Confirm MiroClaw gateway is listening, path ends with /ws/chat, pairing token if enabled, and no proxy stripping WebSockets.";
      if (transportErrored || !wasClean) {
        fail(
          new Error(
            `WebSocket failed (${bits.join(" — ")}). Target: ${target}. ${hint}`
          )
        );
      } else {
        fail(new Error(`WebSocket closed before completion (${bits.join(" — ")}). Target: ${target}`));
      }
    };

    ws.onmessage = (ev) => {
      const text = typeof ev.data === "string" ? ev.data : "";
      const msg = parseWsMessage(text);
      const t = msg.type;

      if (t === "session_start" || t === "connected") {
        return;
      }

      if (t === "chunk" && typeof msg.content === "string") {
        accum += msg.content;
        return;
      }

      if (t === "chunk_reset") {
        return;
      }

      if (t === "tool_call" || t === "tool_result") {
        return;
      }

      if (t === "done") {
        const full =
          typeof msg.full_response === "string" && msg.full_response.length > 0 ? msg.full_response : accum;
        try {
          ws.close();
        } catch {
          // ignore
        }
        succeed(full);
        return;
      }

      if (t === "error") {
        try {
          ws.close();
        } catch {
          // ignore
        }
        fail(new Error(msg.message?.trim() || "Gateway error"));
      }
    };

    ws.onopen = () => {
      try {
        ws.send(JSON.stringify({ type: "message", content: userContent }));
      } catch (e) {
        fail(e instanceof Error ? e : new Error("send failed"));
      }
    };
  });
}
