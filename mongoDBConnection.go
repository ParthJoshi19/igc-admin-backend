package main

import (
	"context"
	"fmt"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func SetupMongoDB() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://pccoeigchack:Indradhanu407@cluster0.pg7et7j.mongodb.net/pccoe_IGC"))
	if err != nil {
		return nil, fmt.Errorf("MongoDB connect issue: %v", err)
	}
	
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("MongoDB ping issue: %v", err)
	}
	
	fmt.Println("Successfully connected to MongoDB!")
	return client, nil
}

// GetDatabase returns a specific database
func GetDatabase(client *mongo.Client, dbName string) *mongo.Database {
	return client.Database(dbName)
}

// GetCollection returns a specific collection from a database
func GetCollection(db *mongo.Database, collectionName string) *mongo.Collection {
	return db.Collection(collectionName)
}

// Close the connection
func CloseConnection(client *mongo.Client, context context.Context, cancel context.CancelFunc) {
 defer func() {
  cancel()
  if err := client.Disconnect(context); err != nil {
   panic(err)
  }
  fmt.Println("Close connection is called")
 }()
}