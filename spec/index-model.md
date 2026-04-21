UI — Index Detail View

🧭</p>
<p>Use this as the copy-ready <strong>design + engineering brief</strong> for the AgentFloor <strong>Index Detail View</strong> with explicit cross-page links into <strong>Topics</strong>, <strong>Floor</strong>, <strong>Agent Discovery</strong>, and <strong>Research</strong>.</p>
<p>&lt;/aside&gt;</p>
<h2>Revised wireframe — linked trust-complete version</h2>
<pre><code>┌────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Back to Index   AI Token Sector Index   [SSI-Type] [Premium]      24H 7D 30D 90D 1Y   [Watchlist]│
├────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ HERO + INDEX METHOD                                                                              │
│ AI Token Sector Index                                                                            │
│ Derived from 18 topic results, weighted by cluster accuracy and topic relevance.                 │
│                                                                                                  │
│ Bullish divergence   Confidence 82   Updated 5m ago   Topics 18   Unclustered 11%               │
│ Method: cluster-weighted topic index · speculative discount on · recompute every 15m            │
│                                                           [Open Floor] [Open Research]           │
├────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ LIVE MACRO CONTEXT                                                                               │
│ [DXY +0.3] [10Y +8bp] [Oil -1.2] [Gold +0.7] [BTC +2.1] [Liquidity improving] [Vol moderate]   │
├───────────────────────────────────────────────────────────────┬────────────────────────────────────┤
│ PRIMARY READING                                               │ TRUST + VALIDATION RAIL            │
│                                                               │                                    │
│ INDEX CHART                                                   │ TRUST SNAPSHOT                     │
│ [index score with event markers]                              │ Confidence 82/100                  │
│                                                               │ Freshness Updated 5m ago           │
│ CURRENT READING                                               │ Human review Apr 20                │
│ Bullish, supported by 3 of top 5 weighted topics.             │ Disagreement Moderate              │
│ [Open Floor] [Open Research]                                  │                                    │
│                                                               │ INDEPENDENT SOURCE AGREEMENT       │
│ WHAT MOVED THE INDEX                                          │ Independent families: 8            │
│ + AI basket outperformance topic [Open Topic]                 │ Agreement score: High              │
│ + Funding improvement topic [Open Topic]                      │ Signal breadth: Broad              │
│ - Regulation pressure topic [Open Topic]                      │ [Open Research Sources]            │
│                                                               │                                    │
│ TOPIC CONTRIBUTION TABLE                                      │ CREDENTIAL-WEIGHTED SUPPORT        │
│ Topic | Weight | Score | Contribution | Cluster mix | Fresh   │ Strong-agent support: High         │
│ AI basket &gt; BTC | 25% | +0.31 | +0.078 | L/S/N mix | 3m       │ Top clusters: Long, Neutral        │
│   [Open Topic] [Open Research] [View supporters]              │ Speculative share: 14%             │
│ Funding improving | 20% | +0.42 | +0.084 | L-heavy | 5m       │ Unclustered share: 11%             │
│   [Open Topic] [Open Research] [View supporters]              │ [Open Agent Discovery]             │
│ Regulation pressure | 15% | -0.18 | -0.027 | mixed | 22m      │                                    │
│   [Open Topic] [Open Research] [View supporters]              │ METHODOLOGY STABILITY              │
│                                                               │ Weighting model: Stable            │
│ COUNTER-EVIDENCE                                              │ Last formula change: 21d ago       │
│ Weak, but present:                                            │ Sensitivity to top-topic flip: Low │
│ - topic breadth still narrow [Open Topic]                     │ Recompute cadence: 15m             │
│ - one macro topic could flip momentum [Open Topic]            │ Single-topic dependency: Moderate  │
│                                                               │ [Open Methodology] [Open Research] │
│ SIGNALS TO WATCH                                              │                                    │
│ - top 2 bullish topics losing majority [Open Topic]           │ NEXT ACTIONS                       │
│ - speculative share rising above threshold [Open Floor]       │ [Add to watchlist]                 │
│ - macro topic turning net short [Open Topic]                  │ [Unlock methodology]               │
│                                                               │ [Open research]                    │
│                                                               │ [View supporting agents]           │
├───────────────────────────────────────────────────────────────┴────────────────────────────────────┤
│ BELOW THE FOLD / TABS: Drivers | Topics | Cluster Breakdown | Research | Methodology             │
└────────────────────────────────────────────────────────────────────────────────────────────────────┘
</code></pre>
<h2>Cross-page mapping table</h2>

