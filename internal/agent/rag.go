/*
Example eino agent graph project: https://github.com/cloudwego/eino-examples/blob/main/quickstart/eino_assistant/
*/

package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/gitslim/go-ragger/internal/mem"
	"github.com/gitslim/go-ragger/internal/types"
	"github.com/gitslim/go-ragger/internal/vectordb/milvus"
)

// RAGAgent is a RAG agent that uses a chat model to generate a response to a user message.
type RAGAgent struct {
	runner compose.Runnable[*types.UserMessage, *schema.Message]
	mem    *mem.SimpleMemory
	cb     callbacks.Handler
}

// RagAgentConfig is the configuration for a RAG agent.
type RagAgentConfig struct {
	UserID         string
	MaxSteps       int
	ToolDuckduckgo bool
}

// RagAgentFactory is a factory for a RAG agent.
type RagAgentFactory func(ctx context.Context, config *RagAgentConfig) (*RAGAgent, error)

// NewRAGAgentFactory creates a new RAG agent factory.
func NewRAGAgentFactory(logger *slog.Logger, mem *mem.SimpleMemory, retrieverFactory milvus.MilvusRetrieverFactory, chatTemplate prompt.ChatTemplate, model model.ToolCallingChatModel) RagAgentFactory {
	const (
		NodeInputToQuery   = "InputToQuery"
		NodeChatTemplate   = "ChatTemplate"
		NodeReactAgent     = "ReactAgent"
		NodeRetriever      = "Retriever"
		NodeInputToHistory = "InputToHistory"
	)

	return func(ctx context.Context, config *RagAgentConfig) (*RAGAgent, error) {
		g := compose.NewGraph[*types.UserMessage, *schema.Message]()
		_ = g.AddLambdaNode(NodeInputToQuery, compose.InvokableLambdaWithOption(userMessageToQueryLambda), compose.WithNodeName("UserMessageToQuery"))
		_ = g.AddChatTemplateNode(NodeChatTemplate, chatTemplate)
		reactAgentLambda, err := reactAgentLambda(ctx, &model, config)
		if err != nil {
			return nil, err
		}
		retriever, err := retrieverFactory(ctx, &milvus.MilvusRetrieverConfig{Collection: milvus.ToMilvusName(config.UserID)})
		if err != nil {
			return nil, fmt.Errorf("failed to create retriever: %w", err)
		}
		_ = g.AddLambdaNode(NodeReactAgent, reactAgentLambda, compose.WithNodeName("ReAct Agent"))
		_ = g.AddRetrieverNode(NodeRetriever, retriever, compose.WithOutputKey("documents"))
		_ = g.AddLambdaNode(NodeInputToHistory, compose.InvokableLambdaWithOption(userMessageToVariablesLambda), compose.WithNodeName("UserMessageToVariables"))
		_ = g.AddEdge(compose.START, NodeInputToQuery)
		_ = g.AddEdge(compose.START, NodeInputToHistory)
		_ = g.AddEdge(NodeReactAgent, compose.END)
		_ = g.AddEdge(NodeInputToQuery, NodeRetriever)
		_ = g.AddEdge(NodeRetriever, NodeChatTemplate)
		_ = g.AddEdge(NodeInputToHistory, NodeChatTemplate)
		_ = g.AddEdge(NodeChatTemplate, NodeReactAgent)

		runner, err := g.Compile(ctx, compose.WithGraphName("RAGAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
		if err != nil {
			return nil, err
		}

		cbConfig := &LogCallbackConfig{
			Detail: true,
			Writer: os.Stderr,
			Debug:  true,
		}

		agent := &RAGAgent{
			runner: runner,
			mem:    mem,
			cb:     LogCallback(cbConfig),
		}

		return agent, nil
	}
}

// userMessageToQueryLambda is a lambda function that takes a user message and returns a query string.
func userMessageToQueryLambda(ctx context.Context, input *types.UserMessage, opts ...any) (output string, err error) {
	return input.Query, nil
}

// reactAgentLambda is a lambda function that takes a model and a rag agent config and returns a react agent.
func reactAgentLambda(ctx context.Context, model *model.ToolCallingChatModel, ragAgentConfig *RagAgentConfig) (lambda *compose.Lambda, err error) {
	config := &react.AgentConfig{
		MaxStep: ragAgentConfig.MaxSteps,
		// ToolReturnDirectly: map[string]struct{}{}
	}

	config.ToolCallingModel = *model

	tools, err := GetTools(ctx, ragAgentConfig)
	if err != nil {
		return nil, err
	}

	config.ToolsConfig.Tools = tools

	agent, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	lambda, err = compose.AnyLambda(agent.Generate, agent.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lambda, nil
}

// userMessageToVariablesLambda is a lambda function that takes a user message and returns a map of variables
func userMessageToVariablesLambda(ctx context.Context, input *types.UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"content": input.Query,
		"history": input.History,
		"date":    time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// Run is a function that run agent and return stream reader
func (agent *RAGAgent) Run(ctx context.Context, id string, msg string) (*schema.StreamReader[*schema.Message], error) {

	conversation := agent.mem.GetConversation(id, true)

	userMessage := &types.UserMessage{
		ID:      id,
		Query:   msg,
		History: conversation.GetMessages(),
	}

	sr, err := agent.runner.Stream(ctx, userMessage, compose.WithCallbacks(agent.cb))
	if err != nil {
		return nil, fmt.Errorf("failed to stream: %w", err)
	}

	srs := sr.Copy(2)

	go func() {
		// for save to memory
		fullMsgs := make([]*schema.Message, 0)

		defer func() {
			// close stream if you used it
			srs[1].Close()

			// add user input to history
			conversation.Append(schema.UserMessage(msg))

			fullMsg, err := schema.ConcatMessages(fullMsgs)
			if err != nil {
				fmt.Println("error concatenating messages: ", err.Error())
			}
			// add agent response to history
			conversation.Append(fullMsg)
		}()

	outer:
		for {
			select {
			case <-ctx.Done():
				fmt.Println("context done", ctx.Err())
				return
			default:
				chunk, err := srs[1].Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break outer
					}
				}

				fullMsgs = append(fullMsgs, chunk)
			}
		}
	}()

	return srs[0], nil
}

// LogCallbackConfig is the configuration for the log callback.
type LogCallbackConfig struct {
	Detail bool
	Debug  bool
	Writer io.Writer
}

// LogCallback is a callback that logs the input and output of the component.
func LogCallback(config *LogCallbackConfig) callbacks.Handler {
	if config == nil {
		config = &LogCallbackConfig{
			Detail: true,
			Writer: os.Stdout,
		}
	}
	if config.Writer == nil {
		config.Writer = os.Stdout
	}
	builder := callbacks.NewHandlerBuilder()
	builder.OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
		fmt.Fprintf(config.Writer, "[view]: start [%s:%s:%s]\n", info.Component, info.Type, info.Name)
		if config.Detail {
			var b []byte
			if config.Debug {
				b, _ = json.MarshalIndent(input, "", "  ")
			} else {
				b, _ = json.Marshal(input)
			}
			fmt.Fprintf(config.Writer, "%s\n", string(b))
		}
		return ctx
	})
	builder.OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		fmt.Fprintf(config.Writer, "[view]: end [%s:%s:%s]\n", info.Component, info.Type, info.Name)
		return ctx
	})
	return builder.Build()
}
