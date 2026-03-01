package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"github.com/yourorg/boardgov/config"
	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/pkg/crypto"
	"github.com/yourorg/boardgov/pkg/logger"
)

const (
	ContextUserID    = "user_id"
	ContextEmail     = "email"
	ContextUserEmail = "email"
)

// ─── CORS Middleware ──────────────────────────────────────────────────────────

func CORS(cfg *config.Config) gin.HandlerFunc {
	allowed := make(map[string]bool, len(cfg.Security.CORSAllowedOrigins))
	for _, o := range cfg.Security.CORSAllowedOrigins {
		allowed[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, X-Request-ID")
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ─── JWT Auth Middleware ──────────────────────────────────────────────────────

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := crypto.ValidateToken(tokenStr, cfg.JWT.AccessSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		c.Set(ContextUserID, userID)
		c.Set(ContextEmail, claims.Email)
		c.Next()
	}
}

// ─── RBAC Middleware ──────────────────────────────────────────────────────────

func RequirePermission(roleRepo domain.RoleRepository, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ContextUserID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		uid, ok := userID.(uuid.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		has, err := roleRepo.HasPermission(c.Request.Context(), uid, resource, action)
		if err != nil || !has {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// ─── Rate Limiter ─────────────────────────────────────────────────────────────

type ipLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	rps      rate.Limit
	burst    int
}

func newIPLimiter(rps, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

func (l *ipLimiter) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()
	if lim, ok := l.limiters[ip]; ok {
		return lim
	}
	lim := rate.NewLimiter(l.rps, l.burst)
	l.limiters[ip] = lim
	return lim
}

func RateLimiter(cfg *config.Config) gin.HandlerFunc {
	limiter := newIPLimiter(cfg.Security.RateLimitRPS, cfg.Security.RateLimitBurst)
	return func(c *gin.Context) {
		if !limiter.get(c.ClientIP()).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// ─── IP Whitelist ─────────────────────────────────────────────────────────────

func IPWhitelist(cfg *config.Config) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, ip := range cfg.Security.IPWhitelist {
		whitelist[strings.TrimSpace(ip)] = true
	}
	return func(c *gin.Context) {
		if len(whitelist) == 0 {
			c.Next()
			return
		}
		if !whitelist[c.ClientIP()] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.Next()
	}
}

// ─── Secure Headers ───────────────────────────────────────────────────────────

func SecureHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Next()
	}
}

// ─── Panic Recovery ───────────────────────────────────────────────────────────

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Log.Error().Interface("panic", err).
					Str("path", c.Request.URL.Path).Msg("panic recovered")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

// ─── Structured Request Logger ────────────────────────────────────────────────

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		logger.Log.Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Str("ip", c.ClientIP()).
			Dur("latency", time.Since(start)).
			Str("user_agent", c.Request.UserAgent()).
			Msg("request")
	}
}

// ─── Context Helpers ──────────────────────────────────────────────────────────

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(ContextUserID)
	if !ok {
		return uuid.Nil, false
	}
	uid, ok := v.(uuid.UUID)
	return uid, ok
}
