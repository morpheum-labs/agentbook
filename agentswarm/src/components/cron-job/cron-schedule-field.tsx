import { useId, useState } from "react";
import { CalendarClock } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { CronScheduleModal } from "@/components/cron-job/cron-schedule-modal";
import { cn } from "@/lib/utils";

type CronScheduleFieldProps = {
  id: string;
  name?: string;
  value: string;
  onChange: (next: string) => void;
  disabled?: boolean;
  className?: string;
};

/**
 * Read-only schedule line that opens a modal cron builder (presets, 5 fields, or custom text).
 */
export function CronScheduleField({ id, name, value, onChange, disabled, className }: CronScheduleFieldProps) {
  const [open, setOpen] = useState(false);
  const helpId = useId();

  function openModal() {
    if (!disabled) setOpen(true);
  }

  return (
    <div className={cn("relative w-full", className)}>
      <div className="group relative w-full">
        <Input
          id={id}
          name={name}
          type="text"
          readOnly
          disabled={disabled}
          value={value}
          onClick={openModal}
          onKeyDown={(e) => {
            if (disabled) return;
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              openModal();
            }
          }}
          className={cn(
            "h-10 cursor-pointer rounded-md pr-11 font-mono text-caption",
            "read-only:cursor-pointer",
            "focus-visible:ring-2"
          )}
          placeholder="Click to set schedule"
          autoComplete="off"
          role="combobox"
          aria-expanded={open}
          aria-haspopup="dialog"
          aria-describedby={helpId}
        />
        <p id={helpId} className="sr-only">
          Press Enter or click to open the schedule builder. Your runner may expect a 5-field cron
          or an opaque label.
        </p>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          className="text-muted-foreground absolute top-1/2 right-0.5 h-8 w-8 -translate-y-1/2"
          onClick={openModal}
          disabled={disabled}
          tabIndex={-1}
          aria-label="Open schedule builder"
        >
          <CalendarClock className="size-4" />
        </Button>
      </div>
      <CronScheduleModal open={open} onOpenChange={setOpen} value={value} onApply={onChange} />
    </div>
  );
}
