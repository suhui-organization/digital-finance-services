package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"digital-finance-services/internal/config"
	"digital-finance-services/internal/model"
	"digital-finance-services/internal/repository"
)

// QRCodeService handles QR code persistence for admin.
type QRCodeService struct {
	repo     *repository.QRCodeRepository
	userRepo *repository.UserRepository
	cfg      *config.Config
}

// NewQRCodeService creates a new QRCodeService.
func NewQRCodeService(repo *repository.QRCodeRepository, userRepo *repository.UserRepository, cfg *config.Config) *QRCodeService {
	return &QRCodeService{repo: repo, userRepo: userRepo, cfg: cfg}
}

// Create persists a QR code record.
func (s *QRCodeService) Create(operatorID, operatorRole string, req *model.CreateQRCodeRequest) (*model.QRCodeRecord, error) {
	if operatorRole != "super_admin" && operatorRole != "admin" {
		return nil, errors.New("AUTH_FORBIDDEN: 无权创建二维码")
	}

	targetURL := strings.TrimSpace(req.TargetURL)
	parsed, err := url.Parse(targetURL)
	if err != nil || !parsed.IsAbs() {
		return nil, errors.New("VALIDATION_ERROR: 目标地址不是有效的绝对 URL")
	}

	query := parsed.Query()
	if trimmed := strings.TrimSpace(req.Channel); trimmed != "" {
		query.Set("channel", trimmed)
	}
	if trimmed := strings.TrimSpace(req.Campaign); trimmed != "" {
		query.Set("campaign", trimmed)
	}
	if trimmed := strings.TrimSpace(req.Note); trimmed != "" {
		query.Set("note", trimmed)
	}
	parsed.RawQuery = query.Encode()

	record := &model.QRCodeRecord{
		TargetURL: targetURL,
		Channel:   strings.TrimSpace(req.Channel),
		Campaign:  strings.TrimSpace(req.Campaign),
		Note:      strings.TrimSpace(req.Note),
		FinalURL:  parsed.String(),
		Status:    "active",
		CreatedBy: operatorID,
	}

	if err := s.repo.Create(record); err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	// Rebuild access URL after id is generated.
	record.AccessURL, err = s.buildAccessURL(record.ID, strings.TrimSpace(req.AccessBaseURL))
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateAccessURL(record.ID, record.AccessURL); err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	_ = s.userRepo.CreateAuditLog(operatorID, "create_qrcode", "qrcode", record.ID, recordToAuditJSON(record))
	return record, nil
}

// List returns paginated QR code records.
func (s *QRCodeService) List(operatorRole string, page, pageSize int, status string) (*model.QRCodeListResponse, error) {
	if operatorRole != "super_admin" && operatorRole != "admin" {
		return nil, errors.New("AUTH_FORBIDDEN: 无权查看二维码记录")
	}

	records, total, err := s.repo.List(page, pageSize, status)
	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	for i := range records {
		if records[i].AccessURL == "" {
			records[i].AccessURL, _ = s.buildAccessURL(records[i].ID, "")
		}
	}

	return &model.QRCodeListResponse{
		Data:       records,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	}, nil
}

// UpdateStatus enables/disables a QR code record.
func (s *QRCodeService) UpdateStatus(operatorID, operatorRole, id string, req *model.UpdateQRCodeStatusRequest) error {
	if operatorRole != "super_admin" && operatorRole != "admin" {
		return errors.New("AUTH_FORBIDDEN: 无权修改二维码状态")
	}

	if err := s.repo.UpdateStatus(id, req.Status); err != nil {
		return fmt.Errorf("INTERNAL_ERROR: %w", err)
	}

	_ = s.userRepo.CreateAuditLog(operatorID, "update_qrcode_status", "qrcode", id, fmt.Sprintf(`{"status":"%s"}`, req.Status))
	return nil
}

// ResolveAccess returns redirect target URL when QR code record is active.
func (s *QRCodeService) ResolveAccess(id string) (string, error) {
	record, err := s.repo.GetByID(id)
	if err != nil {
		return "", fmt.Errorf("INTERNAL_ERROR: %w", err)
	}
	if record == nil {
		return "", errors.New("QRCODE_NOT_FOUND: 二维码记录不存在")
	}
	if record.Status != "active" {
		return "", errors.New("QRCODE_DISABLED: 二维码已停用")
	}
	return record.FinalURL, nil
}

func (s *QRCodeService) buildAccessURL(id, accessBaseURL string) (string, error) {
	base := strings.TrimSpace(accessBaseURL)
	if base == "" {
		base = s.cfg.PublicBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil || !parsed.IsAbs() {
		return "", errors.New("VALIDATION_ERROR: 访问基址不是有效的绝对 URL")
	}
	base = strings.TrimRight(base, "/")
	return fmt.Sprintf("%s/api/v1/qrcodes/%s/visit", base, id), nil
}

func recordToAuditJSON(record *model.QRCodeRecord) string {
	return fmt.Sprintf(
		`{"target_url":%q,"channel":%q,"campaign":%q,"note":%q,"final_url":%q,"status":%q}`,
		record.TargetURL,
		record.Channel,
		record.Campaign,
		record.Note,
		record.FinalURL,
		record.Status,
	)
}
