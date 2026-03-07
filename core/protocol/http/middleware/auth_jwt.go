package middleware

import (
	"github.com/IZumLzk/httpf/core/common"
	jwt "github.com/golang-jwt/jwt/v4"
	"strings"
)

type AuthJWTMiddleware struct {
	secret string
}

// NewAuthJWTMW 创建JWT认证中间件
func NewAuthJWTMW() common.HTTPMiddlewareFunc {
	// 从配置中获取JWT密钥，为了演示，这里使用默认值
	secret := "gospacex-default-jwt-secret"

	return (&AuthJWTMiddleware{secret: secret}).Process
}

func (am *AuthJWTMiddleware) Process(ctx common.HTTPContext) bool {
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		ctx.JSON(401, map[string]interface{}{
			"error":   "unauthorized",
			"message": "Missing or invalid authorization header",
		})
		return false
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(am.secret), nil
	})

	if err != nil || !token.Valid {
		ctx.JSON(401, map[string]interface{}{
			"error":   "invalid token",
			"message": "Token is invalid or expired",
		})
		return false
	}

	// 验证成功，可选择将用户信息附加到上下文中
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 在真实的上下文中存储用户信息，比如传递给后续处理
		// 这里暂时不实际使用claims变量以避免未使用变量的错误
		_ = claims
	}

	return true
}
