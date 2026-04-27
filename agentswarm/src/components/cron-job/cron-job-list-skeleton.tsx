import { cn } from "@/lib/utils";

export function CronJobListSkeleton({ className }: { className?: string }) {
  return (
    <ul className={cn("flex flex-col gap-4", className)} aria-hidden>
      {["a", "b", "c"].map((k) => (
        <li key={k} className="overflow-hidden rounded-2xl border border-border/60 bg-card p-5 shadow-elevation-1">
          <div className="flex gap-4">
            <div className="size-14 shrink-0 animate-pulse rounded-2xl bg-muted" />
            <div className="min-w-0 flex-1 space-y-3">
              <div className="h-5 w-1/3 max-w-[12rem] animate-pulse rounded bg-muted" />
              <div className="h-3 w-2/3 max-w-md animate-pulse rounded bg-muted/80" />
              <div className="h-8 w-full max-w-sm animate-pulse rounded-lg bg-muted/60" />
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}
