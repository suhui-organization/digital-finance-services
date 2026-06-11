package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"digital-finance-services/internal/model"
)

// UserRepository handles database operations for users and auth.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// --- User CRUD ---

// CreateUser inserts a new user.
func (r *UserRepository) CreateUser(u *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, display_name, phone, email, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query,
		u.Username, u.PasswordHash, u.DisplayName,
		nullString(u.Phone), nullString(u.Email),
		u.Role, u.Status,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// GetUserByID fetches a user by ID.
func (r *UserRepository) GetUserByID(id string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, display_name, COALESCE(phone, ''), COALESCE(email, ''), role, status, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&u.ID, &u.Username, &u.DisplayName, &u.Phone, &u.Email,
		&u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// GetUserByUsername fetches a user by username.
func (r *UserRepository) GetUserByUsername(username string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, username, password_hash, display_name, COALESCE(phone, ''), COALESCE(email, ''), role, status, last_login_at, created_at, updated_at
		FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(
		&u.ID, &u.Username, &u.PasswordHash, &u.DisplayName,
		&u.Phone, &u.Email, &u.Role, &u.Status,
		&u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// ListUsers returns paginated user list with optional filters.
func (r *UserRepository) ListUsers(page, pageSize int, keyword, role, status string) ([]model.User, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR display_name ILIKE $%d OR phone ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx+1, argIdx+2, argIdx+3))
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw)
		argIdx += 4
	}
	if role != "" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, role)
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// List
	offset := (page - 1) * pageSize
	selectQuery := fmt.Sprintf(`
		SELECT id, username, display_name, COALESCE(phone, ''), COALESCE(email, ''), role, status, last_login_at, created_at, updated_at
		FROM users %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.DisplayName, &u.Phone, &u.Email,
			&u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

// UpdateUser updates user fields.
func (r *UserRepository) UpdateUser(id string, req *model.UpdateUserRequest) (*model.User, error) {
	var setClauses []string
	var args []interface{}
	argIdx := 1

	if req.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, *req.DisplayName)
		argIdx++
	}
	if req.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, nullStringPtr(req.Phone))
		argIdx++
	}
	if req.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, nullStringPtr(req.Email))
		argIdx++
	}
	if req.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *req.Role)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	if len(setClauses) == 1 {
		return r.GetUserByID(id)
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, id)

	if _, err := r.db.Exec(query, args...); err != nil {
		return nil, err
	}
	return r.GetUserByID(id)
}

// UpdateUserStatus updates only the status field.
func (r *UserRepository) UpdateUserStatus(id, status string) error {
	_, err := r.db.Exec("UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	return err
}

// UpdatePassword updates the user's password hash.
func (r *UserRepository) UpdatePassword(id, passwordHash string) error {
	_, err := r.db.Exec("UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2", passwordHash, id)
	return err
}

// UpdateLastLogin updates last_login_at.
func (r *UserRepository) UpdateLastLogin(id string) error {
	_, err := r.db.Exec("UPDATE users SET last_login_at = NOW() WHERE id = $1", id)
	return err
}

// --- Sessions ---

// CreateSession inserts a new refresh token session.
func (r *UserRepository) CreateSession(s *model.UserSession) error {
	query := `
		INSERT INTO user_sessions (user_id, refresh_token_hash, client_type, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return r.db.QueryRow(query,
		s.UserID, s.RefreshTokenHash, s.ClientType, s.ExpiresAt,
	).Scan(&s.ID, &s.CreatedAt)
}

// GetSessionByRefreshHash finds a valid session by refresh token hash.
func (r *UserRepository) GetSessionByRefreshHash(hash string) (*model.UserSession, error) {
	s := &model.UserSession{}
	query := `SELECT id, user_id, refresh_token_hash, client_type, expires_at, revoked_at, created_at
		FROM user_sessions WHERE refresh_token_hash = $1`
	err := r.db.QueryRow(query, hash).Scan(
		&s.ID, &s.UserID, &s.RefreshTokenHash, &s.ClientType,
		&s.ExpiresAt, &s.RevokedAt, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

// RevokeSession marks a session as revoked.
func (r *UserRepository) RevokeSession(id string) error {
	_, err := r.db.Exec("UPDATE user_sessions SET revoked_at = NOW() WHERE id = $1", id)
	return err
}

// RevokeAllUserSessions revokes all sessions for a user.
func (r *UserRepository) RevokeAllUserSessions(userID string) error {
	_, err := r.db.Exec(
		"UPDATE user_sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL",
		userID,
	)
	return err
}

// CleanupExpiredSessions removes expired sessions.
func (r *UserRepository) CleanupExpiredSessions() error {
	_, err := r.db.Exec("DELETE FROM user_sessions WHERE expires_at < NOW()")
	return err
}

// --- Audit Logs ---

// CreateAuditLog inserts an audit log entry.
func (r *UserRepository) CreateAuditLog(operatorID, action, targetType, targetID, payloadJSON string) error {
	_, err := r.db.Exec(
		`INSERT INTO audit_logs (operator_user_id, action, target_type, target_id, payload_snapshot)
		 VALUES ($1, $2, $3, $4, $5::jsonb)`,
		operatorID, action, targetType, targetID, payloadJSON,
	)
	return err
}

// --- Helpers ---

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullStringPtr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}