package config

import (
	"os"
	"testing"
)

// TestLoad 验证配置加载使用默认值
func TestLoad(t *testing.T) {
	// 清除可能影响测试的环境变量
	envVars := []string{"DATABASE_URL", "REDIS_URL", "JWT_SECRET", "AI_SERVICE_URL", "PUBLIC_BASE_URL", "PORT"}
	for _, k := range envVars {
		os.Unsetenv(k)
	}

	cfg := Load()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"DatabaseURL", cfg.DatabaseURL, "postgres://localhost:5432/digital_finance?sslmode=disable"},
		{"RedisURL", cfg.RedisURL, "localhost:6379"},
		{"AIServiceURL", cfg.AIServiceURL, "http://localhost:16081"},
		{"PublicBaseURL", cfg.PublicBaseURL, "http://localhost:16080"},
		{"Port", cfg.Port, "16080"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.expected {
				t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.expected)
			}
		})
	}
}

// TestLoadFromEnv 验证从环境变量读取配置
func TestLoadFromEnv(t *testing.T) {
	os.Setenv("DATABASE_URL", "postgres://test:5432/testdb")
	os.Setenv("REDIS_URL", "test-redis:6379")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("PORT", "9999")

	defer func() {
		for _, k := range []string{"DATABASE_URL", "REDIS_URL", "JWT_SECRET", "PORT"} {
			os.Unsetenv(k)
		}
	}()

	cfg := Load()

	if cfg.DatabaseURL != "postgres://test:5432/testdb" {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://test:5432/testdb")
	}
	if cfg.RedisURL != "test-redis:6379" {
		t.Errorf("RedisURL = %q, want %q", cfg.RedisURL, "test-redis:6379")
	}
	if cfg.JWTSecret != "test-secret" {
		t.Errorf("JWTSecret = %q, want %q", cfg.JWTSecret, "test-secret")
	}
	if cfg.Port != "9999" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9999")
	}
}

// TestGetEnv 验证环境变量回退逻辑
func TestGetEnv(t *testing.T) {
	os.Unsetenv("TEST_KEY")

	// 测试回退值
	val := getEnv("TEST_KEY", "fallback")
	if val != "fallback" {
		t.Errorf("getEnv should return fallback when env is unset, got %q", val)
	}

	// 测试环境变量值
	os.Setenv("TEST_KEY", "from-env")
	defer os.Unsetenv("TEST_KEY")

	val = getEnv("TEST_KEY", "fallback")
	if val != "from-env" {
		t.Errorf("getEnv should return env value when set, got %q", val)
	}
}
