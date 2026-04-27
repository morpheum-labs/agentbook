import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import {
  fetchAgents,
  fetchCronScheduleTimeline,
  type CronScheduleTimelineResponse,
  type SwarmAgent,
} from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { CronScheduleGantt } from "@/components/cron-job/cron-schedule-gantt";
import { cn } from "@/lib/utils";

const LEVEL_ORDER = ["ReadOnly", "Supervised", "Full"] as const;

function countByAutonomy(agents: SwarmAgent[]) {
  const m = new Map<string, number>();
  for (const a of agents) {
    const k = a.AutonomyLevel || "—";
    m.set(k, (m.get(k) ?? 0) + 1);
  }
  return m;
}

export function AgentChartPage() {
  const [agents, setAgents] = useState<SwarmAgent[] | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [timeline, setTimeline] = useState<CronScheduleTimelineResponse | null>(null);
  const [timelineErr, setTimelineErr] = useState<string | null>(null);
  const [timelineLoading, setTimelineLoading] = useState(true);

  useEffect(() => {
    setErr(null);
    setAgents(null);
    fetchAgents()
      .then(setAgents)
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Failed to load");
      });
  }, []);

  useEffect(() => {
    setTimelineErr(null);
    setTimeline(null);
    setTimelineLoading(true);
    fetchCronScheduleTimeline()
      .then((t) => {
        setTimeline(t);
        setTimelineErr(null);
      })
      .catch((e: unknown) => {
        setTimelineErr(e instanceof Error ? e.message : "Failed to load cron timeline");
        setTimeline(null);
      })
      .finally(() => {
        setTimelineLoading(false);
      });
  }, []);

  const { rows, max } = useMemo(() => {
    if (!agents?.length) {
      return { rows: [] as { label: string; n: number }[], max: 0 };
    }
    const m = countByAutonomy(agents);
    const ordered: { label: string; n: number }[] = [];
    for (const l of LEVEL_ORDER) {
      const n = m.get(l);
      if (n) ordered.push({ label: l, n });
    }
    for (const [label, n] of m) {
      if (!LEVEL_ORDER.includes(label as (typeof LEVEL_ORDER)[number])) {
        ordered.push({ label, n });
      }
    }
    const maxN = Math.max(1, ...ordered.map((r) => r.n));
    return { rows: ordered, max: maxN };
  }, [agents]);

  return (
    <div className="container-app max-w-4xl py-10">
        <h2 className="text-body-heading mb-6">Agent chart</h2>
        <p className="text-body text-muted-foreground mb-6">
          Agents by autonomy level and a projected cron timeline (Clawgotcha API).
        </p>

        {err && (
          <p className="text-destructive text-body mb-4" role="alert">
            {err}
          </p>
        )}

        {agents === null && !err && (
          <p className="text-muted-foreground text-body">Loading…</p>
        )}

        {agents && agents.length === 0 && !err && (
          <p className="text-muted-foreground text-body">No agents to chart yet.</p>
        )}

        {agents && agents.length > 0 && (
          <div className="grid gap-8 lg:grid-cols-[1fr,280px]">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-body-heading">By autonomy</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-col gap-4">
                {rows.map(({ label, n }) => (
                  <div key={label}>
                    <div className="flex justify-between gap-2 text-caption mb-1">
                      <span className="text-foreground">{label}</span>
                      <span className="text-muted-foreground tabular-nums">{n}</span>
                    </div>
                    <div
                      className="h-3 w-full overflow-hidden rounded-sm bg-muted/60"
                      role="img"
                      aria-label={`${label}: ${n} agent${n === 1 ? "" : "s"}`}
                    >
                      <div
                        className={cn(
                          "h-full min-w-1 rounded-sm transition-[width] duration-300",
                          "bg-[color-mix(in_srgb,var(--amethyst-link)_88%,var(--pure-white))]",
                          "dark:bg-[color-mix(in_srgb,var(--amethyst-link)_45%,var(--dark-surface))]"
                        )}
                        style={{ width: `${(n / max) * 100}%` }}
                      />
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-body-heading">Roster</CardTitle>
              </CardHeader>
              <CardContent>
                <ul className="flex max-h-80 flex-col gap-1 overflow-y-auto text-caption">
                  {agents.map((a) => (
                    <li key={a.ID}>
                      <Link
                        to={`/agents/${a.ID}`}
                        className="text-link hover:underline break-all"
                      >
                        {a.Name}
                      </Link>
                      <span className="text-muted-foreground"> · {a.AutonomyLevel}</span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          </div>
        )}

        <div className="mt-10">
          <CronScheduleGantt
            data={timeline}
            loading={timelineLoading}
            error={timelineErr}
            className="w-full"
          />
        </div>
    </div>
  );
}
