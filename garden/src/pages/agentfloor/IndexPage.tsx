import { useCallback, useEffect, useMemo, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Lock, Star } from "lucide-react";
import { floorApi } from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  defaultIndexPageModel,
  indexAccessTierLabel,
  indexTypeLabel,
  parseIndexPagePayload,
  resolveSelectedIndexPanel,
  type IndexDirectoryRow,
  type IndexPageModel,
} from "./agentfloorIndexModel";

function rowMatchesFilter(row: IndexDirectoryRow, filter: string): boolean {
  if (filter === "all") return true;
  if (filter === "my_watchlist") return Boolean(row.watchlisted);
  if (filter === "macro" || filter === "hidden_data" || filter === "vq_native" || filter === "ssi_type") {
    return row.type === filter;
  }
  if (filter === "free" || filter === "premium" || filter === "api" || filter === "executable") {
    return row.accessTier === filter;
  }
  return true;
}

export default function AgentFloorIndexPage() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [model, setModel] = useState<IndexPageModel>(defaultIndexPageModel);
  const [filter, setFilter] = useState(() => {
    const q = searchParams.get("filter");
    return q && q !== "" ? q : "all";
  });
  const [selectedId, setSelectedId] = useState<string>(() => {
    const focus = searchParams.get("focus");
    if (focus && defaultIndexPageModel.rows.some((r) => r.indexId === focus)) return focus;
    return defaultIndexPageModel.rows[0]?.indexId ?? "";
  });

  useEffect(() => {
    let cancelled = false;
    void floorApi
      .getIndexPage()
      .then((raw: Record<string, unknown>) => {
        if (cancelled) return;
        const parsed = parseIndexPagePayload(raw);
        if (parsed) setModel(parsed);
      })
      .catch(() => {
        /* demo fallback */
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const focusFromUrl = searchParams.get("focus");
  useEffect(() => {
    if (!focusFromUrl) return;
    const hit = model.rows.some((r) => r.indexId === focusFromUrl);
    if (hit) setSelectedId(focusFromUrl);
  }, [focusFromUrl, model.rows]);

  const filteredRows = useMemo(
    () => model.rows.filter((r) => rowMatchesFilter(r, filter)),
    [model.rows, filter],
  );

  useEffect(() => {
    if (filteredRows.length === 0) return;
    if (!filteredRows.some((r) => r.indexId === selectedId)) {
      setSelectedId(filteredRows[0].indexId);
    }
  }, [filteredRows, selectedId]);

  const selectedRow = useMemo(
    () => filteredRows.find((r) => r.indexId === selectedId) ?? filteredRows[0],
    [filteredRows, selectedId],
  );

  const selectedPanel = useMemo(() => {
    if (!selectedRow) return undefined;
    return resolveSelectedIndexPanel(model, selectedRow.indexId, selectedRow);
  }, [model, selectedRow]);

  const filters = model.filters ?? [];

  const onSelectRow = useCallback(
    (id: string) => {
      setSelectedId(id);
      const next = new URLSearchParams(searchParams);
      next.set("focus", id);
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const onFilter = useCallback(
    (value: string) => {
      setFilter(value);
      const next = new URLSearchParams(searchParams);
      if (value === "all") next.delete("filter");
      else next.set("filter", value);
      setSearchParams(next, { replace: true });
    },
    [searchParams, setSearchParams],
  );

  const h = model.header;
  const chips = model.summaryChips ?? [];
  const lower = model.lowerStrip;

  return (
    <div className="af-index">
      <header className="af-index-pagehead">
        <div className="af-index-pagehead-text">
          <h1 className="af-index-h1">{h.title}</h1>
          <p className="af-index-sub">{h.subtitle}</p>
        </div>
        <div className="af-index-pagehead-actions">
          <Link to="/subscribe" className="btn-paid af-index-subscribe">
            Subscribe
          </Link>
          {h.watchlistTierHint ? (
            <span className="af-index-tier-hint" title="Watchlist requires Analytic or Terminal">
              {h.watchlistTierHint}
            </span>
          ) : null}
        </div>
      </header>

      {chips.length > 0 ? (
        <div className="af-index-summary" role="region" aria-label="Index summary">
          {chips.map((c) => (
            <div key={c.label} className="af-index-summary-chip">
              <span className="af-index-summary-lbl">{c.label}</span>
              <span className="af-index-summary-val">{c.value}</span>
            </div>
          ))}
        </div>
      ) : null}

      <div className="af-index-filterbar" role="group" aria-label="Index filters">
        {filters.map((f) => (
          <button
            key={f.value}
            type="button"
            className={cn("af-index-chip", filter === f.value && "af-index-chip--on")}
            onClick={() => onFilter(f.value)}
          >
            {f.label}
          </button>
        ))}
      </div>

      <div className="af-index-layout">
        <div className="af-index-main">
          <div className="af-index-table-wrap" role="region" aria-label="Index directory">
            <table className="af-index-table">
              <thead>
                <tr>
                  <th scope="col">ID</th>
                  <th scope="col">Index</th>
                  <th scope="col">Type</th>
                  <th scope="col">Signal</th>
                  <th scope="col">Access</th>
                </tr>
              </thead>
              <tbody>
                {filteredRows.map((row) => {
                  const isSel = row.indexId === selectedRow?.indexId;
                  const wlLocked = row.canWatchlist && row.watchlistLocked;
                  const wlDisabled = !row.canWatchlist;
                  return (
                    <tr
                      key={row.indexId}
                      className={cn("af-index-tr", isSel && "af-index-tr--sel")}
                      onClick={() => onSelectRow(row.indexId)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" || e.key === " ") {
                          e.preventDefault();
                          onSelectRow(row.indexId);
                        }
                      }}
                      tabIndex={0}
                      aria-selected={isSel}
                    >
                      <td className="af-index-td af-index-td--mono">{row.indexId}</td>
                      <td className="af-index-td af-index-td--index">
                        <span className="af-index-td-title">{row.title}</span>
                        {row.confidenceLabel ? (
                          <span className="af-index-td-sub">{row.confidenceLabel}</span>
                        ) : null}
                        <div className="af-index-row-actions">
                          <Link
                            className="af-index-row-cta"
                            to={row.openDetailUrl}
                            onClick={(e) => e.stopPropagation()}
                          >
                            View detail
                          </Link>
                          <button
                            type="button"
                            className={cn(
                              "af-index-watch-btn",
                              wlLocked && "af-index-watch-btn--locked",
                              wlDisabled && "af-index-watch-btn--disabled",
                            )}
                            disabled={wlDisabled}
                            title={
                              wlDisabled
                                ? "Watchlist not available for this index"
                                : wlLocked
                                  ? "Upgrade to Analytic or Terminal to add to watchlist"
                                  : "Add to watchlist"
                            }
                            aria-disabled={wlDisabled || wlLocked}
                            onClick={(e) => {
                              e.stopPropagation();
                            }}
                          >
                            {wlLocked || wlDisabled ? (
                              <Lock className="af-index-watch-ico" aria-hidden />
                            ) : (
                              <Star className="af-index-watch-ico" aria-hidden />
                            )}
                            Watchlist
                          </button>
                        </div>
                      </td>
                      <td className="af-index-td af-index-td--muted">{indexTypeLabel(row.type)}</td>
                      <td className="af-index-td af-index-td--mono">{row.signalLabel}</td>
                      <td className="af-index-td af-index-td--tag">
                        <span className={cn("af-index-tier-tag", `af-index-tier-tag--${row.accessTier}`)}>
                          {indexAccessTierLabel(row.accessTier)}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            {filteredRows.length === 0 ? (
              <p className="af-index-empty">No indices match the current filter.</p>
            ) : null}
          </div>
          <p className="af-index-row-rules">
            Row click updates the panel. Watchlist is available on Analytic / Terminal; free tier sees it locked.
          </p>
        </div>

        <aside className="af-index-panel" aria-label="Selected index">
          {selectedPanel && selectedRow ? (
            <>
              <div className="af-index-panel-head">
                <p className="af-index-panel-kicker">
                  {selectedPanel.indexId}
                  {selectedPanel.subtitle ? ` · ${selectedPanel.subtitle}` : ""}
                </p>
                <h2 className="af-index-panel-title">{selectedPanel.title}</h2>
              </div>

              {selectedPanel.whyItMatters ? (
                <section className="af-index-panel-block">
                  <h3 className="af-index-panel-h">Why it matters</h3>
                  <p className="af-index-panel-p">{selectedPanel.whyItMatters}</p>
                </section>
              ) : null}

              {selectedPanel.currentReading ? (
                <section className="af-index-panel-block">
                  <h3 className="af-index-panel-h">Current reading</h3>
                  <p className="af-index-panel-reading">{selectedPanel.currentReading}</p>
                </section>
              ) : null}

              <section className="af-index-panel-block">
                <h3 className="af-index-panel-h">Index health</h3>
                <ul className="af-index-panel-list">
                  <li>
                    <span className="af-index-panel-lbl">Confidence</span>{" "}
                    {selectedPanel.trustSnapshot?.confidenceScore != null
                      ? `${selectedPanel.trustSnapshot.confidenceScore} / 100`
                      : "—"}
                  </li>
                  <li>
                    <span className="af-index-panel-lbl">Freshness</span>{" "}
                    {selectedPanel.trustSnapshot?.freshnessLabel ?? "—"}
                  </li>
                  <li>
                    <span className="af-index-panel-lbl">Triggers today</span>{" "}
                    {selectedPanel.trustSnapshot?.triggersToday != null
                      ? String(selectedPanel.trustSnapshot.triggersToday)
                      : "—"}
                  </li>
                </ul>
              </section>

              <section className="af-index-panel-block">
                <h3 className="af-index-panel-h">Source provenance</h3>
                <p className="af-index-panel-p">
                  {selectedPanel.sourceProvenance?.totalSources != null
                    ? `${selectedPanel.sourceProvenance.totalSources} total`
                    : "—"}
                </p>
                {selectedPanel.sourceProvenance?.breakdownLabel ? (
                  <p className="af-index-panel-muted">{selectedPanel.sourceProvenance.breakdownLabel}</p>
                ) : null}
              </section>

              {selectedPanel.updateLog && selectedPanel.updateLog.length > 0 ? (
                <section className="af-index-panel-block">
                  <h3 className="af-index-panel-h">Live update log</h3>
                  <ul className="af-index-log">
                    {selectedPanel.updateLog.map((u, i) => (
                      <li key={`${u.timestampLabel}-${i}`}>
                        <span className="af-index-log-ts">{u.timestampLabel}</span> {u.text}
                      </li>
                    ))}
                  </ul>
                </section>
              ) : null}

              <section className="af-index-panel-block">
                <h3 className="af-index-panel-h">Verification / trust</h3>
                <ul className="af-index-panel-list">
                  <li>
                    <span className="af-index-panel-lbl">Human review</span>{" "}
                    {selectedPanel.trustSnapshot?.lastHumanReviewLabel ?? "—"}
                  </li>
                  <li>
                    <span className="af-index-panel-lbl">Agent disagreement</span>{" "}
                    {selectedPanel.trustSnapshot?.disagreementLabel ?? "—"}
                  </li>
                  <li>
                    <span className="af-index-panel-lbl">Methodology</span>{" "}
                    {selectedPanel.trustSnapshot?.methodologyReviewedLabel ?? "—"}
                  </li>
                </ul>
              </section>

              {selectedRow.accessTier === "executable" ? (
                <p className="af-index-wallet-hint">
                  Executable indices may require a connected wallet for on-chain actions. Browsing stays
                  off-chain.
                </p>
              ) : null}

              <div className="af-index-panel-actions">
                <Link to="/subscribe" className="af-index-panel-btn af-index-panel-btn--secondary">
                  Unlock full methodology
                </Link>
                <button
                  type="button"
                  className={cn(
                    "af-index-panel-btn",
                    selectedPanel.watchlistLocked && "af-index-panel-btn--locked",
                    !selectedPanel.canWatchlist && "af-index-panel-btn--disabled",
                  )}
                  disabled={!selectedPanel.canWatchlist || Boolean(selectedPanel.watchlistLocked)}
                  title={
                    !selectedPanel.canWatchlist
                      ? "Not eligible for watchlist"
                      : selectedPanel.watchlistLocked
                        ? "Analytic or Terminal required"
                        : "Add to watchlist"
                  }
                >
                  {selectedPanel.watchlistLocked || !selectedPanel.canWatchlist ? (
                    <Lock className="af-index-watch-ico" aria-hidden />
                  ) : (
                    <Star className="af-index-watch-ico" aria-hidden />
                  )}
                  Add to watchlist
                </button>
                <Link to={selectedPanel.openDetailUrl} className="af-index-panel-btn af-index-panel-btn--ghost">
                  View detail
                </Link>
              </div>
            </>
          ) : (
            <p className="af-index-empty">Select an index from the directory.</p>
          )}
        </aside>
      </div>

      {lower && (lower.rebalanceSoonLabel || lower.latestResearchLabel) ? (
        <footer className="af-index-lower">
          {lower.rebalanceSoonLabel ? (
            <span className="af-index-lower-bit">{lower.rebalanceSoonLabel}</span>
          ) : null}
          {lower.latestResearchLabel ? (
            <span className="af-index-lower-bit">{lower.latestResearchLabel}</span>
          ) : null}
          {lower.openResearchUrl ? (
            <Link to={lower.openResearchUrl} className="af-index-lower-link">
              Open Research
            </Link>
          ) : null}
        </footer>
      ) : null}
    </div>
  );
}
