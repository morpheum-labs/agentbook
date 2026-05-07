import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { fetchAgents, fetchInstances, type SwarmAgent, type SwarmRuntimeInstance } from "@/lib/api";
import {
  GATEWAY_CHAT_SLASH_COMMANDS_FALLBACK,
  buildGatewayChatWsUrl,
  fetchGatewayChatSlashCommandCatalog,
  gatewayChatSingleTurn,
  gatewayHttpBaseFromWsBase,
  publicHttpUrlToGatewayWsBase,
  type GatewaySlashCommandItem,
} from "@/lib/gateway-ws-chat";
import { MULTI_CHAT_SESSION } from "@/lib/multi-chat-gateway-session";
import { buildMultiChatSessionId } from "@/lib/multi-chat-session-id";
import {
  loadMultiChatSnapshot,
  multiChatPersistenceNamespace,
  saveMultiChatSnapshot,
} from "@/lib/multi-chat-storage";
import {
  MultiAgentChatPanel,
  type AssistantRow,
  type TurnBlock,
  type UserRow,
} from "@/components/multi-agent-chat-panel";
import { MultiAgentHandsPicker } from "@/components/multi-agent-hands-picker";
import { Card } from "@/components/ui/card";

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

export function MultiAgentChatPage() {
  const [agents, setAgents] = useState<SwarmAgent[] | null>(null);
  const [loadErr, setLoadErr] = useState<string | null>(null);
  const [instances, setInstances] = useState<SwarmRuntimeInstance[] | null>(null);
  const [instancesErr, setInstancesErr] = useState<string | null>(null);
  /** Which hand’s session is shown in the chat panel (each hand keeps its own transcript). */
  const [activeAgentId, setActiveAgentId] = useState<string | null>(null);
  const [turnsByAgent, setTurnsByAgent] = useState<Record<string, TurnBlock[]>>({});
  /** Last persisted WebSocket `session_id` per agent (localStorage). */
  const [sessionIdsByAgent, setSessionIdsByAgent] = useState<Record<string, string>>({});
  /** Index of first turn appended after hydrate; drives “past vs this visit” separator in the panel. */
  const [pastTurnBoundaryByAgent, setPastTurnBoundaryByAgent] = useState<Record<string, number>>({});
  const [draft, setDraft] = useState("");
  const [inFlight, setInFlight] = useState(false);
  const [slashCommands, setSlashCommands] = useState<GatewaySlashCommandItem[]>(
    GATEWAY_CHAT_SLASH_COMMANDS_FALLBACK
  );
  /** Gateway wiring comes from session (Runtime instances → Pair & chat). Re-read on mount / focus return. */
  const [gwSession, setGwSession] = useState(readGatewaySession);

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

  const persistenceNamespace = useMemo(
    () =>
      multiChatPersistenceNamespace({
        source: gwSession.source,
        runtimeInstanceName: gwSession.runtimeInstanceName,
        legacyManualBase: gwSession.legacyManualBase,
      }),
    [gwSession.source, gwSession.runtimeInstanceName, gwSession.legacyManualBase]
  );

  const hydratedNsRef = useRef<string | null>(null);
  const saveEnabledRef = useRef(false);

  useEffect(() => {
    hydratedNsRef.current = null;
    saveEnabledRef.current = false;
  }, [persistenceNamespace]);

  useEffect(() => {
    if (agents === null) return;
    if (hydratedNsRef.current === persistenceNamespace) return;
    hydratedNsRef.current = persistenceNamespace;

    const raw = loadMultiChatSnapshot(persistenceNamespace);
    if (!raw) {
      setPastTurnBoundaryByAgent({});
      saveEnabledRef.current = true;
      return;
    }

    const boundaries: Record<string, number> = {};
    for (const a of agents) {
      const fromStore = raw.turnsByAgent[a.ID];
      if (fromStore?.length) boundaries[a.ID] = fromStore.length;
    }

    setTurnsByAgent((prev) => {
      const next = { ...prev };
      for (const a of agents) {
        const fromStore = raw.turnsByAgent[a.ID];
        if (fromStore?.length) next[a.ID] = fromStore;
      }
      return next;
    });
    setSessionIdsByAgent((prev) => ({ ...prev, ...raw.sessionIdsByAgent }));
    setPastTurnBoundaryByAgent(boundaries);
    saveEnabledRef.current = true;
  }, [persistenceNamespace, agents]);

  useEffect(() => {
    if (!saveEnabledRef.current) return;
    const t = window.setTimeout(() => {
      saveMultiChatSnapshot(persistenceNamespace, { turnsByAgent, sessionIdsByAgent });
    }, 400);
    return () => window.clearTimeout(t);
  }, [persistenceNamespace, turnsByAgent, sessionIdsByAgent]);

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

  useEffect(() => {
    const gw = resolvedGatewayBase.trim();
    if (!gw) {
      setSlashCommands(GATEWAY_CHAT_SLASH_COMMANDS_FALLBACK);
      return;
    }
    let cancelled = false;
    let httpBase: string;
    try {
      httpBase = gatewayHttpBaseFromWsBase(gw);
    } catch {
      setSlashCommands(GATEWAY_CHAT_SLASH_COMMANDS_FALLBACK);
      return;
    }
    void fetchGatewayChatSlashCommandCatalog(httpBase)
      .then((cmds) => {
        if (cancelled || cmds.length === 0) return;
        setSlashCommands(cmds);
      })
      .catch(() => {
        if (!cancelled) setSlashCommands(GATEWAY_CHAT_SLASH_COMMANDS_FALLBACK);
      });
    return () => {
      cancelled = true;
    };
  }, [resolvedGatewayBase]);

  const gatewayStatusIsLoading =
    gwSession.source === "runtime" && instances === null && !instancesErr && !!gatewayBlockReason;

  const loadAgents = useCallback(() => {
    setLoadErr(null);
    setAgents(null);
    fetchAgents()
      .then((list) => {
        setAgents(list);
        setActiveAgentId((prev) => {
          if (prev && list.some((a) => a.ID === prev)) return prev;
          return list[0]?.ID ?? null;
        });
        setTurnsByAgent((prev) => {
          const next: Record<string, TurnBlock[]> = {};
          for (const a of list) {
            if (prev[a.ID]) next[a.ID] = prev[a.ID];
          }
          return next;
        });
        setSessionIdsByAgent((prev) => {
          const next: Record<string, string> = {};
          for (const a of list) {
            if (prev[a.ID]) next[a.ID] = prev[a.ID];
          }
          return next;
        });
        setPastTurnBoundaryByAgent((prev) => {
          const next: Record<string, number> = {};
          for (const a of list) {
            if (prev[a.ID] != null) next[a.ID] = prev[a.ID]!;
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

  const activeAgent = useMemo(() => {
    if (!agents || !activeAgentId) return undefined;
    return agents.find((a) => a.ID === activeAgentId);
  }, [agents, activeAgentId]);

  const turns = activeAgentId ? (turnsByAgent[activeAgentId] ?? []) : [];

  const activeChatSessionId = useMemo(() => {
    if (!activeAgent) return null;
    return (
      sessionIdsByAgent[activeAgent.ID] ??
      buildMultiChatSessionId(sessionInstanceKey, activeAgent)
    );
  }, [activeAgent, sessionIdsByAgent, sessionInstanceKey]);

  async function send() {
    const text = draft.trim();
    const agent = activeAgent;
    if (!text || !agent || inFlight) return;

    const userRow: UserRow = {
      id: newId(),
      content: text,
      recipientsLabel: agent.Name,
    };
    const row: AssistantRow = {
      id: newId(),
      agentId: agent.ID,
      agentName: agent.Name,
      content: "",
      pending: true,
    };

    const block: TurnBlock = { user: userRow, assistants: [row] };
    const sessionId = buildMultiChatSessionId(sessionInstanceKey, agent);
    setSessionIdsByAgent((prev) => ({ ...prev, [agent.ID]: sessionId }));
    setTurnsByAgent((prev) => ({
      ...prev,
      [agent.ID]: [...(prev[agent.ID] ?? []), block],
    }));
    setDraft("");
    setInFlight(true);

    const gw = resolvedGatewayBase.trim();
    const token = gwSession.gatewayToken.trim();

    const patchAssistant = (patch: Partial<AssistantRow>) => {
      setTurnsByAgent((prev) => {
        const list = prev[agent.ID] ?? [];
        return {
          ...prev,
          [agent.ID]: list.map((b) => {
            if (!b.assistants.some((r) => r.id === row.id)) return b;
            return {
              ...b,
              assistants: b.assistants.map((r) =>
                r.id === row.id ? { ...r, ...patch } : r
              ),
            };
          }),
        };
      });
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
      setInFlight(false);
      return;
    }

    try {
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

    setInFlight(false);
  }

  return (
    <div className="container-app flex min-h-0 flex-1 flex-col py-6 pb-8">
      <div className="flex min-h-0 flex-1 flex-col gap-4 lg:flex-row lg:items-stretch">
        <Card className="flex min-h-0 w-full shrink-0 flex-col border-border/80 p-4 lg:w-72">
          <MultiAgentHandsPicker
            agents={agents}
            loadErr={loadErr}
            activeAgentId={activeAgentId}
            activeAgentName={activeAgent?.Name}
            turnsByAgent={turnsByAgent}
            onRefresh={loadAgents}
            onSelectAgent={setActiveAgentId}
          />

          <div className="mt-5 shrink-0 border-t border-border/60 pt-4 space-y-3">
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
            {activeAgent && activeChatSessionId ? (
              <div>
                <span className="text-micro text-muted-foreground mb-1 block">
                  Active hand <span className="font-mono">session_id</span>
                </span>
                <p className="text-micro font-mono text-muted-foreground break-all rounded-md border border-border/60 bg-muted/30 px-2 py-1.5">
                  {activeChatSessionId}
                </p>
              </div>
            ) : null}
          </div>
        </Card>

        <MultiAgentChatPanel
          activeAgentName={activeAgent?.Name}
          showPickHandHint={!activeAgent && !!agents && agents.length > 0}
          showEmptySessionHint={!!activeAgent && turns.length === 0}
          turns={turns}
          pastTurnBoundaryIndex={
            activeAgentId ? pastTurnBoundaryByAgent[activeAgentId] ?? 0 : 0
          }
          slashCommands={slashCommands}
          draft={draft}
          onDraftChange={setDraft}
          onSend={send}
          inFlight={inFlight}
          sendDisabled={!draft.trim() || !activeAgent || gatewayNotReady}
          gatewayBlockReason={gatewayBlockReason}
          gatewayStatusIsLoading={gatewayStatusIsLoading}
          messagePlaceholder={
            activeAgent ? `Message ${activeAgent.Name}…` : "Choose a hand on the left to start…"
          }
        />
      </div>
    </div>
  );
}
