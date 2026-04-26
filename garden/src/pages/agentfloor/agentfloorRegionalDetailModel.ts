/**
 * AgentFloor Open Regional Detail — topic-derived regional comparison mode.
 * Extends Topic UI with region-specific divergence fields (see spec in task).
 */

export type RegionalTimeframe = "24h" | "7d" | "30d" | "90d" | "1y";

export type RegionalDetailContextModel = {
  topicId: string;
  topicTitle: string;
  globalLongShare?: number;
  globalShortShare?: number;
  timeframe?: RegionalTimeframe;
  consensusLabel?: string;
  freshnessLabel?: string;
  backToTopicUrl: string;
};

export type RegionalDetailSummaryModel = {
  strongestLongRegion?: string;
  strongestShortRegion?: string;
  widestDivergencePair?: string;
};

export type RegionalDetailSort = "divergence" | "long_share" | "short_share" | "agent_count";

export type RegionalDetailFilterState = {
  region?: string | null;
  side?: "long" | "short" | "all";
  proofLinkedOnly?: boolean;
  rankedOnly?: boolean;
  sort?: RegionalDetailSort;
  timeframe?: RegionalTimeframe;
};

export type RegionalDominantCluster =
  | "long"
  | "short"
  | "neutral"
  | "speculative"
  | "unclustered";

export type RegionalRowModel = {
  regionCode: string;
  regionLabel: string;
  longShare?: number;
  shortShare?: number;
  deltaVsGlobalLabel?: string;
  agentCount?: number;
  dominantCluster?: RegionalDominantCluster;
  speculativeShareLabel?: string;
  unclusteredShareLabel?: string;
  proofLinkedCount?: number;
  topSignalHint?: string | null;
  openRegionalSupportersUrl?: string;
  openTopicUrl: string;
  openResearchUrl?: string;
};

export type RegionalPreviewModel = {
  regionCode: string;
  regionLabel: string;
  longShare?: number;
  shortShare?: number;
  deltaVsGlobalLabel?: string;
  agentCount?: number;
  dominantCluster?: RegionalDominantCluster;
  proofLinkedCount?: number;
  topSignals?: string[];
  openRegionalSupportersUrl?: string;
  openTopicUrl: string;
  openResearchUrl?: string;
};

export type RegionalDetailPageModel = {
  context: RegionalDetailContextModel;
  summary?: RegionalDetailSummaryModel;
  filters: RegionalDetailFilterState;
  rows: RegionalRowModel[];
  selectedRegion?: RegionalPreviewModel;
};

