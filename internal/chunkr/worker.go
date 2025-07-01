package chunkr

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai"
	"github.com/gitslim/chunkr-ai-sdk/sdk/go/chunkrai/core"
	"github.com/gitslim/go-ragger/internal/db/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type DocumentProcessor struct {
	pool   *pgxpool.Pool
	db     *sqlc.Queries
	logger *slog.Logger
	cfg    *ProcessorConfig
	client *Client
}

type ProcessorConfig struct {
	BatchSize     int
	PollInterval  time.Duration
	StuckTimeout  time.Duration
	WorkerCount   int
	ProcessingTTL time.Duration
}

func NewDocumentProcessor(pool *pgxpool.Pool, db *sqlc.Queries, logger *slog.Logger, cfg *ProcessorConfig, client *Client) *DocumentProcessor {
	return &DocumentProcessor{
		pool:   pool,
		db:     db,
		logger: logger.With("component", "document_processor"),
		cfg:    cfg,
		client: client,
	}
}

func (p *DocumentProcessor) Run(ctx context.Context) error {
	// Запускаем воркеры
	for i := range p.cfg.WorkerCount {
		go p.worker(ctx, i)
	}

	// Запускаем очистку зависших задач
	go p.stuckDocumentsCleaner(ctx)

	<-ctx.Done()
	return nil
}

func (p *DocumentProcessor) worker(ctx context.Context, workerID int) {
	logger := p.logger.With("worker_id", workerID)
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.processBatch(ctx, logger); err != nil {
				logger.Error("failed to process batch", "error", err)
			}
		}
	}
}

func (p *DocumentProcessor) processBatch(ctx context.Context, logger *slog.Logger) error {
	if err := p.resetStuckDocuments(ctx); err != nil {
		logger.Error("failed to reset stuck documents", "error", err)
	}

	docs, err := p.db.GetPendingDocuments(ctx, int32(p.cfg.BatchSize))
	if err != nil {
		return fmt.Errorf("get pending documents: %w", err)
	}

	for _, doc := range docs {
		if err := p.processDocument(ctx, doc.ID); err != nil {
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

func (p *DocumentProcessor) processDocument(ctx context.Context, docID uuid.UUID) error {
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

func (p *DocumentProcessor) resetStuckDocuments(ctx context.Context) error {
	err := p.db.ResetStuckDocuments(ctx)
	return err
}

func (p *DocumentProcessor) stuckDocumentsCleaner(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.StuckTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.resetStuckDocuments(ctx); err != nil {
				p.logger.Error("failed to reset stuck documents", "error", err)
			}
		}
	}
}

func RunDocumentProcessor(lc fx.Lifecycle, logger *slog.Logger, pool *pgxpool.Pool, db *sqlc.Queries, client *Client) {
	cfg := &ProcessorConfig{
		BatchSize:     10,
		PollInterval:  5 * time.Second,
		StuckTimeout:  60 * time.Minute,
		WorkerCount:   3,
		ProcessingTTL: 30 * time.Minute,
	}

	processor := NewDocumentProcessor(pool, db, logger, cfg, client)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := processor.Run(context.Background()); err != nil {
					logger.Error("document processor failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

func (p *DocumentProcessor) createChunkrTask(ctx context.Context, doc sqlc.Document) (string, error) {
	b64 := base64.StdEncoding.EncodeToString(doc.FileData)

	req := &chunkrai.CreateForm{
		File:     b64,
		FileName: &core.Optional[string]{Value: doc.FileName},
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
