package mem

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
)

// SimpleMemory is a simple in-memory memory implementation. It stores conversations in a directory.
type SimpleMemory struct {
	mu            sync.Mutex
	dir           string
	maxWindowSize int
	conversations map[string]*Conversation
}

// SimpleMemoryConfig is the configuration for SimpleMemory
type SimpleMemoryConfig struct {
	Dir           string
	MaxWindowSize int
}

// NewSimpleMemory creates a new SimpleMemory instance
func NewSimpleMemory() (*SimpleMemory, error) {
	return CreateConfiguredSimpleMemory(SimpleMemoryConfig{
		Dir:           "data/memory",
		MaxWindowSize: 6,
	})
}

// CreateConfiguredSimpleMemory creates a new SimpleMemory instance with the given configuration
func CreateConfiguredSimpleMemory(cfg SimpleMemoryConfig) (*SimpleMemory, error) {
	if cfg.Dir == "" {
		cfg.Dir = filepath.Join(os.TempDir(), "eino", "memory")
	}
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create memory directory: %w", err)
	}

	return &SimpleMemory{
		dir:           cfg.Dir,
		maxWindowSize: cfg.MaxWindowSize,
		conversations: make(map[string]*Conversation),
	}, nil
}

// GetConversation returns the conversation with the given ID. If the conversation does not exist and createIfNotExist is true, a new conversation is created.
func (m *SimpleMemory) GetConversation(id string, createIfNotExist bool) *Conversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.conversations[id]

	filePath := filepath.Join(m.dir, id+".jsonl")
	if !ok {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if createIfNotExist {
				if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
					return nil
				}
				m.conversations[id] = &Conversation{
					ID:            id,
					Messages:      make([]*schema.Message, 0),
					filePath:      filePath,
					maxWindowSize: m.maxWindowSize,
				}
			}
		}

		con := &Conversation{
			ID:            id,
			Messages:      make([]*schema.Message, 0),
			filePath:      filePath,
			maxWindowSize: m.maxWindowSize,
		}
		con.load()
		m.conversations[id] = con
	}

	return m.conversations[id]
}

// ListConversations returns a list of conversation IDs.
func (m *SimpleMemory) ListConversations() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	files, err := os.ReadDir(m.dir)
	if err != nil {
		return nil
	}

	ids := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ids = append(ids, strings.TrimSuffix(file.Name(), ".jsonl"))
	}

	return ids
}

// DeleteConversation deletes a conversation from the memory.
func (m *SimpleMemory) DeleteConversation(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filePath := filepath.Join(m.dir, id+".jsonl")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	delete(m.conversations, id)
	return nil
}

// Conversation is a conversation in the memory.
type Conversation struct {
	mu sync.Mutex

	ID       string            `json:"id"`
	Messages []*schema.Message `json:"messages"`

	filePath string

	maxWindowSize int
}

// Append appends a message to the conversation.
func (c *Conversation) Append(msg *schema.Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Messages = append(c.Messages, msg)

	c.save(msg)
}

// GetFullMessages returns all messages in the conversation.
func (c *Conversation) GetFullMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.Messages
}

// GetMessages returns the last messages in the conversation.
func (c *Conversation) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.Messages) > c.maxWindowSize {
		return c.Messages[len(c.Messages)-c.maxWindowSize:]
	}

	return c.Messages
}

// load loads the conversation from the file.
func (c *Conversation) load() error {
	reader, err := os.Open(c.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		var msg schema.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return fmt.Errorf("failed to unmarshal message: %w", err)
		}
		c.Messages = append(c.Messages, &msg)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// save saves a message to the conversation file
func (c *Conversation) save(msg *schema.Message) {
	str, _ := json.Marshal(msg)

	// Append to file
	f, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(str)
	f.WriteString("\n")
}
