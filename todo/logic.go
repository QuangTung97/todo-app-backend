package todo

import (
	"errors"
	"regexp"
	"time"

	"github.com/golang/glog"
	"golang.org/x/crypto/bcrypt"
)

var errAlreadyExisted error = errors.New("already existed")
var errInvalidInput error = errors.New("invalid input")
var errPermissionDenied error = errors.New("permission denied")

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

// Todo List
type todoList struct {
	id        int
	accountID int
	name      string
	createdAt time.Time
	updatedAt time.Time
}

type todoListSaver = func(accountID int, name string) (int, time.Time, error)
type todoListGetter = func(id int) (todoList, error)
type todoListUpdater = func(id int, name string) (time.Time, error)

func validateTodoListName(name string) bool {
	if len(name) < 5 || len(name) > 30 {
		return false
	}
	return true
}

func createTodoList(
	accountID int, name string,
	saver todoListSaver,
) (todoList, error) {
	if validateTodoListName(name) {
		id, createdAt, err := saver(accountID, name)
		return todoList{
			id:        id,
			name:      name,
			accountID: accountID,
			createdAt: createdAt,
			updatedAt: createdAt,
		}, err
	}
	return todoList{}, errInvalidInput
}

func updateTodoList(
	id, accountID int, name string,
	getter todoListGetter,
	updater todoListUpdater,
) (todoList, error) {
	todo, err := getter(id)
	if err != nil {
		return todo, err
	}

	if todo.accountID != accountID {
		return todo, errPermissionDenied
	}

	updatedAt, err := updater(id, name)
	todo.name = name
	todo.updatedAt = updatedAt
	return todo, err
}

type todoListsByAccountGetter = func(accountID int) ([]todoList, error)

type todoListDeleter = func(id int) error

func deleteTodoList(
	id, accountID int,
	getter todoListGetter,
	deleter todoListDeleter,
) error {
	todo, err := getter(id)
	if err != nil {
		return err
	}

	if todo.accountID != accountID {
		return errPermissionDenied
	}

	return deleter(id)
}
