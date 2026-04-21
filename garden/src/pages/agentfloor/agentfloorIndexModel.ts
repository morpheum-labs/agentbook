/**
 * View model for AgentFloor Index — discover / trust / watchlist one-pager.
 * Wire to `GET /api/v1/floor/index`; {@link defaultIndexPageModel} is demo fallback.
 */

export type IndexDirectoryType =
  | "macro"
  | "hidden_data"
  | "vq_native"
  | "real_time"
  | "ssi_type"
  | "regional_divergence";

export type IndexAccessTier = "free" | "premium" | "api" | "executable";

export type IndexSummaryChip = {
  label: string;
  value: string;
};

export type IndexFilterChip = {
  label: string;
  value: string;
  active?: boolean;
};

export type IndexDirectoryRow = {
  indexId: string;
  title: string;
  type: IndexDirectoryType;
  signalLabel: string;
  confidenceLabel?: string;
  accessTier: IndexAccessTier;
  openDetailUrl: string;
  canWatchlist: boolean;
  watchlistLocked?: boolean;
  /** For “My watchlist” filter */
  watchlisted?: boolean;
};

export type SourceProvenanceModel = {
  totalSources?: number;
  breakdownLabel?: string;
};

export type TrustSnapshotModel = {
  confidenceScore?: number;
  freshnessLabel?: string;
  lastHumanReviewLabel?: string;
  disagreementLabel?: string;
  methodologyReviewedLabel?: string;
  /** e.g. trigger count today */
  triggersToday?: number;
};

export type UpdateLogItem = {
  timestampLabel: string;
  text: string;
};

export type SelectedIndexPanelModel = {
  indexId: string;
  title: string;
  subtitle?: string;
  whyItMatters?: string;
  currentReading?: string;
  openDetailUrl: string;
  canWatchlist: boolean;
  watchlistLocked?: boolean;
  trustSnapshot?: TrustSnapshotModel;
  sourceProvenance?: SourceProvenanceModel;
  updateLog?: UpdateLogItem[];
};

export type IndexPageModel = {
  header: {
    title: string;
    subtitle: string;
    /** e.g. “My watchlist — Analytic / Terminal” */
    watchlistTierHint?: string;
  };
  summaryChips?: IndexSummaryChip[];
  filters?: IndexFilterChip[];
  rows: IndexDirectoryRow[];
  /** Initial / server-selected index */
  selectedIndex?: SelectedIndexPanelModel;
  /** Rich panel copy keyed by index id (optional; rows still drive directory) */
  panelsById?: Record<string, SelectedIndexPanelModel>;
  lowerStrip?: {
    rebalanceSoonLabel?: string;
    latestResearchLabel?: string;
    openResearchUrl?: string;
  };
};

function isRecord(v: unknown): v is Record<string, unknown> {
  return v != null && typeof v === "object" && !Array.isArray(v);
}

function str(v: unknown): string | undefined {
  return typeof v === "string" ? v : undefined;
}

function num(v: unknown): number | undefined {
  return typeof v === "number" && Number.isFinite(v) ? v : undefined;
}

const INDEX_TYPES = new Set<string>([
  "macro",
  "hidden_data",
  "vq_native",
  "real_time",
  "ssi_type",
  "regional_divergence",
]);

const ACCESS_TIERS = new Set<string>(["free", "premium", "api", "executable"]);

function parseIndexType(v: unknown): IndexDirectoryType | undefined {
  const s = typeof v === "string" ? v.toLowerCase().replace(/-/g, "_") : "";
  return INDEX_TYPES.has(s) ? (s as IndexDirectoryType) : undefined;
}

function parseAccessTier(v: unknown): IndexAccessTier | undefined {
  const s = typeof v === "string" ? v.toLowerCase() : "";
  return ACCESS_TIERS.has(s) ? (s as IndexAccessTier) : undefined;
}

function parseSummaryChip(raw: unknown): IndexSummaryChip | null {
  if (!isRecord(raw)) return null;
  const label = str(raw.label);
  const value = str(raw.value);
  if (!label || !value) return null;
  return { label, value };
}

