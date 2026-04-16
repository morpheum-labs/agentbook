import { Fragment } from "react";
import type { FactionRow } from "@/lib/quorum-mock-data";

interface FactionLegendProps {
  factions: FactionRow[];
  seatsSummary: string;
}

function factDotClass(tone: FactionRow["tone"]): string {
  switch (tone) {
    case "bull":
      return "fact-dot fact-bull";
    case "bear":
      return "fact-dot fact-bear";
    case "neut":
      return "fact-dot fact-neut";
    case "spec":
      return "fact-dot fact-spec";
    default:
      return "fact-dot fact-neut";
  }
}

function factLblClass(tone: FactionRow["tone"]): string {
  switch (tone) {
    case "bull":
      return "fact-lbl-bull";
    case "bear":
      return "fact-lbl-bear";
    case "neut":
      return "fact-lbl-neut";
    case "spec":
      return "fact-lbl-spec";
    default:
      return "fact-lbl-neut";
  }
}

export function FactionLegend({ factions, seatsSummary }: FactionLegendProps) {
  return (
    <section className="factions" aria-labelledby="quorum-factions-heading">
      <h2 id="quorum-factions-heading" className="fleg">
        FACTIONS
      </h2>
      {factions.map((f, i) => (
        <Fragment key={f.label}>
          {i > 0 ? <span className="fsep">·</span> : null}
          <div className="fact">
            <span className={factDotClass(f.tone)} aria-hidden />
            <span className={factLblClass(f.tone)}>{f.label}</span>
            <span className="fact-count">{f.count}</span>
          </div>
        </Fragment>
      ))}
      <span className="fsep">·</span>
      <p className="factions-meta">{seatsSummary}</p>
    </section>
  );
}
