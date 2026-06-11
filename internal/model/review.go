package model

import "time"

// DebtDetail represents a single debt record in a review.
type DebtDetail struct {
	ID              string  `json:"id" db:"id"`
	ReviewID        string  `json:"review_id" db:"review_id"`
	Institution     string  `json:"institution" db:"institution"`
	TotalAmount     float64 `json:"total_amount" db:"total_amount"`
	Balance         float64 `json:"balance" db:"balance"`
	LoanMethod      string  `json:"loan_method" db:"loan_method"`
	LoanDue         string  `json:"loan_due" db:"loan_due"`
	RepaymentMethod string  `json:"repayment_method" db:"repayment_method"`
	// DESIGN_DOC 17.2: enterprise branch debt ownership
	DebtOwnerType string `json:"debt_owner_type,omitempty" db:"debt_owner_type"`
}

// Review represents a complete financing qualification review.
type Review struct {
	ID              string    `json:"id" db:"id"`
	CustomerName    string    `json:"customer_name" db:"customer_name"`
	Gender          string    `json:"gender" db:"gender"`
	Age             int       `json:"age" db:"age"`
	MaritalStatus   string    `json:"marital_status" db:"marital_status"`
	LoanAmount      float64   `json:"loan_amount" db:"loan_amount"`
	IsEnterprise    bool      `json:"is_enterprise" db:"is_enterprise"`
	MainBank        string    `json:"main_bank" db:"main_bank"`
	TotalDebt       float64   `json:"total_debt" db:"total_debt"`
	CreditStatus    string    `json:"credit_status" db:"credit_status"`
	CreditQuery1M   int       `json:"credit_query_1m" db:"credit_query_1m"`
	CreditQuery3M   int       `json:"credit_query_3m" db:"credit_query_3m"`
	CreditQuery6M   int       `json:"credit_query_6m" db:"credit_query_6m"`
	SpouseInfo      string    `json:"spouse_info" db:"spouse_info"`
	SpouseCooperate bool      `json:"spouse_cooperate" db:"spouse_cooperate"`
	Highlights      []string  `json:"highlights" db:"highlights"`
	CanMatch        bool      `json:"can_match" db:"can_match"`
	VisitTime       time.Time `json:"visit_time" db:"visit_time"`
	CreatedBy       string    `json:"created_by" db:"created_by"`
	AIScore         *float64  `json:"ai_score" db:"ai_score"`
	AIRiskLevel     *string   `json:"ai_risk_level" db:"ai_risk_level"`
	AISummary       *string   `json:"ai_summary" db:"ai_summary"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	DebtDetails     []DebtDetail `json:"debt_details,omitempty"`

	// DESIGN_DOC 17.2: customer type
	CustomerType string `json:"customer_type" db:"customer_type"`

	// DESIGN_DOC 17.2: enterprise fields
	EnterpriseName            *string  `json:"enterprise_name,omitempty" db:"enterprise_name"`
	UnifiedSocialCreditCode   *string  `json:"unified_social_credit_code,omitempty" db:"unified_social_credit_code"`
	EnterpriseYears           *int     `json:"enterprise_years,omitempty" db:"enterprise_years"`
	MainBusiness              *string  `json:"main_business,omitempty" db:"main_business"`
	MonthlyRevenue            *float64 `json:"monthly_revenue,omitempty" db:"monthly_revenue"`
	ControllerCooperate       *bool    `json:"controller_cooperate,omitempty" db:"controller_cooperate"`
	EnterpriseHighlights      []string `json:"enterprise_highlights,omitempty" db:"enterprise_highlights"`
}

// CreateReviewRequest is the V1 flat payload (kept for backward compat).
type CreateReviewRequest struct {
	CustomerName    string       `json:"customer_name" binding:"required"`
	Gender          string       `json:"gender" binding:"required"`
	Age             int          `json:"age" binding:"required,min=18,max=120"`
	MaritalStatus   string       `json:"marital_status" binding:"required"`
	LoanAmount      float64      `json:"loan_amount" binding:"required,min=0"`
	IsEnterprise    bool         `json:"is_enterprise"`
	MainBank        string       `json:"main_bank" binding:"required"`
	CreditStatus    string       `json:"credit_status" binding:"required"`
	CreditQuery1M   int          `json:"credit_query_1m" binding:"min=0"`
	CreditQuery3M   int          `json:"credit_query_3m" binding:"min=0"`
	CreditQuery6M   int          `json:"credit_query_6m" binding:"min=0"`
	SpouseInfo      string       `json:"spouse_info" binding:"required"`
	SpouseCooperate bool         `json:"spouse_cooperate"`
	Highlights      []string     `json:"highlights" binding:"required,min=1"`
	CanMatch        bool         `json:"can_match"`
	VisitTime       string       `json:"visit_time" binding:"required"`
	DebtDetails     []DebtDetail `json:"debt_details" binding:"required,min=1,max=5"`
}

// --- V2 request types (DESIGN_DOC 9.2 / 16.2) ---

// CommonRequest holds fields shared by all customers.
type CommonRequest struct {
	CustomerName  string `json:"customer_name" binding:"required"`
	Gender        string `json:"gender" binding:"required"`
	Age           int    `json:"age" binding:"required,min=18,max=120"`
	MaritalStatus string `json:"marital_status" binding:"required"`
	LoanAmount    float64 `json:"loan_amount" binding:"required,min=0"`
	IsEnterprise  bool    `json:"is_enterprise"`
	CanMatch      bool    `json:"can_match"`
	VisitTime     string  `json:"visit_time" binding:"required"`
}

// IndividualProfileRequest holds individual-only fields.
type IndividualProfileRequest struct {
	MainBank        string       `json:"main_bank" binding:"required"`
	CreditStatus    string       `json:"credit_status" binding:"required"`
	CreditQuery1M   int          `json:"credit_query_1m" binding:"min=0"`
	CreditQuery3M   int          `json:"credit_query_3m" binding:"min=0"`
	CreditQuery6M   int          `json:"credit_query_6m" binding:"min=0"`
	SpouseInfo      string       `json:"spouse_info" binding:"required"`
	SpouseCooperate bool         `json:"spouse_cooperate"`
	Highlights      []string     `json:"highlights" binding:"required,min=1"`
	DebtDetails     []DebtDetail `json:"debt_details" binding:"required,min=1,max=5"`
}

// EnterpriseProfileRequest holds enterprise-only fields.
type EnterpriseProfileRequest struct {
	EnterpriseName          string       `json:"enterprise_name" binding:"required"`
	UnifiedSocialCreditCode string       `json:"unified_social_credit_code" binding:"required"`
	EnterpriseYears         int          `json:"enterprise_years" binding:"min=0"`
	MainBusiness            string       `json:"main_business" binding:"required"`
	MonthlyRevenue          float64      `json:"monthly_revenue" binding:"min=0"`
	CreditStatus            string       `json:"credit_status" binding:"required"`
	CreditQuery1M           int          `json:"credit_query_1m" binding:"min=0"`
	CreditQuery3M           int          `json:"credit_query_3m" binding:"min=0"`
	CreditQuery6M           int          `json:"credit_query_6m" binding:"min=0"`
	ControllerCooperate     bool         `json:"controller_cooperate"`
	Highlights              []string     `json:"highlights" binding:"required,min=1"`
	DebtDetails             []DebtDetail `json:"debt_details" binding:"required,min=1,max=5"`
}

// CreateReviewV2Request is the V2 grouped payload.
type CreateReviewV2Request struct {
	CustomerType      string                    `json:"customer_type" binding:"required,oneof=individual enterprise"`
	Common            CommonRequest             `json:"common" binding:"required"`
	IndividualProfile *IndividualProfileRequest `json:"individual_profile"`
	EnterpriseProfile *EnterpriseProfileRequest `json:"enterprise_profile"`
}

// ReviewListResponse wraps a paginated review list.
type ReviewListResponse struct {
	Data       []Review `json:"data"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalCount int      `json:"total_count"`
}