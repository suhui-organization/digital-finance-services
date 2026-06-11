package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"digital-finance-services/internal/model"
)

// ReviewRepository handles database operations for reviews.
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository creates a ReviewRepository.
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create inserts a new review with its debt details in a transaction.
func (r *ReviewRepository) Create(review *model.Review) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Ensure enterprise_highlights is never nil (column has NOT NULL constraint)
	entHighlights := review.EnterpriseHighlights
	if entHighlights == nil {
		entHighlights = []string{}
	}

	err = tx.QueryRow(`
		INSERT INTO reviews (
			id, customer_name, gender, age, marital_status, loan_amount,
			is_enterprise, main_bank, total_debt, credit_status,
			credit_query_1m, credit_query_3m, credit_query_6m,
			spouse_info, spouse_cooperate, highlights, can_match,
			visit_time, created_by,
			customer_type, enterprise_name, unified_social_credit_code,
			enterprise_years, main_business, monthly_revenue,
			controller_cooperate, enterprise_highlights
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)
		RETURNING created_at
	`,
		review.ID, review.CustomerName, review.Gender, review.Age, review.MaritalStatus,
		review.LoanAmount, review.IsEnterprise, review.MainBank, review.TotalDebt,
		review.CreditStatus, review.CreditQuery1M, review.CreditQuery3M, review.CreditQuery6M,
		review.SpouseInfo, review.SpouseCooperate, pq.Array(review.Highlights), review.CanMatch,
		review.VisitTime, review.CreatedBy,
		review.CustomerType, review.EnterpriseName, review.UnifiedSocialCreditCode,
		review.EnterpriseYears, review.MainBusiness, review.MonthlyRevenue,
		review.ControllerCooperate, pq.Array(entHighlights),
	).Scan(&review.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert review: %w", err)
	}

	for i := range review.DebtDetails {
		dd := &review.DebtDetails[i]
		err = tx.QueryRow(`
			INSERT INTO debt_details (review_id, institution, total_amount, balance, loan_method, loan_due, repayment_method, debt_owner_type)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id
		`, review.ID, dd.Institution, dd.TotalAmount, dd.Balance, dd.LoanMethod, dd.LoanDue, dd.RepaymentMethod, dd.DebtOwnerType).Scan(&dd.ID)
		if err != nil {
			return fmt.Errorf("insert debt detail: %w", err)
		}
		dd.ReviewID = review.ID
	}

	return tx.Commit()
}

// scanFullReview scans all review columns including V2 enterprise fields.
func scanFullReview(row interface{ Scan(...interface{}) error }, review *model.Review) error {
	var highlights, enterpriseHighlights []string
	err := row.Scan(
		&review.ID, &review.CustomerName, &review.Gender, &review.Age, &review.MaritalStatus,
		&review.LoanAmount, &review.IsEnterprise, &review.MainBank, &review.TotalDebt,
		&review.CreditStatus, &review.CreditQuery1M, &review.CreditQuery3M, &review.CreditQuery6M,
		&review.SpouseInfo, &review.SpouseCooperate, pq.Array(&highlights), &review.CanMatch,
		&review.VisitTime, &review.CreatedBy, &review.AIScore, &review.AIRiskLevel, &review.AISummary, &review.CreatedAt,
		&review.CustomerType, &review.EnterpriseName, &review.UnifiedSocialCreditCode,
		&review.EnterpriseYears, &review.MainBusiness, &review.MonthlyRevenue,
		&review.ControllerCooperate, pq.Array(&enterpriseHighlights),
	)
	if err != nil {
		return err
	}
	review.Highlights = highlights
	review.EnterpriseHighlights = enterpriseHighlights
	return nil
}

const fullReviewSelect = `
		SELECT id, customer_name, gender, age, marital_status, loan_amount,
			is_enterprise, main_bank, total_debt, credit_status,
			credit_query_1m, credit_query_3m, credit_query_6m,
			spouse_info, spouse_cooperate, highlights, can_match,
			visit_time, created_by, ai_score, ai_risk_level, ai_summary, created_at,
			customer_type, enterprise_name, unified_social_credit_code,
			enterprise_years, main_business, monthly_revenue,
			controller_cooperate, enterprise_highlights
		FROM reviews `

