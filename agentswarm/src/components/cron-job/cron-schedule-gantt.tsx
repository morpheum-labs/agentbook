import { useEffect, useMemo, useRef, useState } from "react";
import type { ComponentProps, ReactNode } from "react";
import { Link } from "react-router-dom";
import { Gantt, Willow } from "@svar-ui/react-gantt";
import type { IColumnConfig } from "@svar-ui/react-gantt";
import type { IConfig, ITask, IApi, IScaleConfig } from "@svar-ui/gantt-store";
import type { ICellProps } from "@svar-ui/react-grid";
import { Timer } from "lucide-react";
import type { CronScheduleTimelineResponse, CronScheduleTimelineRow } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

import "@svar-ui/react-gantt/all.css";

type TimelineLengthUnit = NonNullable<IConfig["lengthUnit"]>;

type CellRow = ITask & {
  cronName?: string;
  agentName?: string;
  scheduleText?: string;
  note?: string;
};

/**
 * Assumed time for one execution to complete (from each ProjectedRun start),
 * but never past the next scheduled run — so the chart reflects the cron’s repeat.
 */
const MAX_RUN_DURATION_MS = 5 * 60 * 1000;

function collectProjectedRunTimes(
  runs: string[],
  horizonStart: Date,
  horizonEnd: Date
): Date[] {
  const t0n = horizonStart.getTime();
  const t1n = horizonEnd.getTime();
  const seen = new Set<number>();
  const out: Date[] = [];
  for (const iso of runs) {
    const d = new Date(iso);
    if (!Number.isFinite(d.getTime())) continue;
    const t = d.getTime();
    if (t < t0n || t >= t1n || seen.has(t)) continue;
    seen.add(t);
    out.push(d);
  }
  out.sort((a, b) => a.getTime() - b.getTime());
  return out;
}

/** For each start, end = min(start + maxRunMs, nextStart, horizonEnd) so the repeat cadence is visible. */
function runSegmentsFromSchedule(
  startTimes: Date[],
  t1: Date,
  maxRunMs: number
): Partial<ITask>[] {
  const t1n = t1.getTime();
  const segments: Partial<ITask>[] = [];
  for (let i = 0; i < startTimes.length; i++) {
    const s = startTimes[i];
    const sn = s.getTime();
    const next = startTimes[i + 1];
    const capFromNext = next != null ? next.getTime() : t1n;
    const endMs = Math.min(sn + maxRunMs, capFromNext, t1n);
    if (endMs <= sn) continue;
    segments.push({ start: s, end: new Date(endMs), type: "task" });
  }
  return segments;
}

/** Must match the fixed grid column width and `IGanttColumn.width` (no flexgrow) */
const GANTT_GRID_PX = 280;

const SCALE_H_PX = 32 as const;

export type TimeScaleId = "minute" | "hour" | "day";

/**
 * `lengthUnit` / scales must match SVAR rules; we use `format` functions (not strftime) for labels.
 * User-selectable: **Min** = 1-minute grid, **Hour** = by hour, **Day** = by day (coarse).
 */
function buildTimeline(
  t0: Date,
  t1: Date,
  mode: TimeScaleId
): { scales: IScaleConfig[]; lengthUnit: TimelineLengthUnit; cellWidth: number; scaleRowCount: number } {
  if (mode === "minute") {
    return {
      lengthUnit: "minute",
      cellWidth: 3,
      scaleRowCount: 2,
      scales: [
        {
          unit: "hour",
          step: 1,
          format: (d: Date) => d.toLocaleString(undefined, { hour: "2-digit", minute: "2-digit" }),
        },
        {
          unit: "minute",
          step: 1,
          format: (d: Date) => d.getMinutes().toString().padStart(2, "0"),
        },
      ],
    };
  }

  if (mode === "hour") {
    return {
      lengthUnit: "hour",
      cellWidth: 24,
      scaleRowCount: 2,
      scales: [
        {
          unit: "day",
          step: 1,
          format: (d: Date) =>
            d.toLocaleDateString(undefined, { weekday: "short", month: "short", day: "numeric" }),
        },
        {
          unit: "hour",
          step: 1,
          format: (d: Date) => d.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" }),
        },
      ],
    };
  }

  return {
    lengthUnit: "day",
    cellWidth: 7,
    scaleRowCount: 2,
    scales: [
      {
        unit: "month",
        step: 1,
        format: (d: Date) => d.toLocaleString(undefined, { month: "short", year: "2-digit" }),
      },
      {
        unit: "day",
        step: 1,
        format: (d: Date) => d.getDate().toString(),
      },
    ],
  };
}

