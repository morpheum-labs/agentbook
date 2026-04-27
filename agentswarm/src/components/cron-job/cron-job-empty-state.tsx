import { Link } from "react-router-dom";
import { Inbox, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CronJobEmptyStateProps = {
  hasAgents: boolean;
  className?: string;
};

export function CronJobEmptyState({ hasAgents, className }: CronJobEmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center rounded-2xl border border-dashed border-border/80",
        "bg-gradient-to-b from-muted/30 to-card px-6 py-16 text-center shadow-elevation-0",
        className
      )}
    >
      <div
        className={cn(
          "mb-5 flex size-16 items-center justify-center rounded-2xl",
          "bg-gradient-to-br from-[var(--mysteria-purple)]/15 to-[var(--lavender-glow)]/20",
          "text-[var(--amethyst-link)]"
        )}
      >
        <Inbox className="size-8" strokeWidth={1.5} aria-hidden />
      </div>
      <h2 className="text-subheading-lg text-foreground font-medium">No jobs yet</h2>
      <p className="text-caption-body text-muted-foreground mt-2 max-w-md text-pretty">
        {hasAgents
          ? "Create your first scheduled task. It only stores metadata — your runner picks it up from the API."
          : "You need at least one Hand before you can attach a cron job."}
      </p>
      <div className="mt-6 flex flex-wrap items-center justify-center gap-2">
        {hasAgents ? (
          <Button asChild className="rounded-xl">
            <Link to="/cron-jobs/new" className="inline-flex items-center gap-2">
              <Plus className="size-4" />
              Create your first job
            </Link>
          </Button>
        ) : (
          <Button asChild variant="default" className="rounded-xl">
            <Link to="/agents/new" className="inline-flex items-center gap-2">
              <Plus className="size-4" />
              New agent
            </Link>
          </Button>
        )}
        <Button asChild variant="ghost" className="rounded-xl text-muted-foreground">
          <Link to="/">Back home</Link>
        </Button>
      </div>
    </div>
  );
}
