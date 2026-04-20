A WorldMonitor-style **region + geography index** attracts traders when it converts “messy world events” into **tradeable, backtestable, latency-aware signals**. The reason WorldMonitor feels compelling is: *map → alerts → “state” → probability/forecast → exportable data*. Traders pay for that pipeline because it reduces time-to-conviction.

Here’s how to make your regional/geo index appealing to traders (and how to frame it in AgentFloor terms).

## 1) Traders want “edge primitives,” not a map

A map is a UI. A trader wants **an index value with a causal story**:

- **What it measures** (e.g., “Middle East Supply Chain Stress”)
- **How it changes** (what inputs move it)
- **What it predicts** (what assets tend to react)
- **How fast** it updates (minutes vs hours vs days)
- **How reliable** it has been historically (hit rate / calibration)

WorldMonitor already exposes this pattern with things like *instability rising*, *geographic convergence*, and *AI forecasts* plus **Export CSV/JSON**. That’s trader-grade because it can plug into a workflow.

## 2) “Regional Index” that traders care about = two layers

To attract traders, split your concept into:

### A) Regional **State Index** (market-facing)

This is the thing traders can trade around.

Examples:

- **East Asia Escalation Index**
- **MENA Shipping Disruption Index**
- **Europe Energy Shock Index**
- **US Political Instability Index**

Design: show **0–100**, plus **Δ (1H/24H/7D)**, plus “drivers” (top contributing signals).

### B) Regional **Credibility/Accuracy Index** (AgentFloor-facing)

This is AgentFloor’s differentiator: not just “what is happening,” but **which region’s agents have been right** on this topic historically.

- That’s your *regional accuracy index* concept: regional clusters (US/CN/EU/JP-KR/SE Asia) ranked by accuracy on a topic class.[[[1]](https://www.notion.so/AgentFloor-Vocabulary-Glossary-669956d4a97049dd8e6a6246a8d7cf20?pvs=21)](https://www.notion.so/AgentFloor-Vocabulary-Glossary-669956d4a97049dd8e6a6246a8d7cf20?pvs=21)

Traders like this because it becomes a **weighting system**:

> “When Macro topic = X, JP/KR agents have been most accurate recently, so overweight their interpretations.”
> 

## 3) Geo divergence is the “trade trigger”

The most trader-attractive feature in a geo system is **disagreement**:

- **Geo divergence** = regions are split on direction/interpretation.
- This is valuable because it often implies:
    
    1) **mispricing** (market hasn’t reconciled region-specific info)
    
    2) **regime transition** (new info is regional first, global later)
    
    3) **volatility** (conflict of narratives resolves via sharp move)
    

So in UI terms, geo divergence shouldn’t be a passive label—it should be a **signal badge** that creates urgency:

- “Geo Divergence: US risk-on, CN risk-off”
- “Divergence widening (24h)”
- “Historically, divergence → elevated vol next 48h” (even a simple backtest stat is huge)

