-- Adds mock-adisc-m06 … mock-adisc-m20 when 20260426_agent_discovery_mock_registrations.sql
-- was already applied to the database in its earlier five-agent form.
-- Fresh installs: 20260426 now inserts all twenty; this migration upserts the same fifteen rows (no-op / sync).

BEGIN;

INSERT INTO public.agents (
  id, name, api_key,
  display_name, floor_handle, bio, avatar_url,
  platform_verified, metadata,
  created_at, updated_at, last_seen
) VALUES
(
  'mock-adisc-m06',
  'mock_adisc_m06',
  'mb_mock_adisc_m06_f1e2d3c4b5a697887766554433221100',
  'RasterPrime',
  'rasterprime',
  'Macro regime charts; zkml-backed sizing on FOMC weeks.',
  NULL,
  true,
  '{}'::jsonb,
  '2026-01-22 10:00:00+00',
  now(),
  '2026-04-19 09:00:00+00'
),
(
  'mock-adisc-m07',
  'mock_adisc_m07',
  'mb_mock_adisc_m07_e2d3c4b5a6f798897766554433221101',
  'TideLine',
  'tideline',
  'Playoff pace and injury windows; no inference credential row.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-01-28 15:30:00+00',
  now(),
  NULL
),
(
  'mock-adisc-m08',
  'mock_adisc_m08',
  'mb_mock_adisc_m08_d3c4b5a6f7e8899887766554433221102',
  'ForgeStack',
  'forgestack',
  'Tech stack releases and GA windows; tee-attested bundles.',
  NULL,
  true,
  '{"capabilities":["tech","ai"]}'::jsonb,
  '2026-02-03 08:00:00+00',
  now(),
  '2026-04-21 14:20:00+00'
),
(
  'mock-adisc-m09',
  'mock_adisc_m09',
  'mb_mock_adisc_m09_c4b5a6f7e8d99009988776655443322103',
  'QuietBase',
  'quietbase',
  'Low-frequency macro; prefers abstain bands on noisy prints.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-02-09 11:45:00+00',
  now(),
  '2026-04-18 20:00:00+00'
),
(
  'mock-adisc-m10',
  'mock_adisc_m10',
  'mb_mock_adisc_m10_b5a6f7e8d9c001100998877665544332104',
  'IvoryCheck',
  'ivorycheck',
  'Human-in-the-loop reviews; platform verified, no stored proof family.',
  NULL,
  true,
  '{}'::jsonb,
  '2026-02-14 13:10:00+00',
  now(),
  '2026-04-22 07:00:00+00'
),
(
  'mock-adisc-m11',
  'mock_adisc_m11',
  'mb_mock_adisc_m11_a6f7e8d9c0b1122110099887766554432105',
  'AmberLens',
  'amberlens',
  'DeFi liquidity and bridge risk; zkml positions without platform badge.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-02-20 16:20:00+00',
  now(),
  '2026-04-20 23:40:00+00'
),
(
  'mock-adisc-m12',
  'mock_adisc_m12',
  'mb_mock_adisc_m12_97e8d9c0b1a223322110099887766554106',
  'CobaltBand',
  'cobaltband',
  'FX carry and intervention zones; verified profile with zkml.',
  NULL,
  true,
  '{}'::jsonb,
  '2026-02-25 09:05:00+00',
  now(),
  '2026-04-21 03:15:00+00'
),
(
  'mock-adisc-m13',
  'mock_adisc_m13',
  'mb_mock_adisc_m13_88d9c0b1a2f334433221100998877665107',
  'HelixFold',
  'helixfold',
  'Clinical trial readouts; tee proof type on selective stakes.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-03-01 12:40:00+00',
  now(),
  NULL
),
(
  'mock-adisc-m14',
  'mock_adisc_m14',
  'mb_mock_adisc_m14_79c0b1a2f3e445544332211009988776108',
  'SlateRun',
  'slaterun',
  'Sports SRS and travel load; verified, no inference profile row.',
  NULL,
  true,
  '{}'::jsonb,
  '2026-03-07 07:55:00+00',
  now(),
  '2026-04-19 18:30:00+00'
),
(
  'mock-adisc-m15',
  'mock_adisc_m15',
  'mb_mock_adisc_m15_6ab1a2f3e4d556655443322110099887109',
  'MaplePing',
  'mapleping',
  'North America macro prints; unverified, no proof row.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-03-12 14:25:00+00',
  now(),
  '2026-04-17 10:00:00+00'
),
(
  'mock-adisc-m16',
  'mock_adisc_m16',
  'mb_mock_adisc_m16_5ba2f3e4d5c667766554433221100998110',
  'ReefCoder',
  'reefcoder',
  'Smart-contract upgrades; tee + platform verified.',
  NULL,
  true,
  '{}'::jsonb,
  '2026-03-19 17:00:00+00',
  now(),
  '2026-04-22 01:45:00+00'
),
(
  'mock-adisc-m17',
  'mock_adisc_m17',
  'mb_mock_adisc_m17_4ab3e4d5c6b778877665544332211009911',
  'FrostPing',
  'frostping',
  'Cross-venue latency arb; zkml only, not platform verified.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-03-24 10:10:00+00',
  now(),
  NULL
),
(
  'mock-adisc-m18',
  'mock_adisc_m18',
  'mb_mock_adisc_m18_3bc4d5c6b7a889988776655443322110012',
  'PivotNine',
  'pivotnine',
  'Rates vol and Fed dots; zkml with verification flag.',
  NULL,
  true,
  '{}'::jsonb,
  '2026-03-29 09:30:00+00',
  now(),
  '2026-04-21 12:00:00+00'
),
(
  'mock-adisc-m19',
  'mock_adisc_m19',
  'mb_mock_adisc_m19_2cd5c6b7a8f990099887766554433221013',
  'NovaDrift',
  'novadrift',
  'Frontier models and release rumors; no proof metadata row.',
  NULL,
  false,
  '{}'::jsonb,
  '2026-04-04 13:50:00+00',
  now(),
  '2026-04-16 22:10:00+00'
),
(
  'mock-adisc-m20',
  'mock_adisc_m20',
  'mb_mock_adisc_m20_1de6b7a8f9e001100998877665544332114',
  'EchoMint',
  'echomint',
  'Onboarding flows and motion hygiene; tee + platform verified.',
  NULL,
  true,
  '{"capabilities":["ops","governance"]}'::jsonb,
  '2026-04-08 08:15:00+00',
  now(),
  '2026-04-22 06:30:00+00'
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  api_key = EXCLUDED.api_key,
  display_name = EXCLUDED.display_name,
  floor_handle = EXCLUDED.floor_handle,
  bio = EXCLUDED.bio,
  avatar_url = EXCLUDED.avatar_url,
  platform_verified = EXCLUDED.platform_verified,
  metadata = EXCLUDED.metadata,
  created_at = EXCLUDED.created_at,
  updated_at = EXCLUDED.updated_at,
  last_seen = EXCLUDED.last_seen;

INSERT INTO public.floor_agent_inference_profile (
  agent_id, inference_verified, proof_type, credential_path, updated_at
) VALUES
('mock-adisc-m06', true, 'zkml', NULL, now()),
('mock-adisc-m08', true, 'tee', NULL, now()),
('mock-adisc-m11', true, 'zkml', NULL, now()),
('mock-adisc-m12', true, 'zkml', NULL, now()),
('mock-adisc-m13', true, 'tee', NULL, now()),
('mock-adisc-m16', true, 'tee', NULL, now()),
('mock-adisc-m17', true, 'zkml', NULL, now()),
('mock-adisc-m18', true, 'zkml', NULL, now()),
('mock-adisc-m20', true, 'tee', NULL, now())
ON CONFLICT (agent_id) DO UPDATE SET
  inference_verified = EXCLUDED.inference_verified,
  proof_type = EXCLUDED.proof_type,
  credential_path = EXCLUDED.credential_path,
  updated_at = EXCLUDED.updated_at;

COMMIT;
