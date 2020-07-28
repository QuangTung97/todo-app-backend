package todo

import "testing"

func TestValidateUsername(t *testing.T) {
	ok := validateUsername("")
	if ok {
		t.Error("empty string should be false")
	}

	ok = validateUsername("tung")
	if ok {
		t.Error("len should not less than 5")
	}

	ok = validateUsername("abcdefdfkfjkafjdkafdftttttfddddd")
	if ok {
		t.Error("len should not bigger than 30")
	}

	ok = validateUsername("1234567")
	if ok {
		t.Error("should not start by a number")
	}

	ok = validateUsername("ab defg")
	if ok {
		t.Error("should not contains spaces")
	}

	ok = validateUsername("abdd^&efg")
	if ok {
		t.Error("should not contains special characters")
	}
}

func TestCreateAccount(t *testing.T) {
	saveCount := 0
	hash := ""
	saver := func(u, h string) error {
		saveCount++
		hash = h
		return nil
	}

	err := createAccount(saver, "tungquang", "abfd")
	if err != errInvalidInput || saveCount > 0 {
		t.Errorf("should be invalid input, actual: %s", err)
	}

	err = createAccount(saver, "tungquang", "abcde")
	if err != nil || saveCount != 1 || len(hash) != 60 {
		t.Errorf("should be called, len(hash) == 60, actual: %v", len(hash))
	}
}

func TestTodoItemsContain(t *testing.T) {
	items := []todoItem{
		{id: 1},
		{id: 2},
		{id: 3},
	}

	ids := []int{1, 2}
	result := todoItemsContain(items, ids)
	if result != true {
		t.Errorf("should be true")
	}

	ids = []int{1, 2, 4}
	result = todoItemsContain(items, ids)
	if result != false {
		t.Errorf("should be false")
	}
}
