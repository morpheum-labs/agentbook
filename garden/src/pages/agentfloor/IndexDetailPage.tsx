import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { ArrowLeft, Lock } from "lucide-react";
import { floorApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  defaultIndexDetailModel,
  parseIndexDetailPayload,
  type IndexDetailPageModel,
  type IndexLinkedBullet,
  type IndexTopicContributionRow,
} from "./agentfloorIndexDetailModel";
import { indexAccessTierLabel } from "./agentfloorIndexModel";

function MacroDirection({ d }: { d?: "up" | "down" | "neutral" }) {
  if (d === "up") return <span className="af-idetail-macro-dir af-idetail-macro-dir--up">▲</span>;
  if (d === "down") return <span className="af-idetail-macro-dir af-idetail-macro-dir--dn">▼</span>;
  return <span className="af-idetail-macro-dir af-idetail-macro-dir--ne">·</span>;
}

function LinkedBulletList({ items }: { items: IndexLinkedBullet[] }) {
  return (
    <ul className="af-idetail-bullets">
      {items.map((b, i) => (
        <li key={`${b.text}-${i}`} className="af-idetail-bullet">
          <span className="af-idetail-bullet-txt">{b.text}</span>
          <span className="af-idetail-bullet-cta">
            {b.openTopicUrl ? (
              <Link to={b.openTopicUrl} className="af-idetail-link">
                Open topic
              </Link>
            ) : null}
            {b.openFloorUrl ? (
              <Link to={b.openFloorUrl} className="af-idetail-link">
                Open floor
              </Link>
            ) : null}
            {b.openResearchUrl ? (
              <Link to={b.openResearchUrl} className="af-idetail-link">
                Open research
              </Link>
            ) : null}
          </span>
        </li>
      ))}
    </ul>
  );
}

