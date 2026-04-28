import { Activity, ArrowUpRight, Globe, Server } from "lucide-react";
import type { SwarmRuntimeInstance } from "@/lib/api";
import { cn } from "@/lib/utils";

type RuntimeInstanceCardProps = {
  instance: SwarmRuntimeInstance;
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

export function RuntimeInstanceCard({ instance, className }: RuntimeInstanceCardProps) {
  const pub = instance.PublicURL?.trim();
  const statusLower = instance.Status?.toLowerCase() ?? "unknown";

  return (
    <article
      aria-label={`Runtime instance: ${instance.InstanceName}, ${instance.Status}`}
      className={cn(
        "group relative overflow-hidden rounded-2xl border border-border/80 bg-card shadow-elevation-2",
        "transition-shadow duration-300 hover:shadow-elevation-3",
        "dark:border-white/[0.08] dark:shadow-[0_8px_32px_rgba(0,0,0,0.35)]",
        className
      )}
    >
      <div
        className={cn(
          "h-1.5 w-full bg-gradient-to-r from-[var(--mysteria-purple)] via-[var(--amethyst-link)] to-[var(--lavender-glow)]",
          statusLower === "offline" &&
            "from-muted-foreground/40 via-muted-foreground/25 to-muted-foreground/10"
        )}
        aria-hidden
      />
      <div className="relative p-5 sm:p-6">
        <div
          className={cn(
            "absolute -right-16 -top-20 size-48 rounded-full opacity-[0.12]",
            "bg-gradient-to-br from-[var(--amethyst-link)] to-[var(--lavender-glow)]",
            "blur-2xl transition-opacity group-hover:opacity-20"
          )}
          aria-hidden
        />
        <div className="relative flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex min-w-0 flex-1 gap-4">
            <div
              className={cn(
                "flex size-14 shrink-0 items-center justify-center rounded-2xl",
                "bg-gradient-to-br from-[var(--mysteria-purple)] to-[var(--amethyst-link)]",
                "text-[var(--surface-hero-foreground)] shadow-elevation-1",
                "ring-1 ring-white/15 dark:ring-white/10"
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
    </article>
  );
}
