package connectionhelper

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var clientInstance *mongo.Client
var clientInstanceError error
var mongoOnce sync.Once

func GetMongoClient() (*mongo.Client, error) {

	mongoOnce.Do(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			client, err := mongo.Connect(ctx, options.Client().ApplyURI(
				"",
			))
			if err != nil { log.Fatal(err) }
			err = client.Ping( context.TODO(),nil )
			if err != nil {
				clientInstanceError = err
			} else {
				fmt.Println("Connected to MongoDB.....")
			}
			
			clientInstance = client
		})
	
	return clientInstance,clientInstanceError
}

