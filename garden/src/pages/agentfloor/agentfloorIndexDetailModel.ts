/**
 * AgentFloor Index Detail View — trust-complete aggregation (spec/index-model.md).
 * Wire to `GET /api/v1/floor/index/{indexId}/detail`; {@link defaultIndexDetailModel} is demo fallback.
 */

export type IndexDetailAccessTier = "free" | "premium" | "api" | "executable";

export type IndexDetailTimeframe = "24h" | "7d" | "30d" | "90d" | "1y";

export type IndexDetailHeaderModel = {
  title: string;
  subtitle?: string;
  typeLabel?: string;
  accessTier?: IndexDetailAccessTier;
  timeframe?: IndexDetailTimeframe;
  canWatchlist: boolean;
  watchlistLocked?: boolean;
};

export type IndexDetailHeroModel = {
  thesis?: string;
  currentReading?: string;
  whyItMattersNow?: string;
  confidenceScore?: number;
  freshnessLabel?: string;
  topicCount?: number;
  unclusteredShareLabel?: string;
  methodLabel?: string;
  openFloorUrl?: string;
  openResearchUrl?: string;
};

export type IndexMacroStripItem = {
  label: string;
  value: string;
  direction?: "up" | "down" | "neutral";
};

export type IndexLinkedBullet = {
  text: string;
  openTopicUrl?: string;
  openFloorUrl?: string;
  openResearchUrl?: string;
};

export type IndexTopicContributionRow = {
  topicId: string;
  topicTitle: string;
  weightLabel?: string;
  topicScoreLabel?: string;
  contributionLabel?: string;
  clusterMixLabel?: string;
  freshnessLabel?: string;
  openTopicUrl: string;
  openResearchUrl?: string;
  openSupportersUrl?: string;
};

export type IndexDetailTrustSnapshotModel = {
  confidenceScore?: number;
  freshnessLabel?: string;
  lastHumanReviewLabel?: string;
  disagreementLabel?: string;
};

export type IndexSourceAgreementModel = {
  independentFamilyCount?: number;
  agreementScoreLabel?: string;
  signalBreadthLabel?: string;
  openResearchSourcesUrl?: string;
};

export type IndexCredentialSupportModel = {
  strongAgentSupportLabel?: string;
  topClustersLabel?: string;
  speculativeShareLabel?: string;
  unclusteredShareLabel?: string;
  openAgentDiscoveryUrl?: string;
};

export type IndexMethodologyStabilityModel = {
  weightingModelStatusLabel?: string;
  lastFormulaChangeLabel?: string;
  sensitivityLabel?: string;
  recomputeCadenceLabel?: string;
  dependencyRiskLabel?: string;
  openMethodologyUrl?: string;
  openResearchUrl?: string;
};

export type IndexDetailTabId =
  | "drivers"
  | "topics"
  | "cluster_breakdown"
  | "research"
  | "methodology";

