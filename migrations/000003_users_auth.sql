-- =====================================================
-- Migration: 000003_users_auth
-- DESIGN_DOC Chapter 26/30: Account system, RBAC, audit logs
-- Idempotent: safe to re-run (uses IF NOT EXISTS / IF EXISTS)
-- =====================================================

-- 1. Users table (DESIGN_DOC 30.1)
CREATE TABLE IF NOT EXISTS users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username        TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    display_name    TEXT NOT NULL DEFAULT '',
    phone           TEXT DEFAULT '',
    email           TEXT DEFAULT '',
    role            TEXT NOT NULL DEFAULT 'mobile_user'
                    CHECK (role IN ('super_admin', 'admin', 'mobile_user')),
    status          TEXT NOT NULL DEFAULT 'active'
                    CHECK (status IN ('active', 'disabled')),
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. User sessions table (DESIGN_DOC 30.1)
CREATE TABLE IF NOT EXISTS user_sessions (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash  TEXT NOT NULL,
    client_type         TEXT NOT NULL DEFAULT 'admin'
                        CHECK (client_type IN ('admin', 'mobile')),
    expires_at          TIMESTAMPTZ NOT NULL,
    revoked_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 3. Audit logs table (DESIGN_DOC 30.1)
CREATE TABLE IF NOT EXISTS audit_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action          TEXT NOT NULL,
    target_type     TEXT NOT NULL,
    target_id       TEXT NOT NULL,
    payload_snapshot JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_hash ON user_sessions(refresh_token_hash);
CREATE INDEX IF NOT EXISTS idx_audit_logs_operator ON audit_logs(operator_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_target ON audit_logs(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);

-- 5. Seed default super_admin account (password: admin123 — bcrypt hash)
-- IMPORTANT: Change password after production deployment!
INSERT INTO users (username, password_hash, display_name, role, status)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    '超级管理员',
    'super_admin',
    'active'
) ON CONFLICT (username) DO NOTHING;

-- =====================================================
-- DOWN script (run only when rolling back):
-- =====================================================
-- DROP TABLE IF EXISTS audit_logs;
-- DROP TABLE IF EXISTS user_sessions;
-- DROP TABLE IF EXISTS users;