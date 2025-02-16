package main

import (
	"os"

	"github.com/labstack/echo/v4"
)

const defaultAddr = "localhost:8080"

var dataStore GameStore
var logs Logger

// main starts an http server on the $PORT environment variable.
func main() {
	logs.init()

	dataStore = createDataStore()
	dataStore.open()
	defer dataStore.close()
	setupDataStore(dataStore)

	addr := defaultAddr
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	logs.info("Server listening on port %s", addr)

	e := echo.New()
	e.HideBanner = true

	addRoutes(e)

	e.Start(addr)
}

func runningOnGCloud() bool {
	gCloudServiceName := os.Getenv("K_SERVICE")
	return len(gCloudServiceName) > 0
}

func createDataStore() GameStore {
	var store GameStore
	if runningOnGCloud() {
		store.datastore = fireDataStore()
	} else {
		store.datastore = testDataStore()
	}
	return store
}

func setupDataStore(ds GameStore) {
	if ds.isEmpty() {
		addTestGames(ds)
	} else {
		logs.info("Datastore is not empty so not adding test games")
	}
}
