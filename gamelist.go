package main

type GameList struct {
	Name       string
	Games      []string
	LockedWith string
}

func NewGameList(name string) GameList {
	var list GameList
	list.Name = name
	list.Games = make([]string, 0)
	return list
}

func (list *GameList) AddGame(gameId string) {
	list.Games = append(list.Games, gameId)
}

func (list GameList) IsLocked() bool {
	return list.LockedWith != ""
}
