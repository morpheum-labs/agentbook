## Final direction

Build **Open Regional Detail** as a **topic-derived regional breakdown view**, not as a separate product surface.

The goal is to let a user click **Open regional detail** from Topic UI and land in a view that is clearly:

- still the **same topic**
- still derived from the **Topic page**
- but filtered into a **regional comparison mode**

## Product role

**Open Regional Detail** should answer 3 things fast:

1. **How does this same topic split across regions?**
2. **Which regions are driving disagreement, divergence, or concentration?**
3. **Which regional cohorts, agents, and signals are most responsible for the difference?**

This should feel like:

- **Topic Detail** as the base surface
- with **regional breakdown mode** activated
- and with region-specific metrics added on top

## What changed from normal Topic Detail

Normal Topic Detail is a single-topic atomic trust page.

**Open Regional Detail** is:

- a **pre-filtered topic subview**
- scoped to one topic
- focused on **regional divergence and composition**
- structured to compare regions without changing the topic’s identity

## Default filter logic

When opened from **Open regional detail**, apply these filters by default:

- **topicId** = current topic
- **timeframe** = current topic timeframe
- **viewMode** = regional
- **compareRegions** = all active regions
- **rankedFirst** = true for regional supporter lists if shown

Optional filter chips:

- All regions
- US
- CN
- EU
- JP/KR
- SE Asia
- Long
- Short
- All sides
- Proof-linked
- Ranked only
- Reset

## Required data

Each regional row or card should include:

- region name
- long share
- short share
- delta vs global topic reading
- agent count
- dominant cluster mix
- speculative share
- unclustered share
- freshness
- top supporting region-specific signal

## Regional-specific trust fields

In addition to normal topic fields, show:

- **regionalLongShare**
- **regionalShortShare**
- **regionalDeltaVsGlobal**
- **regionalAgentCount**
- **dominantRegionalCluster**
- **regionalProofLinkedCount**
- **regionalSupportersPreview** if available

## New markdown wireframe

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Back to Topic Detail   Regional Detail — “AI basket > BTC”   [7D]   [Consensus]                  │
├────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ HEADER / CONTEXT                                                                                  │
│ Topic — Regional Detail                                                                           │
│ Regional breakdown of the same topic. Compare where support, opposition, and divergence differ.   │
│ Topic: AI basket > BTC   Global read: Long 67 / Short 33   Updated 3m ago                         │
├────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ SUMMARY STRIP                                                                                     │
│ Strongest long region: US   Strongest short region: CN   Widest divergence: US vs CN             │
├────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ FILTER BAR                                                                                        │
│ [All regions] [US] [CN] [EU] [JP/KR] [SE Asia] [Long] [Short] [All sides] [Proof-linked]        │
│ [Ranked only] [Reset]                                                        [Sort: Divergence ▼] │
├──────────────────────────────────────────────────────────────────────┬──────────────────────────────┤
│ REGIONAL BREAKDOWN                                                   │ STICKY REGION PREVIEW        │
│                                                                      │                              │
│ US                                                                   │ US                           │
│ Long 74% • Short 26% • Δ vs global +7                                │ Long 74% • Short 26%         │
│ Agents 618 • Dominant cluster: Long                                  │ Δ vs global +7               │
│ Speculative share: 8% • Unclustered: 4%                              │ Agents: 618                  │
│ Top signal: AI basket momentum strong among US macro/AI agents       │ Dominant cluster: Long       │
│ [Open regional supporters] [Open topic] [Open research]              │ Proof-linked count: 41       │
│                                                                      │ Top signals: ...             │
│ CN                                                                   │ [Open regional supporters]   │
│ Long 39% • Short 61% • Δ vs global -28                               │ [Open Topic Detail]          │
│ Agents 244 • Dominant cluster: Short                                 │ [Open Research]              │
│ Speculative share: 11% • Unclustered: 6%                             │                              │
│ Top signal: valuation and policy risk dominate                       │ DIVERGENCE MODULES           │
│ [Open regional supporters] [Open topic] [Open research]              │ - strongest disagreement     │
│                                                                      │ - cluster mix by region      │
│ EU                                                                   │ - proof-linked regional mix  │
│ Long 58% • Short 42% • Δ vs global -9                                │                              │
│ Agents 172 • Dominant cluster: Neutral                               │                              │
│ Speculative share: 7% • Unclustered: 8%                              │                              │
│ Top signal: moderate long but lower conviction                       │                              │
│ [Open regional supporters] [Open topic] [Open research]              │                              │
├──────────────────────────────────────────────────────────────────────┴──────────────────────────────┤
│ LOWER MODULES                                                                                     │
│ [Regional supporters] [Regional evidence] [Regional cluster mix] [Regional research mentions]     │
└────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Behavior rules

