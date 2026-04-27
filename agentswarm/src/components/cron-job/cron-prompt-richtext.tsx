import { useMemo, type CSSProperties } from "react";
import { RichTextarea, createRegexRenderer, type RichTextareaProps } from "rich-textarea";
import { cn } from "@/lib/utils";

/** MiroClaw / prompt convention: `tool_name` and similar backtick-wrapped tool refs. */
const TOOL_BACKTICK = /`[^`]+`/g;

/**
 * {@link https://github.com/inokawa/rich-textarea rich-textarea} for the cron prompt, with
 * green highlight spans for <code>\`...`</code> (tool names and inline refs).
 */
function makeToolBacktickRenderer() {
  return createRegexRenderer([
    [
      TOOL_BACKTICK,
      ({ children, key }) => (
        <span
          key={key}
          className={cn(
            "rounded-sm px-0.5 [font:inherit] [letter-spacing:inherit] [text-shadow:none]",
            "text-green-800 ring-1 ring-inset ring-green-600/25",
            "bg-emerald-500/15",
            "dark:text-emerald-200 dark:ring-emerald-400/25 dark:bg-emerald-500/12"
          )}
        >
          {children}
        </span>
      ),
    ],
  ]);
}

export function CronJobPromptRichtext({
  className,
  style,
  id,
  name,
  value,
  onChange,
  disabled,
  placeholder,
  "aria-label": ariaLabel,
  "aria-describedby": ariaDescribedBy,
  ...rest
}: Omit<RichTextareaProps, "children" | "defaultValue" | "ref">) {
  const renderer = useMemo(() => makeToolBacktickRenderer(), []);

  /** `rich-textarea` roots with `display: inline-block` — `width: 100%` makes the shell fill the row (see package README). */
  const sizeStyle: CSSProperties = {
    width: "100%",
    minWidth: 0,
    maxWidth: "100%",
    boxSizing: "border-box",
    ...style,
  };

  return (
    <div className="block w-full min-w-0 max-w-full">
      <RichTextarea
        id={id}
        name={name}
        className={cn(
          "placeholder:text-muted-foreground",
          "selection:bg-primary selection:text-primary-foreground",
          "min-h-36 block w-full min-w-0 resize-y font-mono text-caption",
          "border border-border bg-background rounded-md px-3 py-2",
          "shadow-elevation-0 outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/50",
          "disabled:pointer-events-none disabled:opacity-50",
          className
        )}
        style={sizeStyle}
        value={value}
        onChange={onChange}
        disabled={disabled}
        placeholder={placeholder}
        aria-label={ariaLabel}
        aria-describedby={ariaDescribedBy}
        {...rest}
      >
        {renderer}
      </RichTextarea>
    </div>
  );
}
