package service

import (
	"log/slog"

	"digital-finance-services/internal/client"
	"digital-finance-services/internal/model"
	"digital-finance-services/internal/repository"
)

// ReviewService contains business logic for reviews.
type ReviewService struct {
	repo     *repository.ReviewRepository
	aiClient *client.AIClient
}

// NewReviewService creates a ReviewService.
func NewReviewService(repo *repository.ReviewRepository, aiClient *client.AIClient) *ReviewService {
	return &ReviewService{repo: repo, aiClient: aiClient}
}

// Create validates and creates a review, auto-calculating total_debt.
// It also triggers the AI analysis pipeline in a goroutine.
func (s *ReviewService) Create(review *model.Review) error {
	var total float64
	for _, dd := range review.DebtDetails {
		total += dd.Balance
	}
	review.TotalDebt = total

	if err := s.repo.Create(review); err != nil {
		return err
	}

	// Kick off AI analysis asynchronously
	go func() {
		if err := s.aiClient.AnalyzeReview(review); err != nil {
			slog.Error("AI analysis failed", "review_id", review.ID, "error", err)
			return
		}
		if err := s.repo.UpdateAI(review.ID, review.AIScore, review.AIRiskLevel, review.AISummary); err != nil {
			slog.Error("failed to update AI fields", "review_id", review.ID, "error", err)
		}
	}()

	return nil
}

// GetByID returns a single review with debt details.
func (s *ReviewService) GetByID(id string) (*model.Review, error) {
	return s.repo.GetByID(id)
}

// List returns paginated reviews.
func (s *ReviewService) List(page, pageSize int) (*model.ReviewListResponse, error) {
	data, total, err := s.repo.List(page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.ReviewListResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	}, nil
}

// Update allows updating a review with all fields including debt details.
// It recalculates total_debt and re-triggers AI analysis.
func (s *ReviewService) Update(review *model.Review) error {
	var total float64
	for _, dd := range review.DebtDetails {
		total += dd.Balance
	}
	review.TotalDebt = total

	if err := s.repo.Update(review); err != nil {
		return err
	}

	// Re-trigger AI analysis asynchronously after update
	go func() {
		if err := s.aiClient.AnalyzeReview(review); err != nil {
			slog.Error("AI analysis failed", "review_id", review.ID, "error", err)
			return
		}
		if err := s.repo.UpdateAI(review.ID, review.AIScore, review.AIRiskLevel, review.AISummary); err != nil {
			slog.Error("failed to update AI fields", "review_id", review.ID, "error", err)
		}
	}()

	return nil
}

// ListByUser returns paginated reviews for a specific user.
func (s *ReviewService) ListByUser(userID string, page, pageSize int) (*model.ReviewListResponse, error) {
	data, total, err := s.repo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &model.ReviewListResponse{
		Data:       data,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	}, nil
}

// Delete removes a review and its debt details.
func (s *ReviewService) Delete(id string) error {
	return s.repo.Delete(id)
}
