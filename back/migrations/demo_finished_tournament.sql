BEGIN;

-- Optional fallback owner user for the demo tournament.
-- Password hash corresponds to: password
INSERT INTO users (
    id,
    first_name,
    last_name,
    email,
    phone,
    nickname,
    password_hash,
    is_blocked,
    deleted_at
)
VALUES (
    'f1000000-0000-0000-0000-000000000001',
    'Demo',
    'Owner',
    'finished.owner@example.com',
    '+77010000099',
    'finished_owner',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    FALSE,
    NULL
)
ON CONFLICT (email) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    phone = EXCLUDED.phone,
    nickname = EXCLUDED.nickname,
    password_hash = EXCLUDED.password_hash,
    is_blocked = FALSE,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO platform_roles (id, code)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'player'),
    ('22222222-2222-2222-2222-222222222222', 'platform_admin')
ON CONFLICT (code) DO NOTHING;

INSERT INTO user_platform_roles (id, user_id, role_id)
SELECT gen_random_uuid(), u.id, pr.id
FROM users u
JOIN platform_roles pr ON pr.code = 'player'
WHERE u.email = 'finished.owner@example.com'
ON CONFLICT DO NOTHING;

INSERT INTO tournaments (
    id,
    title,
    discipline,
    description,
    rules,
    location,
    max_teams,
    registration_deadline,
    start_at,
    status,
    visibility,
    owner_user_id,
    deleted_at
)
VALUES (
    'd2000000-0000-0000-0000-000000000001',
    'AITU Valorant Spring Cup 2026 - Completed Demo',
    'Valorant',
    'Demo completed tournament with prefilled teams, bracket and match results for frontend showcase.',
    'Single elimination. Results already finalized by manager.',
    'Online',
    4,
    '2026-04-10T18:00:00Z',
    '2026-04-12T10:00:00Z',
    'finished',
    'public',
    'f1000000-0000-0000-0000-000000000001',
    NULL
)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    discipline = EXCLUDED.discipline,
    description = EXCLUDED.description,
    rules = EXCLUDED.rules,
    location = EXCLUDED.location,
    max_teams = EXCLUDED.max_teams,
    registration_deadline = EXCLUDED.registration_deadline,
    start_at = EXCLUDED.start_at,
    status = EXCLUDED.status,
    visibility = EXCLUDED.visibility,
    owner_user_id = EXCLUDED.owner_user_id,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO tournament_user_roles (id, tournament_id, user_id, role, assigned_by)
VALUES (
    'd3000000-0000-0000-0000-000000000001',
    'd2000000-0000-0000-0000-000000000001',
    'f1000000-0000-0000-0000-000000000001',
    'owner',
    'f1000000-0000-0000-0000-000000000001'
)
ON CONFLICT (tournament_id, user_id, role) DO NOTHING;