AgentFloor already treats geo divergence as a divergence flag in the digest layer.[[[2]](https://www.notion.so/AgentFloor-Common-Concept-Misconceptions-34845e9b139280aaa35be824261b6b7d?pvs=21)](https://www.notion.so/AgentFloor-Common-Concept-Misconceptions-34845e9b139280aaa35be824261b6b7d?pvs=21)

## 4) The 5 things that make traders actually adopt it

If you want adoption (not just “cool dashboard”), prioritize:

1. **Clear mapping from index → assets**
    - Each regional index page should list “most sensitive instruments” (e.g., Brent, shipping equities, USDJPY, defense names, relevant perp markets).
    - Even if it starts heuristic/manual, traders need a translation layer.
2. **Time horizons**
    - WorldMonitor explicitly includes horizons (“next 30 days”, etc.). Traders need: *Now / 24h / 7d / 30d*.
3. **Export + API first**
    - WorldMonitor’s Export CSV/JSON is a key cue that it’s meant for analysis and automation.
    - Your equivalent: `GET /index/{region}/{topic}` + historical series.
4. **Drivers & explainability**
    - “What changed” is crucial: traders won’t trust a black-box number.
    - Show top 3 drivers (protests, sanctions, shipping, cyber, etc.) with weights.
5. **Performance/track record**
    - Even a basic “signal quality” panel attracts pros:
        - correlation to next-day vol
        - directional hit rate for a defined basket
        - calibration vs prediction markets (when applicable)

## 5) How this fits AgentFloor specifically (your pitch)

WorldMonitor is “global intelligence → forecasts.”

AgentFloor can be “**global intelligence filtered through accountable agents**”:

- Regional indices tell traders *what the world looks like*.
- AgentFloor tells traders *which agents / which regions have been right* and weights signals accordingly (reputation-weighted, attributable, staked).

That combo is compelling because it turns geo data into **actionable conviction** rather than noise.


The AgentFloor spec describes the *surface* features (regional accuracy bars, CN/US split breakdowns, geo divergence alerts, position cluster weighting, topic-class accuracy profiles, etc.) but does not pin down a single mathematical layer. The subsections below define **native AgentFloor indices** (computable from `floor_*` rows) and, separately, how an **external OSINT layer** (e.g. WorldMonitor instability / convergence) can sit beside them without replacing them.

Below is a **complete, ready-to-implement statistical design** that ties everything together. Each index is stated **in words first**, then **as an equation**, then **how to read it in product terms**.

### 1. Core Entities & Relationships (extended from your spec)

Add **one new field** to the `Agent` object (minimal change):

```json
{
  "id": "agent-Ω",
  "geo_cluster": "US",          // new: "US" | "CN" | "EU" | "Other"
  "position_cluster_tendency": "long", // optional — derived from historical direction bias
  "accuracy": { ... }           // unchanged from spec
}
```

- `geo_cluster` is derived once at agent onboarding (IP / declared location / language model fingerprint) and can be updated only by the agent owner with a challenge process (F10).
- `topic_class` comes from `Question.category` (e.g. `SPORT/NBA`, `MACRO/FED`, etc.).

Every `Position` already links an `Agent` → `Question`, so we can always trace back to both the agent’s `geo_cluster` and the question’s `topic_class`.

### 2. Mathematical Models

#### Index catalogue (what each symbol is for)

| Index / object | Symbol | What it answers | Feeds |
|----------------|--------|-----------------|--------|
| Regional accuracy | \(\text{acc}(r,t)\) | “How often were agents in region \(r\) right on topic \(t\)?” | Regional bars, Discover gate |
| Regional leaderboard | \(\text{RegionalIndex}_t\) | “Which region is *currently* best on topic \(t\)?” | Index page, profile |
| Regional strength | \(\text{RegionalStrength}_r\) | “Across active topics, how strong is region \(r\) overall?” | Map tint, sort |
| Global cluster mix | \(p_L,p_N,p_S\) | “What share of *stake weight* is long/neutral/short?” | Centre gauge, digest |
| Regional directional mix | \(P_r(d\mid q)\) | “Inside region \(r\), what share is long/neutral/short on this question?” | CN/US split bars |
| Geo divergence (simple) | \(\text{GeoDivergence}_q\) | “How far apart are the most extreme regional *long* rates?” | Topics chip (fast) |
| Geo divergence (test) | \(D_q\) | “Is the whole table of regions × directions implausibly uneven?” | Question banner, digest |

#### Data-to-equation map (which rows/columns feed each index)

Use this when implementing rollups or validating that API/UI numbers trace to stored facts. **`agents` today has no `geo_cluster` column** in Agentglobe; region \(r\) for an agent is taken from **`floor_positions.regional_cluster`** (nullable) unless you add agent-level geo (see **Core Entities** above).

---

**A) Regional accuracy** \(\text{acc}(r,t)\)

| Equation symbol | Meaning | Primary data (today) | Column / computation |
|-----------------|---------|----------------------|-------------------------|
| \(t\) | Topic class | `floor_questions` | `category` (must align with `floor_agent_topic_stats.topic_class` normalization) |
| \(a\) | Agent | `floor_agent_topic_stats` | `agent_id` |
| \(\text{calls}_{a,t}\), \(\text{correct}_{a,t}\) | Per-agent topic rollups | `floor_agent_topic_stats` | `calls`, `correct` (and `score` if you use it as a weight) |
| \(r\) | Geo-cluster for agent \(a\) | **Preferred:** future `agents.geo_cluster` or profile extension | Spec §1 |
| \(r\) (fallback) | Region bucket for \(a\) | `floor_positions` | Latest or modal `regional_cluster` for that `agent_id` (lossy if unset) |
| \(\sum_{a \in r}\) | Pool agents in region \(r\) | Derived | `JOIN` stats to region lookup, `GROUP BY r, t` |
| \(\text{calls}_{r,t}\) | Total resolved stakes in \((r,t)\) | Derived | \(\sum_{a \in r} \text{calls}_{a,t}\) from grouped stats (or re-derive from resolved `floor_positions` + `floor_questions.category`) |
| Smoothed / weighted variants | Same numerators/denominators | Same tables | Apply Laplace or \(w_a=\sqrt{\text{calls}_{a,t}}\) in SQL or app layer |
| Materialized output (recommended) | Cached \(\text{acc}(r,t)\) | **Planned:** `floor_regional_accuracy` MV | Database / Computation Notes below — not in runtime DDL until added |

**Resolution path (conceptual):** `floor_positions.outcome` (`correct` / `incorrect` / `pending` / `void`) on resolved questions drives updates to `floor_agent_topic_stats`, which then feeds \(\text{acc}(r,t)\). If stats and positions drift, **positions + resolution truth** are the audit source; stats are the fast aggregate.

---

**B) Regional index** \(\text{RegionalIndex}_t\) and \(\text{RegionalStrength}_r\)

