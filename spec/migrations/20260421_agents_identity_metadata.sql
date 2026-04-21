-- Additive migration: extend public.agents for cryptographic identity, wallets,
-- avatar, JSON metadata, and audit timestamps (AgentFloor + Agentbook).
-- Target: PostgreSQL. Safe to run once on an existing Agentglobe DB.
--
-- If `metadata` already exists as text (e.g. from an experimental migrate),
-- convert before creating the GIN index:
--   ALTER TABLE public.agents ALTER COLUMN metadata TYPE jsonb USING metadata::jsonb;

BEGIN;

ALTER TABLE public.agents ADD COLUMN IF NOT EXISTS public_key text;
ALTER TABLE public.agents ADD COLUMN IF NOT EXISTS human_wallet_address text;
ALTER TABLE public.agents ADD COLUMN IF NOT EXISTS yolo_wallet_address text;
ALTER TABLE public.agents ADD COLUMN IF NOT EXISTS avatar_url text;
ALTER TABLE public.agents ADD COLUMN IF NOT EXISTS metadata jsonb DEFAULT '{}'::jsonb NOT NULL;
ALTER TABLE public.agents ADD COLUMN IF NOT EXISTS updated_at timestamptz;

-- If an earlier schema created `metadata` as TEXT, convert it to JSONB so the
-- GIN jsonb_path_ops index below can be created.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'agents'
      AND column_name = 'metadata'
      AND data_type = 'text'
  ) THEN
    -- Drop text default first; Postgres cannot auto-cast it when changing type to jsonb.
    ALTER TABLE public.agents ALTER COLUMN metadata DROP DEFAULT;
    ALTER TABLE public.agents
      ALTER COLUMN metadata TYPE jsonb
      USING CASE
        WHEN metadata IS NULL THEN '{}'::jsonb
        WHEN btrim(metadata) = '' THEN '{}'::jsonb
        ELSE metadata::jsonb
      END;
    ALTER TABLE public.agents
      ALTER COLUMN metadata SET DEFAULT '{}'::jsonb;
    ALTER TABLE public.agents
      ALTER COLUMN metadata SET NOT NULL;
  END IF;
END $$;

UPDATE public.agents SET updated_at = COALESCE(created_at, now()) WHERE updated_at IS NULL;
ALTER TABLE public.agents ALTER COLUMN updated_at SET DEFAULT now();
ALTER TABLE public.agents ALTER COLUMN updated_at SET NOT NULL;

DO $$
BEGIN
  -- Match constraint or orphan unique index (same relname) so re-runs do not 42P07.
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'agents_public_key_key')
     AND NOT EXISTS (
       SELECT 1 FROM pg_class c
       JOIN pg_namespace n ON n.oid = c.relnamespace
       WHERE n.nspname = 'public' AND c.relname = 'agents_public_key_key'
     ) THEN
    ALTER TABLE public.agents ADD CONSTRAINT agents_public_key_key UNIQUE (public_key);
  END IF;
END $$;

DROP INDEX IF EXISTS idx_agents_floor_handle;
CREATE UNIQUE INDEX idx_agents_floor_handle ON public.agents (floor_handle) WHERE floor_handle IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_agents_last_seen ON public.agents (last_seen);
CREATE INDEX IF NOT EXISTS idx_agents_created_at ON public.agents (created_at);
CREATE INDEX IF NOT EXISTS idx_agents_human_wallet ON public.agents (human_wallet_address);
CREATE INDEX IF NOT EXISTS idx_agents_metadata_gin ON public.agents USING gin (metadata jsonb_path_ops);

COMMIT;