function buildRowTask(
  row: CronScheduleTimelineRow,
  t0: Date,
  t1: Date
): ITask {
  const scheduleText = row.Schedule || "—";
  const name = row.Name;
  const common: ITask = {
    id: row.ID,
    text: name,
    type: "task",
    parent: 0,
    cronName: name,
    agentName: row.AgentName,
    scheduleText,
  } as ITask;

  if (!row.Active) {
    return {
      ...common,
      unscheduled: true,
      start: t0,
      end: t1,
      note: "Paused — no projected runs on the chart.",
    };
  }
  if (row.ParseError) {
    return {
      ...common,
      unscheduled: true,
      start: t0,
      end: t1,
      note: `Not a standard 5-field cron: ${row.ParseError}`,
    };
  }
  if (!row.ProjectedRuns.length) {
    return {
      ...common,
      unscheduled: true,
      start: t0,
      end: t1,
      note: row.ScheduleParsed
        ? "No run in this window (or schedule is very sparse)."
        : "No run in this window.",
    };
  }

  const startTimes = collectProjectedRunTimes(row.ProjectedRuns, t0, t1);
  const segments = runSegmentsFromSchedule(startTimes, t1, MAX_RUN_DURATION_MS);

  if (segments.length === 0) {
    return {
      ...common,
      unscheduled: true,
      start: t0,
      end: t1,
      note: "No run in this window (or schedule is very sparse).",
    };
  }

  return {
    ...common,
    unscheduled: false,
    start: t0,
    end: t1,
    segments,
  };
}

type CronScheduleGanttProps = {
  data: CronScheduleTimelineResponse | null;
  loading: boolean;
  error: string | null;
  className?: string;
};

function formatHeaderRange(t0: number, t1: number) {
  return `${new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(t0))} → ${new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(t1))}`;
}

function timeScaleBlurb(mode: TimeScaleId): string {
  switch (mode) {
    case "minute":
      return "1-minute time scale (finest).";
    case "hour":
      return "Hour / day time scale (default).";
    case "day":
      return "Month / day time scale (coarsest).";
  }
}

type GanttViewProps = Omit<ComponentProps<typeof Gantt>, "init" | "readonly"> & {
  init?: (api: IApi) => void;
  readonly?: boolean;
};

/**
 * "Now" line: position tracks `Date.now()` between `t0`/`t1` (ms); re-syncs every second
 * so the marker stays on real time without re-rendering the full Gantt.
 */
function GanttCurrentTimeLine({
  t0Ms,
  t1Ms,
  leftOffsetPx,
  topOffsetPx,
}: {
  t0Ms: number;
  t1Ms: number;
  leftOffsetPx: number;
  topOffsetPx: number;
}) {
  const [nowMs, setNowMs] = useState(() => Date.now());

  useEffect(() => {
    if (t1Ms <= t0Ms) return;
    setNowMs(Date.now());
    const id = setInterval(() => {
      setNowMs(Date.now());
    }, 1000);
    return () => clearInterval(id);
  }, [t0Ms, t1Ms]);

  if (t1Ms <= t0Ms) return null;
  if (nowMs < t0Ms || nowMs > t1Ms) return null;
  const frac = (nowMs - t0Ms) / (t1Ms - t0Ms);

  return (
    <div
      className="pointer-events-none absolute z-[var(--z-page-raised)] max-md:hidden"
      style={{
        left: leftOffsetPx,
        right: 0,
        top: topOffsetPx,
        bottom: 0,
      }}
      aria-hidden
    >
      <div
        className="bg-destructive/90 absolute bottom-0 w-0.5 -translate-x-1/2 rounded-full"
        style={{
          left: `${frac * 100}%`,
          top: 0,
          boxShadow: "0 0 6px color-mix(in srgb, var(--color-destructive) 55%, transparent)",
        }}
      />
    </div>
  );
}