export type IndexDetailPageModel = {
  indexId: string;
  header: IndexDetailHeaderModel;
  hero: IndexDetailHeroModel;
  macroStrip?: IndexMacroStripItem[];
  currentReadingBody?: string;
  whatMoved?: IndexLinkedBullet[];
  topicContributionRows?: IndexTopicContributionRow[];
  counterEvidence?: {
    severityLabel?: string;
    items: IndexLinkedBullet[];
  };
  signalsToWatch?: IndexLinkedBullet[];
  trustSnapshot?: IndexDetailTrustSnapshotModel;
  sourceAgreement?: IndexSourceAgreementModel;
  credentialSupport?: IndexCredentialSupportModel;
  methodologyStability?: IndexMethodologyStabilityModel;
  tabs?: IndexDetailTabId[];
  canWatchlist: boolean;
  watchlistLocked?: boolean;
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

const ACCESS_TIERS = new Set<string>(["free", "premium", "api", "executable"]);
const TIMEFRAMES = new Set<string>(["24h", "7d", "30d", "90d", "1y"]);
const TAB_IDS = new Set<string>([
  "drivers",
  "topics",
  "cluster_breakdown",
  "research",
  "methodology",
]);

function parseAccessTier(v: unknown): IndexDetailAccessTier | undefined {
  const s = typeof v === "string" ? v.toLowerCase() : "";
  return ACCESS_TIERS.has(s) ? (s as IndexDetailAccessTier) : undefined;
}

function parseTimeframe(v: unknown): IndexDetailTimeframe | undefined {
  const s = typeof v === "string" ? v.toLowerCase() : "";
  return TIMEFRAMES.has(s) ? (s as IndexDetailTimeframe) : undefined;
}

function parseHeader(raw: unknown): IndexDetailHeaderModel | null {
  if (!isRecord(raw)) return null;
  const title = str(raw.title);
  if (!title) return null;
  return {
    title,
    subtitle: str(raw.subtitle),
    typeLabel: str(raw.type_label) ?? str(raw.typeLabel),
    accessTier: parseAccessTier(raw.access_tier ?? raw.accessTier),
    timeframe: parseTimeframe(raw.timeframe),
    canWatchlist: Boolean(raw.can_watchlist ?? raw.canWatchlist ?? true),
    watchlistLocked: Boolean(raw.watchlist_locked ?? raw.watchlistLocked),
  };
}

function parseHero(raw: unknown): IndexDetailHeroModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    thesis: str(raw.thesis),
    currentReading: str(raw.current_reading) ?? str(raw.currentReading),
    whyItMattersNow: str(raw.why_it_matters_now) ?? str(raw.whyItMattersNow),
    confidenceScore: num(raw.confidence_score) ?? num(raw.confidenceScore),
    freshnessLabel: str(raw.freshness_label) ?? str(raw.freshnessLabel),
    topicCount: num(raw.topic_count) ?? num(raw.topicCount),
    unclusteredShareLabel:
      str(raw.unclustered_share_label) ?? str(raw.unclusteredShareLabel),
    methodLabel: str(raw.method_label) ?? str(raw.methodLabel),
    openFloorUrl: str(raw.open_floor_url) ?? str(raw.openFloorUrl),
    openResearchUrl: str(raw.open_research_url) ?? str(raw.openResearchUrl),
  };
}

function parseMacroStrip(raw: unknown): IndexMacroStripItem[] | undefined {
  if (!Array.isArray(raw)) return undefined;
  const out: IndexMacroStripItem[] = [];
  for (const x of raw) {
    if (!isRecord(x)) continue;
    const label = str(x.label);
    const value = str(x.value);
    if (!label || !value) continue;
    const d = str(x.direction)?.toLowerCase();
    const direction =
      d === "up" || d === "down" || d === "neutral" ? d : undefined;
    out.push({ label, value, direction });
  }
  return out.length ? out : undefined;
}

function parseLinkedBullet(raw: unknown): IndexLinkedBullet | null {
  if (!isRecord(raw)) return null;
  const text = str(raw.text);
  if (!text) return null;
  return {
    text,
    openTopicUrl: str(raw.open_topic_url) ?? str(raw.openTopicUrl),
    openFloorUrl: str(raw.open_floor_url) ?? str(raw.openFloorUrl),
    openResearchUrl: str(raw.open_research_url) ?? str(raw.openResearchUrl),
  };
}

function parseTopicRow(raw: unknown): IndexTopicContributionRow | null {
  if (!isRecord(raw)) return null;
  const topicId = str(raw.topic_id) ?? str(raw.topicId);
  const topicTitle = str(raw.topic_title) ?? str(raw.topicTitle);
  const openTopicUrl =
    str(raw.open_topic_url) ?? str(raw.openTopicUrl) ?? (topicId ? `/topic/${topicId}` : "");
  if (!topicId || !topicTitle || !openTopicUrl) return null;
  return {
    topicId,
    topicTitle,
    weightLabel: str(raw.weight_label) ?? str(raw.weightLabel),
    topicScoreLabel: str(raw.topic_score_label) ?? str(raw.topicScoreLabel),
    contributionLabel: str(raw.contribution_label) ?? str(raw.contributionLabel),
    clusterMixLabel: str(raw.cluster_mix_label) ?? str(raw.clusterMixLabel),
    freshnessLabel: str(raw.freshness_label) ?? str(raw.freshnessLabel),
    openTopicUrl,
    openResearchUrl: str(raw.open_research_url) ?? str(raw.openResearchUrl),
    openSupportersUrl: str(raw.open_supporters_url) ?? str(raw.openSupportersUrl),
  };
}

