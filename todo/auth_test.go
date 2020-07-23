package todo

import (
	"testing"
	"time"
)

func TestCheckPasswordWithHash(t *testing.T) {
	ok := checkPasswordWithHash(
		"tung", "$2a$10$CTerPFQ.ECHY5gwlgBHM9ezxlLrt5VEPR5mkZVNG9OFzg2dIWbMu6")
	if ok {
		t.Error("password should not match")
	}

	ok = checkPasswordWithHash(
		"admin123", "$2a$10$CTerPFQ.ECHY5gwlgBHM9ezxlLrt5VEPR5mkZVNG9OFzg2dIWbMu6")
	if !ok {
		t.Error("password should match")
	}
}

func TestParseAccountID(t *testing.T) {
	token := "123:somerandom"
	id, ok := parseAccountID(token)
	if !(ok && id == 123) {
		t.Errorf("wrong answer, token %q, values: %d %v", token, id, ok)
	}

	id, ok = parseAccountID("")
	if ok {
		t.Error("should be an error")
	}

	id, ok = parseAccountID("abc:1233")
	if ok {
		t.Error("should be an error")
	}
}

type mockCallbacks struct {
	setValueCount     int
	setValue          valueSetter
	setValueCalledKey string

	getValueCount int
	getValue      valueGetter

	getAccountCount int
	getAccount      accountGetter
	accountID       int
	passwordHash    string

	setExpireCount int
	setExpire      expirationSetter
}

func newMockCallbacks() *mockCallbacks {
	mock := &mockCallbacks{
		setValueCount:   0,
		getValueCount:   0,
		getAccountCount: 0,
		setExpireCount:  0,
	}

	mock.setValue = func(key, value string, d time.Duration) error {
		mock.setValueCount++
		mock.setValueCalledKey = key
		return nil
	}

	mock.getValue = func(key string) (string, error) {
		mock.getValueCount++
		return "", nil
	}

	mock.getAccount = func(username string) (int, string, error) {
		mock.getAccountCount++
		return mock.accountID, mock.passwordHash, nil
	}

	mock.setExpire = func(key string, d time.Duration) error {
		mock.setExpireCount++
		return nil
	}

	return mock
}

func TestVerifyWithBasicAuth(t *testing.T) {
	basicAuth := basicAuthInfo{
		username: "quangtung",
		password: "admin123",
		ok:       true,
	}

	mock := newMockCallbacks()
	mock.accountID = 2334
	mock.passwordHash = "$2a$10$CTerPFQ.ECHY5gwlgBHM9ezxlLrt5VEPR5mkZVNG9OFzg2dIWbMu6"

	id, _, ok, err := verifyCredentials(
		basicAuth, "",
		mock.setValue, mock.getValue,
		mock.setExpire, mock.getAccount,
	)
	if !ok || err != nil || id != 2334 {
		t.Error("error:", ok, err, id)
	}
	if !(mock.getAccountCount == 1 && mock.setValueCount == 1 &&
		mock.getValueCount == 0 && mock.setExpireCount == 0) {
		t.Error("not called correctly")
	}
	if id, ok := parseAccountID(mock.setValueCalledKey); !ok || id != 2334 {
		t.Errorf("actual accountID: %d", id)
	}

	basicAuth.password = "tung222"
	mock = newMockCallbacks()
	mock.passwordHash = "$2a$10$CTerPFQ.ECHY5gwlgBHM9ezxlLrt5VEPR5mkZVNG9OFzg2dIWbMu6"

	_, _, ok, err = verifyCredentials(
		basicAuth, "",
		mock.setValue, mock.getValue,
		mock.setExpire, mock.getAccount,
	)
	if ok || err != nil {
		t.Error("should unauthenticated and not have error")
	}
	if !(mock.getAccountCount == 1 && mock.setValueCount == 0 &&
		mock.getValueCount == 0 && mock.setExpireCount == 0) {
		t.Error("not called correctly")
	}
}
