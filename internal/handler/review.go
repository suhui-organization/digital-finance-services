package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"digital-finance-services/internal/model"
	"digital-finance-services/internal/service"
)

// ReviewHandler handles HTTP requests for reviews.
type ReviewHandler struct {
	svc *service.ReviewService
}

// NewReviewHandler creates a ReviewHandler.
func NewReviewHandler(svc *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{svc: svc}
}

// Create handles POST /api/v1/reviews with V1/V2 dual-protocol support.
// DESIGN_DOC 9.2 / 16.2: V2 grouped payload preferred; V1 flat payload as fallback.
func (h *ReviewHandler) Create(c *gin.Context) {
	// Read raw body for dual-protocol detection
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var review model.Review

	// Try V2 grouped payload first
	var v2Req model.CreateReviewV2Request
	if err := json.Unmarshal(bodyBytes, &v2Req); err == nil && v2Req.CustomerType != "" {
		// V2 path
		visitTime, err := time.Parse("2006-01-02T15:04", v2Req.Common.VisitTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_time format, expected YYYY-MM-DDTHH:MM"})
			return
		}

		review = model.Review{
			ID:            uuid.New().String(),
			CustomerName:  v2Req.Common.CustomerName,
			Gender:        v2Req.Common.Gender,
			Age:           v2Req.Common.Age,
			MaritalStatus: v2Req.Common.MaritalStatus,
			LoanAmount:    v2Req.Common.LoanAmount,
			IsEnterprise:  v2Req.Common.IsEnterprise,
			CanMatch:      v2Req.Common.CanMatch,
			VisitTime:     visitTime,
			CreatedBy:     uuid.New().String(),
			CustomerType:  v2Req.CustomerType,
		}

		if v2Req.CustomerType == "individual" && v2Req.IndividualProfile != nil {
			p := v2Req.IndividualProfile
			review.MainBank = p.MainBank
			review.CreditStatus = p.CreditStatus
			review.CreditQuery1M = p.CreditQuery1M
			review.CreditQuery3M = p.CreditQuery3M
			review.CreditQuery6M = p.CreditQuery6M
			review.SpouseInfo = p.SpouseInfo
			review.SpouseCooperate = p.SpouseCooperate
			review.Highlights = p.Highlights
			debt := make([]model.DebtDetail, len(p.DebtDetails))
			for i, d := range p.DebtDetails {
				debt[i] = d
				debt[i].ID = uuid.New().String()
			}
			review.DebtDetails = debt
		}

		if v2Req.CustomerType == "enterprise" && v2Req.EnterpriseProfile != nil {
			p := v2Req.EnterpriseProfile
			review.EnterpriseName = &p.EnterpriseName
			review.UnifiedSocialCreditCode = &p.UnifiedSocialCreditCode
			review.EnterpriseYears = &p.EnterpriseYears
			review.MainBusiness = &p.MainBusiness
			review.MonthlyRevenue = &p.MonthlyRevenue
			review.CreditStatus = p.CreditStatus
			review.CreditQuery1M = p.CreditQuery1M
			review.CreditQuery3M = p.CreditQuery3M
			review.CreditQuery6M = p.CreditQuery6M
			review.ControllerCooperate = &p.ControllerCooperate
			review.EnterpriseHighlights = p.Highlights
			debt := make([]model.DebtDetail, len(p.DebtDetails))
			for i, d := range p.DebtDetails {
				debt[i] = d
				debt[i].ID = uuid.New().String()
				debt[i].DebtOwnerType = "enterprise"
			}
			review.DebtDetails = debt
		}
	} else {
		// V1 fallback path
		var v1Req model.CreateReviewRequest
		if err := json.Unmarshal(bodyBytes, &v1Req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		visitTime, err := time.Parse("2006-01-02T15:04", v1Req.VisitTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_time format, expected YYYY-MM-DDTHH:MM"})
			return
		}

		review = model.Review{
			ID:              uuid.New().String(),
			CustomerName:    v1Req.CustomerName,
			Gender:          v1Req.Gender,
			Age:             v1Req.Age,
			MaritalStatus:   v1Req.MaritalStatus,
			LoanAmount:      v1Req.LoanAmount,
			IsEnterprise:    v1Req.IsEnterprise,
			MainBank:        v1Req.MainBank,
			CreditStatus:    v1Req.CreditStatus,
			CreditQuery1M:   v1Req.CreditQuery1M,
			CreditQuery3M:   v1Req.CreditQuery3M,
			CreditQuery6M:   v1Req.CreditQuery6M,
			SpouseInfo:      v1Req.SpouseInfo,
			SpouseCooperate: v1Req.SpouseCooperate,
			Highlights:      v1Req.Highlights,
			CanMatch:        v1Req.CanMatch,
			VisitTime:       visitTime,
			CreatedBy:       uuid.New().String(),
			CustomerType:    "individual", // V1 default
			DebtDetails:     v1Req.DebtDetails,
		}
	}

	// Assign UUIDs to debt details
	for i := range review.DebtDetails {
		if review.DebtDetails[i].ID == "" {
			review.DebtDetails[i].ID = uuid.New().String()
		}
		if review.DebtDetails[i].DebtOwnerType == "" {
			review.DebtDetails[i].DebtOwnerType = review.CustomerType
		}
	}

	if err := h.svc.Create(&review); err != nil {
		log.Printf("failed to create review: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}

	c.JSON(http.StatusCreated, review)
}

// List handles GET /api/v1/reviews.
func (h *ReviewHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.svc.List(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetByID handles GET /api/v1/reviews/:id.
func (h *ReviewHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	review, err := h.svc.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// Update handles PUT /api/v1/reviews/:id with V1/V2 dual-protocol support.
func (h *ReviewHandler) Update(c *gin.Context) {
	id := c.Param("id")

	// Read raw body for dual-protocol detection
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	var review model.Review
	review.ID = id

	// Try V2 grouped payload first
	var v2Req model.CreateReviewV2Request
	if err := json.Unmarshal(bodyBytes, &v2Req); err == nil && v2Req.CustomerType != "" {
		visitTime, err := time.Parse("2006-01-02T15:04", v2Req.Common.VisitTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_time format, expected YYYY-MM-DDTHH:MM"})
			return
		}

		review.CustomerName = v2Req.Common.CustomerName
		review.Gender = v2Req.Common.Gender
		review.Age = v2Req.Common.Age
		review.MaritalStatus = v2Req.Common.MaritalStatus
		review.LoanAmount = v2Req.Common.LoanAmount
		review.IsEnterprise = v2Req.Common.IsEnterprise
		review.CanMatch = v2Req.Common.CanMatch
		review.VisitTime = visitTime
		review.CustomerType = v2Req.CustomerType

		if v2Req.CustomerType == "individual" && v2Req.IndividualProfile != nil {
			p := v2Req.IndividualProfile
			review.MainBank = p.MainBank
			review.CreditStatus = p.CreditStatus
			review.CreditQuery1M = p.CreditQuery1M
			review.CreditQuery3M = p.CreditQuery3M
			review.CreditQuery6M = p.CreditQuery6M
			review.SpouseInfo = p.SpouseInfo
			review.SpouseCooperate = p.SpouseCooperate
			review.Highlights = p.Highlights
			// Clear enterprise fields when switching to individual
			review.EnterpriseName = nil
			review.UnifiedSocialCreditCode = nil
			review.EnterpriseYears = nil
			review.MainBusiness = nil
			review.MonthlyRevenue = nil
			review.ControllerCooperate = nil
			review.EnterpriseHighlights = nil
			debt := make([]model.DebtDetail, len(p.DebtDetails))
			for i, d := range p.DebtDetails {
				debt[i] = d
				if debt[i].ID == "" {
					debt[i].ID = uuid.New().String()
				}
				debt[i].DebtOwnerType = "individual"
			}
			review.DebtDetails = debt
		}

		if v2Req.CustomerType == "enterprise" && v2Req.EnterpriseProfile != nil {
			p := v2Req.EnterpriseProfile
			review.EnterpriseName = &p.EnterpriseName
			review.UnifiedSocialCreditCode = &p.UnifiedSocialCreditCode
			review.EnterpriseYears = &p.EnterpriseYears
			review.MainBusiness = &p.MainBusiness
			review.MonthlyRevenue = &p.MonthlyRevenue
			review.CreditStatus = p.CreditStatus
			review.CreditQuery1M = p.CreditQuery1M
			review.CreditQuery3M = p.CreditQuery3M
			review.CreditQuery6M = p.CreditQuery6M
			review.ControllerCooperate = &p.ControllerCooperate
			review.EnterpriseHighlights = p.Highlights
			debt := make([]model.DebtDetail, len(p.DebtDetails))
			for i, d := range p.DebtDetails {
				debt[i] = d
				if debt[i].ID == "" {
					debt[i].ID = uuid.New().String()
				}
				debt[i].DebtOwnerType = "enterprise"
			}
			review.DebtDetails = debt
		}
	} else {
		// V1 fallback path
		var v1Req model.CreateReviewRequest
		if err := json.Unmarshal(bodyBytes, &v1Req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		visitTime, err := time.Parse("2006-01-02T15:04", v1Req.VisitTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit_time format, expected YYYY-MM-DDTHH:MM"})
			return
		}

		review.CustomerName = v1Req.CustomerName
		review.Gender = v1Req.Gender
		review.Age = v1Req.Age
		review.MaritalStatus = v1Req.MaritalStatus
		review.LoanAmount = v1Req.LoanAmount
		review.IsEnterprise = v1Req.IsEnterprise
		review.MainBank = v1Req.MainBank
		review.CreditStatus = v1Req.CreditStatus
		review.CreditQuery1M = v1Req.CreditQuery1M
		review.CreditQuery3M = v1Req.CreditQuery3M
		review.CreditQuery6M = v1Req.CreditQuery6M
		review.SpouseInfo = v1Req.SpouseInfo
		review.SpouseCooperate = v1Req.SpouseCooperate
		review.Highlights = v1Req.Highlights
		review.CanMatch = v1Req.CanMatch
		review.VisitTime = visitTime
		review.DebtDetails = v1Req.DebtDetails
		// V1 default: individual
		if review.CustomerType == "" {
			review.CustomerType = "individual"
		}
	}

	if err := h.svc.Update(&review); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update review"})
		return
	}

	c.JSON(http.StatusOK, review)
}

// ListByUser handles GET /api/v1/reviews/mine for mobile users to see their own reviews.
func (h *ReviewHandler) ListByUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.svc.ListByUser(userID.(string), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reviews"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Delete handles DELETE /api/v1/reviews/:id.
func (h *ReviewHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "review deleted"})
}