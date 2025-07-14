package types

import "github.com/cloudwego/eino/schema"

// UserMessage is the user message for RAG
type UserMessage struct {
	ID      string            `json:"id"`
	Query   string            `json:"query"`
	History []*schema.Message `json:"history"`
}
