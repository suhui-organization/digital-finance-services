package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"digital-finance-services/internal/config"
	"digital-finance-services/internal/middleware"
	"digital-finance-services/internal/model"
	"digital-finance-services/internal/repository"
)

// AuthService handles authentication, user management, and RBAC.
type AuthService struct {
	repo *repository.UserRepository
	cfg  *config.Config
	db   *sql.DB
}

// NewAuthService creates a new AuthService.
func NewAuthService(repo *repository.UserRepository, cfg *config.Config, db *sql.DB) *AuthService {
	return &AuthService{repo: repo, cfg: cfg, db: db}
}

// --- Auth Constants (DESIGN_DOC 30.3) ---
const (
	AccessTokenTTL  = 2 * time.Hour
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// --- Authentication ---

// Register creates a new admin account (DESIGN_DOC 29.1).
func (s *AuthService) Register(operatorID string, req *model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if username exists
	existing, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("USER_ALREADY_EXISTS: %w", err)
	}
	if existing != nil {
		return nil, errors.New("USER_ALREADY_EXISTS: 用户名已存在")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: failed to hash password: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		Phone:        req.Phone,
		Email:        req.Email,
		Role:         "admin", // Register always creates admin role
		Status:       "active",
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: failed to create user: %w", err)
	}

	// Audit log
	s.auditLog(operatorID, "register", "user", user.ID, user)

	// Issue tokens
	return s.issueTokens(user, "admin")
}

// Login authenticates a user and returns tokens (DESIGN_DOC 29.1).
func (s *AuthService) Login(req *model.LoginRequest) (*model.AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if user == nil {
		return nil, errors.New("AUTH_USER_NOT_FOUND: 用户不存在，请检查账号是否正确")
	}

	// Check status
	if user.Status == "disabled" {
		return nil, errors.New("AUTH_ACCOUNT_DISABLED: 账号已禁用")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("AUTH_INVALID_CREDENTIALS: 密码错误")
	}

	// Validate client_type (DESIGN_DOC 27.2)
	if req.ClientType == "admin" && (user.Role != "super_admin" && user.Role != "admin") {
		return nil, errors.New("AUTH_FORBIDDEN: 无权登录管理后台")
	}
	if req.ClientType == "mobile" && (user.Role != "mobile_user" && user.Role != "admin" && user.Role != "super_admin") {
		return nil, errors.New("AUTH_FORBIDDEN: 无权登录移动端")
	}

	// Update last login
	_ = s.repo.UpdateLastLogin(user.ID)
	s.auditLog(user.ID, "login", "user", user.ID, map[string]string{"client_type": req.ClientType})

	return s.issueTokens(user, req.ClientType)
}

// Refresh exchanges a refresh token for a new access token (DESIGN_DOC 29.1).
func (s *AuthService) Refresh(refreshToken string) (*model.AuthResponse, error) {
	hash := hashToken(refreshToken)
	session, err := s.repo.GetSessionByRefreshHash(hash)
	if err != nil {
		return nil, fmt.Errorf("AUTH_TOKEN_INVALID: %w", err)
	}
	if session == nil {
		return nil, errors.New("AUTH_TOKEN_INVALID: 令牌无效")
	}
	if session.RevokedAt != nil {
		return nil, errors.New("AUTH_TOKEN_INVALID: 令牌已注销")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("AUTH_TOKEN_EXPIRED: 令牌已过期")
	}

	// Revoke old session (rotation)
	_ = s.repo.RevokeSession(session.ID)

	user, err := s.repo.GetUserByID(session.UserID)
	if err != nil || user == nil {
		return nil, errors.New("AUTH_TOKEN_INVALID: 用户不存在")
	}
	if user.Status == "disabled" {
		return nil, errors.New("AUTH_ACCOUNT_DISABLED: 账号已禁用")
	}

	return s.issueTokens(user, session.ClientType)
}

// Logout revokes a refresh token session (DESIGN_DOC 29.1).
func (s *AuthService) Logout(userID, refreshToken string) error {
	hash := hashToken(refreshToken)
	session, err := s.repo.GetSessionByRefreshHash(hash)
	if err != nil || session == nil {
		return nil // idempotent
	}
	if session.UserID != userID {
		return errors.New("AUTH_FORBIDDEN: 无权操作此会话")
	}
	return s.repo.RevokeSession(session.ID)
}

// GetMe returns the current user info (DESIGN_DOC 29.1).
func (s *AuthService) GetMe(userID string) (*model.User, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if user == nil {
		return nil, errors.New("USER_NOT_FOUND: 用户不存在")
	}
	return user, nil
}

// --- Admin User Management (DESIGN_DOC 29.2) ---

