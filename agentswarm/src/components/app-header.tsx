import { NavLink } from "react-router-dom";
import { ThemeToggle } from "@/components/theme-toggle";
import { TerminalFxToggle } from "@/components/terminal-fx-toggle";
import { cn } from "@/lib/utils";
import { useHtmlTheme } from "@/lib/theme-utils";

/** Live token rows — matches preview-dark.html swatch hex lines for light / ocean dark. */
const LIVE_TOKEN_HEX = {
  light: {
    main: "#2d6a32",
    bg: "rgb(45, 106, 50)",
    sel: "#a8c4a4",
  },
  dark: {
    main: "#72b6ff",
    bg: "rgb(15, 129, 236)",
    sel: "#3b6d8b",
  },
} as const;

type NavPreset = { to: string; label: string; end?: boolean };

const NAV_PRESETS: NavPreset[] = [
  { to: "/", label: "Home", end: true },
  { to: "/multi-chat", label: "Multi chat" },
  { to: "/chart", label: "Agent chart" },
  { to: "/cron-jobs", label: "Cron jobs" },
  { to: "/instances", label: "Runtime instances" },
];

/**
 * Structure and class names from design-md/terminal/preview-dark.html:
 * `.toolbar` buttons, `h2#themes` + `.section-desc` + `.theme-grid`,
 * `h2#tokens` + `.swatches` / `.swatch` / `.swatch-color` / `.swatch-label` / `.swatch-hex`.
 */
export function AppHeader() {
  const htmlTheme = useHtmlTheme();
  const hex = LIVE_TOKEN_HEX[htmlTheme];

  return (
    <header
      className={cn(
        "app-header-preview-dark border-b border-dotted backdrop-blur-[var(--terminal-tab-blur)]",
        "sticky top-0 z-[var(--z-app-header)]",
        "border-primary bg-[color-mix(in_srgb,var(--pure-white)_78%,transparent)]",
        "dark:border-primary dark:bg-[rgba(var(--terminal-bg-rgb),0.32)]"
      )}
    >
      <div className="container-app flex flex-col gap-0 py-4 sm:py-5">
        <div className="flex w-full flex-wrap items-start justify-between gap-4">
          <div className="flex min-w-0 flex-col gap-2">
            <h2 id="themes">Clawgotcha</h2>
            <p className="section-desc mb-0">Swarm Hands and cron job metadata</p>
          </div>
          <div
            className="toolbar ml-auto flex shrink-0 flex-col items-end gap-2 sm:flex-row sm:items-center"
            role="group"
            aria-label="Toolbar"
          >
            <TerminalFxToggle />
            <ThemeToggle />
          </div>
        </div>


        <nav className="theme-grid" aria-label="Main">
          {NAV_PRESETS.map(({ to, label, end }) => (
            <NavLink key={to} to={to} end={end}>
              {label}
            </NavLink>
          ))}
        </nav>
      </div>
    </header>
  );
}
