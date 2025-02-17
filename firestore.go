package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type FireDataStore struct {
	Client   *firestore.Client
	Project  string
	Database string
	Err      error
}

func fireDataStore() *FireDataStore {
	store := new(FireDataStore)
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

func (store *FireDataStore) Put(ctx context.Context, collection string, id string, item interface{}) {
	logs.info1(ctx, "Writing Firestore %s %s", collection, id)

	doc := store.Client.Doc(collection + "/" + id)
	_, err := doc.Set(ctx, item)
	if err != nil {
		logs.error1(ctx, "Error writing item %v", err)
	} else {
		logs.debug1(ctx, "Wrote iten %s", id)
	}
}

func (store *FireDataStore) Get(ctx context.Context, collection string, id string, item interface{}) interface{} {
	logs.debug1(ctx, "Fetching Firestore %s %s", collection, id)

	doc := store.Client.Doc(collection + "/" + id)
	data, err := doc.Get(ctx)
	if err != nil {
		logs.error1(ctx, "Error fetching game %s, %v", id, err)
	} else {
		logs.debug1(ctx, "Found game document %s", id)

		data.DataTo(&item)
	}

	return item
}
