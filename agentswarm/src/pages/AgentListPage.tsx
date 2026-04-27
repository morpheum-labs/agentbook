import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Pencil, RefreshCw } from "lucide-react";
import { fetchAgents, type SwarmAgent } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

export function AgentListPage() {
  const [agents, setAgents] = useState<SwarmAgent[] | null>(null);
  const [err, setErr] = useState<string | null>(null);

  function load() {
    setErr(null);
    setAgents(null);
    fetchAgents()
      .then(setAgents)
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Failed to load");
      });
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <div className="min-h-screen">
      <header className="border-b border-border bg-surface-elevated/30">
        <div className="container-app section-y flex max-w-4xl flex-col gap-4 py-8">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h1 className="text-section-heading text-foreground">ClawLaundry</h1>
              <p className="text-caption-body mt-1 text-muted-foreground">
                Swarm Hands — list and edit agent metadata
              </p>
            </div>
            <ThemeToggle />
          </div>
        </div>
      </header>

      <main className="container-app max-w-4xl py-10">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-body-heading">Agents</h2>
          <Button type="button" variant="outline" size="sm" onClick={load} disabled={agents === null && !err}>
            <RefreshCw className="size-4" />
            Refresh
          </Button>
        </div>

        {err && (
          <p className="text-destructive text-body mb-4" role="alert">
            {err}
          </p>
        )}

        {agents === null && !err && (
          <p className="text-muted-foreground text-body">Loading…</p>
        )}

        {agents && agents.length === 0 && (
          <p className="text-muted-foreground text-body">No agents yet.</p>
        )}

        {agents && agents.length > 0 && (
          <ul className="flex flex-col gap-3">
            {agents.map((a) => (
              <li key={a.ID}>
                <Card>
                  <CardHeader className="pb-0">
                    <div className="flex w-full items-start justify-between gap-2">
                      <div>
                        <CardTitle className="text-body-heading">{a.Name}</CardTitle>
                        <p className="text-caption text-muted-foreground font-mono mt-1 break-all">
                          {a.ID}
                        </p>
                      </div>
                      <Button variant="outline" size="sm" asChild>
                        <Link to={`/agents/${a.ID}`}>
                          <Pencil className="size-4" />
                          Edit
                        </Link>
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-2">
                    <dl className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                      <div>
                        <dt className="text-micro text-muted-foreground">Provider / model</dt>
                        <dd className={cn("text-body", (!a.Provider && !a.Model) && "text-muted-foreground")}>
                          {[a.Provider, a.Model].filter(Boolean).join(" · ") || "—"}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-micro text-muted-foreground">Autonomy</dt>
                        <dd className="text-body">{a.AutonomyLevel}</dd>
                      </div>
                    </dl>
                  </CardContent>
                </Card>
              </li>
            ))}
          </ul>
        )}
      </main>
    </div>
  );
}