INSERT INTO teams (
    id,
    tournament_id,
    name,
    status,
    approved_by_manager,
    created_from_import_row_id,
    deleted_at
)
VALUES
    ('d4000000-0000-0000-0000-000000000001', 'd2000000-0000-0000-0000-000000000001', 'Aitu Wolves', 'approved', TRUE, NULL, NULL),
    ('d4000000-0000-0000-0000-000000000002', 'd2000000-0000-0000-0000-000000000001', 'Nomad Five', 'approved', TRUE, NULL, NULL),
    ('d4000000-0000-0000-0000-000000000003', 'd2000000-0000-0000-0000-000000000001', 'Steppe Vipers', 'approved', TRUE, NULL, NULL),
    ('d4000000-0000-0000-0000-000000000004', 'd2000000-0000-0000-0000-000000000001', 'Cyber Barys', 'approved', TRUE, NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    tournament_id = EXCLUDED.tournament_id,
    name = EXCLUDED.name,
    status = EXCLUDED.status,
    approved_by_manager = EXCLUDED.approved_by_manager,
    created_from_import_row_id = EXCLUDED.created_from_import_row_id,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO team_members (
    id,
    team_id,
    user_id,
    nickname,
    member_role,
    is_captain,
    is_substitute,
    confirmation_status,
    invited_at,
    responded_at,
    deleted_at
)
VALUES
    -- Aitu Wolves
    ('d5000000-0000-0000-0000-000000000001', 'd4000000-0000-0000-0000-000000000001', NULL, 'wolf_cap', 'captain', TRUE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000002', 'd4000000-0000-0000-0000-000000000001', NULL, 'wolf_2', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000003', 'd4000000-0000-0000-0000-000000000001', NULL, 'wolf_3', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000004', 'd4000000-0000-0000-0000-000000000001', NULL, 'wolf_4', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000005', 'd4000000-0000-0000-0000-000000000001', NULL, 'wolf_5', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000006', 'd4000000-0000-0000-0000-000000000001', NULL, 'wolf_sub', 'substitute', FALSE, TRUE, 'confirmed', now(), now(), NULL),

    -- Nomad Five
    ('d5000000-0000-0000-0000-000000000007', 'd4000000-0000-0000-0000-000000000002', NULL, 'nomad_cap', 'captain', TRUE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000008', 'd4000000-0000-0000-0000-000000000002', NULL, 'nomad_2', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000009', 'd4000000-0000-0000-0000-000000000002', NULL, 'nomad_3', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000010', 'd4000000-0000-0000-0000-000000000002', NULL, 'nomad_4', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000011', 'd4000000-0000-0000-0000-000000000002', NULL, 'nomad_5', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000012', 'd4000000-0000-0000-0000-000000000002', NULL, 'nomad_sub', 'substitute', FALSE, TRUE, 'confirmed', now(), now(), NULL),

    -- Steppe Vipers
    ('d5000000-0000-0000-0000-000000000013', 'd4000000-0000-0000-0000-000000000003', NULL, 'viper_cap', 'captain', TRUE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000014', 'd4000000-0000-0000-0000-000000000003', NULL, 'viper_2', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000015', 'd4000000-0000-0000-0000-000000000003', NULL, 'viper_3', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000016', 'd4000000-0000-0000-0000-000000000003', NULL, 'viper_4', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000017', 'd4000000-0000-0000-0000-000000000003', NULL, 'viper_5', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000018', 'd4000000-0000-0000-0000-000000000003', NULL, 'viper_sub', 'substitute', FALSE, TRUE, 'confirmed', now(), now(), NULL),

    -- Cyber Barys
    ('d5000000-0000-0000-0000-000000000019', 'd4000000-0000-0000-0000-000000000004', NULL, 'barys_cap', 'captain', TRUE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000020', 'd4000000-0000-0000-0000-000000000004', NULL, 'barys_2', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000021', 'd4000000-0000-0000-0000-000000000004', NULL, 'barys_3', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000022', 'd4000000-0000-0000-0000-000000000004', NULL, 'barys_4', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000023', 'd4000000-0000-0000-0000-000000000004', NULL, 'barys_5', 'player', FALSE, FALSE, 'confirmed', now(), now(), NULL),
    ('d5000000-0000-0000-0000-000000000024', 'd4000000-0000-0000-0000-000000000004', NULL, 'barys_sub', 'substitute', FALSE, TRUE, 'confirmed', now(), now(), NULL)
ON CONFLICT (id) DO UPDATE SET
    team_id = EXCLUDED.team_id,
    user_id = EXCLUDED.user_id,
    nickname = EXCLUDED.nickname,
    member_role = EXCLUDED.member_role,
    is_captain = EXCLUDED.is_captain,
    is_substitute = EXCLUDED.is_substitute,
    confirmation_status = EXCLUDED.confirmation_status,
    invited_at = EXCLUDED.invited_at,
    responded_at = EXCLUDED.responded_at,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO brackets (
    id,
    tournament_id,
    format,
    seeding_method,
    status,
    generated_by,
    generated_at,
    metadata_json
)
VALUES (
    'd6000000-0000-0000-0000-000000000001',
    'd2000000-0000-0000-0000-000000000001',
    'single_elimination',
    'manual',
    'finished',
    'f1000000-0000-0000-0000-000000000001',
    '2026-04-12T10:05:00Z',
    '{"teams_seeded":["d4000000-0000-0000-0000-000000000001","d4000000-0000-0000-0000-000000000002","d4000000-0000-0000-0000-000000000003","d4000000-0000-0000-0000-000000000004"]}'::jsonb
)
ON CONFLICT (tournament_id) DO UPDATE SET
    format = EXCLUDED.format,
    seeding_method = EXCLUDED.seeding_method,
    status = EXCLUDED.status,
    generated_by = EXCLUDED.generated_by,
    generated_at = EXCLUDED.generated_at,
    metadata_json = EXCLUDED.metadata_json;

INSERT INTO matches (
    id,
    tournament_id,
    bracket_id,
    round_number,
    slot_index,
    team1_id,
    team2_id,
    scheduled_at,
    location_or_server,
    status,
    team1_confirmation_status,
    team2_confirmation_status,
    winner_team_id,
    score_text,
    manager_comment,
    next_match_id,
    source_match1_id,
    source_match2_id,
    is_bye,
    deleted_at
)
VALUES
    -- Semifinal 1
    (
        'd7000000-0000-0000-0000-000000000001',
        'd2000000-0000-0000-0000-000000000001',
        'd6000000-0000-0000-0000-000000000001',
        1,
        1,
        'd4000000-0000-0000-0000-000000000001',
        'd4000000-0000-0000-0000-000000000002',
        '2026-04-12T10:30:00Z',
        'EU Server 1',
        'finished',
        'ready_confirmed',
        'ready_confirmed',
        'd4000000-0000-0000-0000-000000000001',
        '13-9, 13-7',
        'Aitu Wolves advanced to the final.',
        'd7000000-0000-0000-0000-000000000003',
        NULL,
        NULL,
        FALSE,
        NULL
    ),
    -- Semifinal 2
    (
        'd7000000-0000-0000-0000-000000000002',
        'd2000000-0000-0000-0000-000000000001',
        'd6000000-0000-0000-0000-000000000001',
        1,
        2,
        'd4000000-0000-0000-0000-000000000003',
        'd4000000-0000-0000-0000-000000000004',
        '2026-04-12T11:20:00Z',
        'EU Server 2',
        'finished',
        'ready_confirmed',
        'ready_confirmed',
        'd4000000-0000-0000-0000-000000000004',
        '10-13, 13-11, 8-13',
        'Cyber Barys advanced to the final.',
        'd7000000-0000-0000-0000-000000000003',
        NULL,
        NULL,
        FALSE,
        NULL
    ),
    -- Final
    (
        'd7000000-0000-0000-0000-000000000003',
        'd2000000-0000-0000-0000-000000000001',
        'd6000000-0000-0000-0000-000000000001',
        2,
        1,
        'd4000000-0000-0000-0000-000000000001',
        'd4000000-0000-0000-0000-000000000004',
        '2026-04-12T13:00:00Z',
        'Main Stream Server',
        'finished',
        'ready_confirmed',
        'ready_confirmed',
        'd4000000-0000-0000-0000-000000000001',
        '13-6, 11-13, 13-8',
        'Aitu Wolves became tournament champions.',
        NULL,
        'd7000000-0000-0000-0000-000000000001',
        'd7000000-0000-0000-0000-000000000002',
        FALSE,
        NULL
    )
ON CONFLICT (id) DO UPDATE SET
    tournament_id = EXCLUDED.tournament_id,
    bracket_id = EXCLUDED.bracket_id,
    round_number = EXCLUDED.round_number,
    slot_index = EXCLUDED.slot_index,
    team1_id = EXCLUDED.team1_id,
    team2_id = EXCLUDED.team2_id,
    scheduled_at = EXCLUDED.scheduled_at,
    location_or_server = EXCLUDED.location_or_server,
    status = EXCLUDED.status,
    team1_confirmation_status = EXCLUDED.team1_confirmation_status,
    team2_confirmation_status = EXCLUDED.team2_confirmation_status,
    winner_team_id = EXCLUDED.winner_team_id,
    score_text = EXCLUDED.score_text,
    manager_comment = EXCLUDED.manager_comment,
    next_match_id = EXCLUDED.next_match_id,
    source_match1_id = EXCLUDED.source_match1_id,
    source_match2_id = EXCLUDED.source_match2_id,
    is_bye = EXCLUDED.is_bye,
    deleted_at = NULL,
    updated_at = now();

INSERT INTO audit_logs (
    id,
    actor_user_id,
    tournament_id,
    entity_type,
    entity_id,
    action_type,
    description,
    metadata_json
)
VALUES
    (
        'd8000000-0000-0000-0000-000000000001',
        'f1000000-0000-0000-0000-000000000001',
        'd2000000-0000-0000-0000-000000000001',
        'tournament',
        'd2000000-0000-0000-0000-000000000001',
        'demo_finished_tournament_seeded',
        'Demo completed tournament inserted directly via SQL seed.',
        '{"source":"demo_finished_tournament.sql"}'::jsonb
    ),
    (
        'd8000000-0000-0000-0000-000000000002',
        'f1000000-0000-0000-0000-000000000001',
        'd2000000-0000-0000-0000-000000000001',
        'match',
        'd7000000-0000-0000-0000-000000000003',
        'result_approved',
        'Final match result approved for completed demo tournament.',
        '{"winner_team_id":"d4000000-0000-0000-0000-000000000001"}'::jsonb
    )
ON CONFLICT (id) DO UPDATE SET
    actor_user_id = EXCLUDED.actor_user_id,
    tournament_id = EXCLUDED.tournament_id,
    entity_type = EXCLUDED.entity_type,
    entity_id = EXCLUDED.entity_id,
    action_type = EXCLUDED.action_type,
    description = EXCLUDED.description,
    metadata_json = EXCLUDED.metadata_json;

COMMIT;
