import type { ReactNode } from "react";
import { AppHeader } from "@/components/app-header";

/**
 * One scrollport for the whole app so the frosted AppHeader stays on top
 * while route content scrolls beneath it.
 */
type AppPageShellProps = {
  children: ReactNode;
};

export function AppPageShell({ children }: AppPageShellProps) {
  return (
    <div className="flex h-dvh w-full min-h-0 flex-col overflow-hidden">
      <AppHeader />
      <main className="flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-y-contain">
        {children}
      </main>
    </div>
  );
}