function parseFilterChip(raw: unknown): IndexFilterChip | null {
  if (!isRecord(raw)) return null;
  const label = str(raw.label);
  const value = str(raw.value);
  if (!label || !value) return null;
  return { label, value, active: Boolean(raw.active) };
}

function parseDirectoryRow(raw: unknown): IndexDirectoryRow | null {
  if (!isRecord(raw)) return null;
  const indexId = str(raw.index_id) ?? str(raw.indexId);
  const title = str(raw.title);
  const type = parseIndexType(raw.type ?? raw.index_type ?? raw.indexType);
  const signalLabel = str(raw.signal_label) ?? str(raw.signalLabel);
  const accessTier = parseAccessTier(raw.access_tier ?? raw.accessTier);
  const openDetailUrl = str(raw.open_detail_url) ?? str(raw.openDetailUrl) ?? "";
  if (!indexId || !title || !type || !signalLabel || !accessTier || !openDetailUrl) return null;
  return {
    indexId,
    title,
    type,
    signalLabel,
    confidenceLabel: str(raw.confidence_label) ?? str(raw.confidenceLabel),
    accessTier,
    openDetailUrl,
    canWatchlist: Boolean(raw.can_watchlist ?? raw.canWatchlist ?? true),
    watchlistLocked: Boolean(raw.watchlist_locked ?? raw.watchlistLocked),
    watchlisted: Boolean(raw.watchlisted ?? raw.on_watchlist),
  };
}

function parseTrustSnapshot(raw: unknown): TrustSnapshotModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    confidenceScore: num(raw.confidence_score) ?? num(raw.confidenceScore),
    freshnessLabel: str(raw.freshness_label) ?? str(raw.freshnessLabel),
    lastHumanReviewLabel: str(raw.last_human_review_label) ?? str(raw.lastHumanReviewLabel),
    disagreementLabel: str(raw.disagreement_label) ?? str(raw.disagreementLabel),
    methodologyReviewedLabel:
      str(raw.methodology_reviewed_label) ?? str(raw.methodologyReviewedLabel),
    triggersToday: num(raw.triggers_today) ?? num(raw.triggersToday),
  };
}

function parseSourceProvenance(raw: unknown): SourceProvenanceModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    totalSources: num(raw.total_sources) ?? num(raw.totalSources),
    breakdownLabel: str(raw.breakdown_label) ?? str(raw.breakdownLabel),
  };
}

function parseUpdateLogItem(raw: unknown): UpdateLogItem | null {
  if (!isRecord(raw)) return null;
  const timestampLabel = str(raw.timestamp_label) ?? str(raw.timestampLabel);
  const text = str(raw.text);
  if (!timestampLabel || !text) return null;
  return { timestampLabel, text };
}

function parseSelectedPanel(raw: unknown): SelectedIndexPanelModel | null {
  if (!isRecord(raw)) return null;
  const indexId = str(raw.index_id) ?? str(raw.indexId);
  const title = str(raw.title);
  const openDetailUrl = str(raw.open_detail_url) ?? str(raw.openDetailUrl) ?? "";
  if (!indexId || !title || !openDetailUrl) return null;
  const ulRaw = raw.update_log ?? raw.updateLog;
  let updateLog: UpdateLogItem[] | undefined;
  if (Array.isArray(ulRaw)) {
    const ul = ulRaw.map(parseUpdateLogItem).filter((x): x is UpdateLogItem => x != null);
    if (ul.length > 0) updateLog = ul;
  }
  return {
    indexId,
    title,
    subtitle: str(raw.subtitle),
    whyItMatters: str(raw.why_it_matters) ?? str(raw.whyItMatters),
    currentReading: str(raw.current_reading) ?? str(raw.currentReading),
    openDetailUrl,
    canWatchlist: Boolean(raw.can_watchlist ?? raw.canWatchlist ?? true),
    watchlistLocked: Boolean(raw.watchlist_locked ?? raw.watchlistLocked),
    trustSnapshot: parseTrustSnapshot(raw.trust_snapshot ?? raw.trustSnapshot),
    sourceProvenance: parseSourceProvenance(raw.source_provenance ?? raw.sourceProvenance),
    updateLog,
  };
}

