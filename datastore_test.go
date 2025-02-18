package main

import (
	"context"
	"testing"
)

func TestRandomId(t *testing.T) {
	id := randomId()

	if len(id) != 9 {
		t.Errorf("Random ID was not the correct length: %s", id)
	}
}

func TestAddGame(t *testing.T) {
	store := GameStore{datastore: testDataStore()}
	var game Game
	id := store.addGame(context.Background(), game)

	if len(id) != RANDOM_ID_LENGTH {
		t.Errorf("Random ID for new game was not the correct length: %s", id)
	}

	if store.isEmpty() {
		t.Error("Store should not be empty after add")
	}
}

func TestMinRandomId(t *testing.T) {
	if RANDOM_1_BASE+RANDOM_1_MAX != RANDOM_2_MAX {
		t.Error("Oops, random ID assumptions are not valid")
	}
}

func TestManyRandomIds(t *testing.T) {
	for i := 0; i < 100000; i++ {
		code := randomId()
		if len(code) != RANDOM_ID_LENGTH {
			t.Errorf("Random code is not correct length: %s", code)
		}
		if code[0] == '0' {
			t.Errorf("Random code starts with 0: %s", code)
		}
		if code[4] != '-' {
			t.Errorf("Random code starts doesn't have correct dash: %s", code)
		}
	}
}
