-- Users & Auth
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT,
    email TEXT UNIQUE NOT NULL,
    api_key_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Deployments (replaces in-memory Sandbox.Containers)
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    repo_url TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    container_id TEXT,
    image_tag TEXT,
    host_port INTEGER,
    container_port INTEGER,
    tier_name TEXT NOT NULL DEFAULT 'tier1',
    tier_cpu REAL,
    tier_memory TEXT,
    env_json TEXT DEFAULT '{}',
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Jobs (replaces in-memory JobStore)
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    deployment_id TEXT,
    status TEXT NOT NULL DEFAULT 'queued',
    step TEXT,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (deployment_id) REFERENCES deployments(id) ON DELETE CASCADE
);

-- Port allocations (replaces in-memory ports.Manager tracking)
CREATE TABLE IF NOT EXISTS port_allocations (
    host_port INTEGER PRIMARY KEY,
    container_id TEXT NOT NULL,
    container_port INTEGER NOT NULL,
    allocated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

--TEMPORARY
INSERT INTO users (id, name, email, api_key_hash)
SELECT 
    'test-user-001',
    'Test User', 
    'test@containerix.local',
    '73e9645393777bd3b23819103a85969c556ed6bfa4728c2de81bb207add9e2aa' -- real sha256 hash
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE email = 'test@containerix.local'
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_deployments_user ON deployments(user_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_jobs_deployment ON jobs(deployment_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
