package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

// GenerateUserID creates a new random uint32 as the user ID.
func GenerateUserID() (uint32, error) {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return 0, err
	}
	// Interpret the 4 bytes as a big-endian uint32
	return binary.BigEndian.Uint32(bytes), nil
}

// SignUserID signs userID, and returns a signed value.
func SignUserID(userID uint32, secretKey []byte) []byte {
	// Convert the uint32 ID to a byte slice (4 bytes, big-endian)
	userIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(userIDBytes, userID)

	h := hmac.New(sha256.New, secretKey)
	h.Write(userIDBytes)

	return h.Sum(nil)
}

// VerifyUserID verifies the signed userID value.
func VerifyUserID(userID uint32, signature, secretKey []byte) bool {
	// Convert the uint32 ID to a byte slice (4 bytes, big-endian)
	userIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(userIDBytes, userID)

	expectedSig := hmac.New(sha256.New, secretKey)
	expectedSig.Write(userIDBytes)
	expectedSignature := expectedSig.Sum(nil)

	return hmac.Equal(signature, expectedSignature)
}

// EncodeUserIDCookie combines userID and signature into a single hex-encoded string
func EncodeUserIDCookie(userID uint32, signature []byte) string {
	buf := make([]byte, 4+len(signature))
	binary.BigEndian.PutUint32(buf[:4], userID)
	copy(buf[4:], signature)
	return hex.EncodeToString(buf)
}

// DecodeUserIDCookie extract userID and signature from a hex-encoded string
func DecodeUserIDCookie(cookieValue string) (uint32, []byte, error) {
	buf, err := hex.DecodeString(cookieValue)
	if err != nil {
		return 0, nil, err
	}
	userID := binary.BigEndian.Uint32(buf[:4])
	signature := make([]byte, len(buf)-4)
	copy(signature, buf[4:])
	return userID, signature, nil
}