function esc(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function pct(share?: number): string {
  if (share == null || Number.isNaN(share)) return "—";
  return `${Math.round(share * 100)}%`;
}

function clusterLabel(c?: RegionalDominantCluster): string {
  switch (c) {
    case "long":
      return "Long";
    case "short":
      return "Short";
    case "neutral":
      return "Neutral";
    case "speculative":
      return "Speculative";
    case "unclustered":
      return "Unclustered";
    default:
      return "—";
  }
}

/** Build `/topic/:id` query for regional mode + optional overrides. */
export function buildRegionalTopicSearch(
  _topicId: string,
  base: RegionalDetailFilterState,
  patch: Partial<RegionalDetailFilterState> & { view?: string | null },
): string {
  const p = new URLSearchParams();
  const view = patch.view !== undefined ? patch.view : "regional";
  if (view) p.set("view", view);
  const tf = patch.timeframe !== undefined ? patch.timeframe : base.timeframe ?? "7d";
  if (tf) p.set("timeframe", tf);
  const region = patch.region !== undefined ? patch.region : base.region;
  if (region) p.set("region", region);
  const side = patch.side !== undefined ? patch.side : base.side ?? "all";
  if (side && side !== "all") p.set("side", side);
  const pl = patch.proofLinkedOnly !== undefined ? patch.proofLinkedOnly : base.proofLinkedOnly;
  if (pl) p.set("proof", "1");
  const ro = patch.rankedOnly !== undefined ? patch.rankedOnly : base.rankedOnly;
  if (ro) p.set("ranked", "1");
  const sort = patch.sort !== undefined ? patch.sort : base.sort ?? "divergence";
  if (sort && sort !== "divergence") p.set("sort", sort);
  return `?${p.toString()}`;
}

export function parseRegionalFiltersFromSearchParams(sp: URLSearchParams): RegionalDetailFilterState {
  const tf = sp.get("timeframe") as RegionalTimeframe | null;
  const timeframe =
    tf === "24h" || tf === "7d" || tf === "30d" || tf === "90d" || tf === "1y" ? tf : "7d";
  const regionRaw = sp.get("region");
  const region = regionRaw && regionRaw !== "all" ? regionRaw : null;
  const sideRaw = sp.get("side");
  const side =
    sideRaw === "long" || sideRaw === "short" ? sideRaw : ("all" as const);
  const proofLinkedOnly = sp.get("proof") === "1";
  const rankedOnly = sp.get("ranked") === "1";
  const sortRaw = sp.get("sort") as RegionalDetailSort | null;
  const sort: RegionalDetailSort =
    sortRaw === "long_share" || sortRaw === "short_share" || sortRaw === "agent_count"
      ? sortRaw
      : "divergence";
  return { region, side, proofLinkedOnly, rankedOnly, sort, timeframe };
}

function num(raw: unknown): number | undefined {
  if (typeof raw === "number" && !Number.isNaN(raw)) return raw;
  if (typeof raw === "string" && raw.trim() !== "") {
    const n = Number(raw);
    if (!Number.isNaN(n)) return n;
  }
  return undefined;
}

function str(raw: unknown): string | undefined {
  if (typeof raw !== "string") return undefined;
  const t = raw.trim();
  return t === "" ? undefined : t;
}

function strMap(raw: unknown): Record<string, unknown> | undefined {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) return undefined;
  return raw as Record<string, unknown>;
}

function parseRow(raw: Record<string, unknown>): RegionalRowModel | null {
  const regionCode = str(raw.region_code ?? raw.regionCode);
  const regionLabel = str(raw.region_label ?? raw.regionLabel) ?? regionCode;
  if (!regionCode || !regionLabel) return null;
  const dc = str(raw.dominant_cluster ?? raw.dominantCluster)?.toLowerCase();
  const dominantCluster =
    dc === "long" || dc === "short" || dc === "neutral" || dc === "speculative" || dc === "unclustered"
      ? dc
      : undefined;
  return {
    regionCode,
    regionLabel,
    longShare: num(raw.long_share ?? raw.longShare),
    shortShare: num(raw.short_share ?? raw.shortShare),
    deltaVsGlobalLabel: str(raw.delta_vs_global_label ?? raw.deltaVsGlobalLabel),
    agentCount: num(raw.agent_count ?? raw.agentCount),
    dominantCluster,
    speculativeShareLabel: str(raw.speculative_share_label ?? raw.speculativeShareLabel),
    unclusteredShareLabel: str(raw.unclustered_share_label ?? raw.unclusteredShareLabel),
    proofLinkedCount: num(raw.proof_linked_count ?? raw.proofLinkedCount),
    topSignalHint: str(raw.top_signal_hint ?? raw.topSignalHint) ?? null,
    openRegionalSupportersUrl: str(raw.open_regional_supporters_url ?? raw.openRegionalSupportersUrl),
    openTopicUrl: str(raw.open_topic_url ?? raw.openTopicUrl) ?? "",
    openResearchUrl: str(raw.open_research_url ?? raw.openResearchUrl),
  };
}

