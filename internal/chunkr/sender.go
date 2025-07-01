package chunkr

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

func (p *DocumentProcessor) documentSenderBatch(ctx context.Context, logger *slog.Logger) error {
	docs, err := p.db.GetPendingDocuments(ctx, int32(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("get pending documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.sendDocument(ctx, doc.ID); err != nil {
			logger.Error("failed to process document",
				"doc_id", doc.ID,
				"error", err)

			if err := p.db.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
				ID:     doc.ID,
				Status: sqlc.DocumentStatusFailed,
			}); err != nil {
				logger.Error("failed to mark document as failed",
					"doc_id", doc.ID,
					"error", err)
			}
		}
	}

	return nil
}

func (p *DocumentProcessor) sendDocument(ctx context.Context, docID uuid.UUID) error {
	tx, err := p.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := p.db.WithTx(tx)

	doc, err := q.LockDocumentForProcessing(ctx, docID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			p.logger.Debug("document already processed or not found",
				"doc_id", docID)
			return nil
		}
		return fmt.Errorf("lock document: %w", err)
	}
	if doc.ID == uuid.Nil {
		return fmt.Errorf("document not found or already processed")
	}

	p.logger.Info("processing document",
		"doc_id", doc.ID,
		"file_name", doc.FileName)

	taskID, err := p.createChunkrTask(ctx, doc)
	if err != nil {
		return fmt.Errorf("create chunkr task: %w", err)
	}

	p.logger.Debug("create task", "taskId", taskID)

	doc, err = q.SetChunkrTaskID(ctx, sqlc.SetChunkrTaskIDParams{
		ID:           doc.ID,
		ChunkrTaskID: pgtype.Text{String: taskID, Valid: true},
	})
	if err != nil {
		if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
			ID:     doc.ID,
			Status: sqlc.DocumentStatusFailed,
		}); err != nil {
			return fmt.Errorf("update status: %w", err)
		}
		return fmt.Errorf("create chunkr task: %w", err)
	}

	if err := q.UpdateDocumentStatus(ctx, sqlc.UpdateDocumentStatusParams{
		ID:     doc.ID,
		Status: sqlc.DocumentStatusProcessing,
	}); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return tx.Commit(ctx)
}

func (p *DocumentProcessor) createChunkrTask(ctx context.Context, doc sqlc.Document) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(doc.FileData)

	req := &chunkrai.CreateForm{
		File:      b64,
		FileName:  &core.Optional[string]{Value: doc.FileName},
		ExpiresIn: &core.Optional[int]{Value: 3600},
		ErrorHandling: &core.Optional[chunkrai.ErrorHandlingStrategy]{
			Value: chunkrai.ErrorHandlingStrategyContinue,
		},
		LlmProcessing: &core.Optional[chunkrai.LlmProcessing]{
			Value: chunkrai.LlmProcessing{
				FallbackStrategy: chunkrai.NewFallbackStrategyWithDefaultStringLiteral(),
			},
		},
	}

	res, err := p.client.Task.CreateTaskRoute(ctx, req)
	if err != nil {
		return "", err
	}

	return res.TaskId, nil
}
