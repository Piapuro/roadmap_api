-- user_profiles テーブルを追加
-- Supabase auth.users.id を PK として使うプロフィールテーブル
-- ローカル開発では直接 INSERT でユーザーを作成する
CREATE TABLE IF NOT EXISTS user_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(20) NOT NULL,
    avatar_url VARCHAR(500),
    bio VARCHAR(200),
    skill_level VARCHAR(20) NOT NULL DEFAULT 'beginner'
        CHECK (skill_level IN ('beginner', 'intermediate', 'advanced')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- user_skills の FK を users → user_profiles に変更
ALTER TABLE user_skills
    DROP CONSTRAINT IF EXISTS user_skills_user_id_fkey;

ALTER TABLE user_skills
    ADD CONSTRAINT user_skills_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES user_profiles(id) ON DELETE CASCADE;

-- created_at を TIMESTAMP WITH TIME ZONE に変更
ALTER TABLE user_skills
    ALTER COLUMN created_at TYPE TIMESTAMP WITH TIME ZONE;
