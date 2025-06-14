package auth

import (
	"fmt"
	"time"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "chirpy"
)

func HashPassword(password string) (string, error) {
	byte_pass_hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(byte_pass_hash), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		//return errors.New("Passowrd not matching")
		return fmt.Errorf("Passowrd not matching %v", err)
	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    string(TokenTypeAccess),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	result_string, err := token.SignedString([]byte(tokenSecret))
	fmt.Println(result_string, err)

	return result_string, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claimsStruct, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}
	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	user_id, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("token did not have user ID: %w", err)
	}

	return_uuid, err := uuid.Parse(user_id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return return_uuid, nil
}
