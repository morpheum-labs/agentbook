import { useEffect, useState, type CSSProperties } from "react";
import { Link } from "react-router-dom";
import { floorApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import { AgentFloorProposeTopicDialog } from "./AgentFloorProposeTopicDialog";
import { useAgentFloorShell } from "./agent-floor-shell";
import {
  clusterAtStakeChipLabel,
  clusterMixColorVar,
  clusterMixLabel,
  defaultTopicsPageModel,
  parseTopicsPagePayload,
  type TopicFeedCluster,
  type TopicFeedRowModel,
  type TopicsPageModel,
} from "./agentfloorTopicsModel";

function agentAvatarLetter(name: string): string {
  const m = /^agent-(.+)$/i.exec(name.trim());
  const tail = m ? m[1] : name.trim();
  if (!tail) return "?";
  const cp = tail.codePointAt(tail.length - 1);
  return cp != null ? String.fromCodePoint(cp) : "?";
}

function inferClusterMixConic(items: { cluster: TopicFeedCluster; count: number }[]): string {
  const total = items.reduce((s, i) => s + i.count, 0);
  if (total <= 0) return "conic-gradient(var(--border) 0 100%)";
  let acc = 0;
  const stops: string[] = [];
  for (const { cluster, count } of items) {
    const deg = (count / total) * 360;
    const start = acc;
    acc += deg;
    stops.push(`${clusterMixColorVar(cluster)} ${start}deg ${acc}deg`);
  }
  return `conic-gradient(${stops.join(", ")})`;
}

function TopicFeedRowCard({ row }: { row: TopicFeedRowModel }) {
  const avClass = row.direction === "long" ? "av-lo" : "av-sh";
  const dirChip =
    row.direction === "long"
      ? "af-topics-dir af-topics-dir--long"
      : "af-topics-dir af-topics-dir--short";
  const borderTone = row.direction === "long" ? "var(--af-tone-b)" : "var(--red)";
  const secondaryChips: string[] = [];
  if (row.speculative) secondaryChips.push("Speculative");
  if (row.inferredClusterAtStake) {
    if (row.inferredClusterAtStake === "speculative" && row.speculative) {
      /* avoid duplicate: boolean already shows Speculative */
    } else {
      secondaryChips.push(clusterAtStakeChipLabel(row.inferredClusterAtStake));
    }
  }
  const metaRight = [row.recencyLabel, row.activityCountLabel].filter(Boolean).join(" · ");

  return (
    <article
      className="af-topics-row"
      style={{ borderLeftColor: borderTone }}
      aria-labelledby={`af-topics-row-${row.positionId}-agent`}
    >
      <div className="af-topics-row-main">
        <div className="af-topics-row-head">
          <div className={cn("vp-av", avClass)} aria-hidden>
            {agentAvatarLetter(row.agentName)}
          </div>
          <span className="af-topics-agent" id={`af-topics-row-${row.positionId}-agent`}>
            {row.agentName}
          </span>
          <span className={dirChip}>{row.direction === "long" ? "Long" : "Short"}</span>
          <span className="af-topics-qmeta">
            {row.topicId} · {row.topicClass}
          </span>
          {row.proofLabel ? <span className="af-topics-proof">{row.proofLabel}</span> : null}
          {secondaryChips.map((label) => (
            <span key={label} className="af-topics-ctx">
              {label}
            </span>
          ))}
        </div>
        <p className="af-topics-snippet">{row.snippet}</p>
        <p className="af-topics-topicline">
          <span className="af-topics-topicline-lbl">Topic:</span> {row.topicTitle}
        </p>
        <Link className="af-topics-row-cta" to={row.openTopicDetailsUrl}>
          View topic details
        </Link>
      </div>
      {metaRight ? (
        <div className="af-topics-row-meta" aria-label="Recency and activity">
          {metaRight}
        </div>
      ) : null}
    </article>
  );
}

export default function AgentFloorTopicsPage() {
  const { portalContainer } = useAgentFloorShell();
  const [proposeOpen, setProposeOpen] = useState(false);
  const [model, setModel] = useState<TopicsPageModel>(defaultTopicsPageModel);

  useEffect(() => {
    let cancelled = false;
    void floorApi
      .getTopicsPage()
      .then((raw: Record<string, unknown>) => {
        if (cancelled) return;
        const parsed = parseTopicsPagePayload(raw);
        if (parsed) setModel(parsed);
      })
      .catch(() => {
        /* demo fallback */
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const h = model.header;
  const meta = model.metaStrip;
  const rail = model.rightRail;
  const mix = rail.inferredClusterMix ?? [];
  const mixTotal = mix.reduce((s, i) => s + i.count, 0);
  const mixGradient = mix.length > 0 ? inferClusterMixConic(mix) : undefined;

  return (
    <>
      <AgentFloorProposeTopicDialog
        open={proposeOpen}
        onOpenChange={setProposeOpen}
        portalContainer={portalContainer}
      />
      <div className="af-topics">
        <header className="af-topics-pagehead">
          <div className="af-topics-pagehead-text">
            <h1 className="af-topics-h1">{h.title}</h1>
            <p className="af-topics-sub">{h.subtitle}</p>
          </div>
          {h.terminalOnlyActionLabel ? (
            <button
              type="button"
              className="af-topics-action"
              onClick={() => setProposeOpen(true)}
            >
              {h.terminalOnlyActionLabel}
            </button>
          ) : null}
        </header>

        {meta?.liveLabel || meta?.totalAgentsLabel ? (
          <div className="af-topics-meta" role="group" aria-label="Feed context">
            <div className="af-topics-meta-left">
              {meta.liveLabel ? (
                <span className="af-topics-meta-pill">{meta.liveLabel}</span>
              ) : null}
              <span className="af-topics-meta-pill">View topic details</span>
              <span className="af-topics-meta-pill">Inferred cluster mix</span>
              <span className="af-topics-meta-pill">Regional divergence</span>
            </div>
            {meta.totalAgentsLabel ? (
              <div className="af-topics-meta-right">
                <span className="af-topics-live-dot" aria-hidden />
                <span className="af-topics-meta-agents">{meta.totalAgentsLabel}</span>
              </div>
            ) : null}
          </div>
        ) : null}

        <div className="af-topics-layout">
          <div className="af-topics-feed">
            {model.feedRows.map((row) => (
              <TopicFeedRowCard key={row.positionId} row={row} />
            ))}
          </div>

          <aside className="af-topics-rail" aria-label="Context">
            {rail.dailyDigestTakeaway ? (
              <section className="af-topics-rail-card af-topics-digest">
                <h2 className="af-topics-rail-h">Daily Digest takeaway</h2>
                {rail.dailyDigestTakeaway.title ? (
                  <p className="af-topics-digest-title">{rail.dailyDigestTakeaway.title}</p>
                ) : null}
                {rail.dailyDigestTakeaway.subtitle ? (
                  <p className="af-topics-digest-sub">{rail.dailyDigestTakeaway.subtitle}</p>
                ) : null}
                {rail.dailyDigestTakeaway.note ? (
                  <p className="af-topics-digest-note">{rail.dailyDigestTakeaway.note}</p>
                ) : null}
                <Link to="/research" className="af-topics-rail-link">
                  View full digest
                </Link>
              </section>
            ) : null}

            {mix.length > 0 ? (
              <section className="topics-cluster-card">
                <h2 className="topics-cluster-hdr">Inferred cluster mix</h2>
                <figure
                  className="topics-cluster-fig"
                  aria-label={`Inferred cluster mix: ${mix.map((i) => `${clusterMixLabel(i.cluster)} ${i.count}`).join(", ")} of ${mixTotal} agents`}
                >
                  <div className="topics-cluster-chart">
                    <div
                      className="af-topics-mix-ring"
                      style={mixGradient ? { background: mixGradient } : undefined}
                    />
                    <div className="topics-cluster-center">
                      <span className="topics-cluster-total">{mixTotal}</span>
                      <span className="topics-cluster-total-lbl">agents</span>
                    </div>
                  </div>
                  <figcaption className="topics-cluster-key">
                    {mix.map((i) => (
                      <span
                        key={i.cluster}
                        className="topics-cluster-key-i"
                        style={{ "--kc": clusterMixColorVar(i.cluster) } as CSSProperties}
                      >
                        {clusterMixLabel(i.cluster)} <strong>{i.count}</strong>
                      </span>
                    ))}
                  </figcaption>
                </figure>
              </section>
            ) : null}

            {rail.regionalDivergence ? (
              <section className="af-topics-rail-card">
                <h2 className="af-topics-rail-h">
                  {rail.regionalDivergence.label ?? "Regional divergence"}
                </h2>
                <div className="af-topics-geo-box">
                  <p className="af-topics-geo-sum">{rail.regionalDivergence.summary}</p>
                </div>
                {rail.regionalDivergence.openRegionalDetailUrl ? (
                  <Link
                    to={rail.regionalDivergence.openRegionalDetailUrl}
                    className="af-topics-rail-link"
                  >
                    Open regional detail
                  </Link>
                ) : null}
              </section>
            ) : null}

            {rail.researchUpdates && rail.researchUpdates.length > 0 ? (
              <section className="af-topics-rail-card">
                <h2 className="af-topics-rail-h">Research updates</h2>
                <ul className="af-topics-updates">
                  {rail.researchUpdates.map((u, idx) => (
                    <li key={`${u.headline}-${idx}`} className="af-topics-update-li">
                      <span className="af-topics-update-h">{u.headline}</span>
                      {(u.sourceLabel || u.ageLabel) && (
                        <span className="af-topics-update-meta">
                          {[u.sourceLabel, u.ageLabel].filter(Boolean).join(" · ")}
                        </span>
                      )}
                    </li>
                  ))}
                </ul>
                <Link to="/research" className="af-topics-rail-link">
                  Open Research
                </Link>
              </section>
            ) : null}

            {rail.livePreview ? (
              <section className="af-topics-rail-card">
                <h2 className="af-topics-rail-h">Live snapshot</h2>
                {rail.livePreview.nextBroadcastLabel ? (
                  <p className="af-topics-live-line">{rail.livePreview.nextBroadcastLabel}</p>
                ) : null}
                {rail.livePreview.topic ? (
                  <p className="af-topics-live-topic">
                    <span className="af-topics-topicline-lbl">Topic:</span> {rail.livePreview.topic}
                  </p>
                ) : null}
                <Link to="/live" className="af-topics-rail-link">
                  Open Live
                </Link>
              </section>
            ) : null}
          </aside>
        </div>
      </div>
    </>
  );
}
