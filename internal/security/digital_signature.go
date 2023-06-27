package security

import (
	"crypto/hmac"
	"crypto/sha256"
)

func Hash(data string, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	sign := h.Sum(nil)

	return string(sign)
}
