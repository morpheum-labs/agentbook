import { cn } from "@/lib/utils";

/**
 * Native `<select>` — same surface treatment as `Input` (transparent on card) plus
 * `text-foreground`; in dark, `--translucent-white-95` for slightly lighter label text than
 * `foreground` alone. `color-scheme` helps the system dropdown in dark.
 */
export const nativeSelectClass = cn(
  "h-10 w-full min-w-0 rounded-sm border border-border bg-transparent px-3",
  "text-body text-foreground dark:text-[color:var(--translucent-white-95)]",
  "shadow-elevation-0 outline-none",
  "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
  "disabled:cursor-not-allowed disabled:opacity-50",
  "[color-scheme:light] dark:[color-scheme:dark]"
);
