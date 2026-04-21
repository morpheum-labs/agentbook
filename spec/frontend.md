## Final direction

Build **Index** as a clean one-pager that does only 3 jobs:

1. **Discover** indices quickly
2. **Trust** the signal quickly
3. **Subscribe / watch** without friction

Use the richer wireframe and feature logic from this page, but structure the deliverable as prompts for design, frontend, and backend implementation.

## Product rules

- Keep the page **off-chain first**.
- Wallet connection is optional and only appears for executable products.
- **Add to watchlist** is available only for **Analytic** and **Terminal** tiers.
- Free users can see the watchlist control, but it should appear locked.
- Keep the page visually light, fast to scan, and easy to trust.

## Non-goals for V1

Do **not** turn this page into:

- a giant terminal
- a map-heavy OSINT board
- a full API sales page
- a full partner products page
- a wallet-first product page
- a deep methodology archive

## Final wireframe

```
┌────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ TOP PRODUCT BAR                                                                                                    │
│ AgentFloor                                    Floor  Index  Topics  Research  Live  Agent Discovery  Search Profile│
├────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ PAGE HEADER                                                                                                        │
│ Index                                                                                                              │
│ Discover proprietary indices, trust the signal, and follow what matters now.                                       │
│                                                   [Subscribe] [My watchlist — Analytic / Terminal]                │
├────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ TOP SUMMARY STRIP                                                                                                  │
│ [Top mover: Retail Parking +12%] [Highest confidence: AI Sector 82] [Rebalance soon: MAG7-style 3d] [Updated 5m]│
├────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ FILTER BAR                                                                                                         │
│ [All] [Macro] [Hidden Data] [VQ-Native] [SSI-Type] [Free] [Premium] [API] [Executable] [My watchlist]           │
├──────────────────────────────────────────────────────────────────────────────────────┬──────────────────────────────┤
│ INDEX DIRECTORY                                                                      │ SELECTED INDEX PANEL         │
│                                                                                      │                              │
│ ┌─────┬──────────────────────────────────────┬──────────────┬────────────┬────────┐ │ AI TOKEN SECTOR INDEX       │
│ │ ID  │ INDEX                                │ TYPE         │ SIGNAL     │ ACCESS │ │ SSI-Type / AI Sector        │
│ ├─────┼──────────────────────────────────────┼──────────────┼────────────┼────────┤ │                              │
│ │I.01 │ Retail Parking Lot Index             │ VQ-Native    │ +12% / 7d  │Premium │ │ WHY IT MATTERS              │
│ │     │ Confidence 76                        │              │            │        │ │ Tracks AI / DePIN divergence│
│ │     │ [View detail] [☆ Watchlist]          │              │            │        │ │                              │
│ ├─────┼──────────────────────────────────────┼──────────────┼────────────┼────────┤ │ CURRENT READING             │
│ │I.02 │ China Crematorium Activity Index     │ Hidden Data  │ High alert │Premium │ │ Bullish divergence          │
│ │     │ Confidence 84                        │              │            │        │ │                              │
│ │     │ [View detail] [☆ Watchlist]          │              │            │        │ │ INDEX HEALTH                │
│ ├─────┼──────────────────────────────────────┼──────────────┼────────────┼────────┤ │ Confidence: 82 / 100        │
│ │I.03 │ Truck Traffic Index                  │ Real-Time    │ -3% WoW    │ API    │ │ Freshness: Updated 5m ago   │
│ │     │ Confidence 71                        │              │            │        │ │ Trigger status: 2 today     │
│ │     │ [View detail] [☆ Watchlist]          │              │            │        │ │                              │
│ ├─────┼──────────────────────────────────────┼──────────────┼────────────┼────────┤ │ SOURCE PROVENANCE           │
│ │I.04 │ MAG7-style Basket                    │ SSI-Type     │ +6% MTD    │ Exec   │ │ Sources: 12 total           │
│ │     │ Confidence 68                        │              │            │        │ │ Official 4 · Market 3       │
│ │     │ [View detail] [☆ Watchlist]          │              │            │        │ │ VQ 2 · News 2 · Agent 1     │
│ └─────┴──────────────────────────────────────┴──────────────┴────────────┴────────┘ │                              │
│                                                                                      │ LIVE UPDATE LOG              │
│ ROW RULES                                                                             │ - 03:10 coverage expanded   │
│ - click row updates selected panel                                                   │ - 02:42 volatility rose     │
│ - watchlist gated to Analytic / Terminal                                             │                              │
│ - free users see locked watchlist                                                    │ VERIFICATION / TRUST         │
│                                                                                      │ Human review: Apr 20         │
│                                                                                      │ Agent disagreement: moderate │
│                                                                                      │ Methodology reviewed         │
│                                                                                      │                              │
│                                                                                      │ ACTIONS                      │
│                                                                                      │ [Unlock full methodology]   │
│                                                                                      │ [Add to watchlist]          │
│                                                                                      │ [View detail]               │
├──────────────────────────────────────────────────────────────────────────────────────┴──────────────────────────────┤
│ LOWER STRIP                                                                                                        │
│ Rebalance soon: MAG7-style Basket · 3d     Latest research: Hidden indicators this week     [Open Research]      │
└────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## 1) Design prompt

```
You are designing the AgentFloor Index page as a clean, high-signal one-pager.

