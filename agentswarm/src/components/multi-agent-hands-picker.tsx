import type { SwarmAgent } from "@/lib/api";
import {
  turnBlocksHavePendingAssistant,
  type TurnBlock,
} from "@/components/multi-agent-chat-panel";
import { TerminalFxHeroDecor } from "@/components/terminal-fx-context";
import { cn } from "@/lib/utils";

export type MultiAgentHandsPickerProps = {
  agents: SwarmAgent[] | null;
  loadErr: string | null;
  activeAgentId: string | null;
  activeAgentName?: string | null;
  turnsByAgent: Record<string, TurnBlock[]>;
  onRefresh: () => void;
  onSelectAgent: (agentId: string) => void;
};

export function MultiAgentHandsPicker({
  agents,
  loadErr,
  activeAgentId,
  activeAgentName,
  turnsByAgent,
  onRefresh,
  onSelectAgent,
}: MultiAgentHandsPickerProps) {
  const refreshDisabled = agents === null && !loadErr;

  return (
    <div className="flex min-h-0 min-w-0 flex-1 flex-col">
      <div className="min-w-0 shrink-0 space-y-1">
        <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
          <h3 className="text-subheading-sm font-medium">Hands</h3>
          <button
            type="button"
            onClick={onRefresh}
            disabled={refreshDisabled}
            className={cn(
              "text-micro text-primary underline-offset-2 hover:underline",
              "disabled:pointer-events-none disabled:opacity-50"
            )}
          >
            Refresh agents
          </button>
        </div>
        {activeAgentName ? (
          <span className="text-micro text-muted-foreground block truncate" title={activeAgentName}>
            Viewing: {activeAgentName}
          </span>
        ) : (
          <span className="text-micro text-muted-foreground block">Pick a hand</span>
        )}
      </div>

      {loadErr && (
        <p className="text-destructive text-caption-body mt-3 shrink-0" role="alert">
          {loadErr}
        </p>
      )}

      {agents === null && !loadErr && (
        <p className="text-muted-foreground text-caption-body mt-4 shrink-0">Loading…</p>
      )}

      {agents && agents.length === 0 && (
        <p className="text-muted-foreground text-caption-body mt-4 shrink-0">No agents yet.</p>
      )}

      {agents && agents.length > 0 && (
        <ul
          className="mt-4 flex min-h-0 flex-1 flex-col gap-1.5 overflow-y-auto overscroll-contain pr-1"
          role="listbox"
          aria-label="Hands — click to open session"
        >
          {agents.map((a) => {
            const isActive = activeAgentId === a.ID;
            const blocks = turnsByAgent[a.ID] ?? [];
            const turnCount = blocks.length;
            const handPending = turnBlocksHavePendingAssistant(blocks);
            return (
              <li key={a.ID}>
                <button
                  type="button"
                  role="option"
                  aria-selected={isActive}
                  onClick={() => onSelectAgent(a.ID)}
                  className={cn(
                    "flex w-full flex-col items-stretch rounded-none border px-3 py-2.5 text-left transition-colors",
                    isActive
                      ? cn(
                          "terminal-fx-hero relative isolate overflow-hidden",
                          "border-primary/55 bg-primary/8 shadow-sm ring-1 ring-primary/20"
                        )
                      : "border-border/60 bg-card hover:bg-accent/35"
                  )}
                >
                  {isActive ? <TerminalFxHeroDecor /> : null}
                  <span className="relative z-10 flex min-w-0 flex-col items-stretch">
                    <span className="text-body font-medium leading-snug">{a.Name}</span>
                    {handPending ? (
                      <span className="text-micro text-muted-foreground mt-1.5">Thinking…</span>
                    ) : turnCount > 0 ? (
                      <span className="text-micro text-muted-foreground mt-1.5">
                        {turnCount} turn{turnCount === 1 ? "" : "s"} in session
                      </span>
                    ) : null}
                  </span>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
