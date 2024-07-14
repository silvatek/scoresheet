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
	History []GameRef
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
	c.GET("/unlockGame", unlockGamePage)
	c.POST("/unlockGame", unlockGamePost)
	c.GET("/error", errorPage)
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return showTemplatePage(name, data, w)
}

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
		History: gameHistory(c),
	}

	return c.Render(http.StatusOK, "index", data)
}

// Redirect from query parameter URL to path parameter URL
func gameRedirect(c echo.Context) error {
	logs.info("Game redirect: %s", c.Path())
	gameId := strings.ToUpper(c.QueryParam("game_id"))
	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func errorMessage(errorCode string) string {
	if errorCode == "8001" {
		return "Unable to unlock game for editing"
	}
	return ""
}

type GameRequestKeyType string

const GameRequestKey = GameRequestKeyType("game_request")

type GameRequestContext struct {
	GameId     string
	RemoteAddr string
	Request    *http.Request
	TraceID    string
	SpanID     string
}

func parseCloudTrace(trace string) (string, string, string) {
	if strings.Contains(trace, "/") {
		parts := strings.Split(trace, "/")

		if len(parts) >= 2 {
			if strings.Contains(parts[1], ";") {
				spanParts := strings.Split(parts[1], ";")
				return parts[0], spanParts[0], spanParts[1]
			} else {
				return parts[0], parts[1], ""
			}
		}
	}
	return "", "", ""
}

func gctx(c echo.Context) context.Context {
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
		Request:    c.Request(),
	}
	if len(c.Request().Header["X-Cloud-Trace-Context"]) > 0 {
		values.TraceID, values.SpanID, _ = parseCloudTrace(c.Request().Header["X-Cloud-Trace-Context"][0])
	}
	return context.WithValue(c.Request().Context(), GameRequestKey, values)
}

func gamePage(c echo.Context) error {
	gameId := c.Param("id")

	ctx := gctx(c)
	logs.info1(ctx, "GET for game ID: %s", gameId)

	var data pageData
	data.Game = dataStore.getGame(ctx, gameId)

	if data.Game.ID != gameId {
		return showErrorPage(fmt.Sprintf("Game not found: %s", gameId), c)
	}

	SortEvents(&(data.Game))
	data.Summary = summarise(data.Game)

	errorCode := c.QueryParam("e")
	if errorCode != "" {
		data.Error = errorMessage(errorCode)
	}

	setGameHistoryCookie(gameId, c)

	return c.Render(http.StatusOK, "game", data)
}

func setGameHistoryCookie(gameId string, c echo.Context) {
	gameList := getExistingGameList(c)

	gameList = gameId + " " + strings.Trim(strings.ReplaceAll(gameList, gameId, " "), " ")
	logs.debug1(gctx(c), "New game history: %s", gameList)

	cookie := http.Cookie{
		Name:     "gameHistory",
		Value:    gameList,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	c.SetCookie(&cookie)
}

func getExistingGameList(c echo.Context) string {
	var gameList string
	current, err := c.Cookie("gameHistory")
	if err != http.ErrNoCookie {
		gameList = current.Value
		logs.debug1(gctx(c), "Loaded game history: %s", gameList)
	}
	return strings.Trim(gameList, " ")
}

type GameRef struct {
	ID    string
	Title string
}

func gameHistory(c echo.Context) []GameRef {
	cookieValue := getExistingGameList(c)
	var ids []string
	var games []GameRef
	if cookieValue != "" {
		ids = strings.Split(cookieValue, " ")
		for _, id := range ids {
			gs := GameRef{id, dataStore.getGame(gctx(c), id).Title}
			games = append(games, gs)
		}
	}
	return games
}

func errorPage(c echo.Context) error {
	errorCode := c.QueryParam("e")
	message := errorMessage(errorCode)
	return showErrorPage(message, c)
}

func showErrorPage(error string, c echo.Context) error {
	logs.info("Showing error page: %s", error)
	var data pageData
	data.Error = error
	return c.Render(http.StatusOK, "error", data)
}

func newEventPage(c echo.Context) error {
	gameId := c.QueryParam("game")

	ctx := gctx(c)
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

	ctx := gctx(c)
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

	ctx := gctx(c)
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
	ctx := gctx(c)

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

	ctx := gctx(c)

	game := dataStore.getGame(ctx, gameId)
	data := pageData{
		Game: game,
	}

	return c.Render(http.StatusOK, "lockgame", data)
}

func lockGamePost(c echo.Context) error {
	gameId := c.FormValue("game_id")

	ctx := gctx(c)

	game := dataStore.getGame(ctx, gameId)

	userKey := c.FormValue("unlock_key")

	if game.LockedWith == "" {
		game.LockedWith = userKey
		dataStore.putGame(ctx, gameId, game)
	}

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func unlockGamePage(c echo.Context) error {
	gameId := c.QueryParam("game")

	game := dataStore.getGame(gctx(c), gameId)
	data := pageData{
		Game: game,
	}
	return c.Render(http.StatusOK, "unlockgame", data)
}

func unlockGamePost(c echo.Context) error {
	gameId := c.FormValue("game_id")

	ctx := gctx(c)

	game := dataStore.getGame(ctx, gameId)

	userKey := c.FormValue("unlock_key")

	var errorSuffix string

	if game.LockedWith == userKey {
		game.LockedWith = ""
		dataStore.putGame(ctx, gameId, game)
	} else {
		errorSuffix = "?e=8001"
	}

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId+errorSuffix)
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