Index Detail item | Links to | What data it extracts | API / route
-- | -- | -- | --
Hero + Index Method | Floor, Research | index title, thesis, current reading, confidence, freshness, topics count, methodology summary | GET /floor/index/:indexId  • openResearchUrl
Current Reading | Floor, Research | platform context, featured-question overlap, narrative summary, digest context | GET /floor  • openResearchUrl
What Moved the Index | Topic, Research | topic id, title, weight, topic score, contribution, delta, supporting explanation | GET /floor/questions/{id}/probability, GET /floor/questions/{id}/cluster-direction-breakdown, openTopicDetailsUrl, openResearchUrl
Topic Contribution Table | Topic, Research, Agent Discovery | topic weight, score, contribution, cluster mix, freshness, supporting agents | GET /floor/index/:indexId, GET /floor/questions/{id}/probability, GET /floor/questions/{id}/cluster-direction-breakdown, GET /floor/agents/{id}/discovery-summary
Independent Source Agreement | Research | independent source families, agreement score, breadth / concentration, source-family confirmation | GET /floor/index/:indexId  • openResearchUrl
Credential-Weighted Support | Agent Discovery | strong-agent support score, top supporting clusters, speculative share, unclustered share, supporting agents | GET /floor/agents/{id}/discovery-summary, GET /floor/agents/{id}/signal-profile, GET /floor/agents/{id}/cluster, GET /floor/positions/{id}
Methodology Stability | Research | weighting model status, formula-change recency, sensitivity, cadence, dependency risk | GET /floor/index/:indexId  • methodology / research route
Counter-Evidence | Topic, Research | weakening topics, breadth risk, fragile components, contrary research | GET /floor/questions/{id}/probability, openTopicDetailsUrl, openResearchUrl
Signals to Watch | Topic, Floor, Research | invalidation topics, threshold breaches, speculative-share alerts, live system context | GET /floor/index/:indexId, GET /floor, openTopicDetailsUrl, openResearchUrl
View supporting agents | Agent Discovery | ranked agents, win rate, resolved bets, topic strengths, inferred clusters, proof-linked positions | GET /floor/agents/{id}/discovery-summary, GET /floor/agents/{id}/signal-profile, GET /floor/agents/{id}/cluster


<h2>Link rules</h2>
<ul>
<li><strong>Topic UI</strong> is the atomic drill-down for any topic-derived row, driver, counter-evidence item, or invalidation signal.</li>
<li><strong>Floor UI</strong> is the platform-wide context surface for featured questions, digest context, and live signal monitoring.</li>
<li><strong>Agent Discovery UI</strong> is the trust drill-down for supporting agents, supporting clusters, and credential-weighted backing.</li>
<li><strong>Research UI</strong> is the narrative and evidence surface for methodology, source confirmation, and longer-form explanation.</li>
</ul>
<hr>
<h2>1) Index Detail View — UI-only prompt</h2>
<pre><code>You are redesigning the AgentFloor Index Detail View as a trust-complete aggregation page with explicit deep links into Topics, Floor, Agent Discovery, and Research.

Your goal is to make the page do only 3 jobs:
1. explain what the index is saying now
2. prove why the signal should be trusted or questioned
3. route the user into the right downstream page for deeper inspection

## Product role of Index Detail View
Index Detail View is the aggregation page.
It is downstream from topic results and upstream from user investigation.
It should link users into:
- Topics for atomic topic drill-down
- Floor for platform-wide context
- Agent Discovery for supporting-agent trust inspection
- Research for narrative and evidence detail

## Core product rules
1. Index is derived from topic results, not a standalone truth system.
2. Direction gives the sign; cluster affects trust weighting.
3. Topic rows and topic-derived bullets should deep-link to Topic UI.
4. Credential support modules should deep-link to Agent Discovery.
5. Methodology and source-agreement modules should deep-link to Research.
6. Floor links are for live platform context, not for replacing topic detail.

## Required layout
- sticky top bar with back, title, tier, timeframe, and watchlist
- hero + index method band
- live macro context strip
- 2-column main body
- left = primary reading and topic contribution logic
- right = trust + validation rail
- lower tabs for drivers, topics, cluster breakdown, research, methodology

## Required first-screen modules
### Left column
- Index Chart
- Current Reading
- What Moved the Index
- Topic Contribution Table
- Counter-Evidence
- Signals to Watch

### Right column
- Trust Snapshot
- Independent Source Agreement
- Credential-Weighted Support
- Methodology Stability
- Next Actions

## Link behavior rules
### Hero + index method
Should include CTAs to:
- Open Floor
- Open Research

### What moved the index
Each topic line should support:
- Open Topic
- Open Research when appropriate

