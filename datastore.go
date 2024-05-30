package main

import (
	"context"
	"fmt"
	"math/rand"
)

type DataStore interface {
	summary() string
	open()
	close()
	getGame(ctx context.Context, id string) Game
	putGame(ctx context.Context, id string, game Game)
	addGame(ctx context.Context, game *Game) string
	isEmpty() bool
}

type TestDataStore struct {
	games map[string]Game
}

func testDataStore() DataStore {
	logs.info("Setting up in-memory test datastore")
	store := new(TestDataStore)
	store.games = make(map[string]Game)
	return store
}

func (store *TestDataStore) summary() string {
	return "TestDataStore"
}

func (store *TestDataStore) getGame(ctx context.Context, id string) Game {
	return store.games[id]
}

func (store *TestDataStore) putGame(ctx context.Context, id string, game Game) {
	logs.debug1(ctx, "Writing game %s, %s", id, game.Title)
	store.games[id] = game
}

func (store *TestDataStore) addGame(ctx context.Context, game *Game) string {
	id := randomId()
	// keep generating IDs until we get an unused one
	for ok := true; ok; ok = store.contains(id) {
		id = randomId()
	}
	game.ID = id
	store.games[id] = *game
	return id
}

func (store *TestDataStore) open()  {}
func (store *TestDataStore) close() {}
func (store *TestDataStore) isEmpty() bool {
	return len(store.games) == 0
}

func (store *TestDataStore) contains(id string) bool {
	_, ok := store.games[id]
	return ok
}

func randomId() string {
	return fmt.Sprintf("%04X-%04X", rand.Intn(0xFFFF), rand.Intn(0xFFFF))
}

const TEST_ID_1 = "CODE1"
const TEST_ID_2 = "CODE2"

func testGame1() Game {
	game1 := Game{
		ID:       TEST_ID_1,
		Title:    "Test Game",
		Period:   1,
		HomeTeam: "Reds",
		AwayTeam: "Blues",
		GameDate: "2024-05-27",
	}
	AddPenalty(&game1, 2, "14:25", AWAY, 50, 2, "Slash")
	AddGoal(&game1, 1, "18:30", HOME, 41, 89, 93, "Even")
	AddPenalty(&game1, 2, "3:45", HOME, 41, 2, "Trip")
	AddGoal(&game1, 3, "18:30", AWAY, 98, 0, 0, "PP")
	return game1
}

func testGame2() Game {
	game2 := Game{
		ID:         TEST_ID_2,
		Title:      "Locked Game",
		Period:     1,
		HomeTeam:   "Greens",
		AwayTeam:   "Greys",
		GameDate:   "2024-05-27",
		LockedWith: "secret123",
	}
	AddGoal(&game2, 1, "18:30", HOME, 41, 89, 93, "Even")

	return game2
}

func addTestGames(store DataStore) {
	logs.info("Adding test games to %s", store.summary())

	game1 := testGame1()
	store.putGame(context.Background(), game1.ID, game1)

	game2 := testGame2()
	store.putGame(context.Background(), game2.ID, game2)

	logs.info("Test games added")
}
