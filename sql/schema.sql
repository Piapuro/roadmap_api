-- =====================
-- 認証・ユーザー管理
-- =====================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(20) NOT NULL,
    password_hash VARCHAR(255),
    avatar_url VARCHAR(500),
    bio VARCHAR(200),
    skill_level VARCHAR(20) NOT NULL DEFAULT 'beginner'
        CHECK (skill_level IN ('beginner', 'intermediate', 'advanced')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_name VARCHAR(30) NOT NULL,
    experience_years DECIMAL(3,1),
    is_learning_goal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS global_roles (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    level SMALLINT NOT NULL
);

INSERT INTO global_roles (id, name, level) VALUES
    (1, 'GUEST', 10),
    (2, 'LOGIN_USER', 20),
    (3, 'SYSTEM_ADMIN', 99)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS user_global_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    global_role_id SMALLINT NOT NULL REFERENCES global_roles(id),
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, global_role_id)
);

-- =====================
-- チーム管理
-- =====================
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(30) NOT NULL,
    goal VARCHAR(200) NOT NULL,
    level VARCHAR(20) NOT NULL DEFAULT 'beginner'
        CHECK (level IN ('beginner', 'mixed', 'advanced')),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL CHECK (end_date >= start_date),
    is_archived BOOLEAN NOT NULL DEFAULT false,
    invite_token VARCHAR(100) UNIQUE,
    invite_token_expires_at TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS team_roles (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    level SMALLINT NOT NULL
);

INSERT INTO team_roles (id, name, level) VALUES
    (1, 'TEAM_MEMBER', 10),
    (2, 'TEAM_OWNER', 20)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS user_team_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    team_role_id SMALLINT NOT NULL REFERENCES team_roles(id),
    functional_role VARCHAR(20)
        CHECK (functional_role IN ('pm', 'frontend', 'backend', 'uiux', 'infra')),
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, team_id)
);

-- =====================
-- 要件定義
-- =====================
CREATE TABLE IF NOT EXISTS requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    product_type VARCHAR(10) NOT NULL
        CHECK (product_type IN ('web', 'app', 'game', 'ai')),
    difficulty_level SMALLINT NOT NULL CHECK (difficulty_level BETWEEN 1 AND 3),
    free_text VARCHAR(1000),
    supplement_url VARCHAR(500),
    status VARCHAR(10) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'locked')),
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS requirement_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requirement_id UUID NOT NULL REFERENCES requirements(id) ON DELETE CASCADE,
    feature_name VARCHAR(100) NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT true
);

-- =====================
-- AI生成・ロードマップ
-- =====================
CREATE TABLE IF NOT EXISTS ai_generation_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id),
    user_id UUID NOT NULL REFERENCES users(id),
    requirement_id UUID NOT NULL REFERENCES requirements(id),
    status VARCHAR(10) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'success', 'failed')),
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roadmaps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    requirement_id UUID NOT NULL REFERENCES requirements(id),
    status VARCHAR(10) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'confirmed', 'archived')),
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tech_stacks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roadmap_id UUID NOT NULL REFERENCES roadmaps(id) ON DELETE CASCADE,
    category VARCHAR(20) NOT NULL
        CHECK (category IN ('frontend', 'backend', 'database', 'infra', 'api')),
    name VARCHAR(100) NOT NULL,
    reason TEXT,
    is_adopted BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS epics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    roadmap_id UUID NOT NULL REFERENCES roadmaps(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    phase_number SMALLINT NOT NULL DEFAULT 0,
    order_index SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS stories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epic_id UUID NOT NULL REFERENCES epics(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    order_index SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID REFERENCES stories(id) ON DELETE SET NULL,
    roadmap_id UUID NOT NULL REFERENCES roadmaps(id) ON DELETE CASCADE,
    title VARCHAR(50) NOT NULL,
    description TEXT,
    estimated_hours DECIMAL(5,1),
    status VARCHAR(10) NOT NULL DEFAULT 'todo'
        CHECK (status IN ('todo', 'doing', 'review', 'done')),
    assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date DATE,
    phase_number SMALLINT NOT NULL DEFAULT 0,
    order_index SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =====================
-- 学習リソース
-- =====================
CREATE TABLE IF NOT EXISTS learning_resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    skill_name VARCHAR(100),
    title VARCHAR(200) NOT NULL,
    url VARCHAR(500) NOT NULL,
    source_type VARCHAR(20) NOT NULL
        CHECK (source_type IN ('zenn', 'youtube', 'udemy', 'official_doc', 'custom')),
    added_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =====================
-- インデックス
-- =====================
CREATE INDEX IF NOT EXISTS idx_tasks_roadmap_status
    ON tasks(roadmap_id, status);

CREATE INDEX IF NOT EXISTS idx_tasks_assigned_user
    ON tasks(assigned_user_id, status);

CREATE INDEX IF NOT EXISTS idx_user_team_roles_lookup
    ON user_team_roles(user_id, team_id);