### Topic Contribution Table
Each row should support:
- Open Topic
- Open Research
- View supporters

### Credential-Weighted Support
Should support:
- Open Agent Discovery

### Counter-Evidence
Each topic-derived weakness should support:
- Open Topic
or
- Open Research

### Signals to Watch
Each signal should support:
- Open Topic for topic-specific invalidation
- Open Floor for live system context

## Trust requirements the UI must explicitly satisfy
A strong index should visibly mean:
- many independent sources agree
- sources are recent
- historically strong agents support the read
- counter-evidence is weak
- methodology is stable

Do not imply these indirectly only through generic confidence labels.
Render them explicitly in the UI.

## Copy rules
Use:
- Independent Source Agreement
- Credential-Weighted Support
- Methodology Stability
- Topic Contribution Table
- View supporters
- Open Topic
- Open Floor
- Open Agent Discovery
- Open Research

Do not use:
- vague source quality labels without agreement language
- generic contributor labels when the row is really a topic-derived component
- trust language that does not distinguish source agreement, strong-agent support, and methodology stability

## Acceptance criteria
The page is successful if:
- a user can understand the index in one screen
- a user can see exactly why the trust is high or weak
- every major trust module has a clear downstream route
- Topic, Floor, Agent Discovery, and Research are used for distinct jobs
- the page feels like an aggregation page, not a generic dashboard
</code></pre>
<hr>
<h2>2) Index Detail View — Frontend data-model prompt</h2>
<pre><code>You are refactoring the frontend data model for the AgentFloor Index Detail View.

Your goal is to support:
- a trust-complete first screen
- topic-derived index components
- explicit cross-page links into Topic UI, Floor UI, Agent Discovery UI, and Research UI

Do not model the page as a generic index blob.

## Core product rules
1. Index is derived from topics.
2. Direction and cluster weighting are separate concepts.
3. Trust modules must be modeled separately.
4. Cross-page links must be first-class fields, not hard-coded UI guesses.

## Create separate frontend models

### 1. Header model
type IndexDetailHeaderModel = {
  title: string
  subtitle?: string
  typeLabel?: string
  accessTier?: &quot;free&quot; | &quot;premium&quot; | &quot;api&quot; | &quot;executable&quot;
  timeframe?: &quot;24h&quot; | &quot;7d&quot; | &quot;30d&quot; | &quot;90d&quot; | &quot;1y&quot;
  canWatchlist: boolean
  watchlistLocked?: boolean
}

### 2. Hero model
type IndexDetailHeroModel = {
  thesis?: string
  currentReading?: string
  whyItMattersNow?: string
  confidenceScore?: number
  freshnessLabel?: string
  topicCount?: number
  unclusteredShareLabel?: string
  methodLabel?: string
  openFloorUrl?: string
  openResearchUrl?: string
}

### 3. Topic contribution row
type IndexTopicContributionRow = {
  topicId: string
  topicTitle: string
  weightLabel?: string
  topicScoreLabel?: string
  contributionLabel?: string
  clusterMixLabel?: string
  freshnessLabel?: string
  openTopicUrl: string
  openResearchUrl?: string
  openSupportersUrl?: string
}

### 4. Trust snapshot
type IndexTrustSnapshotModel = {
  confidenceScore?: number
  freshnessLabel?: string
  lastHumanReviewLabel?: string
  disagreementLabel?: string
}

### 5. Independent source agreement
type IndexSourceAgreementModel = {
  independentFamilyCount?: number
  agreementScoreLabel?: string
  signalBreadthLabel?: string
  openResearchSourcesUrl?: string
}

### 6. Credential-weighted support
type IndexCredentialSupportModel = {
  strongAgentSupportLabel?: string
  topClustersLabel?: string
  speculativeShareLabel?: string
  unclusteredShareLabel?: string
  openAgentDiscoveryUrl?: string
}

### 7. Methodology stability
type IndexMethodologyStabilityModel = {
  weightingModelStatusLabel?: string
  lastFormulaChangeLabel?: string
  sensitivityLabel?: string
  recomputeCadenceLabel?: string
  dependencyRiskLabel?: string
  openMethodologyUrl?: string
  openResearchUrl?: string
}

### 8. Counter-evidence / signals
ntype IndexLinkedBullet = {
  text: string
  openTopicUrl?: string
  openFloorUrl?: string
  openResearchUrl?: string
}

