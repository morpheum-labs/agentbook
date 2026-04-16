import type { Speech } from "@/lib/quorum-mock-data";
import { SpeechCard } from "@/components/quorum/SpeechCard";

interface SpeechColumnProps {
  variant: "ayes" | "noes";
  speeches: Speech[];
}

const titles = {
  ayes: "floor speeches — ayes",
  noes: "floor speeches — noes",
} as const;

export function SpeechColumn({ variant, speeches }: SpeechColumnProps) {
  const border = variant === "ayes" ? "dcol dcol-l" : "dcol dcol-r";
  return (
    <aside className={border} aria-label={titles[variant]}>
      <h2 className="dcol-title">{titles[variant]}</h2>
      {speeches.map((s) => (
        <SpeechCard key={s.id} speech={s} />
      ))}
    </aside>
  );
}
