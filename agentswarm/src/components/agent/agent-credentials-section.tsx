import { useCallback, useEffect, useState } from "react";
import {
  CREDENTIAL_MATERIAL_KINDS,
  deleteAgentCredential,
  fetchAgentCredentials,
  postAgentCredential,
  rotateAgentCredential,
  type AgentCredentialBinding,
} from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { nativeSelectClass } from "@/lib/native-select-class";

type DialogMode = "closed" | "add" | "rotate";

const emptyAdd = {
  provider_slug: "",
  label: "",
  mcp_server_name: "",
  material_kind: "api_key" as string,
  secretText: "",
  jsonText: "{}",
  useJson: false,
};

export function AgentCredentialsSection({ agentId }: { agentId: string }) {
  const [rows, setRows] = useState<AgentCredentialBinding[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);
  const [dialog, setDialog] = useState<DialogMode>("closed");
  const [addForm, setAddForm] = useState(emptyAdd);
  const [rotateBindingId, setRotateBindingId] = useState<string | null>(null);
  const [rotateSecret, setRotateSecret] = useState("");
  const [rotateJson, setRotateJson] = useState("{}");
  const [rotateUseJson, setRotateUseJson] = useState(false);
  const [busy, setBusy] = useState(false);

  const reload = useCallback(async () => {
    setErr(null);
    const list = await fetchAgentCredentials(agentId);
    setRows(list);
  }, [agentId]);

  useEffect(() => {
    if (!agentId) return;
    setLoading(true);
    setErr(null);
    reload()
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Failed to load credentials");
      })
      .finally(() => setLoading(false));
  }, [agentId, reload]);

  function openAdd() {
    setAddForm({ ...emptyAdd });
    setDialog("add");
  }

  function openRotate(row: AgentCredentialBinding) {
    setRotateBindingId(row.id);
    setRotateSecret("");
    setRotateJson("{}");
    setRotateUseJson(false);
    setDialog("rotate");
  }

  async function submitAdd() {
    setBusy(true);
    setErr(null);
    try {
      let plaintext: string | Record<string, unknown>;
      if (addForm.useJson) {
        plaintext = JSON.parse(addForm.jsonText || "{}") as Record<string, unknown>;
      } else {
        plaintext = addForm.secretText;
      }
      await postAgentCredential(agentId, {
        provider_slug: addForm.provider_slug.trim(),
        label: addForm.label.trim(),
        mcp_server_name: addForm.mcp_server_name.trim() || undefined,
        material_kind: addForm.material_kind,
        plaintext,
      });
      setDialog("closed");
      await reload();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : "Create failed");
    } finally {
      setBusy(false);
    }
  }

  async function submitRotate() {
    if (!rotateBindingId) return;
    setBusy(true);
    setErr(null);
    try {
      const plaintext = rotateUseJson
        ? (JSON.parse(rotateJson || "{}") as Record<string, unknown>)
        : rotateSecret;
      await rotateAgentCredential(agentId, rotateBindingId, plaintext);
      setDialog("closed");
      await reload();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : "Rotate failed");
    } finally {
      setBusy(false);
    }
  }

  async function onDelete(row: AgentCredentialBinding) {
    if (!window.confirm(`Delete credential “${row.label}” (${row.provider_slug})?`)) return;
    setBusy(true);
    setErr(null);
    try {
      await deleteAgentCredential(agentId, row.id);
      await reload();
    } catch (e: unknown) {
      setErr(e instanceof Error ? e.message : "Delete failed");
    } finally {
      setBusy(false);
    }
  }

  function fmtTime(iso: string | null | undefined) {
    if (!iso) return "—";
    try {
      return new Date(iso).toLocaleString();
    } catch {
      return iso;
    }
  }

  return (
    <Card className="mt-8">
      <CardHeader>
        <CardTitle className="text-subheading-lg">Credentials</CardTitle>
        <CardDescription>
          API tokens and MCP-related secrets are stored encrypted on Clawgotcha. The UI never shows
          stored values — only metadata and version.{" "}
          <span className="text-muted-foreground">
            Allowed <code className="text-caption">material_kind</code> values match the Clawgotcha
            OpenAPI enum (e.g. <code className="text-caption">api_key</code>,{" "}
            <code className="text-caption">github_pat</code>, <code className="text-caption">oauth_tokens</code>
            ).
          </span>
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {err && (
          <p className="text-destructive text-body" role="alert">
            {err}
          </p>
        )}
        <div className="flex justify-end">
          <Button type="button" size="sm" onClick={openAdd} disabled={loading || busy}>
            Add credential
          </Button>
        </div>
        {loading ? (
          <p className="text-muted-foreground text-body">Loading…</p>
        ) : rows.length === 0 ? (
          <p className="text-muted-foreground text-body">No credentials yet.</p>
        ) : (
          <div className="overflow-x-auto rounded-md border border-border/60">
            <table className="w-full min-w-[640px] text-left text-caption">
              <thead className="border-b border-border/60 bg-muted/40">
                <tr>
                  <th className="p-2 font-medium">Label</th>
                  <th className="p-2 font-medium">Provider</th>
                  <th className="p-2 font-medium">MCP server</th>
                  <th className="p-2 font-medium">Kind</th>
                  <th className="p-2 font-medium">Ver.</th>
                  <th className="p-2 font-medium">Secret updated</th>
                  <th className="p-2 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {rows.map((row) => (
                  <tr key={row.id} className="border-b border-border/40 last:border-0">
                    <td className="p-2 align-top">{row.label}</td>
                    <td className="p-2 align-top">{row.provider_slug}</td>
                    <td className="p-2 align-top text-muted-foreground">
                      {row.mcp_server_name?.trim() || "—"}
                    </td>
                    <td className="p-2 align-top">{row.material_kind ?? "—"}</td>
                    <td className="p-2 align-top">{row.current_version}</td>
                    <td className="p-2 align-top text-muted-foreground">
                      {fmtTime(row.secret_updated_at ?? undefined)}
                    </td>
                    <td className="p-2 align-top text-right whitespace-nowrap">
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        className="mr-2"
                        disabled={busy}
                        onClick={() => openRotate(row)}
                      >
                        Rotate
                      </Button>
                      <Button
                        type="button"
                        variant="secondary"
                        size="sm"
                        disabled={busy}
                        onClick={() => void onDelete(row)}
                      >
                        Delete
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </CardContent>

      <Dialog open={dialog === "add"} onOpenChange={(o) => !o && setDialog("closed")}>
        <DialogContent className="max-w-lg gap-0 border-border/80 p-0">
          <DialogHeader className="border-0 px-6 pt-6">
            <DialogTitle>Add credential</DialogTitle>
            <DialogDescription>
              Single-line secrets use the password field. OAuth-style payloads use JSON (must be valid
              JSON; sent as the encrypted <code className="text-caption">plaintext</code> value).
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4 px-6 pb-2">
            <div>
              <label className="text-caption text-muted-foreground mb-1.5 block" htmlFor="cred_provider">
                Provider slug
              </label>
              <Input
                id="cred_provider"
                value={addForm.provider_slug}
                onChange={(e) => setAddForm((f) => ({ ...f, provider_slug: e.target.value }))}
                placeholder="github"
                autoComplete="off"
              />
            </div>
            <div>
              <label className="text-caption text-muted-foreground mb-1.5 block" htmlFor="cred_label">
                Label
              </label>
              <Input
                id="cred_label"
                value={addForm.label}
                onChange={(e) => setAddForm((f) => ({ ...f, label: e.target.value }))}
                placeholder="default"
                autoComplete="off"
              />
            </div>
            <div>
              <label className="text-caption text-muted-foreground mb-1.5 block" htmlFor="cred_mcp">
                MCP server name (optional)
              </label>
              <Input
                id="cred_mcp"
                value={addForm.mcp_server_name}
                onChange={(e) => setAddForm((f) => ({ ...f, mcp_server_name: e.target.value }))}
                placeholder="matches miroclaw [[mcp.servers]].name"
                autoComplete="off"
              />
            </div>
            <div>
              <span className="text-caption text-muted-foreground mb-1.5 block" id="cred_mk_label">
                Material kind
              </span>
              <select
                id="cred_mk"
                className={nativeSelectClass}
                aria-labelledby="cred_mk_label"
                value={addForm.material_kind}
                onChange={(e) => setAddForm((f) => ({ ...f, material_kind: e.target.value }))}
              >
                {CREDENTIAL_MATERIAL_KINDS.map((k) => (
                  <option key={k} value={k}>
                    {k}
                  </option>
                ))}
              </select>
            </div>
            <label className="flex items-center gap-2 text-body">
              <input
                type="checkbox"
                checked={addForm.useJson}
                onChange={(e) => setAddForm((f) => ({ ...f, useJson: e.target.checked }))}
              />
              Plaintext is JSON object
            </label>
            {addForm.useJson ? (
              <div>
                <label className="text-caption text-muted-foreground mb-1.5 block" htmlFor="cred_json">
                  JSON
                </label>
                <Textarea
                  id="cred_json"
                  value={addForm.jsonText}
                  onChange={(e) => setAddForm((f) => ({ ...f, jsonText: e.target.value }))}
                  rows={6}
                  className="font-mono text-caption"
                />
              </div>
            ) : (
              <div>
                <label className="text-caption text-muted-foreground mb-1.5 block" htmlFor="cred_secret">
                  Secret
                </label>
                <Input
                  id="cred_secret"
                  type="password"
                  value={addForm.secretText}
                  onChange={(e) => setAddForm((f) => ({ ...f, secretText: e.target.value }))}
                  autoComplete="new-password"
                />
              </div>
            )}
          </DialogBody>
          <DialogFooter className="px-6 pb-6">
            <Button type="button" variant="secondary" onClick={() => setDialog("closed")} disabled={busy}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={
                busy ||
                !addForm.provider_slug.trim() ||
                !addForm.label.trim() ||
                (!addForm.useJson && !addForm.secretText) ||
                (addForm.useJson && !addForm.jsonText.trim())
              }
              onClick={() => void submitAdd()}
            >
              {busy ? "Saving…" : "Save"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={dialog === "rotate"} onOpenChange={(o) => !o && setDialog("closed")}>
        <DialogContent className="max-w-lg gap-0 border-border/80 p-0">
          <DialogHeader className="border-0 px-6 pt-6">
            <DialogTitle>Rotate secret</DialogTitle>
            <DialogDescription>
              Replaces the encrypted value with a new version. The material kind stays the same.
            </DialogDescription>
          </DialogHeader>
          <DialogBody className="space-y-4 px-6 pb-2">
            <label className="flex items-center gap-2 text-body">
              <input
                type="checkbox"
                checked={rotateUseJson}
                onChange={(e) => setRotateUseJson(e.target.checked)}
              />
              New value is JSON object
            </label>
            {rotateUseJson ? (
              <Textarea
                value={rotateJson}
                onChange={(e) => setRotateJson(e.target.value)}
                rows={6}
                className="font-mono text-caption"
              />
            ) : (
              <Input
                type="password"
                value={rotateSecret}
                onChange={(e) => setRotateSecret(e.target.value)}
                autoComplete="new-password"
              />
            )}
          </DialogBody>
          <DialogFooter className="px-6 pb-6">
            <Button type="button" variant="secondary" onClick={() => setDialog("closed")} disabled={busy}>
              Cancel
            </Button>
            <Button
              type="button"
              disabled={busy || (!rotateUseJson && !rotateSecret) || (rotateUseJson && !rotateJson.trim())}
              onClick={() => void submitRotate()}
            >
              {busy ? "Rotating…" : "Rotate"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
