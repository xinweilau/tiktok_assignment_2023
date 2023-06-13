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
func (c *RedisClient) InitRedis(ctx context.Context, address, password string) error {
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
