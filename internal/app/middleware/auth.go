package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"go-axesthump-shortener/internal/app/user"
	"log"
	"net/http"
)

type authService struct {
	idGenerator user.IDGenerator
	secretKey   []byte
}

func NewAuthService() *authService {
	as := &authService{
		idGenerator: user.IDGenerator{},
		secretKey:   make([]byte, 0, 16),
	}
	_, err := rand.Read(as.secretKey)
	if err != nil {
		panic(err)
	}
	return as
}

func (a *authService) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			a.GenerateCookie(w)
		} else {
			if !a.validateCookie(cookie) {
				a.GenerateCookie(w)
			}
		}
		next.ServeHTTP(w, r)
		return
	})
}

func (a *authService) GenerateCookie(w http.ResponseWriter) {
	newUserID := a.idGenerator.GetNewUserId()
	log.Printf("Generate new user id - %d\n", newUserID)
	newUserIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(newUserIDBytes, newUserID)

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
}

func (a *authService) validateCookie(cookie *http.Cookie) bool {
	data, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return false
	}
	userID := binary.BigEndian.Uint32(data[:4])
	h := hmac.New(sha256.New, a.secretKey)
	log.Printf("User id - %d\n", userID)
	h.Write(data[:4])
	hash := h.Sum(nil)
	if !hmac.Equal(hash, data[4:]) {
		return false
	}
	if !a.idGenerator.IsCreatedUser(userID) {
		return false
	}
	return true
}
