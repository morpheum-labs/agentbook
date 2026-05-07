import { useEffect, useState } from "react";
import { Moon, Sun } from "lucide-react";
import {
  getStoredTheme,
  setStoredTheme,
  applyTheme,
  getEffectiveTheme,
} from "@/lib/theme-utils";
import { cn } from "@/lib/utils";

type ThemeToggleProps = { className?: string };

export function ThemeToggle({ className }: ThemeToggleProps) {
  const [theme, setTheme] = useState<"light" | "dark">("light");
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
    const stored = getStoredTheme();
    setTheme(getEffectiveTheme(stored));
  }, []);

  function toggle() {
    const next = theme === "dark" ? "light" : "dark";
    setTheme(next);
    setStoredTheme(next);
    applyTheme(next);
  }

  if (!mounted) {
    return (
      <button type="button" className={cn(className)} aria-label="Theme">
        <Sun className="mx-auto size-4" />
      </button>
    );
  }

  return (
    <button
      type="button"
      onClick={toggle}
      title={theme === "dark" ? "Light mode" : "Dark mode"}
      className={cn(className)}
    >
      {theme === "dark" ? <Sun className="mx-auto size-4" /> : <Moon className="mx-auto size-4" />}
    </button>
  );
}
