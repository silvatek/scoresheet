package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

const GAMES_COLLECTION = "Games"

type FireDataStore struct {
	//Context  context.Context
	Client   *firestore.Client
	Project  string
	Database string
	Err      error
}

func fireDataStore() *FireDataStore {
	store := new(FireDataStore)
	//store.Context = context.Background()
	store.Project = os.Getenv("GCLOUD_PROJECT")
	store.Database = os.Getenv("FIRESTORE_DB_NAME")
	logs.info("Opening Firestore datastore %s, %s", store.Project, store.Database)
	return store
}

func (store *FireDataStore) summary() string {
	return fmt.Sprintf("ForeDataStore(%s,%s)", store.Project, store.Database)
}

func createClient(ctx context.Context, projectID string, database string) (*firestore.Client, error) {
	client, err := firestore.NewClientWithDatabase(ctx, projectID, database)
	if err == nil {
		logs.info("Firestore client created: %s %s", projectID, database)
	} else {
		logs.error("Failed to create FireStore client: %v", err)
	}
	// Close client when done with "defer client.Close()"
	return client, err
}

func (store *FireDataStore) getGame(ctx context.Context, id string) Game {
	logs.debug1(ctx, "Fetching Firestore game %s", id)

	var game Game

	doc := store.Client.Doc(GAMES_COLLECTION + "/" + id)
	gameDoc, err := doc.Get(ctx)
	if err != nil {
		logs.error1(ctx, "Error fetching game %s, %v", id, err)
	} else {
		logs.debug1(ctx, "Found game document %s", id)

		gameDoc.DataTo(&game)
	}

	return game
}

func (store *FireDataStore) putGame(ctx context.Context, id string, game Game) {
	logs.info1(ctx, "Writing Firestore game %s", id)

	doc := store.Client.Doc(GAMES_COLLECTION + "/" + id)
	_, err := doc.Set(ctx, game)
	if err != nil {
		logs.error1(ctx, "Error writing game %v", err)
	} else {
		logs.debug1(ctx, "Wrote game document %s", id)
	}
}

func (store *FireDataStore) addGame(ctx context.Context, game *Game) string {
	game.ID = randomId()
	store.putGame(ctx, game.ID, *game)
	return game.ID
}

func (store *FireDataStore) open() {
	store.Client, store.Err = createClient(context.Background(), store.Project, store.Database)
}

func (store *FireDataStore) close() {
	store.Client.Close()
}

func (store *FireDataStore) isEmpty() bool {
	games := store.Client.Collection(GAMES_COLLECTION)
	_, err := games.Documents(context.Background()).Next()
	return err == iterator.Done
}
