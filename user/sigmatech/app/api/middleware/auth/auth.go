package auth

import (
	"errors"
	"net/http"
	"strings"
	"user/sigmatech/app/api/middleware/jwt"
	"user/sigmatech/app/constants"
	"user/sigmatech/app/controller"

	"github.com/gin-gonic/gin"
)

// authentication is a middleware that verify JWT token headers
func UserAuthentication(jwt jwt.IJwtService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := getHeaderToken(ctx)
		if err != nil {
			controller.RespondWithError(ctx, http.StatusUnauthorized, constants.UNAUTHORIZED_ACCESS, err)
			return
		}

		claims, valid := jwt.VerifyUserToken(ctx, token)
		if !valid {
			controller.RespondWithError(ctx, http.StatusUnauthorized, constants.UNAUTHORIZED_ACCESS, err)
			return
		}
		ctx.Set(constants.CTK_CLAIM_KEY.String(), claims)
		ctx.Next()
	}
}

// Authentication is a middleware that verifies JWT token headers for both user and customer
func Authentication(jwt jwt.IJwtService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ""
		t, err := getHeaderToken(ctx)
		if err != nil {
			controller.RespondWithError(ctx, http.StatusUnauthorized, constants.UNAUTHORIZED_ACCESS, err)
			return
		}

		if t == "" {
			controller.RespondWithError(ctx, http.StatusUnauthorized, constants.UNAUTHORIZED_ACCESS, errors.New("x-api-key or token is required"))
			return
		}

		token = t

		claims, valid := jwt.VerifyToken(ctx, token)
		if !valid {
			controller.RespondWithError(ctx, http.StatusUnauthorized, constants.UNAUTHORIZED_ACCESS, errors.New("invalid token"))
			return
		}

		ctx.Set(constants.CTK_CLAIM_KEY.String(), claims)

		ctx.Next()
	}
}

func getHeaderToken(ctx *gin.Context) (string, error) {
	header := string(ctx.GetHeader(constants.AUTHORIZATION))
	return extractToken(header)
}

func extractToken(header string) (string, error) {
	if strings.HasPrefix(header, constants.BEARER) {
		return header[len(constants.BEARER):], nil
	}
	return "", errors.New("token not found")
}
