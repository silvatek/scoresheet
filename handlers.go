package main

import (
	"context"
	"fmt"
	"html"
	"io"

	"os"

	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
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
	History []string
}

type Template struct {
	//templates *template.Template
}

func addRoutes(c *echo.Echo) {
	c.Renderer = &Template{}

	c.Static("/static", "template/static")

	c.GET("/", homePage)
	c.GET("/games", gameRedirect)
	c.GET("/game/:id", gamePage)
	c.GET("/sharegame", shareGame)
	c.GET("/qrcode", qrCodeGenerator)
	c.GET("/newEvent", newEventPage)
	c.POST("/addEvent", addEventPost)
	c.GET("/newGame", newGamePage)
	c.POST("/addGame", addGamePost)
	c.GET("/deleteEvent", deleteEventPage)
	c.POST("/deleteGameEvent", deleteEventPost)
	c.GET("/lockGame", lockGamePage)
	c.POST("/lockGame", lockGamePost)
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return showTemplatePage(name, data, w)
}

// func addHandlers() *mux.Router {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/", homePage)
// 	r.HandleFunc("/game/{id}", gamePage)
// 	r.HandleFunc("/games", gameRedirect)
// 	r.HandleFunc("/newEvent", newEventPage)
// 	r.HandleFunc("/addEvent", addEventPost)
// 	r.HandleFunc("/newGame", newGamePage)
// 	r.HandleFunc("/addGame", addGamePost)
// 	r.HandleFunc("/deleteEvent", deleteEventPage)
// 	r.HandleFunc("/deleteGameEvent", deleteEventPost)
// 	r.HandleFunc("/lockGame", lockGame)
// 	r.HandleFunc("/unlockGame", unlockGame)
// 	r.HandleFunc("/sharegame", shareGame)
// 	r.HandleFunc("/qrcode", qrCodeGenerator)

// 	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", http.FileServer(http.Dir("template/static"))))

// 	return r
// }

func showTemplatePage(templateName string, data any, w io.Writer) error {
	t, err := template.ParseFiles("template/base.html", "template/"+templateName+".html")
	if err != nil {
		logs.error("Error parsing template: %+v", err)
		os.Exit(-2)
	}

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		//msg := http.StatusText(http.StatusInternalServerError)
		logs.error("template.Execute: %v", err)
		//http.Error(w, msg, http.StatusInternalServerError)
	}

	return err
}

func homePage(c echo.Context) error {
	logs.info("Received request: %s", c.Path())

	data := pageData{
		Message: "Ice Hockey Scoresheet",
	}

	return c.Render(http.StatusOK, "index", data)
}

