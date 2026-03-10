-- =====================
-- user_profiles テーブル
-- Supabase auth.users.id をそのまま PK として使用
-- =====================
CREATE TABLE IF NOT EXISTS public.user_profiles (
    id UUID PRIMARY KEY,
    name VARCHAR(20) NOT NULL,
    avatar_url VARCHAR(500),
    bio VARCHAR(200),
    skill_level VARCHAR(20) NOT NULL DEFAULT 'beginner'
        CHECK (skill_level IN ('beginner', 'intermediate', 'advanced')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- =====================
-- global_roles テーブル
-- =====================
CREATE TABLE IF NOT EXISTS public.global_roles (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    level SMALLINT NOT NULL
);

INSERT INTO public.global_roles (id, name, level) VALUES
    (1, 'GUEST', 10),
    (2, 'LOGIN_USER', 20),
    (3, 'SYSTEM_ADMIN', 99)
ON CONFLICT DO NOTHING;

-- =====================
-- user_global_roles テーブル
-- user_profiles(id) を参照（auth.users.id と同じ UUID）
-- =====================
CREATE TABLE IF NOT EXISTS public.user_global_roles (
    user_id UUID NOT NULL REFERENCES public.user_profiles(id) ON DELETE CASCADE,
    global_role_id SMALLINT NOT NULL REFERENCES public.global_roles(id),
    granted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, global_role_id)
);

-- =====================
-- Function 1: user_profiles 自動作成
-- auth.users に INSERT されたとき user_profiles にもレコードを作成する
-- =====================
CREATE OR REPLACE FUNCTION public.handle_new_user_profile()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO public.user_profiles (id, name)
    VALUES (
        NEW.id,
        COALESCE(
            NEW.raw_user_meta_data->>'full_name',
            split_part(NEW.email, '@', 1)
        )
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================
-- Function 2: LOGIN_USER ロール自動付与
-- global_roles から LOGIN_USER の id を取得して user_global_roles に INSERT する
-- =====================
CREATE OR REPLACE FUNCTION public.handle_new_user_global_role()
RETURNS TRIGGER AS $$
DECLARE
    login_user_role_id SMALLINT;
BEGIN
    SELECT id INTO login_user_role_id
    FROM public.global_roles
    WHERE name = 'LOGIN_USER';

    INSERT INTO public.user_global_roles (user_id, global_role_id)
    VALUES (NEW.id, login_user_role_id);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- =====================
-- Triggers
-- auth.users の INSERT 後に上記 2 つの Function を実行する
-- =====================
DROP TRIGGER IF EXISTS on_auth_user_created_profile ON auth.users;
CREATE TRIGGER on_auth_user_created_profile
    AFTER INSERT ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION public.handle_new_user_profile();

DROP TRIGGER IF EXISTS on_auth_user_created_role ON auth.users;
CREATE TRIGGER on_auth_user_created_role
    AFTER INSERT ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION public.handle_new_user_global_role();
