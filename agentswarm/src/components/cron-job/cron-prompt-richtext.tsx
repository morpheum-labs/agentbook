import { useMemo, type CSSProperties, type ReactNode } from "react";
import { RichTextarea, type RichTextareaProps, type Renderer } from "rich-textarea";
import { cn } from "@/lib/utils";

const TOOL_BACKTICK = /`[^`]+`/g;

const classBacktickUnvalidated = cn(
  "rounded-sm px-0.5 [font:inherit] [letter-spacing:inherit] [text-shadow:none]",
  "text-green-800 ring-1 ring-inset ring-green-600/25",
  "bg-emerald-500/15",
  "dark:text-emerald-200 dark:ring-emerald-400/25 dark:bg-emerald-500/12"
);

const classBacktickOnAllowlist = cn(
  "rounded-sm px-0.5 [font:inherit] [letter-spacing:inherit] [text-shadow:none]",
  "text-green-800 ring-1 ring-inset ring-green-600/35",
  "bg-emerald-500/20",
  "dark:text-emerald-200 dark:ring-emerald-400/35 dark:bg-emerald-500/18"
);

const classBacktickNotOnAllowlist = cn(
  "rounded-sm px-0.5 [font:inherit] [letter-spacing:inherit] [text-shadow:none]",
  "text-destructive ring-1 ring-inset ring-destructive/30",
  "bg-destructive/10",
  "dark:text-red-300 dark:ring-red-400/35 dark:bg-red-950/45"
);

function isToolOnAllowlist(inner: string, list: string[]): boolean {
  const t = inner.trim();
  if (!t) return false;
  return list.some((n) => n === t || n.toLowerCase() === t.toLowerCase());
}

/**
 * Renders the mirror layer: plain text and styled `` `name` `` spans. When `list` is set,
 * `` `name` `` is green on the hand’s allowlist, red if not. When `list` is null, backticks
 * all use the legacy green “tool ref” look (no pass/fail).
 */
function makeBacktickRenderer(list: string[] | null): Renderer {
  return (value: string) => {
    if (!value) return null;
    const out: ReactNode[] = [];
    let k = 0;
    let last = 0;
    TOOL_BACKTICK.lastIndex = 0;
    let m: RegExpExecArray | null;
    while ((m = TOOL_BACKTICK.exec(value)) !== null) {
      const full = m[0];
      const start = m.index;
      if (start > last) {
        out.push(value.slice(last, start));
      }
      const inner = full.slice(1, -1);
      let className: string;
      if (list == null) {
        className = classBacktickUnvalidated;
      } else {
        className = isToolOnAllowlist(inner, list)
          ? classBacktickOnAllowlist
          : classBacktickNotOnAllowlist;
      }
      out.push(
        <span key={`b-${k++}`} className={className}>
          {full}
        </span>
      );
      last = start + full.length;
    }
    if (last < value.length) {
      out.push(value.slice(last));
    }
    return <>{out}</>;
  };
}

export type CronJobPromptRichtextProps = Omit<
  RichTextareaProps,
  "children" | "defaultValue" | "ref"
> & {
  /**
   * Selected hand’s `SwarmAgent.Tools` from the control plane. When set, each `` `token` ``
   * is green if it is on the list, red if not. When `undefined` (e.g. no hand to resolve),
   * all backticked tool refs get the same neutral green styling (not validated).
   */
  allowedToolNames?: string[];
};

export function CronJobPromptRichtext({
  className,
  style,
  id,
  name,
  value,
  onChange,
  disabled,
  placeholder,
  allowedToolNames,
  "aria-label": ariaLabel,
  "aria-describedby": ariaDescribedBy,
  ...rest
}: CronJobPromptRichtextProps) {
  const allowKey = allowedToolNames === undefined ? "∅" : allowedToolNames.join("\0");
  const renderer = useMemo(() => {
    return makeBacktickRenderer(allowedToolNames === undefined ? null : allowedToolNames);
  }, [allowKey, allowedToolNames]);

  /** `rich-textarea` roots with `display: inline-block` — `width: 100%` makes the shell fill the row. */
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
