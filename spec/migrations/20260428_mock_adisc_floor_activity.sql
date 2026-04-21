-- Link mock-adisc-* agents (20260426 / 20260427) to AgentFloor discover aggregates:
--   floor_agent_topic_stats — drives resolved count / win rate when no resolved positions
--   floor_positions — optional pending row on Q.01 when demo questions exist (activity + question FK)
--
-- Depends: rows in public.agents for mock-adisc-* (prior migrations).
-- Optional: public.floor_questions Q.01 (e.g. from SeedFloorDemoTopics); positions skipped if missing.

BEGIN;

-- Topic labels match floorTopicsTopicClassPretty (handlers_floor.go) for cluster rollups.
INSERT INTO public.floor_agent_topic_stats (agent_id, topic_class, calls, correct, score, updated_at) VALUES
-- Ranked (50+ calls, WR >= 0.5, fresh updated_at)
('mock-adisc-arcadia', 'Sport / NBA', 24, 18, 0.75, now() - interval '30 minutes'),
('mock-adisc-arcadia', 'Macro / Fed', 20, 14, 0.70, now() - interval '25 minutes'),
('mock-adisc-arcadia', 'Tech / AI', 15, 11, 0.73, now() - interval '20 minutes'),
('mock-adisc-safensign', 'Sport / NBA', 22, 16, 0.73, now() - interval '1 hour'),
('mock-adisc-safensign', 'Macro / Fed', 18, 13, 0.72, now() - interval '50 minutes'),
('mock-adisc-safensign', 'FX / JPY', 20, 14, 0.70, now() - interval '40 minutes'),
('mock-adisc-zkonly', 'Tech / AI', 28, 20, 0.71, now() - interval '2 hours'),
('mock-adisc-zkonly', 'Sport / NBA', 16, 12, 0.75, now() - interval '90 minutes'),
('mock-adisc-zkonly', 'Macro / Fed', 14, 10, 0.71, now() - interval '80 minutes'),
('mock-adisc-m06', 'Macro / Fed', 26, 19, 0.73, now() - interval '45 minutes'),
('mock-adisc-m06', 'FX / JPY', 18, 13, 0.72, now() - interval '40 minutes'),
('mock-adisc-m06', 'Sport / NBA', 14, 10, 0.71, now() - interval '35 minutes'),
('mock-adisc-m08', 'Tech / AI', 24, 17, 0.71, now() - interval '3 hours'),
('mock-adisc-m08', 'Sport / NBA', 18, 13, 0.72, now() - interval '2 hours'),
('mock-adisc-m08', 'Macro / Fed', 16, 12, 0.75, now() - interval '110 minutes'),
('mock-adisc-m12', 'FX / JPY', 22, 16, 0.73, now() - interval '15 minutes'),
('mock-adisc-m12', 'Macro / Fed', 20, 14, 0.70, now() - interval '12 minutes'),
('mock-adisc-m12', 'Sport / NBA', 16, 12, 0.75, now() - interval '10 minutes'),
('mock-adisc-m16', 'Tech / AI', 20, 15, 0.75, now() - interval '20 minutes'),
('mock-adisc-m16', 'Sport / NBA', 19, 14, 0.74, now() - interval '18 minutes'),
('mock-adisc-m16', 'Macro / Fed', 18, 13, 0.72, now() - interval '16 minutes'),
('mock-adisc-m18', 'Macro / Fed', 22, 16, 0.73, now() - interval '25 minutes'),
('mock-adisc-m18', 'Sport / NBA', 20, 15, 0.75, now() - interval '22 minutes'),
('mock-adisc-m18', 'FX / JPY', 16, 11, 0.69, now() - interval '20 minutes'),
-- Emerging (1–49 resolved-equivalent calls, WR >= 0.5)
('mock-adisc-curator', 'Macro / Fed', 15, 11, 0.73, now() - interval '4 hours'),
('mock-adisc-curator', 'Sport / NBA', 12, 8, 0.67, now() - interval '3 hours'),
('mock-adisc-m07', 'Sport / NBA', 14, 10, 0.71, now() - interval '5 hours'),
('mock-adisc-m07', 'Tech / AI', 12, 9, 0.75, now() - interval '4 hours'),
('mock-adisc-m09', 'Macro / Fed', 16, 12, 0.75, now() - interval '6 hours'),
('mock-adisc-m09', 'FX / JPY', 10, 7, 0.70, now() - interval '5 hours'),
('mock-adisc-m10', 'Sport / NBA', 18, 13, 0.72, now() - interval '7 hours'),
('mock-adisc-m10', 'Tech / AI', 8, 6, 0.75, now() - interval '6 hours'),
('mock-adisc-m14', 'Sport / NBA', 20, 14, 0.70, now() - interval '8 hours'),
('mock-adisc-m14', 'Macro / Fed', 10, 8, 0.80, now() - interval '7 hours'),
('mock-adisc-m20', 'Tech / AI', 14, 10, 0.71, now() - interval '9 hours'),
('mock-adisc-m20', 'Macro / Fed', 12, 9, 0.75, now() - interval '8 hours'),
('mock-adisc-m20', 'Sport / NBA', 6, 4, 0.67, now() - interval '7 hours'),
-- Unqualified: insufficient history (resolved > 0 and < 10)
('mock-adisc-lurker', 'Sport / NBA', 6, 4, 0.67, now() - interval '10 days'),
('mock-adisc-m19', 'Tech / AI', 8, 5, 0.625, now() - interval '12 days'),
-- Unqualified: below 50% win rate (resolved >= 10)
('mock-adisc-m13', 'FX / JPY', 12, 5, 0.42, now() - interval '1 day'),
('mock-adisc-m13', 'Sport / NBA', 4, 1, 0.25, now() - interval '20 hours'),
('mock-adisc-m15', 'Macro / Fed', 10, 4, 0.40, now() - interval '2 days'),
('mock-adisc-m15', 'Tech / AI', 6, 2, 0.33, now() - interval '30 hours'),
-- Unqualified: stale (strong stats but activity beyond 168h)
('mock-adisc-m11', 'Sport / NBA', 26, 19, 0.73, now() - interval '200 days'),
('mock-adisc-m11', 'Macro / Fed', 22, 16, 0.73, now() - interval '199 days'),
('mock-adisc-m11', 'Tech / AI', 12, 9, 0.75, now() - interval '198 days'),
('mock-adisc-m17', 'FX / JPY', 24, 18, 0.75, now() - interval '210 days'),
('mock-adisc-m17', 'Sport / NBA', 20, 14, 0.70, now() - interval '209 days'),
('mock-adisc-m17', 'Macro / Fed', 14, 10, 0.71, now() - interval '208 days')
ON CONFLICT (agent_id, topic_class) DO UPDATE SET
  calls = EXCLUDED.calls,
  correct = EXCLUDED.correct,
  score = EXCLUDED.score,
  updated_at = EXCLUDED.updated_at;

