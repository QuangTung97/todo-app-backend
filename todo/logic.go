package todo

import (
	"errors"
	"regexp"

	"github.com/golang/glog"
	"golang.org/x/crypto/bcrypt"
)

var errAlreadyExisted error = errors.New("already existed")
var errInvalidInput error = errors.New("invalid input")

// error can be errAlreadyExisted
// passwordHash should use bcrypt
type accountSaver = func(username, passwordHash string) error

func validateUsername(username string) bool {
	if len(username) < 5 {
		return false
	} else if len(username) > 30 {
		return false
	}

	matched, err := regexp.MatchString("^[a-zA-Z][a-zA-Z0-9]+$", username)
	if err != nil {
		glog.Fatal(err)
	}

	return matched
}

func validatePassword(password string) bool {
	if len(password) < 5 {
		return false
	}
	return true
}

func createAccount(saver accountSaver, username, password string) error {
	if validateUsername(username) && validatePassword(password) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			glog.Fatal(err)
		}
		return saver(username, string(hash))
	}
	return errInvalidInput
}
