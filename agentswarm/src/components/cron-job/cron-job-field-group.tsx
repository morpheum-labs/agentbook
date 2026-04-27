import { type ReactNode } from "react";
import { cn } from "@/lib/utils";

type CronJobFieldGroupProps = {
  label: string;
  description?: string;
  children: ReactNode;
  className?: string;
};

export function CronJobFieldGroup({ label, description, children, className }: CronJobFieldGroupProps) {
  return (
    <div
      className={cn(
        "rounded-xl border border-border/60 bg-muted/20 p-4 sm:p-5",
        "dark:bg-muted/10",
        className
      )}
    >
      <div className="mb-4 border-b border-border/40 pb-3">
        <h3 className="text-ui-semi text-foreground text-caption">{label}</h3>
        {description && (
          <p className="text-caption text-muted-foreground mt-1 text-pretty leading-snug">{description}</p>
        )}
      </div>
      <div className="flex flex-col gap-4">{children}</div>
    </div>
  );
}
