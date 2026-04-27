import { Link } from "react-router-dom";
import { Timer } from "lucide-react";
import type { CronScheduleTimelineResponse } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

type CronScheduleGanttProps = {
  data: CronScheduleTimelineResponse | null;
  loading: boolean;
  error: string | null;
  className?: string;
};

function runLeftPercent(iso: string, t0: number, t1: number): number {
  const t = new Date(iso).getTime();
  if (!Number.isFinite(t)) return 0;
  if (t1 <= t0) return 0;
  return Math.min(100, Math.max(0, ((t - t0) / (t1 - t0)) * 100));
}

function formatAxisLabel(ts: number) {
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(ts));
}

const AXIS_LABELS = 5;

/**
 * Horizontal Gantt-style preview: one row per cron job, tick marks for each
 * projected run in the server window.
 */
export function CronScheduleGantt({ data, loading, error, className }: CronScheduleGanttProps) {
  if (error) {
    return (
      <p className="text-destructive text-body" role="alert">
        {error}
      </p>
    );
  }
  if (loading || data === null) {
    return <p className="text-muted-foreground text-body">Loading cron timeline…</p>;
  }

  const t0 = new Date(data.as_of).getTime();
  const t1 = new Date(data.horizon_ends).getTime();
  const rangeOk = Number.isFinite(t0) && Number.isFinite(t1) && t1 > t0;

  const axisTicks: number[] = [];
  for (let i = 0; i < AXIS_LABELS; i++) {
    const f = AXIS_LABELS === 1 ? 0 : i / (AXIS_LABELS - 1);
    axisTicks.push(t0 + f * (t1 - t0));
  }

  return (
    <Card className={cn("border-border/80", className)}>
      <CardHeader className="pb-2">
        <CardTitle className="text-body-heading flex items-center gap-2">
          <Timer className="text-muted-foreground size-4" aria-hidden />
          Cron schedule timeline
        </CardTitle>
        <p className="text-caption text-muted-foreground">
          Projected from standard 5-field cron, anchored on{" "}
          <span className="text-foreground/90">{data.anchored_by}</span>
          {rangeOk ? (
            <>
              {" "}
              · {formatAxisLabel(t0)} → {formatAxisLabel(t1)} · {data.horizon_hours}h / max{" "}
              {data.max_runs} runs per job
            </>
          ) : null}
        </p>
      </CardHeader>
      <CardContent className="pt-0">
        {data.rows.length === 0 ? (
          <p className="text-muted-foreground text-body">No cron jobs in the API.</p>
        ) : (
          <div className="space-y-5">
            {rangeOk && (
              <div className="border-b border-border/60 pb-2 pl-[min(100%,9rem)] sm:pl-44">
                <div
                  className="relative h-5 text-[0.65rem] text-muted-foreground sm:text-xs"
                  aria-hidden
                >
                  {axisTicks.map((ts) => {
                    const left = runLeftPercent(new Date(ts).toISOString(), t0, t1);
                    return (
                      <span
                        key={ts}
                        className="absolute -translate-x-1/2 tabular-nums"
                        style={{ left: `${left}%` }}
                      >
                        {formatAxisLabel(ts)}
                      </span>
                    );
                  })}
                </div>
              </div>
            )}
            <ul className="space-y-4" role="list">
              {data.rows.map((row) => {
                const subtitle = (() => {
                  if (!row.Active) {
                    return <span className="text-muted-foreground">Paused — no projected runs</span>;
                  }
                  if (row.ParseError) {
                    return <span className="text-amber-700/90 dark:text-amber-300/90">Not a standard 5-field cron: {row.ParseError}</span>;
                  }
                  if (row.ProjectedRuns.length === 0) {
                    return (
                      <span className="text-muted-foreground">
                        No run in this window{row.ScheduleParsed ? " (or schedule is very sparse)" : ""}
                      </span>
                    );
                  }
                  return null;
                })();
                return (
                  <li key={row.ID} className="flex flex-col gap-1.5 sm:flex-row sm:items-center sm:gap-4">
                    <div className="shrink-0 sm:w-44 sm:pt-0.5">
                      <Link
                        to={`/cron-jobs/${row.ID}`}
                        className="text-link text-caption font-medium hover:underline break-words"
                      >
                        {row.Name}
                      </Link>
                      <p className="text-micro text-muted-foreground line-clamp-1 font-mono" title={row.Schedule}>
                        {row.Schedule || "—"}
                      </p>
                    </div>
                    <div className="min-w-0 flex-1">
                      <div
                        className={cn(
                          "relative h-6 w-full overflow-visible rounded-sm",
                          "border border-border/50 bg-muted/30",
                          row.Active ? "ring-0" : "opacity-70"
                        )}
                        role="img"
                        aria-label={`${row.Name}: ${row.ProjectedRuns.length} projected run${row.ProjectedRuns.length === 1 ? "" : "s"} in window`}
                      >
                        {rangeOk &&
                          row.Active &&
                          row.ProjectedRuns.map((iso) => {
                            const left = runLeftPercent(iso, t0, t1);
                            return (
                              <span
                                key={iso}
                                className="absolute top-0.5 block size-2.5 -translate-x-1/2 rounded-sm bg-primary shadow-sm ring-1 ring-border/50"
                                style={{ left: `${left}%` }}
                                title={new Date(iso).toLocaleString()}
                              />
                            );
                          })}
                        {!row.Active && (
                          <span className="text-micro text-muted-foreground/90 absolute left-1.5 top-1/2 -translate-y-1/2">
                            Inactive
                          </span>
                        )}
                      </div>
                      {subtitle}
                    </div>
                  </li>
                );
              })}
            </ul>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