// GetByID retrieves a review with its debt details.
func (r *ReviewRepository) GetByID(id string) (*model.Review, error) {
	review := &model.Review{}

	err := scanFullReview(r.db.QueryRow(fullReviewSelect+` WHERE id = $1`, id), review)
	if err != nil {
		return nil, fmt.Errorf("query review: %w", err)
	}

	rows, err := r.db.Query(`
		SELECT id, review_id, institution, total_amount, balance, loan_method, loan_due, repayment_method, debt_owner_type
		FROM debt_details WHERE review_id = $1
		ORDER BY id
	`, id)
	if err != nil {
		return nil, fmt.Errorf("query debt details: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var dd model.DebtDetail
		if err := rows.Scan(&dd.ID, &dd.ReviewID, &dd.Institution, &dd.TotalAmount, &dd.Balance, &dd.LoanMethod, &dd.LoanDue, &dd.RepaymentMethod, &dd.DebtOwnerType); err != nil {
			return nil, fmt.Errorf("scan debt detail: %w", err)
		}
		review.DebtDetails = append(review.DebtDetails, dd)
	}

	return review, nil
}

// UpdateAI updates only the AI fields of an existing review.
func (r *ReviewRepository) UpdateAI(id string, score *float64, riskLevel *string, summary *string) error {
	_, err := r.db.Exec(`
		UPDATE reviews SET ai_score = $2, ai_risk_level = $3, ai_summary = $4
		WHERE id = $1
	`, id, score, riskLevel, summary)
	if err != nil {
		return fmt.Errorf("update ai fields: %w", err)
	}
	return nil
}

// Update updates a review with all fields including V2 enterprise fields and debt details.
func (r *ReviewRepository) Update(review *model.Review) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Ensure enterprise_highlights is never nil
	entHighlights := review.EnterpriseHighlights
	if entHighlights == nil {
		entHighlights = []string{}
	}

	_, err = tx.Exec(`
		UPDATE reviews SET
			customer_name = $2, gender = $3, age = $4, marital_status = $5,
			loan_amount = $6, is_enterprise = $7, main_bank = $8, total_debt = $9,
			credit_status = $10, credit_query_1m = $11, credit_query_3m = $12,
			credit_query_6m = $13, spouse_info = $14, spouse_cooperate = $15,
			highlights = $16, can_match = $17, visit_time = $18,
			customer_type = $19, enterprise_name = $20, unified_social_credit_code = $21,
			enterprise_years = $22, main_business = $23, monthly_revenue = $24,
			controller_cooperate = $25, enterprise_highlights = $26
		WHERE id = $1
	`,
		review.ID, review.CustomerName, review.Gender, review.Age, review.MaritalStatus,
		review.LoanAmount, review.IsEnterprise, review.MainBank, review.TotalDebt,
		review.CreditStatus, review.CreditQuery1M, review.CreditQuery3M, review.CreditQuery6M,
		review.SpouseInfo, review.SpouseCooperate, pq.Array(review.Highlights), review.CanMatch,
		review.VisitTime,
		review.CustomerType, review.EnterpriseName, review.UnifiedSocialCreditCode,
		review.EnterpriseYears, review.MainBusiness, review.MonthlyRevenue,
		review.ControllerCooperate, pq.Array(entHighlights),
	)
	if err != nil {
		return fmt.Errorf("update review: %w", err)
	}

	// Delete old debt details and re-insert
	_, err = tx.Exec(`DELETE FROM debt_details WHERE review_id = $1`, review.ID)
	if err != nil {
		return fmt.Errorf("delete old debt_details: %w", err)
	}

	for i := range review.DebtDetails {
		dd := &review.DebtDetails[i]
		if dd.ID == "" {
			dd.ID = uuid.New().String()
		}
		dd.ReviewID = review.ID
		if dd.DebtOwnerType == "" {
			dd.DebtOwnerType = review.CustomerType
		}
		_, err = tx.Exec(`
			INSERT INTO debt_details (id, review_id, institution, total_amount, balance, loan_method, loan_due, repayment_method, debt_owner_type)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, dd.ID, dd.ReviewID, dd.Institution, dd.TotalAmount, dd.Balance, dd.LoanMethod, dd.LoanDue, dd.RepaymentMethod, dd.DebtOwnerType)
		if err != nil {
			return fmt.Errorf("insert debt detail: %w", err)
		}
	}

	return tx.Commit()
}

// Delete removes a review (debt_details cascade automatically).
func (r *ReviewRepository) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete review: %w", err)
	}
	return nil
}

// List returns a paginated list of reviews (without debt details for performance).
func (r *ReviewRepository) List(page, pageSize int) ([]model.Review, int, error) {
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM reviews`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count reviews: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(fullReviewSelect+`
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]model.Review, 0)
	for rows.Next() {
		var rv model.Review
		if err := scanFullReview(rows, &rv); err != nil {
			return nil, 0, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, rv)
	}

	return reviews, total, nil
}

// ListByUser returns a paginated list of reviews for a specific user (mobile "已填报" page).
func (r *ReviewRepository) ListByUser(userID string, page, pageSize int) ([]model.Review, int, error) {
	// For now, user's reviews are identified by the review's created_by field.
	// In practice, the mobile user ties to the creator. A dedicated user_reviews table
	// may be introduced later for explicit associations.
	var total int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM reviews WHERE created_by = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count user reviews: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.db.Query(fullReviewSelect+`
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query user reviews: %w", err)
	}
	defer rows.Close()

	reviews := make([]model.Review, 0)
	for rows.Next() {
		var rv model.Review
		if err := scanFullReview(rows, &rv); err != nil {
			return nil, 0, fmt.Errorf("scan review: %w", err)
		}
		reviews = append(reviews, rv)
	}

	return reviews, total, nil
}
