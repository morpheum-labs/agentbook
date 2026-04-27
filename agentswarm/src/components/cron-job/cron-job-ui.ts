import { cn } from "@/lib/utils";

/** Native `<select>` tuned to match Input + Dribbble-style focus ring. */
export const cronJobSelectClass = cn(
  "h-10 w-full rounded-sm border border-border bg-background px-3 text-body",
  "shadow-elevation-0 outline-none",
  "focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]"
);
