import { useEffect, useId, useState } from "react";
import {
  Dialog,
  DialogBody,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import {
  type FivePartCron,
  CRON_FIELD_LABELS,
  CRON_PRESETS,
  buildExpression,
  initialBuilderParts,
  shouldUseCustomExpression,
} from "@/components/cron-job/cron-schedule-utils";

type CronScheduleModalProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  /** Current field value (opaque to API) */
  value: string;
  onApply: (next: string) => void;
};

export function CronScheduleModal({ open, onOpenChange, value, onApply }: CronScheduleModalProps) {
  const baseId = useId();
  const [mode, setMode] = useState<"builder" | "custom">("builder");
  const [parts, setParts] = useState<FivePartCron>(initialBuilderParts(""));
  const [customText, setCustomText] = useState("");

  useEffect(() => {
    if (!open) return;
    const c = shouldUseCustomExpression(value);
    setMode(c ? "custom" : "builder");
    setParts(initialBuilderParts(value));
    setCustomText(value);
  }, [open, value]);

  function setPart(i: number, token: string) {
    setParts((prev) => {
      const next = [...prev] as FivePartCron;
      next[i] = token;
      return next;
    });
  }

  function applyPreset(preset: string) {
    setMode("builder");
    setParts(initialBuilderParts(preset));
  }

  function handleApply() {
    if (mode === "custom") {
      onApply(customText.trim());
    } else {
      onApply(buildExpression(parts));
    }
    onOpenChange(false);
  }

  const preview = buildExpression(parts);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-xl gap-0 border-border/80 p-0">
        <div className="h-1 w-full bg-gradient-to-r from-[var(--mysteria-purple)] via-[var(--amethyst-link)] to-[var(--lavender-glow)]" />
        <DialogHeader className="border-0">
          <DialogTitle>Schedule</DialogTitle>
          <DialogDescription>
            Build a standard 5-field cron, or use a custom label if your runner resolves schedules
            elsewhere.
          </DialogDescription>
        </DialogHeader>
        <DialogBody className="space-y-5">
          <div
            className="flex gap-1 rounded-lg border border-border/70 bg-muted/30 p-1"
            role="group"
            aria-label="Expression mode"
          >
            <button
              type="button"
              className={cn(
                "flex-1 rounded-md px-3 py-2 text-caption transition-colors",
                mode === "builder"
                  ? "bg-card text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
              onClick={() => setMode("builder")}
            >
              Unix cron
            </button>
            <button
              type="button"
              className={cn(
                "flex-1 rounded-md px-3 py-2 text-caption transition-colors",
                mode === "custom"
                  ? "bg-card text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              )}
              onClick={() => setMode("custom")}
            >
              Custom label
            </button>
          </div>

          {mode === "builder" && (
            <>
              <div>
                <p className="text-micro text-muted-foreground mb-2">Quick presets</p>
                <div className="flex flex-wrap gap-2">
                  {CRON_PRESETS.map((p) => (
                    <button
                      key={p.value}
                      type="button"
                      onClick={() => applyPreset(p.value)}
                      className={cn(
                        "rounded-lg border border-border/60 bg-background/80 px-2.5 py-1.5",
                        "text-caption text-foreground",
                        "hover:border-[var(--amethyst-link)]/50 hover:bg-accent/40",
                        "text-left"
                      )}
                    >
                      {p.label}
                    </button>
                  ))}
                </div>
              </div>

              <div>
                <p className="text-micro text-muted-foreground mb-2">Fields (minute · hour · day · month · weekday)</p>
                <div className="grid gap-3 sm:grid-cols-5">
                  {CRON_FIELD_LABELS.map((f, i) => (
                    <div key={f.key} className="space-y-1.5 sm:min-w-0">
                      <label
                        className="text-caption text-muted-foreground block"
                        htmlFor={`${baseId}-cron-${f.key}`}
                      >
                        {f.name}
                      </label>
                      <Input
                        id={`${baseId}-cron-${f.key}`}
                        value={parts[i]}
                        onChange={(e) => setPart(i, e.target.value)}
                        className="font-mono text-caption h-9 rounded-md"
                        autoComplete="off"
                        placeholder="*"
                        title={f.hint}
                        aria-label={`${f.name} (${f.hint})`}
                      />
                    </div>
                  ))}
                </div>
                <p className="text-micro text-muted-foreground mt-2" id={`${baseId}-preview`}>
                  Preview:{" "}
                  <code className="text-caption font-mono text-foreground" aria-live="polite">
                    {preview}
                  </code>
                </p>
              </div>
            </>
          )}

          {mode === "custom" && (
            <div>
              <label
                className="text-caption text-muted-foreground block mb-1.5"
                htmlFor={`${baseId}-custom`}
              >
                Custom string
              </label>
              <Textarea
                id={`${baseId}-custom`}
                className="min-h-24 font-mono text-caption"
                value={customText}
                onChange={(e) => setCustomText(e.target.value)}
                placeholder="e.g. nightly, prod-maintenance, 0 9 * * 1-5 (paste cron too)"
              />
              <p className="text-caption text-muted-foreground mt-2" role="note">
                If this is a full cron line, you can still switch to <strong>Unix cron</strong> to edit
                fields, as long as it is exactly five space-separated parts.
              </p>
            </div>
          )}
        </DialogBody>
        <DialogFooter className="sm:justify-between">
          <Button
            type="button"
            variant="ghost"
            className="order-last sm:order-first"
            onClick={() => onOpenChange(false)}
          >
            Cancel
          </Button>
          <div className="flex gap-2">
            <Button type="button" variant="outline" className="rounded-lg" onClick={handleApply}>
              Apply
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
