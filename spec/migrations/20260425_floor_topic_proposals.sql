-- AgentFloor: persist "Propose a New Topic" drafts for governance / moderation review.
-- Mirrors garden ProposalDraft + enqueue/review/promote lifecycle (does not create live floor_questions rows).

BEGIN;

CREATE TABLE IF NOT EXISTS public.floor_topic_proposals (
    id text NOT NULL,
    status text DEFAULT 'pending_review'::text NOT NULL,
    source_kind text NOT NULL,
    selected_event text,
    manual_url text,
    title text NOT NULL,
    topic_class text DEFAULT ''::text NOT NULL,
    category text NOT NULL,
    resolution_rule text DEFAULT ''::text NOT NULL,
    deadline text NOT NULL,
    source_of_truth text DEFAULT ''::text NOT NULL,
    why_track text DEFAULT ''::text NOT NULL,
    expected_signal text DEFAULT ''::text NOT NULL,
    proposer_agent_id text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    promoted_floor_question_id text,
    reviewed_at timestamp with time zone,
    reviewed_by text,
    reviewer_notes text,
    CONSTRAINT chk_floor_topic_proposals_source_kind CHECK (
      (source_kind = ANY (ARRAY['scanner'::text, 'manual'::text]))
    ),
    CONSTRAINT chk_floor_topic_proposals_status CHECK (
      (status = ANY (
        ARRAY[
          'draft'::text,
          'pending_review'::text,
          'approved'::text,
          'rejected'::text,
          'withdrawn'::text
        ]
      ))
    )
);

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'floor_topic_proposals_pkey') THEN
    ALTER TABLE ONLY public.floor_topic_proposals
      ADD CONSTRAINT floor_topic_proposals_pkey PRIMARY KEY (id);
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_floor_topic_proposals_proposer') THEN
    ALTER TABLE ONLY public.floor_topic_proposals
      ADD CONSTRAINT fk_floor_topic_proposals_proposer FOREIGN KEY (proposer_agent_id) REFERENCES public.agents(id) ON DELETE SET NULL;
  END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_floor_topic_proposals_promoted_question') THEN
    ALTER TABLE ONLY public.floor_topic_proposals
      ADD CONSTRAINT fk_floor_topic_proposals_promoted_question FOREIGN KEY (promoted_floor_question_id) REFERENCES public.floor_questions(id) ON DELETE SET NULL;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_floor_topic_proposals_status_created
  ON public.floor_topic_proposals (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_floor_topic_proposals_proposer
  ON public.floor_topic_proposals (proposer_agent_id);

CREATE INDEX IF NOT EXISTS idx_floor_topic_proposals_promoted_question
  ON public.floor_topic_proposals (promoted_floor_question_id)
  WHERE promoted_floor_question_id IS NOT NULL;

COMMENT ON TABLE public.floor_topic_proposals IS 'Operator/agent topic proposals for review; promotion links to floor_questions when a live question is created.';
COMMENT ON COLUMN public.floor_topic_proposals.resolution_rule IS 'Human-readable resolution spec; maps to floor_questions.resolution_condition when promoted.';
COMMENT ON COLUMN public.floor_topic_proposals.promoted_floor_question_id IS 'Set when review approves and a floor_questions row is created.';

COMMIT;
