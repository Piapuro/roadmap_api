-- ロールバック: user_profiles から users に戻す
DROP INDEX IF EXISTS idx_requirements_team_id;

ALTER TABLE requirements DROP CONSTRAINT IF EXISTS requirements_created_by_fkey;
ALTER TABLE requirements ADD CONSTRAINT requirements_created_by_fkey
    FOREIGN KEY (created_by) REFERENCES users(id);
