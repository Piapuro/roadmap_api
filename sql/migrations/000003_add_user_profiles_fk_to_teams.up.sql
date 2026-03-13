-- teams と user_team_roles の FK を users から user_profiles に変更
-- Supabase Auth の user_profiles.id を参照するため

-- teams.created_by
ALTER TABLE teams DROP CONSTRAINT IF EXISTS teams_created_by_fkey;
ALTER TABLE teams ADD CONSTRAINT teams_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES user_profiles(id);

-- user_team_roles.user_id
ALTER TABLE user_team_roles DROP CONSTRAINT IF EXISTS user_team_roles_user_id_fkey;
ALTER TABLE user_team_roles ADD CONSTRAINT user_team_roles_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES user_profiles(id) ON DELETE CASCADE;