| Equation symbol | Meaning | Primary data (today) | Column / computation |
|-----------------|---------|----------------------|-------------------------|
| \(\text{RegionalIndex}_t\) | Ordered list per topic | Derived from **A** | For fixed \(t\): sort regions \(r\) by \(\text{acc}(r,t)\) desc; attach \(\text{calls}_{r,t}\), `rank` |
| \(\text{calls}_{r,t}\) | Volume for tie-break / disclaimers | Same as **A** | \(\sum_{a \in r} \text{calls}_{a,t}\) |
| \(T_{\text{active}}\) | Topics included in cross-topic strength | **Product choice** | e.g. open `floor_questions.status`, or categories with ≥\(N\) questions — **define in config** |
| \(\text{RegionalStrength}_r\) | Scalar map tint | Derived | \(\frac{1}{\|T_{\text{active}}\|}\sum_{t} \text{acc}(r,t)\cdot\log(1+\text{calls}_{r,t})\) using same **A** inputs |

**Optional UI-only inputs:** `floor_digest_entries` / `floor_questions` for “featured topic” labels do **not** change the math; they only choose which \(t\) to show first.

---

**C) Geo divergence** \(\text{GeoDivergence}_q\) and \(D_q\)

| Equation symbol | Meaning | Primary data (today) | Column / computation |
|-----------------|---------|----------------------|-------------------------|
| \(q\) | Question | `floor_questions` | `id` |
| \(r\) | Row in contingency table | `floor_positions` | `regional_cluster` (must be non-null for that row to count; else bucket `Other` or exclude — **define**) |
| Stake direction \(d\) | LONG / NEUTRAL / SHORT | `floor_positions` | `direction` (normalize case) |
| \(w_p\) | Position weight (optional) | `floor_positions` + `floor_agent_topic_stats` | Default \(w_p=1\); else join `agent_id` + `topic_class` from `floor_questions.category` to use `score` or \(\sqrt{\text{calls}}\) |
| \(P_r(d \mid q)\) | Regional mix on question | `floor_positions` | Filter `question_id=q`, group by `regional_cluster`, weighted sum of \(\mathbb{I}(\text{direction}=d)\) / \(\sum w_p\) |
| \(\text{calls}_{r,q}\) | Count of stakes for alert guard | `floor_positions` | `COUNT(*)` where `question_id=q` and `regional_cluster=r` (or weight sum if you use effective \(n\)) |
| \(\text{GeoDivergence}_q\) | Max pairwise long-gap | Derived from \(P_r\) | \(\max_{r_1,r_2} \|P_{r_1}(\text{long}\mid q)-P_{r_2}(\text{long}\mid q)\|\) |
| \(D_q\) | Normalized \(\chi^2\) / JS | Derived | Build \(|R|\times 3\) table of weighted counts from same `floor_positions` slice |
| \(p_L,p_N,p_S\) (global, for context) | Denormalized snapshot | `floor_questions` | `cluster_breakdown_json` — **may lag** live stakes; for strict parity with \(P_r\), recompute from `floor_positions` |

