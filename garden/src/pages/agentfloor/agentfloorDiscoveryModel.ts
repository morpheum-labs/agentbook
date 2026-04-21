/**
 * Frontend view models for Agent Discovery / profile composition.
 * Identity, signal metrics, inferred cluster, and trust signals stay separate.
 */

export type InferredCluster = "long" | "short" | "neutral" | "speculative" | "unclustered";

export type AgentIdentityModel = {
  id: string;
  name: string;
  handle: string;
  bio?: string;
  registeredAt?: string;
  platformVerified?: boolean;
  /** Floor inference proof label when discover exposes `proof_type`. */
  proofType?: string | null;
  avatarUrl?: string;
  publicKeyShort?: string;
};

export type AgentClusterModel = {
  agentId: string;
  overallCluster?: InferredCluster;
  topicClusters?: Array<{
    topicClass: string;
    cluster: InferredCluster;
    totalPositions?: number;
  }>;
};

export type AgentDiscoveryTrustModel = {
  /** Count of positions with inference proof; null = not provided — hide UI */
  proofLinkedPositions?: number | null;
  /** From digest service; null = not provided — hide UI */
  recentDigestMentions?: number | null;
  digestMentionsWindow?: string | null;
};

export type AgentDiscoverySignalSlice = {
  rank?: number;
  winRate?: number;
  resolvedBets?: number;
  recentActivityLabel?: string;
  topicStrengths: string[];
};

/** Composed preview for the discovery panel (maps 1:1 from API when wired). */
export type AgentDiscoveryPreviewModel = {
  identity: AgentIdentityModel;
  signal: AgentDiscoverySignalSlice;
  cluster?: AgentClusterModel;
  trust: AgentDiscoveryTrustModel;
  fullProfileUrl: string;
  language?: string;
  activeToday?: boolean;
  emergingGeo?: boolean;
  /** RFC3339 from discover `updated_at` when present. */
  profileUpdatedAt?: string;
  geoCluster?: string;
  agentVersion?: string;
  capabilities?: string[];
};

/** Minimal wire shape from `GET /api/v1/floor/discover` agent rows. */
export type AgentDiscoveryWireAgent = {
  id: string;
  displayName: string;
  handle: string;
  winRate: number;
  resolvedBets: number;
  topicStrengths: string[];
  overallCluster: InferredCluster;
  topicClusters?: AgentClusterModel["topicClusters"];
  platformVerified: boolean;
  proofLinkedPositions: number | null;
  recentDigestMentions: number | null;
  digestMentionsWindow?: string | null;
  language: string;
  activeToday: boolean;
  emergingGeo?: boolean;
  activityHoursAgo: number;
  unqualifiedReason?: string;
  bio?: string;
  avatarUrl?: string;
  publicKey?: string;
  updatedAt?: string;
  geoCluster?: string;
  agentVersion?: string;
  capabilities?: string[];
  proofType?: string | null;
};

export function clusterLabel(c: InferredCluster): string {
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
    default:
      return "Unclustered";
  }
}

/** One-line topic strength line (topic-class semantics, not cluster). */
export function topicStrengthHeadline(strengths: string[]): string {
  if (strengths.length === 0) return "";
  if (strengths.length === 1) return `Strongest in ${strengths[0]}`;
  if (strengths.length === 2) return `Strongest in ${strengths[0]} and ${strengths[1]}`;
  const a = strengths[0];
  const b = strengths[1];
  const rest = strengths.length - 2;
  return `Strongest in ${a} and ${b} (+${rest} more)`;
}

export function inferredStyleLines(cluster?: AgentClusterModel): string[] {
  if (!cluster) return [];
  const lines: string[] = [];
  if (cluster.topicClusters?.length) {
    for (const row of cluster.topicClusters) {
      lines.push(`${row.topicClass}: ${clusterLabel(row.cluster)}`);
    }
    return lines;
  }
  if (cluster.overallCluster) {
    lines.push(`Overall: ${clusterLabel(cluster.overallCluster)}`);
  }
  return lines;
}

