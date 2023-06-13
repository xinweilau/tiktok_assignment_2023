package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc"
)

// IMServiceImpl implements the last service interface defined in the IDL.
type IMServiceImpl struct{}

func (s *IMServiceImpl) Send(ctx context.Context, req *rpc.SendRequest) (*rpc.SendResponse, error) {
	// Split the chat ID into the two users
	users := strings.Split(req.Message.Chat, ":")
	if len(users) != 2 {
		err := fmt.Errorf("Invalid Chat ID format: '%s'. Chat ID should be in the format of 'user1:user2'", req.Message.GetChat())
		return nil, err
	}

	// Check if the sender is a member of the chat room
	user1, user2 := users[0], users[1]
	if req.Message.GetSender() != user1 && req.Message.GetSender() != user2 {
		err := fmt.Errorf("User '%s' is not a member of the chat room '%s'", req.Message.GetSender(), req.Message.GetChat())
		return nil, err
	}

	// Get the current timestamp and create a new message object
	timestamp := time.Now().Unix()
	message := &Message{
		Message:   req.Message.GetText(),
		Sender:    req.Message.GetSender(),
		Timestamp: timestamp,
	}

	// Get the chat ID from the request and save the message to Redis
	chatID, err := getChatID(req.Message.GetChat())
	if err != nil {
		return nil, err
	}
	if err := rdb.SaveMessage(ctx, chatID, message); err != nil {
		return nil, err
	}

	// Create the response object
	resp := rpc.NewSendResponse()
	resp.Code = 0
	resp.Msg = "success"

	return resp, nil
}

func (s *IMServiceImpl) Pull(ctx context.Context, req *rpc.PullRequest) (*rpc.PullResponse, error) {
	// Get the chat ID from the request
	chatID, err := getChatID(req.GetChat())
	if err != nil {
		return nil, err
	}

	// Calculate the start and end indices for the messages to retrieve
	start := req.GetCursor()
	end := start + int64(req.GetLimit())

	// Retrieve the messages from Redis
	messages, err := rdb.GetMessagesByChatID(ctx, chatID, start, end, req.GetReverse())
	if err != nil {
		return nil, err
	}

	// Initialize variables for the response
	respMessages := make([]*rpc.Message, 0)
	hasMore := false
	counter := int32(0)
	nextCursor := int64(0)

	// Iterate over the retrieved messages and add them to the response
	for _, msg := range messages {
		if counter+1 >= req.GetLimit() {
			// If we've reached the limit, set hasMore to true and break out of the loop
			hasMore = true
			nextCursor = end
			break
		}

		respMessages = append(respMessages, &rpc.Message{
			Chat:     req.GetChat(),
			Text:     msg.Message,
			Sender:   msg.Sender,
			SendTime: msg.Timestamp,
		})
		counter++
	}

	// Create the response object
	resp := rpc.NewPullResponse()
	resp.Messages = respMessages
	resp.Code = 0
	resp.Msg = "success"
	resp.HasMore = &hasMore
	resp.NextCursor = &nextCursor

	return resp, nil
}

// getChatID takes a chat ID and returns the corresponding chat ID found in Redis.
// The sender and receiver are sorted alphabetically to form the chat ID.
func getChatID(id string) (string, error) {
	lowercase := strings.ToLower(id)

	// Convert the chat to lowercase and split it into the two users
	users := strings.Split(lowercase, ":")
	if len(users) != 2 {
		// Return an error if the chat ID is not in the correct format
		err := fmt.Errorf("Invalid Chat ID format: '%s'. Chat ID should be in the format of 'user1:user2'", id)
		return "", err
	}

	// Sort the users alphabetically and form the room ID
	var chatID string
	user1, user2 := users[0], users[1]
	if comp := strings.Compare(user1, user2); comp == 1 {
		chatID = fmt.Sprintf("%s:%s", user2, user1)
	} else {
		chatID = fmt.Sprintf("%s:%s", user1, user2)
	}

	return chatID, nil
}