// ListUsers returns paginated user list (RBAC: super_admin sees all, admin sees non-super_admin).
func (s *AuthService) ListUsers(operatorID, operatorRole string, page, pageSize int, keyword, role, status string) (*model.UserListResponse, error) {
	// RBAC: admin cannot filter/view super_admin
	if operatorRole == "admin" && role == "super_admin" {
		return &model.UserListResponse{Data: []model.User{}, Page: page, PageSize: pageSize, TotalCount: 0}, nil
	}

	users, total, err := s.repo.ListUsers(page, pageSize, keyword, role, status)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	// Filter out super_admin if operator is admin
	if operatorRole == "admin" {
		filtered := make([]model.User, 0)
		for _, u := range users {
			if u.Role != "super_admin" {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	return &model.UserListResponse{
		Data:       users,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	}, nil
}

// GetUser returns a single user (RBAC enforced).
func (s *AuthService) GetUser(operatorID, operatorRole, targetID string) (*model.User, error) {
	user, err := s.repo.GetUserByID(targetID)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if user == nil {
		return nil, errors.New("USER_NOT_FOUND: 用户不存在")
	}

	// RBAC: admin cannot view super_admin
	if operatorRole == "admin" && user.Role == "super_admin" {
		return nil, errors.New("AUTH_FORBIDDEN: 无权查看超级管理员")
	}
	return user, nil
}

// CreateUser creates a new user (RBAC: only super_admin/admin, cannot escalate role).
func (s *AuthService) CreateUser(operatorID, operatorRole string, req *model.CreateUserRequest) (*model.User, error) {
	// RBAC: only super_admin/admin can create
	if operatorRole != "super_admin" && operatorRole != "admin" {
		return nil, errors.New("AUTH_FORBIDDEN: 无权创建用户")
	}
	// admin cannot create super_admin (DESIGN_DOC 27.3)
	if operatorRole == "admin" && req.Role == "super_admin" {
		return nil, errors.New("USER_ROLE_ILLEGAL: 无权创建超级管理员")
	}

	// Check uniqueness
	existing, _ := s.repo.GetUserByUsername(req.Username)
	if existing != nil {
		return nil, errors.New("USER_ALREADY_EXISTS: 用户名已存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		Phone:        req.Phone,
		Email:        req.Email,
		Role:         req.Role,
		Status:       "active",
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	s.auditLog(operatorID, "create_user", "user", user.ID, user)
	return user, nil
}

// UpdateUser updates user info (RBAC enforced).
func (s *AuthService) UpdateUser(operatorID, operatorRole, targetID string, req *model.UpdateUserRequest) (*model.User, error) {
	target, err := s.repo.GetUserByID(targetID)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if target == nil {
		return nil, errors.New("USER_NOT_FOUND: 用户不存在")
	}

	// RBAC: admin cannot modify super_admin
	if operatorRole == "admin" && target.Role == "super_admin" {
		return nil, errors.New("AUTH_FORBIDDEN: 无权修改超级管理员")
	}
	// admin cannot escalate anyone to super_admin
	if operatorRole == "admin" && req.Role != nil && *req.Role == "super_admin" {
		return nil, errors.New("USER_ROLE_ILLEGAL: 无权提升为超级管理员")
	}

	// Prevent self-disable (but allow self-edit)
	if targetID == operatorID && req.Status != nil && *req.Status == "disabled" {
		return nil, errors.New("AUTH_FORBIDDEN: 不能禁用自己的账号")
	}

	updated, err := s.repo.UpdateUser(targetID, req)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	s.auditLog(operatorID, "update_user", "user", targetID, req)
	return updated, nil
}

// UpdateUserStatus enables/disables a user (RBAC enforced).
func (s *AuthService) UpdateUserStatus(operatorID, operatorRole, targetID string, req *model.UpdateUserStatusRequest) error {
	target, err := s.repo.GetUserByID(targetID)
	if err != nil {
		return fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if target == nil {
		return errors.New("USER_NOT_FOUND: 用户不存在")
	}

	// RBAC
	if operatorRole == "admin" && target.Role == "super_admin" {
		return errors.New("AUTH_FORBIDDEN: 无权操作超级管理员")
	}
	if targetID == operatorID && req.Status == "disabled" {
		return errors.New("AUTH_FORBIDDEN: 不能禁用自己的账号")
	}

	if err := s.repo.UpdateUserStatus(targetID, req.Status); err != nil {
		return fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	// Revoke all sessions if disabled
	if req.Status == "disabled" {
		_ = s.repo.RevokeAllUserSessions(targetID)
	}

	s.auditLog(operatorID, "update_user_status", "user", targetID, map[string]string{"status": req.Status})
	return nil
}

// ResetPassword resets a user's password (RBAC enforced).
func (s *AuthService) ResetPassword(operatorID, operatorRole, targetID string, req *model.ResetPasswordRequest) error {
	target, err := s.repo.GetUserByID(targetID)
	if err != nil {
		return fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if target == nil {
		return errors.New("USER_NOT_FOUND: 用户不存在")
	}

	// RBAC
	if operatorRole == "admin" && target.Role == "super_admin" {
		return errors.New("AUTH_FORBIDDEN: 无权操作超级管理员")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	if err := s.repo.UpdatePassword(targetID, string(hash)); err != nil {
		return fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	// Revoke all sessions (old password invalid immediately; DESIGN_DOC 31.2)
	_ = s.repo.RevokeAllUserSessions(targetID)

	s.auditLog(operatorID, "reset_password", "user", targetID, nil)
	return nil
}

// --- Token helpers ---

func (s *AuthService) issueTokens(user *model.User, clientType string) (*model.AuthResponse, error) {
	// Access token
	accessToken, err := middleware.GenerateToken(s.cfg.JWTSecret, user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: failed to generate access token: %w", err)
	}

	// Refresh token
	refreshToken := generateRandomToken(64)
	refreshHash := hashToken(refreshToken)

	session := &model.UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshHash,
		ClientType:       clientType,
		ExpiresAt:        time.Now().Add(RefreshTokenTTL),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: failed to create session: %w", err)
	}

	// Strip sensitive fields
	userCopy := *user
	userCopy.PasswordHash = ""

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(AccessTokenTTL.Seconds()),
		User:         &userCopy,
	}, nil
}

// --- Audit ---

func (s *AuthService) auditLog(operatorID, action, targetType, targetID string, payload interface{}) {
	var payloadJSON string
	if payload != nil {
		b, _ := json.Marshal(payload)
		payloadJSON = string(b)
	}
	_ = s.repo.CreateAuditLog(operatorID, action, targetType, targetID, payloadJSON)
}

// --- Crypto helpers ---

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
