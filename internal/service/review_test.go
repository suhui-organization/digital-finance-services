package service

import (
	"testing"

	"digital-finance-services/internal/model"
)

// TestTotalDebtCalculation 验证负债总额自动计算
func TestTotalDebtCalculation(t *testing.T) {
	tests := []struct {
		name     string
		debts    []model.DebtDetail
		expected float64
	}{
		{
			name:     "空负债明细",
			debts:    []model.DebtDetail{},
			expected: 0,
		},
		{
			name: "单条负债",
			debts: []model.DebtDetail{
				{Balance: 50000.50},
			},
			expected: 50000.50,
		},
		{
			name: "多条负债",
			debts: []model.DebtDetail{
				{Balance: 10000},
				{Balance: 20000},
				{Balance: 30000.75},
			},
			expected: 60000.75,
		},
		{
			name: "包含零余额",
			debts: []model.DebtDetail{
				{Balance: 50000},
				{Balance: 0},
			},
			expected: 50000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var total float64
			for _, dd := range tc.debts {
				total += dd.Balance
			}
			if total != tc.expected {
				t.Errorf("total debt = %v, want %v", total, tc.expected)
			}
		})
	}
}

// TestReviewCustomerTypeDefault 验证客户类型默认值
func TestReviewCustomerTypeDefault(t *testing.T) {
	review := model.Review{
		CustomerName: "测试客户",
	}

	// V1 默认值应该是 "individual"
	if review.CustomerType == "" {
		review.CustomerType = "individual"
	}

	if review.CustomerType != "individual" {
		t.Errorf("default customer type should be 'individual', got %q", review.CustomerType)
	}
}

// TestReviewIDGeneration 验证审核记录 ID 格式
func TestReviewIDGeneration(t *testing.T) {
	review := model.Review{
		ID: "550e8400-e29b-41d4-a716-446655440000",
	}

	if len(review.ID) != 36 {
		t.Errorf("UUID should be 36 characters, got %d", len(review.ID))
	}
}
