package model

import (
	"time"

	"github.com/google/uuid"
)

// Prize represents a lottery prize.
type Prize struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Probability float64 `json:"probability"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`
	IsActive    bool    `json:"is_active"`
}

// LotteryActivity represents a lottery activity configuration.
type LotteryActivity struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Prizes    []Prize   `json:"prizes"`
}

// LotteryDrawResponse is the response for a draw attempt.
type LotteryDrawResponse struct {
	Won     bool   `json:"won"`
	Prize   *Prize `json:"prize,omitempty"`
	Message string `json:"message,omitempty"`
}

// LotteryStats holds read-only summary statistics.
type LotteryStats struct {
	ActivityName string `json:"activity_name"`
	IsActive     bool   `json:"is_active"`
	PrizeCount   int    `json:"prize_count"`
	TotalPrizes  int    `json:"total_prizes"`
}

// DefaultLotteryStartTime returns the default activity start time.
func DefaultLotteryStartTime() time.Time {
	return time.Now()
}

// DefaultLotteryEndTime returns the default activity end time (30 days).
func DefaultLotteryEndTime() time.Time {
	return time.Now().Add(30 * 24 * time.Hour)
}

// DefaultPrizes returns the default prize list.
func DefaultPrizes() []Prize {
	return []Prize{
		{ID: uuid.New().String(), Name: "利率优惠券 0.5%", Probability: 0.05, Stock: 100, IsActive: true},
		{ID: uuid.New().String(), Name: "服务费减免 200元", Probability: 0.10, Stock: 50, IsActive: true},
		{ID: uuid.New().String(), Name: "精美礼品", Probability: 0.15, Stock: 30, IsActive: true},
		{ID: uuid.New().String(), Name: "谢谢参与", Probability: 0.70, Stock: -1, IsActive: true}, // -1 = unlimited
	}
}
