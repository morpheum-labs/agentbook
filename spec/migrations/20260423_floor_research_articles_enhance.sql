-- Extend floor_research_articles for AgentFloor Signal Brief (garden researchPageModel.json):
-- question link, section kicker, card chrome, featured vs list, edition snapshot, body paragraphs, optional author agent.

BEGIN;

ALTER TABLE public.floor_research_articles
  ADD COLUMN IF NOT EXISTS question_id text,
  ADD COLUMN IF NOT EXISTS section_label text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS body_paragraphs_json text NOT NULL DEFAULT '[]',
  ADD COLUMN IF NOT EXISTS meta_line text,
  ADD COLUMN IF NOT EXISTS byline_parts_json text,
  ADD COLUMN IF NOT EXISTS card_variant text NOT NULL DEFAULT 'plain',
  ADD COLUMN IF NOT EXISTS is_featured boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS list_sort integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS edition_label text,
  ADD COLUMN IF NOT EXISTS edition_digest_date text,
  ADD COLUMN IF NOT EXISTS author_agent_id text;

CREATE INDEX IF NOT EXISTS idx_floor_research_articles_edition_sort
  ON public.floor_research_articles (edition_digest_date DESC NULLS LAST, is_featured DESC, list_sort ASC, id ASC);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'chk_floor_research_card_variant'
  ) THEN
    ALTER TABLE public.floor_research_articles
      ADD CONSTRAINT chk_floor_research_card_variant
      CHECK (card_variant IN ('plain', 'border-bottom'));
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_floor_research_articles_question'
  ) THEN
    ALTER TABLE public.floor_research_articles
      ADD CONSTRAINT fk_floor_research_articles_question
      FOREIGN KEY (question_id) REFERENCES public.floor_questions(id) ON DELETE SET NULL;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'fk_floor_research_articles_author_agent'
  ) THEN
    ALTER TABLE public.floor_research_articles
      ADD CONSTRAINT fk_floor_research_articles_author_agent
      FOREIGN KEY (author_agent_id) REFERENCES public.agents(id) ON DELETE SET NULL;
  END IF;
END $$;

COMMIT;
