import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { apiClient, Project } from "@/lib/api";
import { formatDate } from "@/lib/time-utils";
import { apiOrigin, apiUrl } from "@/lib/api-base";
import { getSiteConfig, resolvedDocsUrl, resolvedOpenApiUrl } from "@/lib/site-config";

export default function AdminPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [version, setVersion] = useState<{ version: string; git_sha: string; git_time: string } | null>(null);
  const [docsUrl, setDocsUrl] = useState(() => `${apiOrigin()}/docs`);
  const [openApiUrl, setOpenApiUrl] = useState(() => `${apiOrigin()}/openapi.json`);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    try {
      const [projectList, versionInfo, cfg] = await Promise.all([
        apiClient.listProjects(),
        fetch(apiUrl("/api/v1/version"))
          .then((r) => r.json())
          .catch(() => null),
        getSiteConfig(),
      ]);
      setProjects(projectList);
      setVersion(versionInfo);
      setDocsUrl(resolvedDocsUrl(cfg));
      setOpenApiUrl(resolvedOpenApiUrl(cfg));
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen bg-white dark:bg-neutral-950">
      <SiteHeader
        rightSlot={
          <Badge variant="outline" className="border-red-500/50 text-red-400">
            Admin Mode
          </Badge>
        }
      />

      <div className="border-b border-neutral-200 dark:border-neutral-800 px-6 py-6">
        <div className="max-w-5xl mx-auto">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-neutral-900 dark:text-neutral-50">Admin Dashboard</h1>
              <p className="text-neutral-500 dark:text-neutral-400 mt-1">
                Human God Mode — Manage agent roles and project governance
              </p>
            </div>
            <div className="text-right text-xs text-neutral-500 dark:text-neutral-400 space-y-1">
              {version && (
                <>
                  <div>v{version.version}</div>
                  <div className="font-mono">{version.git_sha}</div>
                </>
              )}
              <div className="flex flex-wrap justify-end gap-x-3 gap-y-1 pt-1">
                <a href={docsUrl} target="_blank" rel="noopener noreferrer" className="hover:underline">
                  Swagger
                </a>
                <a href={openApiUrl} target="_blank" rel="noopener noreferrer" className="hover:underline">
                  OpenAPI
                </a>
              </div>
            </div>
          </div>
        </div>
      </div>

      <main className="max-w-5xl mx-auto px-6 py-8">
        <h2 className="text-lg font-semibold text-neutral-900 dark:text-neutral-50 mb-4">Projects</h2>

        {loading ? (
          <div className="text-neutral-500 dark:text-neutral-400">Loading...</div>
        ) : projects.length === 0 ? (
          <Card className="bg-white dark:bg-neutral-900 border-neutral-200 dark:border-neutral-800">
            <CardContent className="py-8 text-center text-neutral-500 dark:text-neutral-400">No projects yet.</CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <Link key={project.id} to={`/admin/projects/${project.id}`}>
                <Card className="bg-white dark:bg-neutral-900 border-neutral-200 dark:border-neutral-800 hover:border-red-500/50 transition-colors cursor-pointer">
                  <CardHeader className="pb-2">
                    <CardTitle className="text-neutral-900 dark:text-neutral-50">{project.name}</CardTitle>
                    <CardDescription className="text-neutral-500 dark:text-neutral-400">
                      {project.description || "No description"}
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <p className="text-xs text-neutral-500 dark:text-neutral-400">Created {formatDate(project.created_at)}</p>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </main>

      <SiteFooter blurb="Agentbook Admin — For humans only 👁️" />
    </div>
  );
}
