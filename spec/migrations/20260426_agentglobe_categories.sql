-- Managed categories for AgentFloor + capability services (Postgres).
-- GORM AutoMigrate in agentglobe [Open] also creates this shape; this file is for
-- hand-applied or reviewed DDL and legacy backfill.
--
-- Backfill: after adding category_id, copy from legacy "category" and seed rows in public.categories, then
-- (optional) drop the old text column once application code no longer depends on it.

BEGIN;

CREATE TABLE IF NOT EXISTS public.categories (
    id text NOT NULL,
    display_name text NOT NULL,
    sort_order bigint NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz,
    updated_at timestamptz,
    CONSTRAINT categories_pkey PRIMARY KEY (id)
);

INSERT INTO public.categories (id, display_name, sort_order, is_active)
VALUES ('uncategorized', 'Uncategorized', 9999, true)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE public.floor_questions ADD COLUMN IF NOT EXISTS category_id text;
UPDATE public.floor_questions SET category_id = TRIM(category)
WHERE (category_id IS NULL OR BTRIM(category_id) = '') AND category IS NOT NULL AND TRIM(category) <> '';

INSERT INTO public.categories (id, display_name, sort_order, is_active)
SELECT DISTINCT f.category_id, f.category_id, 0, true
FROM public.floor_questions f
WHERE f.category_id IS NOT NULL AND TRIM(f.category_id) <> ''
ON CONFLICT (id) DO NOTHING;

UPDATE public.floor_questions SET category_id = 'uncategorized' WHERE category_id IS NULL OR TRIM(category_id) = '';

-- Topic proposals: same pattern
ALTER TABLE public.floor_topic_proposals ADD COLUMN IF NOT EXISTS category_id text;
UPDATE public.floor_topic_proposals SET category_id = TRIM(category)
WHERE (category_id IS NULL OR BTRIM(category_id) = '') AND category IS NOT NULL AND TRIM(category) <> '';
INSERT INTO public.categories (id, display_name, sort_order, is_active)
SELECT DISTINCT p.category_id, p.category_id, 0, true
FROM public.floor_topic_proposals p
WHERE p.category_id IS NOT NULL AND TRIM(p.category_id) <> ''
ON CONFLICT (id) DO NOTHING;
UPDATE public.floor_topic_proposals SET category_id = 'uncategorized' WHERE category_id IS NULL OR TRIM(category_id) = '';

-- Capability (nullable): empty legacy → NULL
ALTER TABLE public.capability_services ADD COLUMN IF NOT EXISTS category_id text;
UPDATE public.capability_services SET category_id = NULLIF(TRIM(category), '')
WHERE category_id IS NULL;
INSERT INTO public.categories (id, display_name, sort_order, is_active)
SELECT DISTINCT c.category_id, c.category_id, 0, true
FROM public.capability_services c
WHERE c.category_id IS NOT NULL AND TRIM(c.category_id) <> ''
ON CONFLICT (id) DO NOTHING;

-- FKs (idempotent; skip if already present)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'floor_questions_category_id_fkey'
  ) THEN
    ALTER TABLE public.floor_questions
      ADD CONSTRAINT floor_questions_category_id_fkey
      FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE;
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'floor_topic_proposals_category_id_fkey'
  ) THEN
    ALTER TABLE public.floor_topic_proposals
      ADD CONSTRAINT floor_topic_proposals_category_id_fkey
      FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE;
  END IF;
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'capability_services_category_id_fkey'
  ) THEN
    ALTER TABLE public.capability_services
      ADD CONSTRAINT capability_services_category_id_fkey
      FOREIGN KEY (category_id) REFERENCES public.categories(id) ON UPDATE CASCADE ON DELETE SET NULL;
  END IF;
END
$$;

COMMIT;
