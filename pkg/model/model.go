package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
)

var timeout = 10 * time.Second

type UserMessage struct {
	UserName    string    `bson:"user_name,omitempty"`
	UserID      string    `bson:"user_id,omitempty"`
	MessageText string    `bson:"message_text,omitempty"`
	Timestamp   time.Time `bson:"timestamp,omitempty"`
}

func (msg *UserMessage) Save() error {
	coll := client.Database("linebot").Collection("messages")

	_, err := coll.InsertOne(context.Background(), msg)
	if err != nil {
		return fmt.Errorf("failed of saving message to MongoDB: %w", err)
	}

	return nil
}

func GetHistory(userID string) ([]UserMessage, error) {
	filter := bson.D{primitive.E{Key: "user_id", Value: userID}}

	coll := client.Database("linebot").Collection("messages")
	cursor, err := coll.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed of query messages: %w", err)
	}

	var msgs []UserMessage
	if err = cursor.All(context.Background(), &msgs); err != nil {
		return nil, fmt.Errorf("failed of fetching and decoding messages: %w", err)
	}

	return msgs, nil
}

func FormatHistory(msgs []UserMessage) string {
	tz, _ := time.LoadLocation("Local")
	var b strings.Builder
	b.Grow(200)
	b.WriteString("Your history:\n")
	for _, msg := range msgs {
		fmt.Fprintf(&b, "* %s | %s\n", msg.Timestamp.In(tz).Format("2006/01/02 15:04:05 MST"), msg.MessageText)
	}

	return b.String()
}

func InitClient(uri, username, password string) error {
	var err error

	credential := options.Credential{
		Username: username,
		Password: password,
	}

	clientOpts := options.Client().ApplyURI(uri).SetAuth(credential)

	ctx, cancel = context.WithTimeout(context.Background(), timeout)

	client, err = mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed of connect the MongoDB: %w", err)
	}

	return nil
}

func Disconnect() error {
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func TestConnection() error {
	return client.Ping(ctx, readpref.Primary())
}
