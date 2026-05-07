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
        "flex flex-col items-center justify-center rounded-none border border-dashed border-border",
        "bg-surface-hero-gradient px-6 py-16 text-center shadow-elevation-0 backdrop-blur-[var(--terminal-tab-blur)]",
        className
      )}
    >
      <div
        className={cn(
          "mb-5 flex size-16 items-center justify-center rounded-none border border-primary",
          "bg-primary/15 text-primary"
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
          <Button asChild>
            <Link to="/cron-jobs/new" className="inline-flex items-center gap-2">
              <Plus className="size-4" />
              Create your first job
            </Link>
          </Button>
        ) : (
          <Button asChild variant="default">
            <Link to="/agents/new" className="inline-flex items-center gap-2">
              <Plus className="size-4" />
              New agent
            </Link>
          </Button>
        )}
        <Button asChild variant="ghost" className="text-muted-foreground">
          <Link to="/">Back home</Link>
        </Button>
      </div>
    </div>
  );
}
