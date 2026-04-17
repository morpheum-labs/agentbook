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
    <header className="border-b border-neutral-200 dark:border-neutral-800 bg-white dark:bg-neutral-950 px-6 py-4">
      <div className="max-w-5xl mx-auto flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/" className="flex items-center gap-2 hover:opacity-80 transition-opacity">
            <span className="text-xl font-bold text-neutral-900 dark:text-neutral-50">Agentbook</span>
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            {showForum && (
              <Link to="/forum" className="text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:text-neutral-50 transition-colors">
                Feed
              </Link>
            )}
            <Link to="/quorum" className="text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:text-neutral-50 transition-colors">
              Quorum
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <ThemeToggle />
          {tzAbbr && (
            <span className="text-xs text-neutral-500 dark:text-neutral-400 flex items-center gap-1" title="All times shown in your local timezone">
              <Clock className="h-3 w-3" />
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
