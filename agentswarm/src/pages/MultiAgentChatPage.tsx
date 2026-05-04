import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { MessagesSquare, RefreshCw, Send } from "lucide-react";
import { fetchAgents, fetchInstances, type SwarmAgent, type SwarmRuntimeInstance } from "@/lib/api";
import { buildGatewayChatWsUrl, gatewayChatSingleTurn, publicHttpUrlToGatewayWsBase } from "@/lib/gateway-ws-chat";
import { MULTI_CHAT_SESSION } from "@/lib/multi-chat-gateway-session";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

type GatewaySessionSource = "runtime" | "manual";

function readGatewaySession(): {
  source: GatewaySessionSource;
  runtimeInstanceName: string;
  legacyManualBase: string;
  gatewayToken: string;
} {
  try {
    const srcRaw = sessionStorage.getItem(MULTI_CHAT_SESSION.GATEWAY_SOURCE)?.trim();
    const legacyManualBase = sessionStorage.getItem(MULTI_CHAT_SESSION.GATEWAY_BASE)?.trim() ?? "";
    const runtimeInstanceName = sessionStorage.getItem(MULTI_CHAT_SESSION.GATEWAY_INSTANCE) ?? "";
    const gatewayToken = sessionStorage.getItem(MULTI_CHAT_SESSION.GATEWAY_TOKEN) ?? "";
    let source: GatewaySessionSource;
    if (srcRaw === "manual") {
      source = "manual";
    } else if (srcRaw === "runtime") {
      source = "runtime";
    } else {
      source = legacyManualBase ? "manual" : "runtime";
    }
    return { source, runtimeInstanceName, legacyManualBase, gatewayToken };
  } catch {
    return { source: "runtime", runtimeInstanceName: "", legacyManualBase: "", gatewayToken: "" };
  }
}

function newId(): string {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

/** Readable segment for ZeroClaw `session_id` (miroclaw stores history under `gw_<session_id>`). */
function handSlug(name: string): string {
  const s = name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 48);
  return s.length > 0 ? s : "hand";
}

type AssistantRow = {
  id: string;
  agentId: string;
  agentName: string;
  content: string;
  pending: boolean;
  error?: string;
};

type UserRow = {
  id: string;
  content: string;
  /** Snapshot of who received this broadcast when Send was pressed. */
  recipientsLabel: string;
};

type TurnBlock = {
  user: UserRow;
  assistants: AssistantRow[];
};

