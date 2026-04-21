import { useCallback, useEffect, useMemo, useState, type CSSProperties } from "react";
import { Link } from "react-router-dom";
import { floorApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import { AgentFloorProposeTopicDialog } from "./ProposeTopicDialog";
import { useAgentFloorShell } from "./agent-floor-shell";
import {
  chartFromPreview,
  clusterMixColorVar,
  clusterMixLabel,
  consensusStatusLabel,
  defaultTopicsPageModel,
  formatDelta01,
  formatPct01,
  inferCategoryValueFromTopicClass,
  parseTopicsPagePayload,
  previewFromBrowseRow,
  type TopicsBrowseRowModel,
  type TopicsPageModel,
  type TopicsQuickFilterChip,
  type TopicsSelectedTopicChartModel,
  type TopicsSelectedTopicPreviewModel,
} from "./agentfloorTopicsModel";

function rowCategoryValue(row: TopicsBrowseRowModel): string {
  return row.categoryValue ?? inferCategoryValueFromTopicClass(row.topicClass);
}

function browseSecondaryLine(row: TopicsBrowseRowModel): string {
  const parts = [
    `Δ ${formatDelta01(row.probabilityDelta)}`,
    row.consensusStatus ? consensusStatusLabel(row.consensusStatus) : null,
    row.deadlineLabel ?? null,
  ].filter(Boolean);
  return parts.join(" · ");
}

function topSignalLine(row: TopicsBrowseRowModel): string | null {
  if (!row.topSignalHint && !row.proofHint) return null;
  const hint = row.topSignalHint ?? "";
  const proof = row.proofHint ? ` · ${row.proofHint}` : "";
  if (!hint) return `Top signal: ${row.proofHint}`.trim();
  return `Top signal: ${hint}${proof}`.trim();
}

function selectedDonutGradient(chart: TopicsSelectedTopicChartModel | undefined): string | undefined {
  const lo = chart?.longPercent ?? 0;
  const sh = chart?.shortPercent ?? 0;
  const t = lo + sh;
  if (t <= 0) return undefined;
  const longDeg = (lo / t) * 360;
  return `conic-gradient(var(--af-tone-b) 0deg ${longDeg}deg, var(--red) ${longDeg}deg 360deg)`;
}

const SAVED_VIEW_IDS = new Set(["Q.01", "Q.04"]);

export default function AgentFloorTopicsPage() {
  const { portalContainer } = useAgentFloorShell();
  const [proposeOpen, setProposeOpen] = useState(false);
  const [model, setModel] = useState<TopicsPageModel>(defaultTopicsPageModel);
  const [category, setCategory] = useState("all");
  const [activeQuick, setActiveQuick] = useState<Set<string>>(() => new Set());
  const [selectedTopicId, setSelectedTopicId] = useState<string | null>(
    () => defaultTopicsPageModel.browseRows[0]?.topicId ?? null,
  );

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

  const quickFilters: TopicsQuickFilterChip[] = model.quickFilters ?? [];

  const filteredBrowseRows = useMemo(() => {
    return model.browseRows.filter((row) => {
      const cat = rowCategoryValue(row);
      if (category !== "all" && cat !== category) return false;
      for (const q of activeQuick) {
        if (q === "watchlist" && !row.watchlisted) return false;
        if (q === "saved_view" && !SAVED_VIEW_IDS.has(row.topicId)) return false;
        if (q === "consensus" || q === "divergent" || q === "low_signal" || q === "speculative") {
          if (row.consensusStatus !== q) return false;
        }
      }
      return true;
    });
  }, [model.browseRows, category, activeQuick]);

  useEffect(() => {
    if (filteredBrowseRows.length === 0) return;
    const stillVisible = filteredBrowseRows.some((r) => r.topicId === selectedTopicId);
    if (!selectedTopicId || !stillVisible) {
      setSelectedTopicId(filteredBrowseRows[0]?.topicId ?? null);
    }
  }, [filteredBrowseRows, selectedTopicId]);

  const selectedRow = useMemo(
    () => filteredBrowseRows.find((r) => r.topicId === selectedTopicId) ?? filteredBrowseRows[0],
    [filteredBrowseRows, selectedTopicId],
  );

  const selectedPreview: TopicsSelectedTopicPreviewModel | undefined = useMemo(() => {
    if (!selectedRow) return undefined;
    if (model.selectedTopic?.topicId === selectedRow.topicId) {
      return model.selectedTopic;
    }
    return previewFromBrowseRow(selectedRow);
  }, [model.selectedTopic, selectedRow]);

  const selectedChart: TopicsSelectedTopicChartModel | undefined = useMemo(() => {
    if (!selectedPreview) return undefined;
    if (
      model.selectedTopicChart &&
      model.selectedTopic?.topicId === selectedPreview.topicId
    ) {
      return model.selectedTopicChart;
    }
    return chartFromPreview(selectedPreview, "donut");
  }, [model.selectedTopic?.topicId, model.selectedTopicChart, selectedPreview]);

  const rail = model.rightRail;
  const cluster = rail?.clusterActivity ?? [];
  const mixTotal = cluster.reduce((s, i) => s + i.count, 0);
  const mixGradient =
    cluster.length > 0
      ? (() => {
          let acc = 0;
          const stops: string[] = [];
          for (const { cluster: cl, count } of cluster) {
            const deg = mixTotal > 0 ? (count / mixTotal) * 360 : 0;
            const start = acc;
            acc += deg;
            stops.push(`${clusterMixColorVar(cl)} ${start}deg ${acc}deg`);
          }
          return `conic-gradient(${stops.join(", ")})`;
        })()
      : undefined;

  const donutGradient = selectedDonutGradient(selectedChart);
  const lower = model.lowerAnalytics;

  const toggleQuick = useCallback((value: string) => {
    setActiveQuick((prev) => {
      const next = new Set(prev);
      if (next.has(value)) next.delete(value);
      else next.add(value);
      return next;
    });
  }, []);

  const h = model.header;

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

        <div className="af-topics-catbar" role="group" aria-label="Category selection">
          <div className="af-topics-catbar-chips">
            {model.categories.map((c) => (
              <button
                key={c.value}
                type="button"
                className={cn("af-topics-chip", category === c.value && "af-topics-chip--on")}
                onClick={() => setCategory(c.value)}
              >
                {c.label}
              </button>
            ))}
          </div>
          <span className="af-topics-live-pill" title="Browse freshness">
            <span className="af-topics-live-dot" aria-hidden />
            Updated live
          </span>
        </div>

        {quickFilters.length > 0 ? (
          <div className="af-topics-quick" role="group" aria-label="Quick filters">
            {quickFilters.map((f) => (
              <button
                key={f.value}
                type="button"
                className={cn("af-topics-chip", activeQuick.has(f.value) && "af-topics-chip--on")}
                onClick={() => toggleQuick(f.value)}
              >
                {f.label}
              </button>
            ))}
          </div>
        ) : null}

        <div className="af-topics-layout af-topics-layout--browse">
          <div className="af-topics-main">
            <div className="af-topics-table-wrap" role="region" aria-label="Topic browse table">
              <table className="af-topics-table">
                <thead>
                  <tr>
                    <th scope="col">ID</th>
                    <th scope="col">Topic</th>
                    <th scope="col">Category</th>
                    <th scope="col">Long</th>
                    <th scope="col">Short</th>
                    <th scope="col">Agents</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredBrowseRows.map((row) => {
                    const isSel = row.topicId === selectedRow?.topicId;
                    const sig = topSignalLine(row);
                    return (
                      <tr
                        key={row.topicId}
                        className={cn("af-topics-tr", isSel && "af-topics-tr--sel")}
                        onClick={() => setSelectedTopicId(row.topicId)}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" || e.key === " ") {
                            e.preventDefault();
                            setSelectedTopicId(row.topicId);
                          }
                        }}
                        tabIndex={0}
                        aria-selected={isSel}
                      >
                        <td className="af-topics-td af-topics-td--mono">{row.topicId}</td>
                        <td className="af-topics-td af-topics-td--topic">
                          <span className="af-topics-td-title">{row.title}</span>
                          <span className="af-topics-td-sub">{browseSecondaryLine(row)}</span>
                          {sig ? <span className="af-topics-td-sig">{sig}</span> : null}
                          <Link
                            className="af-topics-row-cta"
                            to={row.openTopicDetailsUrl}
                            onClick={(e) => e.stopPropagation()}
                          >
                            View topic details
                          </Link>
                        </td>
                        <td className="af-topics-td af-topics-td--muted">{row.topicClass}</td>
                        <td className="af-topics-td af-topics-td--mono">{formatPct01(row.probabilityLong)}</td>
                        <td className="af-topics-td af-topics-td--mono">{formatPct01(row.probabilityShort)}</td>
                        <td className="af-topics-td af-topics-td--mono">
                          {row.agentCount != null ? row.agentCount.toLocaleString() : "—"}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              {filteredBrowseRows.length === 0 ? (
                <p className="af-topics-empty">No topics match the current filters.</p>
              ) : null}
            </div>
          </div>

          <aside className="af-topics-rail af-topics-rail--panel" aria-label="Selected topic">
            {selectedPreview ? (
              <>
                <section className="af-topics-panel-card">
                  <p className="af-topics-panel-kicker">
                    {selectedPreview.topicId} · {selectedPreview.topicClass}
                  </p>
                  <h2 className="af-topics-panel-title">{selectedPreview.title}</h2>
                  <ul className="af-topics-panel-meta">
                    <li>
                      <span className="af-topics-panel-lbl">Status</span>{" "}
                      {consensusStatusLabel(selectedPreview.consensusStatus)}
                    </li>
                    <li>
                      <span className="af-topics-panel-lbl">Delta</span>{" "}
                      {formatDelta01(selectedPreview.probabilityDelta)}
                    </li>
                    <li>
                      <span className="af-topics-panel-lbl">Long</span>{" "}
                      {formatPct01(selectedPreview.probabilityLong)} ·{" "}
                      <span className="af-topics-panel-lbl">Short</span>{" "}
                      {formatPct01(selectedPreview.probabilityShort)}
                    </li>
                  </ul>

                  <div className="af-topics-donut-block">
                    <p className="af-topics-donut-cap">Selected topic split</p>
                    <div className="af-topics-donut-wrap">
                      <div className="af-topics-donut-chart">
                        <div
                          className={cn(
                            "af-topics-donut-ring",
                            selectedChart?.kind === "pie" && "af-topics-donut-ring--pie",
                          )}
                          style={
                            donutGradient
                              ? ({ background: donutGradient } as CSSProperties)
                              : undefined
                          }
                        />
                        <div className="af-topics-donut-center" aria-hidden>
                          <span className="af-topics-donut-pct">
                            {formatPct01(selectedPreview.probabilityLong)}
                          </span>
                          <span className="af-topics-donut-sub">long</span>
                        </div>
                      </div>
                      <p className="af-topics-donut-legend">
                        Long {formatPct01(selectedPreview.probabilityLong)} · Short{" "}
                        {formatPct01(selectedPreview.probabilityShort)}
                      </p>
                    </div>
                  </div>

                  <div className="af-topics-panel-block">
                    <h3 className="af-topics-panel-h">Participation context</h3>
                    <ul className="af-topics-panel-list">
                      <li>
                        Speculative participation:{" "}
                        {(() => {
                          const s = selectedPreview.participationContext?.speculativeParticipationShare;
                          if (s == null || !Number.isFinite(s)) return "—";
                          if (s < 0.08) return "low";
                          return `${Math.round(s * 100)}%`;
                        })()}
                      </li>
                      <li>
                        Neutral-cluster visible:{" "}
                        {(() => {
                          const n = selectedPreview.participationContext?.neutralClusterShare;
                          if (n == null || !Number.isFinite(n)) return "—";
                          return n > 0.05 ? "yes" : "low";
                        })()}
                      </li>
                      <li>
                        Unclustered visible:{" "}
                        {(() => {
                          const u = selectedPreview.participationContext?.unclusteredShare;
                          if (u == null || !Number.isFinite(u)) return "—";
                          return u > 0.02 ? "yes" : "low";
                        })()}
                      </li>
                    </ul>
                  </div>

                  <div className="af-topics-panel-block">
                    <h3 className="af-topics-panel-h">Top signal previews</h3>
                    <p className="af-topics-sig-prev">
                      <span className="af-topics-sig-lbl">Long</span>{" "}
                      {selectedPreview.topLongPreview?.agentName ?? "—"}
                      {selectedPreview.topLongPreview?.proofLabel
                        ? ` · ${selectedPreview.topLongPreview.proofLabel}`
                        : ""}
                    </p>
                    <p className="af-topics-sig-prev">
                      <span className="af-topics-sig-lbl">Short</span>{" "}
                      {selectedPreview.topShortPreview?.agentName ?? "—"}
                      {selectedPreview.topShortPreview?.proofLabel
                        ? ` · ${selectedPreview.topShortPreview.proofLabel}`
                        : ""}
                    </p>
                  </div>

                  <div className="af-topics-panel-actions">
                    <Link className="af-topics-action-solid" to={selectedPreview.openTopicDetailsUrl}>
                      View topic details
                    </Link>
                    {selectedPreview.openResearchUrl ? (
                      <Link className="af-topics-action-ghost" to={selectedPreview.openResearchUrl}>
                        Open in Research
                      </Link>
                    ) : null}
                  </div>
                </section>

                {rail?.dailyDigestTakeaway ? (
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

                {cluster.length > 0 ? (
                  <section className="topics-cluster-card">
                    <h2 className="topics-cluster-hdr">Cluster Activity</h2>
                    <figure
                      className="topics-cluster-fig"
                      aria-label={`Cluster Activity: ${cluster.map((i) => `${clusterMixLabel(i.cluster)} ${i.count}`).join(", ")} of ${mixTotal} agents`}
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
                        {cluster.map((i) => (
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

                {rail?.regionalDivergence ? (
                  <section className="af-topics-rail-card">
                    <h2 className="af-topics-rail-h">GEO DIVERGENCE</h2>
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
              </>
            ) : null}
          </aside>
        </div>

        {lower?.regionalContextMap || lower?.regionalAccuracy?.length ? (
          <section className="af-topics-lower" aria-label="Lower analytics">
            <div className="af-topics-lower-grid">
              {lower.regionalContextMap ? (
                <div className="af-topics-lower-card">
                  <h2 className="af-topics-lower-h">Regional context map</h2>
                  <p className="af-topics-lower-gated">
                    {lower.regionalContextMap.gatedLabel ?? "Interactive map — Analyst+"}
                  </p>
                  <button type="button" className="af-topics-lower-upgrade">
                    {lower.regionalContextMap.upgradeLabel ?? "Upgrade"}
                  </button>
                </div>
              ) : null}
              {lower.regionalAccuracy && lower.regionalAccuracy.length > 0 ? (
                <div className="af-topics-lower-card">
                  <h2 className="af-topics-lower-h">Regional accuracy</h2>
                  <ul className="af-topics-acc-list">
                    {lower.regionalAccuracy.map((item) => (
                      <li key={item.region} className="af-topics-acc-li">
                        <span>{item.region}</span>
                        <span className="af-topics-acc-score">{item.score}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              ) : null}
            </div>
          </section>
        ) : null}
      </div>
    </>
  );
}
