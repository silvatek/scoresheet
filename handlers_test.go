package main

import (
	"context"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAddRoutes(t *testing.T) {
	e := echo.New()
	addRoutes(e)

	if len(e.Routes()) < 20 {
		t.Errorf("Unexpected number of routes: %d", len(e.Routes()))
	}
}

func TestHomePage(t *testing.T) {
	wt := webTest(t)
	defer wt.showBodyOnFail()

	homePage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("#intro", "The Ice Hockey Scoresheet web application is designed to help score keepers")
}

func TestGamePage(t *testing.T) {
	wt := webTest(t)
	wt.setParam("id", TEST_ID_1)
	defer wt.showBodyOnFail()

	dataStore = GameStore{datastore: testDataStore()}
	addTestGames(dataStore)

	gamePage(wt.ec)

	wt.confirmSuccessResponse()

	wt.confirmHtmlIncludes("h1", "Blues @ Reds")
	//wt.confirmHtmlIncludes("span", "P1&nbsp;14:25")
}

func TestNewGamePage(t *testing.T) {
	wt := webTest(t)
	defer wt.showBodyOnFail()

	newGamePage(wt.ec)

	wt.confirmSuccessResponse()
	//wt.confirmHtmlIncludes("h1", "New Game")
}

func TestSetupDataStore(t *testing.T) {
	dataStore = GameStore{datastore: testDataStore()}

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

	codeRedirect(wt.ec)

	wt.confirmRedirect("/game/XYZ")
}

func TestNewEventPage(t *testing.T) {
	dataStore = GameStore{datastore: testDataStore()}
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.setQuery("type", "HG")
	wt.setQuery("game", "CODE1")

	newEventPage(wt.ec)

	wt.confirmSuccessResponse()
	//wt.confirmHtmlIncludes("h1", "Home Goal, Blues @ Reds, 27 May 2024")
}

func TestAddEventPost(t *testing.T) {
	dataStore = GameStore{datastore: testDataStore()}
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
	dataStore = GameStore{datastore: testDataStore()}

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

func TestDeleteEventPage(t *testing.T) {
	dataStore = GameStore{datastore: testDataStore()}
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.setQuery("game", TEST_ID_1)
	defer wt.showBodyOnFail()

	deleteEventPage(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("h1", "Delete event for game "+TEST_ID_1)
}

func TestDeleteEventPost(t *testing.T) {
	dataStore = GameStore{datastore: testDataStore()}
	setupDataStore(dataStore)

	wt := webTest(t)
	wt.post("game_id=" + TEST_ID_1 + "&event_summary=01:30 Home Goal")

	deleteEventPost(wt.ec)

	wt.confirmRedirect("/game/" + TEST_ID_1)
}

func TestShareGamePage(t *testing.T) {
	wt := webTest(t)
	wt.setQuery("type", "game")
	wt.setQuery("code", "SHARE-CODE")
	defer wt.showBodyOnFail()

	shareLink(wt.ec)

	wt.confirmSuccessResponse()
	wt.confirmHtmlIncludes("#linkurl", "SHARE-CODE")
}

func TestContentSecurityPolicy(t *testing.T) {
	wt := webTest(t)
	defer wt.showBodyOnFail()

	homePage(wt.ec)

	wt.confirmSuccessResponse()
	if wt.resp.Header().Get("Content-Security-Policy") == "" {
		t.Error("No content security policy found in response")
	}
}