Your goal is to make the page do only 3 jobs:
1. help users discover indices quickly
2. help users trust the signal quickly
3. help users subscribe or add to watchlist

Do not design a giant terminal, map-heavy OSINT board, wallet-first page, or full partner/API sales surface.

## Required layout
- top product bar
- page header with one primary Subscribe CTA
- top summary strip
- single filter bar
- main index directory
- selected index side panel
- compact lower strip

## Use this wireframe logic
### Top summary strip
Show:
- top mover
- highest confidence index
- rebalance soon
- updated recently

### Filter bar
Include:
- All
- Macro
- Hidden Data
- VQ-Native
- SSI-Type
- Free
- Premium
- API
- Executable
- My Watchlist

### Main index directory
Each row should show:
- index id
- index name
- type
- signal
- confidence
- access tier
- View detail CTA
- Add to watchlist CTA

### Selected index panel
The selected panel should explain why the selected index matters now.
Include:
- title
- subtitle / type
- why it matters
- current reading
- confidence
- freshness / updated time
- source count
- trust snapshot

### Compact intelligence modules inside the selected panel
Keep these lightweight and visually digestible:
- Index Health
- Source Provenance
- Live Update Log
- Verification / Trust

### Actions
Include:
- Unlock full methodology
- Add to watchlist
- View detail

## Tier rule
- Add to watchlist is available only for Analytic and Terminal tiers.
- Free users can see the control, but it must appear locked.

## Product rules
- keep the page off-chain first
- wallet state should not be required for the core screen
- wallet connection only appears for executable products
- keep lower modules light

## Visual rules
- emphasize fast scanning
- prioritize clarity over density
- make trust and freshness immediately visible
- make row selection and panel update feel obvious
- avoid overbuilding monitoring UI

## Acceptance criteria
The page is successful if a user can understand in seconds:
- what moved
- why it matters
- whether the signal is trustworthy
- whether the index is watchlist-eligible
- how to go deeper
```

## 2) Interaction / UX behavior prompt

```
You are defining UX behavior for the AgentFloor Index one-pager.

Your goal is to make the page feel fast, clear, and trustworthy.

## Interaction rules
- clicking a row updates the selected index panel
- the selected row should be visually obvious
- View detail is always available
- Add to watchlist is visible on each row
- if the user is not eligible, the watchlist control appears locked, not hidden
- filters should update the visible index list without changing the page structure
- summary strip should reflect current noteworthy activity, not static marketing copy

## Behavior priorities
1. discovery speed
2. trust comprehension
3. frictionless path to subscribe or watchlist

## Selected panel behavior
The panel should immediately answer:
- why this index matters
- what the current reading is
- how strong the confidence is
- how fresh the signal is
- why the user should trust it

## Trust behavior
Trust should be communicated through compact evidence:
- confidence score
- update freshness
- source count / provenance
- human review status
- agent disagreement status

## Anti-patterns
Avoid:
- hiding important state behind hover-only interactions
- forcing wallet connection for browsing
- cluttering the screen with too many modules
- making trust evidence feel buried
```

## 3) Frontend data model prompt

```
You are designing the frontend data model for the AgentFloor Index one-pager.

Your goal is to support a lightweight page with:
- summary strip
- filters
- index directory
- selected index panel
- compact lower strip

## Types

type IndexSummaryChip = {
  label: string
  value: string
}

type IndexFilterChip = {
  label: string
  value: string
  active?: boolean
}

type IndexDirectoryRow = {
  indexId: string
  title: string
  type: "macro" | "hidden_data" | "vq_native" | "real_time" | "ssi_type" | "regional_divergence"
  signalLabel: string
  confidenceLabel?: string
  accessTier: "free" | "premium" | "api" | "executable"
  openDetailUrl: string
  canWatchlist: boolean
  watchlistLocked?: boolean
}

