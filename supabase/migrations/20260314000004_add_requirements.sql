-- =====================
-- 要件定義テーブル
-- チームごとの要件定義（プロダクト種別・機能・難易度）を管理する
-- =====================

CREATE TABLE IF NOT EXISTS public.requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES public.teams(id) ON DELETE CASCADE,
    product_type VARCHAR(10) NOT NULL
        CHECK (product_type IN ('web', 'app', 'game', 'ai')),
    difficulty_level SMALLINT NOT NULL CHECK (difficulty_level BETWEEN 1 AND 3),
    free_text VARCHAR(1000),
    supplement_url VARCHAR(500),
    status VARCHAR(10) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'locked')),
    created_by UUID NOT NULL REFERENCES public.user_profiles(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 機能チェックリストテーブル
CREATE TABLE IF NOT EXISTS public.requirement_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requirement_id UUID NOT NULL REFERENCES public.requirements(id) ON DELETE CASCADE,
    feature_name VARCHAR(100) NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT true
);

-- インデックス
CREATE INDEX IF NOT EXISTS idx_requirements_team_id ON public.requirements(team_id);