function agentColumn(onCell: (props: ICellProps) => ReactNode): IColumnConfig {
  return {
    id: "agent",
    header: "Job",
    flexgrow: 0,
    width: GANTT_GRID_PX,
    resize: false,
    cell: onCell,
  };
}

function renderCell(props: ICellProps) {
  const t = props.row as unknown as CellRow;
  const title = t.cronName ?? t.text ?? "";
  const to = t.id != null ? `/cron-jobs/${String(t.id)}` : null;

  return (
    <div className="flex min-w-0 flex-col gap-1 py-0.5 pr-1.5 pl-0.5">
      {to ? (
        <Link
          to={to}
          className="text-foreground/95 hover:text-foreground [overflow-wrap:anywhere] min-w-0 break-words text-left text-sm font-medium leading-snug hover:underline"
        >
          <span className="line-clamp-2" title={title}>
            {title}
          </span>
        </Link>
      ) : (
        <span
          className="text-foreground/95 [overflow-wrap:anywhere] min-w-0 break-words text-left text-sm font-medium leading-snug"
        >
          <span className="line-clamp-2" title={title}>
            {title}
          </span>
        </span>
      )}
      <div
        className="line-clamp-1 font-mono text-[0.7rem] leading-tight text-muted-foreground"
        title={t.scheduleText}
      >
        {t.scheduleText}
      </div>
      {t.agentName && (
        <p
          className="line-clamp-1 text-[0.7rem] leading-tight text-muted-foreground"
          title={t.agentName}
        >
          {t.agentName}
        </p>
      )}
      {t.note && (
        <p
          className="line-clamp-2 text-pretty text-[0.7rem] leading-snug text-amber-800/90 dark:text-amber-200/90"
          title={t.note}
        >
          {t.note}
        </p>
      )}
    </div>
  );
}

