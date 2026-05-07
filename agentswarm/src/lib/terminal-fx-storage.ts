const STORAGE_KEY = "agentswarm_terminal_fx_mode";

/** RichardApps-style atmosphere: off | full (rain + CRT stack) | winter (snow). */
export type TerminalFxMode = "off" | "full" | "winter";

export function getStoredTerminalFxMode(): TerminalFxMode {
  if (typeof window === "undefined") return "full";
  const v = localStorage.getItem(STORAGE_KEY) as TerminalFxMode | null;
  if (v === "off" || v === "full" || v === "winter") return v;
  return "full";
}

export function setStoredTerminalFxMode(mode: TerminalFxMode): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(STORAGE_KEY, mode);
}

/** Cycle: full → winter → off → full (skip winter when reducing choices is OK). */
export function cycleTerminalFxMode(current: TerminalFxMode): TerminalFxMode {
  if (current === "full") return "winter";
  if (current === "winter") return "off";
  return "full";
}

export function applyTerminalFxDocumentClass(mode: TerminalFxMode, isDark: boolean): void {
  if (typeof document === "undefined") return;
  const root = document.documentElement;
  root.classList.remove("terminal-effects-on", "terminal-fx-winter");
  if (!isDark || mode === "off") return;
  root.classList.add("terminal-effects-on");
  if (mode === "winter") root.classList.add("terminal-fx-winter");
}
