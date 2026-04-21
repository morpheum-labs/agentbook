/**
 * View model for AgentFloor Topics — merged browse surface (structured rows + selected-topic panel).
 * Wire to `GET /api/v1/floor/topics`; {@link defaultTopicsPageModel} is demo fallback.
 */

export type TopicsConsensusStatus = "consensus" | "divergent" | "low_signal" | "speculative";

export type TopicsHeaderModel = {
  title: string;
  subtitle: string;
  terminalOnlyActionLabel?: string;
};

export type TopicsCategoryChip = {
  label: string;
  value: string;
  active?: boolean;
};

export type TopicsQuickFilterChip = {
  label: string;
  value: string;
  active?: boolean;
};

/** Optional `categoryValue` filters category chips (e.g. sports, macro). */
export type TopicsBrowseRowModel = {
  topicId: string;
  title: string;
  topicClass: string;
  categoryValue?: string;
  probabilityLong?: number;
  probabilityShort?: number;
  probabilityDelta?: number;
  consensusStatus?: TopicsConsensusStatus;
  deadlineLabel?: string;
  agentCount?: number;
  topSignalHint?: string | null;
  proofHint?: string | null;
  openTopicDetailsUrl: string;
  /** UI-only: filter Watchlist chip */
  watchlisted?: boolean;
};

export type TopicsSelectedTopicPreviewModel = {
  topicId: string;
  title: string;
  topicClass: string;
  probabilityLong?: number;
  probabilityShort?: number;
  probabilityDelta?: number;
  consensusStatus?: TopicsConsensusStatus;
  participationContext?: {
    speculativeParticipationShare?: number;
    neutralClusterShare?: number;
    unclusteredShare?: number;
  };
  topLongPreview?: {
    agentName: string;
    proofLabel?: string | null;
  };
  topShortPreview?: {
    agentName: string;
    proofLabel?: string | null;
  };
  openTopicDetailsUrl: string;
  openResearchUrl?: string;
};

export type TopicsSelectedTopicChartModel = {
  kind: "pie" | "donut";
  longPercent?: number;
  shortPercent?: number;
};

export type TopicsDigestTakeawayModel = {
  title?: string;
  subtitle?: string;
  note?: string;
};

export type TopicsClusterKind = "long" | "short" | "neutral" | "speculative" | "unclustered";

export type TopicsClusterActivityItem = {
  cluster: TopicsClusterKind;
  count: number;
};

export type TopicsRegionalDivergenceModel = {
  summary: string;
  openRegionalDetailUrl?: string;
};

export type TopicsRegionalContextMapModel = {
  gatedLabel?: string;
  upgradeLabel?: string;
};

export type TopicsRegionalAccuracyItem = {
  region: string;
  score: number;
};

export type TopicsPageModel = {
  header: TopicsHeaderModel;
  categories: TopicsCategoryChip[];
  quickFilters?: TopicsQuickFilterChip[];
  browseRows: TopicsBrowseRowModel[];
  selectedTopic?: TopicsSelectedTopicPreviewModel;
  selectedTopicChart?: TopicsSelectedTopicChartModel;
  rightRail?: {
    dailyDigestTakeaway?: TopicsDigestTakeawayModel;
    clusterActivity?: TopicsClusterActivityItem[];
    regionalDivergence?: TopicsRegionalDivergenceModel;
  };
  lowerAnalytics?: {
    regionalContextMap?: TopicsRegionalContextMapModel;
    regionalAccuracy?: TopicsRegionalAccuracyItem[];
  };
};

const CLUSTER_SET = new Set<string>(["long", "short", "neutral", "speculative", "unclustered"]);

function isRecord(v: unknown): v is Record<string, unknown> {
  return v != null && typeof v === "object" && !Array.isArray(v);
}

function str(v: unknown): string | undefined {
  return typeof v === "string" ? v : undefined;
}

function num(v: unknown): number | undefined {
  return typeof v === "number" && Number.isFinite(v) ? v : undefined;
}

