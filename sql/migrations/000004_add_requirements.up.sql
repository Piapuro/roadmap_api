-- requirements.created_by の FK を users → user_profiles に変更
-- Supabase Auth の user_profiles.id を参照するため

ALTER TABLE requirements DROP CONSTRAINT IF EXISTS requirements_created_by_fkey;
ALTER TABLE requirements ADD CONSTRAINT requirements_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES user_profiles(id);

-- requirements テーブルに team_id インデックスを追加
CREATE INDEX IF NOT EXISTS idx_requirements_team_id ON requirements(team_id);
