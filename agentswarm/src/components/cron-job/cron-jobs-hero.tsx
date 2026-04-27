import { Link } from "react-router-dom";
import { Plus, RefreshCw, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type CronJobsHeroProps = {
  onRefresh: () => void;
  refreshDisabled?: boolean;
  className?: string;
};

/**
 * Reference-style “hero” band: deep purple → cream gradient, high-contrast actions.
 */
export function CronJobsHero({ onRefresh, refreshDisabled, className }: CronJobsHeroProps) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-2xl border border-surface-hero-border bg-surface-hero-gradient",
        "px-6 py-8 sm:px-8 sm:py-10 text-surface-hero-foreground",
        "shadow-elevation-2",
        className
      )}
    >
      <div
        className="pointer-events-none absolute -right-24 -top-24 size-64 rounded-full bg-[var(--lavender-glow)]/20 blur-3xl dark:bg-[var(--lavender-glow)]/10"
        aria-hidden
      />
      <div
        className="pointer-events-none absolute -bottom-16 -left-16 size-48 rounded-full bg-white/10 blur-2xl"
        aria-hidden
      />
      <div className="relative flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
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
            className={cn(
              "h-9 rounded-lg border-0",
              "bg-pure-white/10 text-surface-hero-foreground",
              "hover:bg-pure-white/18",
              "disabled:opacity-50"
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
