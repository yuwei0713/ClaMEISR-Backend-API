package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// 定义JWT签名密钥
var jwtKey = []byte("my_secret_key")

// 定义Claims结构体，包含了我们想要在JWT中存储的信息
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// 生成JWT
func GenerateJWT(username string) (string, error) {
	// 设置JWT的过期时间为24小时
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// 使用签名方法HS256创建一个新的Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名Token并返回Token字符串
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// 验证JWT
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// 解析Token，并将结果存储在claims中
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}
