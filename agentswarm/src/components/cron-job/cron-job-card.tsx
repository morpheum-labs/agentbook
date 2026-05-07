import { Link } from "react-router-dom";
import { ArrowUpRight, Bot, CirclePause, CirclePlay, Clock, Timer } from "lucide-react";
import type { SwarmCronJob } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CronJobCardProps = {
  job: SwarmCronJob;
  className?: string;
};

/**
 * Terminal-style scheduler card: accent rule, inverted icon chip, schedule row.
 */
export function CronJobCard({ job, className }: CronJobCardProps) {
  const hasSchedule = Boolean(job.Schedule?.trim());
  const isActive = job.Active !== false;
  return (
    <article
      aria-label={`Cron job: ${job.Name}, ${isActive ? "active" : "paused"}`}
      className={cn(
        "group relative overflow-hidden rounded-none border border-border bg-card text-card-foreground shadow-elevation-2",
        "transition-[box-shadow,filter] duration-300 hover:shadow-elevation-3 hover:contrast-[1.02]",
        !isActive && "opacity-95"
      )}
    >
        <div
          className={cn(
            "h-1.5 w-full bg-primary",
            !isActive && "bg-muted-foreground/35"
          )}
          aria-hidden
        />
        <div className="relative p-5 sm:p-6">
          <div className="relative flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex min-w-0 flex-1 gap-4">
              <div
                className={cn(
                  "flex size-14 shrink-0 items-center justify-center rounded-none border border-primary",
                  "bg-primary text-primary-foreground shadow-elevation-1"
                )}
              >
                <Timer className="size-7" strokeWidth={1.75} aria-hidden />
              </div>
              <div className="min-w-0 space-y-3">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="text-card-title text-foreground font-medium tracking-tight">
                      {job.Name}
                    </h3>
                    <span
                      className={cn(
                        "inline-flex items-center gap-1 rounded-full border px-2 py-0.5",
                        "text-[0.6875rem] font-medium leading-tight",
                        isActive
                          ? "border-emerald-500/35 bg-emerald-500/10 text-emerald-800 dark:text-emerald-200"
                          : "border-border/80 bg-muted/60 text-muted-foreground"
                      )}
                    >
                      {isActive ? (
                        <>
                          <CirclePlay
                            className="size-3.5 text-emerald-600 dark:text-emerald-400/90"
                            strokeWidth={2}
                            aria-hidden
                          />
                          Active
                        </>
                      ) : (
                        <>
                          <CirclePause
                            className="size-3.5 opacity-80"
                            strokeWidth={2}
                            aria-hidden
                          />
                          Paused
                        </>
                      )}
                    </span>
                  </div>
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
                      "inline-flex items-center gap-1.5 rounded-none",
                      "border border-border/80 bg-muted/50 px-2.5 py-1",
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
              <Button size="sm" variant="secondary" asChild>
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
