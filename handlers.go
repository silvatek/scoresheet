package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"strconv"
	"time"

	"os"

	"html/template"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/skip2/go-qrcode"
)

type pageData struct {
	Message     string
	Error       string
	Game        Game
	Summary     GameSummary
	GameID      string
	GameURL     string
	Encoded     string
	EventType   string
	EventHA     string
	PageHeading string
	Stylesheet  string
	ItemType    string
	ItemCode    string
	Csrf        interface{}
	History     []HistoryItem
	Detail      interface{}
}

type Template struct {
	//templates *template.Template
}

func addRoutes(e *echo.Echo) {
	e.Renderer = &Template{}

	e.Use(middleware.Recover())
	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{TokenLookup: "form:_csrf"}))

	e.Static("/static", "template/static")
	e.File("/robots.txt", "template/static/robots.txt")

	AddBotHandlers(e)

	e.GET("/", homePage)
	e.GET("/games", codeRedirect)
	e.GET("/lists", codeRedirect)
	e.GET("/game/:id", gamePage)
	e.GET("/sharegame", shareLink)
	e.GET("/share", shareLink)
	e.GET("/qrcode", qrCodeGenerator)
	e.GET("/newEvent", newEventPage)
	e.POST("/addEvent", addEventPost)
	e.GET("/newGame", newGamePage)
	e.POST("/addGame", addGamePost)
	e.GET("/deleteEvent", deleteEventPage)
	e.POST("/deleteGameEvent", deleteEventPost)
	e.GET("/lockGame", lockGamePage)
	e.POST("/lockGame", lockGamePost)
	e.GET("/unlockGame", unlockGamePage)
	e.POST("/unlockGame", unlockGamePost)
	e.GET("/error", errorPage)
	e.GET("/setstyle", styleSet)
	e.GET("/addPlayer", addPlayerPage)
	e.POST("/addPlayer", addPlayerPost)
	e.GET("/list/:id", gameListPage)
	e.GET("/newList", newListPage)
	e.POST("/addList", addListPost)
	e.POST("/addListGame", addListGamePost)
	e.GET("/lock", lockItemPage)
	e.POST("/lock", lockItemPost)
	e.GET("/delete", deleteItemPage)
	e.POST("/delete", deleteItemPost)
	e.GET("/deleted", deletedItemPage)
	e.GET("/help", helpPage)
	e.GET("/cookies", cookiePage)
	e.GET("/privacy", privacyPage)
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return showTemplatePage(name, data, w, c)
}

func showTemplatePage(templateName string, data any, w io.Writer, c echo.Context) error {
	t, err := template.ParseFiles("template/base.html", "template/"+templateName+".html")
	if err != nil {
		logs.error("Error parsing template: %+v", err)
		os.Exit(-2)
	}

	if data == nil {
		data = pageData{}
	}

	data1, ok := data.(pageData)
	if ok {
		data1.Csrf = c.Get(middleware.DefaultCSRFConfig.ContextKey)

		if data1.PageHeading == "" {
			data1.PageHeading = "Ice Hockey Scoresheet"
		}

		stylecookie, err := getStyleCookie(c)
		if err == nil && stylecookie.Value != "" {
			data1.Stylesheet = stylecookie.Value
		} else {
			data1.Stylesheet = "scoresheet-simple"
			setStyleCookie(data1.Stylesheet, c)
		}

		data = data1
	}

	setSecurityHeaders(c)

	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		//msg := http.StatusText(http.StatusInternalServerError)
		logs.error("template.Execute: %v", err)
		//http.Error(w, msg, http.StatusInternalServerError)
	}

	return err
}

func setSecurityHeaders(c echo.Context) {
	c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; "+
		"img-src 'self' cdn.jsdelivr.net/; "+
		"script-src 'self' cdn.jsdelivr.net/; "+
		"style-src 'self' cdn.jsdelivr.net/; ",
	)
	c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
}

func getStyleCookie(c echo.Context) (*http.Cookie, error) {
	return c.Cookie("scoresheetstyle")
}

