-- =====================
-- user_global_roles の FK を users から user_profiles に変更
-- Supabase Auth では auth.users に登録されるため、user_profiles を参照する必要がある
--
-- 前提: 20260308000001_add_global_role_trigger.sql が適用済みであること（user_profiles が存在すること）
-- =====================

-- 既存の FK 制約を削除
ALTER TABLE public.user_global_roles
DROP CONSTRAINT IF EXISTS user_global_roles_user_id_fkey;

-- user_profiles を参照する新しい FK を追加
ALTER TABLE public.user_global_roles
ADD CONSTRAINT user_global_roles_user_id_fkey
FOREIGN KEY (user_id) REFERENCES public.user_profiles(id) ON DELETE CASCADE;