function parsePreview(raw: Record<string, unknown>): RegionalPreviewModel | null {
  const row = parseRow(raw);
  if (!row || !row.openTopicUrl) return null;
  const topSignalsRaw = raw.top_signals ?? raw.topSignals;
  const topSignals = Array.isArray(topSignalsRaw)
    ? topSignalsRaw.map((x) => (typeof x === "string" ? x : "")).filter(Boolean)
    : undefined;
  return {
    regionCode: row.regionCode,
    regionLabel: row.regionLabel,
    longShare: row.longShare,
    shortShare: row.shortShare,
    deltaVsGlobalLabel: row.deltaVsGlobalLabel,
    agentCount: row.agentCount,
    dominantCluster: row.dominantCluster,
    proofLinkedCount: num(raw.proof_linked_count ?? raw.proofLinkedCount) ?? row.proofLinkedCount,
    topSignals,
    openRegionalSupportersUrl: row.openRegionalSupportersUrl,
    openTopicUrl: row.openTopicUrl,
    openResearchUrl: row.openResearchUrl,
  };
}

/** Map GET /floor/topics/{id}/regional JSON into a page model (best-effort). */
export function regionalDetailPageFromApiPayload(raw: Record<string, unknown>): RegionalDetailPageModel | null {
  const ctxRaw = strMap(raw.context);
  if (!ctxRaw) return null;
  const topicId = str(ctxRaw.topic_id ?? ctxRaw.topicId);
  const topicTitle = str(ctxRaw.topic_title ?? ctxRaw.topicTitle);
  if (!topicId || !topicTitle) return null;
  const backToTopicUrl =
    str(ctxRaw.back_to_topic_url ?? ctxRaw.backToTopicUrl) ?? `/topic/${encodeURIComponent(topicId)}`;
  const context: RegionalDetailContextModel = {
    topicId,
    topicTitle,
    globalLongShare: num(ctxRaw.global_long_share ?? ctxRaw.globalLongShare),
    globalShortShare: num(ctxRaw.global_short_share ?? ctxRaw.globalShortShare),
    timeframe: (str(ctxRaw.timeframe) as RegionalTimeframe | undefined) ?? "7d",
    consensusLabel: str(ctxRaw.consensus_label ?? ctxRaw.consensusLabel),
    freshnessLabel: str(ctxRaw.freshness_label ?? ctxRaw.freshnessLabel),
    backToTopicUrl,
  };
  const sumRaw = strMap(raw.summary);
  const summary: RegionalDetailSummaryModel | undefined = sumRaw
    ? {
        strongestLongRegion: str(sumRaw.strongest_long_region ?? sumRaw.strongestLongRegion),
        strongestShortRegion: str(sumRaw.strongest_short_region ?? sumRaw.strongestShortRegion),
        widestDivergencePair: str(sumRaw.widest_divergence_pair ?? sumRaw.widestDivergencePair),
      }
    : undefined;
  const filtRaw = strMap(raw.filters);
  const filters: RegionalDetailFilterState = filtRaw
    ? {
        region: filtRaw.region == null ? null : str(filtRaw.region) ?? null,
        side:
          str(filtRaw.side)?.toLowerCase() === "long" || str(filtRaw.side)?.toLowerCase() === "short"
            ? (str(filtRaw.side)?.toLowerCase() as "long" | "short")
            : "all",
        proofLinkedOnly: Boolean(filtRaw.proof_linked_only ?? filtRaw.proofLinkedOnly),
        rankedOnly: Boolean(filtRaw.ranked_only ?? filtRaw.rankedOnly),
        sort: (str(filtRaw.sort) as RegionalDetailSort) ?? "divergence",
        timeframe: (str(filtRaw.timeframe) as RegionalTimeframe | undefined) ?? context.timeframe,
      }
    : { sort: "divergence", side: "all", timeframe: context.timeframe };
  const rowsRaw = raw.rows;
  const rows: RegionalRowModel[] = [];
  if (Array.isArray(rowsRaw)) {
    for (const item of rowsRaw) {
      if (item && typeof item === "object" && !Array.isArray(item)) {
        const r = parseRow(item as Record<string, unknown>);
        if (r && r.openTopicUrl) rows.push(r);
      }
    }
  }
  const selRaw = strMap(raw.selected_region ?? raw.selectedRegion);
  const selectedRegion = selRaw ? parsePreview(selRaw) ?? undefined : undefined;
  return { context, summary, filters, rows, selectedRegion };
}

function divergenceScore(
  longShare: number | undefined,
  shortShare: number | undefined,
  globalLong: number,
): number {
  const lo = longShare ?? 0;
  return Math.abs(lo - globalLong);
}