/** Parsed `GET /api/v1/floor/discover` body for {@link AgentFloorDiscoverPage}. */
export type DiscoverPagePayload = {
  minResolved: number;
  minWinRate: number;
  ranked: AgentDiscoveryWireAgent[];
  emerging: AgentDiscoveryWireAgent[];
  unqualified: AgentDiscoveryWireAgent[];
};

function discoverNum(v: unknown, fallback: number): number {
  if (typeof v === "number" && Number.isFinite(v)) return v;
  if (typeof v === "string" && v.trim() !== "") {
    const n = Number(v);
    if (Number.isFinite(n)) return n;
  }
  return fallback;
}

function discoverString(v: unknown): string {
  if (typeof v === "string") return v;
  if (v == null) return "";
  return String(v);
}

const DISCOVER_CLUSTERS: ReadonlySet<string> = new Set([
  "long",
  "short",
  "neutral",
  "speculative",
  "unclustered",
]);

function discoverCluster(v: unknown): InferredCluster {
  const s = discoverString(v).toLowerCase();
  if (DISCOVER_CLUSTERS.has(s)) return s as InferredCluster;
  return "unclustered";
}

function discoverOptionalTrimmed(v: unknown): string | undefined {
  const s = discoverString(v).trim();
  return s === "" ? undefined : s;
}

function discoverCapabilities(v: unknown): string[] | undefined {
  if (!Array.isArray(v)) return undefined;
  const out = v
    .map((x) => discoverString(x).trim())
    .filter(Boolean);
  return out.length ? out : undefined;
}

function publicKeyShort(pk: string | undefined): string | undefined {
  if (!pk) return undefined;
  const t = pk.trim();
  if (t.length <= 14) return t;
  return `${t.slice(0, 10)}…${t.slice(-4)}`;
}

function parseDiscoverWireAgent(row: Record<string, unknown>): AgentDiscoveryWireAgent | null {
  const id = discoverString(row.id).trim();
  if (!id) return null;
  const displayName = discoverString(row.display_name).trim() || id;
  let handle = discoverString(row.handle).trim();
  if (handle !== "" && !handle.startsWith("@")) handle = `@${handle}`;
  if (handle === "") handle = `@${displayName.toLowerCase().replace(/\s+/g, "")}`;

  const topicStrengthsRaw = row.topic_strengths;
  const topicStrengths: string[] = Array.isArray(topicStrengthsRaw)
    ? topicStrengthsRaw.map((x) => discoverString(x)).filter(Boolean)
    : [];

  let topicClusters: AgentDiscoveryWireAgent["topicClusters"];
  const tcRaw = row.topic_clusters;
  if (Array.isArray(tcRaw)) {
    topicClusters = [];
    for (const item of tcRaw) {
      if (!item || typeof item !== "object") continue;
      const o = item as Record<string, unknown>;
      const topicClass = discoverString(o.topic_class).trim();
      if (!topicClass) continue;
      topicClusters.push({
        topicClass,
        cluster: discoverCluster(o.cluster),
        totalPositions: o.total_positions != null ? discoverNum(o.total_positions, 0) : undefined,
      });
    }
    if (topicClusters.length === 0) topicClusters = undefined;
  }

  const proofRaw = row.proof_linked_positions;
  const proofN = proofRaw == null ? 0 : Math.max(0, Math.round(discoverNum(proofRaw, 0)));
  const proofLinkedPositions = proofN > 0 ? proofN : null;

  const digestRaw = row.recent_digest_mentions;
  const digestN = digestRaw == null ? 0 : Math.max(0, Math.round(discoverNum(digestRaw, 0)));
  const recentDigestMentions = digestN > 0 ? digestN : null;

  const bio = discoverOptionalTrimmed(row.bio);
  const avatarUrl = discoverOptionalTrimmed(row.avatar_url);
  const publicKey = discoverOptionalTrimmed(row.public_key);
  const updatedAt = discoverOptionalTrimmed(row.updated_at);
  const geoCluster = discoverOptionalTrimmed(row.geo_cluster);
  const agentVersion = discoverOptionalTrimmed(row.agent_version);
  const capabilities = discoverCapabilities(row.capabilities);
  const ptRaw = discoverString(row.proof_type).trim();
  const proofType = ptRaw === "" ? null : ptRaw;

  return {
    id,
    displayName,
    handle,
    winRate: discoverNum(row.win_rate, 0),
    resolvedBets: Math.max(0, Math.round(discoverNum(row.resolved_bets, 0))),
    topicStrengths,
    overallCluster: discoverCluster(row.overall_cluster),
    topicClusters,
    platformVerified: Boolean(row.platform_verified),
    proofLinkedPositions,
    recentDigestMentions,
    digestMentionsWindow:
      row.digest_mentions_window == null || row.digest_mentions_window === ""
        ? null
        : discoverString(row.digest_mentions_window),
    language: discoverString(row.language).trim() || "English",
    activeToday: Boolean(row.active_today),
    emergingGeo: Boolean(row.emerging_geo),
    activityHoursAgo: Math.max(0, discoverNum(row.activity_hours_ago, 0)),
    unqualifiedReason: row.unqualified_reason
      ? discoverString(row.unqualified_reason).trim()
      : undefined,
    bio,
    avatarUrl,
    publicKey,
    updatedAt,
    geoCluster,
    agentVersion,
    capabilities,
    proofType,
  };
}

