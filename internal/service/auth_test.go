package service

import (
	"testing"
)

// TestHashToken 验证哈希函数产生一致的输出
func TestHashToken(t *testing.T) {
	token := "test-refresh-token-12345"
	hash1 := hashToken(token)
	hash2 := hashToken(token)

	if hash1 != hash2 {
		t.Errorf("hashToken should be deterministic: got %s and %s", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("SHA-256 hex digest should be 64 characters, got %d", len(hash1))
	}
}

// TestHashTokenDifferentInputs 验证不同输入产生不同哈希
func TestHashTokenDifferentInputs(t *testing.T) {
	hash1 := hashToken("token-a")
	hash2 := hashToken("token-b")

	if hash1 == hash2 {
		t.Error("different tokens should produce different hashes")
	}
}

// TestGenerateRandomToken 验证随机 token 的长度和格式
func TestGenerateRandomToken(t *testing.T) {
	// 测试默认长度 64
	token := generateRandomToken(64)
	if len(token) != 128 { // hex encoding doubles the length
		t.Errorf("expected 128 hex chars from 64 random bytes, got %d", len(token))
	}

	// 测试自定义长度
	token32 := generateRandomToken(32)
	if len(token32) != 64 {
		t.Errorf("expected 64 hex chars from 32 random bytes, got %d", len(token32))
	}
}

// TestGenerateRandomTokenUniqueness 验证连续生成的 token 是唯一的
func TestGenerateRandomTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token := generateRandomToken(32)
		if tokens[token] {
			t.Error("random tokens should be unique, found collision")
		}
		tokens[token] = true
	}
}

// TestAccessTokenTTL 验证常量值符合规范
func TestAccessTokenTTL(t *testing.T) {
	if AccessTokenTTL.Hours() != 2 {
		t.Errorf("access token TTL should be 2 hours, got %v", AccessTokenTTL)
	}
}

// TestRefreshTokenTTL 验证刷新令牌过期时间
func TestRefreshTokenTTL(t *testing.T) {
	if RefreshTokenTTL.Hours() != 7*24 {
		t.Errorf("refresh token TTL should be 7 days (168 hours), got %v", RefreshTokenTTL)
	}
}
