package core

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/seip25/Go-Blue-bird/config"
)

func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func IsDev() bool {
	return config.IsDev()
}

func IsProd() bool {
	return config.IsProd()
}
