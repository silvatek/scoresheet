package main

import (
	"context"
	"fmt"
	"html"

	"os"

	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/skip2/go-qrcode"
)

type pageData struct {
	Message string
	Error   string
	Game    Game
	Summary GameSummary
	GameID  string
	GameURL string
	Encoded string
}

func addHandlers() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/game/", gamePage)
	http.HandleFunc("/games", gameRedirect)
	http.HandleFunc("/newEvent", newEventPage)
	http.HandleFunc("/addEvent", addEventPost)
	http.HandleFunc("/newGame", newGamePage)
	http.HandleFunc("/addGame", addGamePost)
	http.HandleFunc("/deleteEvent", deleteEventPage)
	http.HandleFunc("/deleteGameEvent", deleteEventPost)
	http.HandleFunc("/lockGame", lockGame)
	http.HandleFunc("/unlockGame", unlockGame)
	http.HandleFunc("/sharegame", shareGame)
	http.HandleFunc("/qrcode", qrCodeGenerator)

	addStaticAssetHandler()
}

func addStaticAssetHandler() {
	fs := http.FileServer(http.Dir("template/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

// show home/index page
func homePage(w http.ResponseWriter, r *http.Request) {
	logs.info("Received request: %s %s", r.Method, r.URL.Path)

	data := pageData{
		Message: "Ice Hockey Scoresheet",
	}

	showTemplatePage("index", data, w)
}

// Redirect from query parameter URL to path parameter URL
func gameRedirect(w http.ResponseWriter, r *http.Request) {
	logs.info("Game redirect: %s", r.RequestURI)
	gameId := strings.ToUpper(strings.TrimPrefix(r.RequestURI, "/games?game_id="))
	http.Redirect(w, r, "/game/"+gameId, http.StatusSeeOther)
}

func lastPathElement(uri string) string {
	// strip query parameters
	queryStart := strings.Index(uri, "?")
	if queryStart > -1 {
		uri = uri[:queryStart]
	}
	// return everything after the last slash
	lastSlash := strings.LastIndex(uri, "/")
	if lastSlash == -1 {
		return uri
	}
	return uri[lastSlash+1:]
}

func queryParam(uri string, param string) string {
	queryStart := strings.Index(uri, "?")
	if queryStart == -1 {
		return ""
	}
	uri = uri[queryStart+1:]

	paramStart := strings.Index(uri, param+"=")
	if paramStart == -1 {
		return ""
	}
	paramVal := uri[paramStart:]

	valueStart := strings.Index(uri, "=")
	paramVal = paramVal[valueStart+1:]

	nextStart := strings.Index(paramVal, "&")
	if nextStart > 0 {
		paramVal = paramVal[0:nextStart]
	}

	return paramVal
}

func errorMessage(errorCode string) string {
	if errorCode == "8001" {
		return "Unable to unlock game for editing"
	}
	return ""
}

type GameIdKeyType string
type RemoteAddrKeyType string

const GameIdKey = GameIdKeyType("game_id")
const RemoteAddrKey = RemoteAddrKeyType("remote_addr")

type GameRequestContext struct {
	GameId     string
	RemoteAddr string
	Request    *http.Request
}

func gameRequestContext(gameId string, r *http.Request) context.Context {
	values := GameRequestContext{
		GameId:     gameId,
		RemoteAddr: r.RemoteAddr,
	}
	if r.Header["X-Forwarded-For"] != nil {
		values.RemoteAddr = r.Header["X-Forwarded-For"][0]
	}
	return context.WithValue(r.Context(), GameIdKey, values)
}

func gamePage(w http.ResponseWriter, r *http.Request) {
	gameId := lastPathElement(r.URL.Path)

	ctx := gameRequestContext(gameId, r)
	logs.info1(ctx, "GET for game ID: %s", gameId)

	var data pageData
	data.Game = dataStore.getGame(ctx, gameId)

	if data.Game.ID != gameId {
		showErrorPage(fmt.Sprintf("Game not found: %s", gameId), w)
		return
	}

	SortEvents(&(data.Game))
	data.Summary = summarise(data.Game)

	errorCode := queryParam(r.RequestURI, "e")
	if errorCode != "" {
		data.Error = errorMessage(errorCode)
	}

	showTemplatePage("game", data, w)
}

func showErrorPage(error string, w http.ResponseWriter) {
	logs.info("Showing error page: %s", error)
	var data pageData
	data.Error = error
	showTemplatePage("error", data, w)
}

func showTemplatePage(templateName string, data any, w http.ResponseWriter) {
	t, err := template.ParseFiles("template/" + templateName + ".html")
	if err != nil {
		logs.error("Error parsing template: %+v", err)
		os.Exit(-2)
	}

	if err := t.Execute(w, data); err != nil {
		msg := http.StatusText(http.StatusInternalServerError)
		logs.debug("template.Execute: %v", err)
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

func newEventPage(w http.ResponseWriter, r *http.Request) {
	gameId := gameIdParameter(r)

	ctx := gameRequestContext(gameId, r)
	logs.debug1(ctx, "Showing new event page for game %s", gameId)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		showErrorPage(fmt.Sprintf("Game not found when adding event: %s", gameId), w)
		return
	}

	data := pageData{
		Game: game,
	}

	showTemplatePage("newevent", data, w)
}

func newGamePage(w http.ResponseWriter, r *http.Request) {
	showTemplatePage("newgame", "", w)
}

func addEventPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	gameId := r.Form.Get("game_id")
	playerId, _ := strconv.Atoi(r.Form.Get("player"))

	ctx := gameRequestContext(gameId, r)
	logs.debug1(ctx, "Received new event data, Game ID = %s, Player = %d, Event = %s",
		gameId, playerId, r.Form.Get("event_type"))

	game := dataStore.getGame(ctx, gameId)

	period, _ := strconv.Atoi(r.Form.Get("period"))
	clockTime := r.Form.Get("minutes") + ":" + r.Form.Get("seconds")
	homeAway := r.Form.Get("home_away")
	category := r.Form.Get("category")

	assist1, _ := strconv.Atoi(r.Form.Get("assist1"))
	assist2, _ := strconv.Atoi(r.Form.Get("assist2"))
	minutes, _ := strconv.Atoi(r.Form.Get("minutes"))

	if r.Form.Get("event_type") == "Penalty" {
		AddPenalty(&game, period, EventTime(clockTime), homeAway, playerId, minutes, category)
	} else {
		AddGoal(&game, period, EventTime(clockTime), homeAway, playerId, assist1, assist2, category)
	}

	dataStore.putGame(ctx, gameId, game)

	http.Redirect(w, r, "/game/"+gameId, http.StatusSeeOther)
}

func addGamePost(w http.ResponseWriter, r *http.Request) {
	logs.debug("Received new game data")

	r.ParseForm()

	var game Game

	game.HomeTeam = r.Form.Get("home_team")
	game.AwayTeam = r.Form.Get("away_team")
	game.GameDate = r.Form.Get("game_date")
	game.Title = game.AwayTeam + " @ " + game.HomeTeam + " on " + game.GameDate

	gameId := dataStore.addGame(context.Background(), &game)

	http.Redirect(w, r, "/game/"+gameId, http.StatusSeeOther)
}

func deleteEventPage(w http.ResponseWriter, r *http.Request) {
	gameId := gameIdParameter(r)

	ctx := gameRequestContext(gameId, r)
	logs.debug1(ctx, "Showing delete event page for game %s", gameId)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		showErrorPage(fmt.Sprintf("Game not found when deleting event: %s", gameId), w)
		return
	}

	SortEvents(&game)

	data := pageData{
		Game: game,
	}

	showTemplatePage("deleteevent", data, w)
}

func deleteEventPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	gameId := r.Form.Get("game_id")
	ctx := gameRequestContext(gameId, r)

	requestedEvent := r.Form.Get("event_summary")
	logs.debug("Received delete event request for %s, %s", gameId, requestedEvent)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		showErrorPage(fmt.Sprintf("Game not found when deleting event: %s", gameId), w)
		return
	}
	if game.LockedWith != "" {
		showErrorPage(fmt.Sprintf("Attempting to delete event from locked game: %s", gameId), w)
		return
	}

	for n, event := range game.Events {
		event_summary := fmt.Sprintf("%s %s %s", event.GameTime, event.HomeAway, event.EventType)
		if event_summary == requestedEvent {
			logs.debug("Found event to delete")
			event.GameTime = "99:99"
			event.Period = 0
			event.ClockTime = ""
			event.EventType = "DELETED"
			event.HomeAway = ""
			event.Category = ""
			event.Player = 0
			event.Assist1 = 0
			event.Assist2 = 0
			event.Minutes = 0
		}
		game.Events[n] = event
	}

	SortEvents(&game)

	eventCount := len(game.Events)
	if game.Events[eventCount-1].EventType == "DELETED" {
		game.Events = game.Events[:eventCount-1]
	} else {
		logs.error("Last event not marked for deletion")
	}

	dataStore.putGame(ctx, gameId, game)

	http.Redirect(w, r, "/game/"+gameId, http.StatusSeeOther)
}

