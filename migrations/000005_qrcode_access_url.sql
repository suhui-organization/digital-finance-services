-- =====================================================
-- Migration: 000005_qrcode_access_url
-- DESIGN_DOC Chapter 35.11: Persist access_url for configurable visit base URL
-- =====================================================

ALTER TABLE qrcode_records
    ADD COLUMN IF NOT EXISTS access_url TEXT;

CREATE INDEX IF NOT EXISTS idx_qrcode_records_access_url ON qrcode_records(access_url);
