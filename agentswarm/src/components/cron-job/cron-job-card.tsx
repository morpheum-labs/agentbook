import { Link } from "react-router-dom";
import { ArrowUpRight, Bot, Clock, Timer } from "lucide-react";
import type { SwarmCronJob } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CronJobCardProps = {
  job: SwarmCronJob;
  className?: string;
};

/**
 * Dribbble-adjacent scheduler card: gradient accent, glass row, schedule chip, agent target.
 */
export function CronJobCard({ job, className }: CronJobCardProps) {
  const hasSchedule = Boolean(job.Schedule?.trim());
  return (
    <article
      className={cn(
        "group relative overflow-hidden rounded-2xl border border-border/80 bg-card shadow-elevation-2",
        "transition-shadow duration-300 hover:shadow-elevation-3",
        "dark:border-white/[0.08] dark:shadow-[0_8px_32px_rgba(0,0,0,0.35)]",
        className
      )}
    >
        <div
          className="h-1.5 w-full bg-gradient-to-r from-[var(--mysteria-purple)] via-[var(--amethyst-link)] to-[var(--lavender-glow)]"
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
                <Timer className="size-7" strokeWidth={1.75} aria-hidden />
              </div>
              <div className="min-w-0 space-y-3">
                <div>
                  <h3 className="text-card-title text-foreground font-medium tracking-tight">
                    {job.Name}
                  </h3>
                  <p
                    className="text-caption font-mono text-muted-foreground mt-1.5 break-all"
                    title={job.ID}
                  >
                    {job.ID}
                  </p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <span
                    className={cn(
                      "inline-flex max-w-full items-center gap-1.5 rounded-full border px-3 py-1",
                      "border-border/80 bg-muted/50 text-caption backdrop-blur-sm",
                      "text-foreground/90"
                    )}
                  >
                    <Clock className="text-muted-foreground size-3.5 shrink-0" aria-hidden />
                    <span
                      className={cn(
                        "min-w-0 font-mono text-[0.8125rem] leading-tight",
                        hasSchedule && "text-foreground",
                        !hasSchedule && "text-muted-foreground italic"
                      )}
                    >
                      {hasSchedule ? job.Schedule : "No schedule set"}
                    </span>
                  </span>
                </div>
                <div className="flex flex-wrap items-center gap-2 text-caption">
                  <span className="text-micro uppercase text-muted-foreground">Runs as</span>
                  <span
                    className={cn(
                      "inline-flex items-center gap-1.5 rounded-lg",
                      "border border-border/60 bg-surface-elevated/60 px-2.5 py-1",
                      "text-caption-semi text-foreground"
                    )}
                  >
                    <Bot className="text-muted-foreground size-3.5" aria-hidden />
                    {job.AgentName || "—"}
                  </span>
                  {typeof job.TimeoutSeconds === "number" && job.TimeoutSeconds > 0 && (
                    <span className="text-caption text-muted-foreground">
                      · {job.TimeoutSeconds}s cap
                    </span>
                  )}
                </div>
              </div>
            </div>
            <div className="shrink-0 sm:pt-0.5">
              <Button
                size="sm"
                variant="secondary"
                className="h-9 rounded-xl border border-border/60 bg-accent/50 shadow-sm hover:bg-accent/80"
                asChild
              >
                <Link to={`/cron-jobs/${job.ID}`} className="inline-flex items-center gap-1.5">
                  Edit
                  <ArrowUpRight className="size-4" aria-hidden />
                </Link>
              </Button>
            </div>
          </div>
        </div>
    </article>
  );
}
