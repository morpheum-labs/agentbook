import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchInstances, type SwarmRuntimeInstance } from "@/lib/api";
import { RuntimeInstanceCard } from "@/components/runtime-instance/runtime-instance-card";
import { RuntimeInstanceEmptyState } from "@/components/runtime-instance/runtime-instance-empty-state";
import { RuntimeInstanceListSkeleton } from "@/components/runtime-instance/runtime-instance-list-skeleton";
import { RuntimeInstancesHero } from "@/components/runtime-instance/runtime-instances-hero";

export function RuntimeInstancesPage() {
  const navigate = useNavigate();
  const [instances, setInstances] = useState<SwarmRuntimeInstance[] | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  function load() {
    setErr(null);
    setInstances(null);
    void fetchInstances()
      .then((list) => setInstances(list))
      .catch((e: unknown) => {
        setErr(e instanceof Error ? e.message : "Failed to load");
      });
  }

  useEffect(() => {
    void load();
  }, []);

  function heroPairAndChat() {
    if (!instances?.length || !selectedId) return;
    const sel = instances.find((i) => i.ID === selectedId);
    if (!sel) return;
    setErr(null);
    navigate(`/instances/${encodeURIComponent(sel.ID)}/pair`);
  }

  return (
    <div className="container-app max-w-4xl space-y-8 py-8 sm:py-10">
      <RuntimeInstancesHero
        onRefresh={() => void load()}
        refreshDisabled={instances === null && !err}
        onPairAndChat={heroPairAndChat}
        pairChatDisabled={!instances?.length || !selectedId}
      />

      {err && (
        <p
          className="rounded-xl border border-destructive/30 bg-destructive/5 px-4 py-3 text-destructive text-body"
          role="alert"
        >
          {err}
        </p>
      )}

      {instances === null && !err && <RuntimeInstanceListSkeleton />}

      {instances && instances.length === 0 && !err && (
        <RuntimeInstanceEmptyState className="mt-2" />
      )}

      {instances && instances.length > 0 && (
        <p className="text-caption-body text-muted-foreground">
          Click a card to select it for the hero <span className="font-medium text-foreground">Pair & chat</span>{" "}
          action, or use <span className="font-medium text-foreground">Pair gateway & multi-agent chat</span> on a
          card.
        </p>
      )}

      {instances && instances.length > 0 && (
        <ul className="flex flex-col gap-4" aria-label="Runtime instance list">
          {instances.map((inst) => (
            <li key={inst.ID}>
              <RuntimeInstanceCard
                instance={inst}
                selected={selectedId === inst.ID}
                onSelect={() =>
                  setSelectedId((prev) => (prev === inst.ID ? null : inst.ID))
                }
                onPairAndChat={() => navigate(`/instances/${encodeURIComponent(inst.ID)}/pair`)}
              />
            </li>
          ))}
        </ul>
      )}

    </div>
  );
}
