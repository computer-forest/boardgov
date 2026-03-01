package unit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yourorg/boardgov/config"
	"github.com/yourorg/boardgov/internal/middleware"
	"github.com/yourorg/boardgov/pkg/crypto"
	"github.com/google/uuid"
)

func setupTestRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.JWTAuth(cfg))
	r.GET("/protected", func(c *gin.Context) {
		uid, ok := middleware.GetUserID(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})
	return r
}

func TestJWTAuth_ValidToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:  "test-secret-32-chars-long-enough!!",
			RefreshSecret: "refresh-secret-32-chars-long-ok!!",
			AccessExpiry:  15 * 60 * 1e9,  // 15 min in nanoseconds
			RefreshExpiry: 7 * 24 * 3600 * 1e9,
		},
	}

	userID := uuid.New()
	tokens, err := crypto.GenerateTokenPair(userID, "test@example.com",
		cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	if err != nil {
		t.Fatal(err)
	}

	r := setupTestRouter(cfg)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuth_MissingToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{AccessSecret: "test-secret-32-chars-long-enough!!"},
	}
	r := setupTestRouter(cfg)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{AccessSecret: "test-secret-32-chars-long-enough!!"},
	}
	r := setupTestRouter(cfg)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer tampered.invalid.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSecureHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.SecureHeaders())
	r.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{}) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	headers := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"Strict-Transport-Security",
		"Content-Security-Policy",
	}
	for _, h := range headers {
		if w.Header().Get(h) == "" {
			t.Errorf("missing security header: %s", h)
		}
	}
}

func TestRateLimiter_Allows(t *testing.T) {
	cfg := &config.Config{
		Security: config.SecurityConfig{RateLimitRPS: 100, RateLimitBurst: 200},
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RateLimiter(cfg))
	r.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{}) })

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