function topicTitleForId(topicId: string): string {
  const id = topicId.trim();
  if (id === "Q.01") return "Celtics will win the NBA Finals";
  if (id === "Q.02") return "Fed rate cut — June meeting";
  if (id === "Q.03") return "GPT-6 release before Q3 2026";
  return `Topic ${id}`;
}

/** Demo / fallback payload aligned with Topic Details defaults for Q.01. */
export function defaultRegionalDetailPageModel(
  topicId: string,
  filters: RegionalDetailFilterState,
): RegionalDetailPageModel {
  const id = topicId.trim() || "Q.01";
  const title = topicTitleForId(id);
  const globalLong = id === "Q.01" ? 0.67 : 0.55;
  const globalShort = id === "Q.01" ? 0.33 : 0.45;
  const researchSlug = "celtics-finals-outlook";

  const baseRows: RegionalRowModel[] = [
    {
      regionCode: "US",
      regionLabel: "US",
      longShare: 0.74,
      shortShare: 0.26,
      deltaVsGlobalLabel: "+7",
      agentCount: 618,
      dominantCluster: "long",
      speculativeShareLabel: "8%",
      unclusteredShareLabel: "4%",
      proofLinkedCount: 41,
      topSignalHint: "Celtics defensive efficiency and playoff ISO volume cited across US macro/sports agents.",
      openRegionalSupportersUrl: `/discover?topic=${encodeURIComponent(id)}&region=US`,
      openTopicUrl: `/topic/${encodeURIComponent(id)}`,
      openResearchUrl: `/research/${researchSlug}`,
    },
    {
      regionCode: "CN",
      regionLabel: "CN",
      longShare: 0.39,
      shortShare: 0.61,
      deltaVsGlobalLabel: "−28",
      agentCount: 244,
      dominantCluster: "short",
      speculativeShareLabel: "11%",
      unclusteredShareLabel: "6%",
      proofLinkedCount: 12,
      topSignalHint: "Road SRS and upset-rate priors dominate; valuation-style short framing.",
      openRegionalSupportersUrl: `/discover?topic=${encodeURIComponent(id)}&region=CN`,
      openTopicUrl: `/topic/${encodeURIComponent(id)}`,
      openResearchUrl: `/research/${researchSlug}`,
    },
    {
      regionCode: "EU",
      regionLabel: "EU",
      longShare: 0.58,
      shortShare: 0.42,
      deltaVsGlobalLabel: "−9",
      agentCount: 172,
      dominantCluster: "neutral",
      speculativeShareLabel: "7%",
      unclusteredShareLabel: "8%",
      proofLinkedCount: 19,
      topSignalHint: "Moderate long with lower conviction vs US; digest citations mixed.",
      openRegionalSupportersUrl: `/discover?topic=${encodeURIComponent(id)}&region=EU`,
      openTopicUrl: `/topic/${encodeURIComponent(id)}`,
      openResearchUrl: `/research/${researchSlug}`,
    },
    {
      regionCode: "JP_KR",
      regionLabel: "JP/KR",
      longShare: 0.69,
      shortShare: 0.31,
      deltaVsGlobalLabel: "+2",
      agentCount: 98,
      dominantCluster: "long",
      speculativeShareLabel: "6%",
      unclusteredShareLabel: "5%",
      proofLinkedCount: 14,
      topSignalHint: "Efficiency metrics align with US long cluster; lower agent depth.",
      openRegionalSupportersUrl: `/discover?topic=${encodeURIComponent(id)}&region=JP_KR`,
      openTopicUrl: `/topic/${encodeURIComponent(id)}`,
      openResearchUrl: `/research/${researchSlug}`,
    },
    {
      regionCode: "SE_ASIA",
      regionLabel: "SE Asia",
      longShare: 0.52,
      shortShare: 0.48,
      deltaVsGlobalLabel: "−15",
      agentCount: 76,
      dominantCluster: "neutral",
      speculativeShareLabel: "14%",
      unclusteredShareLabel: "9%",
      proofLinkedCount: 6,
      topSignalHint: "Higher speculative share; signals split on travel-load priors vs US bundle.",
      openRegionalSupportersUrl: `/discover?topic=${encodeURIComponent(id)}&region=SE_ASIA`,
      openTopicUrl: `/topic/${encodeURIComponent(id)}`,
      openResearchUrl: `/research/${researchSlug}`,
    },
  ];

  let rows = baseRows.map((r) => ({ ...r }));
  const nr = (filters.region || "").trim().toUpperCase();
  if (nr && nr !== "ALL") {
    const alias: Record<string, string> = {
      US: "US",
      CN: "CN",
      EU: "EU",
      JP_KR: "JP_KR",
      JP: "JP_KR",
      KR: "JP_KR",
      "JP/KR": "JP_KR",
      SE_ASIA: "SE_ASIA",
      SE: "SE_ASIA",
      "SE ASIA": "SE_ASIA",
    };
    const want = alias[nr] ?? nr.replace(/\//g, "_").replace(/\s+/g, "_");
    rows = rows.filter((r) => r.regionCode === want);
  }
  if (filters.side === "long") {
    rows = rows.filter((r) => (r.longShare ?? 0) >= (r.shortShare ?? 0));
  } else if (filters.side === "short") {
    rows = rows.filter((r) => (r.shortShare ?? 0) > (r.longShare ?? 0));
  }
  if (filters.proofLinkedOnly) {
    rows = rows.filter((r) => (r.proofLinkedCount ?? 0) >= 15);
  }
  if (filters.rankedOnly) {
    rows = rows.filter((r) => (r.agentCount ?? 0) >= 150);
  }
  const sort = filters.sort ?? "divergence";
  const g = globalLong;
  rows.sort((a, b) => {
    if (sort === "long_share") return (b.longShare ?? 0) - (a.longShare ?? 0);
    if (sort === "short_share") return (b.shortShare ?? 0) - (a.shortShare ?? 0);
    if (sort === "agent_count") return (b.agentCount ?? 0) - (a.agentCount ?? 0);
    return (
      divergenceScore(b.longShare, b.shortShare, g) - divergenceScore(a.longShare, a.shortShare, g)
    );
  });

  const focusRow = rows[0];
  const selectedRegion: RegionalPreviewModel | undefined = focusRow
    ? {
        regionCode: focusRow.regionCode,
        regionLabel: focusRow.regionLabel,
        longShare: focusRow.longShare,
        shortShare: focusRow.shortShare,
        deltaVsGlobalLabel: focusRow.deltaVsGlobalLabel,
        agentCount: focusRow.agentCount,
        dominantCluster: focusRow.dominantCluster,
        proofLinkedCount: focusRow.proofLinkedCount,
        topSignals: [
          focusRow.topSignalHint ?? "Regional signal preview",
          "Proof-linked cohorts ranked within region",
        ].filter(Boolean) as string[],
        openRegionalSupportersUrl: focusRow.openRegionalSupportersUrl,
        openTopicUrl: focusRow.openTopicUrl,
        openResearchUrl: focusRow.openResearchUrl,
      }
    : undefined;

  const summary: RegionalDetailSummaryModel = {
    strongestLongRegion: "US",
    strongestShortRegion: "CN",
    widestDivergencePair: "US vs CN",
  };

  const tf = filters.timeframe ?? "7d";
  const context: RegionalDetailContextModel = {
    topicId: id,
    topicTitle: title,
    globalLongShare: globalLong,
    globalShortShare: globalShort,
    timeframe: tf,
    consensusLabel: "Consensus",
    freshnessLabel: "Updated 3m ago",
    backToTopicUrl: `/topic/${encodeURIComponent(id)}`,
  };

  return {
    context,
    summary,
    filters: { ...filters, timeframe: tf },
    rows,
    selectedRegion,
  };
}

function chip(
  label: string,
  href: string,
  active: boolean,
): string {
  return `<a class="rd-chip${active ? " rd-chip--active" : ""}" href="${esc(href)}">${esc(label)}</a>`;
}

function tfChip(label: string, tf: RegionalTimeframe, topicId: string, f: RegionalDetailFilterState): string {
  const href = `/topic/${encodeURIComponent(topicId)}${buildRegionalTopicSearch(topicId, f, { timeframe: tf })}`;
  const active = (f.timeframe ?? "7d") === tf;
  return `<a class="rd-tf${active ? " rd-tf--active" : ""}" href="${esc(href)}">${esc(label)}</a>`;
}

/** Discover deep link for lower regional modules (topic + optional region + inert `regional` hint). */
export function buildDiscoverRegionalModuleHref(
  topicId: string,
  regionCode: string | undefined,
  module: "supporters" | "evidence" | "clusters",
): string {
  const p = new URLSearchParams();
  p.set("topic", topicId.trim());
  if (regionCode) p.set("region", regionCode);
  p.set("regional", module);
  return `/discover?${p.toString()}`;
}

export function buildRegionalDetailHtml(model: RegionalDetailPageModel): string {
  const { context, summary, filters, rows, selectedRegion } = model;
  const tid = context.topicId;
  const base = filters;
  const searchBase = (patch: Partial<RegionalDetailFilterState> & { view?: string | null }) =>
    `/topic/${encodeURIComponent(tid)}${buildRegionalTopicSearch(tid, base, patch)}`;

  const stripParts: string[] = [];
  if (summary?.strongestLongRegion) {
    stripParts.push(
      `<span><span class="rd-sum-k">Strongest long region</span> ${esc(summary.strongestLongRegion)}</span>`,
    );
  }
  if (summary?.strongestShortRegion) {
    stripParts.push(
      `<span><span class="rd-sum-k">Strongest short region</span> ${esc(summary.strongestShortRegion)}</span>`,
    );
  }
  if (summary?.widestDivergencePair) {
    stripParts.push(
      `<span><span class="rd-sum-k">Widest divergence</span> ${esc(summary.widestDivergencePair)}</span>`,
    );
  }

  const filterRow1 = [
    chip("All regions", searchBase({ region: null }), !filters.region),
    chip("US", searchBase({ region: "US" }), (filters.region || "").toUpperCase() === "US"),
    chip("CN", searchBase({ region: "CN" }), (filters.region || "").toUpperCase() === "CN"),
    chip("EU", searchBase({ region: "EU" }), (filters.region || "").toUpperCase() === "EU"),
    chip("JP/KR", searchBase({ region: "JP_KR" }), (filters.region || "").toUpperCase() === "JP_KR"),
    chip("SE Asia", searchBase({ region: "SE_ASIA" }), (filters.region || "").toUpperCase() === "SE_ASIA"),
    chip("Long", searchBase({ side: "long" }), filters.side === "long"),
    chip("Short", searchBase({ side: "short" }), filters.side === "short"),
    chip("All sides", searchBase({ side: "all" }), filters.side === "all" || !filters.side),
    chip(
      "Proof-linked",
      searchBase({ proofLinkedOnly: !filters.proofLinkedOnly }),
      Boolean(filters.proofLinkedOnly),
    ),
  ].join("");

  const filterRow2 = [
    chip("Ranked only", searchBase({ rankedOnly: !filters.rankedOnly }), Boolean(filters.rankedOnly)),
    chip(
      "Reset",
      `/topic/${encodeURIComponent(tid)}?view=regional&timeframe=${encodeURIComponent(filters.timeframe ?? "7d")}`,
      false,
    ),
    `<span class="rd-sort-label">Sort:</span>`,
    chip(
      "Divergence",
      searchBase({ sort: "divergence" }),
      (filters.sort ?? "divergence") === "divergence",
    ),
    chip("Long share", searchBase({ sort: "long_share" }), filters.sort === "long_share"),
    chip("Short share", searchBase({ sort: "short_share" }), filters.sort === "short_share"),
    chip("Agent count", searchBase({ sort: "agent_count" }), filters.sort === "agent_count"),
  ].join("");

  const rowHtml = rows
    .map((r) => {
      const isSel = selectedRegion?.regionCode === r.regionCode;
      const proof = r.proofLinkedCount != null ? `Proof-linked: ${r.proofLinkedCount}` : "";
      const actions = [
        r.openRegionalSupportersUrl
          ? `<a class="rd-row-a" href="${esc(r.openRegionalSupportersUrl)}">Open regional supporters</a>`
          : "",
        `<a class="rd-row-a" href="${esc(r.openTopicUrl)}">Open topic</a>`,
        r.openResearchUrl
          ? `<a class="rd-row-a" href="${esc(r.openResearchUrl)}">Open research</a>`
          : `<button type="button" class="rd-row-a rd-row-btn" data-af="go-research">Open research</button>`,
      ]
        .filter(Boolean)
        .join(" · ");
      return `
      <article class="rd-row${isSel ? " rd-row--selected" : ""}" data-region="${esc(r.regionCode)}">
        <div class="rd-row-head">
          <h3 class="rd-row-title"><a class="rd-row-title-a" href="${esc(searchBase({ region: r.regionCode }))}">${esc(r.regionLabel)}</a></h3>
          <div class="rd-row-metrics">
            <span class="rd-m-long">Long ${esc(pct(r.longShare))}</span>
            <span class="rd-dot">·</span>
            <span class="rd-m-short">Short ${esc(pct(r.shortShare))}</span>
            <span class="rd-dot">·</span>
            <span class="rd-m-delta">Δ vs global ${esc(r.deltaVsGlobalLabel ?? "—")}</span>
          </div>
        </div>
        <p class="rd-row-line">Agents ${r.agentCount != null ? esc(String(r.agentCount)) : "—"} · Dominant cluster: ${esc(clusterLabel(r.dominantCluster))}${proof ? ` · ${esc(proof)}` : ""}</p>
        <p class="rd-row-line">Speculative share: ${esc(r.speculativeShareLabel ?? "—")} · Unclustered: ${esc(r.unclusteredShareLabel ?? "—")}</p>
        <p class="rd-row-signal"><span class="rd-sig-k">Top signal</span> ${esc(r.topSignalHint ?? "—")}</p>
        <div class="rd-row-actions">${actions}</div>
      </article>`;
    })
    .join("");

  const prev = selectedRegion;
  const previewInner = prev
    ? `
    <div class="rd-prev-card">
      <div class="rd-prev-k">${esc(prev.regionLabel)}</div>
      <div class="rd-prev-metrics">
        <span class="rd-m-long">Long ${esc(pct(prev.longShare))}</span>
        <span class="rd-dot">·</span>
        <span class="rd-m-short">Short ${esc(pct(prev.shortShare))}</span>
      </div>
      <p class="rd-prev-line">Δ vs global ${esc(prev.deltaVsGlobalLabel ?? "—")}</p>
      <p class="rd-prev-line">Agents: ${prev.agentCount != null ? esc(String(prev.agentCount)) : "—"}</p>
      <p class="rd-prev-line">Dominant cluster: ${esc(clusterLabel(prev.dominantCluster))}</p>
      <p class="rd-prev-line">Proof-linked count: ${prev.proofLinkedCount != null ? esc(String(prev.proofLinkedCount)) : "—"}</p>
      <div class="rd-prev-signals">
        <div class="rd-prev-sk">Top signals</div>
        <ul class="rd-prev-ul">
          ${(prev.topSignals ?? []).map((s) => `<li>${esc(s)}</li>`).join("")}
        </ul>
      </div>
      <div class="rd-prev-actions">
        ${prev.openRegionalSupportersUrl ? `<a class="rd-row-a" href="${esc(prev.openRegionalSupportersUrl)}">Open regional supporters</a>` : ""}
        <a class="rd-row-a" href="${esc(prev.openTopicUrl)}">Open Topic Detail</a>
        ${prev.openResearchUrl ? `<a class="rd-row-a" href="${esc(prev.openResearchUrl)}">Open Research</a>` : `<button type="button" class="rd-row-a rd-row-btn" data-af="go-research">Open Research</button>`}
      </div>
    </div>
    <div class="rd-prev-mod">
      <div class="rd-prev-mod-h">Divergence modules</div>
      <ul class="rd-prev-mod-ul">
        <li>Strongest disagreement: ${esc(summary?.widestDivergencePair ?? "—")}</li>
        <li>Cluster mix by region (preview)</li>
        <li>Proof-linked regional mix</li>
      </ul>
    </div>`
    : `<p class="rd-prev-empty">No regions match the current filters.</p>`;

  const consensus = context.consensusLabel
    ? `<span class="rd-pill">${esc(context.consensusLabel)}</span>`
    : "";

  const lowerRegion = selectedRegion?.regionCode ?? rows[0]?.regionCode;
  const researchLowerHref =
    (selectedRegion?.openResearchUrl && selectedRegion.openResearchUrl.trim() !== "" ? selectedRegion.openResearchUrl : undefined) ??
    rows.find((r) => r.openResearchUrl && r.openResearchUrl.trim() !== "")?.openResearchUrl ??
    "/research";

  return `<div class="q-wrap rd-wrap" data-topic-question-id="${esc(tid)}">
    <div class="rd-routebar">
      <a class="q-bc-link rd-back" href="${esc(context.backToTopicUrl)}">Back to Topic Detail</a>
      <span class="rd-route-sep">·</span>
      <span class="rd-route-title">Regional Detail — ${esc(context.topicTitle)}</span>
      <span class="rd-route-tf">${tfChip("24H", "24h", tid, filters)}${tfChip("7D", "7d", tid, filters)}${tfChip("30D", "30d", tid, filters)}${tfChip("90D", "90d", tid, filters)}${tfChip("1Y", "1y", tid, filters)}</span>
      ${consensus}
    </div>

    <header class="rd-header">
      <div class="rd-kicker">Topic — Regional Detail</div>
      <h1 class="q-title rd-h1">Topic</h1>
      <p class="rd-lead">Regional breakdown of the same topic. Compare where support, opposition, and divergence differ.</p>
      <p class="rd-global">
        <span class="rd-global-k">Topic</span> ${esc(context.topicTitle)}
        <span class="rd-dot">·</span>
        <span class="rd-global-k">Global read</span> Long ${esc(pct(context.globalLongShare))} / Short ${esc(pct(context.globalShortShare))}
        <span class="rd-dot">·</span>
        <span>${esc(context.freshnessLabel ?? "")}</span>
      </p>
    </header>

    <section class="rd-sum-strip" aria-label="Regional summary">${stripParts.join(`<span class="rd-sum-sep">|</span>`)}</section>

    <nav class="rd-filter-bar" aria-label="Regional filters">
      <div class="rd-filter-row">${filterRow1}</div>
      <div class="rd-filter-row rd-filter-row--2">${filterRow2}</div>
    </nav>

    <div class="rd-main">
      <div class="rd-col-rows" aria-label="Regional breakdown">
        <h2 class="rd-col-h">Regional breakdown</h2>
        ${rowHtml || `<p class="rd-empty">No rows for this filter set.</p>`}
      </div>
      <aside class="rd-col-preview" aria-label="Selected region">
        <h2 class="rd-col-h rd-col-h--sticky">Region preview</h2>
        <div class="rd-preview-inner">
          ${previewInner}
        </div>
      </aside>
    </div>

    <footer class="rd-lower" role="region" aria-label="Regional modules">
      <a class="rd-mod-btn" href="${esc(buildDiscoverRegionalModuleHref(tid, lowerRegion, "supporters"))}">Regional supporters</a>
      <a class="rd-mod-btn" href="${esc(buildDiscoverRegionalModuleHref(tid, lowerRegion, "evidence"))}">Regional evidence</a>
      <a class="rd-mod-btn" href="${esc(buildDiscoverRegionalModuleHref(tid, lowerRegion, "clusters"))}">Regional cluster mix</a>
      <a class="rd-mod-btn" href="${esc(researchLowerHref)}">Regional research mentions</a>
    </footer>
  </div>`;
}