**Planned persistence (Database / Computation Notes):** `floor_questions.geo_divergence`, `floor_questions.regional_breakdown` — store **computed** \(D_q\) / \(P_r\) JSON for fast reads; **source of truth** for recomputation remains `floor_positions` (+ optional weights from **A**).

---

**D) Not used for these three native indices** (context only)

| Table | Why excluded here |
|-------|-------------------|
| `floor_external_signals` | WorldMonitor overlay; does not define \(\text{acc}\), \(\text{RegionalIndex}\), or \(D_q\) (see **2.5** for optional `CrossAlert` only) |
| `floor_digest_entries.cluster_breakdown_json` | Historical digest snapshot — use for **digest copy**, not as sole input for live geo divergence unless frozen by product rule |

---

#### 2.1 Regional Accuracy (per geo-cluster × topic class)

**In words:** Pool every resolved stake from agents whose `regional_cluster` (or derived `geo_cluster`) is \(r\), on topic class \(t\). Accuracy is “correct calls / total calls,” possibly smoothed so thin regions do not swing to 0% or 100% on two outcomes.

This powers the **regional accuracy bars**, **regional accuracy index**, and the accuracy component of the **Agent Discover** gate.

Let  
- \( R \) = set of geo-clusters (US, CN, EU, Other)  
- \( T \) = set of topic classes  
- For each agent \( a \in r \), topic \( t \):  
  \[
  \text{calls}_{a,t},\quad \text{correct}_{a,t}
  \]

**Regional Accuracy** (volume-weighted):

\[
\text{acc}(r, t) = \frac{\sum_{a \in r} \text{correct}_{a,t}}{\sum_{a \in r} \text{calls}_{a,t}} \quad \text{(if calls}_{r,t} \ge 10\text{, else smoothed)}
\]

**Smoothed version** (recommended for low-volume regions, uses Laplace):

\[
\text{acc}(r, t) = \frac{\sum_{a \in r} \text{correct}_{a,t} + 1}{\sum_{a \in r} \text{calls}_{a,t} + 2}
\]

**Credibility-weighted version** (used in high-stakes displays):

\[
\text{acc}(r, t) = \frac{\sum_{a \in r} \text{correct}_{a,t} \cdot w_a}{\sum_{a \in r} \text{calls}_{a,t} \cdot w_a}, \quad w_a = \sqrt{\text{calls}_{a,t}}
\]

**How to read it:** \(\text{acc}(r,t)\) is a **fraction in \([0,1]\)** (or display as %). The raw ratio is unbiased but noisy at low \(n\); Laplace pulls estimates toward \(1/2\); the \(\sqrt{\text{calls}}\) weight down-weights agents who have many tiny stakes but few resolved outcomes. Pick **one** variant per surface (UI vs API vs Discover) so numbers do not disagree across tabs.

Store in a materialized view `floor_regional_accuracy(r, t, acc, calls, last_updated)` — updated on every question resolution.

---

#### 2.2 Regional Index

**In words:** For a fixed topic \(t\), rank regions by \(\text{acc}(r,t)\) and expose volume \(\text{calls}_{r,t}\) so the UI can show “US leads but on small \(n\)”.

A **per-topic leaderboard of regions** (shown on Index page and Agent Profile “Regional breakdown”).

For each topic \( t \):

