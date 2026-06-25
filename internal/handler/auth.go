package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"digital-finance-services/internal/model"
)

// authService defines the interface for AuthService used by this handler.
type authService interface {
	Register(operatorID string, req *model.RegisterRequest) (*model.AuthResponse, error)
	Login(req *model.LoginRequest) (*model.AuthResponse, error)
	Refresh(refreshToken string) (*model.AuthResponse, error)
	Logout(userID, refreshToken string) error
	GetMe(userID string) (*model.User, error)
	ListUsers(operatorID, operatorRole string, page, pageSize int, keyword, role, status string) (*model.UserListResponse, error)
	GetUser(operatorID, operatorRole, targetID string) (*model.User, error)
	CreateUser(operatorID, operatorRole string, req *model.CreateUserRequest) (*model.User, error)
	UpdateUser(operatorID, operatorRole, targetID string, req *model.UpdateUserRequest) (*model.User, error)
	UpdateUserStatus(operatorID, operatorRole, targetID string, req *model.UpdateUserStatusRequest) error
	ResetPassword(operatorID, operatorRole, targetID string, req *model.ResetPasswordRequest) error
}

// AuthHandler handles auth and user management HTTP requests.
type AuthHandler struct {
	svc authService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(svc authService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// --- Auth endpoints (DESIGN_DOC 29.1) ---

// Register handles POST /api/v1/auth/register (Admin registration).
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	resp, err := h.svc.Register("system", &req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles POST /api/v1/auth/login (Admin/Mobile).
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	if req.ClientType != "admin" && req.ClientType != "mobile" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: client_type must be 'admin' or 'mobile'"})
		return
	}

	resp, err := h.svc.Login(&req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	resp, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req model.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	if err := h.svc.Logout(userID.(string), req.RefreshToken); err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已注销"})
}

// GetMe handles GET /api/v1/auth/me.
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.svc.GetMe(userID.(string))
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// --- Admin User Management endpoints (DESIGN_DOC 29.2) ---

// ListUsers handles GET /api/v1/admin/users.
func (h *AuthHandler) ListUsers(c *gin.Context) {
	operatorID, operatorRole := h.getOperator(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	keyword := c.Query("keyword")
	role := c.Query("role")
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, err := h.svc.ListUsers(operatorID, operatorRole, page, pageSize, keyword, role, status)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUser handles GET /api/v1/admin/users/:id.
func (h *AuthHandler) GetUser(c *gin.Context) {
	operatorID, operatorRole := h.getOperator(c)
	targetID := c.Param("id")

	user, err := h.svc.GetUser(operatorID, operatorRole, targetID)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /api/v1/admin/users.
func (h *AuthHandler) CreateUser(c *gin.Context) {
	operatorID, operatorRole := h.getOperator(c)

	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	user, err := h.svc.CreateUser(operatorID, operatorRole, &req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /api/v1/admin/users/:id.
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	operatorID, operatorRole := h.getOperator(c)
	targetID := c.Param("id")

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	user, err := h.svc.UpdateUser(operatorID, operatorRole, targetID, &req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUserStatus handles PUT /api/v1/admin/users/:id/status.
func (h *AuthHandler) UpdateUserStatus(c *gin.Context) {
	operatorID, operatorRole := h.getOperator(c)
	targetID := c.Param("id")

	var req model.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	if err := h.svc.UpdateUserStatus(operatorID, operatorRole, targetID, &req); err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

// ResetPassword handles PUT /api/v1/admin/users/:id/password/reset.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	operatorID, operatorRole := h.getOperator(c)
	targetID := c.Param("id")

	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	if err := h.svc.ResetPassword(operatorID, operatorRole, targetID, &req); err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码已重置"})
}

// --- Helpers ---

func (h *AuthHandler) getOperator(c *gin.Context) (string, string) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	return userID.(string), role.(string)
}

// handleAuthError maps service errors to HTTP status codes (DESIGN_DOC 29.4).
func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	errStr := err.Error()

	code := "INTERNAL_ERROR"
	status := http.StatusInternalServerError

	// Try to parse structured error
	if idx := strings.Index(errStr, ":"); idx != -1 {
		code = errStr[:idx]
	}

	switch code {
	case "AUTH_INVALID_CREDENTIALS":
		status = http.StatusUnauthorized
	case "AUTH_USER_NOT_FOUND":
		status = http.StatusNotFound
	case "AUTH_ACCOUNT_DISABLED":
		status = http.StatusForbidden
	case "AUTH_TOKEN_EXPIRED", "AUTH_TOKEN_INVALID":
		status = http.StatusUnauthorized
	case "AUTH_FORBIDDEN":
		status = http.StatusForbidden
	case "USER_ALREADY_EXISTS":
		status = http.StatusConflict
	case "USER_NOT_FOUND":
		status = http.StatusNotFound
	case "USER_ROLE_ILLEGAL":
		status = http.StatusBadRequest
	case "VALIDATION_ERROR":
		status = http.StatusBadRequest
	}

	// Strip the code prefix for the message
	msg := errStr
	if idx := strings.Index(errStr, ": "); idx != -1 {
		msg = strings.TrimSpace(errStr[idx+2:])
	}

	// Never expose internal error details (DESIGN_DOC 30.2 item 5)
	if code == "INTERNAL_ERROR" {
		msg = "服务器内部错误"
	}

	c.JSON(status, gin.H{"error": code + ": " + msg})
}
