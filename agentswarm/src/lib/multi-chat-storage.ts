import type { TurnBlock } from "@/components/multi-agent-chat-panel";

const LS_KEY_PREFIX = "agentswarm.multiChat.v1";

export type MultiChatLocalSnapshot = {
  turnsByAgent: Record<string, TurnBlock[]>;
  /** Last WebSocket `session_id` used per agent (for gateway key `gw_<session_id>`). */
  sessionIdsByAgent: Record<string, string>;
  /** Last gateway replay `seq` seen per agent (for `connect.last_event_seq`). */
  wsMaxSeqByAgent?: Record<string, number>;
};

function djb2Hex(s: string): string {
  let h = 5381;
  for (let i = 0; i < s.length; i++) {
    h = (h * 33) ^ s.charCodeAt(i);
  }
  return (h >>> 0).toString(36);
}

/**
 * Stable scope for localStorage: same paired runtime or same manual base → same bucket.
 */
export function multiChatPersistenceNamespace(gw: {
  source: "runtime" | "manual";
  runtimeInstanceName: string;
  legacyManualBase: string;
}): string {
  if (gw.source === "manual") {
    return `m:${djb2Hex(gw.legacyManualBase.trim())}`;
  }
  const n = gw.runtimeInstanceName.trim();
  return n ? `r:${n}` : "r:pending";
}

function storageKey(namespace: string): string {
  return `${LS_KEY_PREFIX}:${namespace}`;
}

function sanitizeTurnsForStorage(turns: Record<string, TurnBlock[]>): Record<string, TurnBlock[]> {
  const out: Record<string, TurnBlock[]> = {};
  for (const [agentId, list] of Object.entries(turns)) {
    out[agentId] = list.map((block) => ({
      ...block,
      assistants: block.assistants.map((a) =>
        a.pending
          ? {
              ...a,
              pending: false,
              content:
                a.content.trim() ||
                "(Reply was still in progress — reopen the chat or send again to continue.)",
            }
          : a
      ),
    }));
  }
  return out;
}

export function loadMultiChatSnapshot(namespace: string): MultiChatLocalSnapshot | null {
  try {
    const raw = localStorage.getItem(storageKey(namespace));
    if (!raw?.trim()) return null;
    const j = JSON.parse(raw) as MultiChatLocalSnapshot;
    if (!j || typeof j !== "object" || !j.turnsByAgent || typeof j.turnsByAgent !== "object") {
      return null;
    }
    const sessionIdsByAgent =
      j.sessionIdsByAgent && typeof j.sessionIdsByAgent === "object" ? j.sessionIdsByAgent : {};
    const wsMaxSeqByAgent =
      j.wsMaxSeqByAgent && typeof j.wsMaxSeqByAgent === "object" ? j.wsMaxSeqByAgent : undefined;
    return { turnsByAgent: j.turnsByAgent, sessionIdsByAgent, ...(wsMaxSeqByAgent ? { wsMaxSeqByAgent } : {}) };
  } catch {
    return null;
  }
}

export function saveMultiChatSnapshot(namespace: string, snapshot: MultiChatLocalSnapshot): void {
  try {
    const payload: MultiChatLocalSnapshot = {
      turnsByAgent: sanitizeTurnsForStorage(snapshot.turnsByAgent),
      sessionIdsByAgent: { ...snapshot.sessionIdsByAgent },
      ...(snapshot.wsMaxSeqByAgent && Object.keys(snapshot.wsMaxSeqByAgent).length > 0
        ? { wsMaxSeqByAgent: { ...snapshot.wsMaxSeqByAgent } }
        : {}),
    };
    localStorage.setItem(storageKey(namespace), JSON.stringify(payload));
  } catch {
    // quota / private mode — ignore
  }
}
