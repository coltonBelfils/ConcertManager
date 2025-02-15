package cookieHelpers

import (
	"crypto"
	"encoding/base64"
	"github.com/cockroachdb/errors"
	"net/http"
	"os"
	"strconv"
)

func ReadCookies(r *http.Request) (string, []byte, []byte, error) {
	username64, err := r.Cookie("username")
	if err != nil {
		return "", nil, nil, errors.Wrap(err, "error reading username cookie")
	}

	username, decodeErr := base64.StdEncoding.DecodeString(username64.Value)
	if decodeErr != nil {
		return "", nil, nil, errors.Wrap(decodeErr, "error decoding username")
	}

	passwordHash64, err := r.Cookie("passwordHash")
	if err != nil {
		return "", nil, nil, errors.Wrap(err, "error reading passwordHash cookie")
	}

	passwordHash, decodeErr := base64.StdEncoding.DecodeString(passwordHash64.Value)
	if decodeErr != nil {
		return "", nil, nil, errors.Wrap(decodeErr, "error decoding passwordHash")
	}

	check, err := r.Cookie("check")
	if err != nil {
		return "", nil, nil, errors.Wrap(err, "error reading check cookie")
	}

	checkDecode, checkDecodeErr := base64.StdEncoding.DecodeString(check.Value)
	if checkDecodeErr != nil {
		return "", nil, nil, errors.Wrap(checkDecodeErr, "error decoding check")
	}

	return string(username), passwordHash, checkDecode, nil
}

func WriteCookies(w http.ResponseWriter, username string, passwordHash []byte, check []byte) error {
	username64 := base64.StdEncoding.EncodeToString([]byte(username))
	passwordHash64 := base64.StdEncoding.EncodeToString(passwordHash)
	check64 := base64.StdEncoding.EncodeToString(check)

	http.SetCookie(w, &http.Cookie{
		Name:  "username",
		Value: username64,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "passwordHash",
		Value: passwordHash64,
	})
	http.SetCookie(w, &http.Cookie{
		Name:  "check",
		Value: check64,
	})

	return nil
}

func CookieCheck(username string, passwordHash []byte) ([]byte, error) {
	hasher := crypto.SHA256.New()

	cookieCheckSalt, ok := os.LookupEnv("COOKIE_CHECK_SALT")
	if !ok {
		return nil, errors.New("could not validate login")
	}

	for i := range 11 + len(cookieCheckSalt) + len(username) {
		hasher.Write([]byte(strconv.Itoa(i)))
		hasher.Write([]byte(username))
		hasher.Write(passwordHash)
		hasher.Write([]byte(cookieCheckSalt))
	}

	return hasher.Sum(nil), nil
}