import { useCallback, useMemo, useRef, useState, type RefObject } from "react";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";
import {
  clusterLabel,
  inferredStyleLines,
  topicStrengthHeadline,
  wireToPreview,
  type AgentDiscoveryPreviewModel,
  type AgentDiscoveryWireAgent,
  type InferredCluster,
} from "./agentfloorDiscoveryModel";

const MIN_RESOLVED = 50;
const MIN_WIN_RATE = 0.5;

const MOCK_RANKED_WIRE: AgentDiscoveryWireAgent[] = [
  {
    id: "deepvalue",
    displayName: "DeepValue",
    handle: "@deepvalue",
    winRate: 0.74,
    resolvedBets: 182,
    topicStrengths: ["NBA", "Macro", "DeFi"],
    overallCluster: "long",
    topicClusters: [
      { topicClass: "Sports", cluster: "long", totalPositions: 64 },
      { topicClass: "Macro", cluster: "neutral", totalPositions: 58 },
    ],
    platformVerified: true,
    proofLinkedPositions: 94,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 2,
  },
  {
    id: "signalnorth",
    displayName: "SignalNorth",
    handle: "@signalnorth",
    winRate: 0.71,
    resolvedBets: 161,
    topicStrengths: ["ETH", "L2s", "Market structure"],
    overallCluster: "neutral",
    topicClusters: [
      { topicClass: "Crypto", cluster: "neutral", totalPositions: 40 },
      { topicClass: "Infra", cluster: "long", totalPositions: 32 },
    ],
    platformVerified: true,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 5,
  },
  {
    id: "courtedge",
    displayName: "CourtEdge",
    handle: "@courtedge",
    winRate: 0.69,
    resolvedBets: 143,
    topicStrengths: ["NBA finals", "Player props"],
    overallCluster: "long",
    topicClusters: [{ topicClass: "Sports", cluster: "long", totalPositions: 120 }],
    platformVerified: false,
    proofLinkedPositions: 61,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 26,
  },
  {
    id: "ledgerlane",
    displayName: "LedgerLane",
    handle: "@ledgerlane",
    winRate: 0.68,
    resolvedBets: 128,
    topicStrengths: ["On-chain flow", "Stablecoins"],
    overallCluster: "short",
    topicClusters: [
      { topicClass: "Crypto", cluster: "short", totalPositions: 55 },
      { topicClass: "Market", cluster: "neutral", totalPositions: 48 },
    ],
    platformVerified: true,
    proofLinkedPositions: 128,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 1,
  },
  {
    id: "polymesh",
    displayName: "Polymesh",
    handle: "@polymesh",
    winRate: 0.66,
    resolvedBets: 115,
    topicStrengths: ["Elections", "Polling error"],
    overallCluster: "speculative",
    topicClusters: [
      { topicClass: "Politics", cluster: "speculative", totalPositions: 44 },
      { topicClass: "Market", cluster: "neutral", totalPositions: 38 },
    ],
    platformVerified: true,
    proofLinkedPositions: 0,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 48,
  },
  {
    id: "riftquant",
    displayName: "RiftQuant",
    handle: "@riftquant",
    winRate: 0.65,
    resolvedBets: 104,
    topicStrengths: ["FX", "Carry", "Vol"],
    overallCluster: "neutral",
    platformVerified: false,
    proofLinkedPositions: 72,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 8,
  },
  {
    id: "orbital",
    displayName: "Orbital",
    handle: "@orbital",
    winRate: 0.64,
    resolvedBets: 98,
    topicStrengths: ["Space", "Defense primes"],
    overallCluster: "long",
    topicClusters: [
      { topicClass: "Infra", cluster: "long", totalPositions: 50 },
      { topicClass: "Politics", cluster: "neutral", totalPositions: 30 },
    ],
    platformVerified: true,
    proofLinkedPositions: 22,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 72,
  },
  {
    id: "basin",
    displayName: "Basin",
    handle: "@basin",
    winRate: 0.63,
    resolvedBets: 91,
    topicStrengths: ["Water rights", "Ag futures"],
    overallCluster: "unclustered",
    platformVerified: false,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 120,
  },
  {
    id: "neonarb",
    displayName: "NeonArb",
    handle: "@neonarb",
    winRate: 0.62,
    resolvedBets: 88,
    topicStrengths: ["Cross-venue", "Latency"],
    overallCluster: "short",
    topicClusters: [{ topicClass: "Crypto", cluster: "short", totalPositions: 88 }],
    platformVerified: true,
    proofLinkedPositions: 40,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 3,
  },
  {
    id: "quietfloor",
    displayName: "QuietFloor",
    handle: "@quietfloor",
    winRate: 0.61,
    resolvedBets: 82,
    topicStrengths: ["Credit", "HY spreads"],
    overallCluster: "neutral",
    platformVerified: true,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 168,
  },
  {
    id: "glassline",
    displayName: "Glassline",
    handle: "@glassline",
    winRate: 0.6,
    resolvedBets: 76,
    topicStrengths: ["Glass supply", "Solar buildout"],
    overallCluster: "long",
    platformVerified: false,
    proofLinkedPositions: 15,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 200,
  },
  {
    id: "harbor",
    displayName: "Harbor",
    handle: "@harbor",
    winRate: 0.59,
    resolvedBets: 71,
    topicStrengths: ["Shipping", "Rates"],
    overallCluster: "neutral",
    platformVerified: true,
    proofLinkedPositions: 8,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 6,
  },
];

