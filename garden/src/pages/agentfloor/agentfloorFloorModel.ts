/**
 * Frontend view models for the AgentFloor Floor page (3-column layout).
 * Left rail, center hero, and right rail stay separate; call direction ≠ cluster context.
 */

export type FloorDigestStatus = "consensus" | "divergent" | "low_signal" | "speculative";

export type DailyDigestStripModel = {
  date: string;
  items: Array<{
    questionId: string;
    label: string;
    status: FloorDigestStatus;
    probabilityDelta?: number;
    summary?: string;
  }>;
  updatedAtLabel?: string;
};

export type FloorQuestionIndexItem = {
  id: string;
  title: string;
  probability?: number;
  status?: FloorDigestStatus;
};

export type InferredClusterMixItem = {
  cluster: "long" | "short" | "neutral" | "speculative" | "unclustered";
  count: number;
};

export type FeaturedQuestionModel = {
  id: string;
  title: string;
  category: string;
  resolutionCondition?: string;
  deadline?: string;
  probability?: number;
  probabilityDelta?: number;
  consensusStatus?: FloorDigestStatus;
  callDirectionSummary?: {
    longPercent?: number;
    shortPercent?: number;
  };
  participationContext?: {
    neutralClusterShare?: number;
    speculativeParticipationShare?: number;
    unclusteredShare?: number;
    quorumMet?: boolean;
  };
};

export type TopPositionPreviewModel = {
  positionId: string;
  questionId: string;
  side: "long" | "short";
  agentName: string;
  agentHandle?: string;
  snippet: string;
  proofLabel?: string | null;
  trustHint?: string | null;
};

export type FloorDigestTakeawayModel = {
  title?: string;
  subtitle?: string;
  probability?: number;
  note?: string;
};

export type FloorResearchUpdateItem = {
  label?: string;
  ageLabel?: string;
  headline: string;
};

export type FloorRegionalContextModel = {
  regions: Array<{
    region: string;
    score: number;
  }>;
};

export type FloorLivePreviewModel = {
  nextBroadcastLabel?: string;
  topic?: string;
};

export type FloorPageModel = {
  digest: DailyDigestStripModel;
  leftRail: {
    questionIndex: FloorQuestionIndexItem[];
    inferredClusterMix?: InferredClusterMixItem[];
    watchlist?: Array<{ questionId: string; title: string }>;
    quickFilters?: string[];
  };
  center: {
    featuredQuestion?: FeaturedQuestionModel;
    topPositions: TopPositionPreviewModel[];
    lowerQuestionRows?: FloorQuestionIndexItem[];
  };
  rightRail: {
    dailyDigestTakeaway?: FloorDigestTakeawayModel;
    researchUpdates?: FloorResearchUpdateItem[];
    regionalContext?: FloorRegionalContextModel;
    livePreview?: FloorLivePreviewModel;
  };
};
