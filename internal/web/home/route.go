package home

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/cloudwego/eino/schema"
	"github.com/davecgh/go-spew/spew"
	"github.com/gitslim/go-ragger/internal/agent"
	"github.com/gitslim/go-ragger/internal/config"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/gitslim/go-ragger/internal/vectordb/milvus"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"

	datastar "github.com/starfederation/datastar/sdk/go"
)

var md = goldmark.New(
	goldmark.WithRendererOptions(
		html.WithXHTML(),
		html.WithUnsafe(),
	),
)

func SetupRoutes(rtr chi.Router, logger *slog.Logger, config *config.ServerConfig, q *sqlc.Queries, retrieverFactory milvus.MilvusRetrieverFactory, agentFacrory agent.RagAgentFactory) {

	rtr.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := util.UserFromContext(ctx)

		s := &Signals{Question: "", ShowThink: true, Duckduckgo: false}

		docCounts, err := q.GetDocumentsStatusCounts(ctx, user.ID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		status := &Status{
			Ready:      docCounts.ReadyCount,
			Processing: docCounts.ProcessingCount,
			Failed:     docCounts.FailedCount,
		}

		PageHome(r, user, s, config, status).Render(ctx, w)
	})

	rtr.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := util.UserFromContext(ctx)
		if user == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		docCounts, err := q.GetDocumentsStatusCounts(ctx, user.ID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		status := &Status{
			Ready:      docCounts.ReadyCount,
			Processing: docCounts.ProcessingCount,
			Failed:     docCounts.FailedCount,
		}

		sse := datastar.NewSSE(w, r)
		sse.MergeFragmentTempl(
			NavbarStatus(r, status),
			datastar.WithSelectorID("navbar-status"),
			datastar.WithMergeMode(datastar.FragmentMergeModeInner))
	})

	rtr.Post("/ask", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, ok := util.UserFromContext(ctx)
		if user == nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		reqID, ok := util.RequestIDFromContext(ctx)
		if !ok {
			reqID = uuid.New().String()
		}

		signals := &Signals{}
		if err := datastar.ReadSignals(r, &signals); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		questionID := fmt.Sprintf("question-%s", reqID)
		answerID := fmt.Sprintf("answer-%s", reqID)
		question := signals.Question

		sse := datastar.NewSSE(w, r)

		signals.Question = ""
		sse.MarshalAndMergeSignals(signals)

		// append question
		appendElement(sse, "chat", UserRequest(questionID, question))

		// append answer element
		appendElement(sse, "chat", AssistantResponse(
			answerID,
			Response{
				Stage:   ResponseStarting,
				Content: "Обработка запроса...",
			}))

		// create agent for current user with config
		agent, err := agentFacrory(ctx, &agent.RagAgentConfig{
			UserID:         user.ID.String(),
			MaxSteps:       25,
			ToolDuckduckgo: signals.Duckduckgo})
		if err != nil {
			logger.Error("failed to create agent with factory", "error", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}

		// run agent
		sr, err := agent.Run(ctx, reqID, question)
		if err != nil {
			logger.Error("failed to run rag agent", "error", err)
			appendElement(sse, "chat", AssistantResponse(
				answerID,

				Response{
					Stage: ResponseFinished,
					Error: "Сбой системы! Повторите попытку!"}))
			return
		}

		errorMsg, mainContent, reasonContent := "", "", ""
		var builder strings.Builder
		builder.Grow(1000)

		// streamreader loop
		for {
			msg, err := sr.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				logger.Error("failed to recv from rag agent stream: %v\n", "error", err)
				errorMsg = "Ошибка! ответ сгенерирован не полностью!"
			}

			// spew.Dump(msg)

			content := msg.Content
			if len(msg.MultiContent) > 0 {
				content = getMultiContent(msg.MultiContent)
			}

			builder.WriteString(content)
			reasonContent, mainContent = parseStreamedContent(builder.String())

			mainContent, err = renderMarkdown(mainContent)
			if err != nil {
				mainContent = ""
			}
			reasonContent, err = renderMarkdown(reasonContent)
			if err != nil {
				reasonContent = ""
			}

			replaceElement(sse, answerID, AssistantResponse(
				answerID,
				Response{
					Stage:         ResponseStreaming,
					Content:       mainContent,
					ReasonContent: reasonContent}))

		}
		// was an error
		if errorMsg != "" {
			replaceElement(sse, answerID, AssistantResponse(
				answerID,
				Response{
					Stage:         ResponseFinished,
					Content:       mainContent,
					ReasonContent: reasonContent,
					Error:         errorMsg}))
		} else {
			// finished normally
			replaceElement(sse, answerID, AssistantResponse(
				answerID,
				Response{
					Stage:         ResponseFinished,
					Content:       mainContent,
					ReasonContent: reasonContent}))
		}
		spew.Dump(mainContent, reasonContent)
	})
}

func parseStreamedContent(content string) (reasonContent, mainContent string) {
	re := regexp.MustCompile(`(?s)<think>(.*?)(</think>|$)`)
	matches := re.FindStringSubmatchIndex(content)

	if len(matches) >= 4 {
		start := matches[2]
		end := matches[3]

		reasonContent = strings.TrimSpace(content[start:end])

		if len(matches) >= 6 && matches[4] != -1 {
			mainStart := matches[1]
			mainContent = strings.TrimSpace(content[mainStart:])
		} else {
			mainContent = ""
		}
	} else {
		mainContent = strings.TrimSpace(content)
	}

	return reasonContent, mainContent
}

func appendElement(sse *datastar.ServerSentEventGenerator, elementID string, tpl templ.Component) {
	sse.MergeFragmentTempl(tpl,
		datastar.WithSelectorID(elementID),
		datastar.WithMergeMode(datastar.FragmentMergeModeAppend))
}

func replaceElement(sse *datastar.ServerSentEventGenerator, elementID string, tpl templ.Component) {
	sse.MergeFragmentTempl(tpl,
		datastar.WithSelectorID(elementID),
		datastar.WithMergeMode(datastar.FragmentMergeModeOuter))
}

func getMultiContent(parts []schema.ChatMessagePart) string {
	var builder strings.Builder
	builder.Grow(300)
	for _, part := range parts {
		var s string
		switch part.Type {
		case schema.ChatMessagePartTypeText:
			s = part.Text
		case schema.ChatMessagePartTypeImageURL:
			s = fmt.Sprintf("<img src=\"%s\" alt=\"%s\"/>", part.ImageURL.URL, part.ImageURL.Detail)
		case schema.ChatMessagePartTypeAudioURL:
			s = fmt.Sprintf("<audio controls><source src=\"%s\" type=\"%s\">Your browser does not support the audio tag.</audio>", part.AudioURL.URL, part.AudioURL.MIMEType)
		case schema.ChatMessagePartTypeVideoURL:
			s = fmt.Sprintf("<video  width=\"320\" height=\"240\" controls><source src=\"%s\" type=\"%s\">Your browser does not support the video tag.</video>", part.VideoURL.URL, part.VideoURL.MIMEType)
		}
		builder.WriteString(s)
	}
	return builder.String()
}

func renderMarkdown(markdown string) (string, error) {
	var buf strings.Builder
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
