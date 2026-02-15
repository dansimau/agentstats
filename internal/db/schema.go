package db

const schema = `
CREATE TABLE IF NOT EXISTS projects (
    id          TEXT PRIMARY KEY,
    git_origin  TEXT,
    directory   TEXT NOT NULL UNIQUE,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id),
    agent_type  TEXT NOT NULL DEFAULT 'claude-code',
    started_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS prompts (
    id              TEXT PRIMARY KEY,
    session_id      TEXT NOT NULL REFERENCES sessions(id),
    project_id      TEXT NOT NULL REFERENCES projects(id),
    prompt_text     TEXT,
    submitted_at    DATETIME NOT NULL,
    completed_at    DATETIME,
    git_hash_start  TEXT,
    git_hash_end    TEXT,
    agent_type      TEXT NOT NULL DEFAULT 'claude-code'
);

CREATE INDEX IF NOT EXISTS idx_prompts_session   ON prompts(session_id);
CREATE INDEX IF NOT EXISTS idx_prompts_project   ON prompts(project_id);
CREATE INDEX IF NOT EXISTS idx_prompts_submitted ON prompts(submitted_at);
`