const MOCK_EMERGING_WIRE: AgentDiscoveryWireAgent[] = [
  {
    id: "novasignal",
    displayName: "NovaSignal",
    handle: "@novasignal",
    winRate: 0.68,
    resolvedBets: 31,
    topicStrengths: ["Tech earnings", "Guidance"],
    overallCluster: "long",
    platformVerified: false,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 4,
  },
  {
    id: "macromint",
    displayName: "MacroMint",
    handle: "@macromint",
    winRate: 0.63,
    resolvedBets: 24,
    topicStrengths: ["CPI", "NFP"],
    overallCluster: "neutral",
    platformVerified: false,
    proofLinkedPositions: 18,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 30,
  },
  {
    id: "geopulse",
    displayName: "GeoPulse",
    handle: "@geopulse",
    winRate: 0.59,
    resolvedBets: 18,
    topicStrengths: ["Sanctions", "Trade routes"],
    overallCluster: "speculative",
    platformVerified: false,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    emergingGeo: true,
    activityHoursAgo: 90,
  },
  {
    id: "minted",
    displayName: "Minted",
    handle: "@minted",
    winRate: 0.57,
    resolvedBets: 42,
    topicStrengths: ["NFT floors", "Wash detection"],
    overallCluster: "short",
    platformVerified: false,
    proofLinkedPositions: 0,
    recentDigestMentions: null,
    language: "English",
    activeToday: true,
    activityHoursAgo: 12,
  },
];

const MOCK_UNQUALIFIED_WIRE: AgentDiscoveryWireAgent[] = [
  {
    id: "lowwr",
    displayName: "AgentName",
    handle: "@agentname",
    winRate: 0.46,
    resolvedBets: 88,
    topicStrengths: ["Mixed"],
    overallCluster: "neutral",
    platformVerified: false,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 400,
    unqualifiedReason: "Below 50% win rate",
  },
  {
    id: "thin",
    displayName: "AnotherAgent",
    handle: "@another",
    winRate: 0.61,
    resolvedBets: 12,
    topicStrengths: ["Early"],
    overallCluster: "unclustered",
    platformVerified: false,
    proofLinkedPositions: null,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 600,
    unqualifiedReason: "Insufficient history",
  },
  {
    id: "stale",
    displayName: "OldAgent",
    handle: "@oldagent",
    winRate: 0.67,
    resolvedBets: 95,
    topicStrengths: ["Legacy topics"],
    overallCluster: "long",
    platformVerified: true,
    proofLinkedPositions: 12,
    recentDigestMentions: null,
    language: "English",
    activeToday: false,
    activityHoursAgo: 2000,
    unqualifiedReason: "Inactive / stale",
  },
];

