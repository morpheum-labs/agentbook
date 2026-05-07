import { Link } from "react-router-dom";
import { Plus, RefreshCw, Sparkles } from "lucide-react";
import { TerminalFxHeroDecor } from "@/components/terminal-fx-context";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CronJobsHeroProps = {
  onRefresh: () => void;
  refreshDisabled?: boolean;
  className?: string;
};

/**
 * Terminal hero panel: diagonal wash + dotted chrome, RichardApps-style actions.
 */
export function CronJobsHero({ onRefresh, refreshDisabled, className }: CronJobsHeroProps) {
  return (
    <div
      className={cn(
        "terminal-fx-hero relative isolate overflow-hidden rounded-none border border-surface-hero-border bg-surface-hero-gradient",
        "px-6 py-8 sm:px-8 sm:py-10 text-surface-hero-foreground",
        "shadow-elevation-2",
        className
      )}
    >
      <TerminalFxHeroDecor />
      <div className="relative z-10 flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div className="max-w-2xl space-y-2">
          <p className="text-caption text-surface-hero-muted inline-flex items-center gap-2">
            <Sparkles className="size-3.5 shrink-0 opacity-90" aria-hidden />
            Scheduler
          </p>
          <h1 className="text-feature font-medium tracking-[-0.02em] text-balance sm:text-4xl">Cron jobs</h1>
          <p className="text-caption-body text-surface-hero-muted max-w-xl text-pretty leading-relaxed">
            Plan when each Hand runs: store a schedule, timeout, and the prompt your runner will inject.
            Execution stays in your stack — this is the control plane.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2 lg:shrink-0">
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={onRefresh}
            disabled={refreshDisabled}
          >
            <RefreshCw className="size-4" />
            Refresh
          </Button>
          <Button size="sm" asChild>
            <Link to="/cron-jobs/new" className="inline-flex items-center gap-1.5">
              <Plus className="size-4" />
              New job
            </Link>
          </Button>
        </div>
      </div>
    </div>
  );
}
