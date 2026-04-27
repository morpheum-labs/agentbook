import { NavLink } from "react-router-dom";
import { BarChart3, Home, Timer } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";

type AppHeaderProps = {
  maxWidthClassName?: string;
};

/**
 * Superhuman design system: light = Mysteria “paid chrome” strip (design-md/superhuman/preview.html);
 * dark = dark glass bar (preview-dark.html).
 */
function navLinkClass({ isActive }: { isActive: boolean }) {
  return cn(
    "text-nav inline-flex items-center gap-1.5 rounded-sm px-4 py-2 transition-[color,background-color]",
    "text-[color:var(--translucent-white-80)] hover:text-[color:var(--translucent-white-95)]",
    "dark:text-[color:var(--dark-text-secondary)] dark:hover:text-[color:var(--dark-text-primary)]",
    isActive
      ? "bg-white/10 text-[color:var(--translucent-white-95)] dark:bg-[var(--dark-surface-elevated)] dark:text-[color:var(--dark-text-primary)]"
      : "hover:bg-white/5 dark:hover:bg-white/5"
  );
}

const themeToggleOnStripClass =
  "hover:opacity-100 text-[color:var(--translucent-white-80)] hover:bg-white/10 hover:text-[color:var(--translucent-white-95)] dark:text-[color:var(--dark-text-secondary)] dark:hover:bg-white/5 dark:hover:text-[color:var(--dark-text-primary)] focus-visible:ring-white/30 dark:focus-visible:ring-[var(--lavender-glow)]/40";

export function AppHeader({ maxWidthClassName = "max-w-4xl" }: AppHeaderProps) {
  return (
    <header
      className={cn(
        "border-b backdrop-blur-md",
        "sticky top-0 z-sticky",
        "border-[rgba(255,255,255,0.1)] bg-[rgba(27,25,56,0.95)]",
        "dark:border-border dark:bg-[rgba(18,17,17,0.92)]"
      )}
    >
      <div
        className={cn(
          "container-app flex flex-col gap-3 py-4 sm:py-5",
          maxWidthClassName
        )}
      >
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1
              className={cn(
                "text-section-heading",
                "text-[color:var(--translucent-white-95)]",
                "dark:text-[color:var(--dark-text-primary)]"
              )}
            >
              ClawLaundry
            </h1>
            <p
              className={cn(
                "text-caption-body mt-1",
                "text-[color:var(--translucent-white-80)]",
                "dark:text-[color:var(--dark-text-secondary)]"
              )}
            >
              Swarm Hands and cron job metadata
            </p>
          </div>
          <ThemeToggle className={themeToggleOnStripClass} />
        </div>
        <nav
          className={cn(
            "flex flex-wrap items-center gap-2",
            "border-t border-white/10 pt-3",
            "dark:border-border/80"
          )}
          aria-label="Main"
        >
          <NavLink to="/" className={navLinkClass} end>
            <Home className="size-4 shrink-0 opacity-90" />
            Home
          </NavLink>
          <NavLink to="/chart" className={navLinkClass}>
            <BarChart3 className="size-4 shrink-0 opacity-90" />
            Agent chart
          </NavLink>
          <NavLink to="/cron-jobs" className={navLinkClass}>
            <Timer className="size-4 shrink-0 opacity-90" />
            Cron jobs
          </NavLink>
        </nav>
      </div>
    </header>
  );
}
