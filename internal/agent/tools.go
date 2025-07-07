package agent

import (
	"context"
	"time"

	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/tool/gitclone"
	"github.com/cloudwego/eino-examples/quickstart/eino_assistant/pkg/tool/open"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	// duckduckgo "github.com/gitslim/eino-tool-duckduckgo-v2"
)

func GetTools(ctx context.Context, ragAgentConfig *RagAgentConfig) ([]tool.BaseTool, error) {
	tools := make([]tool.BaseTool, 0)

	// toolOpen, err := NewOpenFileTool(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// toolGitClone, err := NewGitCloneFile(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// tools = append(tools, toolOpen)
	// tools = append(tools, toolGitClone)

	if ragAgentConfig.ToolDuckduckgo {
		toolDDGSearch, err := NewDDGSearch(ctx, nil)
		if err != nil {
			return nil, err
		}
		toolDDGSearch.Info(ctx)
		tools = append(tools, toolDDGSearch)
	}

	if len(tools) == 0 {
		// add dummy tool because tools list cannot be empty
		tools = append(tools, DummyTool())
	}

	return tools, nil
}

// DummyTool - tool that do nothing
func DummyTool() tool.BaseTool {
	type Input struct{}
	type Result struct {
		// Msg string `json:"msg"`
	}

	fn := func(ctx context.Context, input *Input) (*Result, error) { return &Result{}, nil }

	return utils.NewTool(&schema.ToolInfo{
		Name: "dummy",
		Desc: "dummy_tool",
		// ParamsOneOf: schema.NewParamsOneOfByParams(
		// 	map[string]*schema.ParameterInfo{},
		// ),
	}, fn)
}

func defaultDDGSearchConfig(_ context.Context) (*duckduckgo.Config, error) {
	config := &duckduckgo.Config{
		Region:     duckduckgo.RegionRU,
		MaxResults: 5,
		Timeout:    30 * time.Second,
	}
	return config, nil
}

func NewDDGSearch(ctx context.Context, config *duckduckgo.Config) (tn tool.BaseTool, err error) {
	if config == nil {
		config, err = defaultDDGSearchConfig(ctx)
		if err != nil {
			return nil, err
		}
	}
	tn, err = duckduckgo.NewTextSearchTool(ctx, config)
	if err != nil {
		return nil, err
	}
	return tn, nil
}

func NewOpenFileTool(ctx context.Context) (tn tool.BaseTool, err error) {
	return open.NewOpenFileTool(ctx, nil)
}

func NewGitCloneFile(ctx context.Context) (tn tool.BaseTool, err error) {
	return gitclone.NewGitCloneFile(ctx, nil)
}
