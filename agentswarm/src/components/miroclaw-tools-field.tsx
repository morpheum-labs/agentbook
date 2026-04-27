import { useMemo, useState } from "react";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import miroclawTools from "@/data/miroclaw-tools.json";
import { MIROCLAW_TOOL_PRESETS } from "@/lib/miroclaw-tool-presets";
import { Button } from "@/components/ui/button";
import { Search } from "lucide-react";

type ToolDef = (typeof miroclawTools)["tools"][number];

function textToTools(text: string): string[] {
  return text
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean);
}

const catalogNamesOrdered = miroclawTools.tools.map((t) => t.name);
const catalogNameSet = new Set(catalogNamesOrdered);

/** Stable order: catalog in doc order, then any extra names in prior list order, then new extras. */
const EMPTY_TOOLS: string[] = [];

function orderToolNames(
  selected: Set<string>,
  previousOrder: string[]
): string[] {
  const fromCatalog = catalogNamesOrdered.filter((n) => selected.has(n));
  const seen = new Set(fromCatalog);
  const rest: string[] = [];
  for (const n of previousOrder) {
    if (selected.has(n) && !seen.has(n)) {
      rest.push(n);
      seen.add(n);
    }
  }
  for (const n of selected) {
    if (!seen.has(n)) {
      rest.push(n);
      seen.add(n);
    }
  }
  return [...fromCatalog, ...rest];
}

export function MiroclawToolsField({
  id: baseId,
  value,
  onChange,
  disabled = false,
}: {
  id: string;
  value: string[] | undefined;
  onChange: (next: string[]) => void;
  disabled?: boolean;
}) {
  const tools = value ?? EMPTY_TOOLS;
  const [filter, setFilter] = useState("");

  const byCategory = useMemo(() => {
    const map = new Map<string, ToolDef[]>();
    for (const t of miroclawTools.tools) {
      const c = t.category?.trim() || "Other";
      if (!map.has(c)) map.set(c, []);
      map.get(c)!.push(t);
    }
    return map;
  }, []);

  const categories = useMemo(
    () => Array.from(byCategory.keys()),
    [byCategory]
  );

  const q = filter.trim().toLowerCase();
  const matches = (t: ToolDef) => {
    if (!q) return true;
    return (
      t.name.toLowerCase().includes(q) ||
      (t.description && t.description.toLowerCase().includes(q))
    );
  };

  const customText = useMemo(
    () => tools.filter((n) => !catalogNameSet.has(n)).join("\n"),
    [tools]
  );

  const selected = useMemo(() => new Set(tools), [tools]);

  function setSelectedAndOrder(next: Set<string>) {
    onChange(orderToolNames(next, tools));
  }

  function toggleName(name: string) {
    const s = new Set(tools);
    if (s.has(name)) s.delete(name);
    else s.add(name);
    setSelectedAndOrder(s);
  }

  function applyPreset(toolsInPreset: readonly string[]) {
    const valid = toolsInPreset.filter((n) => catalogNameSet.has(n));
    onChange(orderToolNames(new Set(valid), []));
  }

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-caption text-muted-foreground">Presets</span>
        {MIROCLAW_TOOL_PRESETS.map((p) => (
          <Button
            key={p.id}
            type="button"
            variant="secondary"
            size="xs"
            title={p.title}
            disabled={disabled}
            onClick={() => applyPreset(p.tools)}
            className="border border-border/60 bg-accent/50 text-foreground shadow-sm hover:bg-accent/80"
          >
            {p.label}
          </Button>
        ))}
      </div>
      <div className="relative" role="search">
        <Search
          className="text-muted-foreground pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2"
          aria-hidden
        />
        <Input
          id={`${baseId}_filter`}
          className="pl-9"
          type="search"
          placeholder="Filter by tool name or description"
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          autoComplete="off"
          disabled={disabled}
        />
      </div>

      <div className="max-h-[min(420px,55vh)] space-y-5 overflow-y-auto rounded-sm border border-border p-3">
        {categories.map((cat) => {
          const list = (byCategory.get(cat) ?? []).filter(matches);
          if (!list.length) return null;
          return (
            <div key={cat}>
              <h4 className="text-caption text-muted-foreground mb-2 font-medium">
                {cat}
              </h4>
              <ul className="space-y-2.5" role="list">
                {list.map((t) => (
                  <li key={t.name}>
                    <label
                      className={cn(
                        "flex cursor-pointer gap-2.5",
                        disabled && "cursor-not-allowed opacity-50"
                      )}
                    >
                      <input
                        type="checkbox"
                        className="border-border text-primary focus-visible:ring-ring mt-0.5 h-4 w-4 shrink-0 rounded-sm"
                        checked={selected.has(t.name)}
                        onChange={() => !disabled && toggleName(t.name)}
                        disabled={disabled}
                        aria-describedby={
                          t.description
                            ? `${baseId}_desc_${t.name}`
                            : undefined
                        }
                      />
                      <span className="min-w-0">
                        <span className="text-body font-mono text-sm">{t.name}</span>
                        {t.description ? (
                          <span
                            id={`${baseId}_desc_${t.name}`}
                            className="text-caption text-muted-foreground mt-0.5 line-clamp-2 block"
                          >
                            {t.description}
                          </span>
                        ) : null}
                      </span>
                    </label>
                  </li>
                ))}
              </ul>
            </div>
          );
        })}
      </div>

      {q && !miroclawTools.tools.some(matches) ? (
        <p className="text-caption text-muted-foreground">No tools match the filter.</p>
      ) : null}

      <div>
        <label
          className="text-caption text-muted-foreground block mb-1.5"
          htmlFor={`${baseId}_custom`}
        >
          Additional tools (one per line, not in the list above: MCP, runtime, or
          custom names)
        </label>
        <Textarea
          id={`${baseId}_custom`}
          className="min-h-20 font-mono text-caption"
          value={customText}
          onChange={(e) => {
            const custom = textToTools(e.target.value);
            const s = new Set(selectedCatalogNames(tools).concat(custom));
            onChange(orderToolNames(s, tools));
          }}
          disabled={disabled}
        />
        <p
          className="text-caption text-muted-foreground mt-1.5"
          id={`${baseId}_ref`}
        >
          Reference: MiroClaw in-process tool registry and parameter docs (
          {miroclawTools.sourcePath}).
        </p>
      </div>
    </div>
  );
}

function selectedCatalogNames(tools: string[]): string[] {
  return catalogNamesOrdered.filter((n) => tools.includes(n));
}