export function MultiAgentChatPage() {
  const [agents, setAgents] = useState<SwarmAgent[] | null>(null);
  const [loadErr, setLoadErr] = useState<string | null>(null);
  const [instances, setInstances] = useState<SwarmRuntimeInstance[] | null>(null);
  const [instancesErr, setInstancesErr] = useState<string | null>(null);
  const [selected, setSelected] = useState<Set<string>>(() => new Set());
  const [turns, setTurns] = useState<TurnBlock[]>([]);
  const [draft, setDraft] = useState("");
  const [inFlight, setInFlight] = useState(false);
  /** Gateway wiring comes from session (Runtime instances → Pair & chat). Re-read on mount / focus return. */
  const [gwSession, setGwSession] = useState(readGatewaySession);
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setGwSession(readGatewaySession());
  }, []);

  useEffect(() => {
    function onVis() {
      if (document.visibilityState === "visible") {
        setGwSession(readGatewaySession());
      }
    }
    document.addEventListener("visibilitychange", onVis);
    return () => document.removeEventListener("visibilitychange", onVis);
  }, []);

  const loadInstances = useCallback(() => {
    setInstancesErr(null);
    setInstances(null);
    fetchInstances()
      .then((list) => setInstances(list))
      .catch((e: unknown) => {
        setInstancesErr(e instanceof Error ? e.message : "Failed to load runtime instances");
      });
  }, []);

  useEffect(() => {
    void loadInstances();
  }, [loadInstances]);

  const selectedRuntime = useMemo(() => {
    if (!instances?.length || !gwSession.runtimeInstanceName.trim()) return undefined;
    return instances.find((i) => i.InstanceName === gwSession.runtimeInstanceName.trim());
  }, [instances, gwSession.runtimeInstanceName]);

  const resolvedGatewayBase = useMemo(() => {
    if (gwSession.source === "manual") {
      return gwSession.legacyManualBase.trim();
    }
    const pu = selectedRuntime?.PublicURL?.trim();
    if (!pu) return "";
    try {
      return publicHttpUrlToGatewayWsBase(pu);
    } catch {
      return "";
    }
  }, [gwSession.source, gwSession.legacyManualBase, selectedRuntime]);

  const sessionInstanceKey = useMemo(() => {
    if (gwSession.source === "runtime" && selectedRuntime?.InstanceName?.trim()) {
      return selectedRuntime.InstanceName.trim();
    }
    return "manual";
  }, [gwSession.source, selectedRuntime]);

  /** Blocks Send until gateway session + clawgotcha data are coherent. */
  const gatewayBlockReason = useMemo((): string | null => {
    if (gwSession.source === "manual") {
      if (!gwSession.legacyManualBase.trim()) {
        return "No gateway base in session. Open Runtime instances, use Pair & chat, or restore a legacy manual ws:// base in session storage.";
      }
      return null;
    }
    if (instances === null) {
      if (instancesErr) {
        return `Could not load runtime instances: ${instancesErr}`;
      }
      return "Loading runtime instances from clawgotcha…";
    }
    if (!gwSession.runtimeInstanceName.trim()) {
      return "No runtime instance selected in session. Go to Runtime instances → Pair & chat for the gateway you want.";
    }
    const name = gwSession.runtimeInstanceName.trim();
    const inst = instances.find((i) => i.InstanceName === name);
    if (!inst) {
      return `Instance “${name}” is not in clawgotcha anymore. Run Pair & chat again from Runtime instances.`;
    }
    const pu = inst.PublicURL?.trim();
    if (!pu) {
      return `Instance “${name}” has no public_url. Fix registration or infer from callback_url on the control plane, then pair again from Runtime instances.`;
    }
    try {
      publicHttpUrlToGatewayWsBase(pu);
      return null;
    } catch (e) {
      const msg = e instanceof Error ? e.message : "invalid URL";
      return `public_url “${pu}” is not usable as a WebSocket origin (${msg}).`;
    }
  }, [gwSession.source, gwSession.runtimeInstanceName, gwSession.legacyManualBase, instances, instancesErr]);

  const gatewayNotReady =
    gwSession.source === "manual"
      ? !gwSession.legacyManualBase.trim()
      : !!gatewayBlockReason;

  const gatewayStatusIsLoading =
    gwSession.source === "runtime" && instances === null && !instancesErr && !!gatewayBlockReason;

  const loadAgents = useCallback(() => {
    setLoadErr(null);
    setAgents(null);
    fetchAgents()
      .then((list) => {
        setAgents(list);
        setSelected((prev) => {
          const next = new Set<string>();
          for (const id of prev) {
            if (list.some((a) => a.ID === id)) next.add(id);
          }
          return next;
        });
      })
      .catch((e: unknown) => {
        setLoadErr(e instanceof Error ? e.message : "Failed to load agents");
      });
  }, []);

  useEffect(() => {
    void loadAgents();
  }, [loadAgents]);

  const selectedAgents = useMemo(() => {
    if (!agents) return [];
    return agents.filter((a) => selected.has(a.ID));
  }, [agents, selected]);

  const toggle = useCallback((id: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const selectAll = useCallback(() => {
    if (!agents) return;
    setSelected(new Set(agents.map((a) => a.ID)));
  }, [agents]);

  const clearSelection = useCallback(() => {
    setSelected(new Set());
  }, []);

  useEffect(() => {
    const el = scrollRef.current;
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  }, [turns, inFlight]);

  async function send() {
    const text = draft.trim();
    if (!text || selectedAgents.length === 0 || inFlight) return;

    const recipientsLabel =
      selectedAgents.length === 1
        ? selectedAgents[0].Name
        : `${selectedAgents.length} hands · ${selectedAgents.map((a) => a.Name).join(", ")}`;
    const userRow: UserRow = { id: newId(), content: text, recipientsLabel };
    const assistantRows: AssistantRow[] = selectedAgents.map((a) => ({
      id: newId(),
      agentId: a.ID,
      agentName: a.Name,
      content: "",
      pending: true,
    }));

    setTurns((t) => [...t, { user: userRow, assistants: assistantRows }]);
    setDraft("");
    setInFlight(true);

    const gw = resolvedGatewayBase.trim();
    const token = gwSession.gatewayToken.trim();

    await Promise.all(
      assistantRows.map(async (row) => {
        const agent = selectedAgents.find((a) => a.ID === row.agentId);
        if (!agent) return;

        const patchAssistant = (patch: Partial<AssistantRow>) => {
          setTurns((prev) =>
            prev.map((block) => {
              if (!block.assistants.some((r) => r.id === row.id)) return block;
              return {
                ...block,
                assistants: block.assistants.map((r) =>
                  r.id === row.id ? { ...r, ...patch } : r
                ),
              };
            })
          );
        };

        if (!gw) {
          const hint = [
            gatewayBlockReason ?? "Could not resolve a WebSocket base for this gateway session.",
            "Each hand uses its own session_id on GET /ws/chat (zeroclaw.v1); transcripts are gw_<session_id> on the gateway.",
          ].join(" ");
          patchAssistant({
            pending: false,
            content: hint,
          });
          return;
        }

        try {
          const sessionId = `agentswarm:${sessionInstanceKey}:${handSlug(agent.Name)}:${agent.ID}`;
          const wsUrl = buildGatewayChatWsUrl(gw, {
            sessionId,
            name: agent.Name,
            token: token || undefined,
            fresh: false,
          });
          const { fullResponse } = await gatewayChatSingleTurn(wsUrl, text);
          patchAssistant({ pending: false, content: fullResponse });
        } catch (e) {
          patchAssistant({
            pending: false,
            content: "",
            error: e instanceof Error ? e.message : "Request failed",
          });
        }
      })
    );

    setInFlight(false);
  }

  return (
    <div className="container-app max-w-6xl py-8 pb-12">
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <h2 className="text-body-heading flex items-center gap-2">
            <MessagesSquare className="size-5 opacity-90" />
            Multi-agent chat
          </h2>
          <p className="text-caption-body text-muted-foreground mt-1 max-w-xl">
            Choose clawgotcha hands, send one prompt, and read side‑by‑side replies over{" "}
            <span className="font-mono text-xs">GET /ws/chat</span> (one connection per hand, distinct{" "}
            <span className="font-mono text-xs">session_id</span>). Pair the gateway and pick a runtime on{" "}
            <Link to="/instances" className="text-primary underline-offset-2 hover:underline">
              Runtime instances
            </Link>
            .
          </p>
        </div>
        <Button
          type="button"
          variant="secondary"
          size="sm"
          onClick={loadAgents}
          disabled={agents === null && !loadErr}
          className={cn(
            "h-9 rounded-lg border border-border/60 bg-accent/50 shadow-sm",
            "text-foreground hover:bg-accent/80"
          )}
        >
          <RefreshCw className="size-4" />
          Refresh agents
        </Button>
      </div>

      <div className="flex flex-col gap-4 lg:flex-row lg:items-stretch">
        <Card className="lg:w-72 shrink-0 border-border/80 p-4">
          <div className="flex items-center justify-between gap-2">
            <h3 className="text-subheading-sm font-medium">Hands</h3>
            <span className="text-micro text-muted-foreground">{selected.size} selected</span>
          </div>
          <div className="mt-3 flex flex-wrap gap-2">
            <Button type="button" variant="outline" size="xs" onClick={selectAll} disabled={!agents?.length}>
              Select all
            </Button>
            <Button type="button" variant="outline" size="xs" onClick={clearSelection} disabled={selected.size === 0}>
              Clear
            </Button>
          </div>

          {loadErr && (
            <p className="text-destructive text-caption-body mt-3" role="alert">
              {loadErr}
            </p>
          )}

          {agents === null && !loadErr && (
            <p className="text-muted-foreground text-caption-body mt-4">Loading…</p>
          )}

          {agents && agents.length === 0 && (
            <p className="text-muted-foreground text-caption-body mt-4">No agents yet.</p>
          )}

          {agents && agents.length > 0 && (
            <ul className="mt-4 flex max-h-[min(50vh,28rem)] flex-col gap-2 overflow-y-auto pr-1">
              {agents.map((a) => {
                const isOn = selected.has(a.ID);
                return (
                  <li key={a.ID}>
                    <label
                      className={cn(
                        "flex cursor-pointer items-start gap-3 rounded-xl border px-3 py-2.5 transition-colors",
                        isOn
                          ? "border-primary/50 bg-primary/5"
                          : "border-border/60 bg-card hover:bg-accent/30"
                      )}
                    >
                      <input
                        type="checkbox"
                        checked={isOn}
                        onChange={() => toggle(a.ID)}
                        className="mt-1 size-4 shrink-0 rounded border-border accent-primary"
                      />
                      <span className="min-w-0">
                        <span className="text-body block font-medium leading-snug">{a.Name}</span>
                        <span className="text-micro font-mono text-muted-foreground break-all">{a.ID}</span>
                      </span>
                    </label>
                  </li>
                );
              })}
            </ul>
          )}

          <div className="mt-5 border-t border-border/60 pt-4 space-y-3">
            <h3 className="text-subheading-sm font-medium">Gateway</h3>
            <p className="text-micro text-muted-foreground leading-relaxed">
              Pair or switch runtimes on{" "}
              <Link to="/instances" className="text-primary underline-offset-2 hover:underline">
                Runtime instances
              </Link>
              .
            </p>
            {gwSession.source === "runtime" && gwSession.runtimeInstanceName.trim() ? (
              <div>
                <span className="text-micro text-muted-foreground block">Runtime</span>
                <p className="text-caption-body font-medium">{gwSession.runtimeInstanceName}</p>
              </div>
            ) : gwSession.source === "manual" && gwSession.legacyManualBase.trim() ? (
              <div>
                <span className="text-micro text-muted-foreground block">Legacy manual base</span>
                <p className="text-micro font-mono break-all rounded-md border border-border/60 bg-muted/30 px-2 py-1.5">
                  {gwSession.legacyManualBase}
                </p>
              </div>
            ) : (
              <p className="text-micro text-muted-foreground">No gateway session yet.</p>
            )}
            <div>
              <span className="text-micro text-muted-foreground block mb-1">
                WebSocket base for <span className="font-mono">/ws/chat</span>
              </span>
              <p className="text-micro font-mono text-muted-foreground break-all rounded-md border border-border/60 bg-muted/30 px-2 py-1.5">
                {resolvedGatewayBase || "—"}
              </p>
            </div>
            <p className="text-micro text-muted-foreground leading-relaxed">
              {gwSession.gatewayToken.trim()
                ? "Pairing token present for this tab."
                : "No pairing token in session — required if the gateway enforces pairing."}
            </p>
            <p className="text-micro text-muted-foreground leading-relaxed">
              Each hand uses <span className="font-mono">Sec-WebSocket-Protocol: zeroclaw.v1</span> and its own{" "}
              <span className="font-mono">session_id</span>; gateway transcripts use keys{" "}
              <span className="font-mono">gw_…</span>.
            </p>
          </div>
        </Card>

        <Card className="min-h-[min(70vh,36rem)] flex-1 flex flex-col border-border/80 overflow-hidden">
          <div
            ref={scrollRef}
            className="flex-1 space-y-4 overflow-y-auto p-4 sm:p-5"
          >
            {turns.length === 0 && (
              <p className="text-muted-foreground text-body">
                Select hands, pair a gateway on{" "}
                <Link to="/instances" className="text-primary underline-offset-2 hover:underline">
                  Runtime instances
                </Link>{" "}
                if needed, then type a message and press Send.
              </p>
            )}
            {turns.map((block) => (
              <div key={block.user.id} className="space-y-3">
                <div className="rounded-xl bg-muted/50 border border-border/50 px-4 py-3">
                  <p className="text-micro text-muted-foreground mb-1">You → {block.user.recipientsLabel}</p>
                  <p className="text-body whitespace-pre-wrap">{block.user.content}</p>
                </div>
                <div className="grid gap-3 sm:grid-cols-2">
                  {block.assistants.map((r) => (
                    <div
                      key={r.id}
                      className={cn(
                        "rounded-xl border px-3 py-3",
                        r.error ? "border-destructive/40 bg-destructive/5" : "border-border/60 bg-card"
                      )}
                    >
                      <p className="text-micro font-medium text-muted-foreground mb-2">{r.agentName}</p>
                      {r.pending ? (
                        <p className="text-caption-body text-muted-foreground animate-pulse">Thinking…</p>
                      ) : r.error ? (
                        <p className="text-destructive text-caption-body">{r.error}</p>
                      ) : (
                        <p className="text-caption-body whitespace-pre-wrap">{r.content}</p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>

          <div className="border-t border-border/60 p-4 sm:p-5">
            {gatewayBlockReason && (
              <p
                role={gatewayStatusIsLoading ? "status" : "alert"}
                className={cn(
                  "mb-3 rounded-xl border px-3 py-2 text-caption-body leading-snug",
                  gatewayStatusIsLoading
                    ? "text-muted-foreground bg-muted/40 border-border/60"
                    : "text-destructive bg-destructive/5 border-destructive/30"
                )}
              >
                {gatewayBlockReason}
              </p>
            )}
            <Textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="Message all selected hands…"
              rows={3}
              disabled={inFlight}
              className="resize-y min-h-[5rem] rounded-xl"
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  void send();
                }
              }}
            />
            <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
              <p className="text-micro text-muted-foreground">
                <kbd className="rounded border border-border/60 px-1">Enter</kbd> send ·{" "}
                <kbd className="rounded border border-border/60 px-1">Shift</kbd>+
                <kbd className="rounded border border-border/60 px-1">Enter</kbd> newline
              </p>
              <Button
                type="button"
                size="sm"
                disabled={
                  inFlight || !draft.trim() || selectedAgents.length === 0 || gatewayNotReady
                }
                onClick={() => void send()}
                className="rounded-lg"
              >
                <Send className="size-4" />
                Send
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