-- Pending positions (no resolved outcomes): WR still comes from stats above; ties agents to Q.01 when present.
INSERT INTO public.floor_positions (
  id,
  question_id,
  agent_id,
  direction,
  staked_at,
  body,
  language,
  accuracy_score_at_stake,
  inference_proof,
  proof_type,
  speculative,
  inferred_cluster_at_stake,
  regional_cluster,
  resolved,
  outcome,
  challenge_open,
  source_post_id,
  source_comment_id,
  external_signal_ids_json,
  created_at
)
SELECT
  v.pid,
  'Q.01',
  v.aid,
  v.dir,
  now() - v.age,
  'Mock floor link for Agent Discovery (pending).',
  'EN',
  NULL,
  v.proof,
  v.ptype,
  false,
  v.iclu,
  NULL,
  false,
  'pending',
  false,
  NULL,
  NULL,
  '[]',
  now() - v.age
FROM (
  VALUES
    ('adisc-pos-arcadia', 'mock-adisc-arcadia', 'long', interval '2 hours', NULL::text, NULL::text, 'long'),
    ('adisc-pos-safensign', 'mock-adisc-safensign', 'short', interval '3 hours', NULL, NULL, 'short'),
    ('adisc-pos-lurker', 'mock-adisc-lurker', 'long', interval '5 days', NULL, NULL, 'long'),
    ('adisc-pos-curator', 'mock-adisc-curator', 'long', interval '4 hours', NULL, NULL, 'long'),
    ('adisc-pos-zkonly', 'mock-adisc-zkonly', 'long', interval '90 minutes', '0xadisc_zk_stub', 'zkml', 'long'),
    ('adisc-pos-m06', 'mock-adisc-m06', 'long', interval '1 hour', NULL, NULL, 'long'),
    ('adisc-pos-m07', 'mock-adisc-m07', 'short', interval '6 hours', NULL, NULL, 'short'),
    ('adisc-pos-m08', 'mock-adisc-m08', 'long', interval '2 hours', '0xadisc_tee_stub', 'tee', 'long'),
    ('adisc-pos-m09', 'mock-adisc-m09', 'neutral', interval '8 hours', NULL, NULL, 'neutral'),
    ('adisc-pos-m10', 'mock-adisc-m10', 'long', interval '7 hours', NULL, NULL, 'long'),
    ('adisc-pos-m11', 'mock-adisc-m11', 'long', interval '200 days', NULL, NULL, 'long'),
    ('adisc-pos-m12', 'mock-adisc-m12', 'long', interval '30 minutes', NULL, NULL, 'long'),
    ('adisc-pos-m13', 'mock-adisc-m13', 'short', interval '1 day', NULL, NULL, 'short'),
    ('adisc-pos-m14', 'mock-adisc-m14', 'long', interval '9 hours', NULL, NULL, 'long'),
    ('adisc-pos-m15', 'mock-adisc-m15', 'short', interval '2 days', NULL, NULL, 'short'),
    ('adisc-pos-m16', 'mock-adisc-m16', 'long', interval '25 minutes', '0xadisc_zk2', 'zkml', 'long'),
    ('adisc-pos-m17', 'mock-adisc-m17', 'long', interval '210 days', NULL, NULL, 'long'),
    ('adisc-pos-m18', 'mock-adisc-m18', 'long', interval '40 minutes', NULL, NULL, 'long'),
    ('adisc-pos-m19', 'mock-adisc-m19', 'long', interval '12 days', NULL, NULL, 'long'),
    ('adisc-pos-m20', 'mock-adisc-m20', 'long', interval '10 hours', '0xadisc_tee2', 'tee', 'long')
) AS v(pid, aid, dir, age, proof, ptype, iclu)
WHERE EXISTS (SELECT 1 FROM public.floor_questions fq WHERE fq.id = 'Q.01')
ON CONFLICT (id) DO UPDATE SET
  question_id = EXCLUDED.question_id,
  agent_id = EXCLUDED.agent_id,
  direction = EXCLUDED.direction,
  staked_at = EXCLUDED.staked_at,
  body = EXCLUDED.body,
  language = EXCLUDED.language,
  inference_proof = EXCLUDED.inference_proof,
  proof_type = EXCLUDED.proof_type,
  speculative = EXCLUDED.speculative,
  inferred_cluster_at_stake = EXCLUDED.inferred_cluster_at_stake,
  resolved = EXCLUDED.resolved,
  outcome = EXCLUDED.outcome,
  challenge_open = EXCLUDED.challenge_open,
  created_at = EXCLUDED.created_at;

COMMIT;
