package chunker

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai"
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai/core"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (p *Chunker) senderBatch(ctx context.Context, logger *slog.Logger) error {
	docs, err := p.db.GetPendingDocuments(ctx, int32(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("get pending documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.sendDocument(ctx, doc.ID); err != nil {
			logger.Error("failed to send document to chunking",
				"doc_id", doc.ID,
				"error", err)

			if err := p.db.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
				ID:     doc.ID,
				Status: sqlc.DocumentStatusChunkfail,
			}); err != nil {
				logger.Error("failed to mark document as chunkfail",
					"doc_id", doc.ID,
					"error", err)
			}
		}
	}

	return nil
}

func (p *Chunker) sendDocument(ctx context.Context, docID uuid.UUID) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := p.db.WithTx(tx)

	doc, err := q.LockDocumentForChunking(ctx, docID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("lock document for chunking: %w", err)
	}

	taskID, err := p.createChunkrTask(ctx, doc)
	if err != nil {
		return fmt.Errorf("create chunkr task: %w", err)
	}

	p.logger.Info("chunkr task created",
		"taskId", taskID,
		"doc_id", doc.ID,
		"file_name", doc.FileName)

	doc, err = q.SetChunkrTaskID(ctx, sqlc.SetChunkrTaskIDParams{
		ID:           doc.ID,
		ChunkrTaskID: pgtype.Text{String: taskID, Valid: true},
	})
	if err != nil {
		if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
			ID:     doc.ID,
			Status: sqlc.DocumentStatusChunkfail,
		}); err != nil {
			return fmt.Errorf("failed to mark document as chunkfail: %w", err)
		}
		return fmt.Errorf("create chunkr task: %w", err)
	}

	if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
		ID:     doc.ID,
		Status: sqlc.DocumentStatusChunking,
	}); err != nil {
		return fmt.Errorf("failed to mark document as chunking: %w", err)
	}

	return tx.Commit(ctx)
}

func (p *Chunker) createChunkrTask(ctx context.Context, doc sqlc.Document) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(doc.FileData)
	// modelId := "ollama-qwen3-8b"

	req := &chunkrai.CreateForm{
		File:      b64,
		FileName:  &core.Optional[string]{Value: doc.FileName},
		ExpiresIn: &core.Optional[int]{Value: int(p.cfg.ProcessingTTL.Seconds())},
		ErrorHandling: &core.Optional[chunkrai.ErrorHandlingStrategy]{
			Value: chunkrai.ErrorHandlingStrategyContinue,
		},
		OcrStrategy: &core.Optional[chunkrai.OcrStrategy]{Value: chunkrai.OcrStrategyAuto},
		LlmProcessing: &core.Optional[chunkrai.LlmProcessing]{
			Value: chunkrai.LlmProcessing{
				// ModelId:          &modelId,
				FallbackStrategy: chunkrai.NewFallbackStrategyWithDefaultStringLiteral(),
			},
		},
	}

	res, err := p.chunkrAIClient.Task.CreateTaskRoute(ctx, req)
	if err != nil {
		return "", err
	}

	return res.TaskId, nil
}
