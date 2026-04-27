import type { ReactNode } from "react";
import { useLocation } from "react-router-dom";
import { AppHeader } from "@/components/app-header";

/**
 * One scrollport for the whole app so the frosted AppHeader stays on top
 * while route content scrolls beneath it.
 */
function headerMaxWidthForPath(pathname: string) {
  if (pathname === "/agents/new" || /^\/agents\/[^/]+$/.test(pathname)) {
    return "max-w-3xl";
  }
  if (pathname === "/cron-jobs/new" || /^\/cron-jobs\/[^/]+$/.test(pathname)) {
    return "max-w-5xl";
  }
  return "max-w-4xl";
}

type AppPageShellProps = {
  children: ReactNode;
};

export function AppPageShell({ children }: AppPageShellProps) {
  const { pathname } = useLocation();
  return (
    <div className="h-dvh w-full min-h-0 overflow-y-auto overscroll-y-contain">
      <AppHeader maxWidthClassName={headerMaxWidthForPath(pathname)} />
      <main className="min-h-0">{children}</main>
    </div>
  );
}
