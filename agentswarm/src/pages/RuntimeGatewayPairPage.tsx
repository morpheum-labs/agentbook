import { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { fetchInstances, type SwarmRuntimeInstance } from "@/lib/api";
import { RuntimeGatewayPairView } from "@/components/runtime-instance/runtime-gateway-pair-view";
import { Button } from "@/components/ui/button";

export function RuntimeGatewayPairPage() {
  const { instanceId = "" } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState<string | null>(null);
  const [instance, setInstance] = useState<SwarmRuntimeInstance | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setErr(null);
    setInstance(null);
    void fetchInstances()
      .then((list) => {
        if (cancelled) return;
        const found = list.find((i) => i.ID === instanceId);
        if (!found) {
          setErr("Runtime instance not found.");
          setInstance(null);
        } else {
          setInstance(found);
        }
      })
      .catch((e: unknown) => {
        if (!cancelled) setErr(e instanceof Error ? e.message : "Failed to load");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [instanceId]);

  function onBack() {
    navigate("/instances");
  }

  return (
    <div className="container-app space-y-6 py-8 sm:py-10">
      {loading && (
        <div className="rounded-2xl border border-border/80 bg-card p-8 text-caption-body text-muted-foreground">
          Loading runtime…
        </div>
      )}

      {!loading && err && (
        <div className="rounded-2xl border border-destructive/30 bg-destructive/5 p-6 space-y-4">
          <p className="text-destructive text-body" role="alert">
            {err}
          </p>
          <Button type="button" variant="secondary" size="sm" className="rounded-lg" asChild>
            <Link to="/instances">Back to runtimes</Link>
          </Button>
        </div>
      )}

      {!loading && !err && instance && (
        <RuntimeGatewayPairView instance={instance} onBack={onBack} />
      )}
    </div>
  );
}
