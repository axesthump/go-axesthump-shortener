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

type userKeyID string

const UserIDKey userKeyID = "id"

type authService struct {
	idGenerator *generator.IDGenerator
	secretKey   []byte
}

func NewAuthService(generator *generator.IDGenerator) *authService {
	as := &authService{
		idGenerator: generator,
		secretKey:   []byte("secret_key"),
	}
	return as
}

func (a *authService) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		var userID uint32
		var ok bool
		if err != nil {
			userID = a.GenerateCookie(w)
		} else {
			if ok, userID = a.validateCookie(cookie); !ok {
				userID = a.GenerateCookie(w)
			}
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *authService) GenerateCookie(w http.ResponseWriter) uint32 {
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
	return true, userID
}
