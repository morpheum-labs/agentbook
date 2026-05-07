import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import { Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";

export type AssistantRow = {
  id: string;
  agentId: string;
  agentName: string;
  content: string;
  pending: boolean;
  error?: string;
};

export type UserRow = {
  id: string;
  content: string;
  /** Snapshot of who received this broadcast when Send was pressed. */
  recipientsLabel: string;
};

export type TurnBlock = {
  user: UserRow;
  assistants: AssistantRow[];
};

export type MultiAgentChatPanelProps = {
  activeAgentName?: string;
  showPickHandHint: boolean;
  showEmptySessionHint: boolean;
  turns: TurnBlock[];
  /** First index of turns added after this page load; shows a divider before it when &gt; 0 and there are newer turns. */
  pastTurnBoundaryIndex?: number;
  draft: string;
  onDraftChange: (next: string) => void;
  onSend: () => void;
  inFlight: boolean;
  sendDisabled: boolean;
  gatewayBlockReason: string | null;
  gatewayStatusIsLoading: boolean;
  messagePlaceholder: string;
};

export function MultiAgentChatPanel({
  activeAgentName,
  showPickHandHint,
  showEmptySessionHint,
  turns,
  pastTurnBoundaryIndex = 0,
  draft,
  onDraftChange,
  onSend,
  inFlight,
  sendDisabled,
  gatewayBlockReason,
  gatewayStatusIsLoading,
  messagePlaceholder,
}: MultiAgentChatPanelProps) {
  const scrollRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = scrollRef.current;
    if (el) {
      el.scrollTop = el.scrollHeight;
    }
  }, [turns, inFlight, activeAgentName]);

  return (
    <Card className="flex min-h-0 flex-1 flex-col overflow-hidden border-border/80">
      <div className="shrink-0 border-b border-border/60 px-4 py-2.5 sm:px-5">
        <p className="text-micro text-muted-foreground">
          {activeAgentName ? (
            <>
              Session for <span className="font-medium text-foreground">{activeAgentName}</span>
            </>
          ) : (
            "Choose a hand to load its session."
          )}
        </p>
      </div>

      <div
        ref={scrollRef}
        className="min-h-0 flex-1 space-y-4 overflow-y-auto overscroll-contain p-4 sm:p-5"
      >
        {showPickHandHint && (
          <p className="text-body text-muted-foreground">Click a hand in the list to open its chat.</p>
        )}
        {showEmptySessionHint && (
          <p className="text-body text-muted-foreground">
            No messages in this session yet. Pair a gateway on{" "}
            <Link to="/instances" className="text-primary underline-offset-2 hover:underline">
              Runtime instances
            </Link>{" "}
            if needed, then send a prompt.
          </p>
        )}
        {turns.map((block, turnIndex) => (
          <div key={block.user.id} className="space-y-3">
            {pastTurnBoundaryIndex > 0 &&
              turnIndex === pastTurnBoundaryIndex &&
              turns.length > pastTurnBoundaryIndex && (
                <div
                  className="flex items-center gap-3 py-1"
                  role="separator"
                  aria-label="New messages after reload"
                >
                  <div className="h-px flex-1 bg-border/70" />
                  <span className="shrink-0 rounded-md border border-border/60 bg-muted/50 px-2.5 py-1 text-micro text-muted-foreground">
                    New this visit
                  </span>
                  <div className="h-px flex-1 bg-border/70" />
                </div>
              )}
            <div className="rounded-xl border border-border/50 bg-muted/50 px-4 py-3">
              <p className="text-micro mb-1 text-muted-foreground">You → {block.user.recipientsLabel}</p>
              <p className="text-body whitespace-pre-wrap">{block.user.content}</p>
            </div>
            <div className="space-y-3">
              {block.assistants.map((r) => (
                <div
                  key={r.id}
                  className={cn(
                    "rounded-xl border px-3 py-3",
                    r.error ? "border-destructive/40 bg-destructive/5" : "border-border/60 bg-card"
                  )}
                >
                  <p className="text-micro mb-2 font-medium text-muted-foreground">{r.agentName}</p>
                  {r.pending ? (
                    <p className="animate-pulse text-caption-body text-muted-foreground">Thinking…</p>
                  ) : r.error ? (
                    <p className="text-caption-body text-destructive">{r.error}</p>
                  ) : (
                    <p className="text-caption-body whitespace-pre-wrap">{r.content}</p>
                  )}
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="shrink-0 border-t border-border/60 bg-card p-4 sm:p-5">
        {gatewayBlockReason && (
          <p
            role={gatewayStatusIsLoading ? "status" : "alert"}
            className={cn(
              "mb-3 rounded-xl border px-3 py-2 text-caption-body leading-snug",
              gatewayStatusIsLoading
                ? "border-border/60 bg-muted/40 text-muted-foreground"
                : "border-destructive/30 bg-destructive/5 text-destructive"
            )}
          >
            {gatewayBlockReason}
          </p>
        )}
        <Textarea
          value={draft}
          onChange={(e) => onDraftChange(e.target.value)}
          placeholder={messagePlaceholder}
          rows={3}
          disabled={inFlight || !activeAgentName}
          className="min-h-[5rem] resize-y rounded-xl"
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              void onSend();
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
            disabled={inFlight || sendDisabled}
            onClick={() => void onSend()}
            className="rounded-lg"
          >
            <Send className="size-4" />
            Send
          </Button>
        </div>
      </div>
    </Card>
  );
}
