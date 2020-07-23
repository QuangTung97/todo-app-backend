package todo

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenSecretSize = 20
)

var tokenExpiration = 60 * time.Minute

var errKeyNotExist = errors.New("key does not exist")
var errAccountNotExist = errors.New("account does not exist")

type valueSetter = func(key, value string, expiration time.Duration) error

// return errKeyNotExist if key not exist
type valueGetter = func(key string) (string, error)

type expirationSetter = func(key string, expiration time.Duration) error

// return errAccountNotExist if no account has this username
type accountGetter = func(username string) (int, string, error)

func checkPasswordWithHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false
	}
	if err != nil {
		glog.Fatal(err)
	}
	return true
}

// token has a form "1234:somesecret"
func parseAccountID(token string) (int, bool) {
	index := strings.Index(token, ":")
	if index == -1 {
		return 0, false
	}
	numStr := token[0:index]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, false
	}

	return num, true
}

type basicAuthInfo struct {
	username string
	password string
	ok       bool
}

func verifyCredentials(
	basicAuth basicAuthInfo,
	token string,
	setValue valueSetter,
	getValue valueGetter,
	setExpire expirationSetter,
	getAccount accountGetter,
) (int, string, bool, error) {
	if basicAuth.ok {
		id, hash, err := getAccount(basicAuth.username)
		if err == errAccountNotExist {
			return 0, "", false, nil
		}
		if err != nil {
			return 0, "", false, err
		}
		ok := checkPasswordWithHash(basicAuth.password, hash)
		if !ok {
			return 0, "", false, nil
		}

		b := make([]byte, tokenSecretSize)
		_, err = rand.Read(b)
		if err != nil {
			return 0, "", false, err
		}
		randStr := base64.StdEncoding.EncodeToString(b)

		token = fmt.Sprintf("%d:%s", id, randStr)
		err = setValue(token, "ok", tokenExpiration)
		if err != nil {
			return 0, "", false, err
		}
		return id, token, true, nil
	}

	value, err := getValue(token)
	if err == errKeyNotExist {
		return 0, "", false, nil
	}
	if err != nil {
		return 0, "", false, err
	}
	if value != "ok" {
		return 0, "", false, nil
	}

	err = setExpire(token, tokenExpiration)
	if err != nil {
		return 0, "", false, err
	}

	id, ok := parseAccountID(token)
	if !ok {
		glog.Fatal("error while parsing account id")
	}

	return id, token, true, nil
}
