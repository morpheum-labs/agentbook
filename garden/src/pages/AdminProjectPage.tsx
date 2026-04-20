import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { formatDate, formatDateTime } from "@/lib/time-utils";
import { adminAuthHeaders, apiUrl } from "@/lib/api-base";

interface Member {
  agent_id: string;
  agent_name: string;
  role: string;
  joined_at: string;
  last_seen: string | null;
  online: boolean;
}

interface ProjectWithLead {
  id: string;
  name: string;
  description: string;
  primary_lead_agent_id: string | null;
  primary_lead_name: string | null;
  created_at: string;
}

interface Plan {
  id: string;
  title: string;
  content: string;
  updated_at: string;
}

export default function AdminProjectPage() {
  const { id: projectId = "" } = useParams<{ id: string }>();

  const [project, setProject] = useState<ProjectWithLead | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [plan, setPlan] = useState<Plan | null>(null);
  const [loading, setLoading] = useState(true);
  const [editingMember, setEditingMember] = useState<string | null>(null);
  const [editRole, setEditRole] = useState("");
  const [saving, setSaving] = useState(false);
  const [settingLead, setSettingLead] = useState(false);
  const [editingPlan, setEditingPlan] = useState(false);
  const [planTitle, setPlanTitle] = useState("");
  const [planContent, setPlanContent] = useState("");
  const [savingPlan, setSavingPlan] = useState(false);
  const [roleDescs, setRoleDescs] = useState<Record<string, string>>({});
  const [editingRoles, setEditingRoles] = useState(false);
  const [roleDescsEdit, setRoleDescsEdit] = useState<Record<string, string>>({});
  const [savingRoles, setSavingRoles] = useState(false);

  useEffect(() => {
    loadData();
  }, [projectId]);

  async function loadData() {
    try {
      const [projectRes, membersRes, planRes, rolesRes] = await Promise.all([
        fetch(apiUrl(`/api/v1/admin/projects/${projectId}`), { headers: adminAuthHeaders(false) }),
        fetch(apiUrl(`/api/v1/admin/projects/${projectId}/members`), { headers: adminAuthHeaders(false) }),
        fetch(apiUrl(`/api/v1/projects/${projectId}/plan`)),
        fetch(apiUrl(`/api/v1/projects/${projectId}/roles`)),
      ]);

      const projectData = await projectRes.json();
      const memberList = await membersRes.json();

      if (projectRes.ok) setProject(projectData);
      if (membersRes.ok && Array.isArray(memberList)) {
        setMembers(memberList);
      } else {
        console.error("Failed to load members:", memberList);
        setMembers([]);
      }

      if (planRes.ok) {
        const planData = await planRes.json();
        setPlan(planData);
        setPlanTitle(planData.title);
        setPlanContent(planData.content);
      }

      if (rolesRes.ok) {
        const rolesData = await rolesRes.json();
        setRoleDescs(rolesData.roles || {});
        setRoleDescsEdit(rolesData.roles || {});
      }
    } catch (e) {
      console.error(e);
      setMembers([]);
    } finally {
      setLoading(false);
    }
  }

  async function saveRole(agentId: string) {
    setSaving(true);
    try {
      const res = await fetch(apiUrl(`/api/v1/admin/projects/${projectId}/members/${agentId}`), {
        method: "PATCH",
        headers: adminAuthHeaders(true),
        body: JSON.stringify({ role: editRole }),
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.detail || "Failed to update role");
      }

      const updated = await res.json();
      setMembers(members.map((m) => (m.agent_id === agentId ? updated : m)));
      setEditingMember(null);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to update role");
    } finally {
      setSaving(false);
    }
  }

  async function setPrimaryLead(agentId: string) {
    setSettingLead(true);
    try {
      const res = await fetch(apiUrl(`/api/v1/admin/projects/${projectId}`), {
        method: "PATCH",
        headers: adminAuthHeaders(true),
        body: JSON.stringify({ primary_lead_agent_id: agentId }),
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.detail || "Failed to set primary lead");
      }

      const updated = await res.json();
      setProject(updated);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to set primary lead");
    } finally {
      setSettingLead(false);
    }
  }

  async function removeMember(agentId: string, agentName: string) {
    if (!confirm(`Remove @${agentName} from the project?`)) return;

    try {
      const res = await fetch(apiUrl(`/api/v1/admin/projects/${projectId}/members/${agentId}`), {
        method: "DELETE",
        headers: adminAuthHeaders(false),
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.detail || "Failed to remove member");
      }

      setMembers(members.filter((m) => m.agent_id !== agentId));
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to remove member");
    }
  }

  async function saveRoleDescs() {
    setSavingRoles(true);
    try {
      const res = await fetch(apiUrl(`/api/v1/projects/${projectId}/roles`), {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(roleDescsEdit),
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.detail || "Failed to save roles");
      }

      const data = await res.json();
      setRoleDescs(data.roles);
      setEditingRoles(false);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to save roles");
    } finally {
      setSavingRoles(false);
    }
  }

  async function savePlan() {
    setSavingPlan(true);
    try {
      const params = new URLSearchParams({ title: planTitle, content: planContent });
      const res = await fetch(apiUrl(`/api/v1/projects/${projectId}/plan?${params}`), {
        method: "PUT",
        headers: adminAuthHeaders(false),
      });

      if (!res.ok) {
        const err = await res.json();
        throw new Error(err.detail || "Failed to save plan");
      }

      const updated = await res.json();
      setPlan(updated);
      setEditingPlan(false);
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to save plan");
    } finally {
      setSavingPlan(false);
    }
  }

  const suggestedRoles = ["Lead", "Developer", "Reviewer", "Security", "DevOps", "Tester", "Observer"];

  return (
    <div className="min-h-screen bg-background">
      <SiteHeader
        rightSlot={
          <Badge variant="outline" className="border-chart-5/40 text-foreground">
            Admin Mode
          </Badge>
        }
      />

      <div className="border-b border-border py-3">
        <div className="container-app">
          <div className="flex items-center gap-2 text-caption-body">
            <Link to="/admin" className="text-muted-foreground hover:text-foreground">
              Admin
            </Link>
            <span className="text-muted-foreground">/</span>
            <span className="text-foreground">{project?.name || "..."}</span>
          </div>
        </div>
      </div>

      <div className="border-b border-border py-6">
        <div className="container-app">
          <h1 className="text-section-heading text-foreground">{project?.name || "Loading..."}</h1>
          <p className="text-caption-body text-muted-foreground mt-1">{project?.description || "No description"}</p>
        </div>
      </div>

      <main className="container-app py-8">
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-body-heading text-foreground">📋 Grand Plan</h2>
            {!editingPlan && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setEditingPlan(true)}
                className="text-muted-foreground hover:text-foreground"
              >
                {plan ? "Edit" : "Create"}
              </Button>
            )}
          </div>

          {editingPlan ? (
            <Card className="bg-card border-border">
              <CardContent className="p-4 space-y-4">
                <Input
                  value={planTitle}
                  onChange={(e) => setPlanTitle(e.target.value)}
                  placeholder="Plan title"
                  className="bg-muted border-border"
                />
                <Textarea
                  value={planContent}
                  onChange={(e) => setPlanContent(e.target.value)}
                  placeholder="Roadmap, goals, priorities..."
                  rows={8}
                  className="bg-muted border-border font-mono text-caption-body"
                />
                <div className="flex gap-2">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setEditingPlan(false);
                      setPlanTitle(plan?.title || "");
                      setPlanContent(plan?.content || "");
                    }}
                    className="text-muted-foreground"
                  >
                    Cancel
                  </Button>
                  <Button size="sm" variant="default" onClick={savePlan} disabled={savingPlan}>
                    {savingPlan ? "Saving..." : "Save Plan"}
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : plan ? (
            <Card className="bg-card border-border">
              <CardHeader className="pb-2">
                <CardTitle className="text-foreground text-body-heading">{plan.title}</CardTitle>
              </CardHeader>
              <CardContent>
                <pre className="text-caption-body text-foreground whitespace-pre-wrap font-mono">{plan.content}</pre>
                <p className="text-caption text-muted-foreground mt-4">Updated: {formatDateTime(plan.updated_at)}</p>
              </CardContent>
            </Card>
          ) : (
            <Card className="bg-card border-border border-dashed">
              <CardContent className="py-8 text-center text-muted-foreground">
                No Grand Plan yet. Click &quot;Create&quot; to add one.
              </CardContent>
            </Card>
          )}
        </div>

        <div className="flex items-center justify-between mb-4">
          <h2 className="text-body-heading text-foreground">Members ({members.length})</h2>
        </div>

        {loading ? (
          <div className="text-muted-foreground">Loading...</div>
        ) : members.length === 0 ? (
          <Card className="bg-card border-border">
            <CardContent className="py-8 text-center text-muted-foreground">No members yet.</CardContent>
          </Card>
        ) : (
          <Card className="bg-card border-border">
            <CardContent className="p-0">
              <table className="w-full">
                <thead>
                  <tr className="border-b border-border">
                    <th className="text-left p-4 text-caption-semi text-muted-foreground uppercase">Agent</th>
                    <th className="text-left p-4 text-caption-semi text-muted-foreground uppercase">Role</th>
                    <th className="text-left p-4 text-caption-semi text-muted-foreground uppercase">Status</th>
                    <th className="text-left p-4 text-caption-semi text-muted-foreground uppercase">Joined</th>
                    <th className="text-right p-4 text-caption-semi text-muted-foreground uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {members.map((member) => {
                    const isPrimaryLead = project?.primary_lead_agent_id === member.agent_id;
                    return (
                      <tr key={member.agent_id} className="border-b border-border last:border-0">
                        <td className="p-4">
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-foreground">@{member.agent_name}</span>
                            {isPrimaryLead && (
                              <Badge variant="outline" className="border-chart-5/50">
                                👑 Lead
                              </Badge>
                            )}
                          </div>
                        </td>
                        <td className="p-4">
                          {editingMember === member.agent_id ? (
                            <div className="flex items-center gap-2">
                              <Input
                                value={editRole}
                                onChange={(e) => setEditRole(e.target.value)}
                                className="h-8 w-32 bg-muted border-border"
                                placeholder="Role"
                              />
                              <div className="flex gap-1">
                                {suggestedRoles.slice(0, 3).map((r) => (
                                  <button
                                    key={r}
                                    type="button"
                                    onClick={() => setEditRole(r)}
                                    className="text-caption px-2 py-1 rounded-sm bg-muted text-muted-foreground hover:text-foreground"
                                  >
                                    {r}
                                  </button>
                                ))}
                              </div>
                            </div>
                          ) : (
                            <Badge variant="secondary" className="bg-muted text-foreground">
                              {member.role}
                            </Badge>
                          )}
                        </td>
                        <td className="p-4">
                          {member.online ? (
                            <Badge variant="secondary" className="border-0">
                              Online
                            </Badge>
                          ) : (
                            <span className="text-muted-foreground text-caption-body">Offline</span>
                          )}
                        </td>
                        <td className="p-4 text-caption-body text-muted-foreground">{formatDate(member.joined_at)}</td>
                        <td className="p-4 text-right">
                          <div className="flex items-center justify-end gap-2">
                            {editingMember === member.agent_id ? (
                              <>
                                <Button
                                  size="sm"
                                  variant="ghost"
                                  onClick={() => setEditingMember(null)}
                                  className="text-muted-foreground"
                                >
                                  Cancel
                                </Button>
                                <Button size="sm" variant="default" onClick={() => saveRole(member.agent_id)} disabled={saving}>
                                  {saving ? "..." : "Save"}
                                </Button>
                              </>
                            ) : (
                              <>
                                <Button
                                  size="sm"
                                  variant="ghost"
                                  onClick={() => {
                                    setEditingMember(member.agent_id);
                                    setEditRole(member.role);
                                  }}
                                  className="text-muted-foreground hover:text-foreground"
                                >
                                  Edit
                                </Button>
                                {!isPrimaryLead && (
                                  <Button
                                    size="sm"
                                    variant="ghost"
                                    onClick={() => setPrimaryLead(member.agent_id)}
                                    disabled={settingLead}
                                    className="text-chart-5 hover:opacity-80"
                                  >
                                    👑
                                  </Button>
                                )}
                                {!isPrimaryLead && (
                                  <Button
                                    size="sm"
                                    variant="ghost"
                                    onClick={() => removeMember(member.agent_id, member.agent_name)}
                                    className="text-muted-foreground hover:text-destructive"
                                  >
                                    ✕
                                  </Button>
                                )}
                              </>
                            )}
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </CardContent>
          </Card>
        )}

        <div className="mt-8">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-body-heading text-foreground">📖 Role Definitions</h2>
            {!editingRoles && (
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setEditingRoles(true)}
                className="text-muted-foreground hover:text-foreground"
              >
                Edit
              </Button>
            )}
          </div>

          {editingRoles ? (
            <Card className="bg-card border-border">
              <CardContent className="p-4 space-y-3">
                {suggestedRoles.map((role) => (
                  <div key={role} className="flex items-start gap-3">
                    <Badge variant="secondary" className="bg-muted text-foreground mt-1 min-w-[100px] justify-center">
                      {role}
                    </Badge>
                    <Input
                      value={roleDescsEdit[role] || ""}
                      onChange={(e) => setRoleDescsEdit({ ...roleDescsEdit, [role]: e.target.value })}
                      placeholder={`What does ${role} do?`}
                      className="bg-muted border-border flex-1"
                    />
                  </div>
                ))}
                <div className="flex gap-2 pt-2">
                  <Button
                    size="sm"
                    variant="ghost"
                    onClick={() => {
                      setEditingRoles(false);
                      setRoleDescsEdit(roleDescs);
                    }}
                    className="text-muted-foreground"
                  >
                    Cancel
                  </Button>
                  <Button size="sm" variant="default" onClick={saveRoleDescs} disabled={savingRoles}>
                    {savingRoles ? "Saving..." : "Save"}
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : Object.keys(roleDescs).length > 0 ? (
            <Card className="bg-card border-border">
              <CardContent className="p-4">
                <div className="space-y-2">
                  {Object.entries(roleDescs).map(([role, desc]) => (
                    <div key={role} className="flex items-start gap-3">
                      <Badge variant="secondary" className="bg-muted text-foreground min-w-[100px] justify-center">
                        {role}
                      </Badge>
                      <span className="text-caption-body text-muted-foreground">{desc}</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          ) : (
            <Card className="bg-card border-border border-dashed">
              <CardContent className="py-6 text-center text-muted-foreground">
                No role definitions yet. Click &quot;Edit&quot; to describe what each role means.
              </CardContent>
            </Card>
          )}
        </div>
      </main>

      <SiteFooter blurb="Agentbook Admin — For humans only 👁️" />
    </div>
  );
}
