-- 融资资质审查系统 - 初始数据库迁移

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 审查报告主表
CREATE TABLE IF NOT EXISTS reviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_name   TEXT NOT NULL,
    gender          TEXT NOT NULL,
    age             INT NOT NULL CHECK (age >= 18 AND age <= 120),
    marital_status  TEXT NOT NULL,
    loan_amount     DECIMAL NOT NULL DEFAULT 0 CHECK (loan_amount >= 0),
    is_enterprise   BOOLEAN NOT NULL DEFAULT FALSE,
    main_bank       TEXT NOT NULL,
    total_debt      DECIMAL NOT NULL DEFAULT 0 CHECK (total_debt >= 0),
    credit_status   TEXT NOT NULL,
    credit_query_1m INT NOT NULL DEFAULT 0 CHECK (credit_query_1m >= 0),
    credit_query_3m INT NOT NULL DEFAULT 0 CHECK (credit_query_3m >= 0),
    credit_query_6m INT NOT NULL DEFAULT 0 CHECK (credit_query_6m >= 0),
    spouse_info     TEXT NOT NULL,
    spouse_cooperate BOOLEAN NOT NULL DEFAULT FALSE,
    highlights      TEXT[] NOT NULL DEFAULT '{}',
    can_match       BOOLEAN NOT NULL DEFAULT FALSE,
    visit_time      TIMESTAMPTZ NOT NULL,
    created_by      UUID NOT NULL,
    ai_score        DECIMAL,
    ai_risk_level   TEXT,
    ai_summary      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 负债明细表
CREATE TABLE IF NOT EXISTS debt_details (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    review_id        UUID NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    institution      TEXT NOT NULL,
    total_amount     DECIMAL NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    balance          DECIMAL NOT NULL DEFAULT 0 CHECK (balance >= 0),
    loan_method      TEXT NOT NULL,
    loan_due         TEXT NOT NULL,
    repayment_method TEXT NOT NULL
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reviews_credit_status ON reviews(credit_status);
CREATE INDEX IF NOT EXISTS idx_debt_details_review_id ON debt_details(review_id);