INSERT INTO platform_roles (id, code) VALUES
('11111111-1111-1111-1111-111111111111', 'player'),
('22222222-2222-2222-2222-222222222222', 'platform_admin')
ON CONFLICT (code) DO NOTHING;

INSERT INTO users (id, first_name, last_name, email, phone, nickname, password_hash, is_blocked)
VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Platform', 'Admin', 'admin@example.com', '+77010000001', 'admin_kz', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', FALSE),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Demo', 'Player', 'player@example.com', '+77010000002', 'demo_player', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', FALSE)
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_platform_roles (id, user_id, role_id)
VALUES
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111'),
(gen_random_uuid(), 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '22222222-2222-2222-2222-222222222222'),
(gen_random_uuid(), 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111')
ON CONFLICT DO NOTHING;

INSERT INTO tournaments (id, title, discipline, description, rules, location, max_teams, status, visibility, owner_user_id)
VALUES
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Demo Valorant Cup', 'Valorant', 'Seed tournament for local development', 'Manager approves final results', 'Online', 8, 'draft', 'public', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa')
ON CONFLICT DO NOTHING;

INSERT INTO tournament_user_roles (id, tournament_id, user_id, role, assigned_by)
VALUES
(gen_random_uuid(), 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'owner', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa')
ON CONFLICT DO NOTHING;
