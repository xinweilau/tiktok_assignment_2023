package main

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	cli *redis.Client
}

type Message struct {
	Sender    string `json:"sender"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

// InitRedis initializes a Redis client with the given address and password.
// It returns an error if the client fails to connect to the Redis server.
func (c *RedisClient) InitRedisClient(ctx context.Context, address, password string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       0,
	})

	if _, err := client.Ping(ctx).Result(); err != nil {
		return err
	}

	c.cli = client

	return nil
}

// SaveMessage saves a message to Redis with the given chat ID.
// It marshals the message to JSON and adds it to a sorted set with the message timestamp as the score.
// It returns an error if the Redis operation fails.
func (c *RedisClient) SaveMessage(ctx context.Context, chatId string, message *Message) error {
	// Marshal the message to JSON.
	var (
		text []byte
		err  error
	)
	if text, err = json.Marshal(message); err != nil {
		return err
	}

	// Create a Redis sorted set member with the message timestamp as the score and the JSON-encoded message as the member.
	member := &redis.Z{
		Score:  float64(message.Timestamp),
		Member: text,
	}

	// Add the member to the sorted set with the given chat ID.
	if _, err = c.cli.ZAdd(ctx, chatId, *member).Result(); err != nil {
		return err
	}

	return nil
}

// GetMessagesByChatID retrieves messages from Redis with the given chat ID and timestamp range.
// It returns a slice of messages and an error if the Redis operation fails.
func (c *RedisClient) GetMessagesByChatID(ctx context.Context, chatID string, start, end int64, reverse bool) ([]*Message, error) {
	var (
		rawMessages []string
		messages    []*Message
		err         error
	)

	// Retrieve messages from Redis with the given chat ID and timestamp range.
	if reverse {
		// Retrieve messages in descending order with the message timestamp as the score.
		// The first message in the slice is the latest message.
		rawMessages, err = c.cli.ZRevRange(ctx, chatID, start, end).Result()
	} else {
		// Retrieve messages in ascending order with the message timestamp as the score.
		// The first message in the slice is the earliest message.
		rawMessages, err = c.cli.ZRange(ctx, chatID, start, end).Result()
	}

	// Return an error if the Redis operation fails.
	if err != nil {
		return nil, err
	}

	// Unmarshal each message from JSON and append it to the messages slice.
	for _, rawMessage := range rawMessages {
		message := &Message{}
		if err := json.Unmarshal([]byte(rawMessage), message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}