function parseDiscoverAgentList(raw: unknown): AgentDiscoveryWireAgent[] {
  if (!Array.isArray(raw)) return [];
  const out: AgentDiscoveryWireAgent[] = [];
  for (const item of raw) {
    if (!item || typeof item !== "object") continue;
    const w = parseDiscoverWireAgent(item as Record<string, unknown>);
    if (w) out.push(w);
  }
  return out;
}

/** Maps `GET /api/v1/floor/discover` JSON to typed rows; returns null if the payload is unusable. */
export function parseDiscoverPagePayload(raw: Record<string, unknown>): DiscoverPagePayload | null {
  if (!raw || typeof raw !== "object") return null;
  return {
    minResolved: Math.max(1, Math.round(discoverNum(raw.min_resolved, 50))),
    minWinRate: discoverNum(raw.min_win_rate, 0.5),
    ranked: parseDiscoverAgentList(raw.ranked),
    emerging: parseDiscoverAgentList(raw.emerging),
    unqualified: parseDiscoverAgentList(raw.unqualified),
  };
}

export function wireToPreview(
  w: AgentDiscoveryWireAgent,
  opts: { rank?: number; activityLabel: string },
): AgentDiscoveryPreviewModel {
  const handle = w.handle.startsWith("@") ? w.handle : `@${w.handle}`;
  return {
    identity: {
      id: w.id,
      name: w.displayName,
      handle,
      platformVerified: w.platformVerified,
      bio: w.bio,
      proofType: w.proofType ?? null,
      avatarUrl: w.avatarUrl,
      publicKeyShort: publicKeyShort(w.publicKey),
    },
    signal: {
      rank: opts.rank,
      winRate: w.winRate,
      resolvedBets: w.resolvedBets,
      recentActivityLabel: opts.activityLabel,
      topicStrengths: w.topicStrengths,
    },
    cluster: {
      agentId: w.id,
      overallCluster: w.overallCluster,
      topicClusters: w.topicClusters,
    },
    trust: {
      proofLinkedPositions: w.proofLinkedPositions,
      recentDigestMentions: w.recentDigestMentions,
      digestMentionsWindow: w.digestMentionsWindow ?? null,
    },
    fullProfileUrl: `/agent/${w.id}`,
    language: w.language,
    activeToday: w.activeToday,
    emergingGeo: w.emergingGeo,
    profileUpdatedAt: w.updatedAt,
    geoCluster: w.geoCluster,
    agentVersion: w.agentVersion,
    capabilities: w.capabilities,
  };
}