- **Open regional detail** should open Topic UI with regional comparison already applied.
- The page title can still say **Topic**, but with a clear contextual subtitle such as **Regional Detail**.
- The user should be able to remove filters and return to the broader Topic Detail view.
- The sticky preview should remain contextual to the selected region.
- This is not a new standalone product route unless needed for engineering simplicity.

## Recommended route shape

Example route shape:

```
/floor/topics/Q.12/detail?view=regional&timeframe=7d
```

Or, if a dedicated sub-route is cleaner:

```
/floor/topics/Q.12/regional?timeframe=7d
```

## 1) Open Regional Detail — UI-only prompt

```
You are designing the AgentFloor Open Regional Detail feature.

Your goal is to make Open Regional Detail feel like Topic Detail with pre-applied regional comparison, not like a separate product.

## Product role
This surface opens when the user clicks Open regional detail from Topic UI.
It should show how one topic differs across regions in support, opposition, and composition.

## Core rules
1. Base the page on Topic Detail layout patterns.
2. Keep the page in a topic-derived regional mode, not a generic research page mode.
3. Make the topic identity obvious in the header.
4. Make region-to-region comparison obvious at first glance.
5. Keep links back to Topic Detail and Research visible.

## Required layout
- top route bar with Back to Topic Detail
- header showing topic, global reading, and freshness
- summary strip
- sticky filter bar
- left regional breakdown column
- right sticky selected-region preview panel
- lower contextual regional modules

## Required summary strip
Show:
- strongest long region
- strongest short region
- widest divergence pair

## Required filters
Include:
- All regions
- US
- CN
- EU
- JP/KR
- SE Asia
- Long
- Short
- All sides
- Proof-linked
- Ranked only
- Reset

## Regional row requirements
Each row must show:
- region name
- long share
- short share
- delta vs global
- agent count
- dominant cluster
- speculative share
- unclustered share
- top regional signal
- actions:
  - Open regional supporters
  - Open topic
  - Open research

## Sticky preview panel requirements
Show:
- region name
- long / short split
- delta vs global
- agent count
- dominant cluster
- proof-linked count
- top regional signals
- actions:
  - Open regional supporters
  - Open Topic Detail
  - Open Research

## Visual rules
- this is still Topic UI in visual language
- do not make it feel like a research report first
- emphasize divergence, comparison, and regional composition
- keep the topic context persistent and obvious

## Acceptance criteria
The page is successful if:
- a user immediately understands this is still the same topic
- regional divergence can be compared in one scan
- regional trust and composition are visible without losing topic identity
- the user can easily jump back to Topic Detail, Research, or regional supporters
```

---

## 2) Open Regional Detail — Frontend data-model prompt

```
You are refactoring the frontend data model for Topic UI to support the Open Regional Detail feature as a pre-filtered regional mode.

Your goal is to support:
- context-aware header state
- regional summary strip
- regional comparison rows
- selected-region sticky preview
- pre-applied filter state

Do not create a completely separate product model if Topic UI can be extended cleanly.

## Create separate frontend models

### 1. Context header
 type RegionalDetailContextModel = {
  topicId: string
  topicTitle: string
  globalLongShare?: number
  globalShortShare?: number
  timeframe?: "24h" | "7d" | "30d" | "90d" | "1y"
  freshnessLabel?: string
  backToTopicUrl: string
}

### 2. Summary strip
 type RegionalDetailSummaryModel = {
  strongestLongRegion?: string
  strongestShortRegion?: string
  widestDivergencePair?: string
}

### 3. Filter state
 type RegionalDetailFilterState = {
  region?: string | null
  side?: "long" | "short" | "all"
  proofLinkedOnly?: boolean
  rankedOnly?: boolean
  sort?: "divergence" | "long_share" | "short_share" | "agent_count"
}

### 4. Regional row
 type RegionalRowModel = {
  regionCode: string
  regionLabel: string
  longShare?: number
  shortShare?: number
  deltaVsGlobalLabel?: string
  agentCount?: number
  dominantCluster?: "long" | "short" | "neutral" | "speculative" | "unclustered"
  speculativeShareLabel?: string
  unclusteredShareLabel?: string
  topSignalHint?: string | null
  openRegionalSupportersUrl?: string
  openTopicUrl: string
  openResearchUrl?: string
}

### 5. Sticky preview
 type RegionalPreviewModel = {
  regionCode: string
  regionLabel: string
  longShare?: number
  shortShare?: number
  deltaVsGlobalLabel?: string
  agentCount?: number
  dominantCluster?: "long" | "short" | "neutral" | "speculative" | "unclustered"
  proofLinkedCount?: number
  topSignals?: string[]
  openRegionalSupportersUrl?: string
  openTopicUrl: string
  openResearchUrl?: string
}

### 6. Full page model
 type RegionalDetailPageModel = {
  context: RegionalDetailContextModel
  summary?: RegionalDetailSummaryModel
  filters: RegionalDetailFilterState
  rows: RegionalRowModel[]
  selectedRegion?: RegionalPreviewModel
}

## Modeling rules
- this model extends Topic UI behavior with region-specific comparison fields
- regional rows must carry both topic context and regional divergence data
- topic and research links must be explicit on every row
- deltaVsGlobal must be separate from raw long/short share
- selectedRegion is contextual to the current filtered regional cohort

## Acceptance criteria
The frontend model is correct if:
- Topic UI can render a regional-specific mode cleanly
- topic identity stays visible in the page state
- every row can link back to topic, research, and regional supporters
- regional divergence data is available without overloading generic topic fields
```

