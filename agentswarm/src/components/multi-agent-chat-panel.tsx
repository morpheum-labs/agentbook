import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import { Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import type { GatewaySlashCommandItem } from "@/lib/gateway-ws-chat";
import { cn } from "@/lib/utils";

/** Scroll `list` so the vertical center of `option` aligns with the list viewport center; clamp `scrollTop`. */
function scrollSlashOptionToCenter(list: HTMLElement, option: HTMLElement) {
  const listRect = list.getBoundingClientRect();
  const optRect = option.getBoundingClientRect();
  const listMidY = listRect.top + listRect.height / 2;
  const optMidY = optRect.top + optRect.height / 2;
  const delta = optMidY - listMidY;
  const maxScroll = Math.max(0, list.scrollHeight - list.clientHeight);
  list.scrollTop = Math.min(maxScroll, Math.max(0, list.scrollTop + delta));
}

function slashTokenAt(
  draft: string,
  caret: number
): { lineStart: number; replaceEnd: number; slashToken: string } | null {
  const lineStart = draft.lastIndexOf("\n", Math.max(0, caret - 1)) + 1;
  const lineToCursor = draft.slice(lineStart, caret);
  const m = /^(\/[\w-]*)$/.exec(lineToCursor);
  if (!m) return null;
  return { lineStart, replaceEnd: caret, slashToken: m[1] };
}

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
  /** Gateway WebSocket slash commands (from GET /api/chat-slash-commands or static fallback). */
  slashCommands: GatewaySlashCommandItem[];
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
  slashCommands,
}: MultiAgentChatPanelProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const slashListRef = useRef<HTMLDivElement>(null);
  const slashOptionRefs = useRef<(HTMLButtonElement | null)[]>([]);
  const [caretPos, setCaretPos] = useState(0);
  const [slashHighlight, setSlashHighlight] = useState(0);

  const slashPicker = useMemo(() => {
    const tok = slashTokenAt(draft, caretPos);
    if (!tok || !slashCommands.length) {
      return { open: false as const, filtered: [] as GatewaySlashCommandItem[], ctx: null };
    }
    const q = tok.slashToken.slice(1).toLowerCase();
    const filtered = slashCommands.filter((c) =>
      c.name.slice(1).toLowerCase().startsWith(q)
    );
    if (!filtered.length) {
      return { open: false as const, filtered, ctx: null };
    }
    return { open: true as const, filtered, ctx: tok };
  }, [draft, caretPos, slashCommands]);

  const slashNamesKey = useMemo(
    () =>
      slashPicker.open ? slashPicker.filtered.map((c) => c.name).join("\0") : "",
    [slashPicker.open, slashPicker.filtered]
  );

  useLayoutEffect(() => {
    if (!slashPicker.open || slashPicker.filtered.length === 0) return;
    const safeIdx = Math.min(Math.max(0, slashHighlight), slashPicker.filtered.length - 1);
    const list = slashListRef.current;
    const option = slashOptionRefs.current[safeIdx];
    if (!list || !option) return;
    scrollSlashOptionToCenter(list, option);
  }, [slashHighlight, slashPicker.open, slashNamesKey]);

  useEffect(() => {
    setSlashHighlight(0);
  }, [slashPicker.ctx?.slashToken, slashPicker.filtered.length]);

  function applySlashPick(cmd: GatewaySlashCommandItem) {
    const ctx = slashPicker.ctx;
    if (!ctx) return;
    const lineStart = ctx.lineStart;
    const replaceEnd = ctx.replaceEnd;
    const before = draft.slice(0, lineStart);
    const after = draft.slice(replaceEnd);
    const insert = `${cmd.name} `;
    const next = before + insert + after;
    const newCaret = lineStart + insert.length;
    onDraftChange(next);
    requestAnimationFrame(() => {
      const el = inputRef.current;
      if (!el) return;
      el.focus();
      el.selectionStart = el.selectionEnd = newCaret;
      setCaretPos(newCaret);
    });
  }

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
        className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-4 sm:p-5"
      >
        {showPickHandHint && (
          <p className="mb-4 text-body text-muted-foreground">Click a hand in the list to open its chat.</p>
        )}
        {showEmptySessionHint && (
          <p className="mb-4 text-body text-muted-foreground">
            No messages in this session yet. Pair a gateway on{" "}
            <Link to="/instances" className="text-primary underline-offset-2 hover:underline">
              Runtime instances
            </Link>{" "}
            if needed, then send a prompt.
          </p>
        )}
        {turns.map((block, turnIndex) => (
          <div key={block.user.id}>
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
            <div className="rounded-none bg-muted/50 px-4 py-3">
              <p className="text-micro mb-1 text-muted-foreground">You → {block.user.recipientsLabel}</p>
              <p className="text-body whitespace-pre-wrap">{block.user.content}</p>
            </div>
            <div>
              {block.assistants.map((r) => (
                <div
                  key={r.id}
                  className={cn(
                    "rounded-none px-3 py-3",
                    r.error ? "bg-destructive/5" : "bg-card"
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
              "mb-3 rounded-none border px-3 py-2 text-caption-body leading-snug",
              gatewayStatusIsLoading
                ? "border-border/60 bg-muted/40 text-muted-foreground"
                : "border-destructive/30 bg-destructive/5 text-destructive"
            )}
          >
            {gatewayBlockReason}
          </p>
        )}
        <div className="relative">
          {slashPicker.open && slashPicker.filtered.length > 0 ? (
            <div
              ref={slashListRef}
              className="absolute bottom-full left-0 right-0 z-10 mb-1.5 max-h-36 overflow-y-auto rounded-none border border-border bg-[var(--terminal-html-bg)] shadow-sm divide-y divide-border"
              role="listbox"
              aria-label="Gateway slash commands"
            >
              {slashPicker.filtered.map((cmd, i) => (
                <button
                  key={cmd.name}
                  ref={(el) => {
                    slashOptionRefs.current[i] = el;
                  }}
                  type="button"
                  role="option"
                  aria-selected={i === slashHighlight}
                  className={cn(
                    "grid w-full grid-cols-[auto_minmax(0,1fr)] items-center gap-x-3 px-2 py-1 text-left text-micro transition-colors",
                    i === slashHighlight
                      ? "bg-[var(--terminal-selection)] text-[var(--terminal-selection-text)]"
                      : "bg-transparent text-foreground dark:text-zinc-400 hover:bg-[var(--terminal-selection)] hover:text-[var(--terminal-selection-text)] dark:hover:text-[var(--terminal-selection-text)]"
                  )}
                  onMouseEnter={() => setSlashHighlight(i)}
                  onMouseDown={(ev) => {
                    ev.preventDefault();
                    applySlashPick(cmd);
                  }}
                >
                  <span className="font-mono shrink-0">{cmd.name}</span>
                  <span className="min-w-0 truncate opacity-80">{cmd.description}</span>
                </button>
              ))}
            </div>
          ) : null}
          <Input
            ref={inputRef}
            value={draft}
            onChange={(e) => {
              const next = e.target.value.replace(/\r?\n/g, " ");
              onDraftChange(next);
              setCaretPos(e.target.selectionStart);
            }}
            onSelect={(e) => setCaretPos(e.currentTarget.selectionStart)}
            onClick={(e) => setCaretPos(e.currentTarget.selectionStart)}
            onKeyUp={(e) => setCaretPos(e.currentTarget.selectionStart)}
            placeholder={messagePlaceholder}
            disabled={inFlight || !activeAgentName}
            className="rounded-none border-x-0 border-y border-border px-0 shadow-none focus:!border-x-0 focus-visible:!border-x-0 aria-invalid:focus:!border-x-0 aria-invalid:focus-visible:!border-x-0"
            onKeyDown={(e) => {
              if (slashPicker.open && slashPicker.filtered.length > 0) {
                if (e.key === "ArrowDown") {
                  e.preventDefault();
                  setSlashHighlight((h) => Math.min(h + 1, slashPicker.filtered.length - 1));
                  return;
                }
                if (e.key === "ArrowUp") {
                  e.preventDefault();
                  setSlashHighlight((h) => Math.max(h - 1, 0));
                  return;
                }
                if (e.key === "Enter" || e.key === "Tab") {
                  e.preventDefault();
                  const pick = slashPicker.filtered[slashHighlight];
                  if (pick) applySlashPick(pick);
                  return;
                }
                if (e.key === "Escape") {
                  e.preventDefault();
                  const el = inputRef.current;
                  const ctx = slashPicker.ctx;
                  if (el && ctx) {
                    el.selectionStart = el.selectionEnd = ctx.lineStart;
                    setCaretPos(ctx.lineStart);
                  }
                  return;
                }
              }
              if (e.key === "Enter") {
                e.preventDefault();
                void onSend();
              }
            }}
          />
        </div>
        <div className="mt-3 flex flex-wrap items-center justify-between gap-2">
          <p className="text-micro text-muted-foreground">
            <kbd className="rounded border border-border/60 px-1">/</kbd> gateway commands ·{" "}
            <kbd className="rounded border border-border/60 px-1">Enter</kbd> send
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