export function CronScheduleGantt({ data, loading, error, className }: CronScheduleGanttProps) {
  const ganttRef = useRef<IApi | null>(null);
  const [timeScale, setTimeScale] = useState<TimeScaleId>("hour");
  const ganttKey = data ? `${data.as_of}-${data.horizon_ends}-${timeScale}` : "empty";

  const { t0, t1, ganttConfig, timelineTopPx } = useMemo(() => {
    if (data == null) {
      return {
        t0: 0,
        t1: 0,
        ganttConfig: null as GanttViewProps | null,
        timelineTopPx: 0,
      };
    }
    const startD = new Date(data.as_of);
    const endD = new Date(data.horizon_ends);
    const t0m = startD.getTime();
    const t1m = endD.getTime();
    const rangeOk = Number.isFinite(t0m) && Number.isFinite(t1m) && t1m > t0m;
    if (!rangeOk) {
      return {
        t0: 0,
        t1: 0,
        ganttConfig: null as GanttViewProps | null,
        timelineTopPx: 0,
      };
    }

    const { scales, lengthUnit, cellWidth, scaleRowCount } = buildTimeline(startD, endD, timeScale);
    const topPx = scaleRowCount * SCALE_H_PX;

    const taskList = data.rows.map((row) => buildRowTask(row, startD, endD));

    const c: IConfig = {
      tasks: taskList,
      links: [],
      start: startD,
      end: endD,
      autoScale: false,
      lengthUnit,
      durationUnit: "hour",
      cellWidth,
      cellHeight: 76,
      scaleHeight: SCALE_H_PX,
      scales,
      columns: [agentColumn(renderCell)],
      splitTasks: true,
      baselines: false,
      unscheduledTasks: true,
      rollups: false,
      /** Lets users zoom the time axis when the window is long; `false` was forcing a hard-to-read default. */
      zoom: { minCellWidth: 4, maxCellWidth: 64 },
    };

    const gantt: GanttViewProps = { ...c, markers: [] };

    return { t0: t0m, t1: t1m, ganttConfig: gantt, timelineTopPx: topPx };
  }, [data, timeScale]);

  const rangeOk = data != null && t0 > 0 && t1 > t0;

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

  if (!ganttConfig || !rangeOk) {
    return (
      <Card className={cn("border-border/80", className)}>
        <CardHeader className="pb-2">
          <CardTitle className="text-body-heading flex items-center gap-2">
            <Timer className="text-muted-foreground size-4" aria-hidden />
            Cron schedule timeline
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-body text-muted-foreground">The timeline has no valid range from the API.</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card
      className={cn("border-border/80 overflow-x-auto overflow-y-hidden", className)}
    >
      <CardHeader className="pb-2">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between sm:gap-3">
          <div className="min-w-0">
            <CardTitle className="text-body-heading flex items-center gap-2">
              <Timer className="text-muted-foreground size-4" aria-hidden />
              Cron schedule timeline
            </CardTitle>
            <p className="text-caption text-muted-foreground mt-0.5">
              Projected from standard 5-field cron, anchored on{" "}
              <span className="text-foreground/90">{data.anchored_by}</span>
              {rangeOk && (
                <>
                  {" "}
                  · {formatHeaderRange(t0, t1)} · {data.horizon_hours}h / max {data.max_runs} runs per job.{" "}
                  {timeScaleBlurb(timeScale)}
                </>
              )}
            </p>
          </div>
          {data.rows.length > 0 && (
            <div
              className="flex shrink-0 flex-wrap items-center gap-1.5"
              role="group"
              aria-label="Time axis scale"
            >
              <span className="text-caption text-muted-foreground hidden min-[400px]:inline">Scale</span>
              {(
                [
                  { id: "minute" as const, label: "Min" },
                  { id: "hour" as const, label: "Hour" },
                  { id: "day" as const, label: "Day" },
                ] as const
              ).map(({ id, label }) => (
                <Button
                  key={id}
                  type="button"
                  size="xs"
                  variant={timeScale === id ? "default" : "secondary"}
                  className={cn(
                    "font-medium",
                    timeScale !== id &&
                      "border border-border/60 bg-accent/50 text-foreground shadow-sm hover:bg-accent/80"
                  )}
                  onClick={() => setTimeScale(id)}
                >
                  {label}
                </Button>
              ))}
            </div>
          )}
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        {data.rows.length === 0 ? (
          <p className="text-body text-muted-foreground">No cron jobs in the API.</p>
        ) : (
          <div
            className="cron-gantt-timeline relative w-full min-w-0 [color-scheme:light] dark:[color-scheme:dark]"
            style={{
              height: "min(32rem, 70dvh)",
              minHeight: 320,
            }}
          >
            <Willow fonts>
              <div className="h-full min-h-0 w-full overflow-hidden">
                <Gantt
                  key={ganttKey}
                  ref={ganttRef}
                  {...ganttConfig}
                  init={(api) => {
                    ganttRef.current = api;
                  }}
                  readonly
                  cellBorders="column"
                />
              </div>
            </Willow>
            <GanttCurrentTimeLine
              t0Ms={t0}
              t1Ms={t1}
              leftOffsetPx={GANTT_GRID_PX}
              topOffsetPx={timelineTopPx}
            />
          </div>
        )}
      </CardContent>
    </Card>
  );
}
