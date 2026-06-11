package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"digital-finance-services/internal/model"
)

// QRCodeRepository handles database operations for QR code records.
type QRCodeRepository struct {
	db *sql.DB
}

// NewQRCodeRepository creates a new QRCodeRepository.
func NewQRCodeRepository(db *sql.DB) *QRCodeRepository {
	return &QRCodeRepository{db: db}
}

// Create inserts a new QR code record.
func (r *QRCodeRepository) Create(record *model.QRCodeRecord) error {
	query := `
		INSERT INTO qrcode_records (target_url, channel, campaign, note, final_url, access_url, status, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(query,
		record.TargetURL,
		nullString(record.Channel),
		nullString(record.Campaign),
		nullString(record.Note),
		record.FinalURL,
		record.AccessURL,
		record.Status,
		record.CreatedBy,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
}

// List returns paginated records ordered by creation time desc.
func (r *QRCodeRepository) List(page, pageSize int, status string) ([]model.QRCodeRecord, int64, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM qrcode_records %s", whereClause)
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT id, target_url, COALESCE(channel, ''), COALESCE(campaign, ''), COALESCE(note, ''), final_url, COALESCE(access_url, ''), status, created_by, created_at, updated_at
		FROM qrcode_records %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]model.QRCodeRecord, 0)
	for rows.Next() {
		var item model.QRCodeRecord
		if err := rows.Scan(
			&item.ID,
			&item.TargetURL,
			&item.Channel,
			&item.Campaign,
			&item.Note,
			&item.FinalURL,
			&item.AccessURL,
			&item.Status,
			&item.CreatedBy,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, item)
	}

	return records, total, rows.Err()
}

// UpdateStatus changes record status.
func (r *QRCodeRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(
		"UPDATE qrcode_records SET status = $1, updated_at = NOW() WHERE id = $2",
		status, id,
	)
	return err
}

// UpdateAccessURL updates persisted access_url.
func (r *QRCodeRepository) UpdateAccessURL(id, accessURL string) error {
	_, err := r.db.Exec(
		"UPDATE qrcode_records SET access_url = $1, updated_at = NOW() WHERE id = $2",
		accessURL, id,
	)
	return err
}

// GetByID fetches a QR code record by id.
func (r *QRCodeRepository) GetByID(id string) (*model.QRCodeRecord, error) {
	var item model.QRCodeRecord
	err := r.db.QueryRow(`
		SELECT id, target_url, COALESCE(channel, ''), COALESCE(campaign, ''), COALESCE(note, ''), final_url, COALESCE(access_url, ''), status, created_by, created_at, updated_at
		FROM qrcode_records
		WHERE id = $1`, id,
	).Scan(
		&item.ID,
		&item.TargetURL,
		&item.Channel,
		&item.Campaign,
		&item.Note,
		&item.FinalURL,
		&item.AccessURL,
		&item.Status,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
