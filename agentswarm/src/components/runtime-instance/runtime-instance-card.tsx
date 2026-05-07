import { Activity, ArrowUpRight, Globe, MessagesSquare, Server } from "lucide-react";
import type { SwarmRuntimeInstance } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type RuntimeInstanceCardProps = {
  instance: SwarmRuntimeInstance;
  /** Highlights card when this runtime is selected for hero actions. */
  selected?: boolean;
  onSelect?: () => void;
  /** Open gateway pairing, then multi-agent chat. */
  onPairAndChat?: () => void;
  className?: string;
};

function formatRelativeHeartbeat(iso: string | null | undefined): string {
  if (!iso) return "Never";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  const diffSec = Math.round((Date.now() - d.getTime()) / 1000);
  const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });
  const abs = Math.abs(diffSec);
  if (abs < 45) return rtf.format(-diffSec, "second");
  const diffMin = Math.round(diffSec / 60);
  if (Math.abs(diffMin) < 60) return rtf.format(-diffMin, "minute");
  const diffHr = Math.round(diffSec / 3600);
  if (Math.abs(diffHr) < 48) return rtf.format(-diffHr, "hour");
  const diffDay = Math.round(diffSec / 86400);
  return rtf.format(-diffDay, "day");
}

function statusChipClass(status: string): string {
  const s = status.toLowerCase();
  if (s === "online") {
    return "border-emerald-500/35 bg-emerald-500/10 text-emerald-800 dark:text-emerald-200";
  }
  if (s === "offline") {
    return "border-border/80 bg-muted/60 text-muted-foreground";
  }
  if (s === "degraded") {
    return "border-amber-500/35 bg-amber-500/10 text-amber-900 dark:text-amber-200";
  }
  return "border-border/80 bg-muted/40 text-muted-foreground";
}

export function RuntimeInstanceCard({
  instance,
  selected,
  onSelect,
  onPairAndChat,
  className,
}: RuntimeInstanceCardProps) {
  const pub = instance.PublicURL?.trim();
  const statusLower = instance.Status?.toLowerCase() ?? "unknown";
  const selectable = Boolean(onSelect);

  return (
    <article
      aria-label={`Runtime instance: ${instance.InstanceName}, ${instance.Status}`}
      className={cn(
        "group relative overflow-hidden rounded-none border border-border bg-card shadow-elevation-2",
        "transition-[box-shadow,filter] duration-300 hover:shadow-elevation-3 hover:contrast-[1.02]",
        selected && "border-primary/50 ring-2 ring-primary/25",
        className
      )}
    >
      <div
        className={cn(
          "h-1.5 w-full bg-primary",
          statusLower === "offline" && "bg-muted-foreground/35"
        )}
        aria-hidden
      />
      <div
        role={selectable ? "button" : undefined}
        tabIndex={selectable ? 0 : undefined}
        onClick={selectable ? () => onSelect?.() : undefined}
        onKeyDown={
          selectable
            ? (e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  onSelect?.();
                }
              }
            : undefined
        }
        aria-pressed={selectable ? selected : undefined}
        aria-label={selectable ? `Select ${instance.InstanceName}` : undefined}
        className={cn(
          "relative p-5 sm:p-6",
          selectable &&
            "cursor-pointer outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-card rounded-none"
        )}
      >
        <div className="relative flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 flex-1 gap-4">
            <div
              className={cn(
                "flex size-14 shrink-0 items-center justify-center rounded-none border border-primary",
                "bg-primary text-primary-foreground shadow-elevation-1"
              )}
            >
              <Server className="size-7" strokeWidth={1.75} aria-hidden />
            </div>
            <div className="min-w-0 space-y-3">
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <h3 className="text-card-title text-foreground font-medium tracking-tight">
                    {instance.InstanceName}
                  </h3>
                  <span
                    className={cn(
                      "inline-flex items-center gap-1 rounded-full border px-2 py-0.5",
                      "text-[0.6875rem] font-medium leading-tight",
                      statusChipClass(instance.Status)
                    )}
                  >
                    <Activity className="size-3 shrink-0 opacity-90" aria-hidden />
                    {instance.Status || "unknown"}
                  </span>
                </div>
                <p
                  className="text-caption font-mono text-muted-foreground mt-1.5 break-all"
                  title={instance.ID}
                >
                  {instance.ID}
                </p>
              </div>
              <div className="flex flex-wrap gap-x-4 gap-y-1 text-caption text-muted-foreground">
                <span>
                  <span className="text-micro uppercase text-muted-foreground/90">Host </span>
                  <span className="text-foreground/90">{instance.Hostname}</span>
                </span>
                <span>
                  <span className="text-micro uppercase text-muted-foreground/90">Type </span>
                  <span className="text-foreground/90">{instance.InstanceType}</span>
                </span>
                <span>
                  <span className="text-micro uppercase text-muted-foreground/90">Version </span>
                  <span className="font-mono text-[0.8125rem] text-foreground/90">
                    {instance.Version}
                  </span>
                </span>
              </div>
              <div className="flex flex-wrap items-center gap-2 text-caption">
                <span className="inline-flex items-center gap-1.5 rounded-full border border-border/80 bg-muted/50 px-3 py-1 text-foreground/90">
                  Heartbeat {formatRelativeHeartbeat(instance.LastHeartbeatAt)}
                </span>
                {pub && (
                  <a
                    href={pub}
                    target="_blank"
                    rel="noopener noreferrer"
                    onClick={(e) => e.stopPropagation()}
                    className={cn(
                      "inline-flex max-w-full items-center gap-1 rounded-full border border-border/80",
                      "bg-muted/40 px-3 py-1 text-caption-semi text-primary hover:underline",
                      "min-w-0"
                    )}
                  >
                    <Globe className="size-3.5 shrink-0" aria-hidden />
                    <span className="min-w-0 truncate">{pub}</span>
                    <ArrowUpRight className="size-3.5 shrink-0 opacity-80" aria-hidden />
                  </a>
                )}
              </div>
              {instance.Capabilities?.length ? (
                <div className="flex flex-wrap gap-1.5">
                  {instance.Capabilities.map((c) => (
                    <span
                      key={c}
                      className="rounded-md border border-border/60 bg-muted/30 px-2 py-0.5 text-[0.6875rem] text-muted-foreground"
                    >
                      {c}
                    </span>
                  ))}
                </div>
              ) : null}
            </div>
          </div>
        </div>
      </div>

      {onPairAndChat && (
        <div className="relative border-t border-border/60 bg-muted/20 px-5 py-4 sm:px-6">
          <Button type="button" size="sm" variant="secondary" className="rounded-none" onClick={onPairAndChat}>
            <MessagesSquare className="size-4" />
            Pair gateway & multi-agent chat
          </Button>
          <p className="text-micro text-muted-foreground mt-2 max-w-xl leading-relaxed">
            Generates or uses a code via{" "}
            <span className="font-mono text-[0.65rem]">GET /admin/paircode</span>, exchanges at{" "}
            <span className="font-mono text-[0.65rem]">POST /pair</span>, then opens chat with this instance selected.
          </p>
        </div>
      )}
    </article>
  );
}