func setStyleCookie(stylesheetName string, c echo.Context) {
	stylecookie := new(http.Cookie)
	stylecookie.Name = "scoresheetstyle"
	stylecookie.Path = "/"
	stylecookie.Value = stylesheetName
	stylecookie.SameSite = http.SameSiteStrictMode
	stylecookie.Secure = true
	stylecookie.Expires = time.Now().Add(2 * 365 * time.Hour)
	c.SetCookie(stylecookie)
}

func homePage(c echo.Context) error {
	logs.info("Received request: %s", c.Path())

	history := getHistory(c)

	data := pageData{
		Message: "Ice Hockey Scoresheet",
		History: history,
	}

	writeHistoryCookie(HistoryString(history), c)

	return c.Render(http.StatusOK, "index", data)
}

// Redirect from query parameter URL to path parameter URL
func codeRedirect(c echo.Context) error {
	logs.info("Code redirect: %s", c.Path())

	gameId := strings.ToUpper(c.QueryParam("game_id"))
	if gameId != "" {
		return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
	}

	listId := strings.ToUpper(c.QueryParam("list_id"))
	if listId != "" {
		return c.Redirect(http.StatusSeeOther, "/list/"+listId)
	}

	return c.Redirect(http.StatusSeeOther, "/")
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
	ReqPath    string
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
	gameId := c.Param("id")
	if gameId == "" {
		gameId = c.QueryParam("game_id")
	}
	if gameId == "" {
		gameId = c.QueryParam("game")
	}
	values := GameRequestContext{
		GameId:     gameId,
		ReqPath:    c.Path(),
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
	data.PageHeading = data.Game.Title

	if data.Game.ID != gameId {
		return showErrorPage(fmt.Sprintf("Game not found: %s", gameId), c)
	}

	SortEvents(&(data.Game))
	data.Summary = summarise(data.Game)

	errorCode := c.QueryParam("e")
	if errorCode != "" {
		data.Error = errorMessage(errorCode)
	}

	setGameHistoryCookie(data.Game.LinkCode(), c)

	return c.Render(http.StatusOK, "game", data)
}

func setGameHistoryCookie(newItem string, c echo.Context) {
	history := getExistingHistory(c)

	history = AddToHistory(newItem, history)

	logs.debug1(gctx(c), "New game history: %s", history)

	writeHistoryCookie(history, c)
}

func writeHistoryCookie(history string, c echo.Context) {
	cookie := http.Cookie{
		Name:     "gameHistory",
		Value:    history,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   365 * 24 * 60 * 60, // 1 year
	}

	c.SetCookie(&cookie)
}

func getExistingHistory(c echo.Context) string {
	var gameList string
	current, err := c.Cookie("gameHistory")
	if err != http.ErrNoCookie {
		gameList = current.Value
		logs.debug1(gctx(c), "Loaded game history: %s", gameList)
	}
	return strings.Trim(gameList, " ")
}

func getHistory(c echo.Context) []HistoryItem {
	cookieValue := getExistingHistory(c)
	return GetHistory(gctx(c), cookieValue)
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
	eventType := c.QueryParam("type")

	ctx := gctx(c)
	logs.debug1(ctx, "Showing new event page for game %s", gameId)

	game := dataStore.getGame(ctx, gameId)

	if game.ID != gameId {
		return c.Redirect(http.StatusNotFound, fmt.Sprintf("Game not found when adding event: %s", gameId))
	}

	data := pageData{
		Game: game,
	}

	if eventType[0:1] == "A" {
		data.EventHA = "Away"
	} else {
		data.EventHA = "Home"
	}

	if eventType[1:2] == "P" {
		data.EventType = "Penalty"
	} else {
		data.EventType = "Goal"
	}

	data.PageHeading = data.EventHA + " " + data.EventType + ", " + game.Title

	return c.Render(http.StatusOK, "newevent", data)
}

func addEventPost(c echo.Context) error {
	gameId := c.FormValue("game_id")

	ctx := gctx(c)

	game := dataStore.getGame(ctx, gameId)

	var event Event

	err := c.Bind(&event)
	logs.debug("Bind errors: %v", err)

	event.ClockTime = EventTime(c.FormValue("minutes") + ":" + c.FormValue("seconds"))
	event.GameTime = ClockToGameTime(event.Period, event.ClockTime)

	AddEvent(&game, event)

	dataStore.putGame(ctx, gameId, game)

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func newGamePage(c echo.Context) error {
	data := pageData{}
	return c.Render(http.StatusOK, "newgame", data)
}

func addGamePost(c echo.Context) error {
	logs.debug("Received new game data")

	var game Game

	c.Bind(&game)
	game.Created = time.Now()

	gameDate, err := time.Parse("2006-01-02", game.GameDate)
	if err == nil {
		game.Title = game.AwayTeam + " @ " + game.HomeTeam + ", " + gameDate.Format("2 Jan 2006")
	} else {
		logs.debug("Could not parse game date `%s`, %v", game.GameDate, err)
		game.Title = game.AwayTeam + " @ " + game.HomeTeam + " on " + game.GameDate
	}

	gameId := dataStore.addGame(context.Background(), game)

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

type lockData struct {
	Type   string
	Code   string
	Action string
	Error  string
}

func lockItemPage(c echo.Context) error {
	errorText := ""
	errorCode := c.QueryParam("error")
	if errorCode == "1001" {
		errorText = "Incorrect unlock key"
	} else if errorCode == "1002" {
		errorText = "Unlock key must not be empty"
	}
	lockdata := lockData{
		Type:   c.QueryParam("type"),
		Code:   c.QueryParam("code"),
		Action: c.QueryParam("action"),
		Error:  errorText,
	}

	pageData := pageData{
		Detail: lockdata,
	}

	return c.Render(http.StatusOK, "lockitem", pageData)
}

func lockItemPost(c echo.Context) error {
	action := strings.ToLower(c.FormValue("action"))
	itemType := strings.ToLower(c.FormValue("item_type"))
	itemCode := strings.ToUpper(c.FormValue("item_code"))
	unlockKey := strings.TrimSpace(c.FormValue("unlock_key"))

	ctx := gctx(c)

	itemUrl := "/"

	if itemType == "list" {
		list := dataStore.getList(ctx, itemCode)

		if action == "lock" {
			if unlockKey == "" {
				return c.Redirect(http.StatusSeeOther, "/lock?error=1002&action=Lock&type=list&code="+itemCode)
			}
			list.LockedWith = unlockKey
		} else if action == "unlock" {
			if unlockKey != list.LockedWith {
				return c.Redirect(http.StatusSeeOther, "/lock?error=1001&action=Unlock&type=list&code="+itemCode)
			}
			list.LockedWith = ""
		}
		dataStore.putList(ctx, itemCode, list)
		itemUrl = "/list/" + itemCode
	}

	return c.Redirect(http.StatusSeeOther, itemUrl)
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

func itemUrl(path string, r *http.Request) string {
	schemeHeader, ok := r.Header["X-Forwarded-Proto"]
	var scheme string
	if ok {
		scheme = schemeHeader[0]
	} else {
		scheme = strings.ToLower(strings.Split(r.Proto, "/")[0])
	}
	return scheme + "://" + r.Host + path
}

type ShareInfo struct {
	Url     string
	Title   string
	Encoded string
	Type    string
	Code    string
}

func shareLink(c echo.Context) error {
	itemType := c.QueryParam("type")
	itemCode := c.QueryParam("code")

	itemUrl := itemUrl("/"+itemType+"/"+itemCode, c.Request())

	share := ShareInfo{
		Type:    itemType,
		Code:    itemCode,
		Url:     itemUrl,
		Encoded: html.EscapeString(itemUrl),
	}

	var data pageData
	data.GameID = itemCode
	data.GameURL = itemUrl
	data.Detail = share

	return c.Render(http.StatusOK, "sharelink", data)
}

func qrCodeGenerator(c echo.Context) error {
	itemType := c.QueryParam("type")
	itemCode := c.QueryParam("code")

	url := itemUrl("/"+itemType+"/"+itemCode, c.Request())

	headers := c.Response().Header()
	headers.Add("Content-Type", "image/png")
	c.Response().WriteHeader(http.StatusOK)

	q, _ := qrcode.New(url, qrcode.High)
	return q.Write(320, c.Response())
}

func styleSet(c echo.Context) error {
	styleName := c.QueryParam("style")

	setStyleCookie("scoresheet-"+styleName, c)

	return c.Redirect(http.StatusSeeOther, "/")
}

func addPlayerPage(c echo.Context) error {
	gameId := c.QueryParam("game")

	var data pageData
	data.GameID = gameId

	return c.Render(http.StatusOK, "newplayer", data)
}

func addPlayerPost(c echo.Context) error {
	gameId := c.FormValue("game_id")

	ctx := gctx(c)

	game := dataStore.getGame(ctx, gameId)

	if game.LockedWith != "" {
		return c.Redirect(http.StatusSeeOther, "/game/"+gameId+"?e=8001")
	}

	homeAway := c.FormValue("home_away")
	playerNum, err := strconv.Atoi(c.FormValue("player_number"))
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/game/"+gameId+"?e=8002")
	}
	playerName := c.FormValue("player_name")

	AddPlayer(&game, homeAway, playerNum, playerName)

	dataStore.putGame(ctx, gameId, game)

	return c.Redirect(http.StatusSeeOther, "/game/"+gameId)
}

func helpPage(c echo.Context) error {
	return c.Render(http.StatusOK, "help", nil)
}

func cookiePage(c echo.Context) error {
	return c.Render(http.StatusOK, "cookies", nil)
}

func privacyPage(c echo.Context) error {
	return c.Render(http.StatusOK, "datapolicy", nil)
}

type ListPageData struct {
	List  GameList
	Games []Game
}

func gameListPage(c echo.Context) error {
	listId := c.Param("id")

	ctx := gctx(c)
	logs.info1(ctx, "GET for list ID: %s", listId)

	var listData ListPageData
	listData.List = dataStore.getList(ctx, listId)

	for _, gameId := range listData.List.Games {
		game := dataStore.getGame(ctx, gameId)
		listData.Games = append(listData.Games, game)
	}

	var data pageData
	data.Detail = listData
	data.PageHeading = listData.List.Name

	if listData.List.ID != listId {
		return showErrorPage(fmt.Sprintf("List not found: %s", listId), c)
	}

	errorCode := c.QueryParam("e")
	if errorCode != "" {
		data.Error = errorMessage(errorCode)
	}

	setGameHistoryCookie(listData.List.LinkCode(), c)

	return c.Render(http.StatusOK, "gamelist", data)
}

func newListPage(c echo.Context) error {
	return c.Render(http.StatusOK, "newlist", nil)
}

func addListPost(c echo.Context) error {
	ctx := gctx(c)

	var list GameList
	list.Name = c.FormValue("list_name")
	id := dataStore.addList(ctx, list)

	return c.Redirect(http.StatusSeeOther, "/list/"+id)
}

func addListGamePost(c echo.Context) error {
	ctx := gctx(c)

	listId := c.FormValue("list_id")
	gameId := c.FormValue("game_id")

	list := dataStore.getList(ctx, listId)

	list.AddGame(gameId)

	dataStore.putList(ctx, listId, list)

	return c.Redirect(http.StatusSeeOther, "/list/"+listId)
}

func deleteItemPage(c echo.Context) error {
	var data pageData
	data.ItemType = c.QueryParam("type")
	data.ItemCode = c.QueryParam("code")

	return c.Render(http.StatusOK, "deleteitem", data)
}

// Deletes an item, but only if deleteCode matches itemCode.
func deleteItemPost(c echo.Context) error {
	itemCode := c.FormValue("item_code")
	itemType := c.FormValue("item_type")
	confirmCode := strings.ToUpper(strings.TrimSpace(c.FormValue("confirm_code")))

	if itemCode != confirmCode {
		return echo.NewHTTPError(http.StatusBadRequest, "Code does not match")
	}

	logs.info("Deleting %s %s at user's request", itemType, confirmCode)

	dataStore.deleteItem(gctx(c), itemType, confirmCode)

	return c.Redirect(http.StatusSeeOther, "/deleted?type="+itemType+"&code="+confirmCode)
}

func deletedItemPage(c echo.Context) error {
	var data pageData
	data.ItemType = c.QueryParam("type")
	data.ItemCode = c.QueryParam("code")

	return c.Render(http.StatusOK, "deleted", data)
}
