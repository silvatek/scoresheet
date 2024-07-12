# Ice Hockey Scoresheet Web Application

"Scoresheet" is a Google Cloud Run application that  allows the user to record the scoresheet details of ice hockey games.

It does automatic conversion of clock time to game time, and calculates the game totals such as how many goals were scored by each player.

The site is designed to operate without logging in, so it is not possible to get a list of games; the user needs to know the game ID. Games can also be locked to prevent anyone from editing them without knowing the edit code.

## Code structure

`main.go` is the primary entry point, but delegates most of the work to the http request handlers in `handlers.go`. 

The primary game logic is in `game.go`, backed by a "datastore" interface defined in `datastore.go` which includes a test datastore implementation. 

A "real" datastore built using Google Cloud Platform's Firestore is implemented in `firestore.go`, and support for Google Cloud logging (with fallback to console if not running on GCP) is in `logging.go`.

The templates for html pages are in `templates` and static content (stylesheet, images, etc) is in 
`templates/static`.

## Commands
Run tests and show coverage...

```go test -coverprofile cover.out```

```go tool cover -html cover.out```

## Tracking

### To-Do

* Cache parsed templates

### Done
* Migrate to echo web framework
* Store game history in a cookie and allow user to go back to recent ones
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
* Site logo
* Share game via QR code