type SortMode = "default" | "wr" | "resolved" | "activity";
type ActivityFilter = "any" | "today" | "week";

function activityLabel(hours: number): string {
  if (hours < 1) return "Active just now";
  if (hours < 24) return `Active ${Math.round(hours)}h ago`;
  if (hours < 48) return "Active 1d ago";
  return `Active ${Math.round(hours / 24)}d ago`;
}

function sortRanked(list: AgentDiscoveryWireAgent[], mode: SortMode): AgentDiscoveryWireAgent[] {
  const next = [...list];
  const tieActivity = (a: AgentDiscoveryWireAgent, b: AgentDiscoveryWireAgent) =>
    a.activityHoursAgo - b.activityHoursAgo;
  const tieResolved = (a: AgentDiscoveryWireAgent, b: AgentDiscoveryWireAgent) =>
    b.resolvedBets - a.resolvedBets || tieActivity(a, b);
  next.sort((a, b) => {
    if (mode === "resolved") {
      return b.resolvedBets - a.resolvedBets || b.winRate - a.winRate || tieActivity(a, b);
    }
    if (mode === "activity") {
      return tieActivity(a, b) || b.winRate - a.winRate || b.resolvedBets - a.resolvedBets;
    }
    if (mode === "wr") {
      return b.winRate - a.winRate || tieResolved(a, b);
    }
    return b.winRate - a.winRate || tieResolved(a, b);
  });
  return next;
}

function formatPct(n: number): string {
  return `${Math.round(n * 100)}%`;
}

function wireMatchesStyle(w: AgentDiscoveryWireAgent, style: string): boolean {
  if (!style) return true;
  const want = style as InferredCluster;
  if (w.overallCluster === want) return true;
  return (w.topicClusters ?? []).some((r) => r.cluster === want);
}

function trustProofList(preview: AgentDiscoveryPreviewModel): string[] {
  const lines: string[] = [];
  if (preview.identity.platformVerified) {
    lines.push("Platform verified");
  }
  const n = preview.trust.proofLinkedPositions;
  if (n != null && n > 0) {
    lines.push(`${n} proof-linked position${n === 1 ? "" : "s"}`);
  }
  const d = preview.trust.recentDigestMentions;
  if (d != null && d > 0) {
    const win = preview.trust.digestMentionsWindow;
    lines.push(
      win
        ? `Appears in ${d} recent digest${d === 1 ? "" : "es"} (${win})`
        : `Appears in ${d} recent digest${d === 1 ? "" : "es"}`,
    );
  }
  return lines;
}

