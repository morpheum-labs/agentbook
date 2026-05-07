import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, MessagesSquare } from "lucide-react";
import type { SwarmRuntimeInstance } from "@/lib/api";
import {
  exchangeGatewayPairingCode,
  fetchGatewayPairingCode,
  gatewayHttpBaseFromPublicUrl,
  joinGatewayHttpPath,
  postGatewayPairingCodeNew,
} from "@/lib/gateway-pairing";
import { persistMultiChatGatewayRuntime } from "@/lib/multi-chat-gateway-session";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type RuntimeGatewayPairViewProps = {
  instance: SwarmRuntimeInstance;
  onBack: () => void;
};

export function RuntimeGatewayPairView({ instance, onBack }: RuntimeGatewayPairViewProps) {
  const navigate = useNavigate();
  const httpBase = useMemo(() => {
    const pu = instance.PublicURL?.trim();
    if (!pu) return "";
    try {
      return gatewayHttpBaseFromPublicUrl(pu);
    } catch {
      return "";
    }
  }, [instance]);

  const [pairingCode, setPairingCode] = useState("");
  const [tokenPaste, setTokenPaste] = useState("");
  const [busy, setBusy] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    setPairingCode("");
    setTokenPaste("");
    setBusy(null);
    setErr(null);
  }, [instance.ID]);

  /** Prefill from gateway: GET /admin/paircode (browser may fail without CORS). */
  useEffect(() => {
    if (!httpBase) return;
    let cancelled = false;
    setBusy("Reading pairing code…");
    void fetchGatewayPairingCode(httpBase)
      .then((code) => {
        if (cancelled) return;
        if (code) setPairingCode(code);
      })
      .catch(() => {
        if (cancelled) return;
      })
      .finally(() => {
        if (!cancelled) setBusy(null);
      });
    return () => {
      cancelled = true;
    };
  }, [httpBase, instance.ID]);

  async function onGenerateNew() {
    if (!httpBase) return;
    setErr(null);
    setBusy("Generating…");
    try {
      const code = await postGatewayPairingCodeNew(httpBase);
      setPairingCode(code);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Request failed");
    } finally {
      setBusy(null);
    }
  }

  async function onFetchCurrent() {
    if (!httpBase) return;
    setErr(null);
    setBusy("Fetching…");
    try {
      const code = await fetchGatewayPairingCode(httpBase);
      if (code) setPairingCode(code);
      else setErr("No active pairing code (generate a new one).");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Request failed");
    } finally {
      setBusy(null);
    }
  }

  /** GET /admin/paircode (if needed), then POST /pair — completes gateway pairing. */
  async function onPairAndOpenChat() {
    if (!httpBase) return;
    setErr(null);
    setBusy("Pairing…");
    try {
      let code = pairingCode.trim();
      if (!code) {
        code = (await fetchGatewayPairingCode(httpBase)) ?? "";
        if (code) setPairingCode(code);
      }
      if (!code) {
        setErr(
          "No pairing code from GET /admin/paircode. Generate a new code on the gateway (POST /admin/paircode/new) or paste a token below."
        );
        return;
      }
      const token = await exchangeGatewayPairingCode(httpBase, code);
      persistMultiChatGatewayRuntime({
        instanceName: instance.InstanceName,
        gatewayPairingToken: token,
      });
      navigate("/multi-chat");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Pair exchange failed");
    } finally {
      setBusy(null);
    }
  }

  function onSavePastedToken() {
    const t = tokenPaste.trim();
    if (!t) {
      setErr("Paste a bearer token from POST /pair.");
      return;
    }
    setErr(null);
    persistMultiChatGatewayRuntime({
      instanceName: instance.InstanceName,
      gatewayPairingToken: t,
    });
    navigate("/multi-chat");
  }

  const curlGet = httpBase ? `curl -s "${joinGatewayHttpPath(httpBase, "/admin/paircode")}"` : "";
  const curlNew = httpBase
    ? `curl -s -X POST "${joinGatewayHttpPath(httpBase, "/admin/paircode/new")}"`
    : "";
  const curlPair = httpBase
    ? `curl -s -X POST "${joinGatewayHttpPath(httpBase, "/pair")}" -H "X-Pairing-Code: YOUR_CODE"`
    : "";

  const missingPublic = !instance.PublicURL?.trim();

  return (
    <div className="rounded-2xl border border-border/80 bg-card shadow-elevation-2 overflow-hidden">
      <div className="border-b border-border/60 bg-muted/20 px-5 py-4 sm:px-6">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          className="-ml-2 mb-3 rounded-lg gap-1.5 text-muted-foreground hover:text-foreground"
          onClick={onBack}
        >
          <ArrowLeft className="size-4" aria-hidden />
          Back to runtimes
        </Button>
        <div className="space-y-2">
          <h1 className="text-card-title flex items-center gap-2 font-medium tracking-tight">
            <MessagesSquare className="size-5 opacity-90" aria-hidden />
            Gateway pairing
          </h1>
          <p className="text-caption-body text-muted-foreground leading-relaxed">
            MiroClaw requires a bearer token for <span className="font-mono text-xs">/ws/chat</span> when pairing is
            enabled. Use the gateway&apos;s admin endpoints on <span className="font-mono text-xs">public_url</span>,
            then continue to multi-agent chat.
          </p>
        </div>
      </div>

      <div className="space-y-4 px-5 py-6 sm:px-6">
        <div className="rounded-xl border border-border/60 bg-muted/30 px-3 py-2 text-caption-body">
          <p className="text-micro text-muted-foreground">Runtime</p>
          <p className="font-medium">{instance.InstanceName}</p>
          <p className="text-micro font-mono text-muted-foreground mt-1 break-all">{httpBase || "—"}</p>
        </div>

        {missingPublic && (
          <p className="text-destructive text-caption-body" role="alert">
            This instance has no <span className="font-mono text-xs">public_url</span>. Pair against a reachable
            gateway HTTP origin first.
          </p>
        )}

        <p className="text-micro text-muted-foreground leading-relaxed">
          Pairing uses <span className="font-mono text-xs">GET /admin/paircode</span> then{" "}
          <span className="font-mono text-xs">POST /pair</span>. Browser calls to{" "}
          <span className="font-mono text-xs">/admin/*</span> often fail without gateway CORS or when admin is
          localhost-only — use curl on the gateway host and paste the token below if needed.
        </p>

        {err && (
          <p className="text-destructive text-caption-body rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2">
            {err}
          </p>
        )}

        <div className="flex flex-wrap gap-2">
          <Button
            type="button"
            size="sm"
            variant="secondary"
            className="rounded-lg"
            disabled={!httpBase || !!busy}
            onClick={() => void onFetchCurrent()}
          >
            Refresh code (GET /admin/paircode)
          </Button>
          <Button
            type="button"
            size="sm"
            className="rounded-lg"
            disabled={!httpBase || !!busy}
            onClick={() => void onGenerateNew()}
          >
            Generate new code
          </Button>
        </div>
        <p className="text-micro text-muted-foreground">
          <span className="font-mono text-[0.65rem] break-all">{curlGet}</span>
        </p>
        <p className="text-micro text-muted-foreground">
          <span className="font-mono text-[0.65rem] break-all">{curlNew}</span>
        </p>

        <div>
          <label htmlFor="pair-code" className="text-micro text-muted-foreground">
            Pairing code
          </label>
          <input
            id="pair-code"
            value={pairingCode}
            onChange={(e) => setPairingCode(e.target.value)}
            placeholder="One-time code from gateway"
            autoComplete="off"
            className={cn(
              "border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring mt-1 flex h-10 w-full rounded-md border px-3 py-2 font-mono text-body shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
            )}
          />
        </div>

        <Button
          type="button"
          size="sm"
          className="w-full rounded-lg sm:w-auto"
          disabled={!httpBase || !!busy}
          onClick={() => void onPairAndOpenChat()}
        >
          {busy === "Pairing…" ? "Pairing…" : "Pair & open chat"}
        </Button>
        <p className="text-micro text-muted-foreground break-all font-mono">{curlPair}</p>

        <div className="border-t border-border/60 pt-4">
          <label htmlFor="token-paste" className="text-micro text-muted-foreground">
            Or paste bearer token manually
          </label>
          <textarea
            id="token-paste"
            value={tokenPaste}
            onChange={(e) => setTokenPaste(e.target.value)}
            placeholder="Token from POST /pair JSON"
            rows={2}
            className="border-input bg-background placeholder:text-muted-foreground focus-visible:ring-ring mt-1 flex w-full resize-y rounded-md border px-3 py-2 font-mono text-caption-body shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-offset-2"
          />
          <Button type="button" variant="outline" size="sm" className="mt-2 rounded-lg" onClick={onSavePastedToken}>
            Save token & open chat
          </Button>
        </div>
      </div>
    </div>
  );
}
