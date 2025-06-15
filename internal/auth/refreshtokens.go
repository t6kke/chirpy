package auth

import (
	"strconv"
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand_int, _ := rand.Read(key)

	encoded_string := hex.EncodeToString([]byte(strconv.Itoa(rand_int)))

	return encoded_string, nil
}
