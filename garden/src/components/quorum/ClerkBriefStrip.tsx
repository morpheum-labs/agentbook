import type { ClerkItem } from "@/lib/quorum-mock-data";

interface ClerkBriefStripProps {
  items: ClerkItem[];
  timestamp: string;
}

function clerkToneClass(tone: ClerkItem["tone"]): string {
  switch (tone) {
    case "c":
      return "citem ci-c";
    case "d":
      return "citem ci-d";
    case "n":
      return "citem ci-n";
    case "r":
      return "citem ci-r";
    default:
      return "citem ci-n";
  }
}

export function ClerkBriefStrip({ items, timestamp }: ClerkBriefStripProps) {
  return (
    <section className="clerk" aria-labelledby="quorum-clerk-heading">
      <h2 id="quorum-clerk-heading" className="clerk-title">
        CLERK&apos;S BRIEF
      </h2>
      <div className="clerk-items">
        {items.map((item) => (
          <div key={item.id} className={clerkToneClass(item.tone)}>
            {item.text}
          </div>
        ))}
      </div>
      <div className="clerk-ts">{timestamp}</div>
    </section>
  );
}
