package jwt

import (
	"encoding/json"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/senslabs/alpha/sens/errors"
	"github.com/senslabs/alpha/sens/logger"
)

const accessSigningKey = "L7zBtmHybEkjubZfvkAs-3gklypIZGO5WZKZZQuLQ"
const refreshSigningKey = "T8YmLUVa9BR66RbNR5YV-zLDM3cuQMz8gKismkDCw"

func generateToken(subject interface{}, duration time.Duration, signingMethod *jwt.SigningMethodHMAC, signingKey string) (string, error) {
	s, err := json.Marshal(subject)
	if err != nil {
		logger.Error(err)
		return "", errors.FromError(errors.GO_ERROR, err)
	}
	issuedAt := time.Now().UTC()
	claims := &jwt.StandardClaims{
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: issuedAt.Unix() + int64(duration.Seconds()),
		Issuer:    "senslabs.io",
		Subject:   string(s),
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString([]byte(signingKey))
}

func GenerateAccessToken(subject interface{}, expiry time.Duration) (string, error) {
	return generateToken(subject, expiry, jwt.SigningMethodHS256, accessSigningKey)
}

func GenerateRefreshToken(subject interface{}) (string, error) {
	return generateToken(subject, 90*24*time.Hour, jwt.SigningMethodHS512, refreshSigningKey)
}

func GenerateTemporaryToken(subject interface{}) (string, error) {
	return GenerateAccessToken(subject, 15*time.Minute)
}

func verifyToken(tokenText string, signingMethod *jwt.SigningMethodHMAC, signingKey string) (map[string]interface{}, error) {
	var m map[string]interface{}
	token, err := jwt.Parse(tokenText, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			logger.Errorf("Method: %v, Unexpected signing method: %v", method, token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		logger.Error("Error: ", err)
		return m, errors.New(errors.GO_ERROR, err.Error())
	}
	if token == nil || !token.Valid {
		logger.Error("Token: ", token)
		return m, errors.New(errors.GO_ERROR, "Invalid Token Received")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// err = types.JsonUnmarshal(claims["sub"].([]byte), &m)
		err = json.Unmarshal([]byte(claims["sub"].(string)), &m)
	}
	return m, err
}

func VerifyToken(token interface{}) (map[string]interface{}, error) {
	if token == nil {
		return nil, errors.New(errors.GO_ERROR, "No token found")
	}
	return VerifyTokenString(token.(string))
}

func VerifyTokenString(tokenText string) (map[string]interface{}, error) {
	return verifyToken(tokenText, jwt.SigningMethodHS256, accessSigningKey)
}

func RefreshAccessToken(tokenText string) (string, error) {
	if subject, err := verifyToken(tokenText, jwt.SigningMethodHS512, refreshSigningKey); err != nil {
		return "", err
	} else {
		return GenerateAccessToken(subject, 24*time.Hour)
	}
}
