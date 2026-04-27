import { useEffect, useId, useMemo, useState } from "react";
import { Wrench } from "lucide-react";
import miroclawTools from "@/data/miroclaw-tools.json";
import { cn } from "@/lib/utils";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

type ToolDef = (typeof miroclawTools)["tools"][number];

const catalogByName = (() => {
  const m = new Map<string, ToolDef>();
  for (const t of miroclawTools.tools) m.set(t.name, t);
  return m;
})();

/** Heuristic: pull name='value' / name="value" examples from the reference blurb. */
function paramExamplesFromDescription(description: string): { key: string; example: string }[] {
  const re = /\b([a-zA-Z_][a-zA-Z0-9_]*)\s*=\s*['"`]([^'"`]+)['"`]/g;
  const seen = new Set<string>();
  const out: { key: string; example: string }[] = [];
  for (const match of description.matchAll(re)) {
    const k = match[1] ?? "";
    const v = match[2] ?? "";
    const u = k.toLowerCase();
    if (seen.has(u)) continue;
    seen.add(u);
    out.push({ key: k, example: v.length > 120 ? `${v.slice(0, 117)}…` : v });
    if (out.length >= 32) break;
  }
  return out;
}

type AgentHandToolsInspectorProps = {
  toolNames: string[] | undefined;
  /**
   * When a hand is selected, true if that name exists in the last `fetchAgents` payload.
   * If false, we cannot show the allowlist (stale or orphan name).
   */
  handOnRecord?: boolean;
  /** e.g. empty when no agent picked */
  disabled?: boolean;
  className?: string;
};

/**
 * When a hand (agent) is selected: lists that agent’s MiroClaw tool names, click to show
 * description + heuristically parsed parameter examples from the bundled reference JSON.
 */
export function AgentHandToolsInspector({
  toolNames,
  handOnRecord = true,
  disabled,
  className,
}: AgentHandToolsInspectorProps) {
  const listId = useId();
  const [selected, setSelected] = useState<string | null>(null);
  const names = useMemo(() => {
    const t = toolNames?.filter((n) => n.trim().length) ?? [];
    return [...t].sort((a, b) => a.localeCompare(b));
  }, [toolNames]);
  const namesKey = useMemo(() => names.join("\0"), [names]);

  useEffect(() => {
    setSelected(null);
  }, [namesKey]);

  const def = useMemo(
    () => (selected && catalogByName.get(selected)) ?? null,
    [selected]
  );

  const paramExamples = def ? paramExamplesFromDescription(def.description) : [];
  const unknown = Boolean(selected && !def);

  if (disabled) return null;
  if (!handOnRecord) {
    return (
      <div
        className={cn(
          "rounded-xl border border-amber-500/30 bg-amber-500/5 px-4 py-3 text-caption text-foreground/90",
          className
        )}
        role="status"
      >
        This hand name is not in the last agent list from the API, so the tool allowlist cannot be
        shown. Refresh, or change the name to a known hand.
      </div>
    );
  }
  if (names.length === 0) {
    return (
      <div
        className={cn(
          "rounded-xl border border-dashed border-border/70 bg-muted/20 px-4 py-3 text-caption text-muted-foreground",
          className
        )}
      >
        This hand has no tools in its MiroClaw allowlist in the control plane.
      </div>
    );
  }

  return (
    <div className={cn("space-y-3", className)} role="region" aria-label="This hand’s tools">
      <p className="text-micro text-muted-foreground">Tools for this hand (from control plane allowlist)</p>
      <div
        className="flex max-h-32 flex-wrap gap-2 overflow-y-auto rounded-lg border border-border/60 bg-background/50 p-2"
        id={listId}
        role="group"
        aria-label="Tool allowlist"
      >
        {names.map((n) => {
          const inCatalog = catalogByName.has(n);
          return (
            <Button
              key={n}
              type="button"
              variant={selected === n ? "default" : "secondary"}
              size="sm"
              onClick={() => {
                setSelected(n);
              }}
              className={cn(
                "h-8 max-w-full shrink-0 cursor-pointer rounded-md px-2.5 text-caption",
                "font-mono",
                !inCatalog && "border-amber-500/40 bg-amber-500/5"
              )}
              title={inCatalog ? "In local reference" : "Not in bundled reference — click for name only"}
              aria-pressed={selected === n}
            >
              {n}
            </Button>
          );
        })}
      </div>

      {selected && (
        <Card className="border-border/70">
          <CardContent className="p-4 pt-4 sm:p-5">
            <div className="mb-2 flex items-start justify-between gap-2">
              <h4 className="text-body-emphasis text-foreground flex min-w-0 items-center gap-2 font-mono break-all">
                <Wrench className="text-muted-foreground size-4 shrink-0" aria-hidden />
                {selected}
              </h4>
              {def && (
                <span className="text-caption text-muted-foreground shrink-0 rounded-md bg-muted/50 px-2 py-0.5">
                  {def.category}
                </span>
              )}
            </div>

            {unknown && (
              <p className="text-caption text-muted-foreground text-pretty">
                This name is on the hand’s list but is not in the local <code>miroclaw-tools</code>{" "}
                bundle, so we do not have a full description. Check the MiroClaw docs in your
                environment.
              </p>
            )}

            {def && (
              <>
                <p className="text-caption text-foreground/95 leading-relaxed text-pretty whitespace-pre-wrap">
                  {def.description}
                </p>
                {paramExamples.length > 0 && (
                  <div className="mt-4 border-t border-border/50 pt-3">
                    <p className="text-micro text-muted-foreground mb-1.5">Parameters (examples in reference text)</p>
                    <ul className="text-caption text-foreground/90 max-h-40 list-none space-y-1.5 overflow-y-auto" role="list">
                      {paramExamples.map((p) => (
                        <li key={p.key} className="flex flex-col gap-0.5 sm:flex-row sm:items-baseline sm:gap-2" role="listitem">
                          <code className="shrink-0 text-[var(--amethyst-link)]">{p.key}</code>
                          <span className="text-muted-foreground min-w-0">
                            <span className="text-border mr-1.5" aria-hidden>
                              →
                            </span>
                            {p.example}
                          </span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
                {paramExamples.length === 0 && (
                  <p className="text-caption text-muted-foreground mt-3" role="note">
                    The reference blurb has no <code>name=&apos;…&apos;</code>-style parameter examples;
                    use the full description for calling details.
                  </p>
                )}
              </>
            )}
          </CardContent>
        </Card>
      )}

      {!selected && (
        <p className="text-caption text-muted-foreground" id={`${listId}-help`} role="note">
          Click a tool to see what it does and which parameters are mentioned in the MiroClaw
          reference.
        </p>
      )}
    </div>
  );
}
