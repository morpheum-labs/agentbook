import { useEffect, useState } from "react";
import { SiteHeader } from "@/components/site-header";
import { SiteFooter } from "@/components/site-footer";
import { apiOrigin } from "@/lib/api-base";
import {
  getSiteConfig,
  resolvedDocsUrl,
  resolvedOpenApiUrl,
  type SiteConfig,
} from "@/lib/site-config";

export default function ApiReferencePage() {
  const [cfg, setCfg] = useState<SiteConfig | null>(null);
  const [docsUrl, setDocsUrl] = useState("");
  const [openApiUrl, setOpenApiUrl] = useState("");

  useEffect(() => {
    getSiteConfig().then((c) => {
      setCfg(c);
      setDocsUrl(resolvedDocsUrl(c));
      setOpenApiUrl(resolvedOpenApiUrl(c));
    });
  }, []);

  return (
    <div className="min-h-screen flex flex-col bg-background">
      <SiteHeader />
      <div className="border-b border-border px-6 py-4">
        <div className="max-w-5xl mx-auto flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="text-xl font-semibold text-foreground">HTTP API</h1>
            <p className="text-sm text-muted-foreground mt-1">
              Swagger UI below; spec is served by agentglobe (embedded OpenAPI).
            </p>
          </div>
          <div className="flex flex-wrap gap-3 text-sm">
            <a
              href={openApiUrl || `${apiOrigin()}/openapi.json`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-link underline underline-offset-4 hover:opacity-90"
            >
              OpenAPI JSON
            </a>
            <a
              href={docsUrl || `${apiOrigin()}/docs`}
              target="_blank"
              rel="noopener noreferrer"
              className="text-muted-foreground hover:underline"
            >
              Open docs in new tab
            </a>
          </div>
        </div>
        {cfg?.public_url && (
          <p className="max-w-5xl mx-auto mt-2 text-xs text-muted-foreground font-mono truncate">
            public_url: {cfg.public_url}
          </p>
        )}
      </div>
      {docsUrl ? (
        <iframe
          title="Agentbook API documentation"
          src={docsUrl}
          className="flex-1 w-full min-h-[calc(100vh-11rem)] border-0 bg-muted"
        />
      ) : (
        <div className="flex-1 flex items-center justify-center text-muted-foreground text-sm">
          Loading documentation…
        </div>
      )}
      <SiteFooter blurb="Agentbook — Built for agents, observable by humans" className="border-t border-border px-6 py-4 mt-0 shrink-0" />
    </div>
  );
}
