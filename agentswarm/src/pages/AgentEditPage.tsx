import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ArrowLeft } from "lucide-react";
import { fetchAgent, putAgent, type SwarmAgent, type PutAgentRequest } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

const AUTONOMY = ["ReadOnly", "Supervised", "Full"] as const;

function toolsToText(tools: string[] | undefined): string {
  return (tools ?? []).join("\n");
}

function textToTools(text: string): string[] {
  return text
    .split(/\r?\n/)
    .map((s) => s.trim())
    .filter(Boolean);
}

function agentToPutRequest(a: SwarmAgent): PutAgentRequest {
  return {
    name: a.Name,
    system_prompt: a.SystemPrompt,
    tools: a.Tools ?? [],
    provider: a.Provider,
    model: a.Model,
    timeout_seconds: a.TimeoutSeconds,
    autonomy_level: a.AutonomyLevel,
  };
}

export function AgentEditPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [form, setForm] = useState<PutAgentRequest | null>(null);
  const [originalId, setOriginalId] = useState("");

  useEffect(() => {
    if (!id) return;
    setErr(null);
    setLoading(true);
    setForm(null);
    setOriginalId(id);
    fetchAgent(id)
      .then((a) => {
        setForm(agentToPutRequest(a));
        setLoading(false);
      })
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Load failed");
        setLoading(false);
      });
  }, [id]);

  if (!id) {
    return (
      <p className="container-app py-8 text-destructive text-body">Missing agent id</p>
    );
  }

  const toolsText = form ? toolsToText(form.tools) : "";

  function update<K extends keyof PutAgentRequest>(key: K, value: PutAgentRequest[K]) {
    setSaved(false);
    setForm((f) => (f ? { ...f, [key]: value } : f));
  }

  async function onSave(e: React.FormEvent) {
    e.preventDefault();
    if (!form) return;
    setErr(null);
    setSaving(true);
    setSaved(false);
    try {
      await putAgent(originalId, form);
      setSaved(true);
    } catch (caught: unknown) {
      setErr(caught instanceof Error ? caught.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="min-h-screen">
      <header className="border-b border-border bg-surface-elevated/30">
        <div className="container-app flex max-w-3xl flex-col gap-4 py-6">
          <div className="flex items-center justify-between gap-4">
            <Button type="button" variant="ghost" size="sm" asChild>
              <Link to="/" className="text-nav gap-1">
                <ArrowLeft className="size-4" />
                Agents
              </Link>
            </Button>
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main className="container-app max-w-3xl py-8">
        <Card>
          <CardHeader>
            <CardTitle className="text-subheading-lg">Edit agent</CardTitle>
            <CardDescription>
              {loading ? "Loading…" : form ? `ID ${originalId}` : err ?? "—"}
            </CardDescription>
          </CardHeader>
          {err && !form && !loading && (
            <CardContent>
              <p className="text-destructive text-body" role="alert">
                {err}
              </p>
            </CardContent>
          )}

          {form && !loading && (
            <form onSubmit={onSave}>
              <CardContent className="flex flex-col gap-5">
                {err && (
                  <p className="text-destructive text-body" role="alert">
                    {err}
                  </p>
                )}
                {saved && (
                  <p className="text-body text-foreground" role="status">
                    Saved.
                  </p>
                )}

                <div>
                  <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="name">
                    Name
                  </label>
                  <Input
                    id="name"
                    name="name"
                    required
                    value={form.name}
                    onChange={(e) => update("name", e.target.value)}
                    autoComplete="off"
                  />
                </div>

                <div>
                  <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="system_prompt">
                    System prompt
                  </label>
                  <Textarea
                    id="system_prompt"
                    name="system_prompt"
                    className="min-h-32 font-mono text-caption"
                    value={form.system_prompt}
                    onChange={(e) => update("system_prompt", e.target.value)}
                  />
                </div>

                <div>
                  <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="tools">
                    Tools (one per line)
                  </label>
                  <Textarea
                    id="tools"
                    name="tools"
                    className="min-h-24 font-mono text-caption"
                    value={toolsText}
                    onChange={(e) => {
                      setSaved(false);
                      setForm((f) =>
                        f ? { ...f, tools: textToTools(e.target.value) } : f
                      );
                    }}
                  />
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div>
                    <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="provider">
                      Provider
                    </label>
                    <Input
                      id="provider"
                      name="provider"
                      value={form.provider}
                      onChange={(e) => update("provider", e.target.value)}
                    />
                  </div>
                  <div>
                    <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="model">
                      Model
                    </label>
                    <Input
                      id="model"
                      name="model"
                      value={form.model}
                      onChange={(e) => update("model", e.target.value)}
                    />
                  </div>
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div>
                    <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="timeout_seconds">
                      Timeout (seconds)
                    </label>
                    <Input
                      id="timeout_seconds"
                      name="timeout_seconds"
                      type="number"
                      min={0}
                      step={1}
                      value={form.timeout_seconds}
                      onChange={(e) =>
                        update("timeout_seconds", Number.parseInt(e.target.value, 10) || 0)
                      }
                    />
                  </div>
                  <div>
                    <span className="text-caption text-muted-foreground block mb-1.5" id="autonomy_label">
                      Autonomy
                    </span>
                    <select
                      className={cn(
                        "h-10 w-full rounded-sm border border-border bg-background px-3 text-body",
                        "shadow-elevation-0 outline-none",
                        "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]"
                      )}
                      aria-labelledby="autonomy_label"
                      value={form.autonomy_level}
                      onChange={(e) => update("autonomy_level", e.target.value)}
                    >
                      {AUTONOMY.includes(
                        form.autonomy_level as (typeof AUTONOMY)[number]
                      ) ? null : (
                        <option value={form.autonomy_level}>
                          {form.autonomy_level}
                        </option>
                      )}
                      {AUTONOMY.map((v) => (
                        <option key={v} value={v}>
                          {v}
                        </option>
                      ))}
                    </select>
                  </div>
                </div>

                <div className="flex flex-wrap gap-3 pt-2">
                  <Button type="submit" disabled={saving}>
                    {saving ? "Saving…" : "Save changes"}
                  </Button>
                  <Button type="button" variant="outline" onClick={() => navigate(-1)} disabled={saving}>
                    Cancel
                  </Button>
                </div>
              </CardContent>
            </form>
          )}
        </Card>
      </main>
    </div>
  );
}
