package mem

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSimpleMemory tests create, get, list and delete conversations
func TestSimpleMemory(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test_memory")
	defer os.RemoveAll(testDir)

	cfg := SimpleMemoryConfig{
		Dir:           testDir,
		MaxWindowSize: 3,
	}
	mem, err := CreateConfiguredSimpleMemory(cfg)
	require.NoError(t, err)

	t.Run("GetConversation - new conversation", func(t *testing.T) {
		conv := mem.GetConversation("conv1", true)
		assert.NotNil(t, conv)
		assert.Equal(t, "conv1", conv.ID)
		assert.Empty(t, conv.Messages)
	})

	t.Run("GetConversation - existing conversation", func(t *testing.T) {
		conv1 := mem.GetConversation("conv2", true)
		require.NotNil(t, conv1)
		conv1.Append(&schema.Message{Content: "test"})

		conv2 := mem.GetConversation("conv2", false)
		assert.NotNil(t, conv2)
		assert.Equal(t, 1, len(conv2.Messages))
		assert.Equal(t, "test", conv2.Messages[0].Content)
	})

	t.Run("ListConversations", func(t *testing.T) {
		mem.GetConversation("conv3", true)
		mem.GetConversation("conv4", true)

		ids := mem.ListConversations()
		assert.Contains(t, ids, "conv3")
		assert.Contains(t, ids, "conv4")
	})

	t.Run("DeleteConversation", func(t *testing.T) {
		conv := mem.GetConversation("conv5", true)
		require.NotNil(t, conv)

		err := mem.DeleteConversation("conv5")
		assert.NoError(t, err)

		ids := mem.ListConversations()
		assert.NotContains(t, ids, "conv5")
	})
}

// TestConversation tests the Conversation struct functions
func TestConversation(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test_conversations")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	file, err := os.Create(filepath.Join(testDir, "test_conv.jsonl"))
	require.NoError(t, err)
	file.Close()
	fmt.Printf("filename: %s\n", file.Name())

	conv := &Conversation{
		ID:            "test_conv",
		Messages:      []*schema.Message{},
		filePath:      file.Name(),
		maxWindowSize: 2,
	}

	t.Run("Append and GetMessages", func(t *testing.T) {
		conv.Append(&schema.Message{Content: "msg1"})
		conv.Append(&schema.Message{Content: "msg2"})
		conv.Append(&schema.Message{Content: "msg3"})

		messages := conv.GetMessages()
		assert.Equal(t, 2, len(messages))
		assert.Equal(t, "msg2", messages[0].Content)
		assert.Equal(t, "msg3", messages[1].Content)
	})

	t.Run("GetFullMessages", func(t *testing.T) {
		messages := conv.GetFullMessages()
		assert.Equal(t, 3, len(messages))
	})

	t.Run("Load and Save", func(t *testing.T) {
		file, err := os.Create(filepath.Join(testDir, "test_conv.jsonl"))
		require.NoError(t, err)
		file.Close()

		maxWindowSize := 2
		newConv := &Conversation{
			ID:            "test_load",
			Messages:      []*schema.Message{},
			filePath:      file.Name(),
			maxWindowSize: maxWindowSize,
		}

		newConv.Append(&schema.Message{Content: "saved1"})
		newConv.Append(&schema.Message{Content: "saved2"})

		loadedConv := &Conversation{
			ID:            "test_load",
			Messages:      []*schema.Message{},
			filePath:      file.Name(),
			maxWindowSize: maxWindowSize,
		}
		err = loadedConv.load()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(loadedConv.Messages))
		assert.Equal(t, "saved1", loadedConv.Messages[0].Content)
	})
}

// TestEdgeCases tests some edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Create with invalid dir", func(t *testing.T) {
		cfg := SimpleMemoryConfig{
			Dir: "/dev/null",
		}
		_, err := CreateConfiguredSimpleMemory(cfg)
		assert.Error(t, err)
	})

	t.Run("Get non-existent conversation with create", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "test_create")
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		mem, err := CreateConfiguredSimpleMemory(SimpleMemoryConfig{Dir: testDir})
		assert.NoError(t, err)
		conv := mem.GetConversation("nonexistent", true)
		assert.Empty(t, conv.Messages)
	})

	t.Run("Get non-existent conversation without create", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "test_create")
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		mem, err := CreateConfiguredSimpleMemory(SimpleMemoryConfig{Dir: testDir})
		assert.NoError(t, err)
		conv := mem.GetConversation("nonexistent", false)
		assert.Empty(t, conv.Messages)
	})

	t.Run("Delete non-existent conversation", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "test_delete")
		os.MkdirAll(testDir, 0755)
		defer os.RemoveAll(testDir)

		mem, err := CreateConfiguredSimpleMemory(SimpleMemoryConfig{Dir: testDir})
		assert.NoError(t, err)
		err = mem.DeleteConversation("nonexistent")
		assert.Error(t, err)
	})
}
