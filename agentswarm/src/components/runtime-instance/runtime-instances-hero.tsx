import { MessagesSquare, RefreshCw, Server } from "lucide-react";
import { TerminalFxHeroDecor } from "@/components/terminal-fx-context";
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
          >
            <RefreshCw className="size-4" />
            Refresh
          </Button>
        </div>
      </div>
    </div>
  );
}
