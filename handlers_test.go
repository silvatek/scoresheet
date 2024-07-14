package main

import (
	"context"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAddRoutes(t *testing.T) {
	e := echo.New()
	addRoutes(e)

	if len(e.Routes()) != 17 {
		t.Errorf("Unexpected number of routes: %d", len(e.Routes()))
	}
}

func TestHomePage(t *testing.T) {
	wt := webTest(t)
	defer wt.showBodyOnFail()

	homePage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("#intro", "Use this site to record details of UK recreational ice hockey games.")
}

func TestGamePage(t *testing.T) {
	wt := webTest(t)
	wt.setParam("id", "CODE1")
	defer wt.showBodyOnFail()

	dataStore = testDataStore()
	addTestGames(dataStore)

	gamePage(wt.ec)

	wt.confirmSuccessResponse()

	wt.confirmHtmlIncludes("h1", "Blues @ Reds, 27 May 2024")
	wt.confirmHtmlIncludes("td", "14:25 (25:35)")
}

func TestNewGamePage(t *testing.T) {
	wt := webTest(t)
	defer wt.showBodyOnFail()

	newGamePage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("h1", "New Game")
}

func TestSetupDataStore(t *testing.T) {
	dataStore := testDataStore()

	if !dataStore.isEmpty() {
		t.Error("Datastore should be empty before setup")
	}

	setupDataStore(dataStore)

	if dataStore.isEmpty() {
		t.Error("Datastore should not be empty after setup")
	}
}

func TestGameRedirect(t *testing.T) {
	wt := webTest(t)
	wt.setQuery("game_id", "xyz")

	gameRedirect(wt.ec)

	wt.confirmRedirect("/game/XYZ")
}

func TestNewEventPage(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.setQuery("type", "HG")
	wt.setQuery("game", "CODE1")

	newEventPage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("h1", "Home Goal for game CODE1")
}

func TestAddEventPost(t *testing.T) {
	dataStore = testDataStore()
	dataStore.putGame(context.TODO(), "CODE1", Game{ID: "CODE1"})

	wt := webTest(t)
	wt.post("game_id=CODE1&period=2&minutes=5&seconds=0")

	addEventPost(wt.ec)

	wt.confirmRedirect("/game/CODE1")

	game := dataStore.getGame(context.TODO(), "CODE1")
	if len(game.Events) < 1 {
		t.Error("Game has no events after addEventPost")
		return
	}
	event := game.Events[0]
	if event.Period != 2 {
		t.Errorf("Unexpected event period: %d", event.Period)
	}
	if event.Minutes != 0 {
		t.Errorf("Unexpected event minutes: %d", event.Minutes)
	}
}

func TestAddGamePost(t *testing.T) {
	dataStore = testDataStore()

	if !dataStore.isEmpty() {
		t.Error("Datastore should be empty before test")
	}

	wt := webTest(t)
	wt.post("home_team='ABC'&away_test='XYZ'")

	addGamePost(wt.ec)

	if dataStore.isEmpty() {
		t.Error("Datastore should not be empty after test")
	}
}

func TestLockGamePage(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.setQuery("game", "CODE1")
	wt.showBodyOnFail()

	lockGamePage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("h1", "Lock Game CODE1")
}

func TestLockGamePost(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.post("game_id=CODE1&unlock_key=testing")

	lockGamePost(wt.ec)

	wt.confirmRedirect("/game/CODE1")
}

func TestUnlockGamePage(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.setQuery("game", "CODE1")
	wt.showBodyOnFail()

	unlockGamePage(wt.ec)

	wt.confirmSuccessResponse()

	wt.confirmHtmlIncludes("h1", "Unlock Game CODE1")
}

func TestUnlockGamePost(t *testing.T) {
	dataStore = testDataStore()
	game := testGame2()
	dataStore.putGame(context.Background(), game.ID, game)

	wt := webTest(t)
	wt.post("game_id=CODE2&unlock_key=secret123")

	unlockGamePost(wt.ec)

	wt.confirmRedirect("/game/CODE2")
}

func TestDeleteEventPage(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.setQuery("game", "CODE1")
	defer wt.showBodyOnFail()

	deleteEventPage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("h1", "Delete event for game CODE1")
}

func TestDeleteEventPost(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.post("game_id=CODE1&event_summary=01:30 Home Goal")

	deleteEventPost(wt.ec)

	wt.confirmRedirect("/game/CODE1")
}

func TestShareGamePage(t *testing.T) {
	wt := webTest(t)
	wt.setQuery("game", "SHARE-CODE")
	defer wt.showBodyOnFail()

	shareGame(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("#gameurl", "SHARE-CODE")
}
