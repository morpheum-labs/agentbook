import { NavLink } from "react-router-dom";
import { BarChart3, Home, Timer } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

type AppHeaderProps = {
  maxWidthClassName?: string;
};

function navLinkClass({ isActive }: { isActive: boolean }) {
  return cn(
    buttonVariants({ variant: isActive ? "secondary" : "ghost", size: "sm" }),
    "text-nav"
  );
}

export function AppHeader({ maxWidthClassName = "max-w-4xl" }: AppHeaderProps) {
  return (
    <header className="border-b border-border bg-surface-elevated/30">
      <div
        className={cn(
          "container-app section-y flex flex-col gap-4 py-6",
          maxWidthClassName
        )}
      >
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-section-heading text-foreground">ClawLaundry</h1>
            <p className="text-caption-body mt-1 text-muted-foreground">
              Swarm Hands and cron job metadata
            </p>
          </div>
          <ThemeToggle />
        </div>
        <nav
          className="flex flex-wrap items-center gap-2 border-t border-border/60 pt-4"
          aria-label="Main"
        >
          <NavLink to="/" className={navLinkClass} end>
            <Home className="size-4" />
            Home
          </NavLink>
          <NavLink to="/chart" className={navLinkClass}>
            <BarChart3 className="size-4" />
            Agent chart
          </NavLink>
          <NavLink to="/cron-jobs" className={navLinkClass}>
            <Timer className="size-4" />
            Cron jobs
          </NavLink>
        </nav>
      </div>
    </header>
  );
}
