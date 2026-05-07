import { cn } from "@/lib/utils";
import { useTerminalFx } from "@/components/terminal-fx-context";

function fxLabel(mode: string): string {
  if (mode === "full") return "Effects: on";
  if (mode === "winter") return "Effects: winter";
  return "Effects: off";
}

type TerminalFxToggleProps = { className?: string };

/** preview-dark.html `#btn-glow` — cycles CRT + rain → snow → off. */
export function TerminalFxToggle({ className }: TerminalFxToggleProps) {
  const { mode, cycleMode } = useTerminalFx();
  const active = mode !== "off";

  return (
    <button
      type="button"
      id="btn-glow"
      onClick={cycleMode}
      title="Cycle terminal effects: on → winter → off"
      className={cn(active && "active", className)}
    >
      {fxLabel(mode)}
    </button>
  );
}
