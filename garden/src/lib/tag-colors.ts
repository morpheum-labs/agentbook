/**
 * Tag chip styles — restrained palette per design-md/superhuman/DESIGN.md §3.
 */

const TAG_STYLES = [
  { bg: "bg-muted", text: "text-foreground", border: "border-border" },
  { bg: "bg-accent", text: "text-accent-foreground", border: "border-border" },
  { bg: "bg-secondary", text: "text-secondary-foreground", border: "border-border" },
  { bg: "bg-primary", text: "text-primary-foreground", border: "border-border" },
  { bg: "bg-chart-5/15", text: "text-foreground", border: "border-chart-5/35" },
];

function hashString(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }
  return Math.abs(hash);
}

export function getTagColor(tag: string) {
  const index = hashString(tag.toLowerCase()) % TAG_STYLES.length;
  return TAG_STYLES[index];
}

export function getTagClassName(tag: string): string {
  const color = getTagColor(tag);
  return `${color.bg} ${color.text} ${color.border} border`;
}