// Redirect from query parameter URL to path parameter URL
func gameRedirect(c echo.Context) error {
	logs.info("Game redirect: %s", c.Path())
	gameId := strings.ToUpper(c.QueryParam("game_id"))
	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
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

func ctx(c echo.Context) context.Context {
	gameId := c.Param("gameId")
	if gameId == "" {
		gameId = c.QueryParam("game_id")
	}
	if gameId == "" {
		gameId = c.QueryParam("game")
	}
	values := GameRequestContext{
		GameId:     gameId,
		RemoteAddr: c.RealIP(),
	}
	return context.WithValue(c.Request().Context(), GameIdKey, values)
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

func gamePage(c echo.Context) error {
	gameId := c.Param("id")

	ctx := ctx(c)
	logs.info1(ctx, "GET for game ID: %s", gameId)

	var data pageData
	data.Game = dataStore.getGame(ctx, gameId)

	if data.Game.ID != gameId {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Game not found: %s", gameId))
	}

	SortEvents(&(data.Game))
	data.Summary = summarise(data.Game)

	// errorCode := queryParam(r.RequestURI, "e")
	// if errorCode != "" {
	// 	data.Error = errorMessage(errorCode)
	// }

	// setGameHistoryCookie(gameId, w, r)

	return c.Render(http.StatusOK, "game", data)
}

func setGameHistoryCookie(gameId string, w http.ResponseWriter, r *http.Request) {
	gameList := getExistingGameList(r)

	gameList = gameId + " " + strings.Trim(strings.ReplaceAll(gameList, gameId, " "), " ")
	logs.debug1(context.Background(), "New game history: %s", gameList)

	cookie := http.Cookie{
		Name:     "gameHistory",
		Value:    gameList,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
}

func getExistingGameList(r *http.Request) string {
	var gameList string
	current, err := r.Cookie("gameHistory")
	if err != http.ErrNoCookie {
		gameList = current.Value
		logs.debug1(context.Background(), "Loaded game history: %s", gameList)
	}
	return strings.Trim(gameList, " ")
}

func gameHistory(r *http.Request) []string {
	cookieValue := getExistingGameList(r)
	var games []string
	if cookieValue != "" {
		games = strings.Split(cookieValue, " ")
	}
	return games
}

func showErrorPage(error string, w http.ResponseWriter) {
	logs.info("Showing error page: %s", error)
	var data pageData
	data.Error = error
	showTemplatePage("error", data, w)
}

func newEventPage(c echo.Context) error {
	gameId := c.QueryParam("game")

	ctx := context.Background() //gameRequestContext(gameId, r)
	logs.debug1(ctx, "Showing new event page for game %s", gameId)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		return c.Redirect(http.StatusNotFound, fmt.Sprintf("Game not found when adding event: %s", gameId))
	}

	data := pageData{
		Game: game,
	}

	return c.Render(http.StatusOK, "newevent", data)
}

func addEventPost(c echo.Context) error {
	gameId := c.FormValue("game_id")
	playerId, _ := strconv.Atoi(c.FormValue("player"))

	ctx := ctx(c)
	logs.debug1(ctx, "Received new event data, Game ID = %s, Player = %d, Event = %s",
		gameId, playerId, c.FormValue("event_type"))

	game := dataStore.getGame(ctx, gameId)

	period, _ := strconv.Atoi(c.FormValue("period"))
	clockTime := c.FormValue("minutes") + ":" + c.FormValue("seconds")
	homeAway := c.FormValue("home_away")
	category := c.FormValue("category")

	assist1, _ := strconv.Atoi(c.FormValue("assist1"))
	assist2, _ := strconv.Atoi(c.FormValue("assist2"))
	minutes, _ := strconv.Atoi(c.FormValue("minutes"))

	if c.FormValue("event_type") == "Penalty" {
		AddPenalty(&game, period, EventTime(clockTime), homeAway, playerId, minutes, category)
	} else {
		AddGoal(&game, period, EventTime(clockTime), homeAway, playerId, assist1, assist2, category)
	}

	dataStore.putGame(ctx, gameId, game)

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func newGamePage(c echo.Context) error {
	return c.Render(http.StatusOK, "newgame", "")
}

func addGamePost(c echo.Context) error {
	logs.debug("Received new game data")

	var game Game

	game.HomeTeam = c.FormValue("home_team")
	game.AwayTeam = c.FormValue("away_team")
	game.GameDate = c.FormValue("game_date")
	game.Title = game.AwayTeam + " @ " + game.HomeTeam + " on " + game.GameDate

	gameId := dataStore.addGame(context.Background(), &game)

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func deleteEventPage(c echo.Context) error {
	gameId := c.QueryParam("game")

	ctx := ctx(c)
	logs.debug1(ctx, "Showing delete event page for game %s", gameId)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		return c.Redirect(http.StatusNotFound, fmt.Sprintf("Game not found when deleting event: %s", gameId))
	}

	SortEvents(&game)

	data := pageData{
		Game: game,
	}

	return c.Render(http.StatusOK, "deleteevent", data)
}

func deleteEventPost(c echo.Context) error {

	gameId := c.FormValue("game_id")
	ctx := ctx(c)

	requestedEvent := c.FormValue("event_summary")
	logs.debug("Received delete event request for %s, %s", gameId, requestedEvent)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("Game not found when deleting event: %s", gameId))
	}
	if game.LockedWith != "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Attempting to delete event from locked game: %s", gameId))
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

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func lockGamePage(c echo.Context) error {
	gameId := c.QueryParam("game")

	ctx := ctx(c)

	game := dataStore.getGame(ctx, gameId)
	data := pageData{
		Game: game,
	}

	return c.Render(http.StatusOK, "lockgame", data)
}

func lockGamePost(c echo.Context) error {
	gameId := c.FormValue("game_id")

	ctx := ctx(c)

	game := dataStore.getGame(ctx, gameId)

	userKey := c.FormValue("unlock_key")

	if game.LockedWith == "" {
		game.LockedWith = userKey
		dataStore.putGame(ctx, gameId, game)
	}

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
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
	schemeHeader, ok := r.Header["X-Forwarded-Proto"]
	var scheme string
	if ok {
		scheme = schemeHeader[0]
	} else {
		scheme = strings.ToLower(strings.Split(r.Proto, "/")[0])
	}
	return scheme + "://" + r.Host + "/game/" + gameId
}

func shareGame(c echo.Context) error {
	gameId := c.QueryParam("game")

	gameUrl := gameUrl(gameId, c.Request())

	var data pageData
	data.GameID = gameId
	data.GameURL = gameUrl
	data.Encoded = html.EscapeString(gameUrl)

	return c.Render(http.StatusOK, "sharegame", data)
}

func qrCodeGenerator(c echo.Context) error {
	gameId := c.QueryParam("game")

	gameUrl := gameUrl(gameId, c.Request())

	headers := c.Response().Header()
	headers.Add("Content-Type", "image/png")
	c.Response().WriteHeader(http.StatusOK)

	q, _ := qrcode.New(gameUrl, qrcode.High)
	return q.Write(320, c.Response())
}
