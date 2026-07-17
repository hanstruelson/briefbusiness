package utilities

import (
	"appeals/db"
	"crypto/rand"
	"encoding/hex"
)

func GenerateAccessToken(userId int) string {
	b := make([]byte, 32)
	// rand.Read fills the slice with cryptographically secure random bytes
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	// Convert bytes to a hexadecimal string
	token := hex.EncodeToString(b)
	_, err := db.DB.Exec("INSERT INTO AccessToken (Token, UserId) VALUES (?, ?)", token, userId)
	if err != nil {
		panic(err)
	}
	return token
}