function parsePanelsMap(raw: unknown): Record<string, SelectedIndexPanelModel> | undefined {
  if (!isRecord(raw)) return undefined;
  const out: Record<string, SelectedIndexPanelModel> = {};
  for (const [k, v] of Object.entries(raw)) {
    const p = parseSelectedPanel(v);
    if (p) out[k] = p;
  }
  return Object.keys(out).length ? out : undefined;
}

/** Maps snake_case / camelCase API payloads into {@link IndexPageModel}. */
export function parseIndexPagePayload(raw: unknown): IndexPageModel | null {
  if (!isRecord(raw)) return null;
  const headerRaw = raw.header;
  if (!isRecord(headerRaw)) return null;
  const title = str(headerRaw.title);
  const subtitle = str(headerRaw.subtitle);
  if (!title || !subtitle) return null;

  const rowsArr = raw.rows;
  if (!Array.isArray(rowsArr)) return null;
  const rows = rowsArr.map(parseDirectoryRow).filter((r): r is IndexDirectoryRow => r != null);
  if (rows.length === 0) return null;

  const scRaw = raw.summary_chips ?? raw.summaryChips;
  let summaryChips: IndexSummaryChip[] | undefined;
  if (Array.isArray(scRaw)) {
    const sc = scRaw.map(parseSummaryChip).filter((c): c is IndexSummaryChip => c != null);
    if (sc.length > 0) summaryChips = sc;
  }

  const fRaw = raw.filters;
  let filters: IndexFilterChip[] | undefined;
  if (Array.isArray(fRaw)) {
    const f = fRaw.map(parseFilterChip).filter((c): c is IndexFilterChip => c != null);
    if (f.length > 0) filters = f;
  }

  const siRaw = raw.selected_index ?? raw.selectedIndex;
  const selectedIndex = siRaw ? parseSelectedPanel(siRaw) ?? undefined : undefined;

  const panelsRaw = raw.index_panels ?? raw.indexPanels;
  const panelsById = parsePanelsMap(panelsRaw);

  const lsRaw = raw.lower_strip ?? raw.lowerStrip;
  let lowerStrip: IndexPageModel["lowerStrip"] | undefined;
  if (isRecord(lsRaw)) {
    lowerStrip = {
      rebalanceSoonLabel: str(lsRaw.rebalance_soon_label) ?? str(lsRaw.rebalanceSoonLabel),
      latestResearchLabel: str(lsRaw.latest_research_label) ?? str(lsRaw.latestResearchLabel),
      openResearchUrl: str(lsRaw.open_research_url) ?? str(lsRaw.openResearchUrl),
    };
  }

  return {
    header: {
      title,
      subtitle,
      watchlistTierHint:
        str(headerRaw.watchlist_tier_hint) ?? str(headerRaw.watchlistTierHint),
    },
    summaryChips,
    filters,
    rows,
    selectedIndex,
    panelsById,
    lowerStrip,
  };
}

export function indexTypeLabel(t: IndexDirectoryType): string {
  switch (t) {
    case "macro":
      return "Macro";
    case "hidden_data":
      return "Hidden Data";
    case "vq_native":
      return "VQ-Native";
    case "real_time":
      return "Real-Time";
    case "ssi_type":
      return "SSI-Type";
    case "regional_divergence":
      return "Regional";
    default:
      return t;
  }
}

export function indexAccessTierLabel(tier: IndexAccessTier): string {
  switch (tier) {
    case "free":
      return "Free";
    case "premium":
      return "Premium";
    case "api":
      return "API";
    case "executable":
      return "Exec";
    default:
      return tier;
  }
}

