import { useEffect, useState } from "react";

const STORAGE_KEY = "agentswarm_theme";

export type Theme = "light" | "dark" | "system";

export function getStoredTheme(): Theme {
  if (typeof window === "undefined") return "light";
  const v = localStorage.getItem(STORAGE_KEY) as Theme | null;
  return v || "light";
}

export function setStoredTheme(theme: Theme): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(STORAGE_KEY, theme);
}

export function getEffectiveTheme(theme: Theme): "light" | "dark" {
  if (theme === "system") {
    if (typeof window === "undefined") return "light";
    return window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }
  return theme;
}

export function applyTheme(theme: Theme): void {
  if (typeof document === "undefined") return;
  const effective = getEffectiveTheme(theme);
  document.documentElement.classList.remove("light", "dark");
  document.documentElement.classList.add(effective);
}

/** Effective light/dark from `<html class>` — updates when theme toggles. */
export function useHtmlTheme(): "light" | "dark" {
  const [mode, setMode] = useState<"light" | "dark">(() =>
    typeof document !== "undefined" && document.documentElement.classList.contains("dark")
      ? "dark"
      : "light"
  );

  useEffect(() => {
    const el = document.documentElement;
    const sync = () => setMode(el.classList.contains("dark") ? "dark" : "light");
    sync();
    const obs = new MutationObserver(sync);
    obs.observe(el, { attributes: true, attributeFilter: ["class"] });
    return () => obs.disconnect();
  }, []);

  return mode;
}
