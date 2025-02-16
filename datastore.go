package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
)

type GameStore struct {
	datastore DataStore
}

const GAMES_COLLECTION = "Games"

func (store GameStore) getGame(ctx context.Context, id string) Game {
	var game Game
	store.datastore.Get(ctx, GAMES_COLLECTION, id, &game)
	return game
}

func (store GameStore) putGame(ctx context.Context, id string, game Game) {
	store.datastore.Put(ctx, GAMES_COLLECTION, id, &game)
}

func (store GameStore) addGame(ctx context.Context, game Game) string {
	id := randomId()
	// keep generating IDs until we get an unused one
	// for ok := true; ok; ok = store.datastore.Contains(id) {
	// 	id = randomId()
	// }
	game.ID = id
	store.putGame(ctx, id, game)
	// store.games[id] = *game
	return id
}

func (store GameStore) summary() string {
	return store.datastore.summary()
}

func (store GameStore) isEmpty() bool {
	return store.datastore.isEmpty()
}

func (store GameStore) open() {
	store.datastore.open()
}

func (store GameStore) close() {
	store.datastore.close()
}

type DataStore interface {
	summary() string
	open()
	close()
	Get(ctx context.Context, collection string, id string, item interface{}) interface{}
	Put(ctx context.Context, collection string, id string, item interface{})
	// getGame(ctx context.Context, id string) Game
	// putGame(ctx context.Context, id string, game Game)
	// addGame(ctx context.Context, game *Game) string
	isEmpty() bool
}

type TestDataStore struct {
	items map[string](map[string][]byte)
}

func testDataStore() DataStore {
	logs.info("Setting up in-memory test datastore")
	store := new(TestDataStore)
	store.items = make(map[string]map[string][]byte)
	store.items[GAMES_COLLECTION] = make(map[string][]byte)
	return store
}

func (store *TestDataStore) summary() string {
	return "TestDataStore"
}

func (store *TestDataStore) Get(ctx context.Context, collection string, id string, item interface{}) interface{} {
	data := store.items[collection][id]

	_ = json.Unmarshal(data, item)
	return item
}

func (store *TestDataStore) Put(ctx context.Context, collection string, id string, item interface{}) {
	data, _ := json.Marshal(item)
	store.items[collection][id] = data
}

// func (store *TestDataStore) getGame(ctx context.Context, id string) Game {
// 	return store.games[id]
// }

// func (store *TestDataStore) putGame(ctx context.Context, id string, game Game) {
// 	logs.debug1(ctx, "Writing game %s, %s", id, game.Title)
// 	store.games[id] = game
// }

// func (store *TestDataStore) addGame(ctx context.Context, game *Game) string {
// 	id := randomId()
// 	// keep generating IDs until we get an unused one
// 	for ok := true; ok; ok = store.contains(id) {
// 		id = randomId()
// 	}
// 	game.ID = id
// 	store.games[id] = *game
// 	return id
// }

func (store *TestDataStore) open()  {}
func (store *TestDataStore) close() {}
func (store *TestDataStore) isEmpty() bool {
	return len(store.items[GAMES_COLLECTION]) == 0
}

// func (store *TestDataStore) contains(id string) bool {
// 	_, ok := store.games[id]
// 	return ok
// }

func randomId() string {
	return fmt.Sprintf("%04X-%04X", rand.Intn(0xFFFF), rand.Intn(0xFFFF))
}

const TEST_ID_1 = "CODE1"
const TEST_ID_2 = "CODE2"

func testGame1() Game {
	game1 := Game{
		ID:       TEST_ID_1,
		Title:    "Blues @ Reds, 27 May 2024",
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

func addTestGames(store GameStore) {
	logs.info("Adding test games to %s", store.summary())

	game1 := testGame1()
	store.putGame(context.Background(), game1.ID, game1)

	game2 := testGame2()
	store.putGame(context.Background(), game2.ID, game2)

	logs.info("Test games added")
}