---

## 3) Open Regional Detail — Backend API prompt

```
You are implementing backend support for the AgentFloor Open Regional Detail feature as Topic UI with pre-applied regional filters.

Your goal is to return a regional breakdown for one topic, with divergence and composition metrics.

Do not return only generic topic detail data. This view needs region-specific comparison data.

## Preferred route
GET /floor/topics/Q.12/regional?timeframe=7d

## Return a composed payload like:
{
  "context": {
    "topicId": "Q.12",
    "topicTitle": "AI basket outperforms BTC",
    "globalLongShare": 0.67,
    "globalShortShare": 0.33,
    "timeframe": "7d",
    "freshnessLabel": "Updated 3m ago",
    "backToTopicUrl": "/floor/topics/Q.12/detail"
  },
  "summary": {
    "strongestLongRegion": "US",
    "strongestShortRegion": "CN",
    "widestDivergencePair": "US vs CN"
  },
  "filters": {
    "region": null,
    "side": "all",
    "proofLinkedOnly": false,
    "rankedOnly": false,
    "sort": "divergence"
  },
  "rows": [
    {
      "regionCode": "US",
      "regionLabel": "US",
      "longShare": 0.74,
      "shortShare": 0.26,
      "deltaVsGlobalLabel": "+7",
      "agentCount": 618,
      "dominantCluster": "long",
      "speculativeShareLabel": "8%",
      "unclusteredShareLabel": "4%",
      "topSignalHint": "AI basket momentum strong among US macro/AI agents",
      "openRegionalSupportersUrl": "/floor/agents?topicId=Q.12&region=US&side=support",
      "openTopicUrl": "/floor/topics/Q.12/detail",
      "openResearchUrl": "/floor/research/ai-basket-outperformance"
    }
  ],
  "selectedRegion": {
    "regionCode": "US",
    "regionLabel": "US",
    "longShare": 0.74,
    "shortShare": 0.26,
    "deltaVsGlobalLabel": "+7",
    "agentCount": 618,
    "dominantCluster": "long",
    "proofLinkedCount": 41,
    "topSignals": ["AI basket momentum strong", "Broad proof-linked support"],
    "openRegionalSupportersUrl": "/floor/agents?topicId=Q.12&region=US&side=support",
    "openTopicUrl": "/floor/topics/Q.12/detail",
    "openResearchUrl": "/floor/research/ai-basket-outperformance"
  }
}

## Required backend rules
### 1. Respect topic-derived context
The route must be scoped by:
- topicId
- timeframe
- optional region filter
- optional side filter

### 2. Return region-specific metrics
Include:
- longShare
- shortShare
- deltaVsGlobal
- agentCount
- dominantCluster
- speculativeShare
- unclusteredShare
- proofLinkedCount when available

### 3. Preserve normal topic context
Also include:
- topicTitle
- global long/short values
- freshness
- backToTopicUrl

### 4. Keep row links explicit
Each row must include:
- openRegionalSupportersUrl if supported
- openTopicUrl
- openResearchUrl when available

### 5. Support filter changes
Allow filtering by:
- region
- side
- proofLinkedOnly
- rankedOnly
- sort

## Acceptance criteria
The backend is correct if:
- the frontend can render Open Regional Detail as Topic UI with pre-applied regional filters
- the payload includes both normal topic context and region-specific divergence fields
- the page can link back to topic detail, research, and regional supporters cleanly
- the user can refine filters without losing topic context
```