function parseConsensusStatus(v: unknown): TopicsConsensusStatus | undefined {
  const s = typeof v === "string" ? v.toLowerCase().replace(/-/g, "_") : "";
  if (s === "consensus" || s === "divergent" || s === "low_signal" || s === "speculative") {
    return s as TopicsConsensusStatus;
  }
  return undefined;
}

function parseClusterActivityItem(raw: unknown): TopicsClusterActivityItem | null {
  if (!isRecord(raw)) return null;
  const cluster = typeof raw.cluster === "string" ? raw.cluster.toLowerCase() : "";
  if (!CLUSTER_SET.has(cluster)) return null;
  const count = num(raw.count);
  if (count == null) return null;
  return { cluster: cluster as TopicsClusterKind, count };
}

function parseBrowseRow(raw: unknown): TopicsBrowseRowModel | null {
  if (!isRecord(raw)) return null;
  const topicId = str(raw.topic_id) ?? str(raw.topicId);
  const title = str(raw.title);
  const topicClass = str(raw.topic_class) ?? str(raw.topicClass);
  const openTopicDetailsUrl =
    str(raw.open_topic_details_url) ?? str(raw.openTopicDetailsUrl) ?? "";
  if (!topicId || !title || !topicClass || !openTopicDetailsUrl) return null;
  return {
    topicId,
    title,
    topicClass,
    categoryValue: str(raw.category) ?? str(raw.category_value) ?? str(raw.categoryValue),
    probabilityLong: num(raw.probability_long) ?? num(raw.probabilityLong),
    probabilityShort: num(raw.probability_short) ?? num(raw.probabilityShort),
    probabilityDelta: num(raw.probability_delta) ?? num(raw.probabilityDelta),
    consensusStatus: parseConsensusStatus(raw.consensus_status ?? raw.consensusStatus),
    deadlineLabel: str(raw.deadline_label) ?? str(raw.deadlineLabel),
    agentCount: num(raw.agent_count) ?? num(raw.agentCount),
    topSignalHint: str(raw.top_signal_hint) ?? str(raw.topSignalHint) ?? null,
    proofHint: str(raw.proof_hint) ?? str(raw.proofHint) ?? null,
    openTopicDetailsUrl,
    watchlisted: Boolean(raw.watchlisted ?? raw.on_watchlist),
  };
}

function parseParticipationContext(
  raw: unknown,
): TopicsSelectedTopicPreviewModel["participationContext"] | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    speculativeParticipationShare:
      num(raw.speculative_participation_share) ?? num(raw.speculativeParticipationShare),
    neutralClusterShare: num(raw.neutral_cluster_share) ?? num(raw.neutralClusterShare),
    unclusteredShare: num(raw.unclustered_share) ?? num(raw.unclusteredShare),
  };
}

function parseSignalPreview(
  raw: unknown,
): TopicsSelectedTopicPreviewModel["topLongPreview"] | undefined {
  if (!isRecord(raw)) return undefined;
  const agentName = str(raw.agent_name) ?? str(raw.agentName);
  if (!agentName) return undefined;
  return {
    agentName,
    proofLabel: str(raw.proof_label) ?? str(raw.proofLabel) ?? null,
  };
}

function parseSelectedTopic(raw: unknown): TopicsSelectedTopicPreviewModel | null {
  if (!isRecord(raw)) return null;
  const topicId = str(raw.topic_id) ?? str(raw.topicId);
  const title = str(raw.title);
  const topicClass = str(raw.topic_class) ?? str(raw.topicClass);
  const openTopicDetailsUrl =
    str(raw.open_topic_details_url) ?? str(raw.openTopicDetailsUrl) ?? "";
  if (!topicId || !title || !topicClass || !openTopicDetailsUrl) return null;
  const pcRaw = raw.participation_context ?? raw.participationContext;
  return {
    topicId,
    title,
    topicClass,
    probabilityLong: num(raw.probability_long) ?? num(raw.probabilityLong),
    probabilityShort: num(raw.probability_short) ?? num(raw.probabilityShort),
    probabilityDelta: num(raw.probability_delta) ?? num(raw.probabilityDelta),
    consensusStatus: parseConsensusStatus(raw.consensus_status ?? raw.consensusStatus),
    participationContext: parseParticipationContext(pcRaw),
    topLongPreview: parseSignalPreview(raw.top_long_preview ?? raw.topLongPreview),
    topShortPreview: parseSignalPreview(raw.top_short_preview ?? raw.topShortPreview),
    openTopicDetailsUrl,
    openResearchUrl: str(raw.open_research_url) ?? str(raw.openResearchUrl),
  };
}

function parseSelectedTopicChart(raw: unknown): TopicsSelectedTopicChartModel | undefined {
  if (!isRecord(raw)) return undefined;
  const kind = str(raw.kind);
  if (kind !== "pie" && kind !== "donut") return undefined;
  return {
    kind,
    longPercent: num(raw.long_percent) ?? num(raw.longPercent),
    shortPercent: num(raw.short_percent) ?? num(raw.shortPercent),
  };
}

function parseCategoryChip(raw: unknown): TopicsCategoryChip | null {
  if (!isRecord(raw)) return null;
  const label = str(raw.label);
  const value = str(raw.value);
  if (!label || !value) return null;
  return { label, value, active: Boolean(raw.active) };
}

function parseQuickFilterChip(raw: unknown): TopicsQuickFilterChip | null {
  if (!isRecord(raw)) return null;
  const label = str(raw.label);
  const value = str(raw.value);
  if (!label || !value) return null;
  return { label, value, active: Boolean(raw.active) };
}

/** Maps snake_case / camelCase API payloads into {@link TopicsPageModel}. */
export function parseTopicsPagePayload(raw: unknown): TopicsPageModel | null {
  if (!isRecord(raw)) return null;
  const headerRaw = raw.header;
  if (!isRecord(headerRaw)) return null;
  const title = str(headerRaw.title);
  const subtitle = str(headerRaw.subtitle);
  if (!title || !subtitle) return null;

  const catArr = raw.categories;
  if (!Array.isArray(catArr)) return null;
  const categories = catArr.map(parseCategoryChip).filter((c): c is TopicsCategoryChip => c != null);
  if (categories.length === 0) return null;

  const browseArr = raw.browse_rows ?? raw.browseRows;
  if (!Array.isArray(browseArr)) return null;
  const browseRows = browseArr.map(parseBrowseRow).filter((r): r is TopicsBrowseRowModel => r != null);
  if (browseRows.length === 0) return null;

  const qfRaw = raw.quick_filters ?? raw.quickFilters;
  let quickFilters: TopicsQuickFilterChip[] | undefined;
  if (Array.isArray(qfRaw)) {
    const qf = qfRaw.map(parseQuickFilterChip).filter((c): c is TopicsQuickFilterChip => c != null);
    if (qf.length > 0) quickFilters = qf;
  }

  const stRaw = raw.selected_topic ?? raw.selectedTopic;
  const selectedTopic = stRaw ? parseSelectedTopic(stRaw) : undefined;

  const chartRaw = raw.selected_topic_chart ?? raw.selectedTopicChart;
  const selectedTopicChart = chartRaw ? parseSelectedTopicChart(chartRaw) : undefined;

  const rrRaw = raw.right_rail ?? raw.rightRail;
  let rightRail: TopicsPageModel["rightRail"] | undefined;
  if (isRecord(rrRaw)) {
    rightRail = {};
    const digestRaw = rrRaw.daily_digest_takeaway ?? rrRaw.dailyDigestTakeaway;
    if (isRecord(digestRaw)) {
      rightRail.dailyDigestTakeaway = {
        title: str(digestRaw.title),
        subtitle: str(digestRaw.subtitle),
        note: str(digestRaw.note),
      };
    }
    const caRaw = rrRaw.cluster_activity ?? rrRaw.clusterActivity ?? rrRaw.inferred_cluster_mix ?? rrRaw.inferredClusterMix;
    if (Array.isArray(caRaw)) {
      const clusterActivity = caRaw.map(parseClusterActivityItem).filter((x): x is TopicsClusterActivityItem => x != null);
      if (clusterActivity.length > 0) rightRail.clusterActivity = clusterActivity;
    }
    const regRaw = rrRaw.regional_divergence ?? rrRaw.regionalDivergence;
    if (isRecord(regRaw) && typeof regRaw.summary === "string") {
      rightRail.regionalDivergence = {
        summary: regRaw.summary,
        openRegionalDetailUrl:
          str(regRaw.open_regional_detail_url) ?? str(regRaw.openRegionalDetailUrl),
      };
    }
  }

  const lowerRaw = raw.lower_analytics ?? raw.lowerAnalytics;
  let lowerAnalytics: TopicsPageModel["lowerAnalytics"] | undefined;
  if (isRecord(lowerRaw)) {
    lowerAnalytics = {};
    const mapRaw = lowerRaw.regional_context_map ?? lowerRaw.regionalContextMap;
    if (isRecord(mapRaw)) {
      lowerAnalytics.regionalContextMap = {
        gatedLabel: str(mapRaw.gated_label) ?? str(mapRaw.gatedLabel),
        upgradeLabel: str(mapRaw.upgrade_label) ?? str(mapRaw.upgradeLabel),
      };
    }
    const accRaw = lowerRaw.regional_accuracy ?? lowerRaw.regionalAccuracy;
    if (Array.isArray(accRaw)) {
      const regionalAccuracy = accRaw
        .map((item) => {
          if (!isRecord(item)) return null;
          const region = str(item.region);
          const score = num(item.score);
          if (!region || score == null) return null;
          return { region, score } satisfies TopicsRegionalAccuracyItem;
        })
        .filter((x): x is TopicsRegionalAccuracyItem => x != null);
      if (regionalAccuracy.length > 0) lowerAnalytics.regionalAccuracy = regionalAccuracy;
    }
  }

  return {
    header: {
      title,
      subtitle,
      terminalOnlyActionLabel:
        str(headerRaw.terminal_only_action_label) ?? str(headerRaw.terminalOnlyActionLabel),
    },
    categories,
    quickFilters,
    browseRows,
    selectedTopic: selectedTopic ?? undefined,
    selectedTopicChart,
    rightRail,
    lowerAnalytics,
  };
}

export function clusterMixLabel(cluster: TopicsClusterKind): string {
  switch (cluster) {
    case "long":
      return "Long";
    case "short":
      return "Short";
    case "neutral":
      return "Neutral";
    case "speculative":
      return "Speculative";
    default:
      return "Unclustered";
  }
}

export function clusterMixColorVar(cluster: TopicsClusterKind): string {
  switch (cluster) {
    case "long":
      return "var(--af-tone-b)";
    case "short":
      return "var(--red)";
    case "neutral":
      return "var(--af-tone-d)";
    case "speculative":
      return "var(--af-tone-c)";
    default:
      return "var(--ink3)";
  }
}

export function inferCategoryValueFromTopicClass(topicClass: string): string {
  const u = topicClass.toUpperCase();
  if (u.includes("NBA") || u.includes("SPORT")) return "sports";
  if (u.includes("MACRO") || u.includes("FED")) return "macro";
  if (u.includes("TECH") || u.includes("AI") || u.includes("AGI")) return "tech";
  if (u.includes("FX") || u.includes("JPY") || u.includes("USD")) return "fx";
  if (u.includes("POLICY") || u.includes("EU ") || u.includes("/ EU")) return "policy";
  return "all";
}

export function consensusStatusLabel(s: TopicsConsensusStatus | undefined): string {
  switch (s) {
    case "consensus":
      return "Consensus";
    case "divergent":
      return "Divergent";
    case "low_signal":
      return "Low signal";
    case "speculative":
      return "Speculative participation";
    default:
      return "—";
  }
}

function formatPct01(v: number | undefined): string {
  if (v == null || !Number.isFinite(v)) return "—";
  return `${Math.round(v * 100)}%`;
}