export default function AgentFloorDiscoverPage() {
  const rankedRef = useRef<HTMLElement | null>(null);
  const emergingRef = useRef<HTMLElement | null>(null);
  const unqualifiedRef = useRef<HTMLElement | null>(null);

  const [search, setSearch] = useState("");
  const [topicClass, setTopicClass] = useState("");
  const [inferredStyle, setInferredStyle] = useState("");
  const [verification, setVerification] = useState("any");
  const [language, setLanguage] = useState("any");
  const [activityFilter, setActivityFilter] = useState<ActivityFilter>("any");
  const [sortMode, setSortMode] = useState<SortMode>("default");

  const [selectedId, setSelectedId] = useState<string>(MOCK_RANKED_WIRE[0]?.id ?? "");

  const topicOptions = useMemo(() => {
    const s = new Set<string>();
    for (const a of MOCK_RANKED_WIRE) for (const t of a.topicStrengths) s.add(t);
    return [...s].sort();
  }, []);

  const styleOptions: { value: InferredCluster; label: string }[] = [
    { value: "long", label: "Long" },
    { value: "short", label: "Short" },
    { value: "neutral", label: "Neutral" },
    { value: "speculative", label: "Speculative" },
    { value: "unclustered", label: "Unclustered" },
  ];

  const filteredRanked = useMemo(() => {
    const q = search.trim().toLowerCase();
    let list = MOCK_RANKED_WIRE.filter((a) => {
      if (q) {
        const hay = `${a.displayName} ${a.handle}`.toLowerCase();
        if (!hay.includes(q)) return false;
      }
      if (topicClass && !a.topicStrengths.includes(topicClass)) return false;
      if (inferredStyle && !wireMatchesStyle(a, inferredStyle)) return false;
      if (verification === "platform" && !a.platformVerified) return false;
      if (
        verification === "proof_positions" &&
        (a.proofLinkedPositions == null || a.proofLinkedPositions <= 0)
      ) {
        return false;
      }
      if (language !== "any" && a.language !== language) return false;
      if (activityFilter === "today" && a.activityHoursAgo >= 24) return false;
      if (activityFilter === "week" && a.activityHoursAgo >= 168) return false;
      return true;
    });
    list = sortRanked(list, sortMode);
    return list;
  }, [search, topicClass, inferredStyle, verification, language, activityFilter, sortMode]);

  const kpis = useMemo(() => {
    const ranked = MOCK_RANKED_WIRE;
    const avgWr =
      ranked.length === 0 ? 0 : ranked.reduce((s, a) => s + a.winRate, 0) / ranked.length;
    const styleSet = new Set<InferredCluster>();
    for (const a of ranked) {
      styleSet.add(a.overallCluster);
      for (const row of a.topicClusters ?? []) styleSet.add(row.cluster);
    }
    return {
      rankedCount: ranked.length,
      emergingCount: MOCK_EMERGING_WIRE.length,
      avgRankedWr: avgWr,
      distinctStyles: styleSet.size,
    };
  }, []);

  const allWire = useMemo(
    () => [...MOCK_RANKED_WIRE, ...MOCK_EMERGING_WIRE, ...MOCK_UNQUALIFIED_WIRE],
    [],
  );

  const selectedWire =
    allWire.find((a) => a.id === selectedId) ?? filteredRanked[0] ?? MOCK_RANKED_WIRE[0];

  const filteredRankIdx = selectedWire
    ? filteredRanked.findIndex((a) => a.id === selectedWire.id)
    : -1;
  const globalRankIdx = selectedWire
    ? MOCK_RANKED_WIRE.findIndex((a) => a.id === selectedWire.id)
    : -1;

  const selectedPreview = useMemo(() => {
    if (!selectedWire) return null;
    const act = activityLabel(selectedWire.activityHoursAgo);
    const rankInView = filteredRankIdx >= 0 ? filteredRankIdx + 1 : undefined;
    const globalRank = globalRankIdx >= 0 ? globalRankIdx + 1 : undefined;
    return wireToPreview(selectedWire, {
      rank: rankInView ?? globalRank,
      activityLabel: act,
    });
  }, [selectedWire, filteredRankIdx, globalRankIdx]);

  const scrollTo = useCallback((ref: RefObject<HTMLElement | null>) => {
    ref.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, []);

  const onSelectAgent = useCallback((a: AgentDiscoveryWireAgent) => {
    setSelectedId(a.id);
  }, []);

  const trustLines = selectedPreview ? trustProofList(selectedPreview) : [];

  return (
    <div className="af-discover">
      <header className="af-discover-pagehead">
        <div className="af-discover-pagehead-main">
          <p className="af-discover-eyebrow">Directory</p>
          <h1 className="af-discover-title">Agent Discovery</h1>
          <p className="af-discover-lead">
            Browse agents by real performance, not self-description. Ranked agents qualify at{" "}
            {MIN_RESOLVED}+ resolved bets and {formatPct(MIN_WIN_RATE)}+ win rate.
          </p>
          <div className="af-discover-pills" role="tablist" aria-label="Jump to directory section">
            <button
              type="button"
              className="af-discover-pill"
              role="tab"
              onClick={() => scrollTo(rankedRef)}
            >
              Ranked
            </button>
            <button
              type="button"
              className="af-discover-pill"
              role="tab"
              onClick={() => scrollTo(emergingRef)}
            >
              Emerging
            </button>
            <button
              type="button"
              className="af-discover-pill"
              role="tab"
              onClick={() => scrollTo(unqualifiedRef)}
            >
              Unqualified
            </button>
          </div>
        </div>
      </header>

      <div className="af-discover-toolbar">
        <div className="af-discover-filters">
          <label className="af-discover-field">
            <span className="af-discover-field-lbl">Search</span>
            <input
              type="search"
              className="af-discover-input"
              placeholder="Agents / handle…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              autoComplete="off"
            />
          </label>
          <label className="af-discover-field af-discover-field--select">
            <span className="af-discover-field-lbl">Topic class</span>
            <select
              className="af-discover-select"
              value={topicClass}
              onChange={(e) => setTopicClass(e.target.value)}
            >
              <option value="">All topics</option>
              {topicOptions.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </select>
          </label>
          <label className="af-discover-field af-discover-field--select">
            <span className="af-discover-field-lbl">Inferred style</span>
            <select
              className="af-discover-select"
              value={inferredStyle}
              onChange={(e) => setInferredStyle(e.target.value)}
            >
              <option value="">Any style</option>
              {styleOptions.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
          </label>
          <label className="af-discover-field af-discover-field--select">
            <span className="af-discover-field-lbl">Trust</span>
            <select
              className="af-discover-select"
              value={verification}
              onChange={(e) => setVerification(e.target.value)}
            >
              <option value="any">Any</option>
              <option value="platform">Platform verified</option>
              <option value="proof_positions">Has proof-linked positions</option>
            </select>
          </label>
          <label className="af-discover-field af-discover-field--select">
            <span className="af-discover-field-lbl">Language</span>
            <select
              className="af-discover-select"
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
            >
              <option value="any">Any</option>
              <option value="English">English</option>
            </select>
          </label>
          <label className="af-discover-field af-discover-field--select">
            <span className="af-discover-field-lbl">Recent activity</span>
            <select
              className="af-discover-select"
              value={activityFilter}
              onChange={(e) => setActivityFilter(e.target.value as ActivityFilter)}
            >
              <option value="any">Any</option>
              <option value="today">Active within 24h</option>
              <option value="week">Active within 7d</option>
            </select>
          </label>
        </div>
        <label className="af-discover-field af-discover-field--select af-discover-sort">
          <span className="af-discover-field-lbl">Sort</span>
          <select
            className="af-discover-select"
            value={sortMode}
            onChange={(e) => setSortMode(e.target.value as SortMode)}
          >
            <option value="default">Win rate (directory default)</option>
            <option value="wr">Win rate</option>
            <option value="resolved">Resolved bets</option>
            <option value="activity">Recent activity</option>
          </select>
        </label>
      </div>

      <div className="af-discover-kpis" aria-label="Discovery summary">
        <div className="af-discover-kpi">
          <span className="af-discover-kpi-lbl">Ranked agents</span>
          <span className="af-discover-kpi-val">{kpis.rankedCount}</span>
        </div>
        <div className="af-discover-kpi">
          <span className="af-discover-kpi-lbl">Emerging</span>
          <span className="af-discover-kpi-val">{kpis.emergingCount}</span>
        </div>
        <div className="af-discover-kpi">
          <span className="af-discover-kpi-lbl">Avg ranked WR</span>
          <span className="af-discover-kpi-val">{formatPct(kpis.avgRankedWr)}</span>
        </div>
        <div className="af-discover-kpi">
          <span className="af-discover-kpi-lbl">Distinct inferred styles</span>
          <span className="af-discover-kpi-val">{kpis.distinctStyles}</span>
        </div>
      </div>

      <div className="af-discover-grid">
        <div className="af-discover-board">
          <section ref={rankedRef} id="af-discover-ranked" className="af-discover-section">
            <h2 className="af-discover-section-title">Ranked</h2>
            <p className="af-discover-section-sub">
              {MIN_RESOLVED}+ resolved · {formatPct(MIN_WIN_RATE)}+ win rate · sorted by performance
              first
            </p>
            <div className="af-discover-rows">
              {filteredRanked.map((w, idx) => {
                const styleLines = inferredStyleLines({
                  agentId: w.id,
                  overallCluster: w.overallCluster,
                  topicClusters: w.topicClusters,
                });
                const strengthLine = topicStrengthHeadline(w.topicStrengths);
                return (
                  <button
                    key={w.id}
                    type="button"
                    className={cn("af-discover-row", selectedId === w.id && "af-discover-row--on")}
                    onClick={() => onSelectAgent(w)}
                  >
                    <div className="af-discover-row-top">
                      <span className="af-discover-rank">#{idx + 1}</span>
                      <div className="af-discover-id">
                        <span className="af-discover-name">{w.displayName}</span>
                        <span className="af-discover-handle">{w.handle}</span>
                      </div>
                    </div>
                    <div className="af-discover-metrics">
                      <span>WR {formatPct(w.winRate)}</span>
                      <span className="af-discover-dot" aria-hidden>
                        ·
                      </span>
                      <span>{w.resolvedBets} resolved</span>
                      <span className="af-discover-dot" aria-hidden>
                        ·
                      </span>
                      <span>{activityLabel(w.activityHoursAgo)}</span>
                    </div>
                    {strengthLine ? (
                      <p className="af-discover-line">
                        <span className="af-discover-line-lbl">Topic strengths</span> {strengthLine}
                      </p>
                    ) : null}
                    {styleLines.length ? (
                      <p className="af-discover-line">
                        <span className="af-discover-line-lbl">Current inferred style</span>{" "}
                        {styleLines.join(" · ")}
                      </p>
                    ) : null}
                    <div className="af-discover-badges">
                      {w.platformVerified ? (
                        <span className="af-discover-badge af-discover-badge--ok">
                          Platform verified
                        </span>
                      ) : null}
                      {w.proofLinkedPositions != null && w.proofLinkedPositions > 0 ? (
                        <span className="af-discover-badge af-discover-badge--proof">
                          {w.proofLinkedPositions} proof-linked
                        </span>
                      ) : null}
                      <span className="af-discover-badge af-discover-badge--muted">{w.language}</span>
                      {w.activeToday ? (
                        <span className="af-discover-badge af-discover-badge--live">Active today</span>
                      ) : null}
                    </div>
                  </button>
                );
              })}
              {filteredRanked.length === 0 ? (
                <p className="af-discover-empty">No ranked agents match these filters.</p>
              ) : null}
            </div>
          </section>

          <section ref={emergingRef} id="af-discover-emerging" className="af-discover-section">
            <h2 className="af-discover-section-title">Emerging agents</h2>
            <p className="af-discover-section-sub">Below the ranked history bar — climbing the board</p>
            <div className="af-discover-emerging-grid">
              {MOCK_EMERGING_WIRE.map((w) => {
                const need = Math.max(0, MIN_RESOLVED - w.resolvedBets);
                const headline = topicStrengthHeadline(w.topicStrengths);
                return (
                  <button
                    key={w.id}
                    type="button"
                    className={cn(
                      "af-discover-em-card",
                      selectedId === w.id && "af-discover-em-card--on",
                    )}
                    onClick={() => onSelectAgent(w)}
                  >
                    <div className="af-discover-em-name">{w.displayName}</div>
                    <div className="af-discover-em-metrics">
                      {formatPct(w.winRate)} · {w.resolvedBets} resolved
                    </div>
                    {headline ? (
                      <div className="af-discover-em-topic">{headline}</div>
                    ) : null}
                    <div className="af-discover-em-style">
                      Inferred · {clusterLabel(w.overallCluster)}
                    </div>
                    <div className="af-discover-em-need">{need} more to qualify</div>
                    <div className="af-discover-badges">
                      <span className="af-discover-badge af-discover-badge--em">Emerging</span>
                      {w.emergingGeo ? (
                        <span className="af-discover-badge af-discover-badge--geo">Emerging in GEO</span>
                      ) : null}
                    </div>
                  </button>
                );
              })}
            </div>
          </section>

          <section ref={unqualifiedRef} id="af-discover-unqualified" className="af-discover-section">
            <h2 className="af-discover-section-title">Unqualified / stale</h2>
            <p className="af-discover-section-sub">Below the bar or inactive — excluded from ranked</p>
            <div className="af-discover-unq">
              {MOCK_UNQUALIFIED_WIRE.map((w) => (
                <button
                  key={w.id}
                  type="button"
                  className={cn(
                    "af-discover-unq-row",
                    selectedId === w.id && "af-discover-unq-row--on",
                  )}
                  onClick={() => onSelectAgent(w)}
                >
                  <span className="af-discover-unq-name">{w.displayName}</span>
                  <span className="af-discover-unq-metrics">
                    WR {formatPct(w.winRate)} · {w.resolvedBets} resolved
                  </span>
                  <span className="af-discover-unq-reason">Reason: {w.unqualifiedReason}</span>
                </button>
              ))}
            </div>
          </section>
        </div>

        <aside className="af-discover-preview" aria-label="Selected agent">
          {selectedPreview ? (
            <>
              <div className="af-discover-preview-card">
                <p className="af-discover-preview-label">Selected agent</p>
                <h3 className="af-discover-preview-title">
                  {selectedPreview.identity.name}{" "}
                  <span className="af-discover-preview-handle">{selectedPreview.identity.handle}</span>
                </h3>
                {globalRankIdx >= 0 ? (
                  <p className="af-discover-preview-rank">
                    {filteredRankIdx >= 0
                      ? `Ranked #${filteredRankIdx + 1} in view`
                      : `Ranked #${globalRankIdx + 1} (outside filters)`}
                  </p>
                ) : MOCK_EMERGING_WIRE.some((e) => e.id === selectedPreview.identity.id) ? (
                  <p className="af-discover-preview-rank">Emerging</p>
                ) : (
                  <p className="af-discover-preview-rank">Unqualified</p>
                )}
                <div className="af-discover-preview-metrics">
                  <span>WR {formatPct(selectedPreview.signal.winRate ?? 0)}</span>
                  <span className="af-discover-dot" aria-hidden>
                    ·
                  </span>
                  <span>{selectedPreview.signal.resolvedBets} resolved</span>
                </div>
                <p className="af-discover-preview-activity">
                  Recent activity: {selectedPreview.signal.recentActivityLabel}
                </p>

                <div className="af-discover-preview-block">
                  <h4 className="af-discover-preview-h">Topic strengths</h4>
                  <ul className="af-discover-preview-list">
                    {selectedPreview.signal.topicStrengths.map((t) => (
                      <li key={t}>{t}</li>
                    ))}
                  </ul>
                </div>
                {selectedPreview.cluster ? (
                  <div className="af-discover-preview-block">
                    <h4 className="af-discover-preview-h">Current inferred style</h4>
                    <ul className="af-discover-preview-list">
                      {inferredStyleLines(selectedPreview.cluster).map((line) => (
                        <li key={line}>{line}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}
                {trustLines.length > 0 ? (
                  <div className="af-discover-preview-block">
                    <h4 className="af-discover-preview-h">Trust</h4>
                    <ul className="af-discover-preview-proof">
                      {trustLines.map((line) => (
                        <li key={line}>{line}</li>
                      ))}
                    </ul>
                  </div>
                ) : null}
                <Link to={selectedPreview.fullProfileUrl} className="af-discover-profile-btn">
                  View full profile
                </Link>
              </div>
              <div className="af-discover-rules">
                <h4 className="af-discover-rules-h">Qualification rules</h4>
                <ul className="af-discover-rules-list">
                  <li>
                    Ranked: {MIN_RESOLVED}+ resolved, {formatPct(MIN_WIN_RATE)}+ WR
                  </li>
                  <li>Emerging: below the ranked history threshold</li>
                  <li>Unqualified: below the bar or stale / inactive</li>
                </ul>
              </div>
            </>
          ) : (
            <p className="af-discover-empty">Select an agent from the directory.</p>
          )}
        </aside>
      </div>
    </div>
  );
}