func lockGame(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		gameId := r.Form.Get("game_id")

		ctx := gameRequestContext(gameId, r)

		game := dataStore.getGame(ctx, gameId)

		userKey := r.Form.Get("unlock_key")

		if game.LockedWith == "" {
			game.LockedWith = userKey
			dataStore.putGame(ctx, gameId, game)
		}

		http.Redirect(w, r, "/game/"+gameId, http.StatusSeeOther)
	} else {
		gameId := gameIdParameter(r)

		ctx := gameRequestContext(gameId, r)

		game := dataStore.getGame(ctx, gameId)
		data := pageData{
			Game: game,
		}
		showTemplatePage("lockgame", data, w)
	}
}

func unlockGame(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		gameId := r.Form.Get("game_id")

		ctx := gameRequestContext(gameId, r)

		game := dataStore.getGame(ctx, gameId)

		userKey := r.Form.Get("unlock_key")

		var errorSuffix string

		if game.LockedWith == userKey {
			game.LockedWith = ""
			dataStore.putGame(ctx, gameId, game)
		} else {
			errorSuffix = "?e=8001"
		}

		http.Redirect(w, r, "/game/"+gameId+errorSuffix, http.StatusSeeOther)
	} else {
		gameId := gameIdParameter(r)

		ctx := gameRequestContext(gameId, r)

		game := dataStore.getGame(ctx, gameId)
		data := pageData{
			Game: game,
		}
		showTemplatePage("unlockgame", data, w)
	}
}

func gameIdParameter(r *http.Request) string {
	params := strings.Split(r.RequestURI, "?")[1]
	return strings.ToUpper(strings.TrimPrefix(params, "game="))
}

func gameUrl(gameId string, r *http.Request) string {
	proto := strings.ToLower(strings.Split(r.Proto, "/")[0])
	return proto + "://" + r.Host + "/game/" + gameId
}

func shareGame(w http.ResponseWriter, r *http.Request) {
	gameId := gameIdParameter(r)

	gameUrl := gameUrl(gameId, r)

	var data pageData
	data.GameID = gameId
	data.GameURL = gameUrl
	data.Encoded = html.EscapeString(gameUrl)

	showTemplatePage("sharegame", data, w)
}

func qrCodeGenerator(w http.ResponseWriter, r *http.Request) {
	gameId := gameIdParameter(r)

	gameUrl := gameUrl(gameId, r)

	tempFileName := os.Getenv("TMPDIR") + "/" + gameId + ".png"

	qrcode.WriteFile(gameUrl, qrcode.High, 256, tempFileName)
	defer os.Remove(tempFileName)

	content, _ := os.ReadFile(tempFileName)

	headers := w.Header()
	headers.Add("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(content))
}
