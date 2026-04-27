export type FivePartCron = [string, string, string, string, string];

export const DEFAULT_FIVE: FivePartCron = ["*", "*", "*", "*", "*"];

/** Return five cron tokens if the string is exactly five whitespace-separated parts. */
export function tryParseFivePart(value: string): FivePartCron | null {
  const t = value.trim();
  if (!t) return null;
  const parts = t.split(/\s+/);
  if (parts.length === 5 && parts.every((p) => p.length > 0)) {
    return [parts[0]!, parts[1]!, parts[2]!, parts[3]!, parts[4]!];
  }
  return null;
}

export function initialBuilderParts(value: string): FivePartCron {
  return tryParseFivePart(value) ?? DEFAULT_FIVE;
}

/** Freeform / non-5-field label: use custom textarea mode. */
export function shouldUseCustomExpression(value: string): boolean {
  const t = value.trim();
  if (!t) return false;
  return tryParseFivePart(value) === null;
}

export const CRON_FIELD_LABELS = [
  { key: "min" as const, name: "Minute", hint: "0–59, *, */n" },
  { key: "hour" as const, name: "Hour", hint: "0–23, *" },
  { key: "dom" as const, name: "Day", hint: "1–31, *" },
  { key: "mon" as const, name: "Month", hint: "1–12, *" },
  { key: "dow" as const, name: "Weekday", hint: "0–6, *" },
] as const;

export const CRON_PRESETS: { label: string; value: string }[] = [
  { label: "Every minute", value: "* * * * *" },
  { label: "Every 5 min", value: "*/5 * * * *" },
  { label: "Every 15 min", value: "*/15 * * * *" },
  { label: "Hourly", value: "0 * * * *" },
  { label: "Daily (midnight)", value: "0 0 * * *" },
  { label: "Daily (9:00)", value: "0 9 * * *" },
  { label: "Weekdays 9:00", value: "0 9 * * 1-5" },
  { label: "Sundays midnight", value: "0 0 * * 0" },
  { label: "1st of month, midnight", value: "0 0 1 * *" },
];

export function buildExpression(parts: FivePartCron): string {
  return parts.join(" ");
}