/** Fallback panel when server did not send `index_panels[id]`. */
export function minimalPanelFromRow(row: IndexDirectoryRow): SelectedIndexPanelModel {
  const conf = row.confidenceLabel?.replace(/^Confidence\s*/i, "").trim();
  const score = conf && /^\d+$/.test(conf) ? Number(conf) : undefined;
  return {
    indexId: row.indexId,
    title: row.title,
    subtitle: indexTypeLabel(row.type),
    whyItMatters: "Open the full index detail for drivers, methodology, and historical context.",
    currentReading: row.signalLabel,
    openDetailUrl: row.openDetailUrl,
    canWatchlist: row.canWatchlist,
    watchlistLocked: row.watchlistLocked,
    trustSnapshot: {
      confidenceScore: score,
      freshnessLabel: "Freshness on demand",
      lastHumanReviewLabel: "—",
      disagreementLabel: "—",
      methodologyReviewedLabel: "See detail",
    },
    sourceProvenance: { totalSources: undefined, breakdownLabel: "Provenance available in detail view." },
    updateLog: [],
  };
}

export function resolveSelectedIndexPanel(
  model: IndexPageModel,
  indexId: string,
  row: IndexDirectoryRow | undefined,
): SelectedIndexPanelModel | undefined {
  const fromMap = model.panelsById?.[indexId];
  const baseRow = row ?? model.rows.find((r) => r.indexId === indexId);
  if (fromMap && baseRow) {
    return {
      ...fromMap,
      indexId: fromMap.indexId || indexId,
      openDetailUrl: baseRow.openDetailUrl || fromMap.openDetailUrl,
      canWatchlist: baseRow.canWatchlist,
      watchlistLocked: baseRow.watchlistLocked,
    };
  }
  if (fromMap) return fromMap;
  if (baseRow) return minimalPanelFromRow(baseRow);
  return undefined;
}

function demoPanels(): Record<string, SelectedIndexPanelModel> {
  return {
    "I.00": panelI("I.00", "Global Liquidity Pulse", "macro", "Broad risk-on / risk-off pressure gauge."),
    "I.01": panelI("I.01", "Retail Parking Lot Index", "vq_native", "Leads retail earnings by weeks."),
    "I.02": panelI("I.02", "China Crematorium Activity Index", "hidden_data", "Non-traditional macro stress signal."),
    "I.03": panelI("I.03", "Truck Traffic Index", "real_time", "Freight pulse for goods demand."),
    "I.04": panelI("I.04", "MAG7-style Basket", "ssi_type", "Concentration + rebalance risk in one lens."),
  };
}

function panelI(
  indexId: string,
  title: string,
  type: IndexDirectoryType,
  why: string,
): SelectedIndexPanelModel {
  return {
    indexId,
    title,
    subtitle: indexTypeLabel(type),
    whyItMatters: why,
    currentReading:
      indexId === "I.00"
        ? "Neutral"
        : indexId === "I.02"
          ? "High alert"
          : indexId === "I.03"
            ? "Softening WoW"
            : indexId === "I.04"
              ? "Bullish drift MTD"
              : "Bullish divergence",
    openDetailUrl: `/index?focus=${encodeURIComponent(indexId)}`,
    canWatchlist: true,
    watchlistLocked: true,
    trustSnapshot: {
      confidenceScore:
        indexId === "I.00" ? 62 : indexId === "I.02" ? 84 : indexId === "I.01" ? 76 : indexId === "I.03" ? 71 : 68,
      freshnessLabel: "Updated 5m ago",
      lastHumanReviewLabel: "Apr 20",
      disagreementLabel: indexId === "I.02" ? "Low" : "Moderate",
      methodologyReviewedLabel: "Reviewed",
      triggersToday: indexId === "I.01" ? 2 : 0,
    },
    sourceProvenance: {
      totalSources: 12,
      breakdownLabel: "Official 4 · Market 3 · VQ 2 · News 2 · Agent 1",
    },
    updateLog: [
      { timestampLabel: "03:10", text: "Coverage expanded" },
      { timestampLabel: "02:42", text: "Volatility rose" },
    ],
  };
}

