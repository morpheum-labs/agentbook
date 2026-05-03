import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { fetchAgent, putAgent, type SwarmAgent, type PutAgentRequest } from "@/lib/api";
import { AgentCredentialsSection } from "@/components/agent/agent-credentials-section";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { AutonomyLevelNote } from "@/components/autonomy-level-note";
import { nativeSelectClass } from "@/lib/native-select-class";
import { MiroclawToolsField } from "@/components/miroclaw-tools-field";

const AUTONOMY = ["ReadOnly", "Supervised", "Full"] as const;

function agentToPutRequest(a: SwarmAgent): PutAgentRequest {
  return {
    name: a.Name,
    identity: a.identity ?? "",
    soul: a.soul ?? "",
    user_context: a.user_context ?? "",
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
  const [modularFromApi, setModularFromApi] = useState(true);
  const [originalId, setOriginalId] = useState("");

  useEffect(() => {
    if (!id) return;
    setErr(null);
    setLoading(true);
    setForm(null);
    setOriginalId(id);
    fetchAgent(id)
      .then((a) => {
        setModularFromApi(a.modular_prompt !== false);
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
      const updated = await putAgent(originalId, form);
      setModularFromApi(updated.modular_prompt !== false);
      setSaved(true);
    } catch (caught: unknown) {
      setErr(caught instanceof Error ? caught.message : "Save failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="container-app max-w-3xl py-8">
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
                  <p
                    className="text-caption text-muted-foreground mb-2"
                    id="prompt_parts_help"
                    role="note"
                  >
                    {modularFromApi
                      ? "MiroClaw hand prompt in three parts (also stored with === markers in SystemPrompt in the API)."
                      : "This hand used a non-modular system prompt. The full text is in User context; saving rewrites to modular format."}
                  </p>
                  <div className="flex flex-col gap-3">
                    <div>
                      <label
                        className="text-caption text-muted-foreground block mb-1.5"
                        htmlFor="identity"
                      >
                        Identity
                      </label>
                      <Textarea
                        id="identity"
                        name="identity"
                        className="min-h-24 font-mono text-caption"
                        value={form.identity}
                        onChange={(e) => update("identity", e.target.value)}
                        aria-describedby="prompt_parts_help"
                      />
                    </div>
                    <div>
                      <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="soul">
                        Soul
                      </label>
                      <Textarea
                        id="soul"
                        name="soul"
                        className="min-h-24 font-mono text-caption"
                        value={form.soul}
                        onChange={(e) => update("soul", e.target.value)}
                        aria-describedby="prompt_parts_help"
                      />
                    </div>
                    <div>
                      <label
                        className="text-caption text-muted-foreground block mb-1.5"
                        htmlFor="user_context"
                      >
                        User context
                      </label>
                      <Textarea
                        id="user_context"
                        name="user_context"
                        className="min-h-24 font-mono text-caption"
                        value={form.user_context}
                        onChange={(e) => update("user_context", e.target.value)}
                        aria-describedby="prompt_parts_help"
                      />
                    </div>
                  </div>
                </div>

                <fieldset>
                  <legend className="text-caption text-muted-foreground mb-2">Tools</legend>
                  <MiroclawToolsField
                    id="edit_tools"
                    value={form.tools}
                    onChange={(next) => {
                      setSaved(false);
                      setForm((f) => (f ? { ...f, tools: next } : f));
                    }}
                    disabled={saving}
                  />
                </fieldset>

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
                      className={nativeSelectClass}
                      aria-labelledby="autonomy_label"
                      aria-describedby="edit_autonomy_help"
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
                    <AutonomyLevelNote id="edit_autonomy_help" />
                  </div>
                </div>

                <div className="flex flex-wrap gap-3 pt-2">
                  <Button
                    type="submit"
                    disabled={saving}
                    className="h-9 rounded-lg border-0 bg-primary text-primary-foreground hover:opacity-95 shadow-sm"
                  >
                    {saving ? "Saving…" : "Save changes"}
                  </Button>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => navigate(-1)}
                    disabled={saving}
                    className="h-9 rounded-lg border border-border/60 bg-accent/50 text-foreground shadow-sm hover:bg-accent/80"
                  >
                    Cancel
                  </Button>
                </div>
              </CardContent>
            </form>
          )}
        </Card>
        {form && !loading ? <AgentCredentialsSection agentId={originalId} /> : null}
    </div>
  );
}
