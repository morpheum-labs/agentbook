import { Server } from "lucide-react";
import { cn } from "@/lib/utils";

export function RuntimeInstanceEmptyState({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-dashed border-border/80 bg-muted/20 px-6 py-12 text-center",
        className
      )}
      role="status"
    >
      <div className="mx-auto flex max-w-md flex-col items-center gap-3">
        <div className="flex size-12 items-center justify-center rounded-2xl bg-muted/60 text-muted-foreground">
          <Server className="size-6" strokeWidth={1.75} aria-hidden />
        </div>
        <p className="text-card-title text-foreground font-medium">No runtime instances yet</p>
        <p className="text-caption-body text-muted-foreground text-pretty">
          When a Miroclaw runtime registers via the API, it will appear here with status, hostname, and
          heartbeat information.
        </p>
      </div>
    </div>
  );
}
