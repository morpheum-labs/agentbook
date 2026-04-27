import { useEffect, useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { fetchAgents, postCronJob, type CreateOrReplaceCronJobRequest } from "@/lib/api";
import { AppHeader } from "@/components/app-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CronScheduleField } from "@/components/cron-job/cron-schedule-field";
import { CronJobPromptRichtext } from "@/components/cron-job/cron-prompt-richtext";
import { CronJobFormLayout } from "@/components/cron-job/cron-job-form-layout";
import { CronJobFieldGroup } from "@/components/cron-job/cron-job-field-group";
import { cronJobSelectClass } from "@/components/cron-job/cron-job-ui";

const initial: CreateOrReplaceCronJobRequest = {
  name: "",
  agent_name: "",
  schedule: "",
  timeout_seconds: 0,
  prompt: "",
};

export function CronJobNewPage() {
  const navigate = useNavigate();
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [form, setForm] = useState<CreateOrReplaceCronJobRequest>(initial);
  const [agentNames, setAgentNames] = useState<string[]>([]);
  const [loadingAgents, setLoadingAgents] = useState(true);

  useEffect(() => {
    setLoadingAgents(true);
    fetchAgents()
      .then((a) => setAgentNames(a.map((x) => x.Name).sort((x, y) => x.localeCompare(y))))
      .catch(() => setAgentNames([]))
      .finally(() => setLoadingAgents(false));
  }, []);

  function update<K extends keyof CreateOrReplaceCronJobRequest>(
    key: K,
    value: CreateOrReplaceCronJobRequest[K]
  ) {
    setForm((f) => ({ ...f, [key]: value }));
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.agent_name.trim()) {
      setErr("Choose a target agent");
      return;
    }
    setErr(null);
    setSaving(true);
    try {
      const created = await postCronJob({
        name: form.name.trim(),
        agent_name: form.agent_name.trim(),
        schedule: form.schedule?.trim() || undefined,
        timeout_seconds: form.timeout_seconds,
        prompt: form.prompt,
      });
      void navigate(`/cron-jobs/${encodeURIComponent(created.ID)}`, { replace: true });
    } catch (caught: unknown) {
      setErr(caught instanceof Error ? caught.message : "Create failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="min-h-screen">
      <AppHeader maxWidthClassName="max-w-5xl" />

      <CronJobFormLayout
        title="New cron job"
        description="The runner lives outside this UI — you only define the schedule, agent, and prompt the worker should use."
        headerExtra={
          err ? (
            <p className="text-destructive text-caption pt-1" role="alert">
              {err}
            </p>
          ) : null
        }
      >
        {loadingAgents ? (
          <p className="text-muted-foreground text-body py-2">Loading agents…</p>
        ) : agentNames.length === 0 ? (
          <p className="text-body">
            <span className="text-destructive">No agents in the control plane.</span>{" "}
            <Link to="/agents/new" className="text-link font-medium hover:underline underline-offset-2">
              Create a Hand
            </Link>{" "}
            first, then return here.
          </p>
        ) : null}

        <form onSubmit={onSubmit} className="flex flex-col gap-6">
          {agentNames.length > 0 && (
            <>
              <CronJobFieldGroup
                label="What & who"
                description="Pick a display name and the agent that will execute the job."
              >
                <div>
                  <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="cj_name">
                    Job name
                  </label>
                  <Input
                    id="cj_name"
                    name="name"
                    required
                    value={form.name}
                    onChange={(e) => update("name", e.target.value)}
                    autoComplete="off"
                    disabled={saving}
                    className="h-10 rounded-md"
                    placeholder="e.g. Nightly triage"
                  />
                </div>
                <div>
                  <span
                    className="text-caption text-muted-foreground block mb-1.5"
                    id="cj_agent_label"
                  >
                    Target hand
                  </span>
                  <select
                    className={cronJobSelectClass}
                    aria-labelledby="cj_agent_label"
                    value={form.agent_name}
                    onChange={(e) => update("agent_name", e.target.value)}
                    required
                    disabled={saving}
                  >
                    <option value="">Select agent…</option>
                    {agentNames.map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                </div>
              </CronJobFieldGroup>

              <CronJobFieldGroup
                label="Prompt"
                description="The instruction payload when the schedule fires. Keep it scoped to one goal. Backtick text like `tool_name` is highlighted in green in the editor."
              >
                <div>
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="cj_prompt"
                  >
                    Prompt text
                  </label>
                  <CronJobPromptRichtext
                    id="cj_prompt"
                    name="prompt"
                    value={form.prompt ?? ""}
                    onChange={(e) => update("prompt", e.target.value)}
                    disabled={saving}
                    placeholder="You are the nightly triage hand. Use `file_read` and …"
                  />
                </div>
              </CronJobFieldGroup>

              <CronJobFieldGroup
                label="When & budget"
                description="Schedule is opaque to the API — use whatever your executor expects (cron, RRULE, or a tag)."
              >
                <div>
                  <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="cj_schedule">
                    Schedule
                  </label>
                  <CronScheduleField
                    id="cj_schedule"
                    name="schedule"
                    value={form.schedule ?? ""}
                    onChange={(v) => update("schedule", v)}
                    disabled={saving}
                  />
                  <p className="text-micro text-muted-foreground mt-1.5" role="note">
                    Click the field to open the cron builder (presets, 5 fields, or a custom label).
                  </p>
                </div>
                <div className="max-w-xs">
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="cj_timeout_seconds"
                  >
                    Timeout (seconds)
                  </label>
                  <Input
                    id="cj_timeout_seconds"
                    name="timeout_seconds"
                    type="number"
                    min={0}
                    step={1}
                    value={form.timeout_seconds ?? 0}
                    onChange={(e) =>
                      update("timeout_seconds", Number.parseInt(e.target.value, 10) || 0)
                    }
                    disabled={saving}
                    className="h-10 rounded-md"
                  />
                </div>
              </CronJobFieldGroup>
            </>
          )}

          <div className="flex flex-wrap gap-2 border-t border-border/60 pt-5">
            <Button type="submit" disabled={saving || agentNames.length === 0} className="rounded-xl">
              {saving ? "Creating…" : "Create job"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => void navigate(-1)}
              disabled={saving}
              className="rounded-xl"
            >
              Cancel
            </Button>
          </div>
        </form>
      </CronJobFormLayout>
    </div>
  );
}
