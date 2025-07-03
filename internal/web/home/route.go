package home

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/gitslim/go-ragger/internal/util"
	"github.com/go-chi/chi/v5"

	datastar "github.com/starfederation/datastar/sdk/go"
)

func SetupRoutes(rtr chi.Router, logger *slog.Logger, retriever retriever.Retriever) {

	rtr.Get("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		u, _ := util.UserFromContext(ctx)

		s := &Signals{}

		PageHome(r, u, s).Render(ctx, w)
	})

	rtr.Post("/ask", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		s := &Signals{}
		if err := datastar.ReadSignals(r, &s); err != nil {
			http.Error(w, fmt.Sprintf("error unmarshalling form: %s", err), http.StatusBadRequest)
		}

		logger.Debug("chat", "question", &s.Question)

		sse := datastar.NewSSE(w, r)

		message := ChatMessage(ActorUser, s.Question)
		sse.MergeFragmentTempl(message,
			datastar.WithSelectorID("chat"),
			datastar.WithMergeMode(datastar.FragmentMergeModeAppend),
		)

		// Retrieve documents
		documents, err := retriever.Retrieve(ctx, s.Question)
		if err != nil {
			fmt.Printf("Failed to retrieve: %v", err)
		}

		// Print the documents
		for i, doc := range documents {
			fmt.Printf("Document %d:\n", i)
			fmt.Printf("title: %s\n", doc.ID)
			fmt.Printf("content: %s\n", doc.Content)
			fmt.Printf("metadata: %v\n", doc.MetaData)

			message := ChatMessage(ActorAssistant, doc.Content)
			sse.MergeFragmentTempl(message,
				datastar.WithSelectorID("chat"),
				datastar.WithMergeMode(datastar.FragmentMergeModeAppend),
			)
		}

		s.Question = ""
		ss, _ := json.Marshal(&s)
		sse.MergeSignals(ss)

	})
}
