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

	if len(id) != 9 {
		t.Errorf("Random ID for new game was not the correct length: %s", id)
	}

	if store.isEmpty() {
		t.Error("Store should not be empty after add")
	}
}
