package auth

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/ginqi7/bsimp/internal/config"
	"net/http"
)

type AuthLibrary struct {
	config *config.Config
}

func NewAuthLibrary(config *config.Config) *AuthLibrary {
	return &AuthLibrary{
		config: config,
	}
}

func Md5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func (al *AuthLibrary) CheckCookie(r *http.Request) error {
	secret_cookie, err := r.Cookie("password")
	if err != nil {
		return err
	}
	if !(Md5(al.config.S3.Credentials.Secret) == secret_cookie.Value) {
		return errors.New("Password Error!")
	}
	return nil

}
