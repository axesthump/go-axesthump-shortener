package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"go-axesthump-shortener/internal/app/generator"
	"log"
	"net/http"
)

// userKeyID type for store user id in context.
type userKeyID string

// UserIDKey key for store user id in context.
const UserIDKey userKeyID = "id"

// authService contains data for auth.
type authService struct {
	// idGenerator - service for generation unique id.
	idGenerator *generator.IDGenerator
	// secretKey - secret key for hash.
	secretKey []byte
}

// NewAuthService returns new authService
func NewAuthService(generator *generator.IDGenerator) *authService {
	as := &authService{
		idGenerator: generator,
		secretKey:   []byte("secret_key"),
	}
	return as
}

// Auth middleware for auth. Returns handler with user id in context.
func (a *authService) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		var userID uint32
		var ok bool
		if err != nil {
			userID = a.generateCookie(w)
		} else {
			if ok, userID = a.validateCookie(cookie); !ok {
				userID = a.generateCookie(w)
			}
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateCookie generates new cookie.
// Hash new user id with secret key, then concatenate user id and hash and convert in to hex.
func (a *authService) generateCookie(w http.ResponseWriter) uint32 {
	newUserID := a.idGenerator.GetID()
	log.Printf("Generate new user id - %d\n", newUserID)
	newUserIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(newUserIDBytes, uint32(newUserID))

	h := hmac.New(sha256.New, a.secretKey)
	h.Write(newUserIDBytes)
	hash := h.Sum(nil)
	res := append(newUserIDBytes, hash...)
	token := hex.EncodeToString(res)
	newCookie := &http.Cookie{
		Name:  "auth",
		Value: token,
	}
	http.SetCookie(w, newCookie)
	return uint32(newUserID)
}

// validateCookie validates cookie.
func (a *authService) validateCookie(cookie *http.Cookie) (bool, uint32) {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return false, 0
	}
	userID := binary.BigEndian.Uint32(data[:4])
	h := hmac.New(sha256.New, a.secretKey)
	log.Printf("User id - %d\n", userID)
	h.Write(data[:4])
	hash := h.Sum(nil)
	if !hmac.Equal(hash, data[4:]) {
		return false, 0
	}
	if !a.idGenerator.IsCreatedID(userID) {
		return false, 0
	}
	return true, userID
}
