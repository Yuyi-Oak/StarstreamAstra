package router

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"Zjmf-kvm/internal/config"
	"Zjmf-kvm/internal/db"
	"Zjmf-kvm/internal/handler"
)

func RegisterRoutes(r *gin.Engine, dbConn *db.DBConn, cfg *config.Config) {
	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", handler.RegisterHandler(dbConn.Gorm))

	jwtSecret, tokenTTL := resolveJWTConfig(cfg)
	_ = tokenTTL
	auth.POST("/login", handler.LoginHandler(dbConn.Gorm, jwtSecret, tokenTTL))

	protected := api.Group("")
	protected.Use(AuthMiddleware(jwtSecret))

	vmGroup := protected.Group("/vm")
	handler.RegisterVMHandlers(vmGroup, dbConn.Gorm)

	adminGroup := vmGroup.Group("/admin")
	adminGroup.Use(RequireRole("admin"))
	adminGroup.POST("/create", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"Message": "VM create placeholder (admin)"})
	})
}

func resolveJWTConfig(cfg *config.Config) (string, time.Duration) {
	secret := "please-change-this-secret"
	ttl := 24 * time.Hour

	if s := os.Getenv("JWT_SECRET"); s != "" {
		secret = s
	}
	if t := os.Getenv("JWT_TTL_SECONDS"); t != "" {
		if v, err := strconv.ParseInt(t, 10, 64); err == nil && v > 0 {
			ttl = time.Duration(v) * time.Second
		}
	}

	type jwtProvider interface {
		GetJWTSecret() string
		GetJWTTTLSeconds() int64
	}
	if cfg != nil {
		if p, ok := interface{}(cfg).(jwtProvider); ok {
			if s := p.GetJWTSecret(); s != "" {
				secret = s
			}
			if v := p.GetJWTTTLSeconds(); v > 0 {
				ttl = time.Duration(v) * time.Second
			}
		}
	}
	return secret, ttl
}

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "missing authorization header"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "invalid authorization header"})
			return
		}
		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenMalformed
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error": "invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("claims", claims)
			if sub, ok := claims["sub"].(float64); ok {
				c.Set("user_id", uint(sub))
			}
		}
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsI, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"Error": "no claims found"})
			return
		}
		claims, ok := claimsI.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"Error": "invalid claims"})
			return
		}
		if r, ok := claims["role"].(string); !ok || r != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"Error": "insufficient role"})
			return
		}
		c.Next()
	}
}
