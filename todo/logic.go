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

type todoItem struct {
	id          int
	todoListID  int
	description string
	completed   bool
	createdAt   time.Time
}

type todoItemSaver = func(todoListID int, description string) (int, time.Time, error)
type todoItemSelecter = func(todoListID int) ([]todoItem, error)
type todoItemsCompletedUpdater = func(toBeCompleted []int, toBeActive []int) error
type todoItemsCompletedDeleter = func(todoListID int) error

func createTodoItem(
	todoListID, accountID int,
	description string,
	getter todoListGetter,
	saver todoItemSaver,
) (todoItem, error) {
	if len(description) < 4 || len(description) > 100 {
		return todoItem{}, errInvalidInput
	}

	todo, err := getter(todoListID)
	if err != nil {
		return todoItem{}, err
	}

	if todo.accountID != accountID {
		return todoItem{}, errPermissionDenied
	}

	id, t, err := saver(todoListID, description)

	return todoItem{
		id:          id,
		todoListID:  todoListID,
		description: description,
		completed:   false,
		createdAt:   t,
	}, err
}

func selectTodoItems(
	todoListID, accountID int,
	getter todoListGetter,
	selecter todoItemSelecter,
) ([]todoItem, error) {
	result := make([]todoItem, 0)

	todo, err := getter(todoListID)
	if err != nil {
		return result, err
	}

	if todo.accountID != accountID {
		return result, errPermissionDenied
	}

	return selecter(todoListID)
}

func todoItemsContain(items []todoItem, ids []int) bool {
	m := make(map[int]struct{})
	for _, item := range items {
		m[item.id] = struct{}{}
	}

	for _, id := range ids {
		if _, exists := m[id]; !exists {
			return false
		}
	}
	return true
}

func updateTodoItemsCompleted(
	todoListID, accountID int,
	toBeCompleted, toBeActive []int,
	getter todoListGetter,
	selecter todoItemSelecter,
	updater todoItemsCompletedUpdater,
) error {
	todo, err := getter(todoListID)
	if err != nil {
		return err
	}

	if todo.accountID != accountID {
		return errPermissionDenied
	}

	items, err := selecter(todoListID)
	if err != nil {
		return err
	}
	if !todoItemsContain(items, toBeCompleted) ||
		!todoItemsContain(items, toBeActive) {
		return errPermissionDenied
	}

	return updater(toBeCompleted, toBeActive)
}

func deleteTodoItemsCompleted(
	todoListID, accountID int,
	getter todoListGetter,
	deleter todoItemsCompletedDeleter,
) error {
	todo, err := getter(todoListID)
	if err != nil {
		return err
	}

	if todo.accountID != accountID {
		return errPermissionDenied
	}

	return deleter(todoListID)
}