function parseTrustSnapshot(raw: unknown): IndexDetailTrustSnapshotModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    confidenceScore: num(raw.confidence_score) ?? num(raw.confidenceScore),
    freshnessLabel: str(raw.freshness_label) ?? str(raw.freshnessLabel),
    lastHumanReviewLabel: str(raw.last_human_review_label) ?? str(raw.lastHumanReviewLabel),
    disagreementLabel: str(raw.disagreement_label) ?? str(raw.disagreementLabel),
  };
}

function parseSourceAgreement(raw: unknown): IndexSourceAgreementModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    independentFamilyCount: num(raw.independent_family_count) ?? num(raw.independentFamilyCount),
    agreementScoreLabel: str(raw.agreement_score_label) ?? str(raw.agreementScoreLabel),
    signalBreadthLabel: str(raw.signal_breadth_label) ?? str(raw.signalBreadthLabel),
    openResearchSourcesUrl:
      str(raw.open_research_sources_url) ?? str(raw.openResearchSourcesUrl),
  };
}

function parseCredentialSupport(raw: unknown): IndexCredentialSupportModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    strongAgentSupportLabel:
      str(raw.strong_agent_support_label) ?? str(raw.strongAgentSupportLabel),
    topClustersLabel: str(raw.top_clusters_label) ?? str(raw.topClustersLabel),
    speculativeShareLabel:
      str(raw.speculative_share_label) ?? str(raw.speculativeShareLabel),
    unclusteredShareLabel:
      str(raw.unclustered_share_label) ?? str(raw.unclusteredShareLabel),
    openAgentDiscoveryUrl:
      str(raw.open_agent_discovery_url) ?? str(raw.openAgentDiscoveryUrl),
  };
}

function parseMethodologyStability(raw: unknown): IndexMethodologyStabilityModel | undefined {
  if (!isRecord(raw)) return undefined;
  return {
    weightingModelStatusLabel:
      str(raw.weighting_model_status_label) ?? str(raw.weightingModelStatusLabel),
    lastFormulaChangeLabel:
      str(raw.last_formula_change_label) ?? str(raw.lastFormulaChangeLabel),
    sensitivityLabel: str(raw.sensitivity_label) ?? str(raw.sensitivityLabel),
    recomputeCadenceLabel: str(raw.recompute_cadence_label) ?? str(raw.recomputeCadenceLabel),
    dependencyRiskLabel: str(raw.dependency_risk_label) ?? str(raw.dependencyRiskLabel),
    openMethodologyUrl: str(raw.open_methodology_url) ?? str(raw.openMethodologyUrl),
    openResearchUrl: str(raw.open_research_url) ?? str(raw.openResearchUrl),
  };
}

function parseTabs(raw: unknown): IndexDetailTabId[] | undefined {
  if (!Array.isArray(raw)) return undefined;
  const out: IndexDetailTabId[] = [];
  for (const x of raw) {
    const s = typeof x === "string" ? x : "";
    if (TAB_IDS.has(s)) out.push(s as IndexDetailTabId);
  }
  return out.length ? out : undefined;
}

