import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { postAgent, type CreateAgentRequest } from "@/lib/api";
import { AutonomyLevelNote } from "@/components/autonomy-level-note";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { nativeSelectClass } from "@/lib/native-select-class";
import { MiroclawToolsField } from "@/components/miroclaw-tools-field";

const AUTONOMY = ["ReadOnly", "Supervised", "Full"] as const;

const initial: CreateAgentRequest = {
  name: "",
  identity: "",
  soul: "",
  user_context: "",
  tools: [],
  provider: "",
  model: "",
  timeout_seconds: 60,
  autonomy_level: "ReadOnly",
};

export function AgentNewPage() {
  const navigate = useNavigate();
  const [saving, setSaving] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [form, setForm] = useState<CreateAgentRequest>(initial);

  function update<K extends keyof CreateAgentRequest>(key: K, value: CreateAgentRequest[K]) {
    setForm((f) => ({ ...f, [key]: value }));
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErr(null);
    setSaving(true);
    try {
      const created = await postAgent({
        ...form,
        tools: form.tools?.length ? form.tools : undefined,
      });
      void navigate(`/agents/${encodeURIComponent(created.ID)}`, { replace: true });
    } catch (caught: unknown) {
      setErr(caught instanceof Error ? caught.message : "Create failed");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="container-app max-w-3xl py-8">
        <Card>
          <CardHeader>
            <CardTitle className="text-subheading-lg">New agent</CardTitle>
            <CardDescription>Create a Hand, then continue editing on the next screen.</CardDescription>
          </CardHeader>
          <form onSubmit={onSubmit}>
            <CardContent className="flex flex-col gap-5">
              {err && (
                <p className="text-destructive text-body" role="alert">
                  {err}
                </p>
              )}

              <div>
                <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="new_name">
                  Name
                </label>
                <Input
                  id="new_name"
                  name="name"
                  required
                  value={form.name}
                  onChange={(e) => update("name", e.target.value)}
                  autoComplete="off"
                />
              </div>

              <p
                className="text-caption text-muted-foreground -mt-1 mb-2"
                id="new_prompt_help"
                role="note"
              >
                MiroClaw hand prompt: identity, soul, and user context (stored with === markers in the
                control plane).
              </p>
              <div className="flex flex-col gap-3">
                <div>
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="new_identity"
                  >
                    Identity
                  </label>
                  <Textarea
                    id="new_identity"
                    name="identity"
                    className="min-h-24 font-mono text-caption"
                    value={form.identity ?? ""}
                    onChange={(e) => update("identity", e.target.value)}
                    aria-describedby="new_prompt_help"
                  />
                </div>
                <div>
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="new_soul"
                  >
                    Soul
                  </label>
                  <Textarea
                    id="new_soul"
                    name="soul"
                    className="min-h-24 font-mono text-caption"
                    value={form.soul ?? ""}
                    onChange={(e) => update("soul", e.target.value)}
                    aria-describedby="new_prompt_help"
                  />
                </div>
                <div>
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="new_user_context"
                  >
                    User context
                  </label>
                  <Textarea
                    id="new_user_context"
                    name="user_context"
                    className="min-h-24 font-mono text-caption"
                    value={form.user_context ?? ""}
                    onChange={(e) => update("user_context", e.target.value)}
                    aria-describedby="new_prompt_help"
                  />
                </div>
              </div>

              <fieldset>
                <legend className="text-caption text-muted-foreground mb-2">Tools</legend>
                <MiroclawToolsField
                  id="new_tools"
                  value={form.tools}
                  onChange={(next) => setForm((f) => ({ ...f, tools: next }))}
                  disabled={saving}
                />
              </fieldset>

              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="new_provider"
                  >
                    Provider
                  </label>
                  <Input
                    id="new_provider"
                    name="provider"
                    value={form.provider ?? ""}
                    onChange={(e) => update("provider", e.target.value)}
                  />
                </div>
                <div>
                  <label className="text-caption text-muted-foreground block mb-1.5" htmlFor="new_model">
                    Model
                  </label>
                  <Input
                    id="new_model"
                    name="model"
                    value={form.model ?? ""}
                    onChange={(e) => update("model", e.target.value)}
                  />
                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label
                    className="text-caption text-muted-foreground block mb-1.5"
                    htmlFor="new_timeout_seconds"
                  >
                    Timeout (seconds)
                  </label>
                  <Input
                    id="new_timeout_seconds"
                    name="timeout_seconds"
                    type="number"
                    min={0}
                    step={1}
                    value={form.timeout_seconds ?? 0}
                    onChange={(e) =>
                      update("timeout_seconds", Number.parseInt(e.target.value, 10) || 0)
                    }
                  />
                </div>
                <div>
                  <span className="text-caption text-muted-foreground block mb-1.5" id="new_autonomy_label">
                    Autonomy
                  </span>
                  <select
                    className={nativeSelectClass}
                    aria-labelledby="new_autonomy_label"
                    aria-describedby="new_autonomy_help"
                    value={form.autonomy_level}
                    onChange={(e) => update("autonomy_level", e.target.value)}
                    required
                  >
                    {AUTONOMY.map((v) => (
                      <option key={v} value={v}>
                        {v}
                      </option>
                    ))}
                  </select>
                  <AutonomyLevelNote id="new_autonomy_help" />
                </div>
              </div>

              <div className="flex flex-wrap gap-3 pt-2">
                <Button
                  type="submit"
                  disabled={saving}
                  className="h-9 rounded-lg border-0 bg-primary text-primary-foreground hover:opacity-95 shadow-sm"
                >
                  {saving ? "Creating…" : "Create agent"}
                </Button>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={() => void navigate(-1)}
                  disabled={saving}
                  className="h-9 rounded-lg border border-border/60 bg-accent/50 text-foreground shadow-sm hover:bg-accent/80"
                >
                  Cancel
                </Button>
              </div>
            </CardContent>
          </form>
        </Card>
    </div>
  );
}
