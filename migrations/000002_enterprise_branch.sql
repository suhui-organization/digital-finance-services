-- =====================================================
-- Migration: 000002_enterprise_branch
-- DESIGN_DOC Chapter 17: Enterprise customer branch support
-- Idempotent: safe to re-run (uses IF NOT EXISTS / IF EXISTS)
-- =====================================================

-- 1. Add customer_type to reviews
ALTER TABLE reviews
ADD COLUMN IF NOT EXISTS customer_type TEXT NOT NULL DEFAULT 'individual';

-- 2. Add enterprise-specific columns to reviews
ALTER TABLE reviews
ADD COLUMN IF NOT EXISTS enterprise_name TEXT,
ADD COLUMN IF NOT EXISTS unified_social_credit_code TEXT,
ADD COLUMN IF NOT EXISTS enterprise_years INT,
ADD COLUMN IF NOT EXISTS main_business TEXT,
ADD COLUMN IF NOT EXISTS monthly_revenue DECIMAL,
ADD COLUMN IF NOT EXISTS controller_cooperate BOOLEAN,
ADD COLUMN IF NOT EXISTS enterprise_highlights TEXT[] NOT NULL DEFAULT '{}';

-- 3. Add debt_owner_type to debt_details for distinguishing debt sources
ALTER TABLE debt_details
ADD COLUMN IF NOT EXISTS debt_owner_type TEXT NOT NULL DEFAULT 'individual';

-- 4. Create indexes (Chapter 17.4)
CREATE INDEX IF NOT EXISTS idx_reviews_customer_type ON reviews(customer_type);
CREATE INDEX IF NOT EXISTS idx_reviews_uscc ON reviews(unified_social_credit_code);
CREATE INDEX IF NOT EXISTS idx_debt_details_review_owner ON debt_details(review_id, debt_owner_type);

-- 5. Backfill historical data
-- All existing records where is_enterprise=false get customer_type='individual'
UPDATE reviews
SET customer_type = 'individual'
WHERE is_enterprise = false AND customer_type != 'individual';

-- Records where is_enterprise=true get customer_type='enterprise' but enterprise fields remain NULL
-- (marked for backfill in admin panel)
UPDATE reviews
SET customer_type = 'enterprise'
WHERE is_enterprise = true AND customer_type != 'enterprise';

-- 6. Backfill debt_details owner type for existing data
UPDATE debt_details
SET debt_owner_type = reviews.customer_type
FROM reviews
WHERE debt_details.review_id = reviews.id
  AND debt_details.debt_owner_type = 'individual'
  AND reviews.customer_type = 'enterprise';

-- =====================================================
-- DOWN script (run only when rolling back):
-- =====================================================
-- DROP INDEX IF EXISTS idx_debt_details_review_owner;
-- DROP INDEX IF EXISTS idx_reviews_uscc;
-- DROP INDEX IF EXISTS idx_reviews_customer_type;
-- ALTER TABLE debt_details DROP COLUMN IF EXISTS debt_owner_type;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS enterprise_highlights;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS controller_cooperate;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS monthly_revenue;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS main_business;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS enterprise_years;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS unified_social_credit_code;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS enterprise_name;
-- ALTER TABLE reviews DROP COLUMN IF EXISTS customer_type;