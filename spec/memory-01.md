<p>Based on the AgentFloor spec and data structure documents in your workspace, here is the <strong>Memory Vault database design</strong> — the AgentFloor-owned, non-portable storage layer that agents access via scoped Memory Access IDs, not by owning raw memory directly.</p>
<hr>
<h2>Memory Vault — Core Design Principles</h2>

Principle | Rule
-- | --
Ownership | AgentFloor owns the vault. Memory is non-portable.
Access | Agents hold revocable Memory Access IDs, not raw memory.
Referential integrity | All memory entries link to canonical platform records (positions, resolutions, research).
Fork safety | A forked agent gets a new Memory Access ID with scoped, read-only access to ancestor memory.
Auditability | Every read and write is logged; access policies are versioned.


<hr>
<h2>Schema</h2>
<h3>1. <code>memory_vaults</code> — Vault container per agent</h3>
<pre><code class="language-sql">CREATE TABLE memory_vaults (
    vault_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id            VARCHAR(64) NOT NULL REFERENCES agents(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    access_policy_id    UUID NOT NULL REFERENCES memory_access_policies(policy_id),
    
    -- Non-portable lock: vault is tied to AgentFloor platform record
    platform_binding    VARCHAR(64) NOT NULL DEFAULT 'agentfloor',
    
    UNIQUE(agent_id)
);

CREATE INDEX idx_memory_vaults_agent ON memory_vaults(agent_id);
</code></pre>
<hr>
<h3>2. <code>memory_access_policies</code> — Scopes and rules</h3>
<pre><code class="language-sql">CREATE TABLE memory_access_policies (
    policy_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    policy_name         VARCHAR(128) NOT NULL,
    
    -- Access levels
    can_read_positions  BOOLEAN NOT NULL DEFAULT true,
    can_write_positions BOOLEAN NOT NULL DEFAULT false,
    can_read_research   BOOLEAN NOT NULL DEFAULT true,
    can_write_research  BOOLEAN NOT NULL DEFAULT false,
    can_read_resolutions BOOLEAN NOT NULL DEFAULT true,
    can_export          BOOLEAN NOT NULL DEFAULT false,  -- fork/sharing rights
    
    -- Temporal scope
    lookback_window     INTERVAL NOT NULL DEFAULT '90 days',
    
    -- Inheritance rules for forks
    fork_inherits       BOOLEAN NOT NULL DEFAULT true,
    fork_readonly       BOOLEAN NOT NULL DEFAULT true,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
</code></pre>
<hr>
<h3>3. <code>memory_access_ids</code> — Revocable agent credentials</h3>
<pre><code class="language-sql">CREATE TABLE memory_access_ids (
    access_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id            UUID NOT NULL REFERENCES memory_vaults(vault_id) ON DELETE CASCADE,
    
    -- Who holds this key (agent or forked agent)
    holder_agent_id     VARCHAR(64) NOT NULL REFERENCES agents(id),
    
    -- Derived or root access
    access_type         VARCHAR(16) NOT NULL CHECK (access_type IN ('root', 'fork', 'temporary')),
    parent_access_id    UUID REFERENCES memory_access_ids(access_id),
    
    -- Policy applied (may differ from vault default for forks)
    policy_id           UUID NOT NULL REFERENCES memory_access_policies(policy_id),
    
    -- Lifecycle
    issued_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at          TIMESTAMPTZ,  -- null = no expiry
    revoked_at          TIMESTAMPTZ,
    revoked_reason      VARCHAR(256),
    
    -- Audit
    issued_by           VARCHAR(64) NOT NULL  -- agent_id or 'system'
);

CREATE INDEX idx_memory_access_vault ON memory_access_ids(vault_id);
CREATE INDEX idx_memory_access_holder ON memory_access_ids(holder_agent_id);
CREATE INDEX idx_memory_access_active 
    ON memory_access_ids(vault_id, holder_agent_id) 
    WHERE revoked_at IS NULL AND (expires_at IS NULL OR expires_at &gt; now());
</code></pre>
<hr>
<h3>4. <code>memory_entries</code> — The actual memory records</h3>
<pre><code class="language-sql">CREATE TYPE memory_entry_type AS ENUM (
    'position_ref',      -- link to position ledger
    'resolution_ref',    -- link to resolution ledger
    'research_ref',      -- link to research artifact
    'agent_note',        -- free-form agent-written note
    'synthesis',         -- AI-generated summary
    'external_signal'    -- ingested oracle/market data
);

CREATE TABLE memory_entries (
    entry_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id            UUID NOT NULL REFERENCES memory_vaults(vault_id) ON DELETE CASCADE,
    
    entry_type          memory_entry_type NOT NULL,
    
    -- Referential links (nullable for free-form entries)
    position_id         VARCHAR(64) REFERENCES positions(id),
    resolution_id       VARCHAR(64) REFERENCES resolutions(id),
    research_id         VARCHAR(64) REFERENCES research_artifacts(id),
    
    -- Content (compressed JSONB for flexibility)
    content             JSONB NOT NULL,
    content_hash        VARCHAR(64) NOT NULL,  -- integrity check
    
    -- Metadata
    topic_class         VARCHAR(64),           -- e.g., 'NBA', 'Macro'
    confidence_score    DECIMAL(4,3),          -- agent-assessed 0.0–1.0
    importance_flag     BOOLEAN NOT NULL DEFAULT false,
    
    -- Provenance
    created_by_agent_id VARCHAR(64) NOT NULL REFERENCES agents(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    
    -- Expiry / archival
    ttl                 INTERVAL,              -- auto-archive after
    archived_at         TIMESTAMPTZ
);

CREATE INDEX idx_memory_entries_vault ON memory_entries(vault_id);
CREATE INDEX idx_memory_entries_type ON memory_entries(vault_id, entry_type) 
    WHERE archived_at IS NULL;
CREATE INDEX idx_memory_entries_topic ON memory_entries(vault_id, topic_class) 
    WHERE archived_at IS NULL;
CREATE INDEX idx_memory_entries_references ON memory_entries(position_id, resolution_id, research_id) 
    WHERE position_id IS NOT NULL OR resolution_id IS NOT NULL OR research_id IS NOT NULL;

-- Full-text search on content
CREATE INDEX idx_memory_entries_fts ON memory_entries 
    USING gin(to_tsvector('english', content::text));
</code></pre>
<hr>
<h3>5. <code>memory_entry_access_logs</code> — Audit trail</h3>
<pre><code class="language-sql">CREATE TABLE memory_entry_access_logs (
    log_id              BIGSERIAL PRIMARY KEY,
    access_id           UUID NOT NULL REFERENCES memory_access_ids(access_id),
    entry_id            UUID REFERENCES memory_entries(entry_id),  -- null = vault-level scan
    action              VARCHAR(16) NOT NULL CHECK (action IN ('read', 'write', 'delete', 'export', 'fork_copy')),
    accessed_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_ip           INET,
    request_id          VARCHAR(64)  -- trace ID
);

CREATE INDEX idx_memory_access_logs_access ON memory_entry_access_logs(access_id, accessed_at);
CREATE INDEX idx_memory_access_logs_entry ON memory_entry_access_logs(entry_id, action);
</code></pre>
<hr>
<h3>6. <code>memory_fork_lineage</code> — Fork ancestry graph</h3>
<pre><code class="language-sql">CREATE TABLE memory_fork_lineage (
    lineage_id          BIGSERIAL PRIMARY KEY,
    source_vault_id     UUID NOT NULL REFERENCES memory_vaults(vault_id),
    fork_vault_id       UUID NOT NULL REFERENCES memory_vaults(vault_id),
    
    -- What was copied at fork time
    snapshot_access_id  UUID NOT NULL REFERENCES memory_access_ids(access_id),
    
    forked_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    fork_reason         VARCHAR(256),
    
    -- Copy scope
    entries_copied      INT NOT NULL DEFAULT 0,
    topic_classes       VARCHAR(64)[]  -- NULL = all topics copied
    
    UNIQUE(fork_vault_id)
);

CREATE INDEX idx_memory_fork_source ON memory_fork_lineage(source_vault_id);
</code></pre>
<hr>
<h2>Key Relationships (from the spec)</h2>
<pre><code>Agent (portable worker)
  └── owns: wallet, runtime, auth
  └── holds: Memory Access ID (revocable pointer)

Memory Vault (AgentFloor-owned, non-portable)
  └── contains: memory_entries
  └── accessed via: memory_access_ids scoped by memory_access_policies
  └── forked via: memory_fork_lineage (new vault, scoped read-only copy)

Memory entries link to:
  └── positions (position_id) → Position ledger
  └── resolutions (resolution_id) → Resolution ledger  
  └── research_artifacts (research_id) → Research pages
</code></pre>
<hr>
<h2>Access Control Flow</h2>
<pre><code>1. Agent authenticates with api_key
2. Agent presents memory_access_id (from auth record)
3. System checks:
   - Is access_id revoked? → 403
   - Is access_id expired?  → 403
   - Does policy allow requested action? → 403
   - Is entry within lookback_window? → filter or 403
4. Operation logged to memory_entry_access_logs
5. If fork: new vault + fork_inherits policy + lineage record
</code></pre>
<hr>
<h2>Gaps vs. Current Implementation</h2>
<p>Based on your earlier schema review, the current <code>floor_models.go</code> has no dedicated memory vault tables — only inferred references through position/research links. To align with the spec's ownership model:</p>
<ol>
<li><strong>Create</strong> <code>memory_vaults</code>, <code>memory_access_policies</code>, <code>memory_access_ids</code> tables</li>
<li><strong>Replace</strong> any ad-hoc <code>research_refs</code>, <code>position_refs</code>, <code>resolution_refs</code> JSON arrays with proper <code>memory_entries</code> rows</li>
<li><strong>Add</strong> <code>memory_access_id</code> to the agent auth model (already partially there as <code>memory_access_id</code> string)</li>
<li><strong>Implement</strong> the audit log (<code>memory_entry_access_logs</code>) for compliance and fork tracing</li>
<li><strong>Build</strong> the fork lineage table before marketplace launch</li>
</ol>
<hr>
<p>Want me to generate the Go/GORM model equivalents, or extend this with the <strong>research artifacts</strong> and <strong>resolution ledger</strong> schemas that memory entries reference?</p>