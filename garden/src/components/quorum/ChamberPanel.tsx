import type { FeaturedMotion, MotionRow } from "@/lib/quorum-mock-data";
import { SeatMapSvg } from "@/components/quorum/SeatMapSvg";

interface ChamberPanelProps {
  featured: FeaturedMotion;
  motionRows: MotionRow[];
}

export function ChamberPanel({ featured, motionRows }: ChamberPanelProps) {
  const { voteTrack } = featured;
  return (
    <section className="chamber" aria-labelledby="quorum-chamber-heading">
      <h2 id="quorum-chamber-heading" className="sr-only">
        Chamber centre
      </h2>
      <div className="motion-label">
        {featured.label}
        {featured.badges.map((b) => (
          <span key={b.text} className={b.variant === "green" ? "motion-badge g" : "motion-badge"}>
            {b.text}
          </span>
        ))}
      </div>

      <div className="motion-featured">
        <div className="mf-num">{featured.numLine}</div>
        <div className="mf-title">{featured.title}</div>
        <div className="mf-sub">{featured.sub}</div>

        <div className="vbar-wrap">
          <div className="vbar-labels">
            <span className="vbar-label-a">{featured.voteLabels.left}</span>
            <span className="vbar-label-mid">{featured.voteLabels.mid}</span>
            <span className="vbar-label-n">{featured.voteLabels.right}</span>
          </div>
          <div className="vbar-track">
            <div className="vbar-aye" style={{ width: `${voteTrack.aye}%` }} />
            <div className="vbar-abst" style={{ width: `${voteTrack.abstain}%` }} />
            <div className="vbar-noe" style={{ width: `${voteTrack.noe}%` }} />
          </div>
          <div className="vbar-vals">
            <span className="vv-a">{featured.voteVals.left}</span>
            <span className="vv-ab">{featured.voteVals.mid}</span>
            <span className="vv-n">{featured.voteVals.right}</span>
          </div>
        </div>

        <div className="mkt-row">
          <div className="mkt-opt mkt-a">
            <div className="mkt-oname">{featured.marketLeft.name}</div>
            <div className="mkt-opct mkt-opct-a">{featured.marketLeft.pct}</div>
            <div className="mkt-oagents">{featured.marketLeft.agents}</div>
          </div>
          <div className="mkt-opt mkt-b">
            <div className="mkt-oname">{featured.marketRight.name}</div>
            <div className="mkt-opct mkt-opct-b">{featured.marketRight.pct}</div>
            <div className="mkt-oagents">{featured.marketRight.agents}</div>
          </div>
        </div>

        <div className="seats-wrap">
          <div className="seats-label">CHAMBER SEAT MAP · faction positions</div>
          <SeatMapSvg />
        </div>
      </div>

      <div className="mlist" role="list">
        {motionRows.map((row) => (
          <div key={row.id} className="mrow" role="listitem">
            <div className={row.numHighlight ? "mrow-num h" : "mrow-num"}>{row.num}</div>
            <div className="mrow-body">
              <div className="mrow-title">{row.title}</div>
              <div className="mrow-sub">{row.sub}</div>
            </div>
            <div className="mrow-right">
              <div className={`mrow-pct ${row.pctClass}`}>{row.pct}</div>
              <div className="mrow-bar">
                <div className="mrow-fill" style={{ width: `${row.barWidthPct}%`, background: row.barColor }} />
              </div>
              <div className={`mrow-tag ${row.tagClass}`}>{row.tag}</div>
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
