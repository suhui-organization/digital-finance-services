package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"digital-finance-services/internal/model"
)

// AIClient wraps calls to the Python AI microservice.
type AIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAIClient creates a new AIClient.
func NewAIClient(baseURL string) *AIClient {
	return &AIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ScoreRequest mirrors the AI service's scoring input.
type ScoreRequest struct {
	CustomerName    string             `json:"customer_name"`
	Gender          string             `json:"gender"`
	Age             int                `json:"age"`
	MaritalStatus   string             `json:"marital_status"`
	LoanAmount      float64            `json:"loan_amount"`
	IsEnterprise    bool               `json:"is_enterprise"`
	MainBank        string             `json:"main_bank"`
	TotalDebt       float64            `json:"total_debt"`
	CreditStatus    string             `json:"credit_status"`
	CreditQuery1M   int                `json:"credit_query_1m"`
	CreditQuery3M   int                `json:"credit_query_3m"`
	CreditQuery6M   int                `json:"credit_query_6m"`
	SpouseInfo      string             `json:"spouse_info"`
	SpouseCooperate bool               `json:"spouse_cooperate"`
	Highlights      []string           `json:"highlights"`
	CanMatch        bool               `json:"can_match"`
	DebtDetails     []model.DebtDetail `json:"debt_details"`
}

// ScoreResponse mirrors the AI service's scoring response.
type ScoreResponse struct {
	Score  float64 `json:"score"`
	Level  string  `json:"level"`
	Detail string  `json:"detail"`
}

// RiskResponse mirrors the AI service's risk analysis response.
type RiskResponse struct {
	RiskLevel string   `json:"risk_level"`
	Score     float64  `json:"score"`
	Warnings  []string `json:"warnings"`
}

// SummaryResponse mirrors the AI service's summary response.
type SummaryResponse struct {
	Summary string `json:"summary"`
}

// post sends a POST request to the AI service.
func (c *AIClient) post(path string, body interface{}, result interface{}) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+path, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("ai service call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ai service returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// Score calls the AI scoring endpoint.
func (c *AIClient) Score(review *model.Review) (*ScoreResponse, error) {
	req := ScoreRequest{
		CustomerName:    review.CustomerName,
		Gender:          review.Gender,
		Age:             review.Age,
		MaritalStatus:   review.MaritalStatus,
		LoanAmount:      review.LoanAmount,
		IsEnterprise:    review.IsEnterprise,
		MainBank:        review.MainBank,
		TotalDebt:       review.TotalDebt,
		CreditStatus:    review.CreditStatus,
		CreditQuery1M:   review.CreditQuery1M,
		CreditQuery3M:   review.CreditQuery3M,
		CreditQuery6M:   review.CreditQuery6M,
		SpouseInfo:      review.SpouseInfo,
		SpouseCooperate: review.SpouseCooperate,
		Highlights:      review.Highlights,
		CanMatch:        review.CanMatch,
		DebtDetails:     review.DebtDetails,
	}

	var result ScoreResponse
	if err := c.post("/api/v1/scoring/score", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AnalyzeRisk calls the AI risk analysis endpoint.
func (c *AIClient) AnalyzeRisk(review *model.Review) (*RiskResponse, error) {
	req := ScoreRequest{
		CustomerName:    review.CustomerName,
		Gender:          review.Gender,
		Age:             review.Age,
		MaritalStatus:   review.MaritalStatus,
		LoanAmount:      review.LoanAmount,
		IsEnterprise:    review.IsEnterprise,
		MainBank:        review.MainBank,
		TotalDebt:       review.TotalDebt,
		CreditStatus:    review.CreditStatus,
		CreditQuery1M:   review.CreditQuery1M,
		CreditQuery3M:   review.CreditQuery3M,
		CreditQuery6M:   review.CreditQuery6M,
		SpouseInfo:      review.SpouseInfo,
		SpouseCooperate: review.SpouseCooperate,
		Highlights:      review.Highlights,
		CanMatch:        review.CanMatch,
		DebtDetails:     review.DebtDetails,
	}

	var result RiskResponse
	if err := c.post("/api/v1/risk/analyze", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GenerateSummary calls the AI summary endpoint.
func (c *AIClient) GenerateSummary(review *model.Review) (*SummaryResponse, error) {
	req := ScoreRequest{
		CustomerName:    review.CustomerName,
		Gender:          review.Gender,
		Age:             review.Age,
		MaritalStatus:   review.MaritalStatus,
		LoanAmount:      review.LoanAmount,
		IsEnterprise:    review.IsEnterprise,
		MainBank:        review.MainBank,
		TotalDebt:       review.TotalDebt,
		CreditStatus:    review.CreditStatus,
		CreditQuery1M:   review.CreditQuery1M,
		CreditQuery3M:   review.CreditQuery3M,
		CreditQuery6M:   review.CreditQuery6M,
		SpouseInfo:      review.SpouseInfo,
		SpouseCooperate: review.SpouseCooperate,
		Highlights:      review.Highlights,
		CanMatch:        review.CanMatch,
		DebtDetails:     review.DebtDetails,
	}

	var result SummaryResponse
	if err := c.post("/api/v1/summary/generate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AnalyzeReview runs the full AI pipeline: scoring + risk + summary.
// It updates the review's AI fields in place.
func (c *AIClient) AnalyzeReview(review *model.Review) error {
	// Scoring
	scoreRes, err := c.Score(review)
	if err != nil {
		return fmt.Errorf("scoring failed: %w", err)
	}
	review.AIScore = &scoreRes.Score
	level := scoreRes.Level
	review.AIRiskLevel = &level

	// Risk analysis
	riskRes, err := c.AnalyzeRisk(review)
	if err != nil {
		// Risk non-blocking: log and continue
		// Use risk level from scoring as fallback
		if review.AIRiskLevel == nil {
			rl := riskRes.RiskLevel
			review.AIRiskLevel = &rl
		}
	} else {
		rl := riskRes.RiskLevel
		review.AIRiskLevel = &rl
	}

	// Summary
	summaryRes, err := c.GenerateSummary(review)
	if err != nil {
		// Summary non-blocking
		defaultSummary := fmt.Sprintf("客户 %s，申请额度 %.0f 元，征信状态 %s",
			review.CustomerName, review.LoanAmount, review.CreditStatus)
		review.AISummary = &defaultSummary
	} else {
		review.AISummary = &summaryRes.Summary
	}

	return nil
}