/** Demo fallback aligned with `floorComposedIndexPage` (Go). */
export const defaultIndexPageModel: IndexPageModel = {
  header: {
    title: "Index",
    subtitle: "Discover proprietary indices, trust the signal, and follow what matters now.",
    watchlistTierHint: "My watchlist — Analytic / Terminal",
  },
  summaryChips: [
    { label: "Top mover", value: "Retail Parking +12%" },
    { label: "Highest confidence", value: "China Crematorium 84" },
    { label: "Rebalance soon", value: "MAG7-style · 3d" },
    { label: "Updated", value: "5m" },
  ],
  filters: [
    { label: "All", value: "all", active: true },
    { label: "Macro", value: "macro" },
    { label: "Hidden Data", value: "hidden_data" },
    { label: "VQ-Native", value: "vq_native" },
    { label: "SSI-Type", value: "ssi_type" },
    { label: "Free", value: "free" },
    { label: "Premium", value: "premium" },
    { label: "API", value: "api" },
    { label: "Executable", value: "executable" },
    { label: "My watchlist", value: "my_watchlist" },
  ],
  rows: [
    {
      indexId: "I.01",
      title: "Retail Parking Lot Index",
      type: "vq_native",
      signalLabel: "+12% / 7d",
      confidenceLabel: "Confidence 76",
      accessTier: "premium",
      openDetailUrl: "/index?focus=I.01",
      canWatchlist: true,
      watchlistLocked: true,
      watchlisted: true,
    },
    {
      indexId: "I.02",
      title: "China Crematorium Activity Index",
      type: "hidden_data",
      signalLabel: "High alert",
      confidenceLabel: "Confidence 84",
      accessTier: "premium",
      openDetailUrl: "/index?focus=I.02",
      canWatchlist: true,
      watchlistLocked: true,
    },
    {
      indexId: "I.03",
      title: "Truck Traffic Index",
      type: "real_time",
      signalLabel: "-3% WoW",
      confidenceLabel: "Confidence 71",
      accessTier: "api",
      openDetailUrl: "/index?focus=I.03",
      canWatchlist: true,
      watchlistLocked: true,
    },
    {
      indexId: "I.04",
      title: "MAG7-style Basket",
      type: "ssi_type",
      signalLabel: "+6% MTD",
      confidenceLabel: "Confidence 68",
      accessTier: "executable",
      openDetailUrl: "/index?focus=I.04",
      canWatchlist: true,
      watchlistLocked: true,
    },
    {
      indexId: "I.00",
      title: "Global Liquidity Pulse",
      type: "macro",
      signalLabel: "Neutral",
      confidenceLabel: "Confidence 62",
      accessTier: "free",
      openDetailUrl: "/index?focus=I.00",
      canWatchlist: true,
      watchlistLocked: true,
    },
  ],
  selectedIndex: {
    indexId: "I.01",
    title: "Retail Parking Lot Index",
    subtitle: "VQ-Native",
    whyItMatters: "Leads retail earnings by weeks.",
    currentReading: "Bullish divergence",
    openDetailUrl: "/index?focus=I.01",
    canWatchlist: true,
    watchlistLocked: true,
    trustSnapshot: {
      confidenceScore: 82,
      freshnessLabel: "Updated 5m ago",
      lastHumanReviewLabel: "Apr 20",
      disagreementLabel: "Moderate",
      methodologyReviewedLabel: "Reviewed",
      triggersToday: 2,
    },
    sourceProvenance: {
      totalSources: 12,
      breakdownLabel: "Official 4 · Market 3 · VQ 2 · News 2 · Agent 1",
    },
    updateLog: [
      { timestampLabel: "03:10", text: "Coverage expanded" },
      { timestampLabel: "02:42", text: "Volatility rose" },
    ],
  },
  panelsById: demoPanels(),
  lowerStrip: {
    rebalanceSoonLabel: "MAG7-style Basket · 3d",
    latestResearchLabel: "Hidden indicators this week",
    openResearchUrl: "/research",
  },
};
