package helpers

import (
"crypto"
"github.com/cockroachdb/errors"
"os"
"strconv"
)

func HashPassword(password string, username string) ([]byte, error) {
	hasher := crypto.SHA256.New()

	passwordSalt, ok := os.LookupEnv("PASSWORD_SALT")
	if !ok {
		return []byte{}, errors.New("could not hash password")
	}

	for i := range 13 + len(passwordSalt) + len(username) {
		hasher.Write([]byte(strconv.Itoa(i)))
		hasher.Write([]byte(password))
		hasher.Write([]byte(username))
		hasher.Write([]byte(passwordSalt))
	}

	return hasher.Sum(nil), nil
}