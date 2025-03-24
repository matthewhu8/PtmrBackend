package firebase

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "Authorization"
	AuthorizationTypeBearer = "Bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

var ErrUnauthorized = errors.New("user not authorized to access this resource")

func AddAuthorization(t *testing.T, request *http.Request, authorizationType string, role string) {
	mockToken := fmt.Sprintf(`{"uid": "mock-uid", "role": "%s"}`, role)
	encodedToken := base64.StdEncoding.EncodeToString([]byte(mockToken))
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, encodedToken)
	request.Header.Set("Authorization", authorizationHeader)
}

func AuthMiddleware(client AuthClientFirebase, requiredRole string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(AuthorizationHeaderKey)
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			ctx.Abort()
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := client.VerifyIDToken(ctx.Request.Context(), idToken)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid ID token"})
			ctx.Abort()
			return
		}

		claims := token.Claims
		role, ok := claims["role"].(string)
		if !ok || role != requiredRole {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			ctx.Abort()
			return
		}

		ctx.Set(AuthorizationPayloadKey, claims)
		ctx.Next()
	}
}

func GetUIDFromClaims(ctx *gin.Context) (string, error) {
	claims, exists := ctx.Get(AuthorizationPayloadKey)
	if !exists {
		return "", errors.New("authorization payload missing")
	}
	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		return "", errors.New("invalid claims format")
	}
	uid, ok := claimsMap["uid"].(string)
	if !ok {
		return "", errors.New("uid not found in claims")
	}
	return uid, nil
}
