import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Pencil, PlusCircle, RefreshCw } from "lucide-react";
import { fetchAgents, type SwarmAgent } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
    <div className="container-app max-w-4xl py-10">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-body-heading">Agents</h2>
          <div className="flex flex-wrap items-center gap-2">
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={load}
              disabled={agents === null && !err}
              className={cn(
                "h-9 rounded-lg border border-border/60 bg-accent/50 shadow-sm",
                "text-foreground hover:bg-accent/80"
              )}
            >
              <RefreshCw className="size-4" />
              Refresh
            </Button>
            <Button
              size="sm"
              asChild
              className="h-9 rounded-lg border-0 bg-primary text-primary-foreground hover:opacity-95 shadow-sm"
            >
              <Link to="/agents/new" className="inline-flex items-center gap-1.5">
                <PlusCircle className="size-4" />
                New agent
              </Link>
            </Button>
          </div>
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
                      <Button
                        size="sm"
                        variant="secondary"
                        className="h-9 rounded-xl border border-border/60 bg-accent/50 shadow-sm hover:bg-accent/80"
                        asChild
                      >
                        <Link
                          to={`/agents/${a.ID}`}
                          className="inline-flex items-center gap-1.5"
                        >
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
    </div>
  );
}
