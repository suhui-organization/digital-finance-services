package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"digital-finance-services/internal/model"
	"digital-finance-services/internal/service"
)

// LotteryHandler handles lottery-related HTTP requests.
// All business logic is delegated to LotteryService (DDD compliant).
type LotteryHandler struct {
	svc *service.LotteryService
}

// NewLotteryHandler creates a LotteryHandler.
func NewLotteryHandler(svc *service.LotteryService) *LotteryHandler {
	return &LotteryHandler{svc: svc}
}

// GetActivity returns the current lottery activity configuration.
func (h *LotteryHandler) GetActivity(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.GetActivity())
}

// UpdateActivity updates the lottery activity configuration.
func (h *LotteryHandler) UpdateActivity(c *gin.Context) {
	var req model.LotteryActivity
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.svc.UpdateActivity(&req)
	c.JSON(http.StatusOK, h.svc.GetActivity())
}

// Draw performs a lottery draw for the current user.
func (h *LotteryHandler) Draw(c *gin.Context) {
	result, err := h.svc.Draw()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// AddPrize adds a new prize to the activity.
func (h *LotteryHandler) AddPrize(c *gin.Context) {
	var prize model.Prize
	if err := c.ShouldBindJSON(&prize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.svc.AddPrize(prize)
	c.JSON(http.StatusCreated, gin.H{"message": "prize added"})
}

// DeletePrize removes a prize from the activity.
func (h *LotteryHandler) DeletePrize(c *gin.Context) {
	prizeID := c.Param("prizeId")

	if err := h.svc.DeletePrize(prizeID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prize deleted"})
}

// GetStats returns lottery statistics.
func (h *LotteryHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.GetStats())
}
