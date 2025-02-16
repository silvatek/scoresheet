package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Game struct {
	ID          string
	Title       string
	Events      []Event
	Period      int
	GameDate    string `form:"game_date"`
	HomeTeam    string `form:"home_team"`
	AwayTeam    string `form:"away_team"`
	Venue       string
	Competition string
	LockedWith  string
	HomePlayers map[int]string
	AwayPlayers map[int]string
}

type EventTime string

const GOAL = "Goal"
const PENALTY = "Penalty"
const HOME = "Home"
const AWAY = "Away"

type Event struct {
	ClockTime EventTime
	GameTime  EventTime
	Period    int    `form:"period"`
	EventType string `form:"event_type"`
	HomeAway  string `form:"home_away"`
	Category  string `form:"category"`
	Player    int    `form:"player"`
	Assist1   int    `form:"assist1"`
	Assist2   int    `form:"assist2"`
	Minutes   int    `form:"penaltyMinutes"`
}

type PlayerSummary struct {
	Goals   int
	Assists int
	Minutes int
}

type PeriodSummary struct {
	Title         string
	HomeGoals     int
	AwayGoals     int
	HomePenalties int
	AwayPenalties int
}

type GameSummary struct {
	HomeGoals   int
	AwayGoals   int
	HomePlayers map[int]PlayerSummary
	AwayPlayers map[int]PlayerSummary
	Periods     []PeriodSummary
}

func AddEvent(game *Game, event Event) {
	game.Events = append(game.Events, event)

	if event.Period > game.Period {
		game.Period = event.Period
	}
}

func AddGoal(game *Game, period int, clockTime EventTime, homeAway string, player int, assist1 int, assist2 int, category string) {
	goal := Event{
		Period:    period,
		ClockTime: clockTime,
		EventType: GOAL,
		HomeAway:  homeAway,
		Player:    player,
		Assist1:   assist1,
		Assist2:   assist2,
		Category:  category,
	}
	goal.GameTime = ClockToGameTime(period, clockTime)
	game.Events = append(game.Events, goal)

	if period > game.Period {
		game.Period = period
	}
}

func AddPenalty(game *Game, period int, clockTime EventTime, homeAway string, player int, minutes int, category string) {
	penalty := Event{
		Period:    period,
		ClockTime: clockTime,
		EventType: PENALTY,
		HomeAway:  homeAway,
		Player:    player,
		Minutes:   minutes,
		Category:  category,
	}
	penalty.GameTime = ClockToGameTime(period, clockTime)
	game.Events = append(game.Events, penalty)

	if period > game.Period {
		game.Period = period
	}
}

func eventTime(mins int, secs int) EventTime {
	return EventTime(fmt.Sprintf("%02d:%02d", mins, secs))
}

func ClockToGameTime(period int, clockTime EventTime) EventTime {
	mins, secs := parseEventTime(clockTime)
	periodMins := 20 - mins
	var gameSecs int
	if secs > 0 {
		periodMins -= 1
		gameSecs = 60 - secs
	} else {
		gameSecs = 0
	}
	gameMins := periodMins + 20*(period-1)
	return eventTime(gameMins, gameSecs)
}

func GameToClockTime(gameTime EventTime) (clockTime EventTime, period int) {
	mins, secs := parseEventTime(gameTime)

	period = (mins / 20) + 1
	clockMins := 20 - (mins % 20)
	var clockSecs int
	if secs == 0 {
		clockSecs = 0
	} else {
		clockSecs = 60 - secs
		clockMins -= 1
	}

	return eventTime(clockMins, clockSecs), period
}

func parseEventTime(time EventTime) (int, int) {
	parts := strings.Split(string(time), ":")
	mins, _ := strconv.Atoi(parts[0])
	secs, _ := strconv.Atoi(parts[1])
	return mins, secs
}

const GAME_TOTAL = 4

func summarise(game Game) GameSummary {
	var summary GameSummary

	summary.HomePlayers = make(map[int]PlayerSummary)
	summary.AwayPlayers = make(map[int]PlayerSummary)
	summary.Periods = make([]PeriodSummary, 5)

	summary.Periods[0].Title = "P1"
	summary.Periods[1].Title = "P2"
	summary.Periods[2].Title = "P3"
	summary.Periods[3].Title = "OT"
	summary.Periods[4].Title = "Total"

	logs.debug("Summarising %d events in %s", len(game.Events), game.ID)

	for _, event := range game.Events {
		if event.EventType == GOAL && event.HomeAway == HOME {
			summary.HomeGoals++
			summary.Periods[event.Period-1].HomeGoals++
			summary.Periods[GAME_TOTAL].HomeGoals++
			countPlayerEvent(event.Player, summary.HomePlayers, 1, 0, 0)
			countAssists(event, summary.HomePlayers)
		}
		if event.EventType == GOAL && event.HomeAway == AWAY {
			summary.AwayGoals++
			summary.Periods[event.Period-1].AwayGoals++
			summary.Periods[GAME_TOTAL].AwayGoals++
			countPlayerEvent(event.Player, summary.AwayPlayers, 1, 0, 0)
			countAssists(event, summary.AwayPlayers)
		}
		if event.EventType == PENALTY && event.HomeAway == HOME {
			summary.Periods[event.Period-1].HomePenalties += event.Minutes
			summary.Periods[GAME_TOTAL].HomePenalties += event.Minutes
			countPlayerEvent(event.Player, summary.HomePlayers, 0, 0, event.Minutes)
		}
		if event.EventType == PENALTY && event.HomeAway == AWAY {
			summary.Periods[event.Period-1].AwayPenalties += event.Minutes
			summary.Periods[GAME_TOTAL].AwayPenalties += event.Minutes
			countPlayerEvent(event.Player, summary.AwayPlayers, 0, 0, event.Minutes)
		}
	}
	return summary
}

func countAssists(event Event, players map[int]PlayerSummary) {
	if event.Assist1 > 0 {
		countPlayerEvent(event.Assist1, players, 0, 1, 0)
	}
	if event.Assist2 > 0 {
		countPlayerEvent(event.Assist2, players, 0, 1, 0)
	}
}

func countPlayerEvent(playerNum int, playerMap map[int]PlayerSummary, goals int, assists int, minutes int) {
	player, ok := playerMap[playerNum]
	if !ok {
		player = *new(PlayerSummary)
	}

	player.Goals += goals
	player.Assists += assists
	player.Minutes += minutes

	playerMap[playerNum] = player
}

func SortEvents(game *Game) {
	sort.Slice(game.Events, func(i, j int) bool {
		return game.Events[i].GameTime < game.Events[j].GameTime
	})
}

func AddPlayer(game *Game, homeAway string, playerNum int, name string) {
	var team *map[int]string
	if homeAway == HOME {
		if game.HomePlayers == nil {
			game.HomePlayers = make(map[int]string)
		}
		team = &game.HomePlayers
	} else if homeAway == AWAY {
		if game.AwayPlayers == nil {
			game.AwayPlayers = make(map[int]string)
		}
		team = &game.AwayPlayers
	} else {
		return
	}
	(*team)[playerNum] = name
}
