package service

import (
	"errors"
	"math/rand"
	"time"

	"digital-finance-services/internal/model"
	"digital-finance-services/internal/repository"
)

// LotteryService handles lottery business logic.
type LotteryService struct {
	repo *repository.LotteryRepository
}

// NewLotteryService creates a new LotteryService.
func NewLotteryService(repo *repository.LotteryRepository) *LotteryService {
	return &LotteryService{repo: repo}
}

// GetActivity returns the current lottery activity.
func (s *LotteryService) GetActivity() *model.LotteryActivity {
	return s.repo.GetActivity()
}

// UpdateActivity updates the lottery activity configuration.
func (s *LotteryService) UpdateActivity(update *model.LotteryActivity) {
	s.repo.UpdateActivity(update)
}

// AddPrize adds a new prize to the activity.
func (s *LotteryService) AddPrize(prize model.Prize) {
	s.repo.AddPrize(prize)
}

// DeletePrize removes a prize by ID.
func (s *LotteryService) DeletePrize(prizeID string) error {
	if !s.repo.DeletePrize(prizeID) {
		return errors.New("prize not found")
	}
	return nil
}

// Draw performs a weighted random lottery draw.
func (s *LotteryService) Draw() (*model.LotteryDrawResponse, error) {
	activity := s.repo.GetActivity()

	if !activity.IsActive {
		return nil, errors.New("抽奖活动未开启")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	roll := r.Float64()

	cumulative := 0.0
	for i := range activity.Prizes {
		prize := &activity.Prizes[i]
		cumulative += prize.Probability
		if roll <= cumulative {
			// "谢谢参与" is a loss — no stock check needed
			if prize.Name == "谢谢参与" {
				return &model.LotteryDrawResponse{
					Won:     false,
					Message: "谢谢参与，再接再厉！",
				}, nil
			}

			// Check stock before decrement
			if prize.Stock == 0 {
				return &model.LotteryDrawResponse{
					Won:     false,
					Message: "奖品已领完，谢谢参与",
				}, nil
			}

			// Decrement stock (repository handles thread safety internally)
			if !s.repo.DecrementPrizeStock(i) {
				return &model.LotteryDrawResponse{
					Won:     false,
					Message: "奖品不足，谢谢参与",
				}, nil
			}

			return &model.LotteryDrawResponse{
				Won:    true,
				Prize:  prize,
				Message: "恭喜中奖！",
			}, nil
		}
	}

	return &model.LotteryDrawResponse{
		Won:     false,
		Message: "谢谢参与",
	}, nil
}

// GetStats returns lottery statistics.
func (s *LotteryService) GetStats() *model.LotteryStats {
	activity := s.repo.GetActivity()
	total, realPrizes := s.repo.GetPrizeCount()

	return &model.LotteryStats{
		ActivityName: activity.Name,
		IsActive:     activity.IsActive,
		PrizeCount:   realPrizes,
		TotalPrizes:  total,
	}
}