import { cn } from "@/lib/utils";

/**
 * Native `<select>` — same surface treatment as `Input` (transparent on card) plus
 * `text-foreground` in both themes (terminal monospace palette).
 * `color-scheme` helps the system dropdown in dark.
 */
export const nativeSelectClass = cn(
  "h-10 w-full min-w-0 rounded-none border border-border bg-transparent px-3",
  "text-body text-foreground",
  "shadow-elevation-0 outline-none",
  "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]",
  "disabled:cursor-not-allowed disabled:opacity-50",
  "[color-scheme:light] dark:[color-scheme:dark]"
);
