package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	secret        = "99fy8976rhn3uibksdbfu8ib28bus"
	refreshSecret = "ng34u5928fnsd90f230-457239u"
)

type JWT interface {
	GenerateAccessToken(userID int) (string, error)
	GenerateRefreshToken(userID int, ttl time.Duration) (string, error)
	ParseAccessToken(jwtToken string) (int, error)
}

type jwtService struct{}

func NewJWT() JWT {
	return &jwtService{}
}

func (s *jwtService) GenerateAccessToken(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":            userID,
		"exp":           time.Now().Add(time.Minute * 15).Unix(),
		"authorization": true,
	})

	return token.SignedString([]byte(secret))
}

func (s *jwtService) GenerateRefreshToken(userID int, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  userID,
		"exp": time.Now().Add(ttl).Unix(),
	})

	return token.SignedString([]byte(refreshSecret))
}

func (s *jwtService) ParseAccessToken(jwtToken string) (int, error) {
	return parseToken(jwtToken, secret)
}

func parseToken(jwtToken, secretKey string) (int, error) {
	token, err := jwt.Parse(jwtToken, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return 0, fmt.Errorf("token parse error: %w", err)
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, fmt.Errorf("user id not found or invalid")
	}

	return int(idFloat), nil
}
