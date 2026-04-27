import { useEffect, useState } from "react";
import { fetchAgents, fetchCronJobs, type SwarmCronJob } from "@/lib/api";
import { CronJobCard } from "@/components/cron-job/cron-job-card";
import { CronJobEmptyState } from "@/components/cron-job/cron-job-empty-state";
import { CronJobListSkeleton } from "@/components/cron-job/cron-job-list-skeleton";
import { CronJobsHero } from "@/components/cron-job/cron-jobs-hero";

export function CronJobListPage() {
  const [jobs, setJobs] = useState<SwarmCronJob[] | null>(null);
  const [hasAgents, setHasAgents] = useState(true);
  const [err, setErr] = useState<string | null>(null);

  function load() {
    setErr(null);
    setJobs(null);
    void Promise.all([fetchCronJobs(), fetchAgents()])
      .then(([list, agents]) => {
        setJobs(list);
        setHasAgents(agents.length > 0);
      })
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Failed to load");
        setHasAgents(true);
      });
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <div className="container-app max-w-4xl space-y-8 py-8 sm:py-10">
        <CronJobsHero
          onRefresh={() => void load()}
          refreshDisabled={jobs === null && !err}
        />

        {err && (
          <p
            className="rounded-xl border border-destructive/30 bg-destructive/5 px-4 py-3 text-destructive text-body"
            role="alert"
          >
            {err}
          </p>
        )}

        {jobs === null && !err && <CronJobListSkeleton />}

        {jobs && jobs.length === 0 && !err && (
          <CronJobEmptyState hasAgents={hasAgents} className="mt-2" />
        )}

        {jobs && jobs.length > 0 && (
          <ul className="flex flex-col gap-4" aria-label="Cron job list">
            {jobs.map((j) => (
              <li key={j.ID}>
                <CronJobCard job={j} />
              </li>
            ))}
          </ul>
        )}
    </div>
  );
}
