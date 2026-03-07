package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/adithyvisnu/user-go/internal/models"
)

const (
	ctxTenantID = "tenant_id"
	ctxUserID   = "user_id"
)

// Auth validates the Bearer JWT and injects tenant_id + user_id into the context.
// In production, replace the stub below with a real JWT verification library
// (e.g. github.com/golang-jwt/jwt/v5) and validate against your signing key.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing or malformed Authorization header",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")

		// ── TODO: replace with real JWT verification ──────────────────────────
		// claims, err := jwtutil.Verify(token, signingKey)
		// if err != nil { c.AbortWithStatusJSON(401, ...) }
		// tenantID := claims.TenantID
		// userID   := claims.Subject
		// ─────────────────────────────────────────────────────────────────────

		// Stub: decode a fake "tenantID.userID" token for local development
		parts := strings.SplitN(token, ".", 2)
		if len(parts) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid token format",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		tenantID, err := uuid.Parse(parts[0])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid tenant_id in token",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		userID, err := uuid.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid user_id in token",
				Code:  "UNAUTHORIZED",
			})
			return
		}

		c.Set(ctxTenantID, tenantID)
		c.Set(ctxUserID, userID)
		c.Next()
	}
}

// MustTenantID extracts the tenant UUID from gin context.
// Panics if called outside of Auth middleware — intentional to catch wiring bugs.
func MustTenantID(c *gin.Context) uuid.UUID {
	val, exists := c.Get(ctxTenantID)
	if !exists {
		panic("MustTenantID: tenant_id not set — Auth middleware missing?")
	}
	return val.(uuid.UUID)
}

// MustUserID extracts the authenticated user UUID from gin context.
func MustUserID(c *gin.Context) uuid.UUID {
	val, exists := c.Get(ctxUserID)
	if !exists {
		panic("MustUserID: user_id not set — Auth middleware missing?")
	}
	return val.(uuid.UUID)
}

// RequestLogger logs method, path, status, and latency for every request.
func RequestLogger() gin.HandlerFunc {
	return gin.Logger()
}

// Recovery returns 500 on panics instead of crashing.
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}
