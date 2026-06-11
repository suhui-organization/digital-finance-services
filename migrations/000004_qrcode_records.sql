-- =====================================================
-- Migration: 000004_qrcode_records
-- DESIGN_DOC Chapter 35.8: Admin QR code persisted records
-- =====================================================

CREATE TABLE IF NOT EXISTS qrcode_records (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    target_url  TEXT NOT NULL,
    channel     TEXT,
    campaign    TEXT,
    note        TEXT,
    final_url   TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'disabled')),
    created_by  UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qrcode_records_status ON qrcode_records(status);
CREATE INDEX IF NOT EXISTS idx_qrcode_records_created_by ON qrcode_records(created_by);
CREATE INDEX IF NOT EXISTS idx_qrcode_records_created_at ON qrcode_records(created_at DESC);

-- DOWN script (run only when rolling back):
-- DROP TABLE IF EXISTS qrcode_records;
