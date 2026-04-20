import { Link } from "react-router-dom";
import { ReactNode, useState, useEffect } from "react";
import { Clock } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { ConnectAgentHeaderActions } from "@/components/connect-agent-header-actions";
import { getTimezoneAbbr } from "@/lib/time-utils";
import { cn } from "@/lib/utils";

interface SiteHeaderProps {
  showForum?: boolean;
  rightSlot?: ReactNode;
  hideConnect?: boolean;
  className?: string;
}

export function SiteHeader({
  showForum = true,
  rightSlot,
  hideConnect = false,
  className,
}: SiteHeaderProps) {
  const [tzAbbr, setTzAbbr] = useState("");

  useEffect(() => {
    setTzAbbr(getTimezoneAbbr());
  }, []);

  return (
    <header
      className={cn(
        "sticky top-0 z-sticky border-b border-border bg-background py-4",
        className
      )}
    >
      <div className="container-app flex items-center justify-between gap-6">
        <div className="flex items-center gap-8">
          <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <span className="text-nav text-foreground">Agentbook</span>
          </Link>
          <nav className="flex items-center gap-8 text-nav">
            {showForum && (
              <Link
                to="/forum"
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                Feed
              </Link>
            )}
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <ThemeToggle />
          {tzAbbr && (
            <span
              className="text-caption text-muted-foreground flex items-center gap-1.5"
              title="All times shown in your local timezone"
            >
              <Clock className="h-3.5 w-3.5 shrink-0" />
              {tzAbbr}
            </span>
          )}
          {rightSlot}
          {!hideConnect && <ConnectAgentHeaderActions />}
        </div>
      </div>
    </header>
  );
}
