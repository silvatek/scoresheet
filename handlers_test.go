package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestLastPathElement(t *testing.T) {
	confirmLastPathElement(t, "/abc/xyz", "xyz")
	confirmLastPathElement(t, "/123", "123")
	confirmLastPathElement(t, "noslash", "noslash")
	confirmLastPathElement(t, "/abc/xyz?test=1", "xyz")
}

func confirmLastPathElement(t *testing.T, path string, expected string) {
	result := lastPathElement(path)
	if result != expected {
		t.Errorf("lastPathElement returned [%s], expected [%s]", result, expected)
	}
}

func TestQueryParam(t *testing.T) {
	confirmQueryParam(t, "/test?code=123", "code", "123")
	confirmQueryParam(t, "/test?code=123&other=xyz", "code", "123")
	confirmQueryParam(t, "/test", "code", "")
	confirmQueryParam(t, "/test?code=123", "other", "")
	confirmQueryParam(t, "", "code", "")
	confirmQueryParam(t, "/test?code", "code", "")
}

func confirmQueryParam(t *testing.T, uri string, param string, expected string) {
	result := queryParam(uri, param)
	if result != expected {
		t.Errorf("queryParam(%s) returned [%s], expected [%s]", param, result, expected)
	}
}

func TestHomePage(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	homePage(w, r)

	confirmSuccessResponse(w, t)

	doc, _ := goquery.NewDocumentFromReader(w.Body)

	confirmBodyIncludes("#intro", "Use this site to record details of UK recreational ice hockey games.",
		"Home page does not contain standard intro", doc, t)
}

func confirmSuccessResponse(w *httptest.ResponseRecorder, t *testing.T) {
	if w.Code >= 400 {
		t.Errorf("got HTTP status code %d, expected 2xx or 3xx", w.Code)
	}
}

func confirmBodyIncludes(query string, expected string, failMessage string, doc *goquery.Document, t *testing.T) {
	text := doc.Find(query).Text()
	if !strings.Contains(text, expected) {
		t.Errorf("%s => %s", failMessage, text)
	}
}

func TestGameIdParameter(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/testing?game=123", nil)

	id := gameIdParameter(r)

	if id != "123" {
		t.Errorf("Unexpected game ID: %s", id)
	}
}

// func TestGamePage(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest(http.MethodGet, "/game/CODE1", nil)

// 	dataStore = testDataStore()
// 	addTestGames(dataStore)

// 	gamePage(w, r)

// 	confirmSuccessResponse(w, t)

// 	doc, _ := goquery.NewDocumentFromReader(w.Body)

// 	confirmBodyIncludes("#game_summary", "Blues @ Reds, 2024-05-27", "Home page does not contain expected heading", doc, t)
// 	confirmBodyIncludes("td", "14:25 (25:35)", "Home page does not contain expected penalty", doc, t)
// }

func TestNewGamePage(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	newGamePage(w, r)

	confirmSuccessResponse(w, t)

	doc, _ := goquery.NewDocumentFromReader(w.Body)

	confirmBodyIncludes("h1", "New Game", "New game page does not contain expected heading", doc, t)
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

// func TestAddHandlers(t *testing.T) {
// 	addHandlers()
// }

// func TestGameRedirect(t *testing.T) {
// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest(http.MethodGet, "/games?game_id=xyz", nil)

// 	gameRedirect(w, r)

// 	if w.Result().StatusCode != http.StatusSeeOther {
// 		t.Errorf("Unexpected response code, was not redirect: %d", w.Result().StatusCode)
// 	}
// }

func TestNewEventPage(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/newEvent?game=CODE1", nil)

	newEventPage(w, r)

	confirmSuccessResponse(w, t)

	doc, _ := goquery.NewDocumentFromReader(w.Body)

	confirmBodyIncludes("h1", "New event for game CODE1", "New event page does not contain expected heading", doc, t)
}

func confirmRedirectTarget(expected string, w *httptest.ResponseRecorder, t *testing.T) {
	if w.Result().StatusCode != http.StatusSeeOther {
		t.Errorf("Unexpected response code, was not redirect: %d", w.Result().StatusCode)
	}
	if w.Result().Header.Get("Location") != expected {
		t.Errorf("Unexpected redirect target: %s", w.Result().Header.Get("Location"))
	}
}

func TestAddEventPost(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	content := "game_id=CODE1"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(content))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	addEventPost(w, r)

	confirmRedirectTarget("/game/CODE1", w, t)
}

func TestAddGamePost(t *testing.T) {
	dataStore = testDataStore()

	if !dataStore.isEmpty() {
		t.Error("Datastore should be empty before test")
	}

	content := "home_team='ABC'&away_test='XYZ'"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(content))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	addGamePost(w, r)

	if dataStore.isEmpty() {
		t.Error("Datastore should not be empty after test")
	}
}

func TestLockGameGet(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/lockGame?game=CODE1", nil)

	lockGame(w, r)

	confirmSuccessResponse(w, t)

	doc, _ := goquery.NewDocumentFromReader(w.Body)

	confirmBodyIncludes("h1", "Lock Game CODE1", "New event page does not contain expected heading", doc, t)
}

func TestLockGamePost(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	content := "game_id=CODE1&unlock_key=testing"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/lockGame", strings.NewReader(content))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	lockGame(w, r)

	confirmRedirectTarget("/game/CODE1", w, t)
}

func TestUnlockGameGet(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/unlockGame?game=CODE1", nil)

	unlockGame(w, r)

	confirmSuccessResponse(w, t)

	doc, _ := goquery.NewDocumentFromReader(w.Body)

	confirmBodyIncludes("h1", "Unlock Game CODE1", "New event page does not contain expected heading", doc, t)
}

func TestUnlockGamePost(t *testing.T) {
	dataStore = testDataStore()
	game := testGame2()
	dataStore.putGame(context.Background(), game.ID, game)

	content := "game_id=CODE2&unlock_key=secret123"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/unlockGame", strings.NewReader(content))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	unlockGame(w, r)

	confirmRedirectTarget("/game/CODE2", w, t)
}

func TestDeleteEventPage(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/deleteEvent?game=CODE1", nil)

	deleteEventPage(w, r)

	confirmSuccessResponse(w, t)

	doc, _ := goquery.NewDocumentFromReader(w.Body)

	confirmBodyIncludes("h1", "Delete event for game CODE1", "New event page does not contain expected heading", doc, t)
}

func TestDeleteEventPost(t *testing.T) {
	dataStore = testDataStore()
	setupDataStore(dataStore)

	content := "game_id=CODE1&event_summary=01:30 Home Goal"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/deleteEvent", strings.NewReader(content))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteEventPost(w, r)

	confirmRedirectTarget("/game/CODE1", w, t)
}
