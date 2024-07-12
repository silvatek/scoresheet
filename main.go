package main

import (
	//"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

const defaultAddr = "localhost:8080"

var dataStore DataStore
var logs Logger

// main starts an http server on the $PORT environment variable.
func main() {
	logs.init()

	dataStore = createDataStore()
	dataStore.open()
	defer dataStore.close()
	setupDataStore(dataStore)

	addr := defaultAddr
	// $PORT environment variable is provided in the Kubernetes deployment.
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	logs.info("Server listening on port %s", addr)

	e := echo.New()

	addRoutes(e)

	e.Start(addr)
	// r := addHandlers()

	// if err := http.ListenAndServe(addr, r); err != nil {
	// 	logs.error("Server listening error: %+v", err)
	// 	os.Exit(-5)
	// }

}

func runningOnGCloud() bool {
	gCloudServiceName := os.Getenv("K_SERVICE")
	return len(gCloudServiceName) > 0
}

func createDataStore() DataStore {
	if runningOnGCloud() {
		return fireDataStore()
	} else {
		return testDataStore()
	}
}

func setupDataStore(ds DataStore) {
	if ds.isEmpty() {
		addTestGames(ds)
	} else {
		logs.info("Datastore is not empty so not adding test games")
	}
}
