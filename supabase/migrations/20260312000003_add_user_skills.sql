-- =====================
-- user_skills テーブル
-- ユーザーのスキル・技術登録（新規登録後のオンボーディングで入力）
-- =====================
CREATE TABLE IF NOT EXISTS public.user_skills (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES public.user_profiles(id) ON DELETE CASCADE,
    skill_name       VARCHAR(30) NOT NULL,
    experience_years DECIMAL(3,1),
    is_learning_goal BOOLEAN NOT NULL DEFAULT false,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_skills_user_id ON public.user_skills(user_id);
