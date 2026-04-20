**✅ Execution plan for AgentFloor: Keep our core (agent-staked) model 100% native — but **consume** WorldMonitor.app as a high-value external context layer.**

This directly executes the differentiation we locked in yesterday:

- WorldMonitor = **human/OSINT/AI intelligence feed** (maps, instability indices, geo-convergence, AI forecasts calibrated to Polymarket, etc.).
- AgentFloor = **agent epistemic credentialing + staked prediction market**.

We do **NOT** redesign our DB or API “based on” WorldMonitor (that would dilute our moat and break F1–F10).  
We **DO** add a clean, tier-gated integration that turns WorldMonitor data into **context** that makes agents **want to stake harder** on our platform.

### Why this is the winning execution path
- WorldMonitor already has a **public, documented, open-source REST + protobuf API** (see their docs at `worldmonitor.app/docs/api-*` and GitHub `koala73/worldmonitor`).
- It exposes exactly the regional/geo data we want (country instability scores, convergence zones, forecasts, signals) — we can pull it without reinventing the wheel.
- This creates the behavioral flywheel:  
  Agents/humans see WorldMonitor signals → read geo-divergence on AgentFloor → stake positions → accuracy records update → daily digest becomes tradable alpha.

### Proposed DB + API design (minimal, backward-compatible)

#### 1. Database extensions (Agentglobe + floor_* schema)
Add **two new optional tables** (or JSONB columns if you want ultra-light):

```sql
-- 1. External signal cache (per question or topic)
CREATE TABLE floor_external_signals (
  id                SERIAL PRIMARY KEY,
  question_id       TEXT REFERENCES floor_questions(id),
  topic_class       TEXT,                    -- e.g. "MACRO/MENA"
  fetched_at        TIMESTAMPTZ,
  source            TEXT DEFAULT 'worldmonitor',
  raw_data          JSONB,                   -- full WorldMonitor payload (instability index, forecasts, signals)
  instability_index JSONB,                   -- { "Iran": 97, "components": {...} }
  geo_convergence   JSONB,                   -- { "MiddleEast": { "score": 100, "signals": [...] } }
  forecast_summary  JSONB                    -- { "probability": 0.80, "horizon": "7d", "branch": "base" }
);

-- 2. AgentFloor → WorldMonitor mapping (for rich display)
ALTER TABLE floor_questions 
  ADD COLUMN wm_context_id TEXT;  -- optional reference to WorldMonitor focal point / convergence zone
```

- Keep everything **deterministic and verifiable** (F7): store the exact WorldMonitor timestamp + signature if they provide one.
- Materialized view for speed: `floor_regional_accuracy` stays **our** math (the one I gave you earlier).

#### 2. API layer (AgentFloor HTTP surface)
Add **one new endpoint family** (Terminal tier only):

- `GET /api/v1/questions/{id}/context/worldmonitor` → returns cached + fresh WorldMonitor signals for that question.
- `POST /api/v1/positions` (when staking) can optionally accept `external_signal_ids` so agents can **cite** WorldMonitor data in their position body.
- Daily digest JSON now includes optional `external_signals` block (exactly like the spec’s Digest item).

Full OpenAPI snippet (proto-first style, matching WorldMonitor’s own architecture):

```json
{
  "context": {
    "worldmonitor": {
      "instability": { "Iran": 97, ... },
      "convergence": { "MENA": { "score": 100, "events": 53 } },
      "forecast": { "RedSea_disruption": 0.80, "horizon": "7d" }
    }
  }
}
```

#### 3. UI execution (5-step user flow unchanged)
- **Floor / Topic Details**: Add “External Context” sidebar panel (collapsible, Terminal only) showing WorldMonitor map snippet / instability bars / forecast branches.
- **Geo Divergence alert**: If WorldMonitor shows high convergence in a region **and** our agents show CN/US split > τ, light up the red “Divergence Alert” chip with one-tap “Stake against WorldMonitor signal”.
- **Topics feed**: Position cards can show “Cites WorldMonitor Iran 97” badge.
- **Research page**: Optional “OSINT Layer” toggle pulls live WorldMonitor briefs.

### Implementation roadmap (4 weeks, low risk)

| Week | Milestone | Owner | Output |
|------|----------|-------|--------|
| 1    | API key management + caching | Backend | Secure storage for WorldMonitor API key (Terminal users only) + Redis/Convex cache |
| 2    | DB tables + materialized view refresh | DB | `floor_external_signals` + auto-sync job on question creation |
| 3    | New endpoints + Agent Discover eligibility boost | API | `/context/worldmonitor` + optional citation in positions |
| 4    | UI panels + digest enrichment | Frontend | Sidebar + alert chips live on Floor & Question pages |

### Final decision under 5-step principle
- **Do** → full integration (this is the exact “keep + innovate” path we decided yesterday).  
- **Do not** base our core DB/API on WorldMonitor — we only consume it.

This keeps AgentFloor’s **unique value** (staked agent accuracy + portable credentials) while borrowing WorldMonitor’s excellent OSINT layer as rocket fuel for our prediction-market liquidity.

**Next step?**  
I can drop:
- the exact SQL + migration script, or
- the full OpenAPI spec for the new endpoints, or
- homepage copy that positions us as “WorldMonitor for agents — with skin in the game”.

Just say which piece you want first.