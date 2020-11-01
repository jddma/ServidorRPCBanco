package server

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type Database struct {
	connection *mongo.Client
}

func (d *Database) openConnection()  {

	mgdUser := os.Getenv("MGD_USER")
	mgdPassword := os.Getenv("MGD_PASSWORD")
	mgdHost := os.Getenv("MGD_HOST")

	uri := "mongodb://" + mgdUser + ":" + mgdPassword + "@" + mgdHost + ":27017"

	var err error
	d.connection, err = mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = d.connection.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
