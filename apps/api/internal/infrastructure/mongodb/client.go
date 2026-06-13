package mongodb

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(ctx context.Context) (*mongo.Client, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	opts := options.Client().ApplyURI(uri).SetConnectTimeout(10 * time.Second)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}

	return client, nil
}

func Database(client *mongo.Client) *mongo.Database {
	name := os.Getenv("MONGODB_DB")
	if name == "" {
		name = "cms"
	}
	return client.Database(name)
}
