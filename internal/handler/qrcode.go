package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"digital-finance-services/internal/model"
)

// qrCodeService defines QR code service interface used by handler.
type qrCodeService interface {
	Create(operatorID, operatorRole string, req *model.CreateQRCodeRequest) (*model.QRCodeRecord, error)
	List(operatorRole string, page, pageSize int, status string) (*model.QRCodeListResponse, error)
	UpdateStatus(operatorID, operatorRole, id string, req *model.UpdateQRCodeStatusRequest) error
	ResolveAccess(id string) (string, error)
}

// QRCodeHandler handles admin QR code APIs.
type QRCodeHandler struct {
	svc qrCodeService
}

// NewQRCodeHandler creates a QRCodeHandler.
func NewQRCodeHandler(svc qrCodeService) *QRCodeHandler {
	return &QRCodeHandler{svc: svc}
}

// Create handles POST /api/v1/admin/qrcodes.
func (h *QRCodeHandler) Create(c *gin.Context) {
	operatorID, operatorRole := getOperator(c)

	var req model.CreateQRCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	record, err := h.svc.Create(operatorID, operatorRole, &req)
	if err != nil {
		handleCommonError(c, err)
		return
	}

	c.JSON(http.StatusCreated, record)
}

// List handles GET /api/v1/admin/qrcodes.
func (h *QRCodeHandler) List(c *gin.Context) {
	_, operatorRole := getOperator(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	resp, err := h.svc.List(operatorRole, page, pageSize, status)
	if err != nil {
		handleCommonError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateStatus handles PUT /api/v1/admin/qrcodes/:id/status.
func (h *QRCodeHandler) UpdateStatus(c *gin.Context) {
	operatorID, operatorRole := getOperator(c)
	id := c.Param("id")

	var req model.UpdateQRCodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR: " + err.Error()})
		return
	}

	if err := h.svc.UpdateStatus(operatorID, operatorRole, id, &req); err != nil {
		handleCommonError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态已更新"})
}

// Visit handles GET /api/v1/qrcodes/:id/visit.
func (h *QRCodeHandler) Visit(c *gin.Context) {
	id := c.Param("id")
	targetURL, err := h.svc.ResolveAccess(id)
	if err != nil {
		handleVisitError(c, err)
		return
	}

	c.Redirect(http.StatusFound, targetURL)
}

func handleVisitError(c *gin.Context, err error) {
	errStr := err.Error()

	status := http.StatusInternalServerError
	title := "二维码暂不可用"
	message := "服务暂时不可用，请稍后重试。"

	if strings.HasPrefix(errStr, "QRCODE_NOT_FOUND") {
		status = http.StatusNotFound
		title = "二维码不存在"
		message = "该二维码可能已失效，或链接不完整。请联系管理员重新获取二维码。"
	} else if strings.HasPrefix(errStr, "QRCODE_DISABLED") {
		status = http.StatusGone
		title = "二维码已停用"
		message = "该二维码已被管理员停用。请联系管理员获取新的有效二维码。"
	}

	html := fmt.Sprintf(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>%s</title>
  <style>
    :root { color-scheme: light; }
    body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #f1f5f9; color: #0f172a; }
    .wrap { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 24px; }
    .card { width: 100%%; max-width: 460px; background: #fff; border: 1px solid #dbeafe; border-radius: 16px; padding: 24px; box-shadow: 0 8px 28px rgba(15, 23, 42, 0.08); }
    .badge { display: inline-block; font-size: 12px; color: #1d4ed8; background: #dbeafe; border-radius: 999px; padding: 4px 10px; }
    h1 { margin: 14px 0 8px; font-size: 22px; line-height: 1.35; }
    p { margin: 0; font-size: 14px; line-height: 1.7; color: #334155; }
  </style>
</head>
<body>
  <div class="wrap">
    <div class="card">
      <span class="badge">Digital Finance</span>
      <h1>%s</h1>
      <p>%s</p>
    </div>
  </div>
</body>
</html>`, title, title, message)

	c.Data(status, "text/html; charset=utf-8", []byte(html))
}

func getOperator(c *gin.Context) (string, string) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	return userID.(string), role.(string)
}

func handleCommonError(c *gin.Context, err error) {
	errStr := err.Error()
	code := "INTERNAL_ERROR"
	status := http.StatusInternalServerError

	mapping := []struct {
		Code   string
		Status int
	}{
		{"AUTH_FORBIDDEN", http.StatusForbidden},
		{"VALIDATION_ERROR", http.StatusBadRequest},
		{"QRCODE_NOT_FOUND", http.StatusNotFound},
		{"QRCODE_DISABLED", http.StatusGone},
	}
	for _, pair := range mapping {
		if strings.HasPrefix(errStr, pair.Code) {
			code = pair.Code
			status = pair.Status
			break
		}
	}

	msg := errStr
	if idx := strings.Index(errStr, ": "); idx != -1 {
		msg = errStr[idx+2:]
	}
	if code == "INTERNAL_ERROR" {
		msg = "服务器内部错误"
	}

	c.JSON(status, gin.H{"error": code + ": " + msg})
}
