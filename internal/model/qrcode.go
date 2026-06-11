package model

import "time"

// QRCodeRecord represents a persisted QR code configuration in admin.
type QRCodeRecord struct {
	ID        string    `json:"id"`
	TargetURL string    `json:"target_url"`
	Channel   string    `json:"channel"`
	Campaign  string    `json:"campaign"`
	Note      string    `json:"note"`
	FinalURL  string    `json:"final_url"`
	AccessURL string    `json:"access_url,omitempty"`
	Status    string    `json:"status"` // active / disabled
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateQRCodeRequest is request body for creating QR code record.
type CreateQRCodeRequest struct {
	TargetURL     string `json:"target_url" binding:"required,url"`
	Channel       string `json:"channel,omitempty"`
	Campaign      string `json:"campaign,omitempty"`
	Note          string `json:"note,omitempty"`
	AccessBaseURL string `json:"access_base_url,omitempty"`
}

// UpdateQRCodeStatusRequest is request body for changing QR code status.
type UpdateQRCodeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active disabled"`
}

// QRCodeListResponse wraps paginated qr code list.
type QRCodeListResponse struct {
	Data       []QRCodeRecord `json:"data"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalCount int64          `json:"total_count"`
}
