import { MessagesSquare, RefreshCw, Server } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type RuntimeInstancesHeroProps = {
  onRefresh: () => void;
  refreshDisabled?: boolean;
  onPairAndChat?: () => void;
  pairChatDisabled?: boolean;
  className?: string;
};

export function RuntimeInstancesHero({
  onRefresh,
  refreshDisabled,
  onPairAndChat,
  pairChatDisabled,
  className,
}: RuntimeInstancesHeroProps) {
  return (
    <div
      className={cn(
        "relative overflow-hidden rounded-none border border-surface-hero-border bg-surface-hero-gradient",
        "px-6 py-8 sm:px-8 sm:py-10 text-surface-hero-foreground",
        "shadow-elevation-2",
        className
      )}
    >
      <div className="relative flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
        <div className="max-w-2xl space-y-2">
          <p className="text-caption text-surface-hero-muted inline-flex items-center gap-2">
            <Server className="size-3.5 shrink-0 opacity-90" aria-hidden />
            Control plane
          </p>
          <h1 className="text-feature font-medium tracking-[-0.02em] text-balance sm:text-4xl">
            Runtime instances
          </h1>
          <p className="text-caption-body text-surface-hero-muted max-w-xl text-pretty leading-relaxed">
            Miroclaw runtimes that register with Clawgotcha: heartbeat, callbacks, and webhook delivery.
            Registration happens from your stack — this view is read-only.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2 lg:shrink-0">
          {onPairAndChat && (
            <Button
              type="button"
              variant="secondary"
              size="sm"
              onClick={onPairAndChat}
              disabled={pairChatDisabled}
              className={cn(
                "h-9 rounded-none border border-primary/50 bg-card/40",
                "text-surface-hero-foreground hover:bg-card/70",
                "disabled:opacity-50"
              )}
            >
              <MessagesSquare className="size-4" />
              Pair & chat
            </Button>
          )}
          <Button
            type="button"
            variant="secondary"
            size="sm"
            onClick={onRefresh}
            disabled={refreshDisabled}
            className={cn(
              "h-9 rounded-none border border-primary/50 bg-card/40",
              "text-surface-hero-foreground hover:bg-card/70",
              "disabled:opacity-50"
            )}
          >
            <RefreshCw className="size-4" />
            Refresh
          </Button>
        </div>
      </div>
    </div>
  );
}
