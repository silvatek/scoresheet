# Ice Hockey Scoresheet Web Application

"Scoresheet" is a Google Cloud Run application that contains allows the user to record the scoresheet details of an Ice Hockey game.

## Commands

`go test -coverprofile=cover.out`
`go tool cover -html=cover.out`

## Tracking

### To-Do

* Cache parsed templates

### Done
* Write a doc into the Games collection from the fire page
* Write test games into firestore if the collection is empty
* Read the game from firestore and show on the page
* Web interface to add a new goal to a game
* Web interface to add a new penalty to a game
* Derive scoresheet summary from game events
* Teams and date on game
* Sort events by game time
* Web interface to start a new game
* Delete event
* Edit keys for games

