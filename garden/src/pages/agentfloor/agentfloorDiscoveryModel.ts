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
  proofType?: "zkml" | "tee" | null;
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
};

/** Minimal wire shape the page can receive from GET /floor/agents/... later. */
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
  };
}
