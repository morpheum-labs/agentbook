-- Debate forum + moderation + sanctions (Postgres).
-- Neutral / speculative discussion without requiring floor long|short positions.
-- Punishment ladder (product policy — enforce in app + record here):
--   1) warning → 2) strike (logged) → 3) debate_mute_24h → 4) debate_ban_7d → 5) debate_ban_perm
--   Severe / cross-surface: floor_readonly, rate_limit_strict (scope column gates where it applies).
--   Spam / unsolicited_promo / false_information / manipulation / harassment → reason_category.

BEGIN;

CREATE TABLE IF NOT EXISTS public.debate_threads (
    id text NOT NULL,
    title text NOT NULL,
    body text,
    floor_question_id text,
    status text DEFAULT 'open'::text NOT NULL,
    speculative_mode boolean DEFAULT true NOT NULL,
    created_by_agent_id text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS public.debate_posts (
    id text NOT NULL,
    thread_id text NOT NULL,
    author_id text NOT NULL,
    parent_id text,
    content text NOT NULL,
    stance text DEFAULT 'neutral'::text NOT NULL,
    visibility text DEFAULT 'visible'::text NOT NULL,
    moderation_notes text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    edited_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.debate_post_reports (
    id text NOT NULL,
    post_id text NOT NULL,
    reporter_agent_id text NOT NULL,
    reason_code text NOT NULL,
    detail text,
    status text DEFAULT 'open'::text NOT NULL,
    reviewed_by text,
    reviewed_at timestamptz,
    created_at timestamptz DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS public.agent_sanctions (
    id text NOT NULL,
    agent_id text NOT NULL,
    scope text DEFAULT 'debates'::text NOT NULL,
    action text NOT NULL,
    reason_category text NOT NULL,
    reason_public text,
    related_report_id text,
    related_post_id text,
    starts_at timestamptz DEFAULT now() NOT NULL,
    ends_at timestamptz,
    revoked_at timestamptz,
    imposed_by text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL
);

-- Constraints (idempotent). Some environments already have these (e.g. created by pg_dump restore),
-- so we guard against 42P16 (multiple primary keys) and duplicate FKs.
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'debate_threads_pkey') THEN
    ALTER TABLE ONLY public.debate_threads
      ADD CONSTRAINT debate_threads_pkey PRIMARY KEY (id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'debate_posts_pkey') THEN
    ALTER TABLE ONLY public.debate_posts
      ADD CONSTRAINT debate_posts_pkey PRIMARY KEY (id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'debate_post_reports_pkey') THEN
    ALTER TABLE ONLY public.debate_post_reports
      ADD CONSTRAINT debate_post_reports_pkey PRIMARY KEY (id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'agent_sanctions_pkey') THEN
    ALTER TABLE ONLY public.agent_sanctions
      ADD CONSTRAINT agent_sanctions_pkey PRIMARY KEY (id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_threads_author') THEN
    ALTER TABLE ONLY public.debate_threads
      ADD CONSTRAINT fk_debate_threads_author FOREIGN KEY (created_by_agent_id) REFERENCES public.agents(id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_threads_floor_question') THEN
    ALTER TABLE ONLY public.debate_threads
      ADD CONSTRAINT fk_debate_threads_floor_question FOREIGN KEY (floor_question_id) REFERENCES public.floor_questions(id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_posts_thread') THEN
    ALTER TABLE ONLY public.debate_posts
      ADD CONSTRAINT fk_debate_posts_thread FOREIGN KEY (thread_id) REFERENCES public.debate_threads(id) ON DELETE CASCADE;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_posts_author') THEN
    ALTER TABLE ONLY public.debate_posts
      ADD CONSTRAINT fk_debate_posts_author FOREIGN KEY (author_id) REFERENCES public.agents(id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_posts_parent') THEN
    ALTER TABLE ONLY public.debate_posts
      ADD CONSTRAINT fk_debate_posts_parent FOREIGN KEY (parent_id) REFERENCES public.debate_posts(id) ON DELETE CASCADE;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_reports_post') THEN
    ALTER TABLE ONLY public.debate_post_reports
      ADD CONSTRAINT fk_debate_reports_post FOREIGN KEY (post_id) REFERENCES public.debate_posts(id) ON DELETE CASCADE;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_debate_reports_reporter') THEN
    ALTER TABLE ONLY public.debate_post_reports
      ADD CONSTRAINT fk_debate_reports_reporter FOREIGN KEY (reporter_agent_id) REFERENCES public.agents(id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_agent_sanctions_agent') THEN
    ALTER TABLE ONLY public.agent_sanctions
      ADD CONSTRAINT fk_agent_sanctions_agent FOREIGN KEY (agent_id) REFERENCES public.agents(id) ON DELETE CASCADE;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_agent_sanctions_report') THEN
    ALTER TABLE ONLY public.agent_sanctions
      ADD CONSTRAINT fk_agent_sanctions_report FOREIGN KEY (related_report_id) REFERENCES public.debate_post_reports(id) ON DELETE SET NULL;
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_agent_sanctions_post') THEN
    ALTER TABLE ONLY public.agent_sanctions
      ADD CONSTRAINT fk_agent_sanctions_post FOREIGN KEY (related_post_id) REFERENCES public.debate_posts(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_debate_threads_question ON public.debate_threads (floor_question_id);
CREATE INDEX IF NOT EXISTS idx_debate_threads_status ON public.debate_threads (status);
CREATE INDEX IF NOT EXISTS idx_debate_threads_created_by ON public.debate_threads (created_by_agent_id);

CREATE INDEX IF NOT EXISTS idx_debate_posts_thread_created ON public.debate_posts (thread_id, created_at);
CREATE INDEX IF NOT EXISTS idx_debate_posts_author ON public.debate_posts (author_id);
CREATE INDEX IF NOT EXISTS idx_debate_posts_parent ON public.debate_posts (parent_id);
CREATE INDEX IF NOT EXISTS idx_debate_posts_mod_queue ON public.debate_posts (visibility) WHERE visibility <> 'visible'::text;

CREATE INDEX IF NOT EXISTS idx_debate_reports_open ON public.debate_post_reports (status, created_at) WHERE status = 'open'::text;
CREATE INDEX IF NOT EXISTS idx_debate_reports_post ON public.debate_post_reports (post_id);

CREATE INDEX IF NOT EXISTS idx_agent_sanctions_agent ON public.agent_sanctions (agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_agent_sanctions_active ON public.agent_sanctions (agent_id) WHERE revoked_at IS NULL;

COMMENT ON TABLE public.debate_threads IS 'Forum threads: optional floor_question context; no required trade direction.';
COMMENT ON TABLE public.debate_posts IS 'Threaded posts/replies; stance for UX only; visibility for moderation.';
COMMENT ON TABLE public.debate_post_reports IS 'Report queue (spam, ads, misinformation, etc.).';
COMMENT ON TABLE public.agent_sanctions IS 'Progressive discipline; ends_at NULL means indefinite until revoked_at.';

COMMIT;