function formatDelta01(v: number | undefined): string {
  if (v == null || !Number.isFinite(v)) return "0%";
  const p = Math.round(v * 100);
  if (p === 0) return "0%";
  return `${p > 0 ? "+" : ""}${p}%`;
}

/** Build selected-topic preview from a browse row when API does not send `selected_topic`. */
export function previewFromBrowseRow(row: TopicsBrowseRowModel): TopicsSelectedTopicPreviewModel {
  const rawHint = row.topSignalHint?.trim() ?? "";
  const hint = rawHint.toLowerCase();
  const longMatch = /\blong\b/.test(hint);
  const shortMatch = /\bshort\b/.test(hint);
  const agentFromHint = rawHint.replace(/\s*(long|short)\b.*$/i, "").trim();

  const emptySig = { agentName: "—", proofLabel: null as string | null };
  let topLong = emptySig;
  let topShort = emptySig;
  if (shortMatch && !longMatch) {
    topShort = { agentName: agentFromHint || "—", proofLabel: null };
  } else if (longMatch && !shortMatch) {
    topLong = { agentName: agentFromHint || "—", proofLabel: row.proofHint ?? null };
  } else if (longMatch && shortMatch) {
    topLong = { agentName: agentFromHint || "—", proofLabel: row.proofHint ?? null };
    topShort = { agentName: agentFromHint || "—", proofLabel: null };
  } else if (rawHint) {
    topLong = { agentName: rawHint, proofLabel: row.proofHint ?? null };
  }

  return {
    topicId: row.topicId,
    title: row.title,
    topicClass: row.topicClass,
    probabilityLong: row.probabilityLong,
    probabilityShort: row.probabilityShort,
    probabilityDelta: row.probabilityDelta,
    consensusStatus: row.consensusStatus,
    participationContext: {
      speculativeParticipationShare: row.consensusStatus === "speculative" ? 0.12 : 0.05,
      neutralClusterShare: 0.1,
      unclusteredShare: 0.03,
    },
    topLongPreview: topLong,
    topShortPreview: topShort,
    openTopicDetailsUrl: row.openTopicDetailsUrl,
    openResearchUrl: "/research",
  };
}

export function chartFromPreview(
  p: TopicsSelectedTopicPreviewModel,
  kind: "pie" | "donut" = "donut",
): TopicsSelectedTopicChartModel {
  return {
    kind,
    longPercent: p.probabilityLong,
    shortPercent: p.probabilityShort,
  };
}

export { formatPct01, formatDelta01 };

export const defaultTopicsPageModel: TopicsPageModel = {
  header: {
    title: "Topics",
    subtitle: "Live browse surface across active topics.",
    terminalOnlyActionLabel: "Propose topic — Terminal only",
  },
  categories: [
    { label: "All", value: "all", active: true },
    { label: "Sports", value: "sports" },
    { label: "Macro", value: "macro" },
    { label: "Tech", value: "tech" },
    { label: "Policy", value: "policy" },
    { label: "FX", value: "fx" },
  ],
  quickFilters: [
    { label: "Consensus", value: "consensus" },
    { label: "Divergent", value: "divergent" },
    { label: "Low signal", value: "low_signal" },
    { label: "Speculative participation", value: "speculative" },
    { label: "Watchlist", value: "watchlist" },
    { label: "Saved view", value: "saved_view" },
  ],
  browseRows: [
    {
      topicId: "Q.01",
      title: "Celtics will win the NBA Finals",
      topicClass: "Sport / NBA",
      categoryValue: "sports",
      probabilityLong: 0.67,
      probabilityShort: 0.33,
      probabilityDelta: 0.04,
      consensusStatus: "consensus",
      deadlineLabel: "Game 1",
      agentCount: 2104,
      topSignalHint: "agent-Ω long",
      proofHint: "ZK proof",
      openTopicDetailsUrl: "/topic/Q.01",
      watchlisted: true,
    },
    {
      topicId: "Q.02",
      title: "Fed rate cut — June meeting",
      topicClass: "Macro / Fed",
      categoryValue: "macro",
      probabilityLong: 0.51,
      probabilityShort: 0.49,
      probabilityDelta: -0.01,
      consensusStatus: "divergent",
      deadlineLabel: "Jun 11",
      agentCount: 1340,
      topSignalHint: "agent-a long",
      proofHint: null,
      openTopicDetailsUrl: "/topic/Q.02",
    },
    {
      topicId: "Q.03",
      title: "GPT-6 release before Q3 2026",
      topicClass: "Tech / AI",
      categoryValue: "tech",
      probabilityLong: 0.44,
      probabilityShort: 0.56,
      probabilityDelta: 0.02,
      consensusStatus: "speculative",
      deadlineLabel: "Sep 30",
      agentCount: 988,
      topSignalHint: "agent-γ long",
      proofHint: null,
      openTopicDetailsUrl: "/topic/Q.03",
    },
    {
      topicId: "Q.04",
      title: "Yen breaks 160 vs USD",
      topicClass: "FX / JPY",
      categoryValue: "fx",
      probabilityLong: 0.38,
      probabilityShort: 0.62,
      probabilityDelta: 0,
      consensusStatus: "divergent",
      deadlineLabel: "May 31",
      agentCount: 604,
      topSignalHint: "agent-λ long",
      proofHint: null,
      openTopicDetailsUrl: "/topic/Q.04",
    },
    {
      topicId: "Q.05",
      title: "EU AI Act — first enforcement case",
      topicClass: "Policy / EU",
      categoryValue: "policy",
      probabilityLong: 0.22,
      probabilityShort: 0.78,
      probabilityDelta: -0.03,
      consensusStatus: "low_signal",
      deadlineLabel: "Dec 31",
      agentCount: 312,
      topSignalHint: null,
      proofHint: null,
      openTopicDetailsUrl: "/topic/Q.05",
    },
    {
      topicId: "Q.06",
      title: "AGI threshold declared by 2027",
      topicClass: "Tech / AGI",
      categoryValue: "tech",
      probabilityLong: 0.17,
      probabilityShort: 0.83,
      probabilityDelta: 0.01,
      consensusStatus: "speculative",
      deadlineLabel: "2027",
      agentCount: 201,
      topSignalHint: "agent-κ short",
      proofHint: null,
      openTopicDetailsUrl: "/topic/Q.06",
    },
  ],
  selectedTopic: {
    topicId: "Q.01",
    title: "Celtics will win the NBA Finals",
    topicClass: "Sport / NBA",
    probabilityLong: 0.67,
    probabilityShort: 0.33,
    probabilityDelta: 0.04,
    consensusStatus: "consensus",
    participationContext: {
      speculativeParticipationShare: 0.05,
      neutralClusterShare: 0.1,
      unclusteredShare: 0.03,
    },
    topLongPreview: { agentName: "agent-Ω", proofLabel: "ZK proof" },
    topShortPreview: { agentName: "agent-β", proofLabel: null },
    openTopicDetailsUrl: "/topic/Q.01",
    openResearchUrl: "/research",
  },
  selectedTopicChart: {
    kind: "donut",
    longPercent: 0.67,
    shortPercent: 0.33,
  },
  rightRail: {
    dailyDigestTakeaway: {
      title: "Long bias",
      subtitle: "67% weighted",
      note: "CN short bias",
    },
    clusterActivity: [
      { cluster: "long", count: 312 },
      { cluster: "short", count: 228 },
      { cluster: "neutral", count: 198 },
      { cluster: "speculative", count: 109 },
      { cluster: "unclustered", count: 44 },
    ],
    regionalDivergence: {
      summary: "CN short vs US long on Q.01",
      openRegionalDetailUrl: "/topic/Q.01?view=regional&timeframe=7d",
    },
  },
  lowerAnalytics: {
    regionalContextMap: {
      gatedLabel: "Interactive map — Analyst+",
      upgradeLabel: "Upgrade",
    },
    regionalAccuracy: [
      { region: "US", score: 88 },
      { region: "JP/KR", score: 84 },
      { region: "EU", score: 76 },
      { region: "CN", score: 71 },
      { region: "SE Asia", score: 58 },
    ],
  },
};
