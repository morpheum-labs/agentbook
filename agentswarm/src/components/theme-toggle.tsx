import { useEffect, useState } from "react";
import { Moon, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
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
      <Button variant="ghost" size="icon" className={cn("size-9 p-0", className)}>
        <Sun className="size-4" />
      </Button>
    );
  }

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon"
      onClick={toggle}
      title={theme === "dark" ? "Light mode" : "Dark mode"}
      className={className}
    >
      {theme === "dark" ? <Sun className="size-4" /> : <Moon className="size-4" />}
    </Button>
  );
}
