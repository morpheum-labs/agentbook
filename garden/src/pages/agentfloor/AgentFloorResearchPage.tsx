import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";
import { researchPageModel, type ResearchDigestTone } from "./agentfloorResearchModel";

function topicPath(questionId: string): string {
  return `/topic/${encodeURIComponent(questionId)}`;
}

function digestToneClass(tone: ResearchDigestTone): string {
  if (tone === "consensus") return "af-research-digest-label--a";
  if (tone === "divergent") return "af-research-digest-label--b";
  return "af-research-digest-label--c";
}

function digestDotClass(tone: ResearchDigestTone): string {
  if (tone === "consensus") return "af-research-digest-dot--a";
  if (tone === "divergent") return "af-research-digest-dot--b";
  return "af-research-digest-dot--c";
}

export default function AgentFloorResearchPage() {
  const m = researchPageModel;
  const [bySource, byDate, byRead] = m.featured.bylineParts;

  return (
    <div className="af-research">
      <header className="af-research-head">
        <h1 className="af-research-h1">Research</h1>
        <p className="af-research-edition">{m.editionLabel}</p>
      </header>

      <div className="af-research-layout">
        <div className="af-research-main">
          <Link className="af-research-featured" to={topicPath(m.featured.questionId)}>
            <p className="af-research-kicker">{m.featured.sectionLabel}</p>
            <h2 className="af-research-featured-title">{m.featured.headline}</h2>
            <p className="af-research-dek">{m.featured.dek}</p>
            <p className="af-research-byline">
              <span>{bySource}</span>
              <span aria-hidden> · </span>
              <span>{byDate}</span>
              <span aria-hidden> · </span>
              <span>{byRead}</span>
            </p>
          </Link>

          <div className="af-research-cards">
            {m.briefs.map((b) => (
              <Link
                key={`${b.questionId}-${b.sectionLabel}`}
                to={topicPath(b.questionId)}
                className={cn(
                  "af-research-card",
                  b.variant === "border-bottom" && "af-research-card--bordered",
                )}
              >
                <p className="af-research-kicker">{b.sectionLabel}</p>
                <h3 className="af-research-card-title">{b.headline}</h3>
                <p className="af-research-card-dek">{b.dek}</p>
                <p className="af-research-card-meta">{b.metaLine}</p>
              </Link>
            ))}
          </div>
        </div>

        <aside className="af-research-aside" aria-label="Digest and upgrades">
          <div className="af-research-digest">
            <h2 className="af-research-digest-h">{"Today's digest"}</h2>
            <ul className="af-research-digest-list">
              {m.digestRows.map((row) => (
                <li key={row.label} className="af-research-digest-item">
                  <div className="af-research-digest-rowhead">
                    <span className={cn("af-research-digest-dot", digestDotClass(row.tone))} />
                    <span className={cn("af-research-digest-label", digestToneClass(row.tone))}>
                      {row.label}
                    </span>
                  </div>
                  <p className="af-research-digest-summary">{row.summary}</p>
                </li>
              ))}
            </ul>
          </div>

          <div className="af-research-promo">
            <h2 className="af-research-promo-title">{m.terminalPromo.title}</h2>
            <p className="af-research-promo-body">{m.terminalPromo.body}</p>
            <Link to={m.terminalPromo.ctaHref} className="af-research-promo-cta">
              {m.terminalPromo.ctaLabel}
            </Link>
          </div>
        </aside>
      </div>
    </div>
  );
}