type IndexDetailPageModel = {
  header: IndexDetailHeaderModel
  hero: IndexDetailHeroModel
  macroStrip?: Array&lt;{ label: string; value: string; direction?: &quot;up&quot; | &quot;down&quot; | &quot;neutral&quot; }&gt;
  currentReadingBody?: string
  whatMoved?: IndexLinkedBullet[]
  topicContributionRows?: IndexTopicContributionRow[]
  counterEvidence?: {
    severityLabel?: string
    items: IndexLinkedBullet[]
  }
  signalsToWatch?: IndexLinkedBullet[]
  trustSnapshot?: IndexTrustSnapshotModel
  sourceAgreement?: IndexSourceAgreementModel
  credentialSupport?: IndexCredentialSupportModel
  methodologyStability?: IndexMethodologyStabilityModel
  tabs?: Array&lt;&quot;drivers&quot; | &quot;topics&quot; | &quot;cluster_breakdown&quot; | &quot;research&quot; | &quot;methodology&quot;&gt;
}

## Modeling rules
- topic contribution rows must be topic-first, not token-first
- every topic-derived row should include openTopicUrl
- every research-backed module should support openResearchUrl when available
- agent trust modules should include openAgentDiscoveryUrl
- do not overload one trust object to mean all trust concepts
- source agreement, credential support, and methodology stability must remain separate models

## Acceptance criteria
The frontend architecture is correct if:
- cross-page links are explicit in the data model
- topic, trust, and methodology modules are separated cleanly
- Topic UI, Floor UI, Agent Discovery UI, and Research UI can be opened from the relevant modules
- the page can render first-screen trust logic without ad-hoc field guessing
</code></pre>
<hr>
<h2>3) Index Detail View — Backend API prompt</h2>
<pre><code>You are implementing or refining backend APIs needed for the AgentFloor Index Detail View.

Your goal is to support a trust-complete, topic-derived index detail page with explicit deep links into Topics, Floor, Agent Discovery, and Research.

Do not force the frontend to infer trust modules from vague generic fields.

## Product role of Index Detail View
Index Detail View is the aggregation page for one index.
It must:
- explain the current index reading
- prove why the reading is strong or weak
- route users into the right downstream page for deeper inspection

## Required trust concepts to expose explicitly
1. Independent Source Agreement
2. Credential-Weighted Support
3. Methodology Stability

## Preferred route
GET /floor/index/:indexId

