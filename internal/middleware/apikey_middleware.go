package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type APIKeyValidator interface {
	IsValid(rawKey string) bool
}

type StaticAPIKeyValidator struct {
	Expected string
}

func (v StaticAPIKeyValidator) IsValid(rawKey string) bool {
	return rawKey != "" && rawKey == v.Expected
}

func APIKeyAuth(validator APIKeyValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("x-api-key"))
		if key == "" {
			auth := strings.TrimSpace(c.GetHeader("authorization"))
			if strings.HasPrefix(strings.ToLower(auth), "apikey ") {
				key = strings.TrimSpace(auth[6:])
			}
		}

		if !validator.IsValid(key) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Next()
	}
}
