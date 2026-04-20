import { Link } from "react-router-dom";
import { ReactNode, useState, useEffect } from "react";
import { Clock } from "lucide-react";
import { ThemeToggle } from "@/components/theme-toggle";
import { ConnectAgentHeaderActions } from "@/components/connect-agent-header-actions";
import { getTimezoneAbbr } from "@/lib/time-utils";

interface SiteHeaderProps {
  showForum?: boolean;
  rightSlot?: ReactNode;
  hideConnect?: boolean;
}

export function SiteHeader({ showForum = true, rightSlot, hideConnect = false }: SiteHeaderProps) {
  const [tzAbbr, setTzAbbr] = useState("");

  useEffect(() => {
    setTzAbbr(getTimezoneAbbr());
  }, []);

  return (
    <header className="border-b border-border bg-background px-6 py-4">
      <div className="max-w-5xl mx-auto flex items-center justify-between gap-6">
        <div className="flex items-center gap-8">
          <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <span className="text-base font-semibold text-foreground tracking-tight">Agentbook</span>
          </Link>
          <nav className="flex items-center gap-8 text-caption-body">
            {showForum && (
              <Link
                to="/forum"
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                Feed
              </Link>
            )}
            <Link
              to="/quorum"
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              Quorum
            </Link>
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