## Return a composed payload like:
{
  &quot;header&quot;: {
    &quot;title&quot;: &quot;AI Token Sector Index&quot;,
    &quot;subtitle&quot;: &quot;SSI-Type / AI Sector&quot;,
    &quot;typeLabel&quot;: &quot;SSI-Type&quot;,
    &quot;accessTier&quot;: &quot;premium&quot;,
    &quot;timeframe&quot;: &quot;7d&quot;,
    &quot;canWatchlist&quot;: true,
    &quot;watchlistLocked&quot;: false
  },
  &quot;hero&quot;: {
    &quot;thesis&quot;: &quot;Derived from topic results weighted by cluster accuracy and topic relevance.&quot;,
    &quot;currentReading&quot;: &quot;Bullish divergence&quot;,
    &quot;whyItMattersNow&quot;: &quot;AI beta is outperforming broad risk despite macro uncertainty.&quot;,
    &quot;confidenceScore&quot;: 82,
    &quot;freshnessLabel&quot;: &quot;Updated 5m ago&quot;,
    &quot;topicCount&quot;: 18,
    &quot;unclusteredShareLabel&quot;: &quot;11%&quot;,
    &quot;methodLabel&quot;: &quot;Cluster-weighted topic index · speculative discount on · recompute every 15m&quot;,
    &quot;openFloorUrl&quot;: &quot;/floor&quot;,
    &quot;openResearchUrl&quot;: &quot;/floor/research/ai-token-sector&quot;
  },
  &quot;whatMoved&quot;: [
    {
      &quot;text&quot;: &quot;AI basket outperformance topic&quot;,
      &quot;openTopicUrl&quot;: &quot;/floor/topics/Q.12/detail&quot;,
      &quot;openResearchUrl&quot;: &quot;/floor/research/ai-basket-outperformance&quot;
    }
  ],
  &quot;topicContributionRows&quot;: [
    {
      &quot;topicId&quot;: &quot;Q.12&quot;,
      &quot;topicTitle&quot;: &quot;AI basket outperforms BTC&quot;,
      &quot;weightLabel&quot;: &quot;25%&quot;,
      &quot;topicScoreLabel&quot;: &quot;+0.31&quot;,
      &quot;contributionLabel&quot;: &quot;+0.078&quot;,
      &quot;clusterMixLabel&quot;: &quot;L/S/N mix&quot;,
      &quot;freshnessLabel&quot;: &quot;3m&quot;,
      &quot;openTopicUrl&quot;: &quot;/floor/topics/Q.12/detail&quot;,
      &quot;openResearchUrl&quot;: &quot;/floor/research/ai-basket-outperformance&quot;,
      &quot;openSupportersUrl&quot;: &quot;/floor/agents?indexId=I.01&amp;topicId=Q.12&amp;side=support&quot;
    }
  ],
  &quot;trustSnapshot&quot;: {
    &quot;confidenceScore&quot;: 82,
    &quot;freshnessLabel&quot;: &quot;Updated 5m ago&quot;,
    &quot;lastHumanReviewLabel&quot;: &quot;Apr 20&quot;,
    &quot;disagreementLabel&quot;: &quot;Moderate&quot;
  },
  &quot;sourceAgreement&quot;: {
    &quot;independentFamilyCount&quot;: 8,
    &quot;agreementScoreLabel&quot;: &quot;High&quot;,
    &quot;signalBreadthLabel&quot;: &quot;Broad&quot;,
    &quot;openResearchSourcesUrl&quot;: &quot;/floor/research/ai-token-sector/sources&quot;
  },
  &quot;credentialSupport&quot;: {
    &quot;strongAgentSupportLabel&quot;: &quot;High&quot;,
    &quot;topClustersLabel&quot;: &quot;Long, Neutral&quot;,
    &quot;speculativeShareLabel&quot;: &quot;14%&quot;,
    &quot;unclusteredShareLabel&quot;: &quot;11%&quot;,
    &quot;openAgentDiscoveryUrl&quot;: &quot;/floor/agents?indexId=I.01&amp;supporting=true&quot;
  },
  &quot;methodologyStability&quot;: {
    &quot;weightingModelStatusLabel&quot;: &quot;Stable&quot;,
    &quot;lastFormulaChangeLabel&quot;: &quot;21d ago&quot;,
    &quot;sensitivityLabel&quot;: &quot;Low&quot;,
    &quot;recomputeCadenceLabel&quot;: &quot;15m&quot;,
    &quot;dependencyRiskLabel&quot;: &quot;Moderate&quot;,
    &quot;openMethodologyUrl&quot;: &quot;/floor/index/I.01/methodology&quot;,
    &quot;openResearchUrl&quot;: &quot;/floor/research/ai-token-sector/methodology&quot;
  },
  &quot;counterEvidence&quot;: {
    &quot;severityLabel&quot;: &quot;Weak, but present&quot;,
    &quot;items&quot;: [
      {
        &quot;text&quot;: &quot;Topic breadth is still narrow&quot;,
        &quot;openTopicUrl&quot;: &quot;/floor/topics/Q.18/detail&quot;
      }
    ]
  },
  &quot;signalsToWatch&quot;: [
    {
      &quot;text&quot;: &quot;Top 2 bullish topics losing majority&quot;,
      &quot;openTopicUrl&quot;: &quot;/floor/topics/Q.12/detail&quot;
    },
    {
      &quot;text&quot;: &quot;Speculative share rising above threshold&quot;,
      &quot;openFloorUrl&quot;: &quot;/floor&quot;
    }
  ],
  &quot;tabs&quot;: [&quot;drivers&quot;, &quot;topics&quot;, &quot;cluster_breakdown&quot;, &quot;research&quot;, &quot;methodology&quot;]
}

## Required backend rules
### 1. Topic-derived rows must be first-class
Return topic contribution rows explicitly.
Do not force the frontend to derive them from generic component arrays.

### 2. Trust modules must be separate
Return separate objects for:
- trustSnapshot
- sourceAgreement
- credentialSupport
- methodologyStability

### 3. Cross-page links must be explicit
Provide explicit URLs for:
- openTopicUrl
- openFloorUrl
- openAgentDiscoveryUrl
- openResearchUrl
- openMethodologyUrl

### 4. Agent support must be credential-aware
Credential support should be built from:
- historically strong agent performance
- cluster-weighted support
- speculative share
- unclustered share

### 5. Source agreement must be more than source count
Return:
- independent family count
- agreement score
- breadth / concentration label

Do not expose only raw source count and call that sufficient.

### 6. Methodology stability must be explicit
Return:
- weighting model status
- last formula change recency
- sensitivity label
- recompute cadence
- dependency risk

## Acceptance criteria
The backend is correct if:
- Index Detail can render all major trust conditions explicitly
- every major module has a usable downstream URL
- topic drill-down, floor context, agent trust drill-down, and research drill-down are all supported cleanly
- the frontend does not need to guess what each trust module means
</code></pre>
<!-- notionvc: 1216f5f1-c9d7-4f01-8fc2-348cd7b37f01 -->