import { useEffect, useMemo, useState } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  deleteCronJob,
  fetchAgents,
  fetchCronJob,
  putCronJob,
  type CreateOrReplaceCronJobRequest,
  type SwarmAgent,
} from "@/lib/api";
import { AgentHandToolsInspector } from "@/components/cron-job/agent-hand-tools-inspector";
import { AppHeader } from "@/components/app-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { CronScheduleField } from "@/components/cron-job/cron-schedule-field";
import { CronJobPromptRichtext } from "@/components/cron-job/cron-prompt-richtext";
import { CronJobFormLayout } from "@/components/cron-job/cron-job-form-layout";
import { CronJobFieldGroup } from "@/components/cron-job/cron-job-field-group";
import { cronJobSelectClass } from "@/components/cron-job/cron-job-ui";

function cronToForm(j: {
  Name: string;
  AgentName: string;
  Schedule: string;
  TimeoutSeconds: number;
  Prompt: string;
  Active?: boolean;
}): CreateOrReplaceCronJobRequest {
  return {
    name: j.Name,
    agent_name: j.AgentName,
    schedule: j.Schedule,
    timeout_seconds: j.TimeoutSeconds,
    prompt: j.Prompt,
    active: j.Active !== false,
  };
}

export function CronJobEditPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [form, setForm] = useState<CreateOrReplaceCronJobRequest | null>(null);
  const [agentNames, setAgentNames] = useState<string[]>([]);
  const [agents, setAgents] = useState<SwarmAgent[]>([]);
  const [originalId, setOriginalId] = useState("");

  const agentNameOptions = useMemo(() => {
    if (!form) return agentNames;
    if (form.agent_name && !agentNames.includes(form.agent_name)) {
      return [...agentNames, form.agent_name].sort((a, b) => a.localeCompare(b));
    }
    return agentNames;
  }, [form, agentNames]);

  const selectedHand = useMemo(
    () => (form?.agent_name ? agents.find((x) => x.Name === form.agent_name) : undefined),
    [form?.agent_name, agents]
  );

  useEffect(() => {
    if (!id) return;
    setErr(null);
    setLoading(true);
    setForm(null);
    setOriginalId(id);
    void Promise.all([fetchCronJob(id), fetchAgents()])
      .then(([cj, agList]) => {
        setForm(cronToForm(cj));
        setAgents(agList);
        setAgentNames(agList.map((a) => a.Name).sort((x, y) => x.localeCompare(y)));
        const names = new Set(agList.map((a) => a.Name));
        if (cj.AgentName && !names.has(cj.AgentName)) {
          setAgentNames((prev) => [...prev, cj.AgentName].sort((x, y) => x.localeCompare(y)));
        }
        setLoading(false);
      })
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Load failed");
        setLoading(false);
      });
  }, [id]);

  if (!id) {
    return <p className="container-app py-8 text-destructive text-body">Missing job id</p>;
  }

  function update<K extends keyof CreateOrReplaceCronJobRequest>(
    key: K,
    value: CreateOrReplaceCronJobRequest[K]
  ) {
    setForm((f) => (f ? { ...f, [key]: value } : f));
  }

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    if (!form) return;
    if (!form.agent_name.trim()) {
      setErr("Choose a target agent");
      return;
    }
    setErr(null);
    setSaving(true);
    try {
      await putCronJob(originalId, {
        name: form.name.trim(),
        agent_name: form.agent_name.trim(),
        schedule: form.schedule?.trim() || undefined,
        timeout_seconds: form.timeout_seconds,
        prompt: form.prompt,
        active: form.active !== false,
      });
    } catch (caught: unknown) {
      setErr(caught instanceof Error ? caught.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  async function onDelete() {
    if (!window.confirm("Delete this cron job? This cannot be undone.")) return;
    setErr(null);
    setDeleting(true);
    try {
      await deleteCronJob(originalId);
      void navigate("/cron-jobs", { replace: true });
    } catch (caught: unknown) {
      setErr(caught instanceof Error ? caught.message : "Delete failed");
    } finally {
      setDeleting(false);
    }
  }

  if (err && !form && !loading) {
    return (
      <div className="min-h-screen">
        <AppHeader maxWidthClassName="max-w-5xl" />
        <div className="container-app max-w-5xl py-12">
          <div className="rounded-2xl border border-border bg-card p-8 text-center shadow-elevation-2">
            <p className="text-destructive text-body" role="alert">
              {err}
            </p>
            <Button className="mt-6 rounded-xl" variant="outline" asChild>
              <Link to="/cron-jobs">Back to cron jobs</Link>
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <AppHeader maxWidthClassName="max-w-5xl" />

      {loading && (
        <div className="container-app max-w-5xl py-16 text-muted-foreground text-body">Loading…</div>
      )}

      {form && !loading && (
        <CronJobFormLayout
          title="Edit cron job"
          description="Tweak the schedule, timeout, and prompt. Changes are persisted to the ClawLaundry API."
          headerExtra={
            <>
              <p className="text-caption text-muted-foreground font-mono pt-0.5 break-all">{originalId}</p>
              {err && (
                <p className="text-destructive text-caption pt-1" role="alert">
                  {err}
                </p>
              )}
            </>
          }
        >
          <form onSubmit={onSave} className="flex flex-col gap-6">
            <CronJobFieldGroup
              label="Status"
              description="When off, the job is paused—your runner should skip it until you turn it back on. The API only stores this flag."
            >
              <div className="flex items-center justify-between gap-4 rounded-lg border border-border/50 bg-background/50 px-4 py-3.5 sm:px-5">
                <div className="min-w-0 pr-2">
                  <label htmlFor="cje_active" className="text-body text-foreground font-medium">
                    Active
                  </label>
                  <p className="text-caption text-muted-foreground mt-1 text-pretty leading-snug" id="cje_active_desc">
                    Paused jobs stay in the list but should not be executed on schedule.
                  </p>
                </div>
                <Switch
                  id="cje_active"
                  checked={form.active !== false}
                  onCheckedChange={(v) => update("active", v)}
                  disabled={saving}
                  aria-describedby="cje_active_desc"
                />
              </div>
            </CronJobFieldGroup>

            <CronJobFieldGroup label="What & who" description="Job label and the Hand that executes it.">
              <div>
                <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="cje_name">
                  Job name
                </label>
                <Input
                  id="cje_name"
                  name="name"
                  required
                  value={form.name}
                  onChange={(e) => update("name", e.target.value)}
                  autoComplete="off"
                  disabled={saving}
                  className="h-10 rounded-md"
                />
              </div>
              <div>
                <span
                  className="text-caption text-muted-foreground block mb-1.5"
                  id="cje_agent_label"
                >
                  Target hand
                </span>
                <select
                  className={cronJobSelectClass}
                  aria-labelledby="cje_agent_label"
                  value={form.agent_name}
                  onChange={(e) => update("agent_name", e.target.value)}
                  required
                  disabled={saving || agentNameOptions.length === 0}
                >
                  {agentNameOptions.map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
                {form.agent_name && !agentNames.includes(form.agent_name) && (
                  <p className="text-caption text-muted-foreground mt-1" role="note">
                    This agent is not in the list; you can still save or pick another after agents sync.
                  </p>
                )}
              </div>
            </CronJobFieldGroup>

            <CronJobFieldGroup
              label="Prompt"
              description="The hand’s tool list is below. A backtick-wrapped tool name in the prompt is green when that tool is on the hand’s allowlist, and red if it is not. If the hand can’t be resolved, backtick text is not validated (all shown in the neutral green ref style)."
            >
              {form.agent_name ? (
                <div>
                  <AgentHandToolsInspector
                    key={form.agent_name}
                    handOnRecord={!!selectedHand}
                    toolNames={selectedHand?.Tools}
                  />
                </div>
              ) : (
                <p className="text-caption text-muted-foreground" role="note">
                  Select a <strong className="font-medium text-foreground/90">Target hand</strong> above
                  to show its MiroClaw allowlist and tool docs while you write the prompt.
                </p>
              )}
              <div>
                <label
                  className="text-caption text-muted-foreground block mb-1.5"
                  htmlFor="cje_prompt"
                >
                  Prompt text
                </label>
                <CronJobPromptRichtext
                  id="cje_prompt"
                  name="prompt"
                  value={form.prompt ?? ""}
                  onChange={(e) => update("prompt", e.target.value)}
                  disabled={saving}
                  allowedToolNames={selectedHand != null ? selectedHand.Tools : undefined}
                />
              </div>
            </CronJobFieldGroup>

            <CronJobFieldGroup
              label="When & budget"
              description="Schedule and timeout for your own runner — the API only stores the strings."
            >
              <div>
                <label
                  className="text-caption text-muted-foreground block mb-1.5"
                  htmlFor="cje_schedule"
                >
                  Schedule
                </label>
                <CronScheduleField
                  id="cje_schedule"
                  name="schedule"
                  value={form.schedule ?? ""}
                  onChange={(v) => update("schedule", v)}
                  disabled={saving}
                />
                <p className="text-micro text-muted-foreground mt-1.5" role="note">
                  Click the field to edit in the schedule builder.
                </p>
              </div>
              <div className="max-w-xs">
                <label
                  className="text-caption text-muted-foreground block mb-1.5"
                  htmlFor="cje_timeout"
                >
                  Timeout (seconds)
                </label>
                <Input
                  id="cje_timeout"
                  name="timeout_seconds"
                  type="number"
                  min={0}
                  step={1}
                  value={form.timeout_seconds ?? 0}
                  onChange={(e) => update("timeout_seconds", Number.parseInt(e.target.value, 10) || 0)}
                  disabled={saving}
                  className="h-10 rounded-md"
                />
              </div>
            </CronJobFieldGroup>

            <div className="flex flex-wrap gap-2 border-t border-border/60 pt-5">
              <Button type="submit" disabled={saving || deleting} className="rounded-xl">
                {saving ? "Saving…" : "Save changes"}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  if (!id) return;
                  void fetchCronJob(id).then((cj) => setForm(cronToForm(cj)));
                }}
                disabled={saving || deleting}
                className="rounded-xl"
              >
                Revert
              </Button>
            </div>
          </form>
          <div className="border-t border-border/50 mt-6 border-dashed pt-5">
            <p className="text-micro text-muted-foreground mb-3">Danger zone</p>
            <Button
              type="button"
              variant="destructive"
              onClick={() => void onDelete()}
              disabled={saving || deleting}
              className="rounded-xl"
            >
              {deleting ? "Deleting…" : "Delete job"}
            </Button>
          </div>
        </CronJobFormLayout>
      )}
    </div>
  );
}