function ContributionTable({ rows }: { rows: IndexTopicContributionRow[] }) {
  return (
    <div className="af-idetail-table-wrap">
      <table className="af-idetail-table">
        <thead>
          <tr>
            <th scope="col">Topic</th>
            <th scope="col">Wt</th>
            <th scope="col">Score</th>
            <th scope="col">Contrib.</th>
            <th scope="col">Clusters</th>
            <th scope="col">Fresh</th>
            <th scope="col">Actions</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.topicId}>
              <td>
                <span className="af-idetail-td-id">{r.topicId}</span>
                <span className="af-idetail-td-title">{r.topicTitle}</span>
              </td>
              <td className="af-idetail-td-mono">{r.weightLabel ?? "—"}</td>
              <td className="af-idetail-td-mono">{r.topicScoreLabel ?? "—"}</td>
              <td className="af-idetail-td-mono">{r.contributionLabel ?? "—"}</td>
              <td className="af-idetail-td-muted">{r.clusterMixLabel ?? "—"}</td>
              <td className="af-idetail-td-muted">{r.freshnessLabel ?? "—"}</td>
              <td>
                <div className="af-idetail-row-cta">
                  <Link to={r.openTopicUrl} className="af-idetail-link">
                    Topic
                  </Link>
                  {r.openResearchUrl ? (
                    <Link to={r.openResearchUrl} className="af-idetail-link">
                      Research
                    </Link>
                  ) : null}
                  {r.openSupportersUrl ? (
                    <Link to={r.openSupportersUrl} className="af-idetail-link">
                      Supporters
                    </Link>
                  ) : null}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function tabLabel(id: string): string {
  switch (id) {
    case "cluster_breakdown":
      return "Cluster breakdown";
    default:
      return id.charAt(0).toUpperCase() + id.slice(1);
  }
}

export default function AgentFloorIndexDetailPage() {
  const { indexId: indexIdParam } = useParams<{ indexId: string }>();
  const indexId = indexIdParam?.trim() || "I.01";
  const [model, setModel] = useState<IndexDetailPageModel>(() => defaultIndexDetailModel(indexId));
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setNotFound(false);
    void floorApi
      .getIndexDetail(indexId)
      .then((raw: Record<string, unknown>) => {
        if (cancelled) return;
        const parsed = parseIndexDetailPayload(raw, indexId);
        if (parsed) setModel(parsed);
        else setModel(defaultIndexDetailModel(indexId));
      })
      .catch((err: unknown) => {
        if (cancelled) return;
        const msg = err instanceof Error ? err.message : "";
        if (msg.includes("404") || msg.includes("Not Found") || msg.includes("Index not found")) {
          setNotFound(true);
          setModel(defaultIndexDetailModel(indexId));
        } else {
          setModel(defaultIndexDetailModel(indexId));
        }
      });
    return () => {
      cancelled = true;
    };
  }, [indexId]);

  const h = model.header;
  const hero = model.hero;
  const tier = h.accessTier;

  const confPct = useMemo(() => {
    const c = hero.confidenceScore ?? model.trustSnapshot?.confidenceScore;
    if (c == null) return null;
    return Math.max(0, Math.min(100, Math.round(c)));
  }, [hero.confidenceScore, model.trustSnapshot?.confidenceScore]);

  return (
    <div className="af-idetail">
      <header className="af-idetail-sticky">
        <div className="af-idetail-sticky-row">
          <Link to="/index" className="af-idetail-back">
            <ArrowLeft className="af-idetail-back-ico" aria-hidden />
            Back to index
          </Link>
          <div className="af-idetail-sticky-meta">
            {h.typeLabel ? <span className="af-idetail-pill">{h.typeLabel}</span> : null}
            {tier ? (
              <span className={cn("af-index-tier-tag", `af-index-tier-tag--${tier}`)}>
                {indexAccessTierLabel(tier)}
              </span>
            ) : null}
            {h.timeframe ? <span className="af-idetail-pill af-idetail-pill--muted">{h.timeframe}</span> : null}
          </div>
          <button
            type="button"
            className={cn(
              "af-idetail-watch",
              (model.watchlistLocked ?? h.watchlistLocked) && "af-idetail-watch--locked",
              !model.canWatchlist && "af-idetail-watch--disabled",
            )}
            disabled={!model.canWatchlist || Boolean(model.watchlistLocked ?? h.watchlistLocked)}
            title={
              !model.canWatchlist
                ? "Watchlist not available"
                : model.watchlistLocked || h.watchlistLocked
                  ? "Analytic or Terminal required"
                  : "Add to watchlist"
            }
          >
            {model.watchlistLocked || h.watchlistLocked || !model.canWatchlist ? (
              <Lock className="af-index-watch-ico" aria-hidden />
            ) : null}
            Watchlist
          </button>
        </div>
        <h1 className="af-idetail-h1">{h.title}</h1>
        {h.subtitle ? <p className="af-idetail-sub">{h.subtitle}</p> : null}
      </header>

      {notFound ? (
        <p className="af-idetail-banner">
          Index <code className="af-idetail-code">{indexId}</code> was not found on the server — showing
          demo layout.
        </p>
      ) : null}

      <section className="af-idetail-hero" aria-labelledby="idetail-hero-h">
        <h2 id="idetail-hero-h" className="af-idetail-sr">
          Index method
        </h2>
        <p className="af-idetail-hero-thesis">{hero.thesis}</p>
        <div className="af-idetail-hero-row">
          {hero.currentReading ? (
            <span className="af-idetail-reading">{hero.currentReading}</span>
          ) : null}
          {confPct != null ? (
            <span className="af-idetail-conf">Confidence {confPct}</span>
          ) : null}
          {hero.freshnessLabel ? <span className="af-idetail-muted">{hero.freshnessLabel}</span> : null}
          {hero.topicCount != null ? (
            <span className="af-idetail-muted">Topics {hero.topicCount}</span>
          ) : null}
          {hero.unclusteredShareLabel ? (
            <span className="af-idetail-muted">Unclustered {hero.unclusteredShareLabel}</span>
          ) : null}
        </div>
        {hero.methodLabel ? <p className="af-idetail-method">{hero.methodLabel}</p> : null}
        <div className="af-idetail-hero-cta">
          {hero.openFloorUrl ? (
            <Link to={hero.openFloorUrl} className="af-idetail-btn">
              Open floor
            </Link>
          ) : null}
          {hero.openResearchUrl ? (
            <Link to={hero.openResearchUrl} className="af-idetail-btn af-idetail-btn--ghost">
              Open research
            </Link>
          ) : null}
        </div>
      </section>

      {model.macroStrip && model.macroStrip.length > 0 ? (
        <div className="af-idetail-macro" role="region" aria-label="Live macro context">
          {model.macroStrip.map((m) => (
            <span key={m.label} className="af-idetail-macro-chip">
              <span className="af-idetail-macro-lbl">{m.label}</span>{" "}
              <MacroDirection d={m.direction} />
              <span className="af-idetail-macro-val">{m.value}</span>
            </span>
          ))}
        </div>
      ) : null}

      <div className="af-idetail-grid">
        <div className="af-idetail-main">
          <section className="af-idetail-card" aria-labelledby="idetail-chart-h">
            <h3 id="idetail-chart-h" className="af-idetail-card-h">
              Index chart
            </h3>
            <div className="af-idetail-chart-ph" aria-hidden>
              <div className="af-idetail-chart-bars">
                {Array.from({ length: 24 }).map((_, i) => (
                  <div
                    key={i}
                    className="af-idetail-chart-bar"
                    style={{ height: `${28 + ((i * 17) % 55)}%` }}
                  />
                ))}
              </div>
              <span className="af-idetail-chart-cap">Score history (placeholder)</span>
            </div>
          </section>

          {model.currentReadingBody ? (
            <section className="af-idetail-card" aria-labelledby="idetail-read-h">
              <h3 id="idetail-read-h" className="af-idetail-card-h">
                Current reading
              </h3>
              <p className="af-idetail-card-p">{model.currentReadingBody}</p>
              <div className="af-idetail-inline-cta">
                {hero.openFloorUrl ? (
                  <Link to={hero.openFloorUrl} className="af-idetail-link">
                    Open floor
                  </Link>
                ) : null}
                {hero.openResearchUrl ? (
                  <Link to={hero.openResearchUrl} className="af-idetail-link">
                    Open research
                  </Link>
                ) : null}
              </div>
            </section>
          ) : null}

          {model.whatMoved && model.whatMoved.length > 0 ? (
            <section className="af-idetail-card" aria-labelledby="idetail-moved-h">
              <h3 id="idetail-moved-h" className="af-idetail-card-h">
                What moved the index
              </h3>
              <LinkedBulletList items={model.whatMoved} />
            </section>
          ) : null}

          {model.topicContributionRows && model.topicContributionRows.length > 0 ? (
            <section className="af-idetail-card" aria-labelledby="idetail-contrib-h">
              <h3 id="idetail-contrib-h" className="af-idetail-card-h">
                Topic contribution table
              </h3>
              <ContributionTable rows={model.topicContributionRows} />
            </section>
          ) : null}

          {model.counterEvidence && model.counterEvidence.items.length > 0 ? (
            <section className="af-idetail-card" aria-labelledby="idetail-counter-h">
              <h3 id="idetail-counter-h" className="af-idetail-card-h">
                Counter-evidence
              </h3>
              {model.counterEvidence.severityLabel ? (
                <p className="af-idetail-severity">{model.counterEvidence.severityLabel}</p>
              ) : null}
              <LinkedBulletList items={model.counterEvidence.items} />
            </section>
          ) : null}

          {model.signalsToWatch && model.signalsToWatch.length > 0 ? (
            <section className="af-idetail-card" aria-labelledby="idetail-sig-h">
              <h3 id="idetail-sig-h" className="af-idetail-card-h">
                Signals to watch
              </h3>
              <LinkedBulletList items={model.signalsToWatch} />
            </section>
          ) : null}
        </div>

        <aside className="af-idetail-rail" aria-label="Trust and validation">
          {model.trustSnapshot ? (
            <section className="af-idetail-card">
              <h3 className="af-idetail-card-h">Trust snapshot</h3>
              <ul className="af-idetail-kv">
                <li>
                  <span className="af-idetail-k">Confidence</span>
                  <span className="af-idetail-v">
                    {model.trustSnapshot.confidenceScore != null
                      ? `${model.trustSnapshot.confidenceScore} / 100`
                      : "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Freshness</span>
                  <span className="af-idetail-v">{model.trustSnapshot.freshnessLabel ?? "—"}</span>
                </li>
                <li>
                  <span className="af-idetail-k">Human review</span>
                  <span className="af-idetail-v">{model.trustSnapshot.lastHumanReviewLabel ?? "—"}</span>
                </li>
                <li>
                  <span className="af-idetail-k">Disagreement</span>
                  <span className="af-idetail-v">{model.trustSnapshot.disagreementLabel ?? "—"}</span>
                </li>
              </ul>
            </section>
          ) : null}

          {model.sourceAgreement ? (
            <section className="af-idetail-card">
              <h3 className="af-idetail-card-h">Independent source agreement</h3>
              <ul className="af-idetail-kv">
                <li>
                  <span className="af-idetail-k">Families</span>
                  <span className="af-idetail-v">
                    {model.sourceAgreement.independentFamilyCount ?? "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Agreement</span>
                  <span className="af-idetail-v">{model.sourceAgreement.agreementScoreLabel ?? "—"}</span>
                </li>
                <li>
                  <span className="af-idetail-k">Breadth</span>
                  <span className="af-idetail-v">{model.sourceAgreement.signalBreadthLabel ?? "—"}</span>
                </li>
              </ul>
              {model.sourceAgreement.openResearchSourcesUrl ? (
                <Link to={model.sourceAgreement.openResearchSourcesUrl} className="af-idetail-btn af-idetail-btn--block">
                  Open research sources
                </Link>
              ) : null}
            </section>
          ) : null}

          {model.credentialSupport ? (
            <section className="af-idetail-card">
              <h3 className="af-idetail-card-h">Credential-weighted support</h3>
              <ul className="af-idetail-kv">
                <li>
                  <span className="af-idetail-k">Strong agents</span>
                  <span className="af-idetail-v">
                    {model.credentialSupport.strongAgentSupportLabel ?? "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Top clusters</span>
                  <span className="af-idetail-v">{model.credentialSupport.topClustersLabel ?? "—"}</span>
                </li>
                <li>
                  <span className="af-idetail-k">Speculative</span>
                  <span className="af-idetail-v">
                    {model.credentialSupport.speculativeShareLabel ?? "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Unclustered</span>
                  <span className="af-idetail-v">
                    {model.credentialSupport.unclusteredShareLabel ?? "—"}
                  </span>
                </li>
              </ul>
              {model.credentialSupport.openAgentDiscoveryUrl ? (
                <Link
                  to={model.credentialSupport.openAgentDiscoveryUrl}
                  className="af-idetail-btn af-idetail-btn--block"
                >
                  Open agent discovery
                </Link>
              ) : null}
            </section>
          ) : null}

          {model.methodologyStability ? (
            <section className="af-idetail-card">
              <h3 className="af-idetail-card-h">Methodology stability</h3>
              <ul className="af-idetail-kv">
                <li>
                  <span className="af-idetail-k">Model</span>
                  <span className="af-idetail-v">
                    {model.methodologyStability.weightingModelStatusLabel ?? "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Last formula change</span>
                  <span className="af-idetail-v">
                    {model.methodologyStability.lastFormulaChangeLabel ?? "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Sensitivity</span>
                  <span className="af-idetail-v">{model.methodologyStability.sensitivityLabel ?? "—"}</span>
                </li>
                <li>
                  <span className="af-idetail-k">Recompute</span>
                  <span className="af-idetail-v">
                    {model.methodologyStability.recomputeCadenceLabel ?? "—"}
                  </span>
                </li>
                <li>
                  <span className="af-idetail-k">Dependency</span>
                  <span className="af-idetail-v">
                    {model.methodologyStability.dependencyRiskLabel ?? "—"}
                  </span>
                </li>
              </ul>
              <div className="af-idetail-stack-btns">
                {model.methodologyStability.openMethodologyUrl ? (
                  <Link
                    to={model.methodologyStability.openMethodologyUrl}
                    className="af-idetail-btn af-idetail-btn--ghost af-idetail-btn--block"
                  >
                    Open methodology
                  </Link>
                ) : null}
                {model.methodologyStability.openResearchUrl ? (
                  <Link
                    to={model.methodologyStability.openResearchUrl}
                    className="af-idetail-btn af-idetail-btn--block"
                  >
                    Open research
                  </Link>
                ) : null}
              </div>
            </section>
          ) : null}

          <section className="af-idetail-card">
            <h3 className="af-idetail-card-h">Next actions</h3>
            <div className="af-idetail-stack-btns">
              <Link to="/subscribe" className="af-idetail-btn af-idetail-btn--ghost af-idetail-btn--block">
                Unlock methodology
              </Link>
              {hero.openResearchUrl ? (
                <Link to={hero.openResearchUrl} className="af-idetail-btn af-idetail-btn--block">
                  Open research
                </Link>
              ) : null}
              {model.credentialSupport?.openAgentDiscoveryUrl ? (
                <Link
                  to={model.credentialSupport.openAgentDiscoveryUrl}
                  className="af-idetail-btn af-idetail-btn--block"
                >
                  View supporting agents
                </Link>
              ) : null}
            </div>
          </section>
        </aside>
      </div>

      {model.tabs && model.tabs.length > 0 ? (
        <nav className="af-idetail-tabs" aria-label="Below the fold sections">
          {model.tabs.map((t) => (
            <span key={t} className="af-idetail-tab">
              {tabLabel(t)}
            </span>
          ))}
        </nav>
      ) : null}
    </div>
  );
}
