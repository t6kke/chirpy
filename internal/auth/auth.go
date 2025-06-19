package auth

import (
	"fmt"
	"time"
	"errors"
	"strings"
	"net/http"

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

func GetBearerToken(headers http.Header) (string, error) {
	auth_token := headers.Get("Authorization")
	if auth_token == "" {
		return "", errors.New("No \"Authorization\" in header")
	}

	token_split := strings.Split(auth_token, " ")
	if len(token_split) != 2 {
		return "", errors.New("Invalid token string format in \"Authorization\"")
	}

	//TODO should also validate that it's Bearer token

	return token_split[1], nil
}

func GetAPIKey(headers http.Header) (string, error) {
	apikey_from_header := headers.Get("Authorization")
	if apikey_from_header == "" {
		return "", errors.New("No \"Authorization\" in header")
	}

	apikey := strings.Split(apikey_from_header, " ")
	if len(apikey) != 2 {
		return "", errors.New("Invalid token string format in \"Authorization\"")
	}

	//TODO should also validate that it's ApiKey token

	return apikey[1], nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Issuer:    string(TokenTypeAccess),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	result_string, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

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