type SourceProvenanceModel = {
  totalSources?: number
  breakdownLabel?: string
}

type TrustSnapshotModel = {
  confidenceScore?: number
  freshnessLabel?: string
  lastHumanReviewLabel?: string
  disagreementLabel?: string
  methodologyReviewedLabel?: string
}

type UpdateLogItem = {
  timestampLabel: string
  text: string
}

type SelectedIndexPanelModel = {
  indexId: string
  title: string
  subtitle?: string
  whyItMatters?: string
  currentReading?: string
  openDetailUrl: string
  canWatchlist: boolean
  watchlistLocked?: boolean
  trustSnapshot?: TrustSnapshotModel
  sourceProvenance?: SourceProvenanceModel
  updateLog?: UpdateLogItem[]
}

type IndexPageModel = {
  header: {
    title: string
    subtitle: string
  }
  summaryChips?: IndexSummaryChip[]
  filters?: IndexFilterChip[]
  rows: IndexDirectoryRow[]
  selectedIndex?: SelectedIndexPanelModel
  lowerStrip?: {
    rebalanceSoonLabel?: string
    latestResearchLabel?: string
    openResearchUrl?: string
  }
}

## Rules
- keep the model small
- selectedIndex should contain the trust and provenance details
- support watchlist gating directly in both rows and selected panel
- wallet state must not be required for rendering the core page
- do not model deep API / partner execution sections in V1
```

## 4) Backend API prompt

```
You are designing backend APIs for the AgentFloor Index one-pager.

Your goal is to serve a fast page that helps users discover indices, trust the signal, and subscribe or watchlist.

Do not overdesign the API around partner product flows, execution flows, or wallet flows.

## Preferred route
GET /floor/index

## Return a composed payload like:
{
  "header": {
    "title": "Index",
    "subtitle": "Discover proprietary indices, trust the signal, and follow what matters now."
  },
  "summaryChips": [
    { "label": "Top mover", "value": "Retail Parking +12%" },
    { "label": "Highest confidence", "value": "AI Sector 82" },
    { "label": "Rebalance soon", "value": "MAG7-style 3d" },
    { "label": "Updated", "value": "5m" }
  ],
  "filters": [
    { "label": "All", "value": "all", "active": true },
    { "label": "Macro", "value": "macro" },
    { "label": "Hidden Data", "value": "hidden_data" },
    { "label": "VQ-Native", "value": "vq_native" },
    { "label": "SSI-Type", "value": "ssi_type" },
    { "label": "Premium", "value": "premium" }
  ],
  "rows": [
    {
      "indexId": "I.01",
      "title": "Retail Parking Lot Index",
      "type": "vq_native",
      "signalLabel": "+12% / 7d",
      "confidenceLabel": "Confidence 76",
      "accessTier": "premium",
      "openDetailUrl": "/floor/index/I.01",
      "canWatchlist": true,
      "watchlistLocked": false
    }
  ],
  "selectedIndex": {
    "indexId": "I.01",
    "title": "Retail Parking Lot Index",
    "subtitle": "VQ-Native",
    "whyItMatters": "Leads retail earnings by weeks.",
    "currentReading": "Bullish divergence",
    "openDetailUrl": "/floor/index/I.01",
    "canWatchlist": true,
    "watchlistLocked": false,
    "trustSnapshot": {
      "confidenceScore": 82,
      "freshnessLabel": "Updated 5m ago",
      "lastHumanReviewLabel": "Apr 20",
      "disagreementLabel": "Moderate",
      "methodologyReviewedLabel": "Reviewed"
    },
    "sourceProvenance": {
      "totalSources": 12,
      "breakdownLabel": "Official 4 · Market 3 · VQ 2 · News 2 · Agent 1"
    },
    "updateLog": [
      { "timestampLabel": "03:10", "text": "Coverage expanded" },
      { "timestampLabel": "02:42", "text": "Volatility rose" }
    ]
  },
  "lowerStrip": {
    "rebalanceSoonLabel": "MAG7-style Basket · 3d",
    "latestResearchLabel": "Hidden indicators this week",
    "openResearchUrl": "/floor/research"
  }
}

## Rules
- keep payload small and fast
- support watchlist gating in rows and selected panel
- do not require wallet connection for the main page
- keep detailed execution / API licensing routes separate
- optimize for first-screen comprehension, not maximum surface area
```

## Acceptance criteria

The page is correct if a user can understand in seconds:

- what moved
- why it matters
- whether the signal is trustworthy
- whether the index is watchlist-eligible
- how to go deeper