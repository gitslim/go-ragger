package chunker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *Chunker) checkerBatch(ctx context.Context, logger *slog.Logger) error {
	docs, err := p.db.GetChunkingDocuments(ctx, int32(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("get processing documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.checkDocument(ctx, doc.ID); err != nil {
			logger.Error("failed to check document",
				"doc_id", doc.ID,
				"error", err)
		}
	}

	return nil
}

func (p *Chunker) checkDocument(ctx context.Context, docID uuid.UUID) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin document checking transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := p.db.WithTx(tx)

	doc, err := q.LockDocumentForChecking(ctx, docID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("lock document for checking: %w", err)
	}

	res, err := p.getChunkrTask(ctx, doc)
	if err != nil {
		return fmt.Errorf("get chunkr result: %w", err)
	}

	switch res.Status {
	case chunkrai.StatusProcessing, chunkrai.StatusStarting:
		if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
			ID:     doc.ID,
			Status: sqlc.DocumentStatusChunking,
		}); err != nil {
			return fmt.Errorf("failed to mark document as chunking: %w", err)
		}

	case chunkrai.StatusFailed, chunkrai.StatusCancelled:
		p.logger.Error("chunkr task failed", "TaskID", res.TaskId)
		if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
			ID:     doc.ID,
			Status: sqlc.DocumentStatusChunkfail,
		}); err != nil {
			return fmt.Errorf("failed to mark document as chunkfail: %w", err)
		}

	case chunkrai.StatusSucceeded:
		p.logger.Info("chunkr task succeeded", "TaskID", res.TaskId)
		data, err := json.Marshal(res)
		if err != nil {
			return fmt.Errorf("failed marshal chunking result: %w", err)
		}
		_, err = q.SetChunkingResult(ctx, sqlc.SetChunkingResultParams{
			ID:           doc.ID,
			ChunkrResult: data})
		if err != nil {
			return fmt.Errorf("failed to save chunking result: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (p *Chunker) getChunkrTask(ctx context.Context, doc sqlc.Document) (*chunkrai.TaskResponse, error) {
	taskID := doc.ChunkrTaskID.String
	ptr := &taskID

	res, err := p.chunkrAIClient.Task.GetTaskRoute(ctx, ptr, &chunkrai.GetTaskRouteRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get chunkr task %v: %w", doc.ChunkrTaskID.String, err)
	}

	return res, nil
}