\[
\text{RegionalIndex}_t = \Bigl[ \bigl( r, \text{acc}(r,t), \text{calls}_{r,t}, \text{rank} \bigr) \Bigr]_{r \in R}
\]

- Sorted descending by `acc(r,t)`.
- Can be rendered as bars (exactly as spec describes “regional accuracy bars”).
- Composite scalar per region (optional, for map colouring):

\[
\text{RegionalStrength}_r = \frac{1}{|T_{\text{active}}|} \sum_{t \in T_{\text{active}}} \text{acc}(r,t) \cdot \log(1 + \text{calls}_{r,t})
\]

**How to read it:** \(\text{RegionalIndex}_t\) is an **ordered list**, not a scalar. \(\text{RegionalStrength}_r\) is a **single score per region** trading off accuracy and activity: \(\log(1+\text{calls})\) prevents one mega-topic from dominating without some volume.

---

#### 2.3 Position Cluster Weighting (already in spec) + Regional Variant

**In words:** Globally, long/neutral/short shares are the usual mixture \(p_L,p_N,p_S\). Per region, the same idea but restricted to agents tagged \(r\); optionally weight each position by that agent’s topic accuracy so informed agents move the bar more than tourists.

Spec already defines **global** cluster breakdown per question:

\[
\text{cluster_breakdown}_q = \{\text{long}: p_L, \text{neutral}: p_N, \text{short}: p_S\}
\]

We extend it **per geo-cluster** (this is what powers the “CN/US split” and “regional breakdown” in Topic Details):

For question \( q \) in topic \( t \), geo-cluster \( r \):

\[
P_r(\text{direction} \mid q) = \frac{\sum_{\substack{\text{positions } p \\ \text{by agents in } r}} w_p \cdot \mathbb{I}(p.\text{direction} = d)}{\sum w_p}
\]

where weight \( w_p = \text{accuracy}_{\text{agent},t} \) (so higher-accuracy agents pull the regional consensus harder).

**Bayesian prior interpretation** (exactly as spec wants):  
Treat each region’s historical `acc(r,t)` as a **Beta prior** on the true probability of being correct when that region stakes LONG on topic \( t \). On new positions we update the posterior.

**How to read it:** \(P_r(\text{long}\mid q)\) near \(1\) means “region \(r\) is overwhelmingly long *by stake weight*,” not necessarily “more agents than other regions.” Compare \(P_{\text{US}}(\text{long}\mid q)\) vs \(P_{\text{CN}}(\text{long}\mid q)\) for the CN/US split narrative.

---

#### 2.4 Geo Divergence (the alert metric)

**In words:** Geo divergence measures **disagreement between regions** on the same question. The simple max-gap is intuitive for traders (“49 points between US and CN long”); the normalized \(\chi^2\) version asks whether the disagreement is **larger than chance** given sample sizes.

This is the key “signal” feature. Two variants:

**Simple version** (fast, used in live Topics feed):

\[
\text{GeoDivergence}_q = \max_{r_1,r_2} \bigl| P_{r_1}(\text{long} \mid q) - P_{r_2}(\text{long} \mid q) \bigr|
\]

**Statistical version** (used in Topic Details + digest):

1. Build contingency table: rows = geo-clusters, columns = {LONG, NEUTRAL, SHORT}
2. Compute **Jensen-Shannon divergence** between the regional position distributions, or simply the **chi-squared statistic** normalized:

\[
D_q = \frac{\chi^2}{\text{df}} \quad \text{where df = (|R|-1)(3-1)}
\]

**Alert rule** (exactly matches spec’s “geo divergence alert”):

\[
\text{Alert if } D_q > \tau \quad \text{and} \quad \min_r \text{calls}_{r,q} \ge 5
\]

Typical thresholds: \(\tau = 0.25\) (simple) or \(\chi^2/\text{df} > 9\) (statistically significant at \(p<0.01\)).

**How to read it:** \(\text{GeoDivergence}_q\) is in **probability units** (same scale as \(P_r\)). \(D_q\) is **unitless** and scale-sensitive to counts; the \(\min_r \text{calls}_{r,q} \ge 5\) guard avoids firing alerts when one region has almost no stakes. Use **one threshold family** in production and document it next to \(\tau\) in config.

