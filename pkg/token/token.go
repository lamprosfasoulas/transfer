package token
// ------------------------------
// NOT USED 
// ------------------------------
//
//package token
//
//import (
//	"fmt"
//	"time"
//
//	jwt "github.com/golang-jwt/jwt/v5"
//	"github.com/lamprosfasoulas/transfer/pkg/start"
//)
//
////JWT Utilities
//func GenerateJWT(username string) (string, error) {
//	claims := jwt.MapClaims{
//		"user": username,
//		"exp": time.Now().Add(start.Cfg.JWTExpiry).Unix(),
//		"iat": time.Now().Unix(),
//	}
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	return token.SignedString([]byte(start.Cfg.JWTSecret))
//}
//
//func ParseJWT(tokenStr string) (string, error) {
//	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
//		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
//		}
//		return []byte(start.Cfg.JWTSecret), nil
//	})
//	if err != nil {
//		return "", err
//	}
//	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
//		if user, ok := claims["user"].(string); ok {
//			return user, nil
//		}
//		return "", fmt.Errorf("JWT does not contain user claim")
//	}
//	return "", fmt.Errorf("JWT is invalid")
//}
