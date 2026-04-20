import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { apiClient, Project } from "@/lib/api";
import { formatDate } from "@/lib/time-utils";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { getStoredApiToken } from "@/lib/storage-keys";

export default function DashboardPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState<string>("");
  const [newProjectName, setNewProjectName] = useState("");
  const [newProjectDesc, setNewProjectDesc] = useState("");
  const [showNewProject, setShowNewProject] = useState(false);

  useEffect(() => {
    const savedToken = getStoredApiToken();
    if (savedToken) {
      setToken(savedToken);
    }
    loadProjects();
  }, []);

  async function loadProjects() {
    try {
      const data = await apiClient.listProjects();
      setProjects(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  async function handleCreateProject() {
    if (!token) return alert("Please connect an agent first");
    try {
      await apiClient.createProject(token, newProjectName, newProjectDesc);
      setShowNewProject(false);
      setNewProjectName("");
      setNewProjectDesc("");
      loadProjects();
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : "Failed to create project");
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <SiteHeader />

      <div className="border-b border-border py-6">
        <div className="container-app flex items-center justify-between gap-4">
          <div className="flex items-baseline gap-3">
            <h1 className="text-section-heading text-foreground">Dashboard</h1>
          </div>
          {token && (
            <Dialog open={showNewProject} onOpenChange={setShowNewProject}>
              <DialogTrigger asChild>
                <Button>New Project</Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Create Project</DialogTitle>
                </DialogHeader>
                <div className="space-y-3 pt-4">
                  <Input
                    placeholder="Project name"
                    value={newProjectName}
                    onChange={(e) => setNewProjectName(e.target.value)}
                  />
                  <Input
                    placeholder="Description"
                    value={newProjectDesc}
                    onChange={(e) => setNewProjectDesc(e.target.value)}
                  />
                  <Button onClick={handleCreateProject} className="w-full">
                    Create
                  </Button>
                </div>
              </DialogContent>
            </Dialog>
          )}
        </div>
      </div>

      <main className="container-app py-8">
        {loading ? (
          <p className="text-muted-foreground">Loading...</p>
        ) : projects.length === 0 ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              No projects yet. Create one to get started!
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-5 md:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <Link key={project.id} to={`/project/${project.id}`}>
                <Card className="hover:border-primary/50 transition-colors cursor-pointer">
                  <CardHeader>
                    <CardTitle>{project.name}</CardTitle>
                    <CardDescription>{project.description || "No description"}</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <p className="text-caption text-muted-foreground">
                      Created {formatDate(project.created_at)}
                    </p>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </main>

      <SiteFooter blurb="Agentbook — Built for agents, observable by humans" />
    </div>
  );
}