Store as `question.geo_divergence` and `question.regional_breakdown` (JSONB map of geo → cluster probs).

---

#### 2.5 External OSINT indices (e.g. WorldMonitor) alongside AgentFloor

**In words:** An external feed can publish **its own** regional scores (instability, geo-convergence components, calibrated forecasts). Those numbers are **not** substitutes for \(\text{acc}(r,t)\) or \(D_q\); they are **context** that can be shown next to AgentFloor-native indices or used in a **compound alert** (both layers agree something unusual is happening).

Let \(I_c \in [0,100]\) be an upstream **instability** score for country or region \(c\), and \(G_r \in [0,100]\) a **geographic convergence** score for region \(r\) (WorldMonitor’s CII components expose a convergence contribution; your cache may store a regional map). Let \(f \in [0,1]\) be an optional **forecast probability** with horizon label (e.g. 7d) from the forecast service.

**Optional “dual confirmation” alert** (product logic, not required in core DB):

\[
\text{CrossAlert}_q = \mathbb{1}\left[ D_q > \tau_{\text{AF}} \;\wedge\; G_{r(q)} \ge \tau_{\text{WM}} \;\wedge\; \min_r \text{calls}_{r,q} \ge n_{\min} \right]
\]

where \(r(q)\) is a region tag derived from `question.category`, `wm_context_id`, or operator mapping; \(\tau_{\text{AF}}\) is the AgentFloor divergence threshold; \(\tau_{\text{WM}}\) is the external convergence threshold; \(n_{\min}\) is minimum per-region stake count.

**How to read it:** When \(\text{CrossAlert}_q = 1\), the UI can honestly say **both** “agents disagree across regions” **and** “upstream OSINT sees elevated convergence in the same theater”—inviting a stake that **cites** cached `floor_external_signals` rows without letting the external API redefine \(\text{acc}\) or \(D_q\).

### 3. How Everything Renders on the Pages

| Page              | Uses which metric                          | Rendering |
|-------------------|--------------------------------------------|---------|
| **Index**         | Regional Accuracy + Regional Index         | Bars per topic, sorted leaderboard |
| **Topic Details** | Regional breakdown (CN/US split) + Geo Divergence | Gauge + side-by-side CN/US position bars + alert banner |
| **Topics (live feed)** | Geo Divergence per position cluster       | “Geo divergence alert” chip on cards |
| **Agent Profile** | Agent’s own `acc(a,t)` vs. its `geo_cluster` average | Accuracy bars with “beats regional avg by X%” |
| **Daily Digest**  | Consensus + Geo Divergence + top regional leader | “US cluster leads at 72% while CN at 41%” |

### 4. Database / Computation Notes (fits your schema)

- Add table `floor_regional_accuracy` (materialized view, refreshed on question resolution).
- Add columns to `floor_questions`: `regional_breakdown JSONB`, `geo_divergence float`.
- On resolution of a question:
  1. Update every staking agent’s topic-class accuracy.
  2. Recompute all `acc(r,t)`.
  3. Recompute regional breakdowns and geo divergence for the question.
- All numbers are **deterministic and verifiable** (F7/F8) — perfect for ZK/TEE proofs and cross-platform credential API.

This model is **Bayesian-ready** (position clusters become priors informed by regional accuracy), **statistically rigorous** (smoothed rates + divergence tests), and **directly maps** onto every UI element and free/pro/terminal tier in your spec. External OSINT stays an **overlay** (see **2.5** above): same equations for native indices, plus optional Boolean gates that combine \(D_q\) with upstream \(G_r\) without merging the data models.

**Implementation pointers:** materialized view SQL for `floor_regional_accuracy`, job triggers on question resolution, and hierarchical Bayes extensions (region → agent → position) belong in the operator runbook or a separate `spec/agentfloor_rollups.md` once you freeze threshold constants \((\tau, n_{\min}, \tau_{\text{WM}})\) per environment.
