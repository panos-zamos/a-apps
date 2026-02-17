package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/panos/a-apps/shared/models"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// Config holds authentication configuration
type Config struct {
	JWTSecret string
	Users     []models.UserFromConfig
}

// LoadUsersFromConfig loads users from a YAML config file
func LoadUsersFromConfig(path string) ([]models.UserFromConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config struct {
		Users []models.UserFromConfig `yaml:"users"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config.Users, nil
}

// ValidateCredentials checks if username and password are valid
func ValidateCredentials(username, password string, users []models.UserFromConfig) (bool, error) {
	for _, user := range users {
		if user.Username == username {
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err == nil {
				return true, nil
			}
			return false, fmt.Errorf("invalid password")
		}
	}
	return false, fmt.Errorf("user not found")
}

// GenerateToken creates a JWT token for the given username
func GenerateToken(username, secret string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken validates a JWT token and returns the username
func ValidateToken(tokenString, secret string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return "", fmt.Errorf("invalid token claims")
		}
		return username, nil
	}

	return "", fmt.Errorf("invalid token")
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
