package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	key := []byte("secret key")

	msg := []byte("correct massage")
	h := hmac.New(sha256.New, key)
	h.Write(msg)
	correctHash := h.Sum(nil)

	tests := []struct {
		name    string
		msg     []byte
		isEqual bool
	}{
		{
			name:    "correct massage",
			msg:     []byte("correct massage"),
			isEqual: true,
		},
		{
			name:    "invalid massage",
			msg:     []byte("invalid massage"),
			isEqual: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			computedHash, err := hex.DecodeString(Hash(string(tt.msg), key))
			require.NoError(t, err)
			require.Equal(t, hmac.Equal(computedHash, correctHash), tt.isEqual)
		})
	}
}
