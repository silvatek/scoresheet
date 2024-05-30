package main

import "testing"

func TestGameStruct(t *testing.T) {
	game := testGame1()

	if len(game.Events) != 4 {
		t.Errorf("Unexpected number of events in game: %d", len(game.Events))
	}
}

func TestParseEventTime(t *testing.T) {
	mins, secs := parseEventTime("12:34")

	if mins != 12 {
		t.Errorf("Mins was not 12: %d", mins)
	}

	if secs != 34 {
		t.Errorf("Secs was not 34: %d", secs)
	}
}

func TestClockToGameTime(t *testing.T) {
	assertGameTimeConversion(t, 1, "20:00", "00:00")
	assertGameTimeConversion(t, 1, "19:45", "00:15")
	assertGameTimeConversion(t, 1, "18:30", "01:30")
	assertGameTimeConversion(t, 1, "15:00", "05:00")
	assertGameTimeConversion(t, 1, "06:12", "13:48")
	assertGameTimeConversion(t, 1, "00:01", "19:59")
	assertGameTimeConversion(t, 2, "19:45", "20:15")
	assertGameTimeConversion(t, 2, "19:23", "20:37")
	assertGameTimeConversion(t, 3, "00:01", "59:59")
}

func TestGameToClockTime(t *testing.T) {
	assertClockTimeConversion(t, "00:00", 1, "20:00")
	assertClockTimeConversion(t, "00:01", 1, "19:59")
	assertClockTimeConversion(t, "15:00", 1, "05:00")
	assertClockTimeConversion(t, "20:37", 2, "19:23")
	assertClockTimeConversion(t, "23:49", 2, "16:11")
	assertClockTimeConversion(t, "25:52", 2, "14:08")
	assertClockTimeConversion(t, "30:21", 2, "09:39")
	assertClockTimeConversion(t, "36:27", 2, "03:33")
	assertClockTimeConversion(t, "38:15", 2, "01:45")
	assertClockTimeConversion(t, "39:35", 2, "00:25")
	assertClockTimeConversion(t, "48:02", 3, "11:58")
	assertClockTimeConversion(t, "54:51", 3, "05:09")
	assertClockTimeConversion(t, "55:00", 3, "05:00")
	assertClockTimeConversion(t, "59:59", 3, "00:01")
}

func assertGameTimeConversion(t *testing.T, period int, clockTime EventTime, expectedGameTime EventTime) {
	gameTime := ClockToGameTime(period, clockTime)

	if gameTime != expectedGameTime {
		t.Errorf("Game time expected %s, got %s", expectedGameTime, gameTime)
	}
}

func assertClockTimeConversion(t *testing.T, gameTime EventTime, expectedPeriod int, expectedClockTime EventTime) {
	clockTime, period := GameToClockTime(gameTime)

	if period != expectedPeriod {
		t.Errorf("Period expected %d, got %d", expectedPeriod, period)
	}

	if clockTime != expectedClockTime {
		t.Errorf("Clock time expected %s, got %s", expectedClockTime, clockTime)
	}
}

func TestAddGoal(t *testing.T) {
	game := Game{
		Period: 1,
		Events: make([]Event, 0),
	}

	AddGoal(&game, 1, "15:00", HOME, 25, 12, 95, "Even")

	if len(game.Events) != 1 {
		t.Errorf("Unexpected updated number of events in game: %d", len(game.Events))
	}

	if game.Events[0].GameTime != "05:00" {
		t.Errorf("Unexpected game time value: %s", game.Events[0].GameTime)
	}
}

func TestAddPenalty(t *testing.T) {
	game := Game{
		Period: 2,
		Events: make([]Event, 0),
	}

	AddPenalty(&game, 2, "12:30", AWAY, 7, 2, "Slash")

	if len(game.Events) != 1 {
		t.Errorf("Unexpected updated number of events in game: %d", len(game.Events))
	}

	if game.Events[0].GameTime != "27:30" {
		t.Errorf("Unexpected game time value: %s", game.Events[0].GameTime)
	}
}

func TestGameSummary(t *testing.T) {
	game := testGame1()
	summary := summarise(game)

	if len(summary.HomePlayers) != 3 {
		t.Errorf("Unexpected home player count: %d", len(summary.HomePlayers))
	}
	ps, ok := summary.HomePlayers[41]
	if !ok {
		t.Error("No summary for home player 41 found")
	}
	if ps.Goals != 1 {
		t.Errorf("Unexpected home player 41 goal count: %d", ps.Goals)
	}
}
