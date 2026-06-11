package repository

import (
	"sync"

	"github.com/google/uuid"

	"digital-finance-services/internal/model"
)

// LotteryRepository provides thread-safe in-memory storage for lottery data.
// MVP: In production this should be backed by a database.
type LotteryRepository struct {
	mu       sync.RWMutex
	activity *model.LotteryActivity
}

// NewLotteryRepository creates a LotteryRepository with default prizes.
func NewLotteryRepository() *LotteryRepository {
	r := &LotteryRepository{}
	r.activity = &model.LotteryActivity{
		ID:        uuid.New().String(),
		Name:      "融资审查抽奖活动",
		IsActive:  true,
		StartTime: model.DefaultLotteryStartTime(),
		EndTime:   model.DefaultLotteryEndTime(),
		Prizes:    model.DefaultPrizes(),
	}
	return r
}

// GetActivity returns the current activity config (thread-safe read).
func (r *LotteryRepository) GetActivity() *model.LotteryActivity {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external mutation
	cp := *r.activity
	cp.Prizes = make([]model.Prize, len(r.activity.Prizes))
	copy(cp.Prizes, r.activity.Prizes)
	return &cp
}

// UpdateActivity updates mutable activity fields (thread-safe write).
func (r *LotteryRepository) UpdateActivity(update *model.LotteryActivity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.activity.IsActive = update.IsActive
	if update.Name != "" {
		r.activity.Name = update.Name
	}
	if len(update.Prizes) > 0 {
		r.activity.Prizes = update.Prizes
	}
}

// AddPrize appends a prize (thread-safe write).
func (r *LotteryRepository) AddPrize(prize model.Prize) {
	r.mu.Lock()
	defer r.mu.Unlock()

	prize.ID = uuid.New().String()
	r.activity.Prizes = append(r.activity.Prizes, prize)
}

// DeletePrize removes a prize by ID (thread-safe write).
// Returns true if deleted, false if not found.
func (r *LotteryRepository) DeletePrize(prizeID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, p := range r.activity.Prizes {
		if p.ID == prizeID {
			r.activity.Prizes = append(r.activity.Prizes[:i], r.activity.Prizes[i+1:]...)
			return true
		}
	}
	return false
}

// DecrementPrizeStock safely decrements a prize's stock (thread-safe write).
// Returns true if stock was decremented, false if out of stock or not found.
func (r *LotteryRepository) DecrementPrizeStock(index int) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if index < 0 || index >= len(r.activity.Prizes) {
		return false
	}

	prize := &r.activity.Prizes[index]
	if prize.Stock == 0 {
		return false
	}
	if prize.Stock > 0 {
		prize.Stock--
	}
	return true
}

// GetPrizeCount returns total and real prize counts (thread-safe read).
func (r *LotteryRepository) GetPrizeCount() (total int, realPrizeCount int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.activity.Prizes {
		total++
		if p.Name != "谢谢参与" {
			realPrizeCount++
		}
	}
	return
}