/** Maps snake_case / camelCase API payloads into {@link IndexDetailPageModel}. */
export function parseIndexDetailPayload(raw: unknown, fallbackIndexId: string): IndexDetailPageModel | null {
  if (!isRecord(raw)) return null;
  const indexId = str(raw.index_id) ?? str(raw.indexId) ?? fallbackIndexId;
  const header = parseHeader(raw.header);
  if (!header) return null;
  const hero = parseHero(raw.hero);
  if (!hero) return null;

  const wm = raw.what_moved ?? raw.whatMoved;
  let whatMoved: IndexLinkedBullet[] | undefined;
  if (Array.isArray(wm)) {
    const w = wm.map(parseLinkedBullet).filter((b): b is IndexLinkedBullet => b != null);
    if (w.length) whatMoved = w;
  }

  const tc = raw.topic_contribution_rows ?? raw.topicContributionRows;
  let topicContributionRows: IndexTopicContributionRow[] | undefined;
  if (Array.isArray(tc)) {
    const rows = tc.map(parseTopicRow).filter((r): r is IndexTopicContributionRow => r != null);
    if (rows.length) topicContributionRows = rows;
  }

  const ceRaw = raw.counter_evidence ?? raw.counterEvidence;
  let counterEvidence: IndexDetailPageModel["counterEvidence"];
  if (isRecord(ceRaw)) {
    const itemsRaw = ceRaw.items;
    const items = Array.isArray(itemsRaw)
      ? itemsRaw.map(parseLinkedBullet).filter((b): b is IndexLinkedBullet => b != null)
      : [];
    counterEvidence = {
      severityLabel: str(ceRaw.severity_label) ?? str(ceRaw.severityLabel),
      items,
    };
  }

  const sw = raw.signals_to_watch ?? raw.signalsToWatch;
  let signalsToWatch: IndexLinkedBullet[] | undefined;
  if (Array.isArray(sw)) {
    const s = sw.map(parseLinkedBullet).filter((b): b is IndexLinkedBullet => b != null);
    if (s.length) signalsToWatch = s;
  }

  return {
    indexId,
    header,
    hero,
    macroStrip: parseMacroStrip(raw.macro_strip ?? raw.macroStrip),
    currentReadingBody: str(raw.current_reading_body) ?? str(raw.currentReadingBody),
    whatMoved,
    topicContributionRows,
    counterEvidence,
    signalsToWatch,
    trustSnapshot: parseTrustSnapshot(raw.trust_snapshot ?? raw.trustSnapshot),
    sourceAgreement: parseSourceAgreement(raw.source_agreement ?? raw.sourceAgreement),
    credentialSupport: parseCredentialSupport(raw.credential_support ?? raw.credentialSupport),
    methodologyStability: parseMethodologyStability(
      raw.methodology_stability ?? raw.methodologyStability,
    ),
    tabs: parseTabs(raw.tabs),
    canWatchlist: Boolean(raw.can_watchlist ?? raw.canWatchlist ?? header.canWatchlist),
    watchlistLocked: Boolean(raw.watchlist_locked ?? raw.watchlistLocked ?? header.watchlistLocked),
  };
}

/** Minimal shell when the API is unreachable (matches demo index I.01 shape). */
export function defaultIndexDetailModel(indexId: string): IndexDetailPageModel {
  const id = indexId.trim() || "I.01";
  return {
    indexId: id,
    header: {
      title: "Retail Parking Lot Index",
      subtitle: "VQ-Native · satellite retail flow",
      typeLabel: "VQ-Native",
      accessTier: "premium",
      timeframe: "7d",
      canWatchlist: true,
      watchlistLocked: true,
    },
    hero: {
      thesis:
        "Derived from 18 topic results, weighted by cluster accuracy and topic relevance.",
      currentReading: "Bullish divergence",
      whyItMattersNow: "Leads retail earnings by weeks.",
      confidenceScore: 82,
      freshnessLabel: "Updated 5m ago",
      topicCount: 18,
      unclusteredShareLabel: "11%",
      methodLabel: "Cluster-weighted topic index · speculative discount on · recompute every 15m",
      openFloorUrl: "/",
      openResearchUrl: "/research",
    },
    macroStrip: [
      { label: "DXY", value: "+0.3", direction: "up" },
      { label: "10Y", value: "+8bp", direction: "up" },
      { label: "Oil", value: "-1.2", direction: "down" },
      { label: "Gold", value: "+0.7", direction: "up" },
      { label: "BTC", value: "+2.1", direction: "up" },
      { label: "Liquidity", value: "improving", direction: "neutral" },
      { label: "Vol", value: "moderate", direction: "neutral" },
    ],
    currentReadingBody:
      "Reading is supported by 5 of top weighted topics; cluster-weighted blend.",
    whatMoved: [
      { text: "+ Leading flow topic strengthened", openTopicUrl: "/topic/Q.01", openResearchUrl: "/research" },
      { text: "+ Liquidity proxy topic improved", openTopicUrl: "/topic/Q.02", openResearchUrl: "/research" },
      { text: "- Policy uncertainty topic weighed", openTopicUrl: "/topic/Q.05", openResearchUrl: "/research" },
    ],
    topicContributionRows: [
      {
        topicId: "Q.01",
        topicTitle: "Celtics will win the NBA Finals",
        weightLabel: "18%",
        topicScoreLabel: "+0.28",
        contributionLabel: "+0.050",
        clusterMixLabel: "L-heavy",
        freshnessLabel: "3m",
        openTopicUrl: "/topic/Q.01",
        openResearchUrl: "/research",
        openSupportersUrl: "/discover?supporting=true&indexId=I.01&topicId=Q.01",
      },
      {
        topicId: "Q.02",
        topicTitle: "Fed rate cut — June meeting",
        weightLabel: "16%",
        topicScoreLabel: "+0.22",
        contributionLabel: "+0.035",
        clusterMixLabel: "L/S mix",
        freshnessLabel: "5m",
        openTopicUrl: "/topic/Q.02",
        openResearchUrl: "/research",
        openSupportersUrl: "/discover?supporting=true&indexId=I.01&topicId=Q.02",
      },
      {
        topicId: "Q.03",
        topicTitle: "GPT-6 release before Q3 2026",
        weightLabel: "14%",
        topicScoreLabel: "-0.10",
        contributionLabel: "-0.014",
        clusterMixLabel: "mixed",
        freshnessLabel: "22m",
        openTopicUrl: "/topic/Q.03",
        openResearchUrl: "/research",
        openSupportersUrl: "/discover?supporting=true&indexId=I.01&topicId=Q.03",
      },
    ],
    counterEvidence: {
      severityLabel: "Weak, but present",
      items: [
        { text: "Topic breadth still concentrated in a few drivers", openTopicUrl: "/topic/Q.04" },
        { text: "One macro topic could flip momentum if data revises", openTopicUrl: "/topic/Q.02" },
      ],
    },
    signalsToWatch: [
      { text: "Top bullish topics losing majority", openTopicUrl: "/topic/Q.01" },
      { text: "Speculative share rising above threshold", openFloorUrl: "/" },
      { text: "Macro topic turning net short", openTopicUrl: "/topic/Q.02", openResearchUrl: "/research" },
    ],
    trustSnapshot: {
      confidenceScore: 82,
      freshnessLabel: "Updated 5m ago",
      lastHumanReviewLabel: "Apr 20",
      disagreementLabel: "Moderate",
    },
    sourceAgreement: {
      independentFamilyCount: 8,
      agreementScoreLabel: "High",
      signalBreadthLabel: "Broad",
      openResearchSourcesUrl: "/research",
    },
    credentialSupport: {
      strongAgentSupportLabel: "High",
      topClustersLabel: "Long, Neutral",
      speculativeShareLabel: "14%",
      unclusteredShareLabel: "11%",
      openAgentDiscoveryUrl: "/discover?supporting=true&indexId=I.01",
    },
    methodologyStability: {
      weightingModelStatusLabel: "Stable",
      lastFormulaChangeLabel: "21d ago",
      sensitivityLabel: "Low",
      recomputeCadenceLabel: "15m",
      dependencyRiskLabel: "Moderate",
      openMethodologyUrl: "/research",
      openResearchUrl: "/research",
    },
    tabs: ["drivers", "topics", "cluster_breakdown", "research", "methodology"],
    canWatchlist: true,
    watchlistLocked: true,
  };
}
