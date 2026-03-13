-- ロールバック: user_profiles から users に戻す（users が存在する場合）
ALTER TABLE teams DROP CONSTRAINT IF EXISTS teams_created_by_fkey;
ALTER TABLE teams ADD CONSTRAINT teams_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id);

ALTER TABLE user_team_roles DROP CONSTRAINT IF EXISTS user_team_roles_user_id_fkey;
ALTER TABLE user_team_roles ADD CONSTRAINT user_team_roles_